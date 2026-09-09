package runner

import (
	"fmt"
	b "mch-light/engine/battle"
	"sort"
)

// ValidateRequired walks changed-skill references too. Cycles are legitimate
// (skills can switch back); a visited set keeps inspection bounded.
func ValidateRequired(skills map[uint32]*b.Skill, required []uint32) error {
	missing := map[uint32]bool{}
	visited := map[uint32]bool{}
	queue := append([]uint32{}, required...)
	for len(queue) > 0 {
		id := queue[0]
		queue = queue[1:]
		if visited[id] {
			continue
		}
		visited[id] = true
		skill, ok := skills[id]
		if !ok || skill == nil {
			missing[id] = true
			continue
		}
		if skill.SkillId != id {
			return fmt.Errorf("skill ID mismatch: %d", id)
		}
		for _, effect := range skill.Effects {
			if effect == nil || effect.Target == nil {
				return fmt.Errorf("skill %d has an invalid effect", id)
			}
			if next, ok := effect.Value.(*b.Skill_Effect_SkillId); ok {
				queue = append(queue, next.SkillId)
			}
		}
	}
	if len(missing) > 0 {
		ids := make([]uint32, 0, len(missing))
		for id := range missing {
			ids = append(ids, id)
		}
		sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
		return fmt.Errorf("missing %d required skills: %v", len(ids), ids)
	}
	return nil
}
