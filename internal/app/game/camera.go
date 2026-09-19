package game

import "Sqwave/internal/domain/world"

// Camera хранит мировую координату левого верхнего угла вьюпорта.
type Camera struct {
	X, Y float64
}

// Follow центрирует камеру на точке (tx, ty) и ограничивает её
// границами мира. Вызывается после обновления игрока.
func (c *Camera) Follow(tx, ty float64) {
	c.X = tx - world.ScreenWidth/2
	c.Y = ty - world.ScreenHeight/2

	if c.X < 0 {
		c.X = 0
	}
	if c.Y < 0 {
		c.Y = 0
	}
	if maxX := world.WorldWidth - world.ScreenWidth; c.X > maxX {
		c.X = maxX
	}
	if maxY := world.WorldHeight - world.ScreenHeight; c.Y > maxY {
		c.Y = maxY
	}
}
