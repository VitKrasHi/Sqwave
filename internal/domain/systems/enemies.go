package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

func StepEnemies(w *world.World) {
	p := &w.Player
	pcx, pcy := p.Center()

	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() {
			continue
		}

		// Отбрасывание — приоритет над ИИ. Враг летит, пока не затухнет.
		if e.KnockbackTimer > 0 {
			moveEnemyX(w, e, e.KnockbackVX)
			moveEnemyY(w, e, e.KnockbackVY)
			e.KnockbackTimer--
			e.KnockbackVX *= 0.85
			e.KnockbackVY *= 0.85
			continue
		}

		stats := e.Type.Stats()
		ecx, ecy := e.Center()
		dx := pcx - ecx
		dy := pcy - ecy
		dist := math.Hypot(dx, dy)

		if dist > 0.01 {
			e.FacingX = dx / dist
			e.FacingY = dy / dist
		}

		// Фаза замаха — враг зафиксирован, меч проворачивается.
		if e.SwingActive {
			e.SwingTimer++
			if e.SwingTimer >= stats.SwingDuration {
				e.SwingActive = false

				// Пересчитываем дистанцию: за время замаха игрок мог отойти.
				ecx, ecy = e.Center()
				pcx, pcy = p.Center()
				dist = math.Hypot(pcx-ecx, pcy-ecy)

				if dist <= stats.MeleeRange && hasLineOfSight(w, ecx, ecy, pcx, pcy) {
					p.TakeDamage(stats.Damage)
				}

				// Кулдаун — случайный, от Min до Max.
				span := stats.AttackCooldownMax - stats.AttackCooldownMin
				e.AttackTimer = stats.AttackCooldownMin + w.Rng.Intn(span+1)
			}
			continue
		}

		if e.AttackTimer > 0 {
			e.AttackTimer--
		}

		// Достаточно близко и кулдаун прошёл — бьём.
		if dist <= stats.MeleeRange && e.AttackTimer == 0 && hasLineOfSight(w, ecx, ecy, pcx, pcy) {
			e.SwingActive = true
			e.SwingTimer = 0
			continue
		}

		// Двигаемся к игроку, если он ещё далеко.
		if dist > stats.MeleeRange {
			speed := e.Speed()
			moveEnemyX(w, e, e.FacingX*speed)
			moveEnemyY(w, e, e.FacingY*speed)
		}
	}

	// Удаляем мёртвых.
	alive := w.Enemies[:0]
	for _, e := range w.Enemies {
		if !e.IsDead() {
			alive = append(alive, e)
		}
	}
	w.Enemies = alive
}

// moveEnemyX двигает врага по X с прижатием к стенам.
// Логика повторяет moveX игрока, но размер берётся из статов врага.
func moveEnemyX(w *world.World, e *world.Enemy, dx float64) {
	if dx == 0 {
		return
	}
	size := e.Type.Stats().Size
	targetX := e.X + dx

	// Стены.
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
			}
		} else {
			if wall.X+wall.W > e.X {
				continue
			}
			if limit := wall.X + wall.W; limit > targetX {
				targetX = limit
			}
		}
	}

	// Другие враги — как стены такого же размера, как их хитбокс.
	// Проверка pointer equality: e указывает внутрь w.Enemies,
	// other == e отсеивает сам себя.
	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e {
			continue
		}
		otherSize := other.Type.Stats().Size
		if e.Y >= other.Y+otherSize || e.Y+size <= other.Y {
			continue
		}
		if dx > 0 {
			if other.X < e.X+size {
				continue
			}
			if limit := other.X - size; limit < targetX {
				targetX = limit
			}
		} else {
			if other.X+otherSize > e.X {
				continue
			}
			if limit := other.X + otherSize; limit > targetX {
				targetX = limit
			}
		}
	}

	if targetX < 0 {
		targetX = 0
	}
	if targetX+size > world.WorldWidth {
		targetX = world.WorldWidth - size
	}
	e.X = targetX
}

func moveEnemyY(w *world.World, e *world.Enemy, dy float64) {
	if dy == 0 {
		return
	}
	size := e.Type.Stats().Size
	targetY := e.Y + dy

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
			}
		} else {
			if wall.Y+wall.H > e.Y {
				continue
			}
			if limit := wall.Y + wall.H; limit > targetY {
				targetY = limit
			}
		}
	}

	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e {
			continue
		}
		otherSize := other.Type.Stats().Size
		if e.X >= other.X+otherSize || e.X+size <= other.X {
			continue
		}
		if dy > 0 {
			if other.Y < e.Y+size {
				continue
			}
			if limit := other.Y - size; limit < targetY {
				targetY = limit
			}
		} else {
			if other.Y+otherSize > e.Y {
				continue
			}
			if limit := other.Y + otherSize; limit > targetY {
				targetY = limit
			}
		}
	}

	if targetY < 0 {
		targetY = 0
	}
	if targetY+size > world.WorldHeight {
		targetY = world.WorldHeight - size
	}
	e.Y = targetY
}

// hasLineOfSight проверяет, что между точками (x1,y1) и (x2,y2)
// нет стен. Используется для melee-удара: если между врагом
// и игроком стена, удар не достаёт.
func hasLineOfSight(w *world.World, x1, y1, x2, y2 float64) bool {
	for _, wall := range w.Walls {
		if _, ok := geometry.RaySegmentIntersectsRect(x1, y1, x2, y2, wall); ok {
			return false
		}
	}
	return true
}
