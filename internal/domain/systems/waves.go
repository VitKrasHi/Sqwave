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
		if ws.PauseTimer > 0 {
			ws.PauseTimer--
			return
		}
		startWave(w)
		return
	}

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

	if ws.TotalPoolRemaining() == 0 {
		ws.Active = false
		ws.PauseTimer = wavePauseTicks
		return
	}

	startGroup(w)
}

func startWave(w *world.World) {
	ws := &w.Wave
	ws.Number++
	size := world.WaveSizeFor(ws.Number, w.Rng)
	comp := world.WaveCompositionFor(ws.Number, size, w.Rng) // ← добавили w.Rng

	ws.Composition = comp
	ws.PoolInfantry = comp.Infantry
	ws.PoolShooter = comp.Shooter
	ws.PoolScout = comp.Scout
	ws.PoolMedic = comp.Medic

	ws.SpawnedCount = 0
	ws.KilledCount = 0
	ws.Active = true
	ws.GroupRemaining = 0
	ws.GroupSpawnTimer = 0

	startGroup(w)
}

func startGroup(w *world.World) {
	ws := &w.Wave

	remaining := ws.TotalPoolRemaining()
	if remaining <= 0 {
		return
	}

	size := 5 + w.Rng.Intn(26)
	if size > remaining {
		size = remaining
	}

	ws.GroupID++
	ws.GroupRemaining = size
	ws.GroupSpawnTimer = 0

	ws.GroupOriginX, ws.GroupOriginY = pickGroupOrigin(w)
}

func spawnOneFromGroup(w *world.World) {
	ws := &w.Wave

	t := pickEnemyType(w, ws)
	if t == -1 {
		ws.GroupRemaining = 0
		return
	}

	pos, ok := findFreeSpotNear(w, ws.GroupOriginX, ws.GroupOriginY, 120)
	if !ok {
		returnTypeToPool(ws, t)
		ws.GroupRemaining--
		return
	}

	e := world.NewEnemy(t, pos.X, pos.Y)
	e.GroupID = ws.GroupID
	w.Enemies = append(w.Enemies, e)

	ws.GroupRemaining--
	ws.SpawnedCount++
}

func pickEnemyType(w *world.World, ws *world.WaveState) world.EnemyType {
	inf := ws.PoolInfantry
	sh := ws.PoolShooter
	sc := ws.PoolScout
	md := ws.PoolMedic
	total := inf + sh + sc + md
	if total == 0 {
		return -1
	}
	pick := w.Rng.Intn(total)
	switch {
	case pick < inf:
		ws.PoolInfantry--
		return world.EnemyInfantry
	case pick < inf+sh:
		ws.PoolShooter--
		return world.EnemyShooter
	case pick < inf+sh+sc:
		ws.PoolScout--
		return world.EnemyScout
	default:
		ws.PoolMedic--
		return world.EnemyMedic
	}
}

func returnTypeToPool(ws *world.WaveState, t world.EnemyType) {
	switch t {
	case world.EnemyInfantry:
		ws.PoolInfantry++
	case world.EnemyShooter:
		ws.PoolShooter++
	case world.EnemyScout:
		ws.PoolScout++
	case world.EnemyMedic:
		ws.PoolMedic++
	}
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

// StartTestRoom — заполняет мир врагами одного типа для теста.
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
