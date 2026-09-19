package systems

import "Sqwave/internal/domain/world"

func StepBullets(w *world.World) {
	alive := w.Bullets[:0]
	for _, b := range w.Bullets {
		b.X += b.VX
		b.Y += b.VY
		b.Life--

		if b.Life <= 0 {
			continue
		}
		if b.X < -b.Size || b.X > world.WorldWidth+b.Size ||
			b.Y < -b.Size || b.Y > world.WorldHeight+b.Size {
			continue
		}

		br := b.Rect()
		hit := false
		for _, wall := range w.Walls {
			if br.Intersects(wall) {
				hit = true
				break
			}
		}
		if hit {
			continue
		}

		alive = append(alive, b)
	}
	w.Bullets = alive
}
