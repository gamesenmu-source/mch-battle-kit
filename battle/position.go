package battle

type Side int32

const Both, Attacker, Defender Side = 0, 1, 2

// battle units only
func (u *Unit) getOwnSide() Side {
	if u.Current.Position < 3 {
		return Attacker
	}

	return Defender
}

func (u *Unit) isAllySide(target *Unit) bool {
	return u.getOwnSide() == target.getOwnSide()
}

func (u *Unit) getSide(team Team) Side {
	if team == Team_BOTH {
		return Both
	}

	side := u.getOwnSide()
	if team == Team_ENEMY {
		if side == Attacker {
			return Defender
		}

		return Attacker
	}

	return side
}

func (u *Unit) getTeamUnits(b *Battle, team Team) []*Unit {
	start, end := 0, 6
	if side := u.getSide(team); side == Attacker {
		end = 3
	} else if side == Defender {
		start = 3
	}

	// ポジションは3枠単位のスロット構造 (FIRST/MIDDLE/BACK等のindex参照が依存)。
	// レイド等で存在しない枠はnilのまま返すので、呼び出し側はnilチェックが必要
	candidates := make([]*Unit, end-start)
	for i := start; i < end; i++ {
		for _, unit := range b.Units {
			if unit.Current.Position == int32(i) {
				candidates[i-start] = unit
				break
			}
		}
	}

	return candidates
}

// first Last random以外は死亡unitも返す
func (b *Battle) getUnitsByTeamPosition(actionUnit *Unit, targets []*Unit, team Team, position Position) []*Unit {
	if position == Position_SELF {
		return []*Unit{actionUnit}
	}

	candidates := actionUnit.getTeamUnits(b, team)
	units := []*Unit{}
	if position == Position_ALL {
		for _, candidate := range candidates {
			if candidate != nil {
				units = append(units, candidate)
			}
		}

		return units
	}

	if position == Position_RANDOM {
		for _, candidate := range candidates {
			if candidate.isAlive() {
				units = append(units, candidate)
			}
		}

		if len(units) == 0 {
			return []*Unit{}
		}

		return []*Unit{units[b.getRandom(int32(len(units)))]}
	}

	if position == Position_TARGET {
		for _, candidate := range candidates {
			if candidate == nil {
				continue
			}

			for _, target := range targets {
				if candidate.Current.Position == target.Current.Position {
					units = append(units, candidate)
					break
				}
			}
		}

		return units
	}

	if position == Position_EXCEPT_SELF {
		for _, candidate := range candidates {
			if candidate != nil && actionUnit.Current.Position != candidate.Current.Position {
				units = append(units, candidate)
			}
		}

		return units
	}

	quo := len(candidates) / 3
	for i := 0; i < quo; i++ {
		index := 3 * i
		if position == Position_FIRST || position == Position_LAST {
			for j := range 3 {
				var unit *Unit
				if position == Position_FIRST {
					unit = candidates[index+j]
				} else {
					unit = candidates[index+2-j]
				}

				if unit.isAlive() {
					units = append(units, unit)
					break
				}
			}
			continue
		}

		switch position {
		case Position_MIDDLE:
			index += 1
		case Position_BACK:
			index += 2
		}

		if unit := candidates[index]; unit != nil {
			units = append(units, unit)
		}
	}

	return units
}

func (b *Battle) getUnitsByParam(actionUnit *Unit, team Team, condition *SkillTarget_ParamCondition) []*Unit {
	candidates := actionUnit.getTeamUnits(b, team)
	var unit *Unit
	for _, candidate := range candidates {
		if candidate.isAlive() {
			if unit == nil {
				unit = candidate
				continue
			}

			p1, p2 := unit.getParam(condition.ParamType, condition.Param), candidate.getParam(condition.ParamType, condition.Param)
			if condition.Calc == SkillTarget_ParamCondition_HIGHEST && p1 < p2 {
				unit = candidate
				continue
			}

			if condition.Calc == SkillTarget_ParamCondition_LOWEST && p1 > p2 {
				unit = candidate
			}
		}
	}

	if unit == nil {
		return []*Unit{}
	}

	return []*Unit{unit}
}