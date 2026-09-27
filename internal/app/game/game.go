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
	menu   Menu

	lastInput input.PlayerInput
	ticks     int // для анимаций представления
}

func New(w *world.World, in InputSource) *Game {
	g := &Game{
		world:  w,
		input:  in,
		camera: NewCamera(),
	}
	cx, cy := w.Player.Center()
	g.camera.Follow(cx, cy, cx, cy, 0)
	if !w.Player.HasWeapon() {
		g.menu.Open = true
		g.menu.Reset(&w.Player)
	}
	return g
}

func (g *Game) Update() error {
	g.ticks++

	in := g.input.Poll()
	in.AimX = (in.AimX-float64(world.ScreenWidth)/2)/g.camera.Zoom + g.camera.X
	in.AimY = (in.AimY-float64(world.ScreenHeight)/2)/g.camera.Zoom + g.camera.Y
	g.lastInput = in

	// Открыть меню.
	if in.Interact && !g.menu.Open && g.world.PlayerInSpawnZone() {
		g.menu.Open = true
		g.menu.Reset(&g.world.Player)
		cx, cy := g.world.Player.Center()
		g.camera.Follow(cx, cy, cx, cy, 0)
		return nil
	}

	if g.menu.Open {
		if handleMenuInput(g.world, &g.menu, in) {
			g.menu.Open = false
		}
		cx, cy := g.world.Player.Center()
		g.camera.Follow(cx, cy, cx, cy, 0)
		return nil
	}

	systems.StepPlayer(g.world, in)
	systems.StepWeaponSelection(g.world, in)
	systems.StepShooting(g.world, in)
	systems.StepBullets(g.world)
	systems.StepRockets(g.world)
	systems.StepExplosions(g.world)
	systems.StepEnemies(g.world)
	systems.StepEnemies(g.world)
	systems.StepEnemyProjectiles(g.world)

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

	drawSpawnZone(r, w, g.ticks)

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

	for _, proj := range w.EnemyProjectiles {
		r.DrawRect(proj.X, proj.Y, proj.Size, proj.Size, ColorEnemyProjectile)
	}

	drawEnemies(r, w)

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
	r.DrawText("Sqwave — WASD move, Space dash, LMB fire, Q swap, E menu", 8, 8)
	r.DrawText(formatLoadout(&w.Player), 8, 24)
	if p.IsAiming() {
		r.DrawText(fmt.Sprintf("Charge: %d%%", int(p.AimChargeRatio()*100)), 8, 40)
	}
	if w.PlayerInSpawnZone() && !g.menu.Open {
		r.DrawScreenText("[E] choose weapon", 8, 60, ColorSpawnZone)
	}

	drawHPBar(r, p.CurrentHP(), p.MaxHP())

	if g.menu.Open {
		drawMenu(r, &g.menu, p)
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

func drawHPBar(r SceneRenderer, hp, maxHP int) {
	const (
		barW = 220.0
		barH = 18.0
	)
	x := 8.0
	y := float64(world.ScreenHeight) - barH - 12

	r.DrawScreenRect(x-2, y-2, barW+4, barH+4, ColorHPBack)

	ratio := 0.0
	if maxHP > 0 {
		ratio = float64(hp) / float64(maxHP)
	}
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	c := ColorHPFull
	switch {
	case ratio < 0.3:
		c = ColorHPLow
	case ratio < 0.6:
		c = ColorHPMid
	}

	if ratio > 0 {
		r.DrawScreenRect(x, y, barW*ratio, barH, c)
	}
	r.DrawScreenText(fmt.Sprintf("HP %d / %d", hp, maxHP), x+6, y+1, ColorMenuText)
}

// drawSpawnZone рисует пульсирующую сферу в центре зоны спавна.
// Слои кругов с разной альфой дают мягкий градиент без шейдеров,
// а синус от ticks задаёт плавную пульсацию размера и яркости.
func drawSpawnZone(r SceneRenderer, w *world.World, ticks int) {
	z := w.SpawnZone
	cx := z.X + z.W/2
	cy := z.Y + z.H/2
	baseR := z.W / 2

	// phase = 0.07 * ticks → период ≈ 90 тиков ≈ 1.5 сек при 60 TPS.
	phase := float64(ticks) * 0.07
	pulse := math.Sin(phase) // -1..1

	// Радиус и общая яркость пульсируют в фазе.
	radius := baseR + 10*pulse
	brightness := 0.55 + 0.45*pulse // 0.1..1.0

	// Слои от внешнего к внутреннему — мягкая сфера.
	// Альфа слоёв и множитель к радиусу подобраны так, чтобы
	// получился градиент: тускло по краю, плотнее к центру.
	layers := []struct {
		scale float64
		alpha float64
	}{
		{1.15, 0.20},
		{1.00, 0.35},
		{0.78, 0.35},
		{0.50, 0.40},
		{0.22, 0.55},
	}
	for _, l := range layers {
		c := ColorSpawnZone
		a := l.alpha * brightness
		if a < 0 {
			a = 0
		}
		if a > 1 {
			a = 1
		}
		c.A = uint8(255 * a)
		r.DrawCircle(cx, cy, radius*l.scale, c)
	}
}

func formatLoadout(p *world.Player) string {
	a := "—"
	b := "—"
	if p.Weapon != world.WeaponNone {
		a = p.Weapon.Stats().Name
	}
	if p.Secondary != world.WeaponNone {
		b = p.Secondary.Stats().Name
	}
	return fmt.Sprintf("[%s]  %s  (Q to swap)", a, b)
}

func drawEnemies(r SceneRenderer, w *world.World) {
	for i := range w.Enemies {
		e := &w.Enemies[i]
		stats := e.Type.Stats()

		c := ColorEnemy
		if e.Type == world.EnemyShooter {
			c = ColorShooter
		}
		if e.SwingActive {
			if e.Type == world.EnemyShooter {
				c = ColorShooterSwing
			} else {
				c = ColorEnemySwing
			}
		}
		r.DrawRect(e.X, e.Y, stats.Size, stats.Size, c)

		// Взгляд — короткий штрих из центра.
		cx, cy := e.Center()
		r.DrawLine(cx, cy,
			cx+e.FacingX*stats.Size/2,
			cy+e.FacingY*stats.Size/2,
			2, ColorEnemyFacing)

		drawEnemyHPBar(r, e)
		drawEnemySwing(r, e)
	}
}

func drawEnemyHPBar(r SceneRenderer, e *world.Enemy) {
	stats := e.Type.Stats()
	const barH = 4.0
	barW := stats.Size
	x := e.X
	y := e.Y - barH - 3

	r.DrawRect(x-1, y-1, barW+2, barH+2, ColorEnemyHPBack)

	ratio := e.HPRatio()
	if ratio <= 0 {
		return
	}
	c := ColorHPFull
	switch {
	case ratio < 0.3:
		c = ColorHPLow
	case ratio < 0.6:
		c = ColorHPMid
	}
	r.DrawRect(x, y, barW*ratio, barH, c)
}

// drawEnemySwing — меч врага проворачивается вперёд в сторону игрока.
// Угол идёт от −45° до +45° относительно направления взгляда.
func drawEnemySwing(r SceneRenderer, e *world.Enemy) {
	if !e.SwingActive {
		return
	}
	stats := e.Type.Stats()
	if stats.SwingDuration <= 0 {
		return
	}

	progress := float64(e.SwingTimer) / float64(stats.SwingDuration)
	baseAngle := math.Atan2(e.FacingY, e.FacingX)
	angle := baseAngle - 0.8 + 1.6*progress // -0.8 → +0.8 рад

	cx, cy := e.Center()
	endX := cx + math.Cos(angle)*stats.MeleeRange
	endY := cy + math.Sin(angle)*stats.MeleeRange
	r.DrawLine(cx, cy, endX, endY, 4, ColorEnemySword)
}
