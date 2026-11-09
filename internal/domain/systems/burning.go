package systems

import "Sqwave/internal/domain/world"

// StepBurning — догорание игроков.
// 1 урон раз в BurnRate тиков, пока BurnTimer > 0.
func StepBurning(w *world.World) {
	const tickRate = 60 // 1 dmg/sec при 60 TPS

	for i := range w.Players {
		p := &w.Players[i]
		if p.IsDead() {
			p.BurnTimer = 0
			p.BurnTick = 0
			continue
		}
		if p.BurnTimer <= 0 {
			p.BurnTick = 0
			continue
		}

		p.BurnTimer--
		p.BurnTick++

		if p.BurnTick >= tickRate {
			p.BurnTick = 0
			p.TakeDamage(1)
		}
	}
}
