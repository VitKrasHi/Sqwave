package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

// stepScout: разведчик — быстрый, агрессивный вблизи, уклоняется
// от прицела игрока, после попадания убегает и стреляет через спину.
func stepScout(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	e.AITimer++
	if e.DodgeCooldown > 0 {
		e.DodgeCooldown--
	}
	if e.DodgeTimer > 0 {
		e.DodgeTimer--
	}

	// После попадания — побег.
	if e.RecentlyHitTimer > 0 {
		scoutFlee(w, e, ecx, ecy, pcx, pcy, dist)
		return
	}

	// Застряли — сбрасываем уклонение и идём через A*.
	if e.StuckTicks > 5 {
		e.DodgeTimer = 0
		e.DodgeCooldown = 0
		e.Path = nil
		e.PathIndex = 0
		e.PathCooldown = 0
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

	// Лёгкий диагональный дрейф — небольшая добавка к движению вперёд.
	drift := math.Sin(float64(e.AITimer)*0.08) * 0.3
	perpX := -baseY
	perpY := baseX

	moveX := baseX + perpX*drift
	moveY := baseY + perpY*drift

	// --- Активное уклонение ---
	speedMult := 1.0
	if e.DodgeTimer > 0 {
		moveX += e.DodgeDirX * 1.5
		moveY += e.DodgeDirY * 1.5
		speedMult = 1.8
	} else if e.DodgeCooldown == 0 {
		if dirX, dirY, ok := dodgeFromAim(w, e, ecx, ecy, pcx, pcy); ok {
			// Проверяем, что в сторону уклонения есть свободное место.
			lookAhead := e.Type.Stats().Size + 20
			checkX := ecx + dirX*lookAhead
			checkY := ecy + dirY*lookAhead
			if !hasClearance(w, ecx, ecy, checkX, checkY, e.Type.Stats().Size/2) {
				// Там стена — уклоняемся в противоположную сторону.
				dirX, dirY = -dirX, -dirY
				checkX = ecx + dirX*lookAhead
				checkY = ecy + dirY*lookAhead
				if !hasClearance(w, ecx, ecy, checkX, checkY, e.Type.Stats().Size/2) {
					// И там стена — не уклоняемся, идём к игроку.
					e.DodgeCooldown = 15
				}
			}
			if e.DodgeCooldown == 0 {
				dot := dirX*baseX + dirY*baseY
				if dot < 0.3 {
					dirX = dirX*0.6 + baseX*0.4
					dirY = dirY*0.6 + baseY*0.4
					l := math.Hypot(dirX, dirY)
					if l > 0.01 {
						dirX /= l
						dirY /= l
					}
				}
				e.DodgeTimer = 10 + w.Rng.Intn(11) // 10..20 тиков
				e.DodgeDirX = dirX
				e.DodgeDirY = dirY
				e.DodgeCooldown = 25 + w.Rng.Intn(15) // 0.4..0.7 сек
			}
		}
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

// dodgeFromAim — если линия прицела игрока проходит близко
// к разведчику, вернуть перпендикулярное направление уклонения.
func dodgeFromAim(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy float64) (float64, float64, bool) {
	p := &w.Player

	aimDX := p.AimX - pcx
	aimDY := p.AimY - pcy
	aimLen := math.Hypot(aimDX, aimDY)
	if aimLen < 1 {
		return 0, 0, false
	}
	aimX := aimDX / aimLen
	aimY := aimDY / aimLen

	toScoutX := ecx - pcx
	toScoutY := ecy - pcy

	// Проекция на линию прицела: если разведчик позади игрока,
	// он не в опасной зоне.
	proj := toScoutX*aimX + toScoutY*aimY
	if proj < 40 {
		return 0, 0, false
	}

	perpX := -aimY
	perpY := aimX
	perpDist := toScoutX*perpX + toScoutY*perpY

	// Порог: чем ближе к линии прицела, тем срочнее уклонение.
	// Если игрок жмёт ЛКМ — порог шире (он явно целится).
	threshold := 28.0
	if p.IsFiring {
		threshold = 42.0
	}
	if math.Abs(perpDist) > threshold {
		return 0, 0, false
	}

	// Сторона уклонения: туда, куда разведчик уже отклонён.
	// Если он ровно на линии — берём случайную.
	side := 1.0
	if perpDist > 0.5 {
		side = 1
	} else if perpDist < -0.5 {
		side = -1
	} else if w.Rng.Float64() < 0.5 {
		side = -1
	}

	return perpX * side, perpY * side, true
}

// scoutFlee — режим побега после попадания: движение от игрока
// с зигзагом и редкими выстрелами через спину.
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

	// Стреляем через спину, пока есть видимость и дистанция разумная.
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
