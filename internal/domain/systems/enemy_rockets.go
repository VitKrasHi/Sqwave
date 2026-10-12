package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

// StepEnemyRockets — движение ракет с лёгким наведением.
// Наведение: вектор скорости плавно поворачивается в сторону игрока
// с коэффициентом HomingRate (0.02..0.05). Чем больше — тем точнее.
func StepEnemyRockets(w *world.World) {
	p := &w.Player
	pcx, pcy := p.Center()
	playerRect := p.Rect()

	alive := w.EnemyRockets[:0]

	for _, r := range w.EnemyRockets {
		speed := math.Hypot(r.VX, r.VY)
		if speed > 0.01 {
			// Направление к игроку.
			toPlayerX := pcx - r.X
			toPlayerY := pcy - r.Y
			toLen := math.Hypot(toPlayerX, toPlayerY)
			if toLen > 0.01 {
				dirX := toPlayerX / toLen
				dirY := toPlayerY / toLen

				// Текущее направление.
				curX := r.VX / speed
				curY := r.VY / speed

				// Смешиваем: newDir = cur + (target - cur) * rate.
				nx := curX + (dirX-curX)*r.HomingRate
				ny := curY + (dirY-curY)*r.HomingRate

				l := math.Hypot(nx, ny)
				if l > 0.01 {
					r.VX = nx / l * speed
					r.VY = ny / l * speed
				}
			}
		}

		r.X += r.VX
		r.Y += r.VY
		r.Life--

		if r.Life <= 0 {
			continue
		}
		if r.X < -r.Size || r.X > world.WorldWidth+r.Size ||
			r.Y < -r.Size || r.Y > world.WorldHeight+r.Size {
			continue
		}

		rect := r.Rect()

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
			p.TakeDamage(r.Damage)
			continue
		}

		alive = append(alive, r)
	}

	w.EnemyRockets = alive
}
