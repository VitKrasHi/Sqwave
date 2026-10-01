package world

import (
	"math/rand"
	"time"

	"Sqwave/internal/domain/geometry"
)

type World struct {
	Player           Player
	Enemies          []Enemy
	Bullets          []Bullet
	Rockets          []Rocket
	Explosions       []Explosion
	EnemyProjectiles []EnemyProjectile
	Walls            []geometry.Rect
	SpawnZone        geometry.Rect
	Rng              *rand.Rand
	NavGrid          *geometry.Grid
	Wave             WaveState
	SpatialHash      *SpatialHash
	EnemyOccupancy   []bool
}

type MapKind int

const (
	MapDefault MapKind = iota
	MapTestInfantry
	MapTestShooter
	MapTestScout
)

func New() *World {
	return NewWithSeed(time.Now().UnixNano())
}

// NewWithSeed — для детерминированных тестов и реплеев.
func NewWithSeed(seed int64) *World {
	return NewWithMap(seed, MapDefault)
}

func NewWithMap(seed int64, kind MapKind) *World {
	spawnSize := SpawnZoneSize

	playerX := WorldWidth/2 - PlayerSize/2
	playerY := WorldHeight/2 - PlayerSize/2
	if kind != MapDefault {
		playerX = 700 - PlayerSize/2
		playerY = 450 - PlayerSize/2
	}

	spawnCX := WorldWidth / 2
	spawnCY := WorldHeight / 2
	if kind != MapDefault {
		spawnCX = 700
		spawnCY = 450
	}

	w := &World{
		Player: Player{
			X:         playerX,
			Y:         playerY,
			Weapon:    WeaponNone,
			Secondary: WeaponNone,
		},
		Walls: wallsForMap(kind),
		SpawnZone: geometry.Rect{
			X: spawnCX - spawnSize/2,
			Y: spawnCY - spawnSize/2,
			W: spawnSize,
			H: spawnSize,
		},
		Rng: rand.New(rand.NewSource(seed)),
	}
	w.NavGrid = buildNavGrid(w.Walls)
	w.EnemyOccupancy = make([]bool, w.NavGrid.Cols*w.NavGrid.Rows)
	w.SpatialHash = NewSpatialHash(WorldWidth, WorldHeight)

	switch kind {
	case MapTestInfantry, MapTestShooter, MapTestScout:
		// Волны не запускаются — врагов расставляет StartTestRoom.
		w.Wave = WaveState{SpawnInterval: 20}
	default:
		w.Wave = WaveState{
			Number:          0,
			SpawnInterval:   8,
			GroupSize:       5,
			GroupPauseTicks: 300,
		}
	}
	return w
}

// PlayerInSpawnZone — центр игрока внутри прямоугольника зоны спавна.
func (w *World) PlayerInSpawnZone() bool {
	cx, cy := w.Player.Center()
	return cx >= w.SpawnZone.X && cx <= w.SpawnZone.X+w.SpawnZone.W &&
		cy >= w.SpawnZone.Y && cy <= w.SpawnZone.Y+w.SpawnZone.H
}

func buildNavGrid(walls []geometry.Rect) *geometry.Grid {
	g := geometry.NewGrid(WorldWidth, WorldHeight, NavCellSize)
	agentSize := EnemyInfantry.Stats().Size + AgentClearance
	for _, wall := range walls {
		g.MarkRectForAgent(wall, agentSize)
	}
	return g
}

func wallsForMap(kind MapKind) []geometry.Rect {
	switch kind {
	case MapTestInfantry, MapTestShooter, MapTestScout:
		return testRoomWalls()
	default:
		return defaultWalls()
	}
}

// ==================== ТЕСТОВАЯ КОМНАТА ====================

func testRoomWalls() []geometry.Rect {
	const (
		t = WallThick
		w = 1400.0
		h = 900.0
	)
	return []geometry.Rect{
		// Границы.
		{X: 0, Y: 0, W: w, H: t},
		{X: 0, Y: h - t, W: w, H: t},
		{X: 0, Y: 0, W: t, H: h},
		{X: w - t, Y: 0, W: t, H: h},
		// Колонны по углам и в центре.
		{X: 300, Y: 300, W: 80, H: 80},
		{X: 1020, Y: 300, W: 80, H: 80},
		{X: 300, Y: 520, W: 80, H: 80},
		{X: 1020, Y: 520, W: 80, H: 80},
		{X: 660, Y: 400, W: 80, H: 80},
	}
}

// ==================== ОСНОВНАЯ КАРТА ====================

func defaultWalls() []geometry.Rect {
	t := WallThick
	return []geometry.Rect{
		// ==================== ГРАНИЦЫ МИРА ====================
		{X: 0, Y: 0, W: WorldWidth, H: t},
		{X: 0, Y: WorldHeight - t, W: WorldWidth, H: t},
		{X: 0, Y: 0, W: t, H: WorldHeight},
		{X: WorldWidth - t, Y: 0, W: t, H: WorldHeight},

		// ==================== ЦЕНТРАЛЬНАЯ АРЕНА ====================
		// Прямоугольник 1000..2200 × 900..1500.
		// Верхняя грань с двумя проходами:
		{X: 1000, Y: 900, W: 200, H: t},
		{X: 1350, Y: 900, W: 500, H: t},
		{X: 2000, Y: 900, W: 200, H: t},
		// Нижняя грань:
		{X: 1000, Y: 1500 - t, W: 200, H: t},
		{X: 1350, Y: 1500 - t, W: 500, H: t},
		{X: 2000, Y: 1500 - t, W: 200, H: t},
		// Левая грань:
		{X: 1000, Y: 900, W: t, H: 200},
		{X: 1000, Y: 1250, W: t, H: 250},
		// Правая грань:
		{X: 2200 - t, Y: 900, W: t, H: 200},
		{X: 2200 - t, Y: 1250, W: t, H: 250},
		// Колонны внутри арены:
		{X: 1150, Y: 1050, W: 70, H: 70},
		{X: 1980, Y: 1050, W: 70, H: 70},
		{X: 1150, Y: 1280, W: 70, H: 70},
		{X: 1980, Y: 1280, W: 70, H: 70},

		// ==================== ВЕРХНЯЯ ЛЕВАЯ КОМНАТА ====================
		{X: 200, Y: 200, W: 400, H: t},
		{X: 750, Y: 200, W: 150, H: t},
		{X: 200, Y: 200, W: t, H: 500},
		{X: 200, Y: 700, W: 250, H: t},
		{X: 600, Y: 700, W: 300, H: t},
		{X: 450, Y: 400, W: 100, H: 100},

		// ==================== ВЕРХНЯЯ ПРАВАЯ КОМНАТА ====================
		{X: 2300, Y: 200, W: 300, H: t},
		{X: 2750, Y: 200, W: 250, H: t},
		{X: 2970 - t, Y: 200, W: t, H: 200},
		{X: 2970 - t, Y: 500, W: t, H: 300},
		{X: 2300, Y: 700, W: 250, H: t},
		{X: 2700, Y: 700, W: 300, H: t},
		{X: 2600, Y: 350, W: t, H: 250},

		// ==================== НИЖНЯЯ ЛЕВАЯ КОМНАТА ====================
		{X: 200, Y: 1600, W: 250, H: t},
		{X: 600, Y: 1600, W: 300, H: t},
		{X: 200, Y: 1600, W: t, H: 200},
		{X: 200, Y: 2000, W: t, H: 200},
		{X: 200, Y: 2170, W: 400, H: t},
		{X: 750, Y: 2170, W: 150, H: t},
		{X: 500, Y: 1800, W: 200, H: t},

		// ==================== НИЖНЯЯ ПРАВАЯ КОМНАТА ====================
		{X: 2300, Y: 1600, W: 300, H: t},
		{X: 2750, Y: 1600, W: 250, H: t},
		{X: 2970 - t, Y: 1600, W: t, H: 200},
		{X: 2970 - t, Y: 2000, W: t, H: 200},
		{X: 2300, Y: 2170, W: 400, H: t},
		{X: 2850, Y: 2170, W: 150, H: t},
		{X: 2450, Y: 1800, W: 300, H: t},
		{X: 2600, Y: 1950, W: t, H: 150},

		// ==================== УКРЫТИЯ В ОТКРЫТЫХ ЗОНАХ ====================
		{X: 1300, Y: 500, W: 300, H: t},
		{X: 1800, Y: 400, W: t, H: 200},
		{X: 2400, Y: 1000, W: 200, H: t},
		{X: 1300, Y: 1900, W: 300, H: t},
		{X: 1800, Y: 1800, W: t, H: 200},
		{X: 2400, Y: 1400, W: 200, H: t},
		{X: 500, Y: 1100, W: 300, H: t},
		{X: 800, Y: 1250, W: t, H: 150},
		{X: 2500, Y: 1100, W: 300, H: t},
		{X: 2400, Y: 1250, W: t, H: 150},
	}
}
