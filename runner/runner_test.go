package runner

import (
	"bytes"
	b "mch-light/engine/battle"
	"strings"
	"testing"
)

func fixture() Input {
	// Test-only skills. Never installed by the production server.
	b.RegisterRepository(&b.CatalogRepository{Skills: map[uint32]*b.Skill{0: {SkillId: 0}, NPCSkill: npcSkill()}})
	return Input{Seed: 12345, CatalogVersion: "test-only", Team: []Fighter{
		{Name: "A", HeroType: 5001, HP: 200, PHY: 100, INT: 100, AGI: 70, Actives: []uint32{NPCSkill, NPCSkill, NPCSkill}},
		{Name: "B", HeroType: 5002, HP: 200, PHY: 100, INT: 100, AGI: 60, Actives: []uint32{NPCSkill, NPCSkill, NPCSkill}},
		{Name: "C", HeroType: 5003, HP: 200, PHY: 100, INT: 100, AGI: 50, Actives: []uint32{NPCSkill, NPCSkill, NPCSkill}},
	}}
}
func TestDeterministicIsolatedBattles(t *testing.T) {
	in := fixture()
	r, err := Run(in)
	if err != nil {
		t.Fatal(err)
	}
	if r.Score <= 0 || len(r.Frames) < 2 {
		t.Fatal("no battle recorded")
	}
	gold, _ := Encode(r)
	for i := 0; i < 100; i++ {
		again, err := Run(in)
		if err != nil {
			t.Fatal(err)
		}
		got, _ := Encode(again)
		if !bytes.Equal(gold, got) {
			t.Fatal("seeded battle changed across requests")
		}
	}
}
func TestRejectReplicasAndUnknownSkills(t *testing.T) {
	in := fixture()
	in.Team[0].HeroType = 15001
	if _, err := Run(in); err == nil {
		t.Fatal("replica accepted")
	}
	in = fixture()
	in.Team[0].Actives[0] = 99999
	if _, err := Run(in); err == nil {
		t.Fatal("unknown skill accepted")
	}
}
func TestRejectDisplayCatalog(t *testing.T) {
	_, err := LoadSkills(strings.NewReader(`{"skill_id":1,"name":{"ja":"display-only"}}`))
	if err == nil {
		t.Fatal("display text accepted as mechanics")
	}
}
func BenchmarkBattle(bm *testing.B) {
	in := fixture()
	bm.ResetTimer()
	for i := 0; i < bm.N; i++ {
		if _, err := Run(in); err != nil {
			bm.Fatal(err)
		}
	}
}
