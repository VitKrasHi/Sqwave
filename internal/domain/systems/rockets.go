package systems

import (
	"Sqwave/internal/domain/world"
)

func StepRockets(w *world.World) {
	alive := w.Rockets[:0]
	for _, r := range w.Rockets {
		r.X += r.VX
		r.Y += r.VY
		r.Life--

		rect := r.Rect()
		hitWall := false
		for _, wall := range w.Walls {
			if rect.Intersects(wall) {
				hitWall = true
				break
			}
		}

		outOfWorld := r.X < 0 || r.X > world.WorldWidth ||
			r.Y < 0 || r.Y > world.WorldHeight

		if hitWall || outOfWorld || r.Life <= 0 {
			spawnExplosion(w, r.X, r.Y, r.ExplosionRadius, r.ExplosionLife)
			continue
		}

		alive = append(alive, r)
	}
	w.Rockets = alive
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
