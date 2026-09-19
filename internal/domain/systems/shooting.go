package systems

import (
	"math"

	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

func StepShooting(w *world.World, in input.PlayerInput) {
	p := &w.Player
	if p.FireCooldownTimer > 0 {
		p.FireCooldownTimer--
	}
	if !in.Fire || p.FireCooldownTimer > 0 {
		return
	}

	weapon := p.Weapon.Stats()

	cx, cy := p.Center()
	dx := in.AimX - cx
	dy := in.AimY - cy
	l := math.Hypot(dx, dy)
	if l == 0 {
		return
	}
	dx /= l
	dy /= l

	w.Bullets = append(w.Bullets, world.Bullet{
		X:      cx - weapon.BulletSize/2,
		Y:      cy - weapon.BulletSize/2,
		VX:     dx * weapon.BulletSpeed,
		VY:     dy * weapon.BulletSpeed,
		Life:   weapon.BulletLifetime,
		Size:   weapon.BulletSize,
		Weapon: p.Weapon,
	})
	p.FireCooldownTimer = weapon.FireCooldown
}
