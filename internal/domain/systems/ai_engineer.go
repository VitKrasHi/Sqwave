package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

const engineerSetupDistance = 250.0

func stepEngineer(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	e.AITimer++

	// Турель поставлена — боевое поведение.
	if e.HasPlacedTurret {
		stepEngineerCombat(w, e, ecx, ecy, pcx, pcy, dist)
		return
	}

	// --- Фаза строительства ---

	// Отступление после ранения.
	if e.FleeTimer > 0 {
		e.FleeTimer--
		awayX := ecx - pcx
		awayY := ecy - pcy
		l := math.Hypot(awayX, awayY)
		if l > 0.01 {
			awayX /= l
			awayY /= l
			speed := e.Speed() * 1.3
			moveEnemyX(w, e, awayX*speed)
			moveEnemyY(w, e, awayY*speed)
		}
		return
	}

	// Урон прерывает установку. Прогресс TurretSetupTimer НЕ сбрасываем.
	if e.TurretSetupTimer > 0 && e.RecentlyHitTimer > 0 {
		e.FleeTimer = stats.FleeDuration
		return
	}

	// Первая инициализация фазы строительства.
	if e.SetupX == 0 && e.SetupY == 0 && e.MoveTimer == 0 && e.TurretSetupTimer == 0 {
		sx, sy := findSetupSpot(w, e, pcx, pcy)
		e.SetupX = sx
		e.SetupY = sy
		e.TurretSetupTimer = stats.TurretSetupTime
		e.MoveTimer = stats.SetupMoveTime
	}

	// Движемся к точке, пока она далеко и есть время.
	dx := e.SetupX - ecx
	dy := e.SetupY - ecy
	d := math.Hypot(dx, dy)

	if d > 25 && e.MoveTimer > 0 {
		e.MoveTimer--
		if hasClearance(w, ecx, ecy, e.SetupX, e.SetupY, stats.Size/2) {
			dx /= d
			dy /= d
			speed := e.Speed()
			moveEnemyX(w, e, dx*speed)
			moveEnemyY(w, e, dy*speed)
		} else {
			moveEnemyToPoint(w, e, ecx, ecy, e.SetupX, e.SetupY)
		}
		return
	}

	// В точке — ставим турель.
	e.TurretSetupTimer--
	if e.TurretSetupTimer <= 0 {
		placeTurretAt(w, e, ecx, ecy)
		e.HasPlacedTurret = true
		e.SetupX = 0
		e.SetupY = 0
		e.TurretSetupTimer = 0
		e.MoveTimer = 0
	}
}

// findSetupSpot ищет точку в ~250 px от игрока, в стороне
// от его текущего направления прицела.
func findSetupSpot(w *world.World, e *world.Enemy, pcx, pcy float64) (float64, float64) {
	// Направление от игрока к инженеру.
	ecx, ecy := e.Center()
	dx := ecx - pcx
	dy := ecy - pcy
	l := math.Hypot(dx, dy)

	var baseAngle float64
	if l < 1 {
		baseAngle = w.Rng.Float64() * math.Pi * 2
	} else {
		baseAngle = math.Atan2(dy, dx)
	}

	// Пробуем несколько углов вокруг базового.
	for attempt := 0; attempt < 12; attempt++ {
		angle := baseAngle + (w.Rng.Float64()*2-1)*math.Pi/3 // ±60°
		dist := engineerSetupDistance + w.Rng.Float64()*60   // 250..310

		tx := pcx + math.Cos(angle)*dist
		ty := pcy + math.Sin(angle)*dist

		if tx < 60 || tx > world.WorldWidth-60 ||
			ty < 60 || ty > world.WorldHeight-60 {
			continue
		}
		if pointInWall(w, tx, ty, 40) {
			continue
		}
		return tx, ty
	}

	// Fallback — точка вдоль базового направления.
	tx := pcx + math.Cos(baseAngle)*engineerSetupDistance
	ty := pcy + math.Sin(baseAngle)*engineerSetupDistance
	if tx < 60 {
		tx = 60
	}
	if tx > world.WorldWidth-60 {
		tx = world.WorldWidth - 60
	}
	if ty < 60 {
		ty = 60
	}
	if ty > world.WorldHeight-60 {
		ty = world.WorldHeight - 60
	}
	return tx, ty
}

// placeTurretAt ставит турель в заданную точку.
func placeTurretAt(w *world.World, e *world.Enemy, tx, ty float64) {
	stats := e.Type.Stats()

	w.Turrets = append(w.Turrets, world.Turret{
		X:            tx,
		Y:            ty,
		HP:           stats.TurretHP,
		MaxHP:        stats.TurretHP,
		Size:         stats.TurretSize,
		DetectRange:  stats.TurretDetectRange,
		ProjSpeed:    stats.TurretProjSpeed,
		ProjSize:     stats.TurretProjSize,
		ProjLife:     stats.TurretProjLife,
		ProjDamage:   stats.TurretProjDamage,
		Cooldown:     0,
		CooldownMax:  stats.TurretCooldown,
		ShotGap:      stats.TurretShotGap,
		BurstSize:    stats.TurretBurstCount,
		ShotsInBurst: 0,
	})
}

// stepEngineerCombat — обычное поведение после установки турели.
func stepEngineerCombat(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	p := &w.Player

	// Стрельба дробью.
	if e.AttackTimer > 0 {
		e.AttackTimer--
	}
	los := hasLineOfSight(w, ecx, ecy, pcx, pcy)
	if e.AttackTimer == 0 && dist < 400 && los {
		fireEnemyProjectile(w, e, ecx, ecy, pcx, pcy)
		span := stats.AttackCooldownMax - stats.AttackCooldownMin
		e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
	}

	// Позиционирование: за спиной игрока.
	facingX := p.AimX - pcx
	facingY := p.AimY - pcy
	fl := math.Hypot(facingX, facingY)
	if fl < 0.01 {
		facingX, facingY = 1, 0
	} else {
		facingX /= fl
		facingY /= fl
	}

	idealDist := (stats.PreferredMin + stats.PreferredMax) / 2
	targetX := pcx - facingX*idealDist
	targetY := pcy - facingY*idealDist

	dx := targetX - ecx
	dy := targetY - ecy
	l := math.Hypot(dx, dy)

	if l < 30 {
		return
	}

	// Игрок смотрит в нашу сторону — уходим вбок.
	toEngX := ecx - pcx
	toEngY := ecy - pcy
	tl := math.Hypot(toEngX, toEngY)
	if tl > 0.01 {
		toEngX /= tl
		toEngY /= tl
		dot := toEngX*facingX + toEngY*facingY
		if dot > 0.7 {
			perpX := -facingY
			perpY := facingX
			cross := facingX*toEngY - facingY*toEngX
			if cross < 0 {
				perpX = -perpX
				perpY = -perpY
			}
			speed := e.Speed() * 1.2
			moveEnemyX(w, e, perpX*speed)
			moveEnemyY(w, e, perpY*speed)
			return
		}
	}

	if hasClearance(w, ecx, ecy, targetX, targetY, stats.Size/2) {
		dx /= l
		dy /= l
		speed := e.Speed()
		moveEnemyX(w, e, dx*speed)
		moveEnemyY(w, e, dy*speed)
		return
	}
	moveEnemyToPoint(w, e, ecx, ecy, targetX, targetY)
}
