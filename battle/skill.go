package battle

import (
	"context"
)

func GetSkill(id uint32) (*Skill, error) {
	return r.GetSkill(id)
}

func SaveSkill(ctx context.Context, skill *Skill) error {
	return r.SaveSkill(ctx, skill)
}

func Reload(ctx context.Context) error {
	return r.Reload(ctx)
}

func (skill *Skill) occurEffects(b *Battle, u *Unit) {
	for _, effect := range skill.Effects {
		skillTarget := effect.Target
		if b.isActivePhase() {
			skillTarget = u.applyCharm(effect.Target)
		}

		targets := b.getTargets(u, skillTarget)
		for _, target := range targets {
			effect.occurEffect(b, u, target, b.authoredSkill(skill.SkillId))
		}
	}

	if usage, ok := u.Current.SkillUsages[skill.SkillId]; ok && usage.RemainingUses > 0 {
		usage.RemainingUses--
	}
}

func (effect *Skill_Effect) occurEffect(b *Battle, u, target *Unit, bounded bool) {
	// targeting可能か判定
	if !effect.targetable(target) {
		return
	}

	target = target.applyDecoy(b, u)
	b.getCurrentAction().EffectPositions = addPosition(b.getCurrentAction().EffectPositions, target.Current.Position)

	// barrier
	if target.applyBarrier(!u.isAllySide(target)) {
		return
	}

	switch v := effect.Value.(type) {
	case *Skill_Effect_StatusEffectType: // !!!success rateに補正が入る特例!!!
		successRate := effect.SuccessRate + effect.AdditionalEffect.getAdditionalEffectRate(b, u, target)
		if effect.IsDamage {
			u.addStatusEffect(b, target, v.StatusEffectType, successRate)
		} else {
			target.cureStatusEffect(b, v.StatusEffectType, successRate)
		}

		return
	case *Skill_Effect_BuffEffectType:
		successRate := effect.SuccessRate + effect.AdditionalEffect.getAdditionalEffectRate(b, u, target)
		if b.getRandom(100) < successRate {
			target.addBuffEffect(v.BuffEffectType)
		}

		return
	case *Skill_Effect_SkillId:
		successRate := effect.SuccessRate + effect.AdditionalEffect.getAdditionalEffectRate(b, u, target)
		if b.getRandom(100) < successRate {
			u.changeSkill(b, v.SkillId)
		}

		return
	}

	if b.getRandom(100) >= effect.SuccessRate {
		return
	}

	point := int64(effect.getCurrentRate(b, u, target))
	if effect.Reference != nil {
		point = b.getReferenceValue(u, target, effect.Reference) * point / 100
	}
	if bounded {
		point = int64(boundedCustomNumber(point))
	}

	switch effect.Param {
	case Skill_Effect_HP:
		if effect.IsDamage {
			// damage計算
			param := effect.getReferenceParam()
			damage := b.computeDamageBounded(u, target, point, param, bounded)

			// shield
			if param.isSpecialReference() {
				damage = target.applyShield(damage)
			}

			u.dealDamage(b, target, damage)
		} else if !target.applyProstration(b) {
			if bounded && point > int64(target.Base.Hp) {
				point = int64(target.Base.Hp)
			}
			u.heal(target, point)
		}

		return
	case Skill_Effect_REVIVE:
		if bounded && point > int64(target.Base.Hp) {
			point = int64(target.Base.Hp)
		}
		u.heal(target, point)
		return
	case Skill_Effect_CHARGE:
		if bounded {
			if effect.IsDamage {
				if point > int64(target.Current.Charge) {
					point = int64(target.Current.Charge)
				}
				point = -point
			} else if target.applyGravity(b, 0) {
				return
			}
			next := boundedCustomNumber(int64(target.Current.Charge) + point)
			delta := int64(next) - int64(target.Current.Charge)
			target.Current.BonusCharge = boundedCustomNumber(int64(target.Current.BonusCharge) + delta)
			target.Current.Charge = next
			return
		}
		if effect.IsDamage {
			target.decreaseCharge(int32(point))
		} else if !target.applyGravity(b, 0) {
			target.increaseCharge(int32(point))
		}

		return
	case Skill_Effect_CRITICAL_PHY:
		target.getDamageModifier(Param_PHY).CriticalRate += int32(point)
		return
	case Skill_Effect_CRITICAL_INT:
		target.getDamageModifier(Param_INT).CriticalRate += int32(point)
		return
	case Skill_Effect_SHIELD:
		if target.Current.Shield < int32(point) {
			target.Current.Shield = int32(point)
		}

		return
	case Skill_Effect_DAMAGE_CUT:
		target.getDamageModifier(Param_PHY).ReductionRateCap += int32(point)
		target.getDamageModifier(Param_INT).ReductionRateCap += int32(point)
		return
	}

	param := effect.Param.getSkillTargetParam(target)
	if param == nil {
		return
	}

	// reduction
	if resistance, ok := target.getPassiveEffect(PassiveEffectType_RESISTANCE_PARAMETER_EFFECT_RATE); ok {
		point = point * int64(resistance) / 100
	}

	if point == 0 {
		return
	}

	if effect.IsDamage {
		target.parameterDown(*param, int32(point))
	} else if !target.applyGravity(b, int32(point)) {
		target.parameterUp(*param, int32(point))
	}
}

func (e *Skill_Effect) targetable(target *Unit) bool {
	return (target.isAlive() && e.Param != Skill_Effect_REVIVE) ||
		(target.isDeath() && e.Param == Skill_Effect_REVIVE)
}

func (e *Skill_Effect) getCurrentRate(b *Battle, actionUnit, target *Unit) int32 {
	var rate int32
	if r := e.GetRate(); r != nil {
		rate = r.Min
		if r.Min < r.Max {
			rate += b.getRandom(r.Max - r.Min)
		}
	}

	return rate + e.AdditionalEffect.getAdditionalEffectRate(b, actionUnit, target)
}

func (effect *Skill_AdditionalEffect) getAdditionalEffectRate(b *Battle, actionUnit, target *Unit) int32 {
	if effect == nil {
		return 0
	}

	switch v := effect.Condition.(type) {
	case *Skill_AdditionalEffect_StatusEffectType:
		if target.getStatusEffect(v.StatusEffectType) != nil {
			return effect.Rate
		}
	case *Skill_AdditionalEffect_BuffEffectType:
		if target.getBuffEffect(v.BuffEffectType) != nil {
			return effect.Rate
		}
	case *Skill_AdditionalEffect_SeriesBonusType:
		switch v.SeriesBonusType {
		case SeriesBonusType_INVERSE:
			return actionUnit.getEnemyMaxDuplicateSeries(b) * effect.Rate
		case SeriesBonusType_DIVERSE:
			return actionUnit.getDefferentSeries(b) * effect.Rate
		}
	}

	return 0
}

func (param Skill_Effect_Param) getSkillTargetParam(u *Unit) *Param {
	res := new(Param)
	switch param {
	case Skill_Effect_PHY:
		*res = Param_PHY
	case Skill_Effect_INT:
		*res = Param_INT
	case Skill_Effect_AGI:
		*res = Param_AGI
	case Skill_Effect_PHY_INT_HIGHER:
		res = u.getMaxParam(Param_PHY, Param_INT)
	case Skill_Effect_PHY_INT_LOWER:
		res = u.getMinParam(Param_PHY, Param_INT)
	case Skill_Effect_ALL_PARAMS_HIGHEST:
		res = u.getMaxParam(Param_PHY, Param_INT, Param_AGI)
	case Skill_Effect_ALL_PARAMS_LOWEST:
		res = u.getMinParam(Param_PHY, Param_INT, Param_AGI)
	}

	return res
}

func (skill *Skill) isAvailable(b *Battle, u *Unit) bool {
	if skill.Trigger == nil || u.Current.SkillUsages[skill.SkillId].Used || u.Current.SkillUsages[skill.SkillId].RemainingUses == 0 {
		return false
	}

	flag := skill.Trigger.isTriggerConditionMet(b, u)
	delete(u.Current.States, int32(State_AFTER_DEATH))
	if !flag {
		return false
	}

	u.Current.SkillUsages[skill.SkillId].Used = true
	return true
}

const Infinity = -1

func addSkillChangeUsage(skill *Skill, usages map[uint32]*SkillUsage, lookup SkillLookup) error {
	for _, effect := range skill.Effects {
		if s, ok := effect.Value.(*Skill_Effect_SkillId); ok {
			if _, exist := usages[s.SkillId]; !exist {
				newSkill, err := lookup(s.SkillId)
				if err != nil {
					return err
				}

				usage := &SkillUsage{RemainingUses: newSkill.RemainingUses}
				if trigger := newSkill.Trigger; trigger != nil {
					usage.DeathTrigger = trigger.isDeathTrigger()
				}

				usages[s.SkillId] = usage
				err = addSkillChangeUsage(newSkill, usages, lookup)
				if err != nil {
					return err
				}
			}
		}
	}

	return nil
}

func NewSkillUsages(skillId uint32) (map[uint32]*SkillUsage, error) {
	return NewSkillUsagesWithLookup(skillId, GetSkill)
}
func NewSkillUsagesWithLookup(skillId uint32, lookup SkillLookup) (map[uint32]*SkillUsage, error) {
	usages := map[uint32]*SkillUsage{}
	skill, err := lookup(skillId)
	if err != nil {
		return nil, err
	}

	usage := &SkillUsage{RemainingUses: skill.RemainingUses}
	if trigger := skill.Trigger; trigger != nil {
		usage.DeathTrigger = trigger.isDeathTrigger()
	}

	usages[skillId] = usage
	err = addSkillChangeUsage(skill, usages, lookup)
	if err != nil {
		return nil, err
	}

	return usages, nil
}
