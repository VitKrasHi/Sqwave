package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepSpinner(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	// --- Режим дальний/ближний (с гистерезисом) ---
	if e.FarInvulnerable {
		if dist < stats.FarRange*0.85 {
			e.FarInvulnerable = false
		}
	} else {
		if dist > stats.FarRange*1.15 {
			e.FarInvulnerable = true
		}
	}

	// Скорость: в дальнем режиме обычная, в ближнем — ускорение.
	if e.FarInvulnerable {
		e.SpeedScale = 1.0
	} else {
		e.SpeedScale = stats.CloseSpeedMul
	}

	// --- Активный проворот топора ---
	if e.SpinActive {
		e.SpinTimer++
		applySpinHit(w, e, stats)
		if e.SpinTimer >= stats.SpinDuration {
			e.SpinActive = false
			e.SpinTimer = 0
			e.SpinHit = false
			e.AttackTimer = stats.SpinCooldown
		}
		return
	}

	if e.AttackTimer > 0 {
		e.AttackTimer--
	}

	// --- Триггер проворота: только в ближнем режиме ---
	if !e.FarInvulnerable && dist <= stats.MeleeRange && e.AttackTimer == 0 {
		e.SpinActive = true
		e.SpinTimer = 0
		e.SpinHit = false
		return
	}

	// --- Движение к игроку ---
	if hasLineOfSight(w, ecx, ecy, pcx, pcy) {
		moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)
	} else {
		moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
	}
}

// applySpinHit — наносит урон и отбрасывает игрока (один раз за проворот).
func applySpinHit(w *world.World, e *world.Enemy, stats world.EnemyStats) {
	if e.SpinHit {
		return
	}

	p := &w.Player
	if p.IsDead() {
		return
	}

	ecx, ecy := e.Center()
	pcx, pcy := p.Center()
	d := math.Hypot(pcx-ecx, pcy-ecy)
	if d > stats.MeleeRange {
		return
	}

	invuln := p.DashActive

	dmg := int(float64(stats.Damage) * e.DamageMul())
	TryMeleeDamage(w, e, dmg)

	// Отбрасывание — только если игрок не в рывке.
	if !invuln && d > 0.01 {
		p.KnockbackVX = (pcx - ecx) / d * stats.KnockbackForce
		p.KnockbackVY = (pcy - ecy) / d * stats.KnockbackForce
		p.KnockbackTimer = stats.KnockbackTicks
	}

	e.SpinHit = true
}
