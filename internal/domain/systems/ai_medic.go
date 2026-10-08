package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepMedic(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	// 1. Держим текущую цель, если она ещё валидна.
	var shield *world.Enemy

	if e.HealTargetIdx >= 0 && e.HealTargetIdx < len(w.Enemies) {
		cur := &w.Enemies[e.HealTargetIdx]
		if !cur.IsDead() && cur != e && cur.Type != world.EnemyMedic {
			// Если где-то появился критически раненый — переключаемся.
			crit := pickCriticalAlly(w, e, ecx, ecy)
			if crit != nil && crit != cur {
				shield = crit
			} else {
				shield = cur
			}
		}
	}

	// 2. Если цели нет — выбираем новую.
	if shield == nil {
		shield = pickShield(w, e, ecx, ecy, pcx, pcy)
	}

	// 3. Совсем никого — бежим от игрока.
	if shield == nil {
		fleeFromPlayer(w, e, ecx, ecy, pcx, pcy)
		return
	}

	e.HealTargetIdx = indexOfEnemy(w, shield)

	// 4. Проверяем, можем ли уже лечить прямо сейчас.
	if MedicCanHeal(w, e, shield) {
		// Стоим на месте — лечение идёт через StepMedicHeal.
		// Не двигаемся, чтобы не терять LOS и не выходить из радиуса.
		return
	}

	// 5. Иначе — двигаемся к позиции за щитом.
	scx, scy := shield.Center()
	behindX, behindY := shieldPosition(w, scx, scy, pcx, pcy)

	dx := behindX - ecx
	dy := behindY - ecy
	behindDist := math.Hypot(dx, dy)

	// Уже на позиции — стоим и ждём, пока лечение дойдёт.
	if behindDist < 12 {
		return
	}

	// Паника при близком игроке — двигаемся быстрее.
	playerDanger := dist < world.MedicPlayerFleeRange*1.5
	speed := e.Speed()
	if playerDanger {
		speed *= 1.5
	}

	// Двигаемся напрямую, если путь свободен, иначе через A*.
	if hasClearance(w, ecx, ecy, behindX, behindY, e.Type.Stats().Size/2) {
		nx := dx / behindDist
		ny := dy / behindDist
		moveEnemyX(w, e, nx*speed)
		moveEnemyY(w, e, ny*speed)
		return
	}
	moveEnemyToPoint(w, e, ecx, ecy, behindX, behindY)
}

// shieldPosition возвращает точку, которая лежит за союзником
// (со стороны, противоположной игроку). Именно там медик
// будет в безопасности: игрок → союзник → медик.
func shieldPosition(w *world.World, scx, scy, pcx, pcy float64) (float64, float64) {
	dx := scx - pcx
	dy := scy - pcy
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		dx, dy, l = 1, 0, 1
	}
	dx /= l
	dy /= l

	dist := 100.0
	targetX := scx + dx*dist
	targetY := scy + dy*dist

	// Если позиция за щитом в стене — сдвигаем к союзнику,
	// пока не выйдем из неё.
	pad := 16.0
	for step := 0; step < 5; step++ {
		if !pointInWallLocal(w, targetX, targetY, pad) {
			return targetX, targetY
		}
		// Сжимаем расстояние к союзнику.
		dist -= 20
		if dist < 30 {
			// Отходим к самому союзнику — он гарантированно в свободной клетке.
			return scx, scy
		}
		targetX = scx + dx*dist
		targetY = scy + dy*dist
	}
	return scx, scy
}

func pointInWallLocal(w *world.World, x, y, pad float64) bool {
	for _, wall := range w.Walls {
		if x+pad > wall.X && x-pad < wall.X+wall.W &&
			y+pad > wall.Y && y-pad < wall.Y+wall.H {
			return true
		}
	}
	return false
}

// pickShield выбирает союзника, за которым прятаться.
// Приоритеты:
//  1. Критически раненый — лечим обязательно.
//  2. Раненый в радиусе луча.
//  3. Ближайший здоровый.
//
// Из кандидатов предпочитаем того, кто «смотрит» в сторону игрока —
// то есть реально может стать щитом.
// pickShield выбирает союзника, за которым прятаться.
// Ищем по всей карте, а не в маленьком радиусе —
// иначе медик паникует, когда рядом никого нет.
func pickShield(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy float64) *world.Enemy {
	var critical *world.Enemy
	var wounded *world.Enemy
	var healthy *world.Enemy

	// Направление от медика к игроку.
	toPlayerX := pcx - ecx
	toPlayerY := pcy - ecy
	toPlayerLen := math.Hypot(toPlayerX, toPlayerY)
	if toPlayerLen > 0.01 {
		toPlayerX /= toPlayerLen
		toPlayerY /= toPlayerLen
	}

	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e || other.IsDead() {
			continue
		}
		if other.Type == world.EnemyMedic {
			continue
		}

		ocx, ocy := other.Center()
		d := math.Hypot(ocx-ecx, ocy-ecy)
		if d > world.MedicAllyRadius {
			continue
		}

		// Приоритет «позади» только для целей лечения в радиусе луча —
		// здоровый союзник подходит даже если он в стороне.
		ratio := other.HPRatio()
		switch {
		case ratio < world.MedicCriticalHPRatio:
			if critical == nil || ratio < critical.HPRatio() {
				critical = other
			}
		case ratio < 1.0 && d < world.MedicHealRange:
			if wounded == nil || ratio < wounded.HPRatio() {
				wounded = other
			}
		default:
			// Здоровый — берём ближайшего, без проверки направления.
			// Он станет «точкой притяжения», даже если не идеальный щит.
			if healthy == nil {
				healthy = other
			} else {
				hcx, hcy := healthy.Center()
				if d < math.Hypot(hcx-ecx, hcy-ecy) {
					healthy = other
				}
			}
		}
	}

	if critical != nil {
		return critical
	}
	if wounded != nil {
		return wounded
	}
	return healthy
}

func distance(a, b *world.Enemy) float64 {
	acx, acy := a.Center()
	bcx, bcy := b.Center()
	return math.Hypot(bcx-acx, bcy-acy)
}

// fleeFromPlayer — союзников нет вообще, бежим от игрока.
// Сначала пробуем найти хоть кого-то на всей карте, чтобы
// использовать его как щит, а не убегать в пустоту.
func fleeFromPlayer(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy float64) {
	// Ищем ближайшего союзника без ограничения радиуса.
	var nearest *world.Enemy
	nearestD := math.Inf(1)
	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e || other.IsDead() || other.Type == world.EnemyMedic {
			continue
		}
		ocx, ocy := other.Center()
		d := math.Hypot(ocx-ecx, ocy-ecy)
		if d < nearestD {
			nearestD = d
			nearest = other
		}
	}

	if nearest != nil {
		// Есть хоть кто-то — идём к нему, а не убегаем.
		ncx, ncy := nearest.Center()
		e.HealTargetIdx = indexOfEnemy(w, nearest)
		moveEnemyToPoint(w, e, ecx, ecy, ncx, ncy)
		return
	}

	// Совсем никого — теперь действительно бежим от игрока.
	awayX := ecx - pcx
	awayY := ecy - pcy
	l := math.Hypot(awayX, awayY)
	if l < 0.01 {
		return
	}
	awayX /= l
	awayY /= l
	speed := e.Speed() * 1.3
	moveEnemyX(w, e, awayX*speed)
	moveEnemyY(w, e, awayY*speed)
}

// indexOfEnemy возвращает индекс врага в слайсе w.Enemies.
// -1, если не найден.
func indexOfEnemy(w *world.World, target *world.Enemy) int {
	for i := range w.Enemies {
		if &w.Enemies[i] == target {
			return i
		}
	}
	return -1
}

// pickCriticalAlly — союзник с HP < 25% в радиусе MedicAllyRadius.
// Если такой есть, медик бросает текущую цель и бежит к нему.
func pickCriticalAlly(w *world.World, e *world.Enemy, ecx, ecy float64) *world.Enemy {
	var best *world.Enemy
	bestRatio := world.MedicCriticalHPRatio

	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e || other.IsDead() {
			continue
		}
		if other.Type == world.EnemyMedic {
			continue
		}
		ratio := other.HPRatio()
		if ratio >= bestRatio {
			continue
		}
		ocx, ocy := other.Center()
		if math.Hypot(ocx-ecx, ocy-ecy) > world.MedicAllyRadius {
			continue
		}
		best = other
		bestRatio = ratio
	}
	return best
}
