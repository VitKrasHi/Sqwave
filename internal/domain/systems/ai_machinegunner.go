package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepMachineGunner(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	p := &w.Player

	// --- Поворот оружия к предсказанной позиции игрока ---
	leadTime := dist / stats.ProjectileSpeed
	if leadTime > 30 {
		leadTime = 30
	}
	aimX := pcx + p.VelX*leadTime*stats.MGPredictLead
	aimY := pcy + p.VelY*leadTime*stats.MGPredictLead

	targetAngle := math.Atan2(aimY-ecy, aimX-ecx)
	if !e.WeaponInit {
		e.WeaponAngle = targetAngle
		e.WeaponInit = true
	}
	diff := angleDiff(targetAngle, e.WeaponAngle)
	maxTurn := stats.WeaponTurnRate
	if diff > maxTurn {
		diff = maxTurn
	}
	if diff < -maxTurn {
		diff = -maxTurn
	}
	e.WeaponAngle += diff

	// --- Стрельба ---
	if e.MGCooldownLeft > 0 {
		e.MGCooldownLeft--
	}

	// LOS — по тонкой линии от оружия, без учёта движения.
	los := hasLineOfSight(w, ecx, ecy, pcx, pcy)
	angleFromAim := math.Abs(angleDiff(targetAngle, e.WeaponAngle))
	arcRad := stats.WeaponArcDegrees * math.Pi / 180
	aimed := angleFromAim <= arcRad

	// Очередь в процессе — продолжаем.
	if e.MGBurstLeft > 0 {
		e.MGBurstTimer--
		if e.MGBurstTimer <= 0 {
			// Если во время очереди потеряли LOS — прекращаем.
			if !los {
				e.MGBurstLeft = 0
				e.MGCooldownLeft = stats.MGCooldown
			} else {
				fireMGRound(w, e, ecx, ecy)
				e.MGBurstLeft--
				e.MGBurstTimer = stats.MGBurstGap
				if e.MGBurstLeft == 0 {
					e.MGCooldownLeft = stats.MGCooldown
				}
			}
		}
	} else if los && aimed && e.MGCooldownLeft == 0 &&
		dist <= stats.PreferredMax && !p.IsDead() {
		e.MGBurstLeft = stats.MGBurstCount
		e.MGBurstTimer = 0
	}

	// --- Движение ---
	// ВСЕГДА через moveEnemyToPoint / moveEnemyViaPath — они учитывают
	// стены и A*. Никаких прямых вызовов moveEnemyDirect с координатой
	// за стеной.

	// Если игрок за стеной — обход через A*.
	if !los {
		moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
		return
	}

	// Есть поддержка — встаём между ней и игроком.
	support := pickSupportToCover(w, e, ecx, ecy)
	if support != nil {
		scx, scy := support.Center()
		tx := scx + (pcx-scx)*0.35
		ty := scy + (pcy-scy)*0.35

		dx := tx - ecx
		dy := ty - ecy
		l := math.Hypot(dx, dy)
		if l < 30 {
			return // уже на месте — стоим
		}

		// Проверяем: прямой путь свободен?
		if hasClearance(w, ecx, ecy, tx, ty, stats.Size/2) {
			dx /= l
			dy /= l
			speed := e.Speed()
			moveEnemyX(w, e, dx*speed)
			moveEnemyY(w, e, dy*speed)
		} else {
			moveEnemyToPoint(w, e, ecx, ecy, tx, ty)
		}
		return
	}

	// Без поддержки — держим среднюю дистанцию до игрока.

	// Слишком далеко — приближаемся. Если игрок за стеной (но los есть),
	// путь прямой. Иначе — через A*.
	if dist > stats.PreferredMax*0.85 {
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
		return
	}

	// Слишком близко — пятимся. При отступлении проверяем, что сзади
	// нет стены — иначе двигаемся в сторону.
	if dist < stats.PreferredMin*0.7 {
		awayX := ecx - pcx
		awayY := ecy - pcy
		l := math.Hypot(awayX, awayY)
		if l < 0.01 {
			return
		}
		awayX /= l
		awayY /= l

		// Проверяем, что отступление свободно. Если нет — обходим
		// через A* в точку подальше.
		backX := ecx + awayX*60
		backY := ecy + awayY*60
		if hasClearance(w, ecx, ecy, backX, backY, stats.Size/2) {
			speed := e.Speed() * 0.6
			moveEnemyX(w, e, awayX*speed)
			moveEnemyY(w, e, awayY*speed)
		} else {
			// Идём в сторону — перпендикулярно направлению на игрока.
			perpX := -awayY
			perpY := awayX
			// Выбираем сторону, где есть свобода.
			if !hasClearance(w, ecx, ecy, ecx+perpX*60, ecy+perpY*60, stats.Size/2) {
				perpX = -perpX
				perpY = -perpY
			}
			speed := e.Speed() * 0.6
			moveEnemyX(w, e, perpX*speed)
			moveEnemyY(w, e, perpY*speed)
		}
		return
	}

	// В комфортной зоне — стоим и стреляем.
}

func mgDamageForDist(stats world.EnemyStats, dist float64) int {
	if dist <= stats.PreferredMin {
		return stats.MGMaxDamage
	}
	if dist >= stats.PreferredMax {
		return stats.MGMinDamage
	}
	t := (dist - stats.PreferredMin) / (stats.PreferredMax - stats.PreferredMin)
	return stats.MGMaxDamage + int(float64(stats.MGMinDamage-stats.MGMaxDamage)*t)
}

func fireMGRound(w *world.World, e *world.Enemy, ecx, ecy float64) {
	stats := e.Type.Stats()
	p := &w.Player

	pcx, pcy := p.Center()
	dist := math.Hypot(pcx-ecx, pcy-ecy)
	leadTime := dist / stats.ProjectileSpeed
	if leadTime > 30 {
		leadTime = 30
	}
	aimX := pcx + p.VelX*leadTime*stats.MGPredictLead
	aimY := pcy + p.VelY*leadTime*stats.MGPredictLead

	dx := aimX - ecx
	dy := aimY - ecy
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		return
	}
	baseAngle := math.Atan2(dy, dx)

	spreadRad := stats.MGSpreadDeg * math.Pi / 180
	angle := baseAngle + (w.Rng.Float64()-0.5)*spreadRad

	cos := math.Cos(angle)
	sin := math.Sin(angle)

	damage := mgDamageForDist(stats, dist)
	half := stats.Size / 2

	w.EnemyProjectiles = append(w.EnemyProjectiles, world.EnemyProjectile{
		X:         ecx + cos*half - stats.ProjectileSize/2,
		Y:         ecy + sin*half - stats.ProjectileSize/2,
		VX:        cos * stats.ProjectileSpeed,
		VY:        sin * stats.ProjectileSpeed,
		Size:      stats.ProjectileSize,
		Life:      stats.ProjectileLife,
		Damage:    damage,
		DamageMul: e.DamageMul(),
	})
}
