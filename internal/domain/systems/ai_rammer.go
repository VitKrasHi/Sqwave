package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

func stepRammer(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	size := stats.Size

	if e.DashCooldown > 0 {
		e.DashCooldown--
	}

	// --- Активный рывок ---
	if e.DashActive {
		e.DashTimer++

		// Движение по фиксированному направлению.
		speed := e.Speed() * stats.DashSpeedMul
		dx := e.DashDirX * speed
		dy := e.DashDirY * speed

		wallHit := false

		targetX := e.X + dx
		if !rammerHitsWall(w, geometry.Rect{X: targetX, Y: e.Y, W: size, H: size}) {
			e.X = targetX
		} else {
			wallHit = true
		}

		targetY := e.Y + dy
		if !rammerHitsWall(w, geometry.Rect{X: e.X, Y: targetY, W: size, H: size}) {
			e.Y = targetY
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

		// Проверка попадания по игроку — урон один раз за рывок,
		// но рывок НЕ прерывается.
		if !e.DashHit && rammerHitsPlayer(e, &w.Player) {
			w.Player.TakeDamage(stats.DashDamage)
			e.DashHit = true
		}

		// Рывок прерывается только по времени или по стене.
		if wallHit || e.DashTimer >= stats.DashDuration {
			endRammerDash(w, e, stats)
		}
		return
	}

	// --- Замах щитом в ближнем бою ---
	if e.SwingActive {
		e.SwingTimer++
		if e.SwingTimer >= stats.SwingDuration {
			e.SwingActive = false

			ecx, ecy = e.Center()
			pcx2, pcy2 := w.Player.Center()
			d := math.Hypot(pcx2-ecx, pcy2-ecy)

			if d <= stats.MeleeRange && hasLineOfSight(w, ecx, ecy, pcx2, pcy2) {
				w.Player.TakeDamage(stats.Damage)
			}

			span := stats.AttackCooldownMax - stats.AttackCooldownMin
			e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
		}
		return
	}

	if e.AttackTimer > 0 {
		e.AttackTimer--
	}

	// --- Ближний бой ---
	if dist <= stats.MeleeRange && e.AttackTimer == 0 &&
		hasLineOfSight(w, ecx, ecy, pcx, pcy) {
		e.SwingActive = true
		e.SwingTimer = 0
		return
	}

	// --- Игрок вне радиуса — тараним ---
	if dist > stats.MeleeRange && e.DashCooldown == 0 &&
		hasLineOfSight(w, ecx, ecy, pcx, pcy) {
		aimX := pcx - ecx
		aimY := pcy - ecy
		l := math.Hypot(aimX, aimY)
		if l > 0.01 {
			e.DashActive = true
			e.DashTimer = 0
			e.DashDirX = aimX / l
			e.DashDirY = aimY / l
			e.DashHit = false
			return
		}
	}

	// --- Обычная ходьба ---
	if hasLineOfSight(w, ecx, ecy, pcx, pcy) {
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

// rammerHitsWall — проверка только стен, без врагов.
// Таран продавливает союзников, но останавливается о стены.
func rammerHitsWall(w *world.World, r geometry.Rect) bool {
	for _, wall := range w.Walls {
		if r.Intersects(wall) {
			return true
		}
	}
	return false
}

// rammerHitsPlayer — пересекается ли хитбокс тарана с игроком.
func rammerHitsPlayer(e *world.Enemy, p *world.Player) bool {
	return e.Rect().Intersects(p.Rect())
}
