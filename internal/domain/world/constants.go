package world

const (
	ScreenWidth  = 960
	ScreenHeight = 640

	WorldWidth  = 3200.0
	WorldHeight = 2400.0
	WallThick   = 30.0

	PlayerSize    = 24.0
	SpawnZoneSize = 320.0

	// i-frames после попадания мечом.
	MeleeHitIFrames = 20 // 0.33 сек при 60 TPS

	// Формула скорости: base + SP * coef.
	PlayerBaseSpeed = 2.0
	SpeedPerSP      = 0.05

	AimSpeedMultiplier  = 0.45
	DashSpeedMultiplier = 3.5

	DashDuration = 8
	DashCooldown = 40

	// Рывок щита.
	DashChargeMinTicks = 60  // 1 сек — минимальный порог активации
	DashChargeMaxTicks = 300 // 5 сек — полная зарядка

	DashSpeedBase      = 20.0 // скорость рывка при минимальной зарядке
	DashSpeedBonus     = 10.0 // +8 при полной → 16 max
	DashDurBase        = 25   // тиков рывка минимум
	DashDurBonus       = 50   // +15 при полной → 40 max
	DashStunBaseTicks  = 30   // 0.5 сек минимум оглушения
	DashStunBonusTicks = 270  // +4.5 сек → 300 max (5 сек)
	DashCdBaseTicks    = 30   // 0.5 сек минимум кулдауна
	DashCdBonusTicks   = 270  // +4.5 сек → 300 max (5 сек)

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
