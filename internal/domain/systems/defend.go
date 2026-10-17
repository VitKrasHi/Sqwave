package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

const defendPlayerLeashDist = 800.0
const defendSpawnAttackRange = 60.0
const defendSpawnAttackCooldown = 30

// StepDefendObjectives — логика Defend-волны.
// Обновляет приоритеты врагов, атакует спавны, проверяет
// условия победы/поражения.
func StepDefendObjectives(w *world.World) {
	ws := &w.Wave
	if ws.Kind != world.WaveKindDefend || !ws.Active {
		return
	}

	// Считаем живые спавны.
	alive := 0
	for i := range w.SpawnPoints {
		if !w.SpawnPoints[i].IsDead() {
			alive++
		}
	}
	ws.DefendSpawnsAlive = alive

	// Проигрыш.
	if alive == 0 {
		ws.DefendOutcome = -1
		ws.Active = false
		ws.PauseTimer = wavePauseTicks
		return
	}

	// Победа — все враги волны мертвы и все группы выпущены.
	if ws.SpawnedCount >= ws.WaveSize && alive > 0 {
		anyAlive := false
		for i := range w.Enemies {
			if w.Enemies[i].GroupID > 0 && !w.Enemies[i].IsDead() {
				anyAlive = true
				break
			}
		}
		if !anyAlive {
			ws.DefendOutcome = 1
			ws.Active = false
			ws.PauseTimer = wavePauseTicks
			return
		}
	}

	// Далеко ли игрок от ближайшего спавна?
	nd := w.NearestSpawnDist()
	playerFar := nd > defendPlayerLeashDist || nd < 0

	// Обновляем приоритеты.
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() || e.GroupID == 0 {
			continue
		}

		if playerFar {
			// Игрок далеко — все идут на спавн.
			if e.TargetSpawnIdx < 0 {
				e.TargetSpawnIdx = pickNearestSpawn(w, e)
			} else {
				// Цель-спавн могла быть уничтожена.
				if e.TargetSpawnIdx >= len(w.SpawnPoints) ||
					w.SpawnPoints[e.TargetSpawnIdx].IsDead() {
					e.TargetSpawnIdx = pickNearestSpawn(w, e)
				}
			}
		} else {
			// Игрок близко — возвращаем изначальное намерение.
			// 75% идут на игрока (OriginalTargetSpawn = -1),
			// 25% остаются на спавне.
			desired := e.OriginalTargetSpawn

			if desired >= 0 {
				// Проверяем, что спавн ещё жив.
				if desired >= len(w.SpawnPoints) ||
					w.SpawnPoints[desired].IsDead() {
					// Целевой спавн убит — ищем другой или на игрока.
					next := pickNearestSpawn(w, e)
					desired = next // может быть -1
				}
			}
			e.TargetSpawnIdx = desired
		}

		// Атака спавна (если цель — спавн и враг рядом).
		if e.TargetSpawnIdx >= 0 && e.TargetSpawnIdx < len(w.SpawnPoints) {
			sp := &w.SpawnPoints[e.TargetSpawnIdx]
			if !sp.IsDead() {
				ecx, ecy := e.Center()
				scx, scy := sp.Center()
				if math.Hypot(scx-ecx, scy-ecy) <= defendSpawnAttackRange {
					if e.AttackTimer == 0 {
						dmg := int(float64(e.Type.Stats().Damage) * e.DamageMul())
						if dmg < 1 {
							dmg = 10
						}
						sp.Damage(dmg)
						e.AttackTimer = defendSpawnAttackCooldown
					}
				}
			}
		}
	}
}

// pickNearestSpawn — индекс ближайшего живого спавна к врагу. -1 если нет.
func pickNearestSpawn(w *world.World, e *world.Enemy) int {
	ecx, ecy := e.Center()
	best := -1
	bestD := math.Inf(1)
	for i := range w.SpawnPoints {
		sp := &w.SpawnPoints[i]
		if sp.IsDead() {
			continue
		}
		scx, scy := sp.Center()
		d := math.Hypot(scx-ecx, scy-ecy)
		if d < bestD {
			bestD = d
			best = i
		}
	}
	return best
}

// spawnSpawnsForWave — размещает спавны для Defend-волны.
func spawnSpawnsForWave(w *world.World, count int, hp int) {
	w.SpawnPoints = w.SpawnPoints[:0]

	// Первый — там, где зона спавна игрока.
	z := w.SpawnZone
	w.SpawnPoints = append(w.SpawnPoints, world.SpawnPoint{
		X:     z.X + z.W/2,
		Y:     z.Y + z.H/2,
		Size:  60,
		HP:    hp,
		MaxHP: hp,
		Index: 0,
	})

	// Остальные — в случайных местах, далеко от игрока и друг от друга.
	for i := 1; i < count; i++ {
		x, y, ok := pickRandomSpawnLocation(w)
		if !ok {
			continue
		}
		w.SpawnPoints = append(w.SpawnPoints, world.SpawnPoint{
			X:     x,
			Y:     y,
			Size:  60,
			HP:    hp,
			MaxHP: hp,
			Index: i,
		})
	}
}

func pickRandomSpawnLocation(w *world.World) (float64, float64, bool) {
	const minDistFromPlayer = 700.0
	const minDistBetween = 500.0
	const margin = 200.0

	pcx, pcy := w.Player.Center()

	for attempt := 0; attempt < 40; attempt++ {
		x := margin + w.Rng.Float64()*(world.WorldWidth-2*margin)
		y := margin + w.Rng.Float64()*(world.WorldHeight-2*margin)

		if math.Hypot(x-pcx, y-pcy) < minDistFromPlayer {
			continue
		}
		if pointInWall(w, x, y, 60) {
			continue
		}

		ok := true
		for i := range w.SpawnPoints {
			scx, scy := w.SpawnPoints[i].Center()
			if math.Hypot(x-scx, y-scy) < minDistBetween {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		return x, y, true
	}
	return 0, 0, false
}

// StepAttackObjectives — проверяет победу в Attack-волне.
// Победа: все вражеские спавны уничтожены.
func StepAttackObjectives(w *world.World) {
	ws := &w.Wave
	if ws.Kind != world.WaveKindAttack || !ws.Active {
		return
	}

	alive := 0
	for i := range w.SpawnPoints {
		sp := &w.SpawnPoints[i]
		if sp.Faction == world.SpawnFactionEnemy && !sp.IsDead() {
			alive++
		}
	}
	ws.AttackSpawnsAlive = alive

	if alive == 0 {
		ws.AttackOutcome = 1
		ws.Active = false
		ws.PauseTimer = wavePauseTicks

		// Все враги волны погибают.
		for i := range w.Enemies {
			if w.Enemies[i].GroupID > 0 {
				w.Enemies[i].HP = 0
			}
		}
	}
}

// spawnEnemySpawnsForWave — размещает вражеские спавны в случайных местах.
func spawnEnemySpawnsForWave(w *world.World, count int, hp int) {
	w.SpawnPoints = w.SpawnPoints[:0]

	pcx, pcy := w.Player.Center()

	for i := 0; i < count; i++ {
		x, y, ok := pickEnemySpawnLocation(w, pcx, pcy)
		if !ok {
			continue
		}
		w.SpawnPoints = append(w.SpawnPoints, world.SpawnPoint{
			X:       x,
			Y:       y,
			Size:    70,
			HP:      hp,
			MaxHP:   hp,
			Index:   i,
			Faction: world.SpawnFactionEnemy,
		})
	}
}

func pickEnemySpawnLocation(w *world.World, pcx, pcy float64) (float64, float64, bool) {
	const minDistFromPlayer = 600.0
	const minDistBetween = 500.0
	const margin = 300.0

	for attempt := 0; attempt < 50; attempt++ {
		x := margin + w.Rng.Float64()*(world.WorldWidth-2*margin)
		y := margin + w.Rng.Float64()*(world.WorldHeight-2*margin)

		if math.Hypot(x-pcx, y-pcy) < minDistFromPlayer {
			continue
		}
		if pointInWall(w, x, y, 80) {
			continue
		}

		ok := true
		for i := range w.SpawnPoints {
			scx, scy := w.SpawnPoints[i].Center()
			if math.Hypot(x-scx, y-scy) < minDistBetween {
				ok = false
				break
			}
		}
		if !ok {
			continue
		}
		return x, y, true
	}
	return 0, 0, false
}
