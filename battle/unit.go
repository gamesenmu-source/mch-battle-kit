package battle

import (
	"context"

	"github.com/golang/protobuf/proto"
)

type paramGetter interface {
	GetHp() int32
	GetPhy() int32
	GetInt() int32
	GetAgi() int32
}

func (u *Unit) getParam(paramType ParamType, param Param) int32 {
	if param == Param_ALL_PARAMS {
		return u.getParam(paramType, Param_PHY) + u.getParam(paramType, Param_INT) + u.getParam(paramType, Param_AGI)
	}

	var val int32
	var getter paramGetter
	switch paramType {
	case ParamType_BASE:
		getter = u.Base
	case ParamType_CURRENT:
		getter = u.Current
	case ParamType_INCREASE:
		val = u.getParam(ParamType_CURRENT, param) - u.getParam(ParamType_BASE, param)
	case ParamType_DECREASE:
		val = u.getParam(ParamType_BASE, param) - u.getParam(ParamType_CURRENT, param)
	}

	if getter != nil {
		switch param {
		case Param_HP:
			val = getter.GetHp()
		case Param_PHY:
			val = getter.GetPhy()
		case Param_INT:
			val = getter.GetInt()
		case Param_AGI:
			val = getter.GetAgi()
		}
	}

	if val < 0 {
		val = 0
	}

	return val
}

func (u *Unit) getMaxParam(params ...Param) *Param {
	var p *Param
	for i := range params {
		if p == nil || u.getParam(ParamType_CURRENT, params[i]) > u.getParam(ParamType_CURRENT, *p) {
			p = &params[i]
		}
	}

	return p
}

func (u *Unit) getMinParam(params ...Param) *Param {
	var p *Param
	for i := range params {
		if p == nil || u.getParam(ParamType_CURRENT, params[i]) < u.getParam(ParamType_CURRENT, *p) {
			p = &params[i]
		}
	}

	return p
}

func (u *Unit) setParam(param Param, val int32) {
	switch param {
	case Param_PHY:
		u.Current.Phy = val
	case Param_INT:
		u.Current.Int = val
	case Param_AGI:
		u.Current.Agi = val
	}
	return
}

func (u *Unit) increaseCharge(point int32) {
	u.Current.Charge += point
	u.Current.BonusCharge += point
}

func (u *Unit) decreaseCharge(point int32) {
	if point < u.Current.Charge {
		u.Current.BonusCharge -= point
		u.Current.Charge -= point
	} else {
		u.Current.BonusCharge -= u.Current.Charge
		u.Current.Charge = 0
	}
}

func (u *Unit) death() {
	u.initialization()
	u.Current.Hp = 0
	u.Current.Charge = 0
	u.Current.BonusCharge = 0
	u.Current.Shield = 0
	u.removeStatusEffectsOnDeath()
	u.removeBuffEffectOnDeath()
	u.Current.States = map[int32]bool{int32(State_AFTER_DEATH): true} // 死亡後フラグ以外をリセット
}

func (u *Unit) initialization() {
	u.Current.Phy = u.Base.Phy
	u.Current.Int = u.Base.Int
	u.Current.Agi = u.Base.Agi
	u.Current.PhyDamageModifier = u.Base.PhyDamageModifier.Copy()
	u.Current.IntDamageModifier = u.Base.IntDamageModifier.Copy()
	u.Current.ActionAddedDamageRecord = &DamageRecord{}
	u.Current.ActionAddedHealingRecord = &DamageRecord{}
	u.Current.ActiveSkillTakenDamageRecord = &DamageRecord{}
}

func (u *Unit) isDeath() bool {
	return u.Current.Hp == 0
}

func (u *Unit) getBattleStat(stat BattleStat) int32 {
	switch stat {
	case BattleStat_ACTIVE_SKILL_TAKEN_DAMAGE:
		return u.Current.ActiveSkillTakenDamageRecord.Amount
	case BattleStat_ACTION_ADDED_DAMAGE:
		return u.Current.ActionAddedDamageRecord.Amount
	case BattleStat_ACTION_ADDED_HEALING:
		return u.Current.ActionAddedHealingRecord.Amount
	case BattleStat_CHARGE:
		return u.Current.Charge
	}

	return 0
}

func (u *Unit) getBuffEffect(buffType BuffEffectType) *BuffEffect {
	if u.Current.BuffEffects == nil {
		return nil
	}

	return u.Current.BuffEffects[int32(buffType)]
}

func (u *Unit) getPassiveEffect(effectType PassiveEffectType) (int32, bool) {
	if len(u.Base.PassiveEffects) == 0 {
		return 0, false
	}

	val, ok := u.Base.PassiveEffects[int32(effectType)]
	return val, ok
}

type PrevUnits []*Unit
type ContinueUnits []*Unit

// クエスト用
func (units PrevUnits) CreateNewTeamUnits(ctx context.Context, option *Option) ([]*Unit, error) {
	newUnits := []*Unit{}
	for _, unit := range units {
		newUnits = append(newUnits, unit.createNewNextBattleUnit())
	}

	return newUnits, nil
}

func (units ContinueUnits) CreateNewTeamUnits(ctx context.Context, option *Option) ([]*Unit, error) {
	newUnits := []*Unit{}
	for _, unit := range units {
		newUnits = append(newUnits, unit.createNewContinueBattleUnit())
	}

	return newUnits, nil
}

func (u *Unit) createNewNextBattleUnit() *Unit {
	unit := &Unit{
		Base: u.Base,
	}

	if unit.Base.IsEnemy {
		unit.Current = u.Current
		unit.Current.SkillUsages = copySkillUsage(u.Current.SkillUsages) // 引継ぐ nil対策
		unit.Current.BuffEffects = copyBuff(u.Current.BuffEffects)
		return unit
	}

	unit.Current = &CurrentUnit{
		Hp:          u.Current.Hp,
		SkillUsages: copySkillUsage(u.Current.SkillUsages), // 引継ぐ nil対策
		BuffEffects: copyBuff(u.Base.BuffEffects),
	}

	return unit
}

func (u *Unit) createNewContinueBattleUnit() *Unit {
	unit := u.createNewNextBattleUnit()
	unit.Current.Hp = unit.Base.Hp // HP回復
	unit.Current.SkillUsages = copySkillUsage(u.Base.SkillUsages)
	return unit
}

func (u *Unit) isAlive() bool {
	return u != nil && u.Current.Hp > 0
}

// ------------------------------------
type BattleDeck interface {
	CreateNewTeamUnits(context.Context, *Option) ([]*Unit, error)
}

type Option struct {
	Uid            uint32
	CryptidEnabled bool
	EmaEnabled     bool
	LockUntil      int64
}

func CreateNewBattleUnits(ctx context.Context, attackerDeck, defenderDeck BattleDeck, options ...*Option) ([]*Unit, error) {
	attackerOption := newOption()
	if len(options) > 0 {
		attackerOption = options[0]
	}

	defenderOption := newOption()
	if len(options) > 1 {
		defenderOption = options[1]
	}

	attacker, err := attackerDeck.CreateNewTeamUnits(ctx, attackerOption)
	if err != nil {
		return nil, err
	}

	for i, unit := range attacker {
		unit.Base.Position = NineSelfPositions[i]
		unit.Current.Position = NineSelfPositions[i]
	}

	defender, err := defenderDeck.CreateNewTeamUnits(ctx, defenderOption)
	if err != nil {
		return nil, err
	}

	for i, unit := range defender {
		unit.Base.Position = NineOpponentPositions[i]
		unit.Current.Position = NineOpponentPositions[i]
	}

	return append(attacker, defender...), nil
}

func (unit *Unit) SetValidatedParam() {
	if unit.Base.Hp < 1 {
		unit.Base.Hp = 1
	}

	for _, p := range []*int32{&unit.Base.Phy, &unit.Base.Int, &unit.Base.Agi} {
		if *p < 1 {
			*p = 1
		} else if *p > 999 {
			*p = 999
		}
	}

	unit.Current.Hp = unit.Base.Hp
	unit.Current.BuffEffects = copyBuff(unit.Base.BuffEffects)
	unit.Current.SkillUsages = copySkillUsage(unit.Base.SkillUsages)
}

func newOption() *Option {
	return &Option{}
}

func copyBuff(src map[int32]*BuffEffect) map[int32]*BuffEffect {
	dst := make(map[int32]*BuffEffect, len(src))
	for k, v := range src {
		dst[k] = proto.Clone(v).(*BuffEffect)
	}

	return dst
}

func copySkillUsage(src map[uint32]*SkillUsage) map[uint32]*SkillUsage {
	dst := make(map[uint32]*SkillUsage, len(src))
	for k, v := range src {
		c := *v
		c.Used = false
		dst[k] = &c
	}

	return dst
}
