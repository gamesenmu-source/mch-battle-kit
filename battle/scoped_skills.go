package battle

import "sync"

// Battle is an upstream protobuf type. Keep request-only lookups outside the
// serialized message and always remove them, including on a recovered panic.
// The immutable original repository is never replaced or written to.
var battleSkillLookups sync.Map // *Battle -> scopedSkills
type SkillLookup func(uint32) (*Skill, error)
type scopedSkills struct {
	lookup   SkillLookup
	authored map[uint32]bool
}

func (b *Battle) getSkill(id uint32) (*Skill, error) {
	if lookup, ok := battleSkillLookups.Load(b); ok {
		return lookup.(scopedSkills).lookup(id)
	}
	return GetSkill(id)
}
func (b *Battle) authoredSkill(id uint32) bool {
	if scope, ok := battleSkillLookups.Load(b); ok {
		return scope.(scopedSkills).authored[id]
	}
	return false
}
func (b *Battle) BattleWithSkills(lookup SkillLookup, authored []uint32) (bool, error) {
	ids := map[uint32]bool{}
	for _, id := range authored {
		ids[id] = true
	}
	battleSkillLookups.Store(b, scopedSkills{lookup, ids})
	defer battleSkillLookups.Delete(b)
	return b.Battle()
}

// At the representation boundary, saturate authored arithmetic instead of
// wrapping a positive charge into a negative number. Original battles retain
// their exact arithmetic and random sequence.
const customNumberLimit int64 = 2000000000

func boundedCustomNumber(n int64) int32 {
	if n > customNumberLimit {
		n = customNumberLimit
	}
	if n < -customNumberLimit {
		n = -customNumberLimit
	}
	return int32(n)
}
