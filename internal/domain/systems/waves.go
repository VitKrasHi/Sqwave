package systems

import (
	"math"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/world"
)

const wavePauseTicks = 240

func StepWaves(w *world.World) {
	ws := &w.Wave

	if !ws.Active {
		if ws.WavesLimit > 0 && ws.Number >= ws.WavesLimit {
			return
		}
		if ws.PauseTimer > 0 {
			ws.PauseTimer--
			return
		}
		startWave(w)
		return
	}

	if ws.Kind == world.WaveKindTimer {
		stepTimerWave(w)
		return
	}

	stepCommonWave(w)
}

// ---------- Common ----------

func stepCommonWave(w *world.World) {
	ws := &w.Wave

	if ws.GroupRemaining > 0 {
		if ws.GroupSpawnTimer > 0 {
			ws.GroupSpawnTimer--
			return
		}
		spawnOneFromGroup(w)
		ws.GroupSpawnTimer = 3
		return
	}

	if groupAlive(w, ws.GroupID) {
		return
	}

	recomputeKilled(w, ws)

	if ws.SpawnedCount >= ws.WaveSize {
		ws.Active = false
		ws.PauseTimer = wavePauseTicks
		return
	}

	startGroup(w)
}

// ---------- Timer ----------

func stepTimerWave(w *world.World) {
	ws := &w.Wave

	// Таймер истёк — все враги умирают.
	if ws.TimerRemaining <= 0 {
		for i := range w.Enemies {
			if w.Enemies[i].GroupID > 0 {
				w.Enemies[i].HP = 0
			}
		}
		ws.Active = false
		ws.PauseTimer = wavePauseTicks
		return
	}

	ws.TimerRemaining--

	// Спавн текущей группы.
	if ws.GroupRemaining > 0 {
		if ws.GroupSpawnTimer > 0 {
			ws.GroupSpawnTimer--
			return
		}
		spawnTimerGroupMember(w)
		ws.GroupSpawnTimer = 3
		return
	}

	// Пауза перед следующей группой.
	if ws.GroupSpawnTimer > 0 {
		ws.GroupSpawnTimer--
		return
	}

	startTimerGroup(w)
}

func timerProgress(ws *world.WaveState) float64 {
	if ws.TimerTotal <= 0 {
		return 0
	}
	elapsed := ws.TimerTotal - ws.TimerRemaining
	p := float64(elapsed) / float64(ws.TimerTotal)
	if p < 0 {
		p = 0
	}
	if p > 1 {
		p = 1
	}
	return p
}

func startTimerGroup(w *world.World) {
	ws := &w.Wave

	// Размер группы: 4..10, к концу чуть крупнее.
	progress := timerProgress(ws)
	base := 4 + int(progress*4)  // 4..8
	size := base + w.Rng.Intn(4) // +0..3

	ws.GroupID++
	ws.GroupRemaining = size
	ws.GroupSpawnTimer = 0

	ws.GroupOriginX, ws.GroupOriginY = pickGroupOrigin(w)

	// Интервал до следующей группы: от 5 сек до 1.5 сек.
	interval := 300 - int(progress*210) // 300..90
	ws.GroupSpawnTimer = interval
}

func spawnTimerGroupMember(w *world.World) {
	ws := &w.Wave

	progress := timerProgress(ws)
	t := world.RandomEnemyTypeTimer(progress, w.Rng)

	pos, ok := findFreeSpotNear(w, ws.GroupOriginX, ws.GroupOriginY, 120)
	if !ok {
		ws.GroupSpawnTimer = 3
		return
	}

	e := world.NewEnemy(t, pos.X, pos.Y)
	e.GroupID = ws.GroupID
	w.Enemies = append(w.Enemies, e)

	switch t {
	case world.EnemyInfantry:
		ws.Composition.Infantry++
	case world.EnemyShooter:
		ws.Composition.Shooter++
	case world.EnemyScout:
		ws.Composition.Scout++
	case world.EnemyMedic:
		ws.Composition.Medic++
	case world.EnemyRammer:
		ws.Composition.Rammer++
	case world.EnemySniper:
		ws.Composition.Sniper++
	case world.EnemyRocketeer:
		ws.Composition.Rocketeer++
	}

	ws.GroupRemaining--
	ws.SpawnedCount++
}

// ---------- Общее ----------

func startWave(w *world.World) {
	ws := &w.Wave
	ws.Number++
	ws.Composition = world.WaveComposition{}

	ws.SpawnedCount = 0
	ws.KilledCount = 0
	ws.Active = true
	ws.GroupRemaining = 0
	ws.GroupSpawnTimer = 0

	// Рандомно: 50/50 common или timer.
	if w.Rng.Intn(2) == 0 {
		startCommonWave(w)
	} else {
		startTimerWave(w)
	}
}

func startCommonWave(w *world.World) {
	ws := &w.Wave
	ws.Kind = world.WaveKindCommon
	ws.WaveSize = world.WaveSizeFor(ws.Number, ws.BaseWaveSize, w.Rng)
	ws.TimerTotal = 0
	ws.TimerRemaining = 0

	lvl := 1
	if ws.MaxWaveLevel > 1 {
		lvl = 1 + w.Rng.Intn(ws.MaxWaveLevel)
	}
	ws.WaveLevel = lvl
	ws.GroupsTotal = lvl
	ws.GroupsRemaining = lvl

	startGroup(w)
}

func startTimerWave(w *world.World) {
	ws := &w.Wave
	ws.Kind = world.WaveKindTimer
	ws.WaveSize = 0
	ws.WaveLevel = 0
	ws.GroupsTotal = 0
	ws.GroupsRemaining = 0

	// 3..5 минут при 60 TPS.
	minutes := 3 + w.Rng.Intn(3)
	ws.TimerTotal = minutes * 60 * 60
	ws.TimerRemaining = ws.TimerTotal

	startTimerGroup(w)
}

func startGroup(w *world.World) {
	ws := &w.Wave

	remaining := ws.WaveSize - ws.SpawnedCount
	if remaining <= 0 || ws.GroupsRemaining <= 0 {
		ws.GroupRemaining = 0
		ws.GroupsRemaining = 0
		return
	}

	base := remaining / ws.GroupsRemaining
	if base < 3 {
		base = 3
	}
	noise := base / 4
	if noise < 1 {
		noise = 1
	}
	size := base + w.Rng.Intn(noise*2+1) - noise
	if size < 3 {
		size = 3
	}
	if size > remaining {
		size = remaining
	}
	if size > 40 {
		size = 40
	}

	ws.GroupsRemaining--
	ws.GroupID++
	ws.GroupRemaining = size
	ws.GroupSpawnTimer = 0

	ws.GroupOriginX, ws.GroupOriginY = pickGroupOrigin(w)
}

func spawnOneFromGroup(w *world.World) {
	ws := &w.Wave

	t := world.RandomEnemyType(w.Rng)

	pos, ok := findFreeSpotNear(w, ws.GroupOriginX, ws.GroupOriginY, 120)
	if !ok {
		ws.GroupSpawnTimer = 3
		return
	}

	e := world.NewEnemy(t, pos.X, pos.Y)
	e.GroupID = ws.GroupID
	w.Enemies = append(w.Enemies, e)

	switch t {
	case world.EnemyInfantry:
		ws.Composition.Infantry++
	case world.EnemyShooter:
		ws.Composition.Shooter++
	case world.EnemyScout:
		ws.Composition.Scout++
	case world.EnemyMedic:
		ws.Composition.Medic++
	case world.EnemyRammer:
		ws.Composition.Rammer++
	case world.EnemySniper:
		ws.Composition.Sniper++
	case world.EnemyRocketeer:
		ws.Composition.Rocketeer++
	case world.EnemyBuffer:
		ws.Composition.Buffer++
	case world.EnemyFleer:
		ws.Composition.Fleer++
	}

	ws.GroupRemaining--
	ws.SpawnedCount++
}

func pickGroupOrigin(w *world.World) (float64, float64) {
	pcx, pcy := w.Player.Center()

	bestX, bestY := 0.0, 0.0
	bestScore := math.Inf(-1)

	for attempt := 0; attempt < 30; attempt++ {
		p := world.SpawnPoints[w.Rng.Intn(len(world.SpawnPoints))]
		x := p.X + (w.Rng.Float64()*2-1)*80
		y := p.Y + (w.Rng.Float64()*2-1)*80

		if pointInWall(w, x, y, 30) {
			continue
		}
		dist := math.Hypot(x-pcx, y-pcy)
		if dist < 500 {
			continue
		}
		if dist > bestScore {
			bestScore = dist
			bestX, bestY = x, y
		}
	}

	if bestScore == math.Inf(-1) {
		p := world.SpawnPoints[w.Rng.Intn(len(world.SpawnPoints))]
		return p.X, p.Y
	}
	return bestX, bestY
}

func findFreeSpotNear(w *world.World, cx, cy, radius float64) (geometry.Point, bool) {
	for attempt := 0; attempt < 15; attempt++ {
		angle := w.Rng.Float64() * math.Pi * 2
		r := math.Sqrt(w.Rng.Float64()) * radius
		x := cx + math.Cos(angle)*r
		y := cy + math.Sin(angle)*r

		if pointInWall(w, x, y, 16) {
			continue
		}
		if !spawnPointFree(w, x, y) {
			continue
		}
		return geometry.Point{X: x, Y: y}, true
	}
	return geometry.Point{}, false
}

func pointInWall(w *world.World, x, y, pad float64) bool {
	for _, wall := range w.Walls {
		if x+pad > wall.X && x-pad < wall.X+wall.W &&
			y+pad > wall.Y && y-pad < wall.Y+wall.H {
			return true
		}
	}
	return false
}

func spawnPointFree(w *world.World, x, y float64) bool {
	const minDist = 40
	for _, e := range w.Enemies {
		ecx, ecy := e.Center()
		dx := ecx - x
		dy := ecy - y
		if dx*dx+dy*dy < minDist*minDist {
			return false
		}
	}
	pcx, pcy := w.Player.Center()
	dx := pcx - x
	dy := pcy - y
	if dx*dx+dy*dy < minDist*minDist {
		return false
	}
	return true
}

func groupAlive(w *world.World, gid int) bool {
	if gid == 0 {
		return false
	}
	for i := range w.Enemies {
		if w.Enemies[i].GroupID == gid {
			return true
		}
	}
	return false
}

func recomputeKilled(w *world.World, ws *world.WaveState) {
	alive := 0
	for i := range w.Enemies {
		if w.Enemies[i].GroupID > 0 {
			alive++
		}
	}
	ws.KilledCount = ws.SpawnedCount - alive
	if ws.KilledCount < 0 {
		ws.KilledCount = 0
	}
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
	case world.MapTestMedic:
		t = world.EnemyMedic
	case world.MapTestRammer:
		t = world.EnemyRammer
	case world.MapTestSniper:
		t = world.EnemySniper
	case world.MapTestRocketeer:
		t = world.EnemyRocketeer
	case world.MapTestBuffer:
		t = world.EnemyBuffer
	case world.MapTestFleer:
		t = world.EnemyFleer
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
