package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepGhost(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	p := &w.Player

	// Атака щитом в упор.
	if e.AttackTimer > 0 {
		e.AttackTimer--
	}

	if dist <= stats.GhostContactRange && e.AttackTimer == 0 && !p.IsDead() {
		p.TakeDamage(int(float64(stats.GhostContactDamage) * e.DamageMul()))
		e.AttackTimer = stats.GhostContactCool
	}

	// --- Движение: напрямую, сквозь стены ---
	dx := pcx - ecx
	dy := pcy - ecy
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		return
	}
	dx /= l
	dy /= l
	speed := e.Speed()

	// Прямое движение без коллизий со стенами.
	newX := e.X + dx*speed
	newY := e.Y + dy*speed
	size := stats.Size

	if newX < 0 {
		newX = 0
	}
	if newY < 0 {
		newY = 0
	}
	if newX+size > world.WorldWidth {
		newX = world.WorldWidth - size
	}
	if newY+size > world.WorldHeight {
		newY = world.WorldHeight - size
	}

	e.X = newX
	e.Y = newY
}
