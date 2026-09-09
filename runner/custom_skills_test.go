package runner

import (
	"bytes"
	"fmt"
	"sync"
	"testing"
)

func testEffect(kind, target string, amount int32) CustomEffect {
	return CustomEffect{Kind: kind, Target: target, Chance: 100, Reference: "fixed", Source: "self", Basis: "CURRENT", Min: amount, Max: amount, Status: "SLEEP", Buff: "DECOY"}
}
func testCustom() CustomSkill {
	return CustomSkill{ID: CustomSkillStart, Name: "手作りスキル", Trigger: "active", Chance: 100, Uses: -1, Threshold: 50, EffectID: 1, Effects: []CustomEffect{testEffect("damage", "enemyAll", 7)}}
}
func customInput(base Input, s CustomSkill) Input {
	in := base
	in.Enemies = []Fighter{}
	for i := 0; i < 3; i++ {
		in.Enemies = append(in.Enemies, Fighter{Name: "boss", HP: 1000000, PHY: 100, INT: 100, AGI: 1, Actives: []uint32{NPCSkill, NPCSkill, NPCSkill}})
	}
	in.Enemies[0].AGI = 999
	in.Enemies[0].CustomSkills = []CustomSkill{s}
	if s.Trigger == "active" {
		in.Enemies[0].Actives[0] = s.ID
	} else {
		in.Enemies[0].Passive = s.ID
	}
	return in
}
func TestAuthoredEffectsRunInOrder(t *testing.T) {
	base := fixture()
	s := testCustom()
	s.Effects = []CustomEffect{testEffect("damage", "self", 100), testEffect("heal", "self", 50), testEffect("damage", "enemyAll", 7), testEffect("status", "enemyAll", 0)}
	result, err := Run(customInput(base, s))
	if err != nil {
		t.Fatal(err)
	}
	frame := result.Frames[1]
	if frame.Skill != s.ID || frame.HP[3] != 999950 {
		t.Fatalf("self damage/heal order ignored: %+v", frame)
	}
	for i := 0; i < 3; i++ {
		if frame.HP[i] != 193 {
			t.Fatalf("custom all-target damage ignored: %+v", frame.HP)
		}
		found := false
		for _, effect := range frame.Units[i].Status {
			if effect.ID == 3 {
				found = true
			}
		}
		if !found {
			t.Fatal("authored sleep missing")
		}
	}
}
func TestAuthoredOpeningPassive(t *testing.T) {
	base := fixture()
	s := testCustom()
	s.Trigger = "opening"
	s.Uses = 1
	s.Effects = []CustomEffect{testEffect("shield", "allyAll", 50)}
	r, err := Run(customInput(base, s))
	if err != nil {
		t.Fatal(err)
	}
	if r.Frames[1].Skill != s.ID {
		t.Fatal("opening passive not run")
	}
	for i := 3; i < 6; i++ {
		if r.Frames[1].Units[i].Shield != 50 {
			t.Fatal("shield missing")
		}
	}
	count := 0
	for _, f := range r.Frames {
		if f.Skill == s.ID {
			count++
		}
	}
	if count != 1 {
		t.Fatal("passive use limit ignored")
	}
}
func TestAuthoredSkillsIsolatedAcrossConcurrentBattles(t *testing.T) {
	base := fixture()
	before, err := Run(base)
	if err != nil {
		t.Fatal(err)
	}
	gold, _ := Encode(before)
	var wg sync.WaitGroup
	fail := make(chan error, 20)
	for i := 0; i < 20; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			s := testCustom()
			amount := int32(7 + n%2*4)
			s.Effects[0].Min = amount
			s.Effects[0].Max = amount
			r, err := Run(customInput(base, s))
			if err != nil {
				fail <- err
				return
			}
			if r.Frames[1].HP[0] != 200-amount {
				fail <- fmt.Errorf("custom skill leaked between battles")
			}
		}(i)
	}
	wg.Wait()
	close(fail)
	for err := range fail {
		t.Fatal(err)
	}
	after, err := Run(base)
	if err != nil {
		t.Fatal(err)
	}
	got, _ := Encode(after)
	if !bytes.Equal(gold, got) {
		t.Fatal("custom skills modified original catalog")
	}
}
func TestAuthoredValidationAndInclusiveRange(t *testing.T) {
	s := testCustom()
	s.Effects[0].Reference = "PHY"
	s.Effects[0].Min = 30
	s.Effects[0].Max = 35
	skill, err := CompileCustomSkill(s)
	if err != nil {
		t.Fatal(err)
	}
	if skill.Effects[0].GetRate().Max != 36 {
		t.Fatal("UI range must include its maximum")
	}
	for _, mutate := range []func(*CustomSkill){func(s *CustomSkill) { s.ID = NPCSkill }, func(s *CustomSkill) { s.Trigger = "eval" }, func(s *CustomSkill) { s.Effects[0].Target = "secret" }, func(s *CustomSkill) { s.Effects[0].Reference = "invalid" }, func(s *CustomSkill) { s.Effects[0].Max = 1001 }, func(s *CustomSkill) { s.Effects[0].Kind = "skillChange" }, func(s *CustomSkill) { s.Effects[0].Status = "fake" }, func(s *CustomSkill) { s.Effects = make([]CustomEffect, 7) }} {
		copy := s
		copy.Effects = append([]CustomEffect{}, s.Effects...)
		mutate(&copy)
		if _, err := CompileCustomSkill(copy); err == nil {
			t.Fatal("invalid authored skill accepted")
		}
	}
	base := fixture()
	in := customInput(base, s)
	in.Enemies[1].Actives[0] = s.ID
	if _, err := Run(in); err == nil {
		t.Fatal("another fighter's definition accepted")
	}
}

func TestAuthoredChargeMultiplicationDoesNotOverflow(t *testing.T) {
	base := fixture()
	s := testCustom()
	s.Effects = []CustomEffect{testEffect("chargeUp", "allyAll", 1000000)}
	for i := 0; i < 5; i++ {
		e := testEffect("chargeUp", "allyAll", 1000)
		e.Reference = "CHARGE"
		e.Source = "target"
		s.Effects = append(s.Effects, e)
	}
	in := customInput(base, s)
	in.Enemies[0].Actives = []uint32{s.ID, s.ID, s.ID}
	r, err := Run(in)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range r.Frames {
		for _, u := range f.Units {
			if u.Charge < 0 {
				t.Fatal("authored charge wrapped negative")
			}
		}
	}
	if len(r.Frames) > 1001 {
		t.Fatal("unbounded action count")
	}
}
