package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepInfantry(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	if e.SwingActive {
		e.SwingTimer++
		if e.SwingTimer >= stats.SwingDuration {
			e.SwingActive = false

			ecx, ecy = e.Center()
			pcx, pcy = w.Player.Center()
			dist = math.Hypot(pcx-ecx, pcy-ecy)

			if dist <= stats.MeleeRange && hasLineOfSight(w, ecx, ecy, pcx, pcy) {
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

	directOK := hasClearance(w, ecx, ecy, pcx, pcy, stats.Size/2+world.AgentPredictPad)

	canHit := dist <= stats.MeleeRange && hasLineOfSight(w, ecx, ecy, pcx, pcy)
	if canHit && e.AttackTimer == 0 {
		e.SwingActive = true
		e.SwingTimer = 0
		e.StuckTicks = 0
		e.LastX, e.LastY = e.X, e.Y
		return
	}

	if e.ForcePathTimer > 0 {
		e.ForcePathTimer--
		moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
		return
	}

	if e.StuckTicks > 20 {
		// Долго не двигаемся — принудительно толкаем в сторону
		// от ближайшей стены. Иначе враг будет вечно стоять
		// в углу, даже если A* построит корректный путь.
		e.ForcePathTimer = 25
		e.StuckTicks = 0
		return
	}

	if directOK {
		moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)
		return
	}
	moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
}
