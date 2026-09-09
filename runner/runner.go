package runner

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/golang/protobuf/jsonpb"
	"io"
	b "mch-light/engine/battle"
)

const NPCSkill uint32 = 4000000000

type Fighter struct {
	Name         string        `json:"name"`
	HeroType     uint32        `json:"heroType"`
	HP           int32         `json:"hp"`
	PHY          int32         `json:"phy"`
	INT          int32         `json:"int"`
	AGI          int32         `json:"agi"`
	Actives      []uint32      `json:"actives"`
	Passive      uint32        `json:"passive"`
	Attributes   []uint32      `json:"attributes"`
	Faction      int32         `json:"faction"`
	CustomSkills []CustomSkill `json:"customSkills,omitempty"`
	EnemyType    uint32        `json:"enemyType,omitempty"`
}
type Input struct {
	Seed           int64     `json:"seed"`
	Team           []Fighter `json:"team"`
	Enemies        []Fighter `json:"enemies,omitempty"`
	CatalogVersion string    `json:"catalogVersion"`
}
type Frame struct {
	Skill   uint32      `json:"skill"`
	Actor   int32       `json:"actor"`
	HP      []int32     `json:"hp"`
	Units   []UnitState `json:"units"`
	Targets []int32     `json:"targets"`
}
type Result struct {
	Score          int64      `json:"score"`
	Outcome        string     `json:"outcome"`
	CatalogVersion string     `json:"catalogVersion"`
	Names          []string   `json:"names"`
	MaxHP          []int32    `json:"maxHp"`
	Frames         []Frame    `json:"frames"`
	Version        int        `json:"version"`
	HeroTypes      []uint32   `json:"heroTypes"`
	BaseParams     [][3]int32 `json:"baseParams"`
}

// Load the exact protobuf JSON exported from skill_master2. Display descriptions
// are deliberately not accepted as skill definitions.
func LoadSkills(reader io.Reader) (map[uint32]*b.Skill, error) {
	scanner := bufio.NewScanner(reader)
	scanner.Buffer(make([]byte, 65536), 4*1024*1024)
	skills := map[uint32]*b.Skill{}
	for scanner.Scan() {
		line := bytes.TrimSpace(scanner.Bytes())
		if len(line) == 0 {
			continue
		}
		s := &b.Skill{}
		if err := jsonpb.Unmarshal(bytes.NewReader(line), s); err != nil {
			return nil, err
		}
		if _, ok := skills[s.SkillId]; ok {
			return nil, fmt.Errorf("duplicate skill %d", s.SkillId)
		}
		for _, e := range s.Effects {
			if e == nil || e.Target == nil {
				return nil, fmt.Errorf("skill %d missing target", s.SkillId)
			}
		}
		skills[s.SkillId] = s
	}
	if err := scanner.Err(); err != nil {
		return nil, err
	}
	if len(skills) < 2 {
		return nil, fmt.Errorf("skill export is empty")
	}
	if _, ok := skills[0]; !ok {
		return nil, fmt.Errorf("missing skill 0")
	}
	if _, ok := skills[NPCSkill]; ok {
		return nil, fmt.Errorf("reserved NPC skill collision")
	}
	skills[NPCSkill] = npcSkill()
	return skills, nil
}
func npcSkill() *b.Skill {
	return &b.Skill{SkillId: NPCSkill, RemainingUses: -1, Effects: []*b.Skill_Effect{{Target: &b.SkillTarget{Team: b.Team_ENEMY, Condition: &b.SkillTarget_Position{Position: b.Position_FRONT}}, Param: b.Skill_Effect_HP, IsDamage: true, SuccessRate: 100, Value: &b.Skill_Effect_Rate_{Rate: &b.Skill_Effect_Rate{Min: 20, Max: 20}}}}}
}
func unit(f Fighter, position int32, lookup b.SkillLookup) (*b.Unit, error) {
	if f.HP <= 0 || f.HP > 1000000 || f.PHY < 1 || f.INT < 1 || f.AGI < 1 || len(f.Actives) != 3 {
		return nil, fmt.Errorf("invalid fighter")
	}
	usages := map[uint32]*b.SkillUsage{}
	for _, id := range append(append([]uint32{}, f.Actives...), f.Passive) {
		u, err := b.NewSkillUsagesWithLookup(id, lookup)
		if err != nil {
			return nil, err
		}
		for k, v := range u {
			usages[k] = v
		}
	}
	u := &b.Unit{Base: &b.BaseUnit{Position: position, HeroType: f.HeroType, Hp: f.HP, Phy: f.PHY, Int: f.INT, Agi: f.AGI, Actives: f.Actives, Passive: f.Passive, AttributeTypes: f.Attributes, Faction: b.Faction(f.Faction), IsEnemy: position >= 3, PhyDamageModifier: &b.DamageModifier{ReductionRateCap: 40}, IntDamageModifier: &b.DamageModifier{ReductionRateCap: 40}, PassiveEffects: map[int32]int32{}, SkillUsages: usages}, Current: &b.CurrentUnit{Position: position}}
	u.SetValidatedParam()
	return u, nil
}
func Run(in Input) (*Result, error) {
	lookup, err := customLookup(in.Enemies)
	if err != nil {
		return nil, err
	}
	if len(in.Team) != 3 {
		return nil, fmt.Errorf("three heroes required")
	}
	units := []*b.Unit{}
	names := []string{}
	hp := []int32{}
	for i, f := range in.Team {
		if f.HeroType == 0 || (f.HeroType >= 10000 && (f.HeroType < 10001 || f.HeroType > 10003)) {
			return nil, fmt.Errorf("replicas are not supported")
		}
		if len(f.CustomSkills) > 0 {
			return nil, fmt.Errorf("custom skills are boss-only")
		}
		u, err := unit(f, int32(i), b.GetSkill)
		if err != nil {
			return nil, err
		}
		units = append(units, u)
		names = append(names, f.Name)
		hp = append(hp, u.Base.Hp)
	}
	enemies := in.Enemies
	if enemies == nil {
		for i := 0; i < 3; i++ {
			enemies = append(enemies, Fighter{Name: []string{"守護像・前衛", "守護像・中衛", "守護像・後衛"}[i], HP: 1000000, PHY: 50, INT: 50, AGI: 40, Actives: []uint32{NPCSkill, NPCSkill, NPCSkill}})
		}
	}
	if len(enemies) != 3 {
		return nil, fmt.Errorf("three enemies required")
	}
	for i, f := range enemies {
		if (f.HeroType >= 10000 && (f.HeroType < 10001 || f.HeroType > 10003)) || f.PHY > 999 || f.INT > 999 || f.AGI > 999 {
			return nil, fmt.Errorf("invalid enemy parameters")
		}
		u, err := unit(f, int32(i+3), lookup)
		if err != nil {
			return nil, err
		}
		units = append(units, u)
		names = append(names, f.Name)
		hp = append(hp, u.Base.Hp)
	}
	battle, err := b.NewBattle(uint32(b.BattleType_RAID_BATTLE), in.Seed, []uint32{1, 0}, units...)
	if err != nil {
		return nil, err
	}
	battle.ActionLimit = 200
	authored := []uint32{}
	for _, f := range in.Enemies {
		for _, s := range f.CustomSkills {
			authored = append(authored, s.ID)
		}
	}
	if _, err = battle.BattleWithSkills(lookup, authored); err != nil {
		return nil, err
	}
	result := &Result{Version: 2, Score: int64(battle.DefenderTakenDamage), Outcome: battle.Result.String(), CatalogVersion: in.CatalogVersion, Names: names, MaxHP: hp}
	for _, u := range units {
		result.HeroTypes = append(result.HeroTypes, u.Base.HeroType)
		result.BaseParams = append(result.BaseParams, [3]int32{u.Base.Phy, u.Base.Int, u.Base.Agi})
	}
	for _, action := range battle.Actions {
		f := Frame{Skill: action.Skill, Actor: action.ActionPosition, Targets: append([]int32{}, action.EffectPositions...)}
		for i, u := range action.Units {
			f.HP = append(f.HP, u.Hp)
			f.Units = append(f.Units, snapshot(u, units[i].Base))
		}
		result.Frames = append(result.Frames, f)
	}
	if result.Score < 0 {
		return nil, fmt.Errorf("score overflow")
	}
	return result, nil
}
func Encode(result *Result) ([]byte, error) { return json.Marshal(result) }
