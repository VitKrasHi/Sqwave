package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepRocketeer(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	if e.RocketFireTimer > 0 {
		e.RocketFireTimer--
	}

	// Стреляем, если в радиусе и есть LOS.
	if e.RocketFireTimer == 0 && dist <= stats.RocketFireRange &&
		hasLineOfSight(w, ecx, ecy, pcx, pcy) {
		fireEnemyRocket(w, e, ecx, ecy, pcx, pcy)
		e.RocketFireTimer = stats.RocketCooldown
	}

	// Идём к игроку. Никаких отступлений и паники.
	if hasLineOfSight(w, ecx, ecy, pcx, pcy) {
		moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)
		return
	}
	moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
}

func fireEnemyRocket(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy float64) {
	stats := e.Type.Stats()

	dx := pcx - ecx
	dy := pcy - ecy
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		return
	}
	dx /= l
	dy /= l

	startX := ecx + dx*stats.Size/2
	startY := ecy + dy*stats.Size/2

	w.EnemyRockets = append(w.EnemyRockets, world.EnemyRocket{
		X:               startX,
		Y:               startY,
		VX:              dx * stats.RocketSpeed,
		VY:              dy * stats.RocketSpeed,
		Size:            stats.RocketSize,
		Life:            stats.RocketLife,
		MaxLife:         stats.RocketLife,
		ExplosionDamage: stats.RocketExplosionDamage,
		ExplosionRadius: stats.RocketExplosionRadius,
		ExplosionLife:   stats.RocketExplosionLife,
		HomingRate:      stats.RocketHomingRate,
	})
}
