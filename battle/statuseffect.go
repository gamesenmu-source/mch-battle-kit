package battle

type StatusEffectMaster struct {
	StatusEffectType   StatusEffectType
	ReferenceParamType ParamType
	ReferenceParams    []Param
	Constructor        func(*Unit) isStatusEffect_Effect
}

type StatusEffectSkill uint32

const (
	minStatusEffectRate, maxStatusEffectRate = 15, 75
	minPoisonDamageRate, maxPoisonDamageRate = 5, 10

	sleepLimit = 3
	bleedLimit = 5
	curseLimit = 5
	charmLimit = 2

	sleepRecoveryRate       = 50
	confusionRecoveryRate   = 60
	stunRecoveryRate        = 40
	fearRecoveryRate        = 30
	prostrationRecoveryRate = 35

	prostrationSkipRate = 50
	bindSkipRate        = 50
	curseDamageRate     = 20

	skipActionId = 0
)

var (
	AllStatusEffects = []StatusEffectType{
		StatusEffectType_POISON, StatusEffectType_SLEEP, StatusEffectType_CONFUSION, StatusEffectType_FEAR,
		StatusEffectType_BARRIER, StatusEffectType_BLEED, StatusEffectType_STUN, StatusEffectType_PROSTRATION,
		StatusEffectType_CURSE, StatusEffectType_CHARM, StatusEffectType_BIND, StatusEffectType_GRAVITY,
	}

	StatusEffects = map[StatusEffectType]*StatusEffectMaster{
		StatusEffectType_POISON: {
			StatusEffectType:   StatusEffectType_POISON,
			ReferenceParamType: ParamType_CURRENT,
			ReferenceParams:    []Param{Param_INT},
			Constructor: func(*Unit) isStatusEffect_Effect {
				return &StatusEffect_Poison_{&StatusEffect_Poison{}}
			},
		},
		StatusEffectType_SLEEP: {
			StatusEffectType:   StatusEffectType_SLEEP,
			ReferenceParamType: ParamType_CURRENT,
			ReferenceParams:    []Param{Param_INT},
			Constructor: func(*Unit) isStatusEffect_Effect {
				return &StatusEffect_Sleep_{&StatusEffect_Sleep{RemainingUses: sleepLimit}}
			},
		},
		StatusEffectType_CONFUSION: {
			StatusEffectType:   StatusEffectType_CONFUSION,
			ReferenceParamType: ParamType_CURRENT,
			ReferenceParams:    []Param{Param_INT},
			Constructor: func(*Unit) isStatusEffect_Effect {
				return &StatusEffect_Confusion_{&StatusEffect_Confusion{}}
			},
		},
		StatusEffectType_FEAR: {
			StatusEffectType:   StatusEffectType_FEAR,
			ReferenceParamType: ParamType_CURRENT,
			ReferenceParams:    []Param{Param_PHY},
			Constructor: func(*Unit) isStatusEffect_Effect {
				return &StatusEffect_Fear_{&StatusEffect_Fear{}}
			},
		},
		StatusEffectType_BARRIER: {
			StatusEffectType: StatusEffectType_BARRIER,
			Constructor: func(*Unit) isStatusEffect_Effect {
				return &StatusEffect_Barrier_{&StatusEffect_Barrier{}}
			},
		},
		StatusEffectType_BLEED: {
			StatusEffectType:   StatusEffectType_BLEED,
			ReferenceParamType: ParamType_CURRENT,
			ReferenceParams:    []Param{Param_PHY},
			Constructor: func(actionUnit *Unit) isStatusEffect_Effect {
				damage := actionUnit.Current.Phy / 10
				return &StatusEffect_Bleed_{&StatusEffect_Bleed{RemainingUses: bleedLimit, Damage: damage}}
			},
		},
		StatusEffectType_STUN: {
			StatusEffectType:   StatusEffectType_STUN,
			ReferenceParamType: ParamType_CURRENT,
			ReferenceParams:    []Param{Param_PHY},
			Constructor: func(*Unit) isStatusEffect_Effect {
				return &StatusEffect_Stun_{&StatusEffect_Stun{}}
			},
		},
		StatusEffectType_PROSTRATION: {
			StatusEffectType:   StatusEffectType_PROSTRATION,
			ReferenceParamType: ParamType_BASE,
			ReferenceParams:    []Param{Param_PHY, Param_INT},
			Constructor: func(*Unit) isStatusEffect_Effect {
				return &StatusEffect_Prostration_{&StatusEffect_Prostration{}}
			},
		},
		StatusEffectType_CURSE: {
			StatusEffectType:   StatusEffectType_CURSE,
			ReferenceParamType: ParamType_BASE,
			ReferenceParams:    []Param{Param_PHY, Param_INT},
			Constructor: func(actionUnit *Unit) isStatusEffect_Effect {
				return &StatusEffect_Curse_{&StatusEffect_Curse{RemainingUses: curseLimit}}
			},
		},
		StatusEffectType_CHARM: {
			StatusEffectType:   StatusEffectType_CHARM,
			ReferenceParamType: ParamType_BASE,
			ReferenceParams:    []Param{Param_PHY, Param_INT},
			Constructor: func(*Unit) isStatusEffect_Effect {
				return &StatusEffect_Charm_{&StatusEffect_Charm{RemainingUses: charmLimit}}
			},
		},
		StatusEffectType_BIND: {
			StatusEffectType:   StatusEffectType_BIND,
			ReferenceParamType: ParamType_CURRENT,
			ReferenceParams:    []Param{Param_AGI},
			Constructor: func(*Unit) isStatusEffect_Effect {
				return &StatusEffect_Bind_{&StatusEffect_Bind{}}
			},
		},
		StatusEffectType_GRAVITY: {
			StatusEffectType:   StatusEffectType_GRAVITY,
			ReferenceParamType: ParamType_BASE,
			ReferenceParams:    []Param{Param_HP},
			Constructor: func(*Unit) isStatusEffect_Effect {
				return &StatusEffect_Gravity_{&StatusEffect_Gravity{}}
			},
		},
	}
	bindRecoveryRate = []int32{70, 20, 10}
)

func (s *StatusEffectMaster) getSuccessRate(actionUnit, target *Unit, rate int32) int32 {
	if rate >= 100 { // 100以上必中
		return 100 - target.getStatusEffectModifier(s.StatusEffectType).ResistanceRate
	}

	successRate := rate
	if n := int32(len(s.ReferenceParams)); n > 0 {
		var sum int32
		for _, p := range s.ReferenceParams {
			sum += actionUnit.getParam(s.ReferenceParamType, p)
		}

		successRate += sum / (n * 40)
	}

	// static := len(s.ReferenceParams) == 0
	// var n, d int64
	// for _, p := range s.ReferenceParams {
	// 	n += int64(actionUnit.getParam(s.ReferenceParamType, p))
	// 	d += int64(target.getParam(s.ReferenceParamType, p))
	// }

	// if d > 0 {
	// 	successRate = int32(int64(successRate) * n / d)
	// } else if n > 0 {
	// 	successRate = 100
	// }

	successRate += actionUnit.getStatusEffectModifier(s.StatusEffectType).InflictBonus - target.getStatusEffectModifier(s.StatusEffectType).ResistanceBonus
	if successRate > 100 {
		successRate = 100
	}

	// if !static { // バリア以外
	// 	if successRate < minStatusEffectRate {
	// 		successRate = minStatusEffectRate
	// 	} else if successRate > maxStatusEffectRate {
	// 		successRate = maxStatusEffectRate
	// 	}
	// }

	return successRate * (100 - target.getStatusEffectModifier(s.StatusEffectType).ResistanceRate) / 100
}

func (u *Unit) addStatusEffect(b *Battle, target *Unit, statusEffectType StatusEffectType, rate int32) {
	if statusEffectType == StatusEffectType_ANY_EFFECT {
		return // 未実装につき
	}

	// 付与済みなら何もしない
	if target.getStatusEffect(statusEffectType) != nil {
		return
	}

	master := StatusEffects[statusEffectType]
	successRate := master.getSuccessRate(u, target, rate)
	if b.getRandom(100) >= successRate {
		return
	}

	statusEffect := &StatusEffect{Effect: master.Constructor(u)}
	target.Current.StatusEffects[int32(statusEffectType)] = statusEffect
}

func (u *Unit) cureStatusEffect(b *Battle, statusEffectType StatusEffectType, rate int32) {
	if statusEffectType == StatusEffectType_ANY_EFFECT {
		for effectType := range u.Current.StatusEffects {
			u.cureStatusEffect(b, StatusEffectType(effectType), rate)
		}
	}

	if u.getStatusEffect(statusEffectType) == nil || b.getRandom(100) >= rate {
		return
	}

	u.removeStatusEffect(statusEffectType)
	return
}

type triggerRateEffect interface{ reduceTriggerRate(*int32) }
type activeSkillSkipEffect interface{ skipActiveSkill(*Battle) bool }
type passiveSkillSkipEffect interface{ skipPassiveSkill() }

// poison
func (poison *StatusEffect_Poison_) applyBeforeActiveSkill(b *Battle, u *Unit) {
	rate := minPoisonDamageRate + b.getRandom(maxPoisonDamageRate-minPoisonDamageRate+1)
	damage := u.getParam(ParamType_BASE, Param_HP) * rate / 100
	b.getCurrentAction().Poison = u.takeDamage(b, damage)
}

// sleep
func (*StatusEffect_Sleep_) skipPassiveSkill()            {}
func (*StatusEffect_Sleep_) skipActiveSkill(*Battle) bool { return true }
func (sleep *StatusEffect_Sleep_) applyBeforeActiveSkill(b *Battle, u *Unit) {
	if sleep.Sleep.RemainingUses == 0 || b.getRandom(100) < sleepRecoveryRate {
		u.removeStatusEffect(StatusEffectType_SLEEP)
		return
	}

	sleep.Sleep.RemainingUses--
}

// confusion
func (*StatusEffect_Confusion_) skipPassiveSkill() {}
func (*StatusEffect_Confusion_) applyBeforeActiveSkill(b *Battle, u *Unit) {
	if b.getRandom(100) < confusionRecoveryRate+u.getStatusEffectModifier(StatusEffectType_CONFUSION).RecoveryBonus {
		u.removeStatusEffect(StatusEffectType_CONFUSION)
	}
}

func (u *Unit) applyConfusion(skill *Skill) *Skill {
	if confusion := u.getStatusEffect(StatusEffectType_CONFUSION).GetConfusion(); confusion == nil {
		return skill
	} else {
		return confusionSkill
	}
}

// fear
func (*StatusEffect_Fear_) skipPassiveSkill() {}
func (*StatusEffect_Fear_) applyBeforeActiveSkill(b *Battle, u *Unit) {
	if b.getRandom(100) < fearRecoveryRate {
		u.removeStatusEffect(StatusEffectType_FEAR)
	}
}

// barrier
func (barrier *StatusEffect_Barrier_) applyAfterAction(u *Unit) {
	if barrier.Barrier.Activated {
		u.removeStatusEffect(StatusEffectType_BARRIER)
	}
}

func (u *Unit) applyBarrier(isEnemySide bool) bool {
	if barrier := u.getStatusEffect(StatusEffectType_BARRIER).GetBarrier(); barrier == nil {
		return false
	} else if isEnemySide {
		barrier.Activated = true
		return true
	}

	return false
}

// bleed
func (bleed *StatusEffect_Bleed_) applyBeforeActiveSkill(b *Battle, u *Unit) {
	if bleed.Bleed.RemainingUses == 0 {
		u.removeStatusEffect(StatusEffectType_BLEED)
		return
	}

	b.getCurrentAction().Bleed = u.takeDamage(b, bleed.Bleed.Damage)
	bleed.Bleed.RemainingUses--
}

// stun
func (*StatusEffect_Stun_) skipPassiveSkill()            {}
func (*StatusEffect_Stun_) skipActiveSkill(*Battle) bool { return true }
func (*StatusEffect_Stun_) applyBeforeActiveSkill(b *Battle, u *Unit) {
	if b.getRandom(100) < stunRecoveryRate {
		u.removeStatusEffect(StatusEffectType_STUN)
	}
}

func (u *Unit) recoverStunOnDamaged() { u.removeStatusEffect(StatusEffectType_STUN) }

// prostration
func (u *Unit) applyProstration(b *Battle) bool {
	if prostration := u.getStatusEffect(StatusEffectType_PROSTRATION).GetProstration(); prostration == nil {
		return false
	} else if b.getRandom(100) < prostrationRecoveryRate {
		u.removeStatusEffect(StatusEffectType_PROSTRATION)
	}

	return true
}

// curse
func (*StatusEffect_Curse_) skipPassiveSkill() {}
func (curse *StatusEffect_Curse_) applyBeforeActiveSkill(b *Battle, u *Unit) {
	if curse.Curse.RemainingUses == 0 {
		u.removeStatusEffect(StatusEffectType_CURSE)
	}

	curse.Curse.RemainingUses--
}

func (curse *StatusEffect_Curse_) applyAfterActiveSkill(b *Battle, u *Unit) {
	b.getCurrentAction().Curse = u.takeDamage(b, curse.Curse.Damage)
	curse.Curse.Damage = 0
}

func (u *Unit) addCurseDamage(point int32) {
	if curse := u.getStatusEffect(StatusEffectType_CURSE).GetCurse(); curse == nil {
		return
	} else {
		curse.Damage += point * curseDamageRate / 100
	}
}

// charm
func (charm *StatusEffect_Charm_) applyBeforeActiveSkill(b *Battle, u *Unit) {
	if charm.Charm.RemainingUses == 0 {
		u.removeStatusEffect(StatusEffectType_CHARM)
	}

	charm.Charm.RemainingUses--
}

func (u *Unit) applyCharm(target *SkillTarget) *SkillTarget {
	if charm := u.getStatusEffect(StatusEffectType_CHARM).GetCharm(); charm == nil {
		return target
	} else {
		team := target.Team
		switch target.Team {
		case Team_ALLY:
			team = Team_ENEMY
		case Team_ENEMY:
			team = Team_ALLY
		}

		return &SkillTarget{Team: team, Condition: target.Condition}
	}
}

// bind
func (*StatusEffect_Bind_) skipActiveSkill(b *Battle) bool { return b.getRandom(100) < bindSkipRate }
func (*StatusEffect_Bind_) reduceTriggerRate(base *int32) {
	*base = *base * (100 - bindSkipRate) / 100
}

func (*StatusEffect_Bind_) applyBeforeActiveSkill(b *Battle, u *Unit) {
	if b.getRandom(100) < bindRecoveryRate[u.getMaxDuplicateFaction(b)] {
		u.removeStatusEffect(StatusEffectType_BIND)
	}
}

// gravity
func (gravity *StatusEffect_Gravity_) applyBeforeActiveSkill(b *Battle, u *Unit) {
	for _, param := range []Param{Param_PHY, Param_INT, Param_AGI} {
		point := (u.getParam(ParamType_INCREASE, param) + 1) / 2
		if resistance, ok := u.getPassiveEffect(PassiveEffectType_RESISTANCE_PARAMETER_EFFECT_RATE); ok {
			point = point * resistance / 100
		}

		u.parameterDown(param, point)
	}
}

func (u *Unit) applyGravity(b *Battle, point int32) bool {
	if gravity := u.getStatusEffect(StatusEffectType_GRAVITY).GetGravity(); gravity == nil {
		return false
	} else {
		delta := u.getParam(ParamType_CURRENT, Param_ALL_PARAMS) - u.getParam(ParamType_BASE, Param_ALL_PARAMS)
		if delta < 0 {
			u.removeStatusEffect(StatusEffectType_GRAVITY)
			return true
		}

		if b.getRandom(100)*delta*4 < point {
			u.removeStatusEffect(StatusEffectType_GRAVITY)
		}

		return true
	}
}

func (u *Unit) getStatusEffectModifier(effect StatusEffectType) *StatusEffectModifier {
	if modifier, ok := u.Base.StatusEffectModifier[int32(effect)]; ok {
		return modifier
	} else {
		u.Base.StatusEffectModifier[int32(effect)] = &StatusEffectModifier{}
		return u.Base.StatusEffectModifier[int32(effect)]
	}
}

func (u *Unit) isNormal() bool {
	return len(u.Current.StatusEffects) == 0
}

func (u *Unit) getStatusEffect(effect StatusEffectType) *StatusEffect {
	return u.Current.StatusEffects[int32(effect)]
}

func (u *Unit) removeStatusEffect(effect StatusEffectType) {
	delete(u.Current.StatusEffects, int32(effect))
}

func (u *Unit) removeStatusEffectsOnDeath() {
	for effect := range StatusEffects {
		if effect != StatusEffectType_CURSE {
			delete(u.Current.StatusEffects, int32(effect))
		}
	}
}

var (
	confusionSkill = &Skill{
		SkillId: confusionSkillId,
		Effects: []*Skill_Effect{{
			Target:      &SkillTarget{Team: Team_BOTH, Condition: &SkillTarget_Position{Position_RANDOM}},
			Param:       Skill_Effect_HP,
			IsDamage:    true,
			SuccessRate: 100,
			Value:       &Skill_Effect_Rate_{&Skill_Effect_Rate{Min: 100, Max: 100}},
			Reference:   newMyBasicStat(ParamType_CURRENT, Param_PHY),
		}},
	}
	skipAction = &Skill{
		SkillId: skipActionId,
	}
)