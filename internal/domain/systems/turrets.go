package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

func StepTurrets(w *world.World) {
	p := &w.Player
	pcx, pcy := p.Center()

	for i := range w.Turrets {
		t := &w.Turrets[i]
		if i == 0 {
			println("turret: cd=", t.Cooldown, "cdmax=", t.CooldownMax,
				"burst=", t.ShotsInBurst, "burstsize=", t.BurstSize,
				"dist=", int(math.Hypot(pcx-t.X, pcy-t.Y)),
				"detect=", int(t.DetectRange))
		}
		if t.IsDead() {
			continue
		}

		// Всегда тикаем кулдаун.
		if t.Cooldown > 0 {
			t.Cooldown--
		}

		// Кулдаун ещё не вышел — ничего не делаем.
		if t.Cooldown > 0 {
			continue
		}

		// Проверяем цель.
		if p.IsDead() {
			continue
		}
		dist := math.Hypot(pcx-t.X, pcy-t.Y)
		if dist > t.DetectRange {
			continue
		}
		if !hasLineOfSight(w, t.X, t.Y, pcx, pcy) {
			continue
		}

		// Один выстрел.
		fireTurretShot(w, t, pcx, pcy)

		t.ShotsInBurst++
		if t.ShotsInBurst >= t.BurstSize {
			// Очередь закончена.
			t.ShotsInBurst = 0
			t.Cooldown = t.CooldownMax
		} else {
			// Ждём следующий выстрел очереди.
			t.Cooldown = t.ShotGap
		}
	}

	alive := w.Turrets[:0]
	for _, t := range w.Turrets {
		if !t.IsDead() {
			alive = append(alive, t)
		}
	}
	w.Turrets = alive
}

func fireTurretShot(w *world.World, t *world.Turret, pcx, pcy float64) {
	dx := pcx - t.X
	dy := pcy - t.Y
	l := math.Hypot(dx, dy)
	if l < 0.01 {
		return
	}
	dx /= l
	dy /= l

	w.EnemyProjectiles = append(w.EnemyProjectiles, world.EnemyProjectile{
		X:         t.X + dx*t.Size/2 - t.ProjSize/2,
		Y:         t.Y + dy*t.Size/2 - t.ProjSize/2,
		VX:        dx * t.ProjSpeed,
		VY:        dy * t.ProjSpeed,
		Size:      t.ProjSize,
		Life:      t.ProjLife,
		Damage:    t.ProjDamage,
		DamageMul: 1,
	})
}
