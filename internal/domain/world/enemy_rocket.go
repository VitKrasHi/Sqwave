package world

import "Sqwave/internal/domain/geometry"

// EnemyRocket — снаряд ракетчика. Летит в направлении VX/VY,
// каждый тик слегка доворачивает скорость к игроку (homing).
// При столкновении с чем угодно или по истечении Life создаёт
// взрыв (world.Explosion) и наносит AoE-урон игроку.
type EnemyRocket struct {
	X, Y    float64
	VX, VY  float64
	Size    float64
	Life    int
	MaxLife int

	ExplosionDamage    int
	ExplosionDamageMul float64
	ExplosionRadius    float64
	ExplosionLife      int

	HomingRate float64
}

func (r *EnemyRocket) Rect() geometry.Rect {
	return geometry.Rect{X: r.X - r.Size/2, Y: r.Y - r.Size/2, W: r.Size, H: r.Size}
}
