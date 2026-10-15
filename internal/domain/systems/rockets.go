package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func StepRockets(w *world.World) {
	alive := w.Rockets[:0]
	for _, r := range w.Rockets {
		r.X += r.VX
		r.Y += r.VY
		r.Life--

		rect := r.Rect()

		// Кто получил прямое попадание?
		var directHit *world.Enemy
		for i := range w.Enemies {
			e := &w.Enemies[i]
			if e.IsDead() {
				continue
			}
			if rect.Intersects(e.Rect()) {
				directHit = e
				break
			}
		}

		// Стены проверяем, только если не попали во врага.
		hitWall := false
		if directHit == nil {
			for _, wall := range w.Walls {
				if rect.Intersects(wall) {
					hitWall = true
					break
				}
			}
		}

		outOfWorld := r.X < 0 || r.X > world.WorldWidth ||
			r.Y < 0 || r.Y > world.WorldHeight

		if directHit != nil || hitWall || outOfWorld || r.Life <= 0 {
			weapon := r.Weapon.Stats()

			if directHit != nil {
				directHit.TakeDamage(weapon.DirectHitDamage)
			}
			applyExplosionDamage(w, r.X, r.Y, r.ExplosionRadius,
				weapon.ExplosionDamage, directHit)

			spawnExplosion(w, r.X, r.Y, r.ExplosionRadius, r.ExplosionLife)
			continue
		}

		alive = append(alive, r)
	}
	w.Rockets = alive
}

// applyExplosionDamage бьёт всех врагов в радиусе, кроме того,
// кто получил прямое попадание (он уже получил DirectHitDamage).
func applyExplosionDamage(w *world.World, x, y, radius float64, damage int, exclude *world.Enemy) {
	if damage <= 0 {
		return
	}
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() || e == exclude {
			continue
		}
		cx, cy := e.Center()
		if math.Hypot(cx-x, cy-y) <= radius {
			e.TakeDamage(damage)
		}
	}
}

func StepExplosions(w *world.World) {
	alive := w.Explosions[:0]
	for _, e := range w.Explosions {
		e.Life--
		if e.Life <= 0 {
			continue
		}
		alive = append(alive, e)
	}
	w.Explosions = alive
}

func spawnExplosion(w *world.World, x, y, radius float64, life int) {
	if radius <= 0 || life <= 0 {
		return
	}
	w.Explosions = append(w.Explosions, world.Explosion{
		X: x, Y: y,
		Radius:  radius,
		Life:    life,
		MaxLife: life,
	})
}
