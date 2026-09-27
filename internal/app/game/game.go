package game

import (
	"fmt"
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

	// Экранные координаты курсора → мировые.
	in.AimX += g.camera.X
	in.AimY += g.camera.Y

	g.lastInput = in

	systems.StepPlayer(g.world, in)
	systems.StepWeaponSelection(g.world, in)
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
		alpha := uint8(int(255) * b.Life / b.MaxLife)
		trail := bulletTrailColor(b.Weapon)
		trail.A = alpha
		r.DrawLine(b.StartX, b.StartY, b.EndX, b.EndY, 2, trail)

		tip := bulletColor(b.Weapon)
		tip.A = alpha
		r.DrawRect(b.EndX-2, b.EndY-2, 4, 4, tip)
	}

	p := &w.Player
	playerColor := ColorPlayer
	if p.Dashing() {
		playerColor = ColorPlayerDash
	}
	r.DrawRect(p.X, p.Y, world.PlayerSize, world.PlayerSize, playerColor)

	pcx, pcy := p.Center()
	dx := g.lastInput.AimX - pcx
	dy := g.lastInput.AimY - pcy
	if l := math.Hypot(dx, dy); l > 0 {
		dx /= l
		dy /= l
	}
	r.DrawLine(pcx, pcy, pcx+dx*aimLineLength, pcy+dy*aimLineLength, 2, ColorAim)

	r.DrawText("Sqwave — WASD: move, Space: dash, LMB: fire, 1/2/3: weapon", 8, 8)
	r.DrawText(fmt.Sprintf("Current weapon: %s", p.Weapon.Stats().Name), 8, 24)
}
