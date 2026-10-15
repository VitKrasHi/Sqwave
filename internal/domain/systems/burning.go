package systems

import "Sqwave/internal/domain/world"

// StepBurning — догорание игрока.
// 1 урон раз в BurnRate тиков, пока BurnTimer > 0.
func StepBurning(w *world.World) {
	p := &w.Player
	if p.IsDead() {
		p.BurnTimer = 0
		p.BurnTick = 0
		return
	}
	if p.BurnTimer <= 0 {
		p.BurnTick = 0
		return
	}

	p.BurnTimer--
	p.BurnTick++

	const tickRate = 60 // 1 dmg/sec при 60 TPS
	if p.BurnTick >= tickRate {
		p.BurnTick = 0
		p.TakeDamage(1)
	}
}
