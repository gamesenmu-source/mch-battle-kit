package runner

import (
	"fmt"
	b "mch-light/engine/battle"
	"strings"
	"unicode"
)

const CustomSkillStart uint32 = 4100000000

type CustomEffect struct {
	Kind      string `json:"kind"`
	Target    string `json:"target"`
	Chance    int32  `json:"chance"`
	Reference string `json:"reference"`
	Source    string `json:"source"`
	Basis     string `json:"basis"`
	Min       int32  `json:"min"`
	Max       int32  `json:"max"`
	Status    string `json:"status"`
	Buff      string `json:"buff"`
}
type CustomSkill struct {
	ID        uint32         `json:"id"`
	Name      string         `json:"name"`
	Trigger   string         `json:"trigger"`
	Chance    int32          `json:"chance"`
	Uses      int32          `json:"uses"`
	Threshold int32          `json:"threshold"`
	EffectID  int32          `json:"effectId"`
	Effects   []CustomEffect `json:"effects"`
}

var effectParams = map[string]b.Skill_Effect_Param{
	"damage": b.Skill_Effect_HP, "heal": b.Skill_Effect_HP, "revive": b.Skill_Effect_REVIVE,
	"phyUp": b.Skill_Effect_PHY, "phyDown": b.Skill_Effect_PHY, "intUp": b.Skill_Effect_INT, "intDown": b.Skill_Effect_INT,
	"agiUp": b.Skill_Effect_AGI, "agiDown": b.Skill_Effect_AGI, "chargeUp": b.Skill_Effect_CHARGE, "chargeDown": b.Skill_Effect_CHARGE,
	"shield": b.Skill_Effect_SHIELD, "status": b.Skill_Effect_STATUS_EFFECT, "cure": b.Skill_Effect_STATUS_EFFECT, "buff": b.Skill_Effect_BUFF_EFFECT,
}

func customTarget(key string) (*b.SkillTarget, error) {
	t := &b.SkillTarget{Team: b.Team_ALLY}
	var position b.Position
	switch key {
	case "self":
		position = b.Position_SELF
	case "previous":
		t.Team = b.Team_BOTH
		position = b.Position_TARGET
	default:
		suffix := ""
		if strings.HasPrefix(key, "enemy") {
			t.Team = b.Team_ENEMY
			suffix = strings.TrimPrefix(key, "enemy")
		} else if strings.HasPrefix(key, "ally") {
			suffix = strings.TrimPrefix(key, "ally")
		} else {
			return nil, fmt.Errorf("invalid target")
		}
		positions := map[string]b.Position{"Front": b.Position_FIRST, "Back": b.Position_LAST, "All": b.Position_ALL, "Random": b.Position_RANDOM, "Other": b.Position_EXCEPT_SELF, "SlotFront": b.Position_FRONT, "SlotMiddle": b.Position_MIDDLE, "SlotBack": b.Position_BACK}
		if p, ok := positions[suffix]; ok && !(suffix == "Other" && t.Team == b.Team_ENEMY) {
			position = p
		} else {
			params := map[string]b.Param{"LowestHp": b.Param_HP, "HighestPhy": b.Param_PHY, "HighestInt": b.Param_INT, "HighestAgi": b.Param_AGI}
			p, ok := params[suffix]
			if !ok {
				return nil, fmt.Errorf("invalid target")
			}
			calc := b.SkillTarget_ParamCondition_HIGHEST
			if suffix == "LowestHp" {
				calc = b.SkillTarget_ParamCondition_LOWEST
			}
			t.Condition = &b.SkillTarget_ParamCondition_{ParamCondition: &b.SkillTarget_ParamCondition{ParamType: b.ParamType_CURRENT, Param: p, Calc: calc}}
			return t, nil
		}
	}
	t.Condition = &b.SkillTarget_Position{Position: position}
	return t, nil
}
func CompileCustomSkill(s CustomSkill) (*b.Skill, error) {
	bad := func() (*b.Skill, error) { return nil, fmt.Errorf("invalid custom skill %d", s.ID) }
	if s.ID < CustomSkillStart || s.ID > CustomSkillStart+11 || len([]rune(strings.TrimSpace(s.Name))) < 1 || len([]rune(s.Name)) > 32 || strings.IndexFunc(s.Name, unicode.IsControl) >= 0 || len(s.Effects) < 1 || len(s.Effects) > 6 || s.Chance < 1 || s.Chance > 100 || s.Threshold < 1 || s.Threshold > 100 || s.EffectID < 1 || s.EffectID > 5 || !(s.Uses == -1 || s.Uses >= 1 && s.Uses <= 20) {
		return bad()
	}
	skill := &b.Skill{SkillId: s.ID, RemainingUses: s.Uses}
	if s.Trigger == "active" {
		if s.Uses != -1 {
			return bad()
		}
	} else {
		condition := &b.Condition{Team: b.Team_ALLY, Position: b.Position_SELF}
		if s.Trigger == "lowHp" {
			condition.Condition = &b.Condition_ParamCondition_{ParamCondition: &b.Condition_ParamCondition{Param: b.Param_HP, Calc: b.Condition_ParamCondition_UNDER, Rate: s.Threshold}}
		} else {
			states := map[string]b.State{"opening": b.State_OPENING, "damaged": b.State_AFTER_ACTIVE_SKILL_TAKEN_DAMAGE, "acted": b.State_AFTER_ACTIVE_SKILL, "abnormal": b.State_HAS_STATUS_EFFECT}
			state, ok := states[s.Trigger]
			if !ok {
				return bad()
			}
			condition.Condition = &b.Condition_StateCondition{StateCondition: state}
		}
		skill.Trigger = &b.Trigger{TriggerRate: s.Chance, Conditions: []*b.Condition{condition}}
	}
	for _, c := range s.Effects {
		param, ok := effectParams[c.Kind]
		if !ok {
			return bad()
		}
		target, err := customTarget(c.Target)
		if err != nil {
			return bad()
		}
		basis, ok := b.ParamType_value[c.Basis]
		if !ok || basis == 0 {
			return bad()
		}
		status, ok := b.StatusEffectType_value[c.Status]
		if !ok || status == 0 || c.Kind == "status" && c.Status == "ANY_EFFECT" {
			return bad()
		}
		buff, ok := b.BuffEffectType_value[c.Buff]
		if !ok || buff == 0 {
			return bad()
		}
		if c.Chance < 1 || c.Chance > 100 || c.Source != "self" && c.Source != "target" || c.Min < 0 || c.Max < c.Min || c.Max > 1000000 || c.Reference != "fixed" && c.Max > 1000 {
			return bad()
		}
		if c.Kind == "revive" && c.Target != "allyAll" && c.Target != "allySlotFront" && c.Target != "allySlotMiddle" && c.Target != "allySlotBack" {
			return bad()
		}
		e := &b.Skill_Effect{Param: param, Target: target, SuccessRate: c.Chance, IsDamage: c.Kind == "damage" || c.Kind == "status" || strings.HasSuffix(c.Kind, "Down")}
		// The sheet converter uses an exclusive maximum; this editor's displayed
		// range includes both endpoints, so convert the upper endpoint here.
		max := c.Max
		if c.Min < c.Max {
			max++
		}
		e.Value = &b.Skill_Effect_Rate_{Rate: &b.Skill_Effect_Rate{Min: c.Min, Max: max}}
		raw := &b.RawStat{Team: b.Team_ALLY, Position: b.Position_SELF}
		if c.Source == "target" {
			raw.Team = b.Team_BOTH
			raw.Position = b.Position_TARGET
		}
		switch c.Reference {
		case "fixed":
		case "HP", "PHY", "INT", "AGI":
			raw.Stat = &b.RawStat_BasicStat{BasicStat: &b.BasicStat{ParamType: b.ParamType(basis), Param: b.Param(b.Param_value[c.Reference])}}
			e.Reference = &b.Skill_Effect_RawStat{RawStat: raw}
		case "CHARGE", "DAMAGE", "HEALING", "TAKEN_DAMAGE":
			stats := map[string]b.BattleStat{"CHARGE": b.BattleStat_CHARGE, "DAMAGE": b.BattleStat_ACTION_ADDED_DAMAGE, "HEALING": b.BattleStat_ACTION_ADDED_HEALING, "TAKEN_DAMAGE": b.BattleStat_ACTIVE_SKILL_TAKEN_DAMAGE}
			raw.Stat = &b.RawStat_BattleStat{BattleStat: stats[c.Reference]}
			e.Reference = &b.Skill_Effect_RawStat{RawStat: raw}
		case "HEAL_INT":
			e.Reference = &b.Skill_Effect_ComputedStat{ComputedStat: b.ComputedStat_INT_HEALING_MODIFIER}
		case "HEAL_PHY":
			e.Reference = &b.Skill_Effect_ComputedStat{ComputedStat: b.ComputedStat_PHY_HEALING_MODIFIER}
		default:
			return bad()
		}
		if c.Kind == "status" || c.Kind == "cure" {
			e.Reference = nil
			e.Value = &b.Skill_Effect_StatusEffectType{StatusEffectType: b.StatusEffectType(status)}
		}
		if c.Kind == "buff" {
			e.Reference = nil
			e.Value = &b.Skill_Effect_BuffEffectType{BuffEffectType: b.BuffEffectType(buff)}
		}
		skill.Effects = append(skill.Effects, e)
	}
	return skill, nil
}
func customLookup(enemies []Fighter) (b.SkillLookup, error) {
	skills := map[uint32]*b.Skill{}
	for index, f := range enemies {
		if len(f.CustomSkills) > 4 {
			return nil, fmt.Errorf("too many custom skills")
		}
		owned := map[uint32]*b.Skill{}
		for _, c := range f.CustomSkills {
			if c.ID < CustomSkillStart+uint32(index*4) || c.ID > CustomSkillStart+uint32(index*4+3) {
				return nil, fmt.Errorf("custom skill belongs to another fighter")
			}
			s, err := CompileCustomSkill(c)
			if err != nil {
				return nil, err
			}
			if skills[c.ID] != nil {
				return nil, fmt.Errorf("duplicate custom skill")
			}
			skills[c.ID] = s
			owned[c.ID] = s
		}
		for _, id := range f.Actives {
			if id >= CustomSkillStart {
				s := owned[id]
				if s == nil || s.Trigger != nil {
					return nil, fmt.Errorf("invalid custom active")
				}
			}
		}
		if f.Passive >= CustomSkillStart {
			s := owned[f.Passive]
			if s == nil || s.Trigger == nil {
				return nil, fmt.Errorf("invalid custom passive")
			}
		}
	}
	return func(id uint32) (*b.Skill, error) {
		if s := skills[id]; s != nil {
			return s, nil
		}
		return b.GetSkill(id)
	}, nil
}
