package world

import "Sqwave/internal/domain/geometry"

type EnemyType int

const (
	EnemyInfantry EnemyType = iota
)

type EnemyStats struct {
	Name   string
	MaxHP  int
	SP     int
	Damage int
	Size   float64

	MeleeRange        float64 // дистанция, с которой враг бьёт
	AttackCooldownMin int     // минимум тиков между ударами
	AttackCooldownMax int     // максимум
	SwingDuration     int     // длительность замаха
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

	// Knockback — импульс от удара щитом.
	KnockbackVX    float64
	KnockbackVY    float64
	KnockbackTimer int

	// LastSwingHitID — номер замаха игрока, который уже попал.
	// Защищает от повторного урона в одном провороте.
	LastSwingHitID int
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
