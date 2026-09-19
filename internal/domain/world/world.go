package world

import "Sqwave/internal/domain/geometry"

type World struct {
	Player            Player
	Bullets           []Bullet
	Walls             []geometry.Rect
	FireCooldownTimer int
}

func New() *World {
	return &World{
		Player: Player{
			X: WorldWidth/2 - PlayerSize/2,
			Y: WorldHeight/2 - PlayerSize/2,
		},
		Walls: defaultWalls(),
	}
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
