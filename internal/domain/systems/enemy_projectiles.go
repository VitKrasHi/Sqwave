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
			dmg := int(float64(proj.Damage) * proj.DamageMul)
			p.TakeDamage(dmg)

			if proj.Flame {
				// Поджигаем игрока.
				// Найти длительность и rate можно из статов огнемётчика.
				// Проще — захардкодить через константы статов.
				fStats := world.EnemyFlamer.Stats()
				p.ApplyBurn(fStats.FlameBurnTicks)
			}
			continue
		}

		alive = append(alive, proj)
	}
	w.EnemyProjectiles = alive
}
