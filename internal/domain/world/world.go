package world

import (
	"math/rand"
	"time"

	"Sqwave/internal/domain/geometry"
)

type World struct {
	Player     Player
	Enemies    []Enemy
	Bullets    []Bullet
	Rockets    []Rocket
	Explosions []Explosion
	Walls      []geometry.Rect
	SpawnZone  geometry.Rect
	Rng        *rand.Rand
}

func New() *World {
	return NewWithSeed(time.Now().UnixNano())
}

// NewWithSeed — для детерминированных тестов и реплеев.
func NewWithSeed(seed int64) *World {
	spawnSize := SpawnZoneSize
	return &World{
		Player: Player{
			X:         WorldWidth/2 - PlayerSize/2,
			Y:         WorldHeight/2 - PlayerSize/2,
			Weapon:    WeaponNone,
			Secondary: WeaponNone,
		},
		Enemies: defaultEnemies(),
		Walls:   defaultWalls(),
		SpawnZone: geometry.Rect{
			X: WorldWidth/2 - spawnSize/2,
			Y: WorldHeight/2 - spawnSize/2,
			W: spawnSize,
			H: spawnSize,
		},
		Rng: rand.New(rand.NewSource(seed)),
	}
}

// PlayerInSpawnZone — центр игрока внутри прямоугольника зоны.
func (w *World) PlayerInSpawnZone() bool {
	cx, cy := w.Player.Center()
	return cx >= w.SpawnZone.X && cx <= w.SpawnZone.X+w.SpawnZone.W &&
		cy >= w.SpawnZone.Y && cy <= w.SpawnZone.Y+w.SpawnZone.H
}

func defaultWalls() []geometry.Rect {
	t := WallThick
	return []geometry.Rect{
		// Внешняя рамка мира.
		{X: 0, Y: 0, W: WorldWidth, H: t},
		{X: 0, Y: WorldHeight - t, W: WorldWidth, H: t},
		{X: 0, Y: 0, W: t, H: WorldHeight},
		{X: WorldWidth - t, Y: 0, W: t, H: WorldHeight},

		// Внутренние препятствия.
		{X: 200, Y: 200, W: 300, H: t},
		{X: 700, Y: 300, W: t, H: 400},
		{X: 1100, Y: 200, W: 400, H: t},
		{X: 400, Y: 600, W: 500, H: t},
		{X: 1300, Y: 500, W: 300, H: t},
		{X: 300, Y: 900, W: t, H: 300},
		{X: 800, Y: 800, W: 400, H: t},
		{X: 1500, Y: 800, W: t, H: 400},
		{X: 500, Y: 1200, W: 400, H: t},
		{X: 1200, Y: 1100, W: 300, H: t},
		{X: 1700, Y: 400, W: t, H: 300},
		{X: 900, Y: 1100, W: t, H: 300},
	}
}

func defaultEnemies() []Enemy {
	positions := []struct{ x, y float64 }{
		{400, 300},
		{1600, 300},
		{1000, 1200},
	}
	enemies := make([]Enemy, 0, len(positions))
	for _, p := range positions {
		enemies = append(enemies, NewEnemy(EnemyInfantry, p.x, p.y))
	}
	return enemies
}
