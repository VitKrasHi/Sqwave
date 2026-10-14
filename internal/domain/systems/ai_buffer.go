package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func stepBuffer(w *world.World, e *world.Enemy, ecx, ecy, pcx, pcy, dist float64) {
	stats := e.Type.Stats()

	// Сбрасываем список целей — пересоберём.
	e.BufferTargets = e.BufferTargets[:0]

	// Собираем до BufferMaxTargets союзников в радиусе.
	type candidate struct {
		idx int
		d   float64
	}
	var cands []candidate

	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e || other.IsDead() {
			continue
		}
		if other.Type == world.EnemyBuffer {
			continue // буферы не усиливают друг друга
		}
		ocx, ocy := other.Center()
		d := math.Hypot(ocx-ecx, ocy-ecy)
		if d > stats.BufferRange {
			continue
		}
		if !hasLineOfSight(w, ecx, ecy, ocx, ocy) {
			continue
		}
		cands = append(cands, candidate{i, d})
	}

	// Сортируем по дистанции и берём ближайших.
	// Простая сортировка — пузырьком, кандидатов обычно мало.
	for i := 1; i < len(cands); i++ {
		for j := i; j > 0 && cands[j-1].d > cands[j].d; j-- {
			cands[j-1], cands[j] = cands[j], cands[j-1]
		}
	}
	max := stats.BufferMaxTargets
	if max > len(cands) {
		max = len(cands)
	}
	for i := 0; i < max; i++ {
		other := &w.Enemies[cands[i].idx]
		other.ApplyBuff(stats.BufferDamageMul, stats.BufferSpeedMul, stats.BufferDuration)
		e.BufferTargets = append(e.BufferTargets, cands[i].idx)
	}

	// --- Движение ---

	// Игрок близко — паника, отходим к ближайшему союзнику.
	if dist < world.MedicPlayerFleeRange*1.5 {
		ally := nearestAlly(w, e, ecx, ecy)
		if ally != nil {
			acx, acy := ally.Center()
			moveEnemyToPoint(w, e, ecx, ecy, acx, acy)
			return
		}
		// Нет союзников — просто убегаем.
		fleeFromPlayer(w, e, ecx, ecy, pcx, pcy)
		return
	}

	// Есть цели — держимся так, чтобы все были в радиусе.
	if len(e.BufferTargets) > 0 {
		// Если все цели в радиусе луча — стоим.
		// Иначе подходим к ближайшей цели.
		// Проверяем: все ли в радиусе?
		allInRange := true
		var farTarget *world.Enemy
		var farDist float64
		for _, idx := range e.BufferTargets {
			t := &w.Enemies[idx]
			tcx, tcy := t.Center()
			d := math.Hypot(tcx-ecx, tcy-ecy)
			if d > stats.BufferRange*0.8 {
				allInRange = false
				if farTarget == nil || d > farDist {
					farTarget = t
					farDist = d
				}
			}
		}
		if allInRange {
			return // стоим и баффаем
		}
		if farTarget != nil {
			tcx, tcy := farTarget.Center()
			moveEnemyToPoint(w, e, ecx, ecy, tcx, tcy)
			return
		}
	}

	// Нет целей — идём к ближайшему союзнику.
	ally := nearestAlly(w, e, ecx, ecy)
	if ally != nil {
		acx, acy := ally.Center()
		moveEnemyToPoint(w, e, ecx, ecy, acx, acy)
		return
	}

	// Совсем никого — стоим.
}

// nearestAlly — ближайший боевой союзник (не медик, не буфер).
// Используется медиком и буфером как «щит» от игрока.
func nearestAlly(w *world.World, e *world.Enemy, ecx, ecy float64) *world.Enemy {
	var nearest *world.Enemy
	nearestDist := math.Inf(1)

	for i := range w.Enemies {
		other := &w.Enemies[i]
		if other == e || other.IsDead() {
			continue
		}
		if other.Type == world.EnemyMedic || other.Type == world.EnemyBuffer {
			continue
		}
		ocx, ocy := other.Center()
		d := math.Hypot(ocx-ecx, ocy-ecy)
		if d < nearestDist {
			nearestDist = d
			nearest = other
		}
	}
	return nearest
}
