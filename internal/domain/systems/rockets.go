package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

// StepRockets — движение ракет игрока. При столкновении со стеной,
// врагом, турелью, вражеским спавном или по истечении Life — взрыв.
func StepRockets(w *world.World) {
	alive := w.Rockets[:0]
	for _, r := range w.Rockets {
		r.X += r.VX
		r.Y += r.VY
		r.Life--

		rect := r.Rect()

		// Прямое попадание во врага.
		var directHit *world.Enemy
		for i := range w.Enemies {
			e := &w.Enemies[i]
			if e.IsDead() {
				continue
			}
			if rect.Intersects(e.Rect()) {
				directHit = e
				break
			}
		}

		// Прямое попадание в турель (блокирует, но не умирает).
		hitTurret := false
		if directHit == nil {
			for i := range w.Turrets {
				t := &w.Turrets[i]
				if t.IsDead() {
					continue
				}
				if rect.Intersects(t.Rect()) {
					hitTurret = true
					break
				}
			}
		}

		// Прямое попадание во вражеский спавн.
		var hitSpawn *world.SpawnPoint
		if directHit == nil && !hitTurret {
			for i := range w.SpawnPoints {
				sp := &w.SpawnPoints[i]
				if sp.IsDead() || sp.Faction != world.SpawnFactionEnemy {
					continue
				}
				if rect.Intersects(sp.Rect()) {
					hitSpawn = sp
					break
				}
			}
		}

		// Стена.
		hitWall := false
		if directHit == nil && !hitTurret && hitSpawn == nil {
			for _, wall := range w.Walls {
				if rect.Intersects(wall) {
					hitWall = true
					break
				}
			}
		}

		outOfWorld := r.X < 0 || r.X > world.WorldWidth ||
			r.Y < 0 || r.Y > world.WorldHeight

		if directHit == nil && !hitTurret && hitSpawn == nil && !hitWall && !outOfWorld && r.Life > 0 {
			alive = append(alive, r)
			continue
		}

		// --- Взрыв ---
		weapon := r.Weapon.Stats()

		// Прямое попадание во врага.
		if directHit != nil {
			if !directHit.Invulnerable && !BlocksDamageFrom(directHit, r.X, r.Y) {
				DealDamageToEnemy(w, directHit, weapon.DirectHitDamage)
			}
		}

		// Прямое попадание в спавн.
		if hitSpawn != nil && !SpawnIsProtected(w, hitSpawn) {
			hitSpawn.Damage(weapon.DirectHitDamage)
		}

		// AoE-урон по всем в радиусе (кроме прямого получателя).
		applyExplosionDamage(w, r.X, r.Y, r.ExplosionRadius,
			weapon.ExplosionDamage, directHit)

		spawnExplosion(w, r.X, r.Y, r.ExplosionRadius, r.ExplosionLife)
	}

	w.Rockets = alive
}

// applyExplosionDamage бьёт всех врагов, турели и вражеские спавны
// в радиусе, кроме того, кто получил прямое попадание.
func applyExplosionDamage(w *world.World, x, y, radius float64, damage int, exclude *world.Enemy) {
	if damage <= 0 {
		return
	}

	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() || e == exclude {
			continue
		}
		if e.Invulnerable {
			continue
		}
		if BlocksDamageFrom(e, x, y) {
			continue
		}
		cx, cy := e.Center()
		if math.Hypot(cx-x, cy-y) <= radius {
			DealDamageToEnemy(w, e, damage)
		}
	}

	// Турели.
	for i := range w.Turrets {
		t := &w.Turrets[i]
		if t.IsDead() {
			continue
		}
		if math.Hypot(t.X-x, t.Y-y) <= radius {
			// t.Damage(damage) — если когда-нибудь станут уязвимы.
		}
	}

	// Вражеские спавны.
	for i := range w.SpawnPoints {
		sp := &w.SpawnPoints[i]
		if sp.IsDead() || sp.Faction != world.SpawnFactionEnemy {
			continue
		}
		if SpawnIsProtected(w, sp) {
			continue
		}
		scx, scy := sp.Center()
		if math.Hypot(scx-x, scy-y) <= radius {
			sp.Damage(damage)
		}
	}
}

// StepExplosions — тикает таймеры визуальных взрывов.
func StepExplosions(w *world.World) {
	alive := w.Explosions[:0]
	for _, e := range w.Explosions {
		e.Life--
		if e.Life <= 0 {
			continue
		}
		alive = append(alive, e)
	}
	w.Explosions = alive
}

// spawnExplosion создаёт визуальный взрыв.
func spawnExplosion(w *world.World, x, y, radius float64, life int) {
	if radius <= 0 || life <= 0 {
		return
	}
	w.Explosions = append(w.Explosions, world.Explosion{
		X: x, Y: y,
		Radius:  radius,
		Life:    life,
		MaxLife: life,
	})
}
