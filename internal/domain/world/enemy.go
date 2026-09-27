package world

import "Sqwave/internal/domain/geometry"

type EnemyType int

const (
	EnemyInfantry EnemyType = iota
	EnemyShooter
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

	// Для стрелка.
	PreferredMin    float64
	PreferredMax    float64
	ProjectileSpeed float64
	ProjectileSize  float64
	ProjectileLife  int
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
		ProjectileSpeed:   25, // было 4 — почти вдвое быстрее
		ProjectileSize:    8,
		ProjectileLife:    90, // было 150 — короче, чтобы не летели через пол-карты
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
}

func NewEnemy(t EnemyType, x, y float64) Enemy {
	return Enemy{
		Type: t,
		X:    x,
		Y:    y,
		HP:   t.Stats().MaxHP,
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
