package world

import "Sqwave/internal/domain/geometry"

type EnemyType int

const (
	EnemyInfantry EnemyType = iota
	EnemyShooter
	EnemyScout
)

type EnemyStats struct {
	Name   string
	MaxHP  int
	SP     int
	Damage int
	Size   float64

	MeleeRange        float64
	AttackCooldownMin int
	AttackCooldownMax int
	SwingDuration     int

	PreferredMin     float64
	PreferredMax     float64
	ProjectileSpeed  float64
	ProjectileSize   float64
	ProjectileLife   int
	ProjectileCount  int     // пеллет за выстрел (дробовик разведчика = 8)
	ProjectileSpread float64 // полный угол разброса в градусах
}

var enemyStats = map[EnemyType]EnemyStats{
	EnemyInfantry: {
		Name:              "Infantry",
		MaxHP:             120,
		SP:                100,
		Damage:            35,
		Size:              24,
		MeleeRange:        55,
		AttackCooldownMin: 24, // 400 мс
		AttackCooldownMax: 60, // 1 сек
		SwingDuration:     5,
	},
	EnemyShooter: {
		Name:              "Shooter",
		MaxHP:             100,
		SP:                80,
		Damage:            25,
		Size:              22,
		MeleeRange:        0,
		AttackCooldownMin: 90,
		AttackCooldownMax: 150,
		PreferredMin:      180,
		PreferredMax:      340,
		ProjectileSpeed:   8, // было 4 — почти вдвое быстрее
		ProjectileSize:    8,
		ProjectileLife:    90, // было 150 — короче, чтобы не летели через пол-карты
	},
	EnemyScout: {
		Name:              "Scout",
		MaxHP:             70,
		SP:                200,
		Damage:            10, // на пеллету, 8 пеллет — до 80 в упор
		Size:              20,
		MeleeRange:        60, // дистанция выстрела дробью
		AttackCooldownMin: 50,
		AttackCooldownMax: 90,
		ProjectileSpeed:   9,
		ProjectileSize:    4,
		ProjectileLife:    20, // короткая жизнь — дробь летит недалеко
		ProjectileCount:   8,
		ProjectileSpread:  30,
	},
}

func (t EnemyType) Stats() EnemyStats {
	if s, ok := enemyStats[t]; ok {
		return s
	}
	return EnemyStats{Name: "Unknown"}
}

type Enemy struct {
	Type EnemyType

	X, Y float64
	HP   int

	FacingX, FacingY float64

	AttackTimer int
	SwingActive bool
	SwingTimer  int

	KnockbackVX    float64
	KnockbackVY    float64
	KnockbackTimer int

	LastSwingHitID int

	// Путь — срез waypoint'ов (мировых координат) для обхода стен.
	// Пока игрок виден напрямую, Path = nil, враг идёт по прямой.
	Path         []geometry.Point
	PathIndex    int
	PathCooldown int
	PathGoalCell [2]int

	// Детекция застревания.
	LastX, LastY float64
	StuckTicks   int

	ForcePathTimer int // если > 0 — игнорировать direct и идти только по A*

	// PrevPlayerDist — расстояние до игрока на прошлом тике.
	// Нужно стрелку, чтобы понимать, с какой скоростью игрок приближается.
	PrevPlayerDist float64
	// Кеш найденной позиции для стрельбы из-за укрытия.
	ShootPosX, ShootPosY float64
	ShootPosTimer        int
	// Трекинг попаданий: если HP упало — недавно ранен.
	LastHP int

	RecentlyHitTimer int

	// Разведчик: таймер для паттернов уклонения.
	AITimer int

	// Множитель скорости (используется разведчиком для рывков).
	SpeedBurst int

	// Разведчик: серия рывков при резком/точном наведении.
	DodgeTimer    int
	DodgeDirX     float64
	DodgeDirY     float64
	DodgeCooldown int
	DodgeQueue    int // сколько ещё рывков в текущей серии
	DodgePause    int // пауза между рывками серии

	// Стрелок: движение по орбите вокруг игрока.
	OrbitDir         int // -1 или +1, направление обхода
	OrbitChangeTimer int // тиков до смены направления
}

func NewEnemy(t EnemyType, x, y float64) Enemy {
	hp := t.Stats().MaxHP
	return Enemy{
		Type:   t,
		X:      x,
		Y:      y,
		LastX:  x,
		LastY:  y,
		HP:     hp,
		LastHP: hp,
	}
}

func (e *Enemy) Rect() geometry.Rect {
	s := e.Type.Stats().Size
	return geometry.Rect{X: e.X, Y: e.Y, W: s, H: s}
}

func (e *Enemy) Center() (float64, float64) {
	s := e.Type.Stats().Size
	return e.X + s/2, e.Y + s/2
}

func (e *Enemy) IsDead() bool {
	return e.HP <= 0
}

func (e *Enemy) TakeDamage(n int) {
	if n <= 0 {
		return
	}
	e.HP -= n
}

func (e *Enemy) Speed() float64 {
	return EnemyBaseSpeed + float64(e.Type.Stats().SP)*EnemySpeedPerSP
}

// HPRatio — для отрисовки полоски здоровья над врагом.
func (e *Enemy) HPRatio() float64 {
	m := e.Type.Stats().MaxHP
	if m <= 0 {
		return 0
	}
	r := float64(e.HP) / float64(m)
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}
