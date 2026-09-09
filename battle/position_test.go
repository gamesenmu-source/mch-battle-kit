package battle

import "testing"

func newTestUnit(position int32) *Unit {
	return &Unit{
		Base: &BaseUnit{Position: position},
		Current: &CurrentUnit{
			Position:    position,
			Hp:          100,
			BuffEffects: map[int32]*BuffEffect{},
		},
	}
}

func newRaidTestBattle() (*Battle, *Unit, *Unit) {
	boss := newTestUnit(3)
	attacker := newTestUnit(0)
	b := &Battle{Units: []*Unit{attacker, newTestUnit(1), newTestUnit(2), boss}}
	return b, attacker, boss
}

// レイド (ボス1体、position 4,5 が空) でも通常攻撃のターゲットにボスが選ばれること
// getTeamUnits はスロット構造 (nil入り) を返す前提で、FIRST/ALL がボスを返すのを確認する
func TestRaidBossTargeting(t *testing.T) {
	b, attacker, boss := newRaidTestBattle()

	for _, position := range []Position{Position_FIRST, Position_ALL} {
		units := b.getUnitsByTeamPosition(attacker, nil, Team_ENEMY, position)
		if len(units) != 1 || units[0] != boss {
			t.Fatalf("position %v should target [boss], got %v", position, units)
		}
	}
}

// レイドアタック時の applyDecoy が panic しないこと (unit.go getBuffEffect の nil deref 再発防止)
func TestApplyDecoyOnRaidBoss(t *testing.T) {
	b, attacker, boss := newRaidTestBattle()

	if got := boss.applyDecoy(b, attacker); got != boss {
		t.Fatalf("applyDecoy should return boss, got %v", got)
	}
}

// チームに空き枠がある場合の getMaxDuplicateFaction が panic しないこと
func TestGetMaxDuplicateFactionWithEmptySlot(t *testing.T) {
	b, _, boss := newRaidTestBattle()
	boss.Base.Faction = Faction(1)

	if got := boss.getMaxDuplicateFaction(b); got != 0 {
		t.Fatalf("boss has no allies, expected 0, got %v", got)
	}
}