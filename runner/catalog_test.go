package runner

import (
	b "mch-light/engine/battle"
	"strings"
	"testing"
)

func TestRequiredChangedSkill(t *testing.T) {
	skills := map[uint32]*b.Skill{1: {SkillId: 1, Effects: []*b.Skill_Effect{{Target: &b.SkillTarget{}, Value: &b.Skill_Effect_SkillId{SkillId: 2}}}}}
	if err := ValidateRequired(skills, []uint32{1}); err == nil || !strings.Contains(err.Error(), "[2]") {
		t.Fatalf("missing dependency not detected: %v", err)
	}
	skills[2] = &b.Skill{SkillId: 2, Effects: []*b.Skill_Effect{{Target: &b.SkillTarget{}, Value: &b.Skill_Effect_SkillId{SkillId: 1}}}}
	if err := ValidateRequired(skills, []uint32{1}); err != nil {
		t.Fatal(err)
	}
	if err := ValidateRequired(skills, []uint32{3}); err == nil {
		t.Fatal("missing root accepted")
	}
}

func TestMalformedImportDoesNotPanic(t *testing.T) {
	for _, raw := range []string{"{}\n{\"skillId\":1,\"effects\":[null]}", "{}\n{}"} {
		if _, err := LoadSkills(strings.NewReader(raw)); err == nil {
			t.Fatal("malformed export accepted")
		}
	}
}
