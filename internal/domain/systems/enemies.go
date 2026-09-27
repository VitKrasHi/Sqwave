package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

func StepEnemies(w *world.World) {
	p := &w.Player

	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() {
			continue
		}

		depenetrate(w, e) // ← добавь эту строку

		// Отбрасывание — приоритет, трекер застревания не работает.
		if e.KnockbackTimer > 0 {
			moveEnemyX(w, e, e.KnockbackVX)
			moveEnemyY(w, e, e.KnockbackVY)
			e.KnockbackTimer--
			e.KnockbackVX *= 0.85
			e.KnockbackVY *= 0.85
			e.LastX, e.LastY = e.X, e.Y
			continue
		}

		stats := e.Type.Stats()
		bodyHalf := stats.Size / 2

		ecx, ecy := e.Center()
		pcx, pcy := p.Center()
		dx := pcx - ecx
		dy := pcy - ecy
		dist := math.Hypot(dx, dy)

		if dist > 0.01 {
			e.FacingX = dx / dist
			e.FacingY = dy / dist
		}

		// Замах.
		if e.SwingActive {
			e.SwingTimer++
			if e.SwingTimer >= stats.SwingDuration {
				e.SwingActive = false

				ecx, ecy = e.Center()
				pcx, pcy = p.Center()
				dist = math.Hypot(pcx-ecx, pcy-ecy)

				if dist <= stats.MeleeRange && hasClearance(w, ecx, ecy, pcx, pcy, bodyHalf) {
					p.TakeDamage(stats.Damage)
				}

				span := stats.AttackCooldownMax - stats.AttackCooldownMin
				e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
			}
			continue
		}

		if e.AttackTimer > 0 {
			e.AttackTimer--
		}

		// Единый трекер застревания.
		moved := math.Hypot(e.X-e.LastX, e.Y-e.LastY)
		if moved < 0.4 {
			e.StuckTicks++
		} else {
			e.StuckTicks = 0
		}
		e.LastX, e.LastY = e.X, e.Y

		// ForcePath — принудительный A* после застревания в direct.
		if e.ForcePathTimer > 0 {
			e.ForcePathTimer--
			moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
			continue
		}

		directOK := hasClearance(w, ecx, ecy, pcx, pcy, bodyHalf+world.AgentPredictPad)

		// Атака.
		if dist <= stats.MeleeRange && e.AttackTimer == 0 && directOK {
			e.SwingActive = true
			e.SwingTimer = 0
			e.StuckTicks = 0
			e.LastX, e.LastY = e.X, e.Y
			continue
		}

		if directOK {
			moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)
		} else {
			moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
		}
	}

	// Удаление мёртвых.
	alive := w.Enemies[:0]
	for _, e := range w.Enemies {
		if !e.IsDead() {
			alive = append(alive, e)
		}
	}
	w.Enemies = alive
}

func hasLineOfSight(w *world.World, x1, y1, x2, y2 float64) bool {
	for _, wall := range w.Walls {
		if _, ok := geometry.RaySegmentIntersectsRect(x1, y1, x2, y2, wall); ok {
			return false
		}
	}
	return true
}

func hasClearance(w *world.World, x1, y1, x2, y2, pad float64) bool {
	for _, wall := range w.Walls {
		expanded := geometry.Rect{
			X: wall.X - pad,
			Y: wall.Y - pad,
			W: wall.W + pad*2,
			H: wall.H + pad*2,
		}
		if _, ok := geometry.RaySegmentIntersectsRect(x1, y1, x2, y2, expanded); ok {
			return false
		}
	}
	return true
}

func moveEnemyDirect(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy float64) {
	e.Path = nil
	e.PathIndex = 0
	e.PathCooldown = 0

	dx := pcx - ecx
	dy := pcy - ecy
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		return
	}
	dx /= l
	dy /= l
	speed := e.Speed()
	moveEnemyX(w, e, dx*speed)
	moveEnemyY(w, e, dy*speed)
}

func moveEnemyViaPath(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy float64) {
	grid := w.NavGrid
	if grid == nil {
		return
	}

	bodyHalf := e.Type.Stats().Size / 2

	// Прошли угол? Возвращаемся к прямому движению.
	if hasClearance(w, ecx, ecy, pcx, pcy, bodyHalf+world.AgentPredictPad) {
		e.Path = nil
		e.PathIndex = 0
		e.PathCooldown = 0
		return
	}

	goalCellX, goalCellY := grid.WorldToCell(pcx, pcy)
	if gx, gy, ok := grid.NearestFree(goalCellX, goalCellY, 2); ok {
		goalCellX, goalCellY = gx, gy
	} else {
		return
	}

	goalCell := [2]int{goalCellX, goalCellY}
	needRecompute := len(e.Path) == 0 ||
		e.PathIndex >= len(e.Path) ||
		e.PathCooldown <= 0 ||
		e.PathGoalCell != goalCell

	if needRecompute {
		startCellX, startCellY := grid.WorldToCell(ecx, ecy)
		if sx, sy, ok := grid.NearestFree(startCellX, startCellY, 2); ok {
			startCellX, startCellY = sx, sy
		} else {
			return
		}

		cells := grid.FindPath(startCellX, startCellY, goalCellX, goalCellY)
		if cells == nil {
			e.Path = nil
			e.PathIndex = 0
			e.PathCooldown = 30
			e.PathGoalCell = [2]int{-1, -1}
			return
		}

		pad := bodyHalf + world.AgentClearance
		e.Path = e.Path[:0]
		cursorX, cursorY := ecx, ecy
		for i := 0; i < len(cells); i++ {
			best := i
			for j := i + 1; j < len(cells); j++ {
				jx, jy := grid.CellCenter(cells[j][0], cells[j][1])
				if !hasClearance(w, cursorX, cursorY, jx, jy, pad) {
					break
				}
				best = j
			}
			bx, by := grid.CellCenter(cells[best][0], cells[best][1])
			e.Path = append(e.Path, geometry.Point{X: bx, Y: by})
			cursorX, cursorY = bx, by
			i = best
		}

		e.PathIndex = 0
		e.PathGoalCell = goalCell
		e.PathCooldown = 15
	}

	if len(e.Path) == 0 {
		return
	}

	// Если текущий waypoint устарел и от нас до него нет clearance —
	// пропускаем его, не дожидаясь stuck-детектора.
	for e.PathIndex < len(e.Path) {
		target := e.Path[e.PathIndex]
		if hasClearance(w, ecx, ecy, target.X, target.Y, bodyHalf) {
			break
		}
		e.PathIndex++
	}
	if e.PathIndex >= len(e.Path) {
		// Все waypoint'ы недостижимы — пересчитываем путь.
		e.Path = nil
		e.PathIndex = 0
		e.PathCooldown = 0
		return
	}

	target := e.Path[e.PathIndex]

	// Достижение waypoint по клетке.
	tcx, tcy := grid.WorldToCell(target.X, target.Y)
	mcx, mcy := grid.WorldToCell(ecx, ecy)
	if tcx == mcx && tcy == mcy {
		e.PathIndex++
		if e.PathIndex >= len(e.Path) {
			e.Path = nil
			e.PathIndex = 0
		}
		return
	}

	// Движение к waypoint.
	dx := target.X - ecx
	dy := target.Y - ecy
	l := math.Hypot(dx, dy)
	if l < 0.5 {
		return
	}
	dx /= l
	dy /= l
	speed := e.Speed()
	moveEnemyX(w, e, dx*speed)
	moveEnemyY(w, e, dy*speed)
}

func moveEnemyX(w *world.World, e *world.Enemy, dx float64) bool {
	if dx == 0 {
		return false
	}
	size := e.Type.Stats().Size
	targetX := e.X + dx
	blocked := false

	for _, wall := range w.Walls {
		if e.Y >= wall.Y+wall.H || e.Y+size <= wall.Y {
			continue
		}
		if dx > 0 {
			if wall.X < e.X+size {
				continue
			}
			if limit := wall.X - size; limit < targetX {
				targetX = limit
				blocked = true
			}
		} else {
			if wall.X+wall.W > e.X {
				continue
			}
			if limit := wall.X + wall.W; limit > targetX {
				targetX = limit
				blocked = true
			}
		}
	}

	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e {
			continue
		}
		otherSize := other.Type.Stats().Size
		if e.Y >= other.Y+otherSize || e.Y+size <= other.Y {
			continue
		}
		if dx > 0 {
			if other.X < e.X+size {
				continue
			}
			if limit := other.X - size; limit < targetX {
				targetX = limit
				blocked = true
			}
		} else {
			if other.X+otherSize > e.X {
				continue
			}
			if limit := other.X + otherSize; limit > targetX {
				targetX = limit
				blocked = true
			}
		}
	}

	if targetX < 0 {
		targetX = 0
		blocked = true
	}
	if targetX+size > world.WorldWidth {
		targetX = world.WorldWidth - size
		blocked = true
	}
	e.X = targetX
	return blocked
}

func moveEnemyY(w *world.World, e *world.Enemy, dy float64) bool {
	if dy == 0 {
		return false
	}
	size := e.Type.Stats().Size
	targetY := e.Y + dy
	blocked := false

	for _, wall := range w.Walls {
		if e.X >= wall.X+wall.W || e.X+size <= wall.X {
			continue
		}
		if dy > 0 {
			if wall.Y < e.Y+size {
				continue
			}
			if limit := wall.Y - size; limit < targetY {
				targetY = limit
				blocked = true
			}
		} else {
			if wall.Y+wall.H > e.Y {
				continue
			}
			if limit := wall.Y + wall.H; limit > targetY {
				targetY = limit
				blocked = true
			}
		}
	}

	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e {
			continue
		}
		otherSize := other.Type.Stats().Size
		if e.X >= other.X+otherSize || e.X+size <= other.X {
			continue
		}
		if dy > 0 {
			if other.Y < e.Y+size {
				continue
			}
			if limit := other.Y - size; limit < targetY {
				targetY = limit
				blocked = true
			}
		} else {
			if other.Y+otherSize > e.Y {
				continue
			}
			if limit := other.Y + otherSize; limit > targetY {
				targetY = limit
				blocked = true
			}
		}
	}

	if targetY < 0 {
		targetY = 0
		blocked = true
	}
	if targetY+size > world.WorldHeight {
		targetY = world.WorldHeight - size
		blocked = true
	}
	e.Y = targetY
	return blocked
}

// depenetrate выталкивает врага из стен по кратчайшей оси,
// если он каким-то образом оказался внутри. Без этого любое
// движение блокируется — враг считает, что станет «ещё хуже».
func depenetrate(w *world.World, e *world.Enemy) {
	size := e.Type.Stats().Size

	// Несколько итераций: выталкивание из одной стены может
	// загнать в другую (в углу). 3 прохода достаточно.
	for iter := 0; iter < 3; iter++ {
		overlapped := false
		for _, wall := range w.Walls {
			if e.X >= wall.X+wall.W || e.X+size <= wall.X ||
				e.Y >= wall.Y+wall.H || e.Y+size <= wall.Y {
				continue
			}
			overlapped = true

			// Глубина проникновения с каждой стороны.
			fromLeft := e.X + size - wall.X    // сколько торчим вправо от левого края
			fromRight := wall.X + wall.W - e.X // сколько торчим влево от правого края
			fromTop := e.Y + size - wall.Y
			fromBottom := wall.Y + wall.H - e.Y

			// Выбираем минимальную — по ней и выталкиваем.
			minPen := fromLeft
			axis := 0
			if fromRight < minPen {
				minPen = fromRight
				axis = 1
			}
			if fromTop < minPen {
				minPen = fromTop
				axis = 2
			}
			if fromBottom < minPen {
				minPen = fromBottom
				axis = 3
			}

			switch axis {
			case 0:
				e.X -= minPen + 0.5
			case 1:
				e.X += minPen + 0.5
			case 2:
				e.Y -= minPen + 0.5
			case 3:
				e.Y += minPen + 0.5
			}
		}
		if !overlapped {
			break
		}
	}

	// Границы мира.
	if e.X < 0 {
		e.X = 0
	}
	if e.Y < 0 {
		e.Y = 0
	}
	if e.X+size > world.WorldWidth {
		e.X = world.WorldWidth - size
	}
	if e.Y+size > world.WorldHeight {
		e.Y = world.WorldHeight - size
	}
}
