package battle

type BuffEffectMaster struct {
	Constructor func() *BuffEffect
}

var (
	BuffEffects = map[BuffEffectType]*BuffEffectMaster{
		BuffEffectType_RESURRECTION: {func() *BuffEffect {
			return &BuffEffect{Effect: &BuffEffect_Resurrection_{&BuffEffect_Resurrection{}}}
		}},
		BuffEffectType_DECOY: {func() *BuffEffect {
			return &BuffEffect{Effect: &BuffEffect_Decoy_{Decoy: &BuffEffect_Decoy{RemainingUses: 1}}}
		}},
		BuffEffectType_AGI_DAMAGE_REDUCTION: {func() *BuffEffect {
			return &BuffEffect{Effect: &BuffEffect_AgiDamageReduction{&BuffEffect_DamageReduction{}}}
		}},
		BuffEffectType_HP_DAMAGE_REDUCTION: {func() *BuffEffect {
			return &BuffEffect{Effect: &BuffEffect_HpDamageReduction{&BuffEffect_DamageReduction{}}}
		}},
	}
)

func (u *Unit) addBuffEffect(buffType BuffEffectType) {
	effect := BuffEffects[buffType].Constructor()
	if buff := u.getBuffEffect(buffType); buff != nil {
		if e, ok := buff.Effect.(Overridable); ok {
			e.override(effect)
		}
	} else {
		u.Current.BuffEffects[int32(buffType)] = effect
	}
}

type Overridable interface{ override(*BuffEffect) }

// decoy
func (decoy *BuffEffect_Decoy_) applyAfterAction(u *Unit) {
	if decoy.Decoy.RemainingUses == 0 {
		delete(u.Current.BuffEffects, int32(BuffEffectType_DECOY))
	}

	decoy.Decoy.Activated = false
}

func (decoy *BuffEffect_Decoy) activate() bool {
	if decoy == nil {
		return false
	}

	if decoy.Activated {
		return true
	}

	decoy.Activated = true
	decoy.RemainingUses--
	return true
}

func (decoy *BuffEffect_Decoy_) override(buff *BuffEffect) {
	if e, ok := buff.Effect.(*BuffEffect_Decoy_); ok {
		decoy.Decoy.RemainingUses = e.Decoy.RemainingUses
	}
}

// damage reduction
func (buff *BuffEffect_DamageReduction) apply(damage *int64) {
	if buff == nil {
		return
	}

	*damage = *damage * (100 - int64(buff.RedunctionRate)) / 100
}

func (buff *BuffEffect_AgiDamageReduction) applyAfterAction(u *Unit) {
	if buff.AgiDamageReduction.RemainingActions == 0 {
		delete(u.Current.BuffEffects, int32(BuffEffectType_AGI_DAMAGE_REDUCTION))
	}

	buff.AgiDamageReduction.RemainingActions--
}

func (buff *BuffEffect_HpDamageReduction) applyAfterAction(u *Unit) {
	if buff.HpDamageReduction.RemainingActions == 0 {
		delete(u.Current.BuffEffects, int32(BuffEffectType_HP_DAMAGE_REDUCTION))
	}

	buff.HpDamageReduction.RemainingActions--
}

func (u *Unit) getDamageReductionBuff(param Param) *BuffEffect_DamageReduction {
	switch param {
	case Param_HP:
		return u.getBuffEffect(BuffEffectType_HP_DAMAGE_REDUCTION).GetHpDamageReduction()
	case Param_AGI:
		return u.getBuffEffect(BuffEffectType_AGI_DAMAGE_REDUCTION).GetAgiDamageReduction()
	}

	return nil
}

func (u *Unit) removeBuffEffectOnDeath() {
	delete(u.Current.BuffEffects, int32(BuffEffectType_DECOY))
}

var (
	resurrectionSkill = &Skill{
		SkillId: resurrectionSkillId,
		Effects: []*Skill_Effect{
			{
				Target: &SkillTarget{Team: Team_ALLY, Condition: &SkillTarget_Position{Position_SELF}},
				Param:  Skill_Effect_REVIVE, SuccessRate: 100,
				Value: &Skill_Effect_Rate_{&Skill_Effect_Rate{Min: 1, Max: 1}}},
			{
				Target: &SkillTarget{Team: Team_ALLY, Condition: &SkillTarget_Position{Position_SELF}},
				Param:  Skill_Effect_CHARGE, SuccessRate: 100,
				Value: &Skill_Effect_Rate_{&Skill_Effect_Rate{Min: 2000, Max: 2000}}},
		},
	}
)