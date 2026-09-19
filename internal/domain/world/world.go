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
			X: ScreenWidth/2 - PlayerSize/2,
			Y: ScreenHeight/2 - PlayerSize/2,
		},
		Walls: defaultWalls(),
	}
}

func defaultWalls() []geometry.Rect {
	return []geometry.Rect{
		{X: 80, Y: 80, W: 220, H: 30},
		{X: 400, Y: 160, W: 30, H: 320},
		{X: 600, Y: 80, W: 260, H: 30},
		{X: 120, Y: 380, W: 320, H: 30},
		{X: 680, Y: 420, W: 200, H: 30},
		{X: 80, Y: 480, W: 30, H: 120},
		{X: 850, Y: 200, W: 30, H: 200},
	}
}
