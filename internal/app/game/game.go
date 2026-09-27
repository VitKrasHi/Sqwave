package game

import (
	"fmt"
	"image/color"
	"math"

	"Sqwave/internal/domain/geometry"
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

	// Меч: клинок + затухающий след проворота.
	if p.Swing.Active {
		if p.Swing.Weapon.Stats().IsThrust {
			drawThrust(r, p, pcx, pcy)
		} else {
			drawSwing(r, w, p, pcx, pcy)
		}
	}

	// HUD.
	r.DrawText("Sqwave — WASD: move, Space: dash(shield), LMB: fire, 1..7: weapon", 8, 8)
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

// clipToWalls возвращает ближайшую точку пересечения отрезка со стенами
// или исходный конец, если пересечений нет. Используется для визуальной
// обрезки клинка и ракетного следа о препятствия.
func clipToWalls(w *world.World, x1, y1, x2, y2 float64) (float64, float64) {
	t := 1.0
	for _, wall := range w.Walls {
		if tt, ok := geometry.RaySegmentIntersectsRect(x1, y1, x2, y2, wall); ok {
			if tt < t {
				t = tt
			}
		}
	}
	return x1 + (x2-x1)*t, y1 + (y2-y1)*t
}

// drawSwing — прежний боковой удар: клинок описывает дугу,
// за ним тянется затухающий след.
func drawSwing(r SceneRenderer, w *world.World, p *world.Player, pcx, pcy float64) {
	swingAngle := p.CurrentSwingAngle()
	progress := p.SwingProgress()

	bladeColor := meleeColor(p.Swing.Weapon)
	trailBase := meleeTrailColor(p.Swing.Weapon)
	bladeWidth := p.Swing.Weapon.Stats().BladeWidth
	if bladeWidth <= 0 {
		bladeWidth = 5
	}

	const trailSteps = 6
	const trailStep = 0.07
	for i := trailSteps; i >= 1; i-- {
		t := progress - float64(i)*trailStep
		if t <= 0 {
			continue
		}
		ang := p.Swing.StartAngle + p.Swing.ArcRadians*t
		ex, ey := clipToWalls(w, pcx, pcy,
			pcx+math.Cos(ang)*p.Swing.Range,
			pcy+math.Sin(ang)*p.Swing.Range)
		c := trailBase
		c.A = uint8(float64(trailBase.A) * (1 - float64(i)/float64(trailSteps)))
		r.DrawLine(pcx, pcy, ex, ey, bladeWidth*0.6, c)
	}

	endX, endY := clipToWalls(w, pcx, pcy,
		pcx+math.Cos(swingAngle)*p.Swing.Range,
		pcy+math.Sin(swingAngle)*p.Swing.Range)
	r.DrawLine(pcx, pcy, endX, endY, bladeWidth, bladeColor)
}

// drawThrust — щит выезжает вперёд и возвращается назад.
// Дистанция меняется по синусоиде: 0 → Range → 0 за SwingDuration.
func drawThrust(r SceneRenderer, p *world.Player, pcx, pcy float64) {
	progress := p.SwingProgress()

	// Синус даёт плавное «туда-обратно» с максимумом ровно посередине.
	reach := math.Sin(progress * math.Pi) // 0 → 1 → 0
	dist := p.Swing.Range * reach
	ang := p.Swing.StartAngle

	hx := pcx + math.Cos(ang)*dist
	hy := pcy + math.Sin(ang)*dist

	// «Рука» от игрока к щиту — тонкая линия, видна только когда щит выдвинут.
	if reach > 0.15 {
		arm := ColorShieldTrail
		arm.A = uint8(float64(ColorShieldTrail.A) * reach)
		r.DrawLine(pcx, pcy, hx, hy, 4, arm)
	}

	// Сам щит — квадрат. Слегка «дышит» по размеру, чтобы чувствовался импульс.
	size := p.Swing.Weapon.Stats().BladeWidth
	size *= 1.0 + 0.15*reach
	r.DrawRect(hx-size/2, hy-size/2, size, size, ColorShield)
}
