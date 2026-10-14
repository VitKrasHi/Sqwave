package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepFleer(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	p := &w.Player

	// --- Уже сцепился с игроком ---
	if e.Latched {
		if p.IsDead() {
			e.Latched = false
			p.Held = false
			p.HeldBy = -1
			return
		}

		// Урон.
		p.TakeDamage(stats.FleerLatchDmg)

		// Подтягиваем игрока к беглецу.
		ecx, ecy := e.Center()
		pcx, pcy := p.Center()
		dx := ecx - pcx
		dy := ecy - pcy
		l := math.Hypot(dx, dy)
		if l > 20 {
			// Тянем игрока к беглецу на 1.5 px/тик.
			pull := 1.5
			if l < 30 {
				pull = l - 20
			}
			p.X += dx / l * pull
			p.Y += dy / l * pull
		}
		return
	}

	// --- Медленный поворот взгляда к игроку ---
	targetAngle := math.Atan2(pcy-ecy, pcx-ecx)
	diff := angleDiff(targetAngle, e.AimAngle)
	maxTurn := stats.FleerTurnRate
	if diff > maxTurn {
		diff = maxTurn
	}
	if diff < -maxTurn {
		diff = -maxTurn
	}
	e.AimAngle += diff

	// --- Проверка захвата ---
	if dist < stats.FleerGrabRange {
		toPlayerAngle := math.Atan2(pcy-ecy, pcx-ecx)
		angleFromAim := math.Abs(angleDiff(toPlayerAngle, e.AimAngle))
		halfCone := stats.FleerConeWidth * math.Pi / 360

		if angleFromAim < halfCone {
			// Захват!
			e.Latched = true
			p.Held = true
			p.HeldBy = indexOfEnemy(w, e)
			return
		}
	}

	// --- Движение в направлении взгляда ---
	// Ключевое отличие от других врагов: движемся туда, КУДА СМОТРИМ,
	// а не напрямую к игроку. Позволяет закрутить и зайти со спины.
	moveDirX := math.Cos(e.AimAngle)
	moveDirY := math.Sin(e.AimAngle)
	speed := e.Speed()
	moveEnemyX(w, e, moveDirX*speed)
	moveEnemyY(w, e, moveDirY*speed)
}

// angleDiff — разница между двумя углами в диапазоне [-π, π].
func angleDiff(a, b float64) float64 {
	d := a - b
	for d > math.Pi {
		d -= 2 * math.Pi
	}
	for d < -math.Pi {
		d += 2 * math.Pi
	}
	return d
}
