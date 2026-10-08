package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

// MedicCanHeal — единая проверка: может ли медик лечить цель прямо сейчас.
// Используется и в системе лечения, и в отрисовке луча, чтобы
// визуально луч был только тогда, когда реально идёт лечение.
func MedicCanHeal(w *world.World, medic, target *world.Enemy) bool {
	if medic == nil || target == nil {
		return false
	}
	if medic.IsDead() || target.IsDead() {
		return false
	}
	if medic == target {
		return false
	}
	if target.Type == world.EnemyMedic {
		return false
	}
	if target.HP >= target.Type.Stats().MaxHP {
		return false // цель уже полностью здорова
	}

	mcx, mcy := medic.Center()
	tcx, tcy := target.Center()
	if math.Hypot(tcx-mcx, tcy-mcy) > world.MedicHealRange {
		return false
	}
	if !hasLineOfSight(w, mcx, mcy, tcx, tcy) {
		return false
	}
	return true
}

func StepMedicHeal(w *world.World) {
	for i := range w.Enemies {
		medic := &w.Enemies[i]
		if medic.IsDead() || medic.Type != world.EnemyMedic {
			continue
		}
		if medic.HealTargetIdx < 0 || medic.HealTargetIdx >= len(w.Enemies) {
			continue
		}
		target := &w.Enemies[medic.HealTargetIdx]
		if target.IsDead() || target == medic {
			medic.HealTargetIdx = -1
			continue
		}

		if !MedicCanHeal(w, medic, target) {
			continue // цель ещё не в зоне лечения — стоим/идём
		}

		medic.HealTick++
		if medic.HealTick%world.MedicHealTickRate != 0 {
			continue
		}

		maxHP := target.Type.Stats().MaxHP
		if target.HP < maxHP {
			target.HP += world.MedicHealPerTick
			if target.HP > maxHP {
				target.HP = maxHP
			}
		}
	}
}

// applyHeal — увеличивает HP с округлением через накопитель.
// Простой вариант: раз в 2 тика +1 HP даёт эффективные 0.5 HP/тик.
// Точнее — через счётчик.
func applyHeal(target *world.Enemy, _ float64) {
	maxHP := target.Type.Stats().MaxHP
	if target.HP >= maxHP {
		return
	}
	target.HP++
	if target.HP > maxHP {
		target.HP = maxHP
	}
}
