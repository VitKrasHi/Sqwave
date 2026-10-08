package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

// maxAStarPerTick — сколько врагов за тик могут пересчитать путь.
// Остальные ждут следующего тика. Ограничивает нагрузку при толпах.
const maxAStarPerTick = 4

func StepEnemies(w *world.World) {
	// 1. Пересобрать spatial hash.
	if w.SpatialHash != nil {
		w.SpatialHash.Clear()
		for i := range w.Enemies {
			e := &w.Enemies[i]
			if e.IsDead() {
				continue
			}
			ecx, ecy := e.Center()
			w.SpatialHash.Insert(i, ecx, ecy)
		}
	}

	// 2. Карта занятости для A* (клетки с врагами).
	if w.EnemyOccupancy != nil {
		for i := range w.EnemyOccupancy {
			w.EnemyOccupancy[i] = false
		}
		for i := range w.Enemies {
			e := &w.Enemies[i]
			if e.IsDead() {
				continue
			}
			size := e.Type.Stats().Size + 8
			markOccupied(w.NavGrid, w.EnemyOccupancy, e.Rect(), size)
		}
	}

	p := &w.Player

	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() {
			continue
		}

		depenetrate(w, e)

		if e.HP < e.LastHP {
			e.RecentlyHitTimer = 60
		}
		e.LastHP = e.HP
		if e.RecentlyHitTimer > 0 {
			e.RecentlyHitTimer--
		}

		if e.KnockbackTimer > 0 {
			moveEnemyX(w, e, e.KnockbackVX)
			moveEnemyY(w, e, e.KnockbackVY)
			e.KnockbackTimer--
			e.KnockbackVX *= 0.85
			e.KnockbackVY *= 0.85
			e.LastX, e.LastY = e.X, e.Y
			continue
		}

		ecx, ecy := e.Center()
		pcx, pcy := p.Center()
		dx := pcx - ecx
		dy := pcy - ecy
		dist := math.Hypot(dx, dy)

		if dist > 0.01 {
			e.FacingX = dx / dist
			e.FacingY = dy / dist
		}

		moved := math.Hypot(e.X-e.LastX, e.Y-e.LastY)
		if moved < 0.4 {
			e.StuckTicks++
		} else {
			e.StuckTicks = 0
		}
		e.LastX, e.LastY = e.X, e.Y

		switch e.Type {
		case world.EnemyInfantry:
			stepInfantry(w, e, ecx, ecy, pcx, pcy, dist)
		case world.EnemyShooter:
			stepShooter(w, e, ecx, ecy, pcx, pcy, dist)
		case world.EnemyScout:
			stepScout(w, e, ecx, ecy, pcx, pcy, dist)
		case world.EnemyMedic:
			stepMedic(w, e, ecx, ecy, pcx, pcy, dist)
		}
	}

	// 3. Расталкивание — после всех движений к цели.
	// Враги уже сдвинулись к игроку; теперь мягко разводим тех,
	// кто наложился.
	applySeparation(w)

	// 4. Удаление мёртвых.
	alive := w.Enemies[:0]
	for _, e := range w.Enemies {
		if !e.IsDead() {
			alive = append(alive, e)
		}
	}
	w.Enemies = alive
}

// ==================== Общие утилиты ====================

func hasLineOfSight(w *world.World, x1, y1, x2, y2 float64) bool {
	for _, wall := range w.Walls {
		if _, ok := geometry.RaySegmentIntersectsRectRaw(x1, y1, x2, y2,
			wall.X, wall.Y, wall.W, wall.H); ok {
			return false
		}
	}
	return true
}

// hasClearance — линия с запасом pad с каждой стороны не пересекает
// ни одну стену. Без аллокаций Rect.
func hasClearance(w *world.World, x1, y1, x2, y2, pad float64) bool {
	for _, wall := range w.Walls {
		if _, ok := geometry.RaySegmentIntersectsRectRaw(x1, y1, x2, y2,
			wall.X-pad, wall.Y-pad, wall.W+pad*2, wall.H+pad*2); ok {
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

	// Если игрок снова виден напрямую — сбрасываем путь,
	// со следующего тика пойдём direct.
	if hasClearance(w, ecx, ecy, pcx, pcy, bodyHalf+world.AgentPredictPad) {
		e.Path = nil
		e.PathIndex = 0
		e.PathCooldown = 0
		return
	}

	// Если cooldown ещё тикает — не строим путь, просто идём
	// по уже проложенному. Это ключевая защита от повторных A*.
	if e.PathCooldown > 0 {
		e.PathCooldown--
		if len(e.Path) > 0 && e.PathIndex < len(e.Path) {
			target := e.Path[e.PathIndex]
			dx := target.X - ecx
			dy := target.Y - ecy
			l := math.Hypot(dx, dy)
			if l > 0.5 {
				dx /= l
				dy /= l
				speed := e.Speed()
				moveEnemyX(w, e, dx*speed)
				moveEnemyY(w, e, dy*speed)
			}
		}
		return
	}

	// --- Построение нового пути ---

	goalCellX, goalCellY := grid.WorldToCell(pcx, pcy)
	if gx, gy, ok := grid.NearestFree(goalCellX, goalCellY, 3); ok {
		goalCellX, goalCellY = gx, gy
	} else {
		return
	}

	startCellX, startCellY := grid.WorldToCell(ecx, ecy)
	if sx, sy, ok := grid.NearestFree(startCellX, startCellY, 3); ok {
		startCellX, startCellY = sx, sy
	} else {
		return
	}

	cells := grid.FindPathAvoid(startCellX, startCellY, goalCellX, goalCellY, w.EnemyOccupancy)
	if cells == nil {
		// Fallback без учёта врагов — иначе в толпе никто не
		// найдёт путь и все замрут.
		cells = grid.FindPath(startCellX, startCellY, goalCellX, goalCellY)
	}
	if cells == nil {
		// Пути нет вообще. Долгий cooldown, чтобы не долбить A*
		// каждый тик. StuckTicks сам переключит на direct,
		// если совсем плохо.
		e.Path = nil
		e.PathIndex = 0
		e.PathCooldown = 60 + w.Rng.Intn(31)
		e.PathGoalCell = [2]int{-1, -1}
		return
	}

	// Сглаживание с ограниченным lookahead: проверяем не больше
	// 8 waypoint'ов вперёд, а не весь путь. Без этого O(N²)
	// на длинных путях убивает производительность.
	pad := bodyHalf + world.AgentClearance
	const maxLookahead = 8

	e.Path = e.Path[:0]
	cursorX, cursorY := ecx, ecy
	for i := 0; i < len(cells); i++ {
		best := i
		limit := i + maxLookahead
		if limit >= len(cells) {
			limit = len(cells) - 1
		}
		for j := i + 1; j <= limit; j++ {
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
	e.PathGoalCell = [2]int{goalCellX, goalCellY}
	e.PathCooldown = 30 + w.Rng.Intn(31) // 0.5..1 сек до следующего пересчёта

	if len(e.Path) == 0 {
		return
	}

	// --- Движение к первому waypoint ---

	// Если до первого waypoint нет clearance — идём к нему
	// «вслепую». Физика сама остановит, если упрёмся в стену,
	// но враг хотя бы сдвинется из мёртвой точки.
	target := e.Path[e.PathIndex]
	if !hasClearance(w, ecx, ecy, target.X, target.Y, bodyHalf) {
		dx := target.X - ecx
		dy := target.Y - ecy
		l := math.Hypot(dx, dy)
		if l > 0.5 {
			dx /= l
			dy /= l
			speed := e.Speed()
			moveEnemyX(w, e, dx*speed)
			moveEnemyY(w, e, dy*speed)
		}
		return
	}

	// Стандартное движение к waypoint'у.
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

// moveEnemyX — движение по X с коллизиями ТОЛЬКО против стен.
// Другие враги игнорируются — их разведёт applySeparation.
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

// depenetrate выталкивает врага из стен по кратчайшей оси.
func depenetrate(w *world.World, e *world.Enemy) {
	size := e.Type.Stats().Size

	for iter := 0; iter < 5; iter++ {
		overlapped := false
		for _, wall := range w.Walls {
			if e.X >= wall.X+wall.W || e.X+size <= wall.X ||
				e.Y >= wall.Y+wall.H || e.Y+size <= wall.Y {
				continue
			}
			overlapped = true

			fromLeft := e.X + size - wall.X
			fromRight := wall.X + wall.W - e.X
			fromTop := e.Y + size - wall.Y
			fromBottom := wall.Y + wall.H - e.Y

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

// pushOutOfCorner — короткий импульс от ближайшей стены.
// Используется, когда враг застрял в углу и никакой путь
// не может его сдвинуть.
func pushOutOfCorner(w *world.World, e *world.Enemy) {
	size := e.Type.Stats().Size
	ecx, ecy := e.Center()

	// Ищем ближайшую стену, которая пересекается с хитбоксом,
	// расширенным на 4 пикселя (то есть враг почти касается).
	var nearWall *geometry.Rect
	nearDist := math.Inf(1)

	for i := range w.Walls {
		wall := &w.Walls[i]
		expanded := geometry.Rect{
			X: wall.X - 4,
			Y: wall.Y - 4,
			W: wall.W + 8,
			H: wall.H + 8,
		}
		if !expanded.Intersects(e.Rect()) {
			continue
		}
		wcx := wall.X + wall.W/2
		wcy := wall.Y + wall.H/2
		d := math.Hypot(wcx-ecx, wcy-ecy)
		if d < nearDist {
			nearDist = d
			nearWall = wall
		}
	}

	if nearWall == nil {
		return
	}

	// Вектор от центра стены к центру врага — куда толкать.
	dx := ecx - (nearWall.X + nearWall.W/2)
	dy := ecy - (nearWall.Y + nearWall.H/2)
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		// Враг ровно в центре стены — это баг, выбираем
		// случайное направление.
		angle := w.Rng.Float64() * math.Pi * 2
		dx = math.Cos(angle)
		dy = math.Sin(angle)
	} else {
		dx /= l
		dy /= l
	}

	// Импульс на 4 тика.
	force := e.Speed() * 2.0
	moveEnemyX(w, e, dx*force)
	moveEnemyY(w, e, dy*force)

	// Заодно сбрасываем path, чтобы пересчитался из новой позиции.
	e.Path = nil
	e.PathIndex = 0
	e.PathCooldown = 0
	_ = size
}

// markOccupied помечает в карте занятости клетки, которые
// перекрываются с прямоугольником r, расширенным на размер агента.
// Реализация совпадает с MarkRectForAgent, но пишет в отдельный срез.
func markOccupied(g *geometry.Grid, occ []bool, r geometry.Rect, agentSize float64) {
	half := agentSize / 2
	x0 := int((r.X - half) / g.CellSize)
	y0 := int((r.Y - half) / g.CellSize)
	x1 := int((r.X + r.W + half) / g.CellSize)
	y1 := int((r.Y + r.H + half) / g.CellSize)

	for cy := y0; cy <= y1; cy++ {
		for cx := x0; cx <= x1; cx++ {
			if !g.InBounds(cx, cy) {
				continue
			}
			centerX, centerY := g.CellCenter(cx, cy)
			agent := geometry.Rect{
				X: centerX - half,
				Y: centerY - half,
				W: agentSize,
				H: agentSize,
			}
			if agent.Intersects(r) {
				occ[cy*g.Cols+cx] = true
			}
		}
	}
}
