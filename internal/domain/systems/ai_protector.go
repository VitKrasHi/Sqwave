package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

// isAttackerType — типы, которые идут прямо на игрока.
// Протектор закрывает их в первую очередь.
func isAttackerType(t world.EnemyType) bool {
	switch t {
	case world.EnemyInfantry, world.EnemyRammer,
		world.EnemyFleer, world.EnemyScout:
		return true
	}
	return false
}

// isPriorityWeakType — типы, которые нужно защищать во вторую очередь.
func isPriorityWeakType(t world.EnemyType) bool {
	switch t {
	case world.EnemySniper, world.EnemyEngineer,
		world.EnemyShooter, world.EnemyMedic,
		world.EnemyBuffer, world.EnemyRocketeer,
		world.EnemyFlamer:
		return true
	}
	return false
}

func stepProtector(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()
	p := &w.Player

	e.ProtectTargets = e.ProtectTargets[:0]

	// --- Выбор целей защиты ---
	// Категория 1: атакующие. Категория 2: слабые.
	// Внутри категории — по дистанции (ближайший).

	type candidate struct {
		idx int
		d   float64
	}
	var attackers []candidate
	var weaks []candidate

	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e || other.IsDead() {
			continue
		}
		if other.Type == world.EnemyProtector {
			continue // протекторы не защищают друг друга
		}
		if other.Invulnerable {
			// Уже защищён — пропускаем, чтобы не тратить слоты.
			continue
		}
		ocx, ocy := other.Center()
		d := math.Hypot(ocx-ecx, ocy-ecy)
		if d > stats.ProtectRange {
			continue
		}
		if !hasLineOfSight(w, ecx, ecy, ocx, ocy) {
			continue
		}

		switch {
		case isAttackerType(other.Type):
			attackers = append(attackers, candidate{i, d})
		case isPriorityWeakType(other.Type):
			weaks = append(weaks, candidate{i, d})
		}
	}

	// Сортировка по дистанции (пузырёк — кандидатов немного).
	sortByDist := func(arr []candidate) {
		for i := 1; i < len(arr); i++ {
			for j := i; j > 0 && arr[j-1].d > arr[j].d; j-- {
				arr[j-1], arr[j] = arr[j], arr[j-1]
			}
		}
	}
	sortByDist(attackers)
	sortByDist(weaks)

	// Собираем итоговый список: сначала атакующие, потом слабые.
	chosen := make([]int, 0, stats.ProtectMaxTargets)
	for _, c := range attackers {
		if len(chosen) >= stats.ProtectMaxTargets {
			break
		}
		chosen = append(chosen, c.idx)
	}
	for _, c := range weaks {
		if len(chosen) >= stats.ProtectMaxTargets {
			break
		}
		chosen = append(chosen, c.idx)
	}

	// Применяем неуязвимость.
	for _, idx := range chosen {
		other := &w.Enemies[idx]
		other.Invulnerable = true
		other.InvulnerableTimer = stats.ProtectDuration
		e.ProtectTargets = append(e.ProtectTargets, idx)
	}

	// --- Движение ---

	// Игрок близко — паника, уходим к ближайшему союзнику (щиту).
	if dist < world.MedicPlayerFleeRange*1.5 {
		ally := nearestAlly(w, e, ecx, ecy)
		if ally != nil {
			acx, acy := ally.Center()
			moveEnemyToPoint(w, e, ecx, ecy, acx, acy)
			return
		}
		fleeFromPlayer(w, e, ecx, ecy, pcx, pcy)
		return
	}

	// Есть цели — двигаемся так, чтобы держать их в радиусе.
	if len(e.ProtectTargets) > 0 {
		// Если самая дальняя цель близко к границе — идём к ней.
		var farTarget *world.Enemy
		var farDist float64
		for _, idx := range e.ProtectTargets {
			t := &w.Enemies[idx]
			tcx, tcy := t.Center()
			d := math.Hypot(tcx-ecx, tcy-ecy)
			if d > farDist {
				farTarget = t
				farDist = d
			}
		}
		// Держим радиус на 80%: если выходим за — идём к дальней цели.
		if farTarget != nil && farDist > stats.ProtectRange*0.8 {
			tcx, tcy := farTarget.Center()
			moveEnemyToPoint(w, e, ecx, ecy, tcx, tcy)
			return
		}
		return // стоим, держим купола
	}

	// Нет целей — идём к ближайшему союзнику.
	ally := nearestAlly(w, e, ecx, ecy)
	if ally != nil {
		acx, acy := ally.Center()
		moveEnemyToPoint(w, e, ecx, ecy, acx, acy)
		return
	}
	_ = p
}
