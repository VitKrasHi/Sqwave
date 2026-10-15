package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

// BlocksDamageFrom — возвращает true, если атака из точки (fromX, fromY)
// заблокирована щитом танка. Только для танков, только спереди.
func BlocksDamageFrom(e *world.Enemy, fromX, fromY float64) bool {
	if e.Type != world.EnemyTank || e.IsDead() {
		return false
	}
	stats := e.Type.Stats()
	ecx, ecy := e.Center()
	attackAngle := math.Atan2(fromY-ecy, fromX-ecx)
	diff := math.Abs(angleDiff(attackAngle, e.WeaponAngle))
	halfArc := stats.ShieldArcDegrees * math.Pi / 360
	return diff <= halfArc
}

func stepTank(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	p := &w.Player

	if e.HammerCooldown > 0 {
		e.HammerCooldown--
	}

	// --- Поворот: смотрим на игрока, но позиционируемся к поддержке ---
	lookX, lookY := pcx, pcy

	support := pickSupportToCover(w, e, ecx, ecy)
	if support != nil {
		scx, scy := support.Center()
		// Смотрим в точку между игроком и поддержкой.
		// Так щит ориентирован на игрока, а тело — между ними.
		dx := pcx - scx
		dy := pcy - scy
		lookX = scx + dx*0.6
		lookY = scy + dy*0.6
	}

	targetAngle := math.Atan2(lookY-ecy, lookX-ecx)
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

	// --- Молот: замах ---
	if e.HammerState == 1 {
		e.HammerTimer++
		if e.HammerTimer >= stats.HammerWindup {
			// Удар! AoE вокруг танка.
			// Урон не отменяется, если игрок успел отойти во время замаха —
			// урон только если он всё ещё в радиусе на момент удара.
			if !p.IsDead() {
				pcx2, pcy2 := p.Center()
				d := math.Hypot(pcx2-ecx, pcy2-ecy)
				if d <= stats.HammerRange {
					p.TakeDamage(stats.HammerDamage)
				}
			}
			e.HammerState = 0
			e.HammerTimer = 0
			e.HammerCooldown = stats.HammerCooldown
		}
		// В замахе не двигаемся.
		return
	}

	// --- Решение о начале удара ---
	los := hasLineOfSight(w, ecx, ecy, pcx, pcy)
	if los && e.HammerCooldown == 0 && dist <= stats.HammerTriggerRange {
		e.HammerState = 1
		e.HammerTimer = 0
		return
	}

	// --- Движение ---

	// Если есть поддержка — встаём между ней и игроком.
	if support != nil {
		scx, scy := support.Center()
		tx := (pcx + scx) / 2
		ty := (pcy + scy) / 2
		dx := tx - ecx
		dy := ty - ecy
		l := math.Hypot(dx, dy)
		if l > 30 {
			dx /= l
			dy /= l
			speed := e.Speed()
			moveEnemyX(w, e, dx*speed)
			moveEnemyY(w, e, dy*speed)
		}
		return
	}

	// Иначе — прямо к игроку.
	if los {
		moveEnemyDirect(w, e, ecx, ecy, pcx, pcy)
		return
	}
	moveEnemyViaPath(w, e, ecx, ecy, pcx, pcy)
}

// pickSupportToCover — ближайший союзник-поддержка (не атакующий),
// которого нужно прикрывать щитом.
func pickSupportToCover(w *world.World, e *world.Enemy, ecx, ecy float64) *world.Enemy {
	stats := e.Type.Stats()
	var best *world.Enemy
	bestDist := math.Inf(1)

	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e || other.IsDead() {
			continue
		}
		if other.Type == world.EnemyTank {
			continue
		}
		if isAttackerType(other.Type) {
			continue // атакующих прикрывать не надо
		}
		ocx, ocy := other.Center()
		d := math.Hypot(ocx-ecx, ocy-ecy)
		if d > stats.SupportCoverRange {
			continue
		}
		if d < bestDist {
			bestDist = d
			best = other
		}
	}
	return best
}
