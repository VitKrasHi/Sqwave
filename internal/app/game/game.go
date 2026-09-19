package game

import (
	"math"

	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/systems"
	"Sqwave/internal/domain/world"
)

const aimLineLength = 70.0

type Game struct {
	world *world.World
	input InputSource

	// lastInput хранится только для презентации (линия прицела),
	// в игровой логике он не нужен.
	lastInput input.PlayerInput
}

func New(w *world.World, in InputSource) *Game {
	return &Game{world: w, input: in}
}

func (g *Game) Update() error {
	in := g.input.Poll()
	g.lastInput = in

	systems.StepPlayer(g.world, in)
	systems.StepShooting(g.world, in)
	systems.StepBullets(g.world)

	return nil
}

func (g *Game) Draw(r SceneRenderer) {
	w := g.world

	r.Clear(ColorBackground)

	for _, wall := range w.Walls {
		r.DrawRect(wall.X, wall.Y, wall.W, wall.H, ColorWall)
	}

	// Сначала все шлейфы, потом все квадраты.
	for _, b := range w.Bullets {
		l := math.Hypot(b.VX, b.VY)
		if l == 0 {
			continue
		}
		headX := b.X + world.BulletSize/2
		headY := b.Y + world.BulletSize/2
		tailX := headX - b.VX/l*world.BulletTrailLen
		tailY := headY - b.VY/l*world.BulletTrailLen
		r.DrawLine(headX, headY, tailX, tailY, 2, ColorTrail)
	}
	for _, b := range w.Bullets {
		r.DrawRect(b.X, b.Y, world.BulletSize, world.BulletSize, ColorBullet)
	}

	p := &w.Player
	playerColor := ColorPlayer
	if p.Dashing() {
		playerColor = ColorPlayerDash
	}
	r.DrawRect(p.X, p.Y, world.PlayerSize, world.PlayerSize, playerColor)

	// Прицел.
	cx, cy := p.Center()
	dx := g.lastInput.AimX - cx
	dy := g.lastInput.AimY - cy
	if l := math.Hypot(dx, dy); l > 0 {
		dx /= l
		dy /= l
	}
	r.DrawLine(cx, cy, cx+dx*aimLineLength, cy+dy*aimLineLength, 2, ColorAim)

	r.DrawText("Sqwave — WASD: move, Space: dash, LMB: fire", 8, 8)
}
