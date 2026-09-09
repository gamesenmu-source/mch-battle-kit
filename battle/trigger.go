package battle

type matchable interface {
	met([]*Unit) []*Unit
}

var unitStateFuncs = map[State]func(u *Unit) bool{
	State_AFTER_ACTIVE_SKILL_TAKEN_DAMAGE: func(u *Unit) bool { return u.Current.ActiveSkillTakenDamageRecord.Amount > 0 },
	State_HAS_STATUS_EFFECT:               func(u *Unit) bool { return !u.isNormal() },
}

func (trigger *Trigger) isTriggerConditionMet(b *Battle, u *Unit) bool {
	var targets []*Unit
	for _, condition := range trigger.Conditions {
		candidates := b.getUnitsByTeamPosition(u, targets, condition.Team, condition.Position)
		if i, ok := condition.Condition.(matchable); ok {
			targets = i.met(candidates)
		} else {
			targets = candidates
		}

		if len(targets) == 0 {
			return false
		}
	}

	return true
}

func (condition *Condition_ParamCondition_) met(units []*Unit) []*Unit {
	var res []*Unit
	var baseTotal, currentTotal int32
	for _, unit := range units {
		base, current := unit.getParam(ParamType_BASE, condition.ParamCondition.Param), unit.getParam(ParamType_CURRENT, condition.ParamCondition.Param)
		baseTotal += base
		currentTotal += current
		if (condition.ParamCondition.Calc == Condition_ParamCondition_SOMEONE_UNDER && current*100 < base*condition.ParamCondition.Rate) ||
			(condition.ParamCondition.Calc == Condition_ParamCondition_SOMEONE_OVER && current*100 >= base*condition.ParamCondition.Rate) {
			res = append(res, unit)
		}
	}

	if (condition.ParamCondition.Calc == Condition_ParamCondition_UNDER && currentTotal*100 < baseTotal*condition.ParamCondition.Rate) ||
		(condition.ParamCondition.Calc == Condition_ParamCondition_OVER && currentTotal*100 >= baseTotal*condition.ParamCondition.Rate) {
		return units
	}

	return res
}

func (condition *Condition_StateCondition) met(units []*Unit) []*Unit {
	var res []*Unit
	for _, unit := range units {
		if unit.hasState(condition.StateCondition) {
			res = append(res, unit)
		}
	}

	return res
}

func (u *Unit) hasState(state State) bool {
	if f, ok := unitStateFuncs[state]; ok {
		return f(u)
	}

	// default
	return u.Current.States[int32(state)]
}

func (u *Unit) canTriggerSkill(skill *Skill) bool {
	if skill.Trigger == nil {
		return u.isAlive()
	}

	deathTrigger := u.Current.SkillUsages[skill.SkillId].DeathTrigger
	return (u.isAlive() && !deathTrigger) || (u.isDeath() && deathTrigger)
}

func (u *Unit) modifyTriggerRate(baseRate int32) int32 {
	triggerRate := baseRate

	// 逆境
	if bonus, ok := u.getPassiveEffect(PassiveEffectType_TRIGGER_RATE_INCREASE_BY_HP); ok {
		triggerRate += bonus * (u.Base.Hp - u.Current.Hp) / u.Base.Hp
	}

	// 衰弱・バインド
	for _, statusEffect := range orderedStatuses(u.Current.StatusEffects) {
		if s, ok := statusEffect.Effect.(triggerRateEffect); ok {
			s.reduceTriggerRate(&triggerRate)
		}
	}

	return triggerRate
}

func (t *Trigger) isDeathTrigger() bool {
	for _, condition := range t.Conditions {
		if c, ok := condition.Condition.(*Condition_StateCondition); ok {
			if c.StateCondition == State_AFTER_DEATH && condition.Position == Position_SELF {
				return true
			}
		}
	}

	return false
}