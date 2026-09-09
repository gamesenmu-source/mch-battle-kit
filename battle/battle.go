package battle

import (
	"context"
	"sort"

	"github.com/golang/protobuf/proto"
)

func (b *Battle) getPassiveOrderSortUnits() []*Unit {
	var units []*Unit
	for _, unit := range b.Units {
		if unit.Current.Position < battleUnitsLimit {
			units = append(units, unit)
		}
	}

	sort.Slice(units, func(i, j int) bool {
		if units[i].Base.Agi == units[j].Base.Agi {
			return b.getRandom(2) == 0
		}

		return units[i].Base.Agi > units[j].Base.Agi
	})

	return units
}

func (b *Battle) next() (bool, error) {
	err := b.passivePhase()
	if err != nil {
		return false, err
	}

	err = b.resetPhase()
	if err != nil {
		return false, err
	}

	if b.isSettled() {
		return false, nil
	}

	u := b.charge()
	err = b.activePhase(u)
	if err != nil {
		return false, err
	}

	if b.isSettled() {
		return false, nil
	}

	return true, nil
}

func (b *Battle) resetPhase() error {
	for _, unit := range b.Units {
		unit.Current.ActiveSkillTakenDamageRecord.Reset()
		unit.Current.States = map[int32]bool{}

		for _, usage := range unit.Current.SkillUsages {
			if usage.RemainingUses != 0 {
				usage.Used = false
			}
		}
	}

	return nil
}

func (b *Battle) activePhase(u *Unit) error {
	b.State = Battle_ACTIVE
	skill, err := b.getSkill(u.Current.Actives[u.Current.ActiveCounts%3])
	if err != nil {
		return err
	}

	u.useActiveSkill(b, skill)
	return nil
}

func (b *Battle) Battle() (bool, error) {
	b.initialization()
	next := true
	for next {
		var err error
		next, err = b.next()
		if err != nil {
			return false, err
		}
	}

	return b.Result == Battle_WIN, nil
}

func (b *Battle) BattleWithSave(ctx context.Context) (uint32, bool, error) {
	win, err := b.Battle()
	if err != nil {
		return 0, false, err
	}

	battleId, err := b.Save(ctx)
	if err != nil {
		return 0, false, err
	}

	return battleId, win, nil
}

type BattleInfo struct {
	BattleId            uint32
	Result              Battle_Result
	AttackerTakenDamage int32
	DefenderTakenDamage int32
}

func (b *Battle) BattleWithResult(ctx context.Context) (*BattleInfo, error) {
	battleId, _, err := b.BattleWithSave(ctx)
	if err != nil {
		return nil, err
	}

	battleInfo := &BattleInfo{
		BattleId:            battleId,
		Result:              b.Result,
		AttackerTakenDamage: b.AttackerTakenDamage,
		DefenderTakenDamage: b.DefenderTakenDamage,
	}

	return battleInfo, nil
}

func (b *Battle) BattleWithTargetResult(ctx context.Context, result Battle_Result) (*BattleInfo, error) {
	_, err := b.Battle()
	if err != nil {
		return nil, err
	}

	var battleId uint32
	if result == b.Result {
		battleId, err = b.Save(ctx)
		if err != nil {
			return nil, err
		}
	}

	battleInfo := &BattleInfo{
		BattleId:            battleId,
		Result:              b.Result,
		AttackerTakenDamage: b.AttackerTakenDamage,
		DefenderTakenDamage: b.DefenderTakenDamage,
	}

	return battleInfo, nil
}

func (b *Battle) BattleWithResultNoSave(ctx context.Context) (*BattleInfo, error) {
	_, err := b.Battle()
	if err != nil {
		return nil, err
	}

	battleInfo := &BattleInfo{
		Result:              b.Result,
		AttackerTakenDamage: b.AttackerTakenDamage,
		DefenderTakenDamage: b.DefenderTakenDamage,
	}

	return battleInfo, nil
}

func (b *Battle) Save(ctx context.Context) (uint32, error) {
	return r.SaveBattle(ctx, b)
}

func GetBattle(ctx context.Context, battleId uint32) (*Battle, error) {
	return r.GetBattle(ctx, battleId)
}

func GetTxBattle(ctx context.Context, battleId uint32) (*Battle, error) {
	return r.GetTxBattle(ctx, battleId)
}

func (b *Battle) passivePhase() error {
	b.State = Battle_FIRST_PASSIVE
	sortedUnits := b.getPassiveOrderSortUnits()
	for {
		var nextLoop bool
		for _, unit := range sortedUnits {
			if nextLoop = unit.useResurrection(b); nextLoop {
				if b.isSettled() {
					return nil
				}

				break
			}
		}

		if nextLoop {
			continue
		}

		for _, unit := range sortedUnits {
			skill, err := b.getSkill(unit.Current.Passive)
			if err != nil {
				return err
			}

			if nextLoop = unit.usePassiveSkill(b, skill); nextLoop {
				if b.isSettled() {
					return nil
				}

				break
			}

			if nextLoop {
				break
			}
		}

		if !nextLoop {
			return nil
		}
	}
}

func (b *Battle) initialization() {
	b.StoredVersion = 1 // 本当はnewにあった方がいい
	initialUnits := make([]*CurrentUnit, len(b.Units))
	for i, unit := range b.Units {
		unit.initialization()
		unit.Current.Actives = append(unit.Current.Actives, unit.Base.Actives...)
		unit.Current.Passive = unit.Base.Passive
		unit.Current.States = map[int32]bool{int32(State_OPENING): true}
		unit.Current.StatusEffects = map[int32]*StatusEffect{}
		if unit.Base.StatusEffectModifier == nil {
			unit.Base.StatusEffectModifier = map[int32]*StatusEffectModifier{}
		}

		initialUnits[i] = proto.Clone(unit.Current).(*CurrentUnit)

	}

	b.Actions = []*BattleAction{{Units: initialUnits}}
}

func (b *Battle) isActivePhase() bool {
	return b.State == Battle_ACTIVE
}

func (b *Battle) isSettled() bool {
	return b.Result != Battle_PROGRESS
}
