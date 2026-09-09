package battle

func (b *Battle) charge() *Unit {

	var minChargeTime int32 = 1000
	var maxCharge int32
	var maxPosition int32 = battleUnitsLimit
	var nextActiveUnit *Unit

	for _, unit := range b.Units {
		if unit.Current.Position >= battleUnitsLimit || unit.isDeath() {
			continue
		}
		unitBaseCharge := baseCharge
		if bonus, ok := unit.getPassiveEffect(PassiveEffectType_BASE_CHARGE_BONUS); ok {
			unitBaseCharge = unitBaseCharge * (100 + bonus) / 100
		}

		var chargeDiff int32 = requiredCharge - unit.Current.Charge
		var chargeTime int32
		if chargeDiff > 0 {
			chargeTime = (chargeDiff-1)/(unitBaseCharge+unit.Current.Agi) + 1 //切り上げ
		}
		if chargeTime < minChargeTime {
			minChargeTime = chargeTime
		}
	}

	for _, unit := range b.Units {
		if unit.Current.Position >= battleUnitsLimit || unit.isDeath() {
			continue
		}

		unitBaseCharge := baseCharge
		if bonus, ok := unit.getPassiveEffect(PassiveEffectType_BASE_CHARGE_BONUS); ok {
			unitBaseCharge = unitBaseCharge * (100 + bonus) / 100
		}

		unit.Current.Charge += (unitBaseCharge + unit.Current.Agi) * minChargeTime
		if unit.Current.Charge > maxCharge || (unit.Current.Charge == maxCharge && unit.Current.Position < maxPosition) {
			maxCharge = unit.Current.Charge
			maxPosition = unit.Current.Position
			nextActiveUnit = unit
		}
	}

	nextActiveUnit.Current.Charge = 0
	b.LastActivePosition = nextActiveUnit.Current.Position

	return nextActiveUnit
}