package runner

import (
	"bytes"
	"testing"
)

func TestCustomBossIsUsedAndLegacyBattleUnchanged(t *testing.T) {
	in := fixture()
	old, err := Run(in)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i < 3; i++ {
		in.Enemies = append(in.Enemies, Fighter{Name: []string{"守護像・前衛", "守護像・中衛", "守護像・後衛"}[i], HP: 1000000, PHY: 50, INT: 50, AGI: 40, Actives: []uint32{NPCSkill, NPCSkill, NPCSkill}})
	}
	same, err := Run(in)
	if err != nil {
		t.Fatal(err)
	}
	a, _ := Encode(old)
	b, _ := Encode(same)
	if !bytes.Equal(a, b) {
		t.Fatal("explicit default changed legacy results")
	}
	in.Enemies[0].Name = "Custom boss"
	in.Enemies[0].PHY = 999
	in.Enemies[0].INT = 999
	in.Enemies[0].AGI = 999
	in.Enemies[0].Actives = []uint32{0, NPCSkill, 0}
	custom, err := Run(in)
	if err != nil {
		t.Fatal(err)
	}
	if custom.Names[3] != "Custom boss" || custom.BaseParams[3] != [3]int32{999, 999, 999} {
		t.Fatal("custom boss ignored")
	}
	again, err := Run(in)
	if err != nil {
		t.Fatal(err)
	}
	a, _ = Encode(custom)
	b, _ = Encode(again)
	if !bytes.Equal(a, b) {
		t.Fatal("custom boss not deterministic")
	}
}
func TestRejectInvalidCustomBoss(t *testing.T) {
	for _, f := range []Fighter{
		{HP: 1000000, PHY: 1000, INT: 50, AGI: 50, Actives: []uint32{NPCSkill, NPCSkill, NPCSkill}},
		{HP: 1000000, PHY: 50, INT: 50, AGI: 50, HeroType: 15001, Actives: []uint32{NPCSkill, NPCSkill, NPCSkill}},
		{HP: 1000000, PHY: 50, INT: 50, AGI: 50, Actives: []uint32{123456789, NPCSkill, NPCSkill}},
	} {
		in := fixture()
		in.Enemies = []Fighter{f, f, f}
		if _, err := Run(in); err == nil {
			t.Fatal("invalid boss accepted")
		}
	}
	in := fixture()
	in.Enemies = []Fighter{}
	if _, err := Run(in); err == nil {
		t.Fatal("empty enemy team accepted")
	}
}
