package systems

import "Sqwave/internal/domain/world"

func StepBullets(w *world.World) {
	alive := w.Bullets[:0]
	for _, b := range w.Bullets {
		b.Life--
		if b.Life <= 0 {
			continue
		}
		alive = append(alive, b)
	}
	w.Bullets = alive
}
