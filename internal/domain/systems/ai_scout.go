package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepScout(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	e.AITimer++
	if e.DodgeCooldown > 0 {
		e.DodgeCooldown--
	}
	if e.DodgeTimer > 0 {
		e.DodgeTimer--
	}
	if e.DodgePause > 0 {
		e.DodgePause--
	}

	// После попадания — побег.
	if e.RecentlyHitTimer > 0 {
		scoutFlee(w, e, ecx, ecy, pcx, pcy, dist)
		return
	}

	// Застряли — сбрасываем серию и идём через A*.
	if e.StuckTicks > 10 && e.PathCooldown == 0 {
		e.DodgeTimer = 0
		e.DodgeQueue = 0
		e.DodgeCooldown = 0
		e.Path = nil
		e.PathIndex = 0
		e.StuckTicks = 0
		moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
		return
	}

	// Стреляем дробью в упор.
	if dist < stats.MeleeRange && e.AttackTimer == 0 &&
		hasLineOfSight(w, ecx, ecy, pcx, pcy) {
		fireEnemyProjectile(w, e, ecx, ecy, pcx, pcy)
		span := stats.AttackCooldownMax - stats.AttackCooldownMin
		e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
	}
	if e.AttackTimer > 0 {
		e.AttackTimer--
	}

	// --- Базовое движение: всегда к игроку ---
	toPlayerX := pcx - ecx
	toPlayerY := pcy - ecy
	toPlayerLen := math.Hypot(toPlayerX, toPlayerY)
	if toPlayerLen < 0.01 {
		return
	}
	baseX := toPlayerX / toPlayerLen
	baseY := toPlayerY / toPlayerLen

	// Лёгкий дрейф.
	drift := math.Sin(float64(e.AITimer)*0.08) * 0.3
	perpX := -baseY
	perpY := baseX

	moveX := baseX + perpX*drift
	moveY := baseY + perpY*drift

	// --- Активное уклонение ---
	speedMult := 1.0
	if e.DodgeTimer > 0 {
		// Рывок: сильная боковая компонента, движение к игроку
		// сохраняется. Отношение примерно 2:1 в пользу бокового.
		moveX += e.DodgeDirX * 2.0
		moveY += e.DodgeDirY * 2.0
		speedMult = 1.8

	} else if e.DodgeQueue > 0 && e.DodgePause == 0 {
		// Серия ещё не кончилась и пауза между рывками прошла.
		// Генерируем следующий рывок.
		dirX, dirY := randomDodgeDirection(w, e, baseX, baseY)
		e.DodgeDirX = dirX
		e.DodgeDirY = dirY
		// Длительность рывка: чередуем короткие и длинные.
		// 5..8 — короткий, 12..20 — длинный.
		if w.Rng.Intn(2) == 0 {
			e.DodgeTimer = 5 + w.Rng.Intn(4) // короткий
		} else {
			e.DodgeTimer = 12 + w.Rng.Intn(9) // длинный
		}
		e.DodgeQueue--
		speedMult = 1.8

	} else if e.DodgeCooldown == 0 {
		// Проверяем угрозы: резкое движение прицела или точное наведение.
		urgency, queued := detectAimThreat(w, e, ecx, ecy, pcx, pcy)

		if urgency {
			// Серия из 2–3 рывков.
			e.DodgeQueue = 2 + w.Rng.Intn(2)
			e.DodgeCooldown = 40 + w.Rng.Intn(20)
			// Первый рывок начнётся со следующего тика.
			// Сразу сгенерируем его параметры.
			dirX, dirY := randomDodgeDirection(w, e, baseX, baseY)
			e.DodgeDirX = dirX
			e.DodgeDirY = dirY
			if w.Rng.Intn(2) == 0 {
				e.DodgeTimer = 5 + w.Rng.Intn(4)
			} else {
				e.DodgeTimer = 12 + w.Rng.Intn(9)
			}
			e.DodgeQueue--
			speedMult = 1.8
		} else if queued {
			// Прицел очень близко, но игрок не двигает резко.
			// Одиночный короткий рывок.
			dirX, dirY := randomDodgeDirection(w, e, baseX, baseY)
			e.DodgeDirX = dirX
			e.DodgeDirY = dirY
			e.DodgeTimer = 6 + w.Rng.Intn(5)
			e.DodgeCooldown = 25 + w.Rng.Intn(15)
			speedMult = 1.8
		}
	}

	// После окончания рывка в серии — небольшая пауза,
	// чтобы рывки не сливались в один непрерывный.
	if e.DodgeTimer == 0 && e.DodgeQueue > 0 && e.DodgePause == 0 {
		e.DodgePause = 3
	}

	// Нормализуем итоговый вектор.
	l := math.Hypot(moveX, moveY)
	if l < 0.01 {
		return
	}
	moveX /= l
	moveY /= l

	speed := e.Speed() * speedMult

	if hasLineOfSight(w, ecx, ecy, pcx, pcy) {
		moveEnemyX(w, e, moveX*speed)
		moveEnemyY(w, e, moveY*speed)
		return
	}
	moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
}

// detectAimThreat проверяет два источника угрозы:
//  1. Игрок резко двинул прицелом — срочное уклонение (urgency=true).
//  2. Прицел точно наведён на скаута — одиночный рывок (queued=true).
func detectAimThreat(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy float64) (bool, bool) {
	p := &w.Player

	aimDX := p.AimX - pcx
	aimDY := p.AimY - pcy
	aimLen := math.Hypot(aimDX, aimDY)
	if aimLen < 1 {
		return false, false
	}
	aimX := aimDX / aimLen
	aimY := aimDY / aimLen

	toScoutX := ecx - pcx
	toScoutY := ecy - pcy

	proj := toScoutX*aimX + toScoutY*aimY
	if proj < 40 {
		return false, false
	}

	perpX := -aimY
	perpY := aimX
	perpDist := toScoutX*perpX + toScoutY*perpY

	// 1. Резкое движение прицела при любой близости к лучу.
	// Чем ближе линия прицела — тем критичнее.
	distToLine := math.Abs(perpDist)
	if p.AimPlayerSpeed > world.ScoutAimSpeedPanic && distToLine < 60 {
		return true, false
	}

	// 2. Точное наведение: прицел стоит близко к скауту, игрок
	// двигает его медленно, но стреляет или готов стрелять.
	if distToLine < world.ScoutPreciseRadius {
		precise := p.AimPlayerSpeed < world.ScoutPreciseSpeed
		if precise && p.IsFiring {
			return false, true
		}
	}

	return false, false
}

// randomDodgeDirection — рывок в сторону, но никогда не от игрока.
// Угол выбирается случайно в диапазоне ±70° от направления на игрока.
func randomDodgeDirection(w *world.World, e *world.Enemy, baseX, baseY float64) (float64, float64) {
	// Случайный угол от -70° до +70° от направления к игроку.
	maxAngle := 70 * math.Pi / 180
	angle := (w.Rng.Float64()*2 - 1) * maxAngle

	cos := math.Cos(angle)
	sin := math.Sin(angle)

	// Поворачиваем базовый вектор на этот угол.
	dx := baseX*cos - baseY*sin
	dy := baseX*sin + baseY*cos

	// Гарантия: всегда есть положительная компонента к игроку.
	// Если из-за округлений получилось перпендикулярно — доворачиваем.
	dot := dx*baseX + dy*baseY
	if dot < 0.15 {
		// Смешиваем с base, чтобы вернуть движение вперёд.
		dx = dx*0.5 + baseX*0.5
		dy = dy*0.5 + baseY*0.5
		l := math.Hypot(dx, dy)
		if l > 0.01 {
			dx /= l
			dy /= l
		}
	}
	return dx, dy
}

func scoutFlee(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	awayX := ecx - pcx
	awayY := ecy - pcy
	awayLen := math.Hypot(awayX, awayY)
	if awayLen < 0.01 {
		awayX, awayY = 1, 0
	} else {
		awayX /= awayLen
		awayY /= awayLen
	}
	perpX := -awayY
	perpY := awayX

	zig := math.Sin(float64(e.AITimer)*0.25) * 0.7
	dirX := awayX + perpX*zig
	dirY := awayY + perpY*zig
	l := math.Hypot(dirX, dirY)
	if l > 0.01 {
		dirX /= l
		dirY /= l
	}

	if e.AttackTimer == 0 && dist < 300 && hasLineOfSight(w, ecx, ecy, pcx, pcy) {
		fireEnemyProjectile(w, e, ecx, ecy, pcx, pcy)
		span := stats.AttackCooldownMax - stats.AttackCooldownMin
		e.AttackTimer = (stats.AttackCooldownMin + w.Rng.Intn(span+1)) * 2
	}
	if e.AttackTimer > 0 {
		e.AttackTimer--
	}

	speed := e.Speed() * 1.1
	moveEnemyX(w, e, dirX*speed)
	moveEnemyY(w, e, dirY*speed)
}
