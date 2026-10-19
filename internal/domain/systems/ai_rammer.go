package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

const rammerDashWindupTicks = 15 // ~0.25 сек замедления перед рывком

func stepRammer(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	size := stats.Size

	if e.DashCooldown > 0 {
		e.DashCooldown--
	}

	// --- Активный рывок (молот) ---
	if e.DashActive {
		e.DashTimer++

		speed := e.Speed() * stats.DashSpeedMul
		dx := e.DashDirX * speed
		dy := e.DashDirY * speed

		wallHit := false

		if !rammerHitsWall(w, geometry.Rect{X: e.X + dx, Y: e.Y, W: size, H: size}) {
			e.X += dx
		} else {
			wallHit = true
		}
		if !rammerHitsWall(w, geometry.Rect{X: e.X, Y: e.Y + dy, W: size, H: size}) {
			e.Y += dy
		} else {
			wallHit = true
		}

		if e.X < 0 {
			e.X = 0
			wallHit = true
		}
		if e.Y < 0 {
			e.Y = 0
			wallHit = true
		}
		if e.X+size > world.WorldWidth {
			e.X = world.WorldWidth - size
			wallHit = true
		}
		if e.Y+size > world.WorldHeight {
			e.Y = world.WorldHeight - size
			wallHit = true
		}

		if !e.DashHit && rammerHitsPlayer(e, &w.Player) {
			dmg := int(float64(stats.DashDamage) * e.DamageMul())
			TryMeleeDamage(w, e, dmg)
			e.DashHit = true
		}

		if wallHit || e.DashTimer >= stats.DashDuration {
			endRammerDash(w, e, stats)
		}
		return
	}

	// --- Замах перед рывком (замедление) ---
	if e.DashWindup {
		e.DashWindupTimer--

		if hasLineOfSight(w, ecx, ecy, pcx, pcy) && dist > 0.01 {
			dx := (pcx - ecx) / dist
			dy := (pcy - ecy) / dist
			speed := e.Speed() * 0.3
			moveEnemyX(w, e, dx*speed)
			moveEnemyY(w, e, dy*speed)
		}

		if e.DashWindupTimer <= 0 {
			e.DashWindup = false
			aimX := pcx - ecx
			aimY := pcy - ecy
			l := math.Hypot(aimX, aimY)
			if l > 0.01 {
				e.DashActive = true
				e.DashTimer = 0
				e.DashDirX = aimX / l
				e.DashDirY = aimY / l
				e.DashHit = false
			}
		}
		return
	}

	// --- Melee-фазы (меч тарана) ---
	if processMeleeAttack(w, e, stats) {
		return
	}

	if e.AttackTimer > 0 {
		e.AttackTimer--
	}

	los := hasLineOfSight(w, ecx, ecy, pcx, pcy)

	if dist <= stats.MeleeRange && e.AttackTimer == 0 && los {
		e.SwingActive = true
		e.SwingTimer = 0
		e.SwingTargetX = pcx
		e.SwingTargetY = pcy
		return
	}

	if dist > stats.MeleeRange && e.DashCooldown == 0 && los {
		e.DashWindup = true
		e.DashWindupTimer = rammerDashWindupTicks
		return
	}

	if los {
		moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)
	} else {
		moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
	}
}

func endRammerDash(w *world.World, e *world.Enemy, stats world.EnemyStats) {
	e.DashActive = false
	e.DashHit = false
	span := stats.DashCooldownMax - stats.DashCooldownMin
	e.DashCooldown = stats.DashCooldownMin + w.Rng.Intn(span+1)
}

func rammerHitsWall(w *world.World, r geometry.Rect) bool {
	for _, wall := range w.Walls {
		if r.Intersects(wall) {
			return true
		}
	}
	return false
}

func rammerHitsPlayer(e *world.Enemy, p *world.Player) bool {
	return e.Rect().Intersects(p.Rect())
}
