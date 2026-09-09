package battle

func (b *Battle) getTargets(actionUnit *Unit, target *SkillTarget) []*Unit {
	lastTarget := b.getUnitByPosition(actionUnit.Current.ActionAddedDamageRecord.Positions...)
	if condition, ok := target.Condition.(*SkillTarget_Position); ok {
		return b.getUnitsByTeamPosition(actionUnit, lastTarget, target.Team, condition.Position)
	}

	if condition, ok := target.Condition.(*SkillTarget_ParamCondition_); ok {
		return b.getUnitsByParam(actionUnit, target.Team, condition.ParamCondition)
	}

	return b.getUnitsByTeamPosition(actionUnit, lastTarget, target.Team, Position_ALL)
}

func (u *Unit) applyDecoy(b *Battle, actionUnit *Unit) *Unit {
	if u.isAllySide(actionUnit) {
		return u
	}

	units := u.getTeamUnits(b, Team_ALLY)
	for _, unit := range units {
		if unit == nil {
			continue
		}

		if unit.getBuffEffect(BuffEffectType_DECOY).GetDecoy().activate() {
			return unit
		}
	}

	return u
}