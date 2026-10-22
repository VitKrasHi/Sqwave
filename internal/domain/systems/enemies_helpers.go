package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

// advanceToward — общее наступление к точке: прямо, если LOS есть,
// иначе через A*.
func advanceToward(w *world.World, e *world.Enemy, ecx, ecy, tx, ty float64) {
	if hasLineOfSight(w, ecx, ecy, tx, ty) {
		moveEnemyToPointDirect(w, e, ecx, ecy, tx, ty, 1.0)
		return
	}
	moveEnemyToPoint(w, e, ecx, ecy, tx, ty)
}

// moveEnemyToPointDirect — движение по прямой с множителем скорости.
func moveEnemyToPointDirect(w *world.World, e *world.Enemy, ecx, ecy, tx, ty, speedMult float64) {
	dx := tx - ecx
	dy := ty - ecy
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		return
	}
	dx /= l
	dy /= l
	speed := e.Speed() * speedMult
	moveEnemyX(w, e, dx*speed)
	moveEnemyY(w, e, dy*speed)
}

// moveEnemyToPoint — движение к произвольной точке через A*.
func moveEnemyToPoint(w *world.World, e *world.Enemy, ecx, ecy, tx, ty float64) {
	baseGrid := w.NavGrid
	if baseGrid == nil {
		return
	}

	if hasClearance(w, ecx, ecy, tx, ty, e.Type.Stats().Size/2) {
		moveEnemyToPointDirect(w, e, ecx, ecy, tx, ty, 1.0)
		return
	}

	grid := gridWithAgents(baseGrid, w.Enemies, e, e.Type.Stats().Size)

	goalCellX, goalCellY := grid.WorldToCell(tx, ty)
	if gx, gy, ok := grid.NearestFree(goalCellX, goalCellY, 3); ok {
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
		if sx, sy, ok := grid.NearestFree(startCellX, startCellY, 3); ok {
			startCellX, startCellY = sx, sy
		} else {
			return
		}

		cells := grid.FindPathAvoid(startCellX, startCellY, goalCellX, goalCellY, w.EnemyOccupancy)
		if cells == nil {
			cells = grid.FindPath(startCellX, startCellY, goalCellX, goalCellY)
		}
		if cells == nil {
			e.Path = nil
			e.PathIndex = 0
			e.PathCooldown = 30
			e.PathGoalCell = [2]int{-1, -1}
			return
		}

		bodyHalf := e.Type.Stats().Size / 2
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

	target := e.Path[e.PathIndex]
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

func fireEnemyProjectile(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy float64) {
	stats := e.Type.Stats()
	dx := pcx - ecx
	dy := pcy - ecy
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		return
	}
	baseAngle := math.Atan2(dy, dx)

	count := stats.ProjectileCount
	if count <= 0 {
		count = 1
	}
	spreadRad := stats.ProjectileSpread * math.Pi / 180

	half := stats.Size / 2

	for i := 0; i < count; i++ {
		var angle float64
		if count == 1 {
			angle = baseAngle
		} else {
			angle = baseAngle + (w.Rng.Float64()-0.5)*spreadRad
		}
		cos := math.Cos(angle)
		sin := math.Sin(angle)

		w.EnemyProjectiles = append(w.EnemyProjectiles, world.EnemyProjectile{
			X:         ecx + cos*half - stats.ProjectileSize/2,
			Y:         ecy + sin*half - stats.ProjectileSize/2,
			VX:        cos * stats.ProjectileSpeed,
			VY:        sin * stats.ProjectileSpeed,
			Size:      stats.ProjectileSize,
			Life:      stats.ProjectileLife,
			Damage:    stats.Damage,
			DamageMul: e.DamageMul(), // ← добавили
		})
	}
}

// losToRect проверяет, видна ли хоть одна точка прямоугольника r
// из точки (fromX, fromY). Возвращает координаты видимой точки
// и ok. Проверяются центр и 4 угла со сдвигом внутрь.
func losToRect(w *world.World, fromX, fromY float64, r geometry.Rect) (float64, float64, bool) {
	const inset = 3.0
	targets := [][2]float64{
		{r.X + r.W/2, r.Y + r.H/2},
		{r.X + inset, r.Y + inset},
		{r.X + r.W - inset, r.Y + inset},
		{r.X + inset, r.Y + r.H - inset},
		{r.X + r.W - inset, r.Y + r.H - inset},
	}
	for _, t := range targets {
		if hasLineOfSight(w, fromX, fromY, t[0], t[1]) {
			return t[0], t[1], true
		}
	}
	return 0, 0, false
}

func fireEnemyProjectileAt(w *world.World, e *world.Enemy, fromX, fromY, toX, toY float64) {
	stats := e.Type.Stats()
	dx := toX - fromX
	dy := toY - fromY
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		return
	}
	baseAngle := math.Atan2(dy, dx)
	cos0 := math.Cos(baseAngle)
	sin0 := math.Sin(baseAngle)

	// Стартовая точка: чуть впереди стрелка, но не в стене.
	half := stats.Size/2 + stats.ProjectileSize/2
	startX := fromX + cos0*half
	startY := fromY + sin0*half
	if pointInAnyWall(w, startX, startY, stats.ProjectileSize/2) {
		// Не выходит вперёд — стреляем из центра.
		startX = fromX
		startY = fromY
	}

	count := stats.ProjectileCount
	if count <= 0 {
		count = 1
	}
	spreadRad := stats.ProjectileSpread * math.Pi / 180

	for i := 0; i < count; i++ {
		var angle float64
		if count == 1 {
			angle = baseAngle
		} else {
			angle = baseAngle + (w.Rng.Float64()-0.5)*spreadRad
		}
		cos := math.Cos(angle)
		sin := math.Sin(angle)

		w.EnemyProjectiles = append(w.EnemyProjectiles, world.EnemyProjectile{
			X:         startX - stats.ProjectileSize/2,
			Y:         startY - stats.ProjectileSize/2,
			VX:        cos * stats.ProjectileSpeed,
			VY:        sin * stats.ProjectileSpeed,
			Size:      stats.ProjectileSize,
			Life:      stats.ProjectileLife,
			Damage:    stats.Damage,
			DamageMul: e.DamageMul(),
		})
	}
}

// pointInAnyWall — точка с запасом r внутри какой-либо стены?
func pointInAnyWall(w *world.World, x, y, r float64) bool {
	for _, wall := range w.Walls {
		if x+r > wall.X && x-r < wall.X+wall.W &&
			y+r > wall.Y && y-r < wall.Y+wall.H {
			return true
		}
	}
	return false
}

// losToRectWithPad — как losToRect, но линия проверяется
// против стен, расширенных на pad. Учитывает размер снаряда.
func losToRectWithPad(w *world.World, fromX, fromY float64, r geometry.Rect, pad float64) (float64, float64, bool) {
	const inset = 3.0
	targets := [][2]float64{
		{r.X + r.W/2, r.Y + r.H/2},
		{r.X + inset, r.Y + inset},
		{r.X + r.W - inset, r.Y + inset},
		{r.X + inset, r.Y + r.H - inset},
		{r.X + r.W - inset, r.Y + r.H - inset},
	}
	for _, t := range targets {
		if hasClearance(w, fromX, fromY, t[0], t[1], pad) {
			return t[0], t[1], true
		}
	}
	return 0, 0, false
}

// gridWithAgents возвращает копию nav-сетки, где клетки,
// занятые другими врагами, помечены как blocked. Это заставляет
// A* прокладывать путь в обход союзников, а не сквозь них.
// skip — сам агент, для которого строим путь; base-сетка не мутируется.
func gridWithAgents(base *geometry.Grid, enemies []world.Enemy, skip *world.Enemy, agentSize float64) *geometry.Grid {
	if base == nil {
		return nil
	}

	copyGrid := &geometry.Grid{
		CellSize: base.CellSize,
		Cols:     base.Cols,
		Rows:     base.Rows,
		Blocked:  make([]bool, len(base.Blocked)),
	}
	copy(copyGrid.Blocked, base.Blocked)

	for i := range enemies {
		e := &enemies[i]
		if e == skip {
			continue
		}
		if e.IsDead() {
			continue
		}
		// Помечаем клетки вокруг врага с запасом в размер агента,
		// чтобы пути не шли впритык к союзникам.
		copyGrid.MarkRectForAgent(e.Rect(), agentSize)
	}
	return copyGrid
}

// processMeleeAttack — замах (стоит, фиксирует точку) → рывок к точке →
// удар в конусе 75° от направления рывка.
// Рывок имеет фиксированную длительность (LungeDur = 12 тиков),
// скорость подбирается под дистанцию.
// processMeleeAttack — замах (стоит, фиксирует точку) → рывок к точке →
// удар. Попадание проверяется как отрезок МЕЧА против хитбокса игрока,
// плюс страховочная проверка конуса в конце.
func processMeleeAttack(w *world.World, e *world.Enemy, stats world.EnemyStats) bool {
	pcx, pcy := w.Player.Center()
	p := &w.Player

	const lungeDur = 12
	const lungeOvershoot = 60.0
	const hitTolerance = 3.0

	playerHitBox := geometry.Rect{
		X: p.X - hitTolerance,
		Y: p.Y - hitTolerance,
		W: world.PlayerSize + 2*hitTolerance,
		H: world.PlayerSize + 2*hitTolerance,
	}

	// meleeHitByBlade — проверка попадания мечом-отрезком.
	// origin — центр врага, dir — направление меча, length — длина.
	meleeHitByBlade := func(originX, originY, dirX, dirY, length float64) bool {
		if p.IsDead() {
			return false
		}
		if length < 1 {
			return false
		}
		endX := originX + dirX*length
		endY := originY + dirY*length
		_, ok := geometry.RaySegmentIntersectsRect(originX, originY, endX, endY, playerHitBox)
		return ok
	}

	// checkConeHit — проверка конуса от фиксированной точки (для страховки).
	halfConeDeg := 75.0 / 2.0
	cosHalfCone := math.Cos(halfConeDeg * math.Pi / 180)
	checkConeHit := func(originX, originY, dirX, dirY float64) bool {
		d := math.Hypot(pcx-originX, pcy-originY)
		if d > stats.MeleeRange {
			return false
		}
		if d < 1 {
			return true
		}
		toPX := (pcx - originX) / d
		toPY := (pcy - originY) / d
		dot := toPX*dirX + toPY*dirY
		return dot >= cosHalfCone
	}

	dealDamage := func() {
		dmg := int(float64(stats.Damage) * e.DamageMul())
		TryMeleeDamage(w, e, dmg)
		e.LungeHit = true
	}

	// --- Фаза рывка ---
	if e.LungeActive {
		e.LungeTimer++

		moveEnemyX(w, e, e.LungeDirX*e.LungeSpeed)
		moveEnemyY(w, e, e.LungeDirY*e.LungeSpeed)
		e.FacingX = e.LungeDirX
		e.FacingY = e.LungeDirY

		// Проверка мечом — отрезок от центра врага в направлении рывка.
		if !e.LungeHit {
			ecx, ecy := e.Center()
			if meleeHitByBlade(ecx, ecy, e.LungeDirX, e.LungeDirY, stats.MeleeRange) {
				dealDamage()
			}
		}

		// Также проверка AABB тела — если меч задел вплотную.
		if !e.LungeHit && e.Rect().Intersects(playerHitBox) {
			dealDamage()
		}

		if e.LungeTimer >= e.LungeDur {
			// Страховка: конус от SwingTarget.
			if !e.LungeHit && checkConeHit(
				e.SwingTargetX, e.SwingTargetY, e.LungeDirX, e.LungeDirY) {
				dealDamage()
			}

			e.LungeActive = false
			e.LungeTimer = 0
			e.LungeHit = false // сброс для следующей атаки
			span := stats.AttackCooldownMax - stats.AttackCooldownMin
			e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
		}
		return true
	}

	// --- Фаза замаха ---
	if e.SwingActive {
		e.SwingTimer++

		ecx, ecy := e.Center()
		dx := pcx - ecx
		dy := pcy - ecy
		l := math.Hypot(dx, dy)
		if l > 0.01 {
			e.FacingX = dx / l
			e.FacingY = dy / l
		}

		// Визуальная длина меча и угол — те же, что в drawEnemySwing.
		progress := 0.0
		if stats.SwingDuration > 0 {
			progress = float64(e.SwingTimer) / float64(stats.SwingDuration)
		}

		var bladeAngle, bladeLen float64
		if progress < 0.6 {
			t := progress / 0.6
			bladeAngle = math.Atan2(e.FacingY, e.FacingX) - t*1.4
			bladeLen = stats.MeleeRange * (0.5 + 0.5*t)
		} else {
			bladeAngle = math.Atan2(e.FacingY, e.FacingX) - 1.4
			bladeLen = stats.MeleeRange
		}

		// Проверка мечом по той же длине/углу, что рисуется.
		if !e.LungeHit {
			if meleeHitByBlade(ecx, ecy, math.Cos(bladeAngle), math.Sin(bladeAngle), bladeLen) {
				dealDamage()
			}
		}

		if e.SwingTimer >= stats.SwingDuration {
			e.SwingActive = false
			e.SwingTimer = 0

			// Если уже попали во время замаха — кулдаун.
			if e.LungeHit {
				e.LungeHit = false
				span := stats.AttackCooldownMax - stats.AttackCooldownMin
				e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
				return true
			}

			ecx, ecy = e.Center()
			dx = e.SwingTargetX - ecx
			dy = e.SwingTargetY - ecy
			l = math.Hypot(dx, dy)

			if l < 2 {
				// Мгновенный удар.
				if e.Rect().Intersects(playerHitBox) {
					dealDamage()
				} else if checkConeHit(e.SwingTargetX, e.SwingTargetY, e.FacingX, e.FacingY) {
					dealDamage()
				}
				e.LungeHit = false
				span := stats.AttackCooldownMax - stats.AttackCooldownMin
				e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
				return true
			}

			flyLen := l + lungeOvershoot
			e.LungeActive = true
			e.LungeTimer = 0
			e.LungeHit = false
			e.LungeDirX = dx / l
			e.LungeDirY = dy / l
			e.LungeSpeed = flyLen / float64(lungeDur)
			if e.LungeSpeed > 22 {
				e.LungeSpeed = 22
			}
			if e.LungeSpeed < 5 {
				e.LungeSpeed = 5
			}
			e.LungeDur = lungeDur
		}
		return true
	}

	return false
}
