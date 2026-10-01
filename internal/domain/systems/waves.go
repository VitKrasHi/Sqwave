package systems

import (
	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

const wavePauseTicks = 180 // 3 секунды между волнами

func StepWaves(w *world.World) {
	ws := &w.Wave

	if ws.Active {
		if !ws.AllSpawned {
			stepSpawning(w)
		} else if len(w.Enemies) == 0 {
			// Волна пройдена — пауза перед следующей.
			ws.Active = false
			ws.PauseTimer = wavePauseTicks
		}
		return
	}

	if ws.PauseTimer > 0 {
		ws.PauseTimer--
		return
	}

	startNextWave(w)
}

func stepSpawning(w *world.World) {
	ws := &w.Wave

	if len(ws.Queue) == 0 {
		ws.AllSpawned = true
		return
	}

	// Пауза между группами.
	if ws.GroupPauseTimer > 0 {
		ws.GroupPauseTimer--
		return
	}

	// Интервал внутри группы.
	if ws.SpawnTimer > 0 {
		ws.SpawnTimer--
		return
	}

	// Спавним одного.
	t := ws.Queue[0]
	ws.Queue = ws.Queue[1:]
	pos := pickSpawnPosition(w)
	w.Enemies = append(w.Enemies, world.NewEnemy(t, pos.X, pos.Y))

	ws.GroupRemaining--

	if ws.GroupRemaining <= 0 {
		// Группа кончилась — пауза перед следующей.
		ws.GroupPauseTimer = ws.GroupPauseTicks
		ws.GroupRemaining = ws.GroupSize
	} else {
		ws.SpawnTimer = ws.SpawnInterval
	}

	if len(ws.Queue) == 0 {
		ws.AllSpawned = true
	}
}

func startNextWave(w *world.World) {
	ws := &w.Wave
	ws.Number++
	ws.Composition = world.WaveCompositionFor(ws.Number)
	ws.Queue = world.QueueFor(ws.Composition, w.Rng)
	ws.AllSpawned = false
	ws.Active = true
	ws.SpawnTimer = 0
	ws.GroupPauseTimer = 0
	ws.GroupRemaining = ws.GroupSize
}

func pickSpawnPosition(w *world.World) geometry.Point {
	for attempt := 0; attempt < 8; attempt++ {
		p := w.Rng.Intn(len(world.SpawnPoints))
		pt := world.SpawnPoints[p]
		if spawnPointFree(w, pt.X, pt.Y) {
			return pt
		}
	}
	return world.SpawnPoints[w.Rng.Intn(len(world.SpawnPoints))]
}

func spawnPointFree(w *world.World, x, y float64) bool {
	const minDist = 60
	for _, e := range w.Enemies {
		ecx, ecy := e.Center()
		dx := ecx - x
		dy := ecy - y
		if dx*dx+dy*dy < minDist*minDist {
			return false
		}
	}
	// Не спавним в игрока.
	pcx, pcy := w.Player.Center()
	dx := pcx - x
	dy := pcy - y
	if dx*dx+dy*dy < minDist*minDist {
		return false
	}
	return true
}

func StartTestRoom(w *world.World, kind world.MapKind) {
	var t world.EnemyType
	switch kind {
	case world.MapTestInfantry:
		t = world.EnemyInfantry
	case world.MapTestShooter:
		t = world.EnemyShooter
	case world.MapTestScout:
		t = world.EnemyScout
	default:
		return
	}
	points := []struct{ x, y float64 }{
		{300, 300}, {1000, 300}, {300, 700}, {1000, 700}, {650, 500},
	}
	for _, p := range points {
		w.Enemies = append(w.Enemies, world.NewEnemy(t, p.x, p.y))
	}
}
