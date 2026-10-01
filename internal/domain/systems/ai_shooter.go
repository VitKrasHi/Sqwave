package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepShooter(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	if e.AttackTimer > 0 {
		e.AttackTimer--
	}

	// Застряли — выходим через A*.
	if e.StuckTicks > 10 {
		e.Path = nil
		e.PathIndex = 0
		e.PathCooldown = 0
		moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
		return
	}

	losX, losY, los := losToRectWithPad(w, ecx, ecy, w.Player.Rect(),
		stats.ProjectileSize/2+1)

	// Стреляем только когда видим игрока.
	if los && e.AttackTimer == 0 {
		fireEnemyProjectileAt(w, e, ecx, ecy, losX, losY)
		span := stats.AttackCooldownMax - stats.AttackCooldownMin
		e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
	}

	// Вектор от игрока к стрелку и его длина.
	awayX := ecx - pcx
	awayY := ecy - pcy
	awayLen := math.Hypot(awayX, awayY)
	if awayLen < 0.01 {
		awayX, awayY = 1, 0
		awayLen = 1
	} else {
		awayX /= awayLen
		awayY /= awayLen
	}

	// Перпендикуляр — для движения по орбите.
	perpX := -awayY
	perpY := awayX

	// Если игрока не видим — обходим стену через поиск позиции.
	if !los {
		sx, sy, ok := shootPosition(w, e, pcx, pcy)
		if !ok {
			moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
			return
		}
		moveEnemyToPoint(w, e, ecx, ecy, sx, sy)
		return
	}

	// --- Игрок виден. Работаем по зонам дистанции ---

	switch {
	case dist > stats.PreferredMax:
		// Игрок убегает — преследуем.
		moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)

	case dist < stats.PreferredMin:
		// Игрок приближается — пятимся назад с половиной скорости.
		// Добавляем лёгкий боковой сдвиг, чтобы не пятиться строго по прямой.
		drift := math.Sin(float64(e.AITimer)*0.1) * 0.3
		dx := awayX + perpX*drift
		dy := awayY + perpY*drift
		l := math.Hypot(dx, dy)
		if l > 0.01 {
			dx /= l
			dy /= l
		}
		speed := e.Speed() * 0.5
		moveEnemyX(w, e, dx*speed)
		moveEnemyY(w, e, dy*speed)

	default:
		// Комфортная зона — движемся по орбите вокруг игрока.
		// Направление выбирается один раз и держится N тиков,
		// потом меняется случайно. Это делает стрелка
		// непредсказуемым, но не позволяет ему метаться.
		e.AITimer++
		if e.AITimer >= e.OrbitChangeTimer {
			e.AITimer = 0
			e.OrbitChangeTimer = 60 + w.Rng.Intn(90) // 1..2.5 сек
			// Направление орбиты: -1 или +1.
			if w.Rng.Intn(2) == 0 {
				e.OrbitDir = -1
			} else {
				e.OrbitDir = 1
			}
		}

		// Двигаемся перпендикулярно к игроку, с лёгким
		// радиальным подтягиванием к центру зоны.
		idealDist := (stats.PreferredMin + stats.PreferredMax) / 2
		radialErr := dist - idealDist
		// Нормализуем ошибку: 0..1 по половине ширины зоны.
		halfWidth := (stats.PreferredMax - stats.PreferredMin) / 2
		radialPull := 0.0
		if halfWidth > 0.01 {
			radialPull = clamp(radialErr/halfWidth, -1, 1) * 0.5
		}

		dx := perpX*float64(e.OrbitDir) + awayX*radialPull
		dy := perpY*float64(e.OrbitDir) + awayY*radialPull
		l := math.Hypot(dx, dy)
		if l < 0.01 {
			return
		}
		dx /= l
		dy /= l
		speed := e.Speed() * 0.8
		moveEnemyX(w, e, dx*speed)
		moveEnemyY(w, e, dy*speed)
	}
}

func clamp(v, lo, hi float64) float64 {
	if v < lo {
		return lo
	}
	if v > hi {
		return hi
	}
	return v
}

// shootPosition — поиск точки с LOS на игрока в комфортной зоне.
// Используется, когда стрелок потерял видимость.
func shootPosition(w *world.World, e *world.Enemy, pcx, pcy float64) (float64, float64, bool) {
	if e.ShootPosTimer > 0 {
		e.ShootPosTimer--
		if _, _, ok := losToRectWithPad(w, e.ShootPosX, e.ShootPosY,
			w.Player.Rect(), e.Type.Stats().ProjectileSize/2+1); ok {
			return e.ShootPosX, e.ShootPosY, true
		}
		e.ShootPosTimer = 0
	}

	grid := w.NavGrid
	if grid == nil {
		return 0, 0, false
	}
	stats := e.Type.Stats()
	ecx, ecy := e.Center()

	ideal := (stats.PreferredMin + stats.PreferredMax) / 2
	distances := []float64{stats.PreferredMin, ideal, stats.PreferredMax}

	baseAngle := w.Rng.Float64() * math.Pi * 2

	bestScore := math.Inf(-1)
	var bestX, bestY float64
	found := false

	for i := 0; i < 8; i++ {
		angle := baseAngle + float64(i)*math.Pi/4
		cos := math.Cos(angle)
		sin := math.Sin(angle)

		for _, d := range distances {
			wx := pcx + cos*d
			wy := pcy + sin*d

			cxCell, cyCell := grid.WorldToCell(wx, wy)
			if grid.IsBlocked(cxCell, cyCell) {
				continue
			}
			if math.Hypot(wx-ecx, wy-ecy) < 40 {
				continue
			}
			if _, _, ok := losToRectWithPad(w, wx, wy, w.Player.Rect(),
				stats.ProjectileSize/2+1); !ok {
				continue
			}

			shooterDist := math.Hypot(wx-ecx, wy-ecy)
			score := -shooterDist
			if score > bestScore {
				bestScore = score
				bestX, bestY = wx, wy
				found = true
			}
		}
	}

	if found {
		e.ShootPosX, e.ShootPosY = bestX, bestY
		e.ShootPosTimer = 30
	}
	return bestX, bestY, found
}
