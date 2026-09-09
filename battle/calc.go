package battle

func newMyBasicStat(paramType ParamType, param Param) isSkill_Effect_Reference {
	return &Skill_Effect_RawStat{&RawStat{Team: Team_ALLY, Position: Position_SELF,
		Stat: &RawStat_BasicStat{&BasicStat{ParamType: paramType, Param: param}}}}
}

func newYourBasicStat(paramType ParamType, param Param) isSkill_Effect_Reference {
	return &Skill_Effect_RawStat{&RawStat{Team: Team_BOTH, Position: Position_TARGET,
		Stat: &RawStat_BasicStat{&BasicStat{ParamType: paramType, Param: param}}}}
}

func newMyBattleStat(battleStat BattleStat) isSkill_Effect_Reference {
	return &Skill_Effect_RawStat{&RawStat{Team: Team_ALLY, Position: Position_SELF, Stat: &RawStat_BattleStat{battleStat}}}
}

func (b *Battle) getReferenceValue(actionUnit, target *Unit, reference isSkill_Effect_Reference) int64 {
	var val int32
	switch r := reference.(type) {
	case *Skill_Effect_RawStat:
		units := b.getUnitsByTeamPosition(actionUnit, []*Unit{target}, r.RawStat.Team, r.RawStat.Position)
		switch s := r.RawStat.Stat.(type) {
		case *RawStat_BasicStat:
			for _, unit := range units {
				val += unit.getParam(s.BasicStat.ParamType, s.BasicStat.Param)
			}
		case *RawStat_BattleStat:
			for _, unit := range units {
				val += unit.getBattleStat(s.BattleStat)
			}
		}
	case *Skill_Effect_ComputedStat:
		switch r.ComputedStat {
		case ComputedStat_INT_HEALING_MODIFIER:
			val = (actionUnit.getParam(ParamType_CURRENT, Param_INT) + target.getParam(ParamType_CURRENT, Param_PHY)) / 2
		case ComputedStat_PHY_HEALING_MODIFIER:
			val = (actionUnit.getParam(ParamType_CURRENT, Param_PHY) + target.getParam(ParamType_CURRENT, Param_INT)) / 2
		}
	}

	return int64(val)
}

func (e *Skill_Effect) getReferenceParam() *Param {
	r := e.GetRawStat()
	if b := r.GetBasicStat(); b != nil {
		switch b.Param {
		case Param_PHY, Param_INT:
			if r.Position == Position_SELF && b.ParamType == ParamType_CURRENT {
				return &b.Param
			}
		case Param_HP, Param_AGI:
			return &b.Param // self以外も念のため
		}
	}

	return nil
}

// damage計算
func (b *Battle) computeDamage(actionUnit, target *Unit, point int64, param *Param) int32 {
	return b.computeDamageBounded(actionUnit, target, point, param, false)
}
func (b *Battle) computeDamageBounded(actionUnit, target *Unit, point int64, param *Param, bounded bool) int32 {
	damage := point
	if param != nil {
		switch *param {
		case Param_PHY, Param_INT:
			attackerModifier, defenderModifier := actionUnit.getDamageModifier(*param), target.getDamageModifier(*param)
			if b.getRandom(100) < attackerModifier.CriticalRate {
				damage += int64(actionUnit.getParam(ParamType_BASE, *param))
				target.Current.CriticalJudg = true
			}

			// damage補正
			damage = damage * int64(100+attackerModifier.DealDamageBonus) / 100 * int64(100+defenderModifier.TakeDamageBonus) / 100

			// cut率
			reductionRate := target.getParam(ParamType_CURRENT, *param)/2 + defenderModifier.ReductionRateBonus
			if reductionRate > defenderModifier.ReductionRateCap {
				reductionRate = defenderModifier.ReductionRateCap
			}

			if reductionRate >= 100 {
				return 0
			}

			damage = damage * int64(100-reductionRate) / 100
		case Param_HP, Param_AGI:
			// オーラ
			target.getDamageReductionBuff(*param).apply(&damage)
		}
	}

	// 全体倍率
	if multi, ok := target.getPassiveEffect(PassiveEffectType_TAKE_DAMAGE_RATE); ok {
		damage = damage * int64(multi) / 100
	}

	if bounded {
		return boundedCustomNumber(damage)
	}
	return int32(damage)
}

func (u *Unit) getDamageModifier(param Param) *DamageModifier {
	switch param {
	case Param_PHY:
		return u.Current.PhyDamageModifier
	case Param_INT:
		return u.Current.IntDamageModifier
	}

	return nil
}

func (d *DamageModifier) Copy() *DamageModifier {
	return &DamageModifier{
		ReductionRateCap:   d.ReductionRateCap,
		ReductionRateBonus: d.ReductionRateBonus,
		CriticalRate:       d.CriticalRate,
		DealDamageBonus:    d.DealDamageBonus,
		TakeDamageBonus:    d.TakeDamageBonus,
	}
}

func (p *Param) isSpecialReference() bool {
	if p == nil {
		return true
	}

	return !(*p == Param_PHY || *p == Param_INT)
}

func (u *Unit) applyShield(damage int32) int32 {
	var remain int32
	if damage < u.Current.Shield {
		u.Current.Shield -= damage
	} else {
		remain = damage - u.Current.Shield
		u.Current.Shield = 0
	}

	return remain
}
