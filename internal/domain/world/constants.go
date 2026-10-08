package world

const (
	ScreenWidth  = 960
	ScreenHeight = 640

	WorldWidth  = 3200.0
	WorldHeight = 2400.0
	WallThick   = 30.0

	PlayerSize    = 24.0
	SpawnZoneSize = 320.0

	// Формула скорости: base + SP * coef.
	PlayerBaseSpeed = 2.0
	SpeedPerSP      = 0.05

	AimSpeedMultiplier  = 0.45
	DashSpeedMultiplier = 3.5

	DashDuration = 8
	DashCooldown = 40

	EnemyBaseSpeed  = 1.0
	EnemySpeedPerSP = 0.03

	NavCellSize     = 24.0
	AgentClearance  = 4.0 // запас вокруг тела при сглаживании пути
	AgentPredictPad = 2.0 // запас при решении «идти напрямую»

	// Поведение стрелка.
	ShooterFastApproach    = 3.5
	ShooterSlowRetreatDist = 80.0
	ShooterFastRetreatDist = 140.0
	ShooterSlowRetreatMul  = 0.5
	ShooterFastRetreatMul  = 1.4

	// Паника: доля от PreferredMin, ниже которой стрелок начинает отступать.
	ShooterPanicThreshold = 0.85
	// Доля, выше которой он бросает стрельбу и сосредотачивается на побеге.
	ShooterFullPanic = 0.95

	// Разведчик: порог скорости прицела, при котором он паникует.
	ScoutAimSpeedPanic = 25.0

	// Разведчик: порог "точного наведения" — прицел в этом радиусе
	// и движется медленно.
	ScoutPreciseRadius = 40.0
	ScoutPreciseSpeed  = 6.0

	// Медик.
	MedicHealTickRate      = 1 // лечим 1 HP раз в N тиков (было 3)
	MedicHealPerTick       = 3 // HP за одно срабатывание
	MedicHealRange         = 220.0
	MedicCriticalHPRatio   = 0.25
	MedicFollowSpeedFactor = 1.05
	MedicAllyRadius        = 900.0 // было 320 — теперь видит почти всю карту
	MedicPlayerFleeRange   = 140.0
)
