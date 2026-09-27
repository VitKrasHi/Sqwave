package game

import (
	"Sqwave/internal/domain/world"
)

const (
	cameraMaxZoom = 1.5

	// Доля пути к цели за один тик. 1.0 — мгновенно (было),
	// 0.2 — плавно (≈10 тиков до почти-цели при 60 TPS).
	cameraFollowLerp = 0.2
)

type Camera struct {
	X, Y float64 // мировая точка в центре экрана
	Zoom float64
}

func NewCamera() *Camera {
	return &Camera{Zoom: 1.0}
}

func (c *Camera) Follow(playerX, playerY, aimX, aimY, aimRatio float64) {
	// Куда камера должна попасть.
	targetX := playerX + (aimX-playerX)*aimRatio
	targetY := playerY + (aimY-playerY)*aimRatio
	targetZoom := 1.0 + aimRatio*(cameraMaxZoom-1.0)

	// Плавно догоняем.
	c.X += (targetX - c.X) * cameraFollowLerp
	c.Y += (targetY - c.Y) * cameraFollowLerp
	c.Zoom += (targetZoom - c.Zoom) * cameraFollowLerp

	c.clampToWorld()
}

func (c *Camera) clampToWorld() {
	visW := float64(world.ScreenWidth) / c.Zoom
	visH := float64(world.ScreenHeight) / c.Zoom

	if visW >= world.WorldWidth {
		c.X = world.WorldWidth / 2
	} else {
		if c.X < visW/2 {
			c.X = visW / 2
		}
		if c.X > world.WorldWidth-visW/2 {
			c.X = world.WorldWidth - visW/2
		}
	}

	if visH >= world.WorldHeight {
		c.Y = world.WorldHeight / 2
	} else {
		if c.Y < visH/2 {
			c.Y = visH / 2
		}
		if c.Y > world.WorldHeight-visH/2 {
			c.Y = world.WorldHeight - visH/2
		}
	}
}
