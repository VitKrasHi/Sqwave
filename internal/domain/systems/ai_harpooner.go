package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepHarpooner(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	p := &w.Player

	if e.HarpoonCooldown > 0 {
		e.HarpoonCooldown--
	}

	// Поворот оружия к игроку.
	targetAngle := math.Atan2(pcy-ecy, pcx-ecx)
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

	// Если игрок уже гарпунирован — не стреляем и не двигаемся.
	if p.Harpooned {
		return
	}

	// Замах.
	if e.HarpoonAiming {
		e.HarpoonAimTimer--
		// Обновляем цель каждый тик.
		e.HarpoonAimTargetX = pcx
		e.HarpoonAimTargetY = pcy

		if e.HarpoonAimTimer <= 0 {
			fireHarpoon(w, e, ecx, ecy)
			e.HarpoonAiming = false
			e.HarpoonCooldown = stats.HarpoonCooldown
		}
		return // во время прицеливания стоим
	}

	los := hasLineOfSight(w, ecx, ecy, pcx, pcy)

	// Начинаем прицеливание.
	if los && e.HarpoonCooldown == 0 &&
		dist >= stats.PreferredMin && dist <= stats.PreferredMax &&
		!p.IsDead() {
		e.HarpoonAiming = true
		e.HarpoonAimTimer = stats.HarpoonAimTime
		e.HarpoonAimTargetX = pcx
		e.HarpoonAimTargetY = pcy
		return
	}

	// --- Движение ---
	if !los {
		moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
		return
	}

	if dist < stats.PreferredMin {
		awayX := ecx - pcx
		awayY := ecy - pcy
		l := math.Hypot(awayX, awayY)
		if l > 0.01 {
			awayX /= l
			awayY /= l
			speed := e.Speed() * 1.1
			moveEnemyX(w, e, awayX*speed)
			moveEnemyY(w, e, awayY*speed)
		}
		return
	}

	if dist > stats.PreferredMax {
		moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)
		return
	}
	// В комфортной зоне стоим.
}

func fireHarpoon(w *world.World, e *world.Enemy, ecx, ecy float64) {
	stats := e.Type.Stats()

	// Если цели нет (0,0) или слишком далеко — берём текущую позицию игрока.
	targetX := e.HarpoonAimTargetX
	targetY := e.HarpoonAimTargetY

	if targetX == 0 && targetY == 0 {
		pcx, pcy := w.Player.Center()
		targetX = pcx
		targetY = pcy
	}

	dx := targetX - ecx
	dy := targetY - ecy
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		return
	}
	dx /= l
	dy /= l

	half := stats.Size / 2
	id := w.AllocHarpoonID()
	w.Harpoons = append(w.Harpoons, world.Harpoon{
		ID:       id,
		X:        ecx + dx*half,
		Y:        ecy + dy*half,
		VX:       dx * stats.HarpoonSpeed,
		VY:       dy * stats.HarpoonSpeed,
		Size:     stats.HarpoonSize,
		Life:     stats.HarpoonLife,
		HP:       stats.HarpoonHP,
		MaxHP:    stats.HarpoonHP,
		OwnerUID: e.UniqueID,
	})
}
