package game

import (
	"math"

	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/systems"
	"Sqwave/internal/domain/world"
)

const aimLineLength = 70.0

type Game struct {
	world  *world.World
	input  InputSource
	camera Camera

	lastInput input.PlayerInput
}

func New(w *world.World, in InputSource) *Game {
	g := &Game{world: w, input: in}
	cx, cy := w.Player.Center()
	g.camera.Follow(cx, cy)
	return g
}

func (g *Game) Update() error {
	in := g.input.Poll()

	// Курсор приходит в экранных координатах. Переводим в мировые,
	// потому что вся игровая логика работает в мировых.
	// Камера — это представление, поэтому перевод делается здесь,
	// а не в domain.
	in.AimX += g.camera.X
	in.AimY += g.camera.Y

	g.lastInput = in

	systems.StepPlayer(g.world, in)
	systems.StepShooting(g.world, in)
	systems.StepBullets(g.world)

	cx, cy := g.world.Player.Center()
	g.camera.Follow(cx, cy)

	return nil
}

func (g *Game) Draw(r SceneRenderer) {
	w := g.world
	r.SetCamera(g.camera.X, g.camera.Y)
	r.Clear(ColorBackground)

	for _, wall := range w.Walls {
		r.DrawRect(wall.X, wall.Y, wall.W, wall.H, ColorWall)
	}

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

	// Прицел — теперь без ручного сложения с камерой:
	pcx, pcy := p.Center()
	dx := g.lastInput.AimX - pcx
	dy := g.lastInput.AimY - pcy
	if l := math.Hypot(dx, dy); l > 0 {
		dx /= l
		dy /= l
	}
	r.DrawLine(pcx, pcy, pcx+dx*aimLineLength, pcy+dy*aimLineLength, 2, ColorAim)

	r.DrawText("Sqwave — WASD: move, Space: dash, LMB: fire", 8, 8)
}
