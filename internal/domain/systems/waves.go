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
	if ws.Kind == world.WaveKindDefend {
		StepDefendObjectives(w) // проверяет win/lose, обновляет приоритеты
		if !ws.Active {
			return
		}
		stepDefendWave(w)
		return
	}
	stepCommonWave(w)
}

// ==================== COMMON ====================

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

	if ws.SpawnedCount >= ws.WaveSize || ws.GroupsRemaining <= 0 {
		ws.Active = false
		ws.PauseTimer = wavePauseTicks
		return
	}

	startGroup(w)
}

// ==================== TIMER ====================

func stepTimerWave(w *world.World) {
	ws := &w.Wave

	// Таймер истёк — все враги волны умирают.
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

	// Идёт спавн текущей группы.
	if ws.GroupRemaining > 0 {
		if ws.GroupSpawnTimer > 0 {
			ws.GroupSpawnTimer--
			return
		}
		spawnOneTimerGroupMember(w)
		ws.GroupSpawnTimer = 3
		return
	}

	// Группа закончилась — тикаем кулдаун.
	if ws.TimerGroupCooldown > 0 {
		ws.TimerGroupCooldown--
		return
	}

	// Кулдаун истёк — спавним новую группу.
	spawnTimerGroup(w)
}

func spawnTimerGroup(w *world.World) {
	ws := &w.Wave

	minSize, maxSize := world.TimerGroupSizeRangeFor(ws.WaveLevel)
	size := minSize
	if maxSize > minSize {
		size += w.Rng.Intn(maxSize - minSize + 1)
	}

	ws.GroupID++
	ws.GroupRemaining = size
	ws.GroupSpawnTimer = 0
	ws.GroupSupports = 0

	ws.GroupOriginX, ws.GroupOriginY = pickGroupOrigin(w)
}

func spawnOneTimerGroupMember(w *world.World) {
	ws := &w.Wave

	t := pickWaveEnemy(w, ws)

	pos, ok := findFreeSpotNear(w, ws.GroupOriginX, ws.GroupOriginY, 120)
	if !ok {
		ws.GroupSpawnTimer = 3
		return
	}

	e := world.NewEnemy(t, pos.X, pos.Y)
	e.GroupID = ws.GroupID
	e.OriginalTargetSpawn = -1
	e.TargetSpawnIdx = -1
	w.Enemies = append(w.Enemies, e)

	if world.IsSupportType(t) {
		ws.WaveSupports++
		ws.GroupSupports++
	}
	if t == world.EnemySniper {
		ws.WaveSnipers++
	}

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
	case world.EnemyEngineer:
		ws.Composition.Engineer++
	case world.EnemyFlamer:
		ws.Composition.Flamer++
	case world.EnemyProtector:
		ws.Composition.Protector++
	case world.EnemyTank:
		ws.Composition.Tank++
	case world.EnemyMachineGunner:
		ws.Composition.MachineGunner++
	case world.EnemySummoner:
		ws.Composition.Summoner++
	case world.EnemyHarpooner:
		ws.Composition.Harpooner++
	}

	ws.GroupRemaining--
	ws.SpawnedCount++

	// Группа заспавнена — ставим кулдаун до следующей.
	if ws.GroupRemaining == 0 {
		ws.TimerGroupCooldown = world.TimerGroupIntervalFor(ws.WaveLevel)
	}
}

// ==================== ОБЩЕЕ ====================

func startWave(w *world.World) {
	ws := &w.Wave
	ws.Number++
	ws.Composition = world.WaveComposition{}
	ws.WaveLevel = world.RandomWaveLevel(ws.Number, w.Rng)

	ws.SpawnedCount = 0
	ws.KilledCount = 0
	ws.Active = true
	ws.GroupRemaining = 0
	ws.GroupSpawnTimer = 0
	ws.WaveSupports = 0
	ws.WaveSnipers = 0
	ws.MaxWaveSnipers = world.SniperLimitFor(w.Rng)
	ws.GroupSupports = 0
	ws.DefendOutcome = 0
	ws.DefendPenalty = false

	// 1/3 Common, 1/3 Timer, 1/3 Defend.
	switch w.Rng.Intn(3) {
	case 0:
		startCommonWave(w)
	case 1:
		startTimerWave(w)
	default:
		startDefendWave(w)
	}
}

func startCommonWave(w *world.World) {
	ws := &w.Wave
	ws.Kind = world.WaveKindCommon
	ws.WaveSize = world.WaveSizeFor(ws.WaveLevel, ws.BaseWaveSize, w.Rng)
	ws.GroupsTotal = world.GroupCountFor(ws.WaveLevel, w.Rng)
	ws.GroupsRemaining = ws.GroupsTotal
	ws.TimerTotal = 0
	ws.TimerRemaining = 0
	ws.TimerGroupCooldown = 0

	if track := world.TrackForLevel(ws.WaveLevel); track != "" {
		w.PendingTrack = track
	}

	startGroup(w)
}

func startTimerWave(w *world.World) {
	ws := &w.Wave
	ws.Kind = world.WaveKindTimer
	ws.WaveSize = 0
	ws.GroupsTotal = 0
	ws.GroupsRemaining = 0

	minutes := 3 + w.Rng.Intn(3) // 3..5 минут
	ws.TimerTotal = minutes * 60 * 60
	ws.TimerRemaining = ws.TimerTotal
	ws.TimerGroupCooldown = 0 // первая группа сразу

	if track := world.TimerTrackForLevel(ws.WaveLevel); track != "" {
		w.PendingTrack = track
	}

	spawnTimerGroup(w)
}

func startDefendWave(w *world.World) {
	ws := &w.Wave
	ws.Kind = world.WaveKindDefend
	ws.WaveSize = world.WaveSizeFor(ws.WaveLevel, ws.BaseWaveSize, w.Rng)
	ws.GroupsTotal = world.GroupCountFor(ws.WaveLevel, w.Rng)
	ws.GroupsRemaining = ws.GroupsTotal
	ws.TimerTotal = 0
	ws.TimerRemaining = 0
	ws.TimerGroupCooldown = 0

	spawnCount := world.DefendSpawnCountFor(ws.WaveLevel)
	spawnHP := world.DefendSpawnHPFor(ws.WaveLevel)
	spawnSpawnsForWave(w, spawnCount, spawnHP)
	ws.DefendSpawnCount = len(w.SpawnPoints)
	ws.DefendSpawnsAlive = len(w.SpawnPoints)

	// Трек.
	var track string
	track = world.DefendTrackForLevel(ws.WaveLevel)

	if track != "" {
		w.PendingTrack = track
	}

	startGroup(w)
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

	ws.GroupsRemaining--
	ws.GroupID++
	ws.GroupRemaining = size
	ws.GroupSpawnTimer = 0
	ws.GroupSupports = 0

	ws.GroupOriginX, ws.GroupOriginY = pickGroupOrigin(w)
}

func spawnOneFromGroup(w *world.World) {
	ws := &w.Wave

	t := pickWaveEnemy(w, ws)

	pos, ok := findFreeSpotNear(w, ws.GroupOriginX, ws.GroupOriginY, 120)
	if !ok {
		ws.GroupSpawnTimer = 3
		return
	}

	e := world.NewEnemy(t, pos.X, pos.Y)
	e.GroupID = ws.GroupID

	if ws.Kind == world.WaveKindDefend && len(w.SpawnPoints) > 0 {
		if w.Rng.Intn(100) < 25 {
			idx := pickNearestSpawnAt(w, e.X, e.Y)
			e.OriginalTargetSpawn = idx
			e.TargetSpawnIdx = idx
		} else {
			e.OriginalTargetSpawn = -1
			e.TargetSpawnIdx = -1
		}
	} else {
		e.OriginalTargetSpawn = -1
		e.TargetSpawnIdx = -1
	}

	w.Enemies = append(w.Enemies, e)

	if world.IsSupportType(t) {
		ws.WaveSupports++
		ws.GroupSupports++
	}
	if t == world.EnemySniper {
		ws.WaveSnipers++
	}

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
	case world.EnemyEngineer:
		ws.Composition.Engineer++
	case world.EnemyFlamer:
		ws.Composition.Flamer++
	case world.EnemyProtector:
		ws.Composition.Protector++
	case world.EnemyTank:
		ws.Composition.Tank++
	case world.EnemyMachineGunner:
		ws.Composition.MachineGunner++
	case world.EnemySummoner:
		ws.Composition.Summoner++
	case world.EnemyHarpooner:
		ws.Composition.Harpooner++
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
	case world.MapTestEngineer:
		t = world.EnemyEngineer
	case world.MapTestFlamer:
		t = world.EnemyFlamer
	case world.MapTestProtector:
		t = world.EnemyProtector
	case world.MapTestTank:
		t = world.EnemyTank
	case world.MapTestMachineGunner:
		t = world.EnemyMachineGunner
	case world.MapTestSummoner:
		t = world.EnemySummoner
	case world.MapTestHarpooner:
		t = world.EnemyHarpooner
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

func pickNearestSpawnAt(w *world.World, x, y float64) int {
	best := -1
	bestD := math.Inf(1)
	for i := range w.SpawnPoints {
		sp := &w.SpawnPoints[i]
		if sp.IsDead() {
			continue
		}
		scx, scy := sp.Center()
		d := math.Hypot(scx-x, scy-y)
		if d < bestD {
			bestD = d
			best = i
		}
	}
	return best
}

func stepDefendWave(w *world.World) {
	ws := &w.Wave

	// Штрафная группа при уничтожении спавна.
	if !ws.DefendPenalty {
		for i := range w.SpawnPoints {
			if w.SpawnPoints[i].IsDead() {
				ws.DefendPenalty = true
				// 20 дополнительных врагов целевых на спавны.
				spawnPenaltyGroup(w, 20)
				break
			}
		}
	}

	stepCommonWave(w) // базовая логика: спавн групп, паузы
}

func spawnPenaltyGroup(w *world.World, count int) {
	ws := &w.Wave
	for i := 0; i < count; i++ {
		t := pickWaveEnemy(w, ws)
		pos, ok := findFreeSpotNear(w,
			w.Player.X+w.Rng.Float64()*800-400,
			w.Player.Y+w.Rng.Float64()*800-400,
			200)
		if !ok {
			continue
		}
		e := world.NewEnemy(t, pos.X, pos.Y)
		e.GroupID = ws.GroupID
		idx := pickNearestSpawnAt(w, e.X, e.Y)
		e.OriginalTargetSpawn = idx
		e.TargetSpawnIdx = idx
		w.Enemies = append(w.Enemies, e)

		if world.IsSupportType(t) {
			ws.WaveSupports++
		}
		if t == world.EnemySniper {
			ws.WaveSnipers++
		}
		ws.SpawnedCount++
	}
}

// pickWaveEnemy — единая точка выбора врага для любой волны.
// Учитывает лимит поддержки и лимит снайперов.
// Возвращает выбранный тип — БЕЗ инкремента счётчиков,
// это делается вызывающей стороной.
func pickWaveEnemy(w *world.World, ws *world.WaveState) world.EnemyType {
	allowSupport := ws.WaveSupports < world.MaxSupportsPerWave &&
		ws.GroupSupports < world.MaxSupportsPerGroup
	allowSniper := ws.WaveSnipers < ws.MaxWaveSnipers

	for attempt := 0; attempt < 12; attempt++ {
		t := world.RandomEnemyTypeForWave(ws.WaveLevel, allowSupport, w.Rng)
		if t == world.EnemySniper && !allowSniper {
			continue
		}
		return t
	}
	return world.EnemyInfantry
}
