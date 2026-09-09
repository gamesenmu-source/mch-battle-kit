package battle

var (
	enchants = map[uint32]func(*Unit, int32){
		101: func(u *Unit, rate int32) {
			u.Base.PassiveEffects[int32(PassiveEffectType_RESISTANCE_PARAMETER_EFFECT_RATE)] += rate
		},
		102: func(u *Unit, rate int32) {
			u.Base.PhyDamageModifier.ReductionRateCap += rate
			u.Base.PhyDamageModifier.ReductionRateBonus += rate
		},
		103: func(u *Unit, rate int32) {
			u.Base.IntDamageModifier.ReductionRateCap += rate
			u.Base.IntDamageModifier.ReductionRateBonus += rate
		},
		104: func(u *Unit, rate int32) {
			u.Base.PassiveEffects[int32(PassiveEffectType_TRIGGER_RATE_INCREASE_BY_HP)] += rate
		},
	}
)

func (u *Unit) ApplyEnchant(enchantType uint32, rate int32) {
	enchants[enchantType](u, rate)
}