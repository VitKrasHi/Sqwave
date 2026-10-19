package systems

import (
	"Sqwave/internal/domain/world"
)

func stepInfantry(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	// Фазы замаха + рывка.
	if processMeleeAttack(w, e, stats) {
		return
	}

	if e.AttackTimer > 0 {
		e.AttackTimer--
	}

	directOK := hasClearance(w, ecx, ecy, pcx, pcy, stats.Size/2+world.AgentPredictPad)
	// Триггер замаха — по MeleeTriggerRange, не по MeleeRange.
	triggerRange := stats.MeleeTriggerRange
	if triggerRange <= 0 {
		triggerRange = stats.MeleeRange
	}
	canTrigger := dist <= triggerRange && hasLineOfSight(w, ecx, ecy, pcx, pcy)

	if canTrigger && e.AttackTimer == 0 {
		e.SwingActive = true
		e.SwingTimer = 0
		e.SwingTargetX = pcx
		e.SwingTargetY = pcy
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
