package game

import (
	"fmt"
	"image/color"
	"math"

	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/systems"
	"Sqwave/internal/domain/world"
)

const aimLineLength = 90.0

type Game struct {
	world  *world.World
	input  InputSource
	camera *Camera

	lastInput input.PlayerInput
}

func New(w *world.World, in InputSource) *Game {
	g := &Game{
		world:  w,
		input:  in,
		camera: NewCamera(),
	}
	cx, cy := w.Player.Center()
	g.camera.Follow(cx, cy, cx, cy, 0)
	return g
}

func (g *Game) Update() error {
	in := g.input.Poll()

	// Экранные координаты курсора → мировые с учётом зума и смещения камеры.
	// Используем состояние камеры с прошлого кадра — задержка в 1 тик
	// незаметна, зато избегаем курицы-яйца.
	in.AimX = (in.AimX-float64(world.ScreenWidth)/2)/g.camera.Zoom + g.camera.X
	in.AimY = (in.AimY-float64(world.ScreenHeight)/2)/g.camera.Zoom + g.camera.Y

	g.lastInput = in

	systems.StepPlayer(g.world, in)
	systems.StepWeaponSelection(g.world, in)
	systems.StepShooting(g.world, in)
	systems.StepBullets(g.world)
	systems.StepRockets(g.world)
	systems.StepExplosions(g.world)

	cx, cy := g.world.Player.Center()
	g.camera.Follow(cx, cy, in.AimX, in.AimY, g.world.Player.AimChargeRatio())

	return nil
}

func (g *Game) Draw(r SceneRenderer) {
	w := g.world
	r.SetCamera(g.camera.X, g.camera.Y, g.camera.Zoom)
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

	// Ракеты: хвост назад по вектору скорости + яркая «голова».
	for _, rocket := range w.Rockets {
		l := math.Hypot(rocket.VX, rocket.VY)
		if l == 0 {
			continue
		}
		tailLen := 14.0
		tailX := rocket.X - rocket.VX/l*tailLen
		tailY := rocket.Y - rocket.VY/l*tailLen
		r.DrawLine(tailX, tailY, rocket.X, rocket.Y, 3, ColorRocketTrail)
		r.DrawRect(rocket.X-rocket.Size/2, rocket.Y-rocket.Size/2, rocket.Size, rocket.Size, ColorRocketBullet)
	}

	// Взрывы: круг расширяется в первой трети жизни, затем плавно гаснет.
	for _, e := range w.Explosions {
		progress := 1 - float64(e.Life)/float64(e.MaxLife) // 0 → 1
		radius := e.Radius * math.Min(progress*3, 1)
		alpha := uint8(255 * (1 - progress))
		c := ColorExplosion
		c.A = alpha
		r.DrawCircle(e.X, e.Y, radius, c)
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

	// Прицел снайперки — маска + перекрестие.
	if p.IsAiming() {
		drawScope(r)
	}

	// HUD.
	r.DrawText("Sqwave — WASD: move, Space: dash, LMB: fire, 1/2/3/4/5: weapon", 8, 8)
	r.DrawText(fmt.Sprintf("Current weapon: %s", p.Weapon.Stats().Name), 8, 24)
	if p.IsAiming() {
		r.DrawText(fmt.Sprintf("Charge: %d%%", int(p.AimChargeRatio()*100)), 8, 40)
	}
}

// drawScope затемняет всё, кроме прямоугольного «окна», и рисует перекрестие.
// Чем больше ratio — тем больше окно и тем сильнее затемнение по краям.
func drawScope(r SceneRenderer) {
	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)
	cx := sw / 2
	cy := sh / 2

	cross := color.NRGBA{R: 255, G: 255, B: 255, A: 220}
	r.DrawScreenLine(cx-18, cy, cx-6, cy, 1, cross)
	r.DrawScreenLine(cx+6, cy, cx+18, cy, 1, cross)
	r.DrawScreenLine(cx, cy-18, cx, cy-6, 1, cross)
	r.DrawScreenLine(cx, cy+6, cx, cy+18, 1, cross)
}
