package world

import "Sqwave/internal/domain/geometry"

type Bullet struct {
	X, Y   float64
	VX, VY float64
	Life   int
}

func (b *Bullet) Rect() geometry.Rect {
	return geometry.Rect{X: b.X, Y: b.Y, W: BulletSize, H: BulletSize}
}
