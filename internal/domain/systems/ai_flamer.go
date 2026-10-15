package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepFlamer(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	// Скорость растёт по мере потери HP: от x1 до x2.
	ratio := float64(e.HP) / float64(stats.MaxHP)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}
	e.SpeedScale = 2.0 - ratio

	// --- Оружие медленно поворачивается к игроку ---
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

	// Проверяем: смотрит ли оружие достаточно близко к игроку.
	angleFromAim := math.Abs(angleDiff(targetAngle, e.WeaponAngle))
	arcRad := stats.WeaponArcDegrees * math.Pi / 180
	aimed := angleFromAim <= arcRad

	if e.AttackTimer > 0 {
		e.AttackTimer--
	}

	los := hasLineOfSight(w, ecx, ecy, pcx, pcy)

	// Стреляем только если прицел в конусе и игрок в диапазоне.
	if los && aimed && e.AttackTimer == 0 &&
		dist <= stats.PreferredMax && dist >= stats.PreferredMin {
		fireFlamerBurst(w, e, ecx, ecy)
		span := stats.AttackCooldownMax - stats.AttackCooldownMin
		e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
	}

	// --- Движение ---
	if los {
		if dist > stats.PreferredMax*0.7 {
			moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)
			return
		}
		if dist < stats.PreferredMin*0.6 {
			awayX := ecx - pcx
			awayY := ecy - pcy
			l := math.Hypot(awayX, awayY)
			if l > 0.01 {
				awayX /= l
				awayY /= l
				speed := e.Speed() * 0.5
				moveEnemyX(w, e, awayX*speed)
				moveEnemyY(w, e, awayY*speed)
			}
			return
		}
		return
	}

	moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
}

// fireFlamerBurst стреляет в направлении ОРУЖИЯ, а не игрока.
// Если игрок закрутил огнемётчика — пламя летит мимо.
func fireFlamerBurst(w *world.World, e *world.Enemy, ecx, ecy float64) {
	stats := e.Type.Stats()
	baseAngle := e.WeaponAngle

	count := stats.ProjectileCount
	if count <= 0 {
		count = 1
	}
	spreadRad := stats.ProjectileSpread * math.Pi / 180
	half := stats.Size / 2

	for i := 0; i < count; i++ {
		angle := baseAngle + (w.Rng.Float64()-0.5)*spreadRad
		cos := math.Cos(angle)
		sin := math.Sin(angle)

		w.EnemyProjectiles = append(w.EnemyProjectiles, world.EnemyProjectile{
			X:         ecx + cos*half - stats.ProjectileSize/2,
			Y:         ecy + sin*half - stats.ProjectileSize/2,
			VX:        cos * stats.ProjectileSpeed,
			VY:        sin * stats.ProjectileSpeed,
			Size:      stats.ProjectileSize,
			Life:      stats.ProjectileLife,
			Damage:    stats.FlameDamage,
			DamageMul: e.DamageMul(),
			Flame:     true,
		})
	}
}
