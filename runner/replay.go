package runner

import (
	b "mch-light/engine/battle"
	"slices"
	"sort"
)

// Compact copies of the engine's AFTER-action snapshots, never inferred from
// skill descriptions. Optional counters retain zero (the next recovery check).
type EffectState struct {
	ID        int32  `json:"id"`
	Remaining *int32 `json:"remaining,omitempty"`
	Rate      *int32 `json:"rate,omitempty"`
}
type UnitState struct {
	Position    int32         `json:"position"`
	Params      [3]int32      `json:"params"`
	Charge      int32         `json:"charge"`
	BonusCharge int32         `json:"bonusCharge,omitempty"`
	Shield      int32         `json:"shield,omitempty"`
	Status      []EffectState `json:"status,omitempty"`
	Buffs       []EffectState `json:"buffs,omitempty"`
	// cap, reduction bonus, critical chance, dealt bonus, received bonus.
	PHY      [5]int32 `json:"phy"`
	INT      [5]int32 `json:"int"`
	Skills   []uint32 `json:"skills,omitempty"`
	Critical bool     `json:"critical,omitempty"`
}

func modifier(m *b.DamageModifier) [5]int32 {
	return [5]int32{m.GetReductionRateCap(), m.GetReductionRateBonus(), m.GetCriticalRate(), m.GetDealDamageBonus(), m.GetTakeDamageBonus()}
}
func snapshot(u *b.CurrentUnit, base *b.BaseUnit) UnitState {
	s := UnitState{Position: u.Position, Params: [3]int32{u.Phy, u.Int, u.Agi}, Charge: u.Charge, BonusCharge: u.BonusCharge, Shield: u.Shield, PHY: modifier(u.PhyDamageModifier), INT: modifier(u.IntDamageModifier), Critical: u.CriticalJudg}
	for id, e := range u.StatusEffects {
		if e == nil {
			continue
		}
		v := EffectState{ID: id}
		switch id {
		case int32(b.StatusEffectType_SLEEP):
			v.Remaining = intPointer(e.GetSleep().GetRemainingUses())
		case int32(b.StatusEffectType_BLEED):
			v.Remaining = intPointer(e.GetBleed().GetRemainingUses())
		case int32(b.StatusEffectType_CURSE):
			v.Remaining = intPointer(e.GetCurse().GetRemainingUses())
		case int32(b.StatusEffectType_CHARM):
			v.Remaining = intPointer(e.GetCharm().GetRemainingUses())
		}
		s.Status = append(s.Status, v)
	}
	for id, e := range u.BuffEffects {
		if e == nil {
			continue
		}
		v := EffectState{ID: id}
		switch id {
		case int32(b.BuffEffectType_DECOY):
			v.Remaining = intPointer(e.GetDecoy().GetRemainingUses())
		case int32(b.BuffEffectType_AGI_DAMAGE_REDUCTION):
			v.Remaining = intPointer(e.GetAgiDamageReduction().GetRemainingActions())
			v.Rate = intPointer(e.GetAgiDamageReduction().GetRedunctionRate())
		case int32(b.BuffEffectType_HP_DAMAGE_REDUCTION):
			v.Remaining = intPointer(e.GetHpDamageReduction().GetRemainingActions())
			v.Rate = intPointer(e.GetHpDamageReduction().GetRedunctionRate())
		}
		s.Buffs = append(s.Buffs, v)
	}
	sort.Slice(s.Status, func(i, j int) bool { return s.Status[i].ID < s.Status[j].ID })
	sort.Slice(s.Buffs, func(i, j int) bool { return s.Buffs[i].ID < s.Buffs[j].ID })
	if !slices.Equal(u.Actives, base.Actives) || u.Passive != base.Passive {
		s.Skills = append(slices.Clone(u.Actives), u.Passive)
	}
	return s
}

func intPointer(value int32) *int32 { return &value }
