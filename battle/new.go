package battle

import (
	"fmt"
)

func NewBattle(battleType uint32, randomSeed int64, uids []uint32, units ...*Unit) (*Battle, error) {
	if len(uids) < 2 {
		return nil, fmt.Errorf("new battle wrong uids=%v", uids)
	}
	newBattle := &Battle{
		RandomSeed:         randomSeed,
		Result:             Battle_PROGRESS,
		LastActivePosition: -1,
		State:              Battle_OPENING,
		Uids:               []uint32{uids[0], uids[1]},
		BattleType:         battleType,
		ActionLimit:        ThreeActionLimit,
	}
	newBattle.Units = make([]*Unit, len(units))
	for i, unit := range units {
		unit.Base.Index = int32(i)
		unit.Current.Index = int32(i)
		newBattle.Units[i] = unit
	}
	return newBattle, nil
}

func NewBattleWithRule(battleType uint32, randomSeed int64, uids []uint32, rules map[int32]bool, units ...*Unit) (*Battle, error) {
	if len(uids) < 2 {
		return nil, fmt.Errorf("new battle wrong uids=%v", uids)
	}
	newBattle := &Battle{
		RandomSeed:         randomSeed,
		Result:             Battle_PROGRESS,
		LastActivePosition: -1,
		State:              Battle_OPENING,
		Uids:               []uint32{uids[0], uids[1]},
		Rules:              rules,
		BattleType:         battleType,
		ActionLimit:        ThreeActionLimit,
	}
	newBattle.Units = make([]*Unit, len(units))
	for i, unit := range units {
		unit.Base.Index = int32(i)
		unit.Current.Index = int32(i)
		newBattle.Units[i] = unit
	}
	return newBattle, nil
}

func NewNineBattle(battleType uint32, randomSeed int64, uids []uint32, units []*Unit) (*Battle, error) {
	if len(uids) < 2 {
		return nil, fmt.Errorf("new battle wrong uids=%v", uids)
	}
	newBattle := &Battle{
		RandomSeed:         randomSeed,
		Result:             Battle_PROGRESS,
		LastActivePosition: -1,
		State:              Battle_OPENING,
		Uids:               uids,
		JinIds:             []uint32{0, 0}, // deprecated
		BattleType:         battleType,
		ActionLimit:        NineActionLimit,
	}
	newBattle.Units = make([]*Unit, len(units))
	for i, unit := range units {
		unit.Base.Index = int32(i)
		unit.Current.Index = int32(i)
		newBattle.Units[i] = unit
	}
	return newBattle, nil
}

func NewNineBattleWithRule(battleType uint32, randomSeed int64, uids []uint32, rules map[int32]bool, units []*Unit) (*Battle, error) {
	if len(uids) < 2 {
		return nil, fmt.Errorf("new battle wrong uids=%v", uids)
	}
	newBattle := &Battle{
		RandomSeed:         randomSeed,
		Result:             Battle_PROGRESS,
		LastActivePosition: -1,
		State:              Battle_OPENING,
		Rules:              rules,
		Uids:               uids,
		JinIds:             []uint32{0, 0}, // deprecated
		BattleType:         battleType,
		ActionLimit:        NineActionLimit,
	}
	newBattle.Units = make([]*Unit, len(units))
	for i, unit := range units {
		unit.Base.Index = int32(i)
		unit.Current.Index = int32(i)
		newBattle.Units[i] = unit
	}
	return newBattle, nil
}

func NewJinBattle(battleType uint32, randomSeed int64, uids []uint32, units []*Unit, jins []*Jin) (*Battle, error) {
	if len(uids) < 2 {
		return nil, fmt.Errorf("new battle wrong uids=%v", uids)
	}
	newBattle := &Battle{
		RandomSeed:         randomSeed,
		Result:             Battle_PROGRESS,
		LastActivePosition: -1,
		State:              Battle_OPENING,
		Uids:               uids,
		JinIds:             []uint32{0, 0}, // deprecated
		BattleType:         battleType,
		ActionLimit:        NineActionLimit,
	}
	newBattle.Units = make([]*Unit, len(units))
	for i, unit := range units {
		unit.Base.Index = int32(i)
		unit.Current.Index = int32(i)
		newBattle.Units[i] = unit
	}
	newBattle.Jins = make([]*Jin, len(jins))
	for i, jin := range jins {
		jin.Base.Index = int32(i)
		jin.Current.Index = int32(i)
		newBattle.Jins[i] = jin
	}
	newBattle.FieldState = make([]*FieldState, 2)
	newBattle.FieldState[0] = &FieldState{}
	newBattle.FieldState[1] = &FieldState{}
	return newBattle, nil
}

func NewJinThreeBattle(battleType uint32, randomSeed int64, uids []uint32, units []*Unit, jins []*Jin) (*Battle, error) {
	if len(uids) < 2 {
		return nil, fmt.Errorf("new battle wrong uids=%v", uids)
	}
	newBattle := &Battle{
		RandomSeed:         randomSeed,
		Result:             Battle_PROGRESS,
		LastActivePosition: -1,
		State:              Battle_OPENING,
		Uids:               uids,
		JinIds:             []uint32{0, 0}, // deprecated
		BattleType:         battleType,
		ActionLimit:        ThreeActionLimit,
	}

	newBattle.Units = make([]*Unit, len(units))
	for i, unit := range units {
		unit.Base.Index = int32(i)
		unit.Current.Index = int32(i)
		newBattle.Units[i] = unit
	}
	newBattle.Jins = make([]*Jin, len(jins))
	for i, jin := range jins {
		jin.Base.Index = int32(i)
		jin.Current.Index = int32(i)
		newBattle.Jins[i] = jin
	}
	newBattle.FieldState = make([]*FieldState, 2)
	newBattle.FieldState[0] = &FieldState{}
	newBattle.FieldState[1] = &FieldState{}
	return newBattle, nil
}

func NewSixBattle(battleType uint32, randomSeed int64, uids []uint32, jinIds []uint32, units []*Unit, bgmId uint32, backgroundId uint32) (*Battle, error) {
	if len(uids) < 2 {
		return nil, fmt.Errorf("new battle wrong uids=%v", uids)
	}

	newBattle := &Battle{
		RandomSeed:         randomSeed,
		Result:             Battle_PROGRESS,
		LastActivePosition: -1,
		State:              Battle_OPENING,
		Uids:               uids,
		JinIds:             jinIds,
		BattleType:         battleType,
		ActionLimit:        SixActionLimit,
		BgmId:              bgmId,
		BackgroundId:       backgroundId,
	}
	newBattle.Units = make([]*Unit, len(units))
	for i, unit := range units {
		unit.Base.Index = int32(i)
		unit.Current.Index = int32(i)
		newBattle.Units[i] = unit
	}
	return newBattle, nil
}
