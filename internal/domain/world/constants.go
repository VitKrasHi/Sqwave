package world

const (
	ScreenWidth  = 960
	ScreenHeight = 640

	WorldWidth  = 2000.0
	WorldHeight = 1500.0
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

	EnemyBaseSpeed  = 1.5
	EnemySpeedPerSP = 0.02
)
