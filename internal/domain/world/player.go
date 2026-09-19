package world

import "Sqwave/internal/domain/geometry"

type Player struct {
	X, Y              float64
	DashTimer         int
	DashCooldownTimer int
	Weapon            WeaponType
	FireCooldownTimer int
}

func (p *Player) Rect() geometry.Rect {
	return geometry.Rect{X: p.X, Y: p.Y, W: PlayerSize, H: PlayerSize}
}

func (p *Player) Center() (float64, float64) {
	return p.X + PlayerSize/2, p.Y + PlayerSize/2
}

func (p *Player) Dashing() bool {
	return p.DashTimer > 0
}
