package world

import "Sqwave/internal/domain/geometry"

type EnemyProjectile struct {
	X, Y   float64
	VX, VY float64
	Size   float64
	Life   int
	Damage int
}

func (p *EnemyProjectile) Rect() geometry.Rect {
	return geometry.Rect{X: p.X, Y: p.Y, W: p.Size, H: p.Size}
}
