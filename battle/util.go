package battle

import (
	"crypto/sha256"
	"encoding/binary"
)

func (b *Battle) getRandom(max int32) int32 {
	num := getSha256Int32(b.RandomSeed + b.RandomCounts)
	b.RandomCounts++
	result := int64(num) % int64(max)
	return int32(result)
}

func getSha256Int32(i int64) uint32 {
	buf := make([]byte, binary.MaxVarintLen64)
	binary.PutVarint(buf, i)
	hash := sha256.Sum256(buf)
	bytes := make([]byte, len(hash))
	for i, b := range hash {
		bytes[i] = b
	}
	result := binary.BigEndian.Uint32(bytes)
	return result
}

func (b *Battle) getUnitByPosition(positions ...int32) []*Unit {
	var units []*Unit
	for _, position := range positions {
		for _, unit := range b.Units {
			if unit.Current.Position == position {
				units = append(units, unit)
				break
			}
		}
	}

	return units
}

func (b *Battle) checkEnd() {
	if b.IsOver() {
		if b.isRaidBattle() {
			b.Result = Battle_TIME_UP
		} else {
			b.battleResultDraw()
		}
	}

	selfTeamDead, opponentTeamDead := true, true
	selfUnits := b.getUnitByPosition(0, 1, 2)
	for _, unit := range selfUnits {
		usage, ok := unit.Current.SkillUsages[unit.Current.Passive]
		if unit.isAlive() || unit.getBuffEffect(BuffEffectType_RESURRECTION) != nil ||
			(ok && usage.DeathTrigger && unit.hasState(State_AFTER_DEATH) && usage.RemainingUses != 0) {
			selfTeamDead = false
			break
		}
	}

	opponentUnits := b.getUnitByPosition(3, 4, 5)
	for _, unit := range opponentUnits {
		usage, ok := unit.Current.SkillUsages[unit.Current.Passive]
		if unit.isAlive() || unit.getBuffEffect(BuffEffectType_RESURRECTION) != nil ||
			(ok && usage.DeathTrigger && unit.hasState(State_AFTER_DEATH) && usage.RemainingUses != 0) {
			opponentTeamDead = false
			break
		}
	}

	if selfTeamDead && opponentTeamDead {
		if b.isRaidBattle() {
			b.Result = Battle_LOSE
		} else {
			b.battleResultDraw()
		}
	} else if selfTeamDead {
		b.Result = Battle_LOSE
	} else if opponentTeamDead {
		b.Result = Battle_WIN
	}
}

func (b *Battle) battleResultDraw() {
	if b.AttackerTakenDamage < b.DefenderTakenDamage {
		b.Result = Battle_WIN
	} else if b.AttackerTakenDamage > b.DefenderTakenDamage {
		b.Result = Battle_LOSE
	} else {
		// 付与ダメージが同値だった場合、乱数で勝利
		if b.getRandom(100) >= 50 {
			b.Result = Battle_WIN
		} else {
			b.Result = Battle_LOSE
		}
	}
}

func (b *Battle) getCurrentAction() *BattleAction {
	return b.Actions[len(b.Actions)-1]
}

func (b *Battle) isRaidBattle() bool {
	return b.BattleType == uint32(BattleType_RAID_BATTLE)
}

func (b *Battle) IsOver() bool {
	return b.ActionCounts > b.ActionLimit
}

func (u *Unit) getMaxDuplicateFaction(b *Battle) int32 {
	allies := u.getTeamUnits(b, Team_ALLY)
	var result int32
	if u.Base.Faction == Faction_NONE_FACTION {
		return result
	}

	for _, ally := range allies {
		if ally == nil || u.Current.Position == ally.Current.Position || !ally.isAlive() {
			continue // 存在しない枠と自分自身と死んでる仲間はスキップ
		} else if u.Base.Faction == ally.Base.Faction {
			result++
		}
	}

	return result
}

func (u *Unit) getEnemyMaxDuplicateSeries(b *Battle) int32 {
	enemies := u.getTeamUnits(b, Team_ENEMY)
	seriesTypes := map[uint32]int32{}
	for _, enemy := range enemies {
		if enemy == nil {
			continue
		}

		for _, series := range enemy.Base.SeriesTypes {
			seriesTypes[series]++
		}
	}

	var result int32
	for _, count := range seriesTypes {
		if result < count {
			result = count
		}
	}

	return result
}

func (u *Unit) getDefferentSeries(b *Battle) int32 {
	allies := u.getTeamUnits(b, Team_ALLY)
	seriesTypes := map[uint32]int32{}
	for _, ally := range allies {
		if ally == nil {
			continue
		}

		for _, series := range ally.Base.SeriesTypes {
			seriesTypes[series]++
		}
	}

	return int32(len(seriesTypes))
}

func addPosition(positions []int32, target int32) []int32 {
	for i, position := range positions {
		if target == position {
			return positions
		}

		if target < position {
			positions = append(positions, 0)
			copy(positions[i+1:], positions[i:])
			positions[i] = target
			return positions
		}
	}

	return append(positions, target)
}