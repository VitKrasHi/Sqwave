package systems

import "Sqwave/internal/domain/world"

func StepEnemyProjectiles(w *world.World) {
	p := &w.Player
	playerRect := p.Rect()

	alive := w.EnemyProjectiles[:0]
	for _, proj := range w.EnemyProjectiles {
		proj.X += proj.VX
		proj.Y += proj.VY
		proj.Life--

		if proj.Life <= 0 {
			continue
		}
		if proj.X < -proj.Size || proj.X > world.WorldWidth+proj.Size ||
			proj.Y < -proj.Size || proj.Y > world.WorldHeight+proj.Size {
			continue
		}

		rect := proj.Rect()

		hitWall := false
		for _, wall := range w.Walls {
			if rect.Intersects(wall) {
				hitWall = true
				break
			}
		}
		if hitWall {
			continue
		}

		if !p.IsDead() && rect.Intersects(playerRect) {
			mul := proj.DamageMul
			if mul <= 0 {
				mul = 1
			}
			dmg := int(float64(proj.Damage) * mul)
			if dmg < 1 {
				dmg = 1
			}
			p.TakeDamage(dmg)

			if proj.Flame {
				fStats := world.EnemyFlamer.Stats()
				p.ApplyBurn(fStats.FlameBurnTicks)
			}
			continue
		}

		alive = append(alive, proj)
	}
	w.EnemyProjectiles = alive
}
