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

	losX, losY, los := losToRect(w, ecx, ecy, w.Player.Rect())

	approachSpeed := 0.0
	if e.PrevPlayerDist > 0.01 {
		approachSpeed = e.PrevPlayerDist - dist
	}
	e.PrevPlayerDist = dist

	panic := 0.0
	if dist < stats.PreferredMin && stats.PreferredMin > 0 {
		panic = 1 - dist/stats.PreferredMin
	}
	if approachSpeed > 0 {
		sf := math.Min(approachSpeed/world.ShooterFastApproach, 1.2)
		if sf > panic {
			panic = sf
		}
	}

	fullPanic := panic >= world.ShooterFullPanic
	// Видимость с учётом размера снаряда: линия проходит не вплотную
	// к стене. Для нахождения видимой точки тела игрока —
	// losToRect, для проверки что снаряд долетит — доп. запас.
	losX, losY, canSee := losToRectWithPad(w, ecx, ecy, w.Player.Rect(),
		stats.ProjectileSize/2+1)

	if canSee && !fullPanic && e.AttackTimer == 0 {
		fireEnemyProjectileAt(w, e, ecx, ecy, losX, losY)
		span := stats.AttackCooldownMax - stats.AttackCooldownMin
		e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
	}

	awayX := ecx - pcx
	awayY := ecy - pcy
	awayLen := math.Hypot(awayX, awayY)
	if awayLen < 0.01 {
		awayX, awayY = 1, 0
	} else {
		awayX /= awayLen
		awayY /= awayLen
	}

	if fullPanic {
		rx := ecx + awayX*world.ShooterFastRetreatDist
		ry := ecy + awayY*world.ShooterFastRetreatDist
		moveEnemyToPoint(w, e, ecx, ecy, rx, ry)
		return
	}

	if panic >= world.ShooterPanicThreshold {
		speed := e.Speed() * world.ShooterSlowRetreatMul
		moveEnemyX(w, e, awayX*speed)
		moveEnemyY(w, e, awayY*speed)
		return
	}

	if los {
		if dist > stats.PreferredMax {
			advanceToward(w, e, ecx, ecy, pcx, pcy)
			return
		}
		e.Path = nil
		e.PathIndex = 0
		e.PathCooldown = 0
		return
	}

	sx, sy, ok := shootPosition(w, e, pcx, pcy)
	if !ok {
		moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
		return
	}
	moveEnemyToPoint(w, e, ecx, ecy, sx, sy)
}

func shootPosition(w *world.World, e *world.Enemy, pcx, pcy float64) (float64, float64, bool) {
	if e.ShootPosTimer > 0 {
		e.ShootPosTimer--
		if hasLineOfSight(w, e.ShootPosX, e.ShootPosY, pcx, pcy) {
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

	pcxCell, pcyCell := grid.WorldToCell(pcx, pcy)
	radiusInCells := int((stats.PreferredMax + grid.CellSize) / grid.CellSize)

	bestScore := math.Inf(-1)
	var bestX, bestY float64
	found := false

	for dy := -radiusInCells; dy <= radiusInCells; dy++ {
		for dx := -radiusInCells; dx <= radiusInCells; dx++ {
			cx := pcxCell + dx
			cy := pcyCell + dy
			if grid.IsBlocked(cx, cy) {
				continue
			}
			wx, wy := grid.CellCenter(cx, cy)

			d := math.Hypot(wx-pcx, wy-pcy)
			if d < stats.PreferredMin || d > stats.PreferredMax {
				continue
			}
			if _, _, ok := losToRectWithPad(w, wx, wy, w.Player.Rect(),
				stats.ProjectileSize/2+1); !ok {
				continue
			}
			if math.Hypot(wx-ecx, wy-ecy) < grid.CellSize {
				continue
			}

			idealDist := (stats.PreferredMin + stats.PreferredMax) / 2
			distScore := -math.Abs(d - idealDist)
			shooterDist := math.Hypot(wx-ecx, wy-ecy)
			pathScore := -shooterDist * 0.3

			score := distScore + pathScore
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
