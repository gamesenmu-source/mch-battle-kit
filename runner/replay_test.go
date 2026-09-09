package runner

import (
	b "mch-light/engine/battle"
	"testing"
)

func TestReplayCopiesEveryStatusAndBuff(t *testing.T) {
	u := &b.CurrentUnit{Position: 2, Phy: 150, Int: 70, Agi: 80, Charge: 1320, BonusCharge: 320, Shield: 90, StatusEffects: map[int32]*b.StatusEffect{}, BuffEffects: map[int32]*b.BuffEffect{}, Actives: []uint32{11, 12, 13}, Passive: 14}
	for _, id := range b.AllStatusEffects {
		u.StatusEffects[int32(id)] = &b.StatusEffect{Effect: b.StatusEffects[id].Constructor(&b.Unit{Current: u})}
	}
	for id, master := range b.BuffEffects {
		u.BuffEffects[int32(id)] = master.Constructor()
	}
	u.StatusEffects[int32(b.StatusEffectType_SLEEP)].GetSleep().RemainingUses = 0
	u.BuffEffects[int32(b.BuffEffectType_AGI_DAMAGE_REDUCTION)].GetAgiDamageReduction().RedunctionRate = 25
	s := snapshot(u, &b.BaseUnit{})
	if len(s.Status) != 12 || len(s.Buffs) != 4 || s.Position != 2 || s.Shield != 90 || s.Charge != 1320 || s.Params[0] != 150 {
		t.Fatalf("incomplete state: %+v", s)
	}
	if s.Status[1].ID != 3 || s.Status[1].Remaining == nil || *s.Status[1].Remaining != 0 {
		t.Fatal("lost the zero recovery counter")
	}
	if s.Buffs[2].Rate == nil || *s.Buffs[2].Rate != 25 {
		t.Fatal("lost damage reduction rate")
	}
	u.StatusEffects[3].GetSleep().RemainingUses = 9
	u.Actives[0] = 99
	if *s.Status[1].Remaining != 0 || s.Skills[0] != 11 {
		t.Fatal("replay aliases mutable engine state")
	}
}

func TestBattleExportsAfterActionStates(t *testing.T) {
	r, err := Run(fixture())
	if err != nil {
		t.Fatal(err)
	}
	if r.Version != 2 || len(r.HeroTypes) != 6 || len(r.BaseParams) != 6 {
		t.Fatal("missing replay metadata")
	}
	charged := false
	targeted := false
	for _, frame := range r.Frames {
		if frame.Targets == nil {
			t.Fatal("target list must distinguish known-empty from old unrecorded logs")
		}
		for _, target := range frame.Targets {
			if target < 0 || target > 5 {
				t.Fatal("invalid target position")
			}
			targeted = true
		}
		if len(frame.Units) != 6 {
			t.Fatal("lost units")
		}
		for _, u := range frame.Units {
			if u.Charge > 0 {
				charged = true
			}
		}
	}
	if !charged {
		t.Fatal("charge not recorded")
	}
	if !targeted {
		t.Fatal("effect targets not recorded")
	}
	if r.Frames[0].Units[0].Charge != 0 {
		t.Fatal("initial state mutated")
	}
	data, _ := Encode(r)
	if len(data) > 500000 {
		t.Fatal("replay exceeds storage bound")
	}
}
