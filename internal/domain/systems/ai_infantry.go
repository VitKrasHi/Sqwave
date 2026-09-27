package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepInfantry(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	bodyHalf := stats.Size / 2

	// Фаза замаха.
	if e.SwingActive {
		e.SwingTimer++
		if e.SwingTimer >= stats.SwingDuration {
			e.SwingActive = false

			ecx, ecy = e.Center()
			pcx, pcy = w.Player.Center() // ← было p.Center()
			dist = math.Hypot(pcx-ecx, pcy-ecy)

			if dist <= stats.MeleeRange && hasClearance(w, ecx, ecy, pcx, pcy, bodyHalf) {
				w.Player.TakeDamage(stats.Damage)
			}

			span := stats.AttackCooldownMax - stats.AttackCooldownMin
			e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
		}
	}

	if e.AttackTimer > 0 {
		e.AttackTimer--
	}

	directOK := hasClearance(w, ecx, ecy, pcx, pcy, bodyHalf+world.AgentPredictPad)

	if dist <= stats.MeleeRange && e.AttackTimer == 0 && directOK {
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
		e.ForcePathTimer = 25
		e.StuckTicks = 0
		return
	}

	if directOK {
		moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)
	} else {
		moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
	}
}
