package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

// StepEnemyRockets — движение ракет с наведением.
// Ракета НЕ взрывается о игрока — пролетает сквозь него. Взрыв
// происходит только по истечении LifeTime, при ударе о стену
// или при выходе за границы мира. Урон наносится только взрывом.
func StepEnemyRockets(w *world.World) {
	p := &w.Player
	pcx, pcy := p.Center()

	alive := w.EnemyRockets[:0]

	for _, r := range w.EnemyRockets {
		// Наведение.
		speed := math.Hypot(r.VX, r.VY)
		if speed > 0.01 {
			toPlayerX := pcx - r.X
			toPlayerY := pcy - r.Y
			toLen := math.Hypot(toPlayerX, toPlayerY)
			if toLen > 0.01 {
				dirX := toPlayerX / toLen
				dirY := toPlayerY / toLen
				curX := r.VX / speed
				curY := r.VY / speed
				nx := curX + (dirX-curX)*r.HomingRate
				ny := curY + (dirY-curY)*r.HomingRate
				l := math.Hypot(nx, ny)
				if l > 0.01 {
					r.VX = nx / l * speed
					r.VY = ny / l * speed
				}
			}
		}

		oldX, oldY := r.X, r.Y
		r.X += r.VX
		r.Y += r.VY
		r.Life--

		// CCD: проверяем отрезок против стен.
		hitWall := false
		for _, wall := range w.Walls {
			if _, ok := geometry.RaySegmentIntersectsRect(oldX, oldY, r.X, r.Y, wall); ok {
				hitWall = true
				break
			}
		}

		outOfWorld := r.X < -r.Size || r.X > world.WorldWidth+r.Size ||
			r.Y < -r.Size || r.Y > world.WorldHeight+r.Size

		expired := r.Life <= 0

		if !hitWall && !outOfWorld && !expired {
			alive = append(alive, r)
			continue
		}

		// --- Взрыв ---
		// Урон наносится только взрывом, если игрок в радиусе.
		if !p.IsDead() {
			if math.Hypot(pcx-r.X, pcy-r.Y) <= r.ExplosionRadius {
				p.TakeDamage(r.ExplosionDamage)
			}
		}

		spawnExplosion(w, r.X, r.Y, r.ExplosionRadius, r.ExplosionLife)
	}

	w.EnemyRockets = alive
}
