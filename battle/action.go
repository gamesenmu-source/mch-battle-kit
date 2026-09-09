package battle

import (
	"github.com/golang/protobuf/proto"
)

// 毒・睡眠・混乱・恐怖・出血・スタン・呪い・魅了・重力
type beforeActiveSkillEffect interface{ applyBeforeActiveSkill(*Battle, *Unit) }
type afterActiveSkillEffect interface{ applyAfterActiveSkill(*Battle, *Unit) }
type afterActionEffect interface{ applyAfterAction(*Unit) }

func (u *Unit) useActiveSkill(b *Battle, skill *Skill) {
	b.processBeforeActiveSkill(u)
	skill = u.applyConfusion(skill)

	// 睡眠・気絶・バインド(確率判定)
	for _, statusEffect := range orderedStatuses(u.Current.StatusEffects) {
		if e, ok := statusEffect.Effect.(activeSkillSkipEffect); !ok {
			continue
		} else if e.skipActiveSkill(b) {
			b.processAfterAction(u, skipAction)
			return
		}
	}

	skill.occurEffects(b, u)

	b.processAfterActiveSkill(u, skill)
}

func (u *Unit) usePassiveSkill(b *Battle, skill *Skill) bool {
	if skill.SkillId == 0 {
		return false
	}

	// 発動判定
	if !u.canTriggerSkill(skill) || !skill.isAvailable(b, u) ||
		b.getRandom(100) >= u.modifyTriggerRate(skill.Trigger.TriggerRate) {
		return false
	}

	b.processBeforeAction()

	// 睡眠・混乱・恐怖・気絶・呪い
	for _, statusEffect := range orderedStatuses(u.Current.StatusEffects) {
		if _, ok := statusEffect.Effect.(passiveSkillSkipEffect); ok {
			b.processAfterAction(u, skipAction)
			return true
		}
	}

	skill.occurEffects(b, u)

	b.processAfterAction(u, skill)
	return true
}

func (u *Unit) useResurrection(b *Battle) bool {
	if u.isAlive() || u.getBuffEffect(BuffEffectType_RESURRECTION) == nil {
		return false
	}

	b.processBeforeAction()
	resurrectionSkill.occurEffects(b, u)
	delete(u.Current.BuffEffects, int32(BuffEffectType_RESURRECTION))
	b.processAfterAction(u, resurrectionSkill)
	return true
}

func (b *Battle) processBeforeAction() {
	b.ActionCounts++
	action := &BattleAction{
		Count: b.ActionCounts,
	}

	b.Actions = append(b.Actions, action)
}

func (b *Battle) processAfterAction(actionUnit *Unit, skill *Skill) {
	action := b.getCurrentAction()
	action.ActionPosition = actionUnit.Current.Position
	action.Skill = skill.SkillId
	action.Units = make([]*CurrentUnit, len(b.Units))
	for i, unit := range b.Units {
		// barrier
		for _, statusEffect := range orderedStatuses(unit.Current.StatusEffects) {
			if i, ok := statusEffect.Effect.(afterActionEffect); ok {
				i.applyAfterAction(unit)
			}
		}

		// decoy
		for _, buffEffect := range orderedBuffs(unit.Current.BuffEffects) {
			if i, ok := buffEffect.Effect.(afterActionEffect); ok {
				i.applyAfterAction(unit)
			}
		}

		action.Units[i] = proto.Clone(unit.Current).(*CurrentUnit)
		unit.Current.CriticalJudg = false
		unit.Current.ActionAddedDamageRecord.Reset()
		unit.Current.ActionAddedHealingRecord.Reset()
		action.AttackerTakenDamage = b.AttackerTakenDamage
		action.DefenderTakenDamage = b.DefenderTakenDamage
	}

	b.checkEnd()
}

func (b *Battle) processBeforeActiveSkill(unit *Unit) {
	b.processBeforeAction()
	unit.Current.ActiveCounts++
	for _, statusEffect := range orderedStatuses(unit.Current.StatusEffects) {
		if i, ok := statusEffect.Effect.(beforeActiveSkillEffect); ok {
			i.applyBeforeActiveSkill(b, unit)
		}
	}
}

func (b *Battle) processAfterActiveSkill(unit *Unit, skill *Skill) {
	// only curse
	for _, statusEffect := range orderedStatuses(unit.Current.StatusEffects) {
		if i, ok := statusEffect.Effect.(afterActiveSkillEffect); ok {
			i.applyAfterActiveSkill(b, unit)
		}
	}

	unit.Current.States[int32(State_AFTER_ACTIVE_SKILL)] = true
	b.ActiveActionCounts++ // 99関連?
	b.processAfterAction(unit, skill)
}

func (u *Unit) dealDamage(b *Battle, target *Unit, damage int32) {
	if damage == 0 {
		return
	}

	diff := target.takeDamage(b, damage)
	target.recoverStunOnDamaged()
	u.Current.ActionAddedDamageRecord.add(diff, target)
	if b.isActivePhase() {
		u.addCurseDamage(diff)
		target.Current.ActiveSkillTakenDamageRecord.add(diff, u)
	}
}

func (u *Unit) heal(target *Unit, point int64) {
	if point == 0 {
		return
	}

	amount := int32(point)
	if bonus, ok := target.getPassiveEffect(PassiveEffectType_TAKE_HEALING_BONUS); ok {
		amount = int32(point * int64(100+bonus) / 100)
	}

	diff := amount
	if target.Current.Hp+amount < target.Base.Hp {
		target.Current.Hp += amount
	} else {
		diff = target.Base.Hp - target.Current.Hp
		target.Current.Hp = target.Base.Hp
	}

	u.Current.ActionAddedHealingRecord.add(diff, target)
	delete(target.Current.States, int32(State_AFTER_DEATH)) // 死亡後フラグのキャンセル
}

func (u *Unit) takeDamage(b *Battle, damage int32) int32 {
	diff := damage
	if damage < u.Current.Hp {
		u.Current.Hp -= damage
	} else {
		diff = u.Current.Hp
		u.death()
	}

	if u.getOwnSide() == Attacker {
		b.AttackerTakenDamage += diff
	} else {
		b.DefenderTakenDamage += diff
	}

	return diff
}

func (u *Unit) parameterDown(param Param, point int32) {
	update := u.getParam(ParamType_CURRENT, param) - point
	if update < 1 {
		update = 1
	}

	u.setParam(param, update)
}

func (u *Unit) parameterUp(param Param, point int32) {
	update := u.getParam(ParamType_CURRENT, param) + point
	if update > 999 {
		update = 999
	}

	u.setParam(param, update)
}

func (u *Unit) changeSkill(b *Battle, skillId uint32) {
	if b.isActivePhase() {
		u.Current.Actives[(u.Current.ActiveCounts-1)%3] = skillId
	} else {
		u.Current.Passive = skillId
	}
}

func (record *DamageRecord) add(amount int32, target *Unit) {
	record.Amount += amount
	record.Positions = addPosition(record.Positions, target.Current.Position)
}
