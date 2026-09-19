package systems

import (
	"math"

	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

func StepShooting(w *world.World, in input.PlayerInput) {
	if w.FireCooldownTimer > 0 {
		w.FireCooldownTimer--
	}
	if !in.Fire || w.FireCooldownTimer > 0 {
		return
	}

	cx, cy := w.Player.Center()
	dx := in.AimX - cx
	dy := in.AimY - cy
	l := math.Hypot(dx, dy)
	if l == 0 {
		return
	}
	dx /= l
	dy /= l

	w.Bullets = append(w.Bullets, world.Bullet{
		X:    cx - world.BulletSize/2,
		Y:    cy - world.BulletSize/2,
		VX:   dx * world.BulletSpeed,
		VY:   dy * world.BulletSpeed,
		Life: world.BulletLifetime,
	})
	w.FireCooldownTimer = world.FireCooldown
}
