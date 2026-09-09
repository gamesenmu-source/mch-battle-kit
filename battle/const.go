package battle

type StateEffectFunc func(b *Battle, actionUnit *Unit, target *Unit, rate int32) bool
type AdditionalCalcFunc func(u *Unit, rate int32, additionalRate int32, isDamage bool) int32
type EffectCalculator func(actionUnit *Unit, rate int32) int32
type EffectCalculatorForTarget func(target *Unit, rate int32) int32
type EffectCalculatorForPhyInt func(actionUnit *Unit, target *Unit, rate int32) int32
type EffectCalculatorForUnits func(actionUnit *Unit, targetUnits []*Unit, rate int32) int32
type skillConditionHandler func(b *Battle, u *Unit, skill *Skill) (bool, error)

const (
	baseCharge           = int32(100)
	requiredCharge       = 1000
	battleUnitsLimit     = 6
	maxUnits             = 7
	defaultMaxDamageCut  = int32(40)
	maxIncreaseDamageCut = int32(100)
	//EmptySkillId     = 1
	NoneSkillId      = 0
	confusionSkillId = 1
	// ChangeSkillId       = 2
	ChangeUnitsSkillId  = 3
	resurrectionSkillId = 4
	JinActivateSkillId  = 5
	JinSkillId          = 6
	// CurseSkillId               = 7
	BerserkerAuraType          = 1037
	ThreeActionLimit           = 200
	SixActionLimit             = 400
	NineActionLimit            = 600
	SectionTimeoutActionCounts = 200 // 同チーム組み合わせにおけるmax action count
	NoneAuraId                 = 0

	SelfJinPosition     = 0
	OpponentJinPosition = 1

	AttackerSidePosition      = 100
	DefenderSidePosition      = 101
	BothSidePosition          = 102
	DamageCutAgiAuraType      = uint32(1001)
	DamageCutHpAuraType       = uint32(1002)
	BattleOpeningCondition100 = 100
	BattleUniqueConditionOnce = 1

	criticalRateThreshold = 1
	baseDamageCutFactor   = 100
	// StanRecoveryRate100        = 50 誤字
	// CurseActionRate100         = 50 deprecated
)

var (
	ThreeSelfPositions     = []int32{0, 1, 2}
	ThreeOpponentPositions = []int32{3, 4, 5}
	SixSelfPositions       = []int32{0, 1, 2, 10, 11, 12}
	SixOpponentPositions   = []int32{3, 4, 5, 20, 21, 22}
	NineSelfPositions      = []int32{0, 1, 2, 10, 11, 12, 13, 14, 15}
	NineOpponentPositions  = []int32{3, 4, 5, 20, 21, 22, 23, 24, 25}
	//休むとコラボは除外している
	EditSkillIds = []uint32{
		3001, 3002, 3003, 3004, 3005, 3006, 3007, 3008,
		3101, 3102, 3103, 3104, 3105, 3106, 3107, 3108, 3109, 3110,
		3111, 3112, 3113, 3114, 3115, 3116, 3117, 3118, 3119, 3120,
		3121, 3122, 3123, 3124, 3125, 3126, 3127, 3128, 3129, 3130,
	}
)