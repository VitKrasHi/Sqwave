package world

import "Sqwave/internal/domain/geometry"

type Player struct {
	X, Y              float64
	DashTimer         int
	DashCooldownTimer int
	Weapon            WeaponType
	FireCooldownTimer int
	AimCharge         int
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

// IsAiming — снайперка заряжается прямо сейчас.
func (p *Player) IsAiming() bool {
	return p.Weapon == WeaponSniper && p.AimCharge > 0
}

// AimChargeRatio — 0 при начале прицеливания, 1 при полном заряде.
func (p *Player) AimChargeRatio() float64 {
	if !p.IsAiming() {
		return 0
	}
	ct := p.Weapon.Stats().ChargeTime
	if ct <= 0 {
		return 0
	}
	r := float64(p.AimCharge) / float64(ct)
	if r > 1 {
		return 1
	}
	return r
}
