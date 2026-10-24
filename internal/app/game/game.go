package game

import (
	"fmt"
	"image/color"
	"math"
	"os"

	"Sqwave/internal/app/settings"
	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/systems"
	"Sqwave/internal/domain/world"
	"Sqwave/internal/infra/audio"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const aimLineLength = 90.0

type Game struct {
	world        *world.World
	input        InputSource
	camera       *Camera
	menu         Menu
	onPlayerDied func()
	deathFired   bool

	paused         bool
	pauseSelection int
	onExitToMenu   func()

	settings *settings.Panel

	lastInput input.PlayerInput
	ticks     int

	debugNav   bool
	hudVisible bool

	music *audio.Player
}

func (g *Game) SetOnPlayerDied(f func()) {
	g.onPlayerDied = f
}

func New(w *world.World, in InputSource, music *audio.Player) *Game {
	// Останавливаем музыку меню при старте игры.
	if music != nil {
		music.Stop()
	}

	g := &Game{
		world:    w,
		input:    in,
		camera:   NewCamera(),
		settings: settings.New(music),
		music:    music,
	}
	cx, cy := w.Player.Center()
	g.camera.Follow(cx, cy, cx, cy, 0)
	if !w.Player.HasWeapon() {
		g.menu.Open = true
		g.menu.Reset(&w.Player)
	}
	return g
}

// SetOnExitToMenu задаёт колбэк для выхода из игры в главное меню.
func (g *Game) SetOnExitToMenu(f func()) {
	g.onExitToMenu = f
}

func (g *Game) Update() error {
	if g.ticks == 0 {
		wd, _ := os.Getwd()
		println("CWD:", wd)
		entries, err := os.ReadDir("cmd/client/assets/music")
		if err != nil {
			println("ReadDir err:", err.Error())
		} else {
			println("Files in cmd/client/assets/music:")
			for _, e := range entries {
				println("  ", e.Name())
			}
		}
	}
	g.ticks++

	if inpututil.IsKeyJustPressed(ebiten.KeyF1) {
		g.debugNav = !g.debugNav
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		g.hudVisible = !g.hudVisible
	}

	// Settings panel — наивысший приоритет.
	if g.settings.Open {
		if g.settings.Update() {
			g.settings.Open = false
		}
		return nil
	}

	// ESC — пауза (только если loadout меню закрыто).
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && !g.menu.Open {
		g.paused = !g.paused
		if g.paused {
			g.pauseSelection = 1 // по умолчанию "Continue"
		}
		return nil
	}

	// Пауза — обрабатываем паузное меню.
	if g.paused {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
			g.pauseSelection = (g.pauseSelection - 1 + 3) % 3
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
			g.pauseSelection = (g.pauseSelection + 1) % 3
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			switch g.pauseSelection {
			case 0: // Settings
				g.settings.Open = true
				g.settings.Selected = settings.SelVolume
			case 1: // Continue
				g.paused = false
			case 2: // Exit to menu
				if g.onExitToMenu != nil {
					g.onExitToMenu()
				}
			}
		}
		return nil
	}

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
	systems.StepModules(g.world)
	systems.StepWeaponSelection(g.world, in)
	systems.StepShooting(g.world, in)
	systems.StepBullets(g.world)
	systems.StepRockets(g.world)
	systems.StepExplosions(g.world)
	systems.StepWaves(g.world)
	// R — досрочно начать волну (только во время паузы, при закрытом меню).
	if !g.menu.Open && inpututil.IsKeyJustPressed(ebiten.KeyR) && !g.world.Wave.Active {
		g.world.StartWaveRequested = true
	}

	if g.world.PendingTrack != "" {
		if g.music != nil {
			if err := g.music.PlayLoopForce(g.world.PendingTrack); err != nil {
				println("AUDIO ERROR:", err.Error())
			}
		}
		g.world.PendingTrack = ""
	}

	systems.StepEnemies(g.world)
	systems.StepMedicHeal(g.world)
	systems.StepEnemyProjectiles(g.world)
	systems.StepEnemyRockets(g.world)
	systems.StepTurrets(g.world)
	systems.StepBurning(g.world)
	systems.StepHarpoons(g.world)
	systems.StepCoins(g.world)

	cx, cy := g.world.Player.Center()
	g.camera.Follow(cx, cy, in.AimX, in.AimY, g.world.Player.AimChargeRatio())

	if !g.deathFired && g.world.Player.IsDead() {
		g.deathFired = true
		if g.onPlayerDied != nil {
			g.onPlayerDied()
		}
	}

	// Defend: если все спавны уничтожены — игрок проигрывает.
	if g.world.Wave.Kind == world.WaveKindDefend &&
		g.world.Wave.DefendOutcome == -1 &&
		!g.world.Player.IsDead() {
		g.world.Player.TakeDamage(999999)
	}

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
		if proj.Flame {
			// Пламя — круг с пульсацией.
			pulse := 0.7 + 0.3*float64(proj.Life%6)/6
			r.DrawCircle(proj.X+proj.Size/2, proj.Y+proj.Size/2,
				proj.Size/2*pulse, ColorFlame)
		} else {
			r.DrawRect(proj.X, proj.Y, proj.Size, proj.Size, ColorEnemyProjectile)
		}
	}

	// Ракеты: хвост + голова.
	for _, rk := range w.EnemyRockets {
		l := math.Hypot(rk.VX, rk.VY)
		if l > 0.01 {
			tailX := rk.X - rk.VX/l*14
			tailY := rk.Y - rk.VY/l*14
			r.DrawLine(tailX, tailY, rk.X, rk.Y, 3, ColorEnemyRocketTrail)
		}
		r.DrawRect(rk.X-rk.Size/2, rk.Y-rk.Size/2, rk.Size, rk.Size, ColorEnemyRocket)
	}

	for _, hp := range w.Harpoons {
		if hp.IsDead() {
			continue
		}

		// След за гарпуном — от позиции минус скорость × N.
		if !hp.Attached {
			trailLen := 30.0
			speed := math.Hypot(hp.VX, hp.VY)
			if speed > 0.01 {
				tx := hp.X - hp.VX/speed*trailLen
				ty := hp.Y - hp.VY/speed*trailLen
				r.DrawLine(tx, ty, hp.X, hp.Y, 3, ColorHarpoonChain)
			}
		}

		// Тело наконечника.
		r.DrawRect(hp.X-hp.Size/2, hp.Y-hp.Size/2, hp.Size, hp.Size, ColorHarpoonBody)

		// Верёвка от гарпуна к владельцу (по UID).
		for i := range w.Enemies {
			owner := &w.Enemies[i]
			if owner.UniqueID == hp.OwnerUID && !owner.IsDead() {
				ox, oy := owner.Center()
				r.DrawLine(hp.X, hp.Y, ox, oy, 2, ColorHarpoonChain)
				break
			}
		}
	}

	if g.debugNav {
		drawNavDebug(r, w)
	}

	drawCoins(r, w, g.ticks)

	drawSpawns(r, w, g.ticks)
	drawMedicBeams(r, w)
	drawSniperBeams(r, w)
	drawBufferBeams(r, w)
	drawProtectorBeams(r, w)
	drawFleerCones(r, w)
	drawTurrets(r, w)
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

	// Индикатор зарядки рывка.
	if p.Weapon.IsShield() && p.DashChargeTicks > 0 {
		pcx, pcy := p.Center()
		const barW = 50.0
		const barH = 6.0
		bx := pcx - barW/2
		by := pcy - world.PlayerSize/2 - 16

		ratio := float64(p.DashChargeTicks) / float64(world.DashChargeMaxTicks)
		ready := p.DashChargeTicks >= world.DashChargeMinTicks

		r.DrawRect(bx-1, by-1, barW+2, barH+2, ColorEnemyHPBack)
		c := color.NRGBA{R: 100, G: 200, B: 255, A: 255}
		if ready {
			// Готов к рывку — цвет от голубого к ярко-красному.
			intensity := (ratio - 0.2) / 0.8
			if intensity < 0 {
				intensity = 0
			}
			c = color.NRGBA{
				R: uint8(100 + 155*intensity),
				G: uint8(200 - 150*intensity),
				B: uint8(255 - 200*intensity),
				A: 255,
			}
		}
		r.DrawRect(bx, by, barW*ratio, barH, c)
	}

	// Индикатор кулдауна.
	if p.Weapon.IsShield() && p.DashCooldown > 0 && !p.DashActive {
		pcx, pcy := p.Center()
		maxCd := float64(world.DashCdBaseTicks + world.DashCdBonusTicks)
		ratio := float64(p.DashCooldown) / maxCd
		rad := world.PlayerSize*0.8 + ratio*6
		ring := color.NRGBA{R: 120, G: 120, B: 140, A: uint8(80 + 80*ratio)}
		r.DrawCircle(pcx, pcy, rad, ring)
	}

	if p.DashActive {
		// След за рывком.
		tailLen := 40.0
		pcx, pcy := p.Center()
		tx := pcx - p.DashActiveDirX*tailLen
		ty := pcy - p.DashActiveDirY*tailLen
		alpha := uint8(150 * float64(p.DashActiveTimer) / float64(p.DashActiveDur))
		trail := color.NRGBA{R: 180, G: 220, B: 255, A: alpha}
		r.DrawLine(pcx, pcy, tx, ty, 8, trail)
	}

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

	// Критические статусы игрока — всегда видны, даже с выключенным HUD.
	if w.Player.Held {
		r.DrawScreenTextScaled("!! CAUGHT !!", 8, float64(world.ScreenHeight)-90, 2.0,
			color.NRGBA{R: 255, G: 60, B: 80, A: 255})
	}
	if w.Player.IsBurning() {
		r.DrawScreenTextScaled("BURNING", 8, 80, 1.4,
			color.NRGBA{R: 255, G: 120, B: 40, A: 255})
	}
	if w.Player.Harpooned {
		r.DrawScreenTextScaled("HARPOONED - SHOOT THE CHAIN",
			8, float64(world.ScreenHeight)-120, 1.4,
			color.NRGBA{R: 255, G: 200, B: 60, A: 255})
	}

	if w.Player.ActiveModule != world.ModuleNone {
		r.DrawScreenTextScaled(
			fmt.Sprintf("MODULE: %s", w.Player.ActiveModule.Stats().Name),
			8, 110, 1.2,
			color.NRGBA{R: 255, G: 100, B: 180, A: 255})

		// Vulnerability: индикатор активной метки.
		if w.Player.ActiveModule == world.ModuleVulnerability && w.Player.VulnArmed {
			ratio := float64(w.Player.VulnTimer) / float64(world.VulnMarkTicks)
			sec := float64(w.Player.VulnTimer) / 60.0

			// Цвет от зелёного к красному по мере истечения.
			rr := uint8(255 * (1 - ratio))
			gg := uint8(255 * ratio)
			c := color.NRGBA{R: rr, G: gg, B: 80, A: 255}

			r.DrawScreenTextScaled(
				fmt.Sprintf("MARK %.1fs — hit for x4", sec),
				8, 130, 1.2, c)
		}
		// Rate of Fire: индикатор стаков.
		if w.Player.ActiveModule == world.ModuleRateOfFire {
			stacks := w.Player.RoFStacks
			c := color.NRGBA{R: 255, G: 140, B: 40, A: 255}
			text := fmt.Sprintf("RoF: %d/%d", stacks, world.RoFMaxStacks)
			if stacks == 0 {
				c = color.NRGBA{R: 150, G: 100, B: 60, A: 255}
			}
			r.DrawScreenTextScaled(text, 8, 130, 1.2, c)
		}
	}

	// Анонс полученного модуля.
	if w.Player.ModuleAnnounce > 0 {
		pulse := uint8(180 + 75*math.Sin(float64(w.Player.ModuleAnnounce)*0.2))
		r.DrawScreenTextScaled(
			fmt.Sprintf("NEW MODULE: %s",
				w.Player.ActiveModule.Stats().Name),
			float64(world.ScreenWidth)/2-140, 40, 1.8,
			color.NRGBA{R: 255, G: 220, B: 100, A: pulse})
	}

	// HUD — рисуется только если включён.
	if g.hudVisible {
		// Подсказки.
		r.DrawText("Sqwave — WASD move, Space dash, LMB fire, Q swap, E menu, F11 fullscreen, Tab HUD", 8, 8)
		r.DrawText(formatLoadout(&w.Player), 8, 24)
		if p.IsAiming() {
			r.DrawText(fmt.Sprintf("Charge: %d%%", int(p.AimChargeRatio()*100)), 8, 40)
			pierce := p.PierceCount()
			if pierce > 1 {
				r.DrawText(fmt.Sprintf("Pierce: %d", pierce), 8, 56)
			}
		}
		if w.PlayerInSpawnZone() && !g.menu.Open {
			r.DrawScreenText("[E] choose weapon", 8, 60, ColorSpawnZone)
		}

		drawHPBar(r, p.CurrentHP(), p.MaxHP())
		drawCoinCounter(r, p)
		drawWavePanel(r, w)

		// Индикатор установки турели.
		engineerSettingUp := false
		for i := range w.Enemies {
			e := &w.Enemies[i]
			if e.Type == world.EnemyEngineer && e.TurretSetupTimer > 0 {
				engineerSettingUp = true
				break
			}
		}
		if engineerSettingUp {
			r.DrawScreenTextScaled("ENGINEER SETTING UP TURRET",
				8, float64(world.ScreenHeight)-150, 1.6,
				color.NRGBA{R: 255, G: 100, B: 100, A: 255})
		}
	}

	if !g.hudVisible {
		r.DrawScreenText("[Tab] show HUD", 8, 8, ColorMenuDim)
	}

	// Меню всегда поверх HUD.
	if g.menu.Open {
		drawMenu(r, &g.menu, p)
	}

	if p.IsBurning() {
		intensity := float64(p.BurnTimer) / 600
		if intensity > 1 {
			intensity = 1
		}
		pulse := 0.7 + 0.3*math.Sin(float64(g.ticks)*0.3)
		alpha := uint8(float64(ColorBurnOverlay.A) * intensity * pulse)
		c := ColorBurnOverlay
		c.A = alpha
		r.DrawScreenRect(0, 0, float64(world.ScreenWidth), float64(world.ScreenHeight), c)
	}

	// Паузное меню — поверх всего.
	if g.paused {
		drawPauseMenu(r, g.pauseSelection)
	}

	// Settings — поверх паузы.
	g.settings.Draw(r)
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
	a := "-"
	b := "-"
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

		var c color.NRGBA
		switch e.Type {
		case world.EnemyShooter:
			c = ColorShooter
		case world.EnemyScout:
			if e.RecentlyHitTimer > 0 {
				c = ColorScoutFlee
			} else {
				c = ColorScout
			}
		case world.EnemyMedic:
			c = ColorMedic
		case world.EnemyRammer:
			if e.DashActive {
				c = ColorRammerDash
			} else if e.SwingActive {
				c = ColorRammerSwing
			} else {
				c = ColorRammer
			}
		case world.EnemySniper:
			c = ColorSniper
		case world.EnemyRocketeer:
			c = ColorRocketeer
		case world.EnemyBuffer:
			c = ColorBuffer
		case world.EnemyFleer:
			c = ColorFleer
		case world.EnemyEngineer:
			c = ColorEngineer
		case world.EnemyFlamer:
			c = ColorFlamer
		case world.EnemyProtector:
			c = ColorProtector
		case world.EnemyTank:
			c = ColorTank
		case world.EnemyMachineGunner:
			c = ColorMachineGunner
		case world.EnemySummoner:
			c = ColorSummoner
		case world.EnemyGhost:
			c = ColorGhost
		case world.EnemyHarpooner:
			c = ColorHarpooner
		default:
			c = ColorEnemy
		}
		if e.SwingActive {
			if e.Type == world.EnemyShooter {
				c = ColorShooterSwing
			} else {
				c = ColorEnemySwing
			}
		}
		r.DrawRect(e.X, e.Y, stats.Size, stats.Size, c)

		if e.Type == world.EnemyRammer && e.DashActive {
			// Хвост за тараном — рисуется назад по направлению рывка.
			const trailLen = 40.0
			cx, cy := e.Center()
			tailX := cx - e.DashDirX*trailLen
			tailY := cy - e.DashDirY*trailLen
			r.DrawLine(cx, cy, tailX, tailY, 8, ColorRammerDash)
		}

		if e.Type == world.EnemyFleer && e.Latched {
			cx, cy := e.Center()
			pcx, pcy := w.Player.Center()
			ang := math.Atan2(pcy-cy, pcx-cx)
			bx := cx + math.Cos(ang)*18
			by := cy + math.Sin(ang)*18
			r.DrawLine(cx, cy, bx, by, 5, ColorFleerBlade)
		}

		if e.Invulnerable {
			cx, cy := e.Center()
			rad := stats.Size*0.9 + 4
			r.DrawCircle(cx, cy, rad, ColorDome)
		}

		if e.Type == world.EnemyEngineer && !e.HasPlacedTurret {
			stats := e.Type.Stats()
			cx, cy := e.Center()

			// Если в режиме отступления — красный маркер.
			if e.FleeTimer > 0 {
				r.DrawCircle(cx, cy, 18, color.NRGBA{R: 255, G: 80, B: 80, A: 120})
			} else if e.TurretSetupTimer > 0 && e.MoveTimer <= 0 {
				// Идёт установка — полоска прогресса.
				const barW = 30.0
				const barH = 5.0
				total := float64(stats.TurretSetupTime)
				progress := 1.0 - float64(e.TurretSetupTimer)/total

				x := e.X + stats.Size/2 - barW/2
				y := e.Y - stats.Size/2 - barH - 5

				r.DrawRect(x-1, y-1, barW+2, barH+2, ColorEnemyHPBack)
				c := color.NRGBA{R: 80, G: 220, B: 220, A: 255}
				r.DrawRect(x, y, barW*progress, barH, c)

				// Пульсирующее кольцо.
				pulse := math.Sin(float64(e.AITimer)*0.4)*2 + 16
				ring := color.NRGBA{R: 80, G: 220, B: 220, A: 80}
				r.DrawCircle(cx, cy, pulse, ring)
			} else if e.MoveTimer > 0 {
				// Идёт к точке установки — маркер «несёт что-то».
				marker := color.NRGBA{R: 220, G: 200, B: 100, A: 200}
				r.DrawRect(e.X+stats.Size/2-4, e.Y-stats.Size/2-8, 8, 4, marker)
			}
		}

		if e.Type == world.EnemyTank {
			cx, cy := e.Center()
			stats := e.Type.Stats()
			arcRad := stats.ShieldArcDegrees * math.Pi / 360

			// Щит — дуга из трёх отрезков по фронту.
			shieldLen := stats.Size * 1.2
			a1 := e.WeaponAngle - arcRad
			a2 := e.WeaponAngle - arcRad*0.5
			a3 := e.WeaponAngle + arcRad*0.5
			a4 := e.WeaponAngle + arcRad

			r.DrawLine(cx, cy, cx+math.Cos(a1)*shieldLen, cy+math.Sin(a1)*shieldLen, 3, ColorTankShield)
			r.DrawLine(cx, cy, cx+math.Cos(a2)*shieldLen, cy+math.Sin(a2)*shieldLen, 3, ColorTankShield)
			r.DrawLine(cx, cy, cx+math.Cos(a3)*shieldLen, cy+math.Sin(a3)*shieldLen, 3, ColorTankShield)
			r.DrawLine(cx, cy, cx+math.Cos(a4)*shieldLen, cy+math.Sin(a4)*shieldLen, 3, ColorTankShield)

			// Контур щита — плотная линия дуги.
			r.DrawLine(
				cx+math.Cos(a1)*shieldLen, cy+math.Sin(a1)*shieldLen,
				cx+math.Cos(a4)*shieldLen, cy+math.Sin(a4)*shieldLen,
				1, ColorTankShield,
			)

			// Индикатор замаха молота.
			if e.HammerState == 1 {
				progress := float64(e.HammerTimer) / float64(stats.HammerWindup)
				rad := stats.HammerRange * progress
				pulse := color.NRGBA{
					R: 255,
					G: 140,
					B: 60,
					A: uint8(40 + 80*progress),
				}
				r.DrawCircle(cx, cy, rad, pulse)
				r.DrawCircle(cx, cy, rad, ColorTankHammerWind)
			}
		}

		if e.Type == world.EnemyMachineGunner {
			cx, cy := e.Center()
			barrelLen := stats.Size * 0.9
			bx := cx + math.Cos(e.WeaponAngle)*barrelLen
			by := cy + math.Sin(e.WeaponAngle)*barrelLen
			r.DrawLine(cx, cy, bx, by, 5, ColorMGBarrel)
		}

		if e.Type == world.EnemyGhost && e.AttackTimer > 40 {
			cx, cy := e.Center()
			pcx, pcy := w.Player.Center()
			ang := math.Atan2(pcy-cy, pcx-cx)
			r.DrawLine(cx, cy,
				cx+math.Cos(ang)*14,
				cy+math.Sin(ang)*14,
				3, ColorGhostShield)
		}

		if e.Type == world.EnemyHarpooner && e.HarpoonAiming {
			cx, cy := e.Center()
			progress := 1.0 - float64(e.HarpoonAimTimer)/float64(e.Type.Stats().HarpoonAimTime)
			alpha := uint8(40 + 120*progress)
			cAim := ColorHarpoonAim
			cAim.A = alpha
			r.DrawLine(cx, cy, e.HarpoonAimTargetX, e.HarpoonAimTargetY, 1.5, cAim)
		}

		if e.StunTimer > 0 {
			cx, cy := e.Center()
			const radius = 4.0
			spin := float64(e.StunTimer) * 0.15
			for k := 0; k < 3; k++ {
				angle := spin + float64(k)*2*math.Pi/3
				sx := cx + math.Cos(angle)*10
				sy := cy - stats.Size/2 - 8 + math.Sin(angle)*3
				r.DrawCircle(sx, sy, radius,
					color.NRGBA{R: 255, G: 220, B: 80, A: 255})
			}
		}

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

func drawEnemySwing(r SceneRenderer, e *world.Enemy) {
	if !e.SwingActive && !e.LungeActive {
		return
	}
	stats := e.Type.Stats()
	cx, cy := e.Center()

	if e.SwingActive && stats.SwingDuration > 0 {
		progress := float64(e.SwingTimer) / float64(stats.SwingDuration)
		baseAngle := math.Atan2(e.FacingY, e.FacingX)

		// Замах: меч медленно поднимается над головой (вверх-назад),
		// достигает верхней точки к 60% замаха, потом держится.
		var angle float64
		var length float64

		if progress < 0.6 {
			// 0..0.6 — поднимаем меч.
			t := progress / 0.6
			angle = baseAngle - t*1.4 // -1.4 рад = ~80° вверх-назад
			length = stats.MeleeRange * (0.5 + 0.5*t)
		} else {
			// 0.6..1.0 — меч над головой, готовится.
			angle = baseAngle - 1.4
			length = stats.MeleeRange
		}

		endX := cx + math.Cos(angle)*length
		endY := cy + math.Sin(angle)*length

		// Замах — тусклый, но с растущей яркостью.
		c := ColorEnemySwing
		c.A = uint8(150 + 100*progress)
		r.DrawLine(cx, cy, endX, endY, 5, c)

		// Индикатор над головой — «готов бьёт».
		if progress > 0.6 {
			warnPulse := 0.7 + 0.3*math.Sin(float64(e.SwingTimer)*0.8)
			rad := 5.0 + 3.0*warnPulse
			cWarn := color.NRGBA{R: 255, G: 80, B: 80, A: 220}
			r.DrawCircle(endX, endY, rad, cWarn)
		}
	} else if e.LungeActive {
		// Рывок — меч резко вперёд.
		angle := math.Atan2(e.LungeDirY, e.LungeDirX)
		endX := cx + math.Cos(angle)*stats.MeleeRange
		endY := cy + math.Sin(angle)*stats.MeleeRange
		r.DrawLine(cx, cy, endX, endY, 6, ColorEnemySword)
	}
}

func drawNavDebug(r SceneRenderer, w *world.World) {
	if w.NavGrid == nil {
		return
	}
	cs := w.NavGrid.CellSize

	// Заблокированные клетки.
	for cy := 0; cy < w.NavGrid.Rows; cy++ {
		for cx := 0; cx < w.NavGrid.Cols; cx++ {
			if !w.NavGrid.Blocked[cy*w.NavGrid.Cols+cx] {
				continue
			}
			r.DrawWorldCell(float64(cx)*cs, float64(cy)*cs, cs, ColorNavGridBlocked)
		}
	}

	// Пути врагов.
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if len(e.Path) == 0 {
			continue
		}
		ecx, ecy := e.Center()
		prevX, prevY := ecx, ecy
		for j := e.PathIndex; j < len(e.Path); j++ {
			wp := e.Path[j]
			r.DrawLine(prevX, prevY, wp.X, wp.Y, 2, ColorNavPath)
			prevX, prevY = wp.X, wp.Y
		}
		// Концы waypoint'ов.
		for j := e.PathIndex; j < len(e.Path); j++ {
			wp := e.Path[j]
			r.DrawRect(wp.X-3, wp.Y-3, 6, 6, ColorNavPathNode)
		}
	}
}

func drawMedicBeams(r SceneRenderer, w *world.World) {
	for i := range w.Enemies {
		medic := &w.Enemies[i]
		if medic.IsDead() || medic.Type != world.EnemyMedic {
			continue
		}
		if medic.HealTargetIdx < 0 || medic.HealTargetIdx >= len(w.Enemies) {
			continue
		}
		target := &w.Enemies[medic.HealTargetIdx]
		if !systems.MedicCanHeal(w, medic, target) {
			continue // луч показывается только когда реально лечим
		}

		mcx, mcy := medic.Center()
		tcx, tcy := target.Center()

		pulse := math.Sin(float64(medic.HealTick)*0.3)*0.5 + 1.0
		thickness := 2.0 + pulse

		r.DrawLine(mcx, mcy, tcx, tcy, thickness, ColorMedicBeam)
	}
}

func drawSniperBeams(r SceneRenderer, w *world.World) {
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() || e.Type != world.EnemySniper {
			continue
		}
		ecx, ecy := e.Center()

		// Яркий луч выстрела.
		if e.SniperShotLife > 0 {
			alpha := uint8(255 * e.SniperShotLife / 12)
			c := ColorSniperShot
			c.A = alpha
			r.DrawLine(ecx, ecy, e.SniperShotX, e.SniperShotY, 3, c)
		}

		// Тонкая линия прицеливания.
		if e.SniperState == world.SniperStateAiming {
			progress := 1.0 - float64(e.SniperTimer)/float64(e.Type.Stats().SniperAimDuration)
			if progress < 0 {
				progress = 0
			}
			if progress > 1 {
				progress = 1
			}
			// Линия обрезается о стены.
			ex, ey := clipToWalls(w, ecx, ecy, e.SniperAimX, e.SniperAimY)

			c := ColorSniperAim
			c.A = uint8(40 + 80*progress)
			thickness := 1.0 + progress*0.5
			r.DrawLine(ecx, ecy, ex, ey, thickness, c)
		}
	}
}

func drawBufferBeams(r SceneRenderer, w *world.World) {
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() || e.Type != world.EnemyBuffer {
			continue
		}
		if len(e.BufferTargets) == 0 {
			continue
		}
		ecx, ecy := e.Center()
		pulse := math.Sin(float64(e.AITimer)*0.15)*0.5 + 1.0
		thickness := 1.5 + pulse*0.8

		for _, idx := range e.BufferTargets {
			if idx < 0 || idx >= len(w.Enemies) {
				continue
			}
			t := &w.Enemies[idx]
			if t.IsDead() {
				continue
			}
			tcx, tcy := t.Center()
			r.DrawLine(ecx, ecy, tcx, tcy, thickness, ColorBufferBeam)
		}
	}
}

func drawFleerCones(r SceneRenderer, w *world.World) {
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() || e.Type != world.EnemyFleer {
			continue
		}
		if e.Latched {
			continue // в захвате конус не нужен
		}

		stats := e.Type.Stats()
		ecx, ecy := e.Center()
		halfCone := stats.FleerConeWidth * math.Pi / 360
		reach := stats.FleerGrabRange

		// Рисуем конус двумя отрезками из центра.
		a1 := e.AimAngle - halfCone
		a2 := e.AimAngle + halfCone
		x1 := ecx + math.Cos(a1)*reach
		y1 := ecy + math.Sin(a1)*reach
		x2 := ecx + math.Cos(a2)*reach
		y2 := ecy + math.Sin(a2)*reach

		r.DrawLine(ecx, ecy, x1, y1, 1.5, ColorFleerCone)
		r.DrawLine(ecx, ecy, x2, y2, 1.5, ColorFleerCone)

		// Направление взгляда — яркая линия.
		ax := ecx + math.Cos(e.AimAngle)*reach
		ay := ecy + math.Sin(e.AimAngle)*reach
		r.DrawLine(ecx, ecy, ax, ay, 2, ColorFleerAim)
	}
}

func drawTurrets(r SceneRenderer, w *world.World) {
	p := &w.Player
	pcx, pcy := p.Center()

	for i := range w.Turrets {
		t := &w.Turrets[i]
		if t.IsDead() {
			continue
		}

		// Радиус обнаружения — слабый круг.
		dist := math.Hypot(pcx-t.X, pcy-t.Y)
		if dist < t.DetectRange {
			r.DrawCircle(t.X, t.Y, t.DetectRange, ColorTurretDetect)
		}

		// Корпус.
		r.DrawRect(t.X-t.Size/2, t.Y-t.Size/2, t.Size, t.Size, ColorTurret)

		// Ствол в сторону игрока, если игрок близко.
		if dist > 0.01 {
			dirX := (pcx - t.X) / dist
			dirY := (pcy - t.Y) / dist
			bx := t.X + dirX*(t.Size/2+8)
			by := t.Y + dirY*(t.Size/2+8)
			r.DrawLine(t.X, t.Y, bx, by, 4, ColorTurretBarrel)
		}

		// HP-бар.
		if t.HP < t.MaxHP {
			const barH = 3.0
			barW := t.Size
			y := t.Y - t.Size/2 - barH - 2
			ratio := float64(t.HP) / float64(t.MaxHP)
			r.DrawRect(t.X-barW/2, y, barW, barH, ColorEnemyHPBack)
			r.DrawRect(t.X-barW/2, y, barW*ratio, barH, ColorHPLow)
		}
	}
}

// DrawFrozen — рисует замороженный мир (без игрока и HUD).
// Используется сценой смерти в первых фазах.
func (g *Game) DrawFrozen(r SceneRenderer) {
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

	for _, e := range w.Explosions {
		progress := 1 - float64(e.Life)/float64(e.MaxLife)
		radius := e.Radius * math.Min(progress*3, 1)
		alpha := uint8(255 * (1 - progress))
		c := ColorExplosion
		c.A = alpha
		r.DrawCircle(e.X, e.Y, radius, c)
	}

	for _, proj := range w.EnemyProjectiles {
		r.DrawRect(proj.X, proj.Y, proj.Size, proj.Size, ColorEnemyProjectile)
	}

	for _, rk := range w.EnemyRockets {
		l := math.Hypot(rk.VX, rk.VY)
		if l > 0.01 {
			tailX := rk.X - rk.VX/l*14
			tailY := rk.Y - rk.VY/l*14
			r.DrawLine(tailX, tailY, rk.X, rk.Y, 3, ColorEnemyRocketTrail)
		}
		r.DrawRect(rk.X-rk.Size/2, rk.Y-rk.Size/2, rk.Size, rk.Size, ColorEnemyRocket)
	}

	drawMedicBeams(r, w)
	drawBufferBeams(r, w)
	drawFleerCones(r, w)
	drawSniperBeams(r, w)
	drawTurrets(r, w)
	drawEnemies(r, w)

	// Игрока НЕТ — он исчез.
	// HUD НЕТ — сцена смерти сама управляет экраном.
}

func drawProtectorBeams(r SceneRenderer, w *world.World) {
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() || e.Type != world.EnemyProtector {
			continue
		}
		if len(e.ProtectTargets) == 0 {
			continue
		}
		ecx, ecy := e.Center()
		pulse := math.Sin(float64(e.AITimer)*0.15)*0.5 + 1.0
		thickness := 1.5 + pulse*0.8

		for _, idx := range e.ProtectTargets {
			if idx < 0 || idx >= len(w.Enemies) {
				continue
			}
			t := &w.Enemies[idx]
			if t.IsDead() {
				continue
			}
			tcx, tcy := t.Center()
			r.DrawLine(ecx, ecy, tcx, tcy, thickness, ColorProtectorBeam)
		}
	}
}

func drawSpawns(r SceneRenderer, w *world.World, ticks int) {
	for i := range w.SpawnPoints {
		sp := &w.SpawnPoints[i]

		isPlayerFaction := sp.Faction == world.SpawnFactionPlayer

		if sp.IsDead() {
			var c color.NRGBA
			if isPlayerFaction {
				c = color.NRGBA{R: 80, G: 20, B: 20, A: 120}
			} else {
				c = color.NRGBA{R: 60, G: 60, B: 60, A: 120}
			}
			r.DrawCircle(sp.X, sp.Y, sp.Size/2, c)
			continue
		}

		pulse := 0.9 + 0.1*math.Sin(float64(i)*1.7+0.5)
		rad := sp.Size / 2 * pulse

		var c, cInner color.NRGBA
		if isPlayerFaction {
			c = color.NRGBA{R: 60, G: 200, B: 100, A: 200}
			cInner = color.NRGBA{R: 100, G: 240, B: 140, A: 220}
		} else {
			protected := systems.SpawnIsProtected(w, sp)
			if protected {
				// Защищён — синий щит.
				c = color.NRGBA{R: 80, G: 90, B: 140, A: 220}
				cInner = color.NRGBA{R: 120, G: 140, B: 200, A: 220}

				// Пульсирующее кольцо щита.
				pulse := 0.85 + 0.15*math.Sin(float64(ticks)*0.2)
				shieldRad := sp.Size/2 + 10*pulse
				shield := color.NRGBA{R: 100, G: 180, B: 255, A: 140}
				r.DrawCircle(sp.X, sp.Y, shieldRad, shield)
			} else {
				// Открыт — можно бить. Пульсирующий красный.
				c = color.NRGBA{R: 200, G: 60, B: 80, A: 220}
				cInner = color.NRGBA{R: 255, G: 100, B: 120, A: 220}
			}
		}

		r.DrawCircle(sp.X, sp.Y, rad, c)
		r.DrawCircle(sp.X, sp.Y, rad*0.6, cInner)

		// HP-бар.
		const barW = 60.0
		const barH = 6.0
		bx := sp.X - barW/2
		by := sp.Y - sp.Size/2 - 14
		r.DrawRect(bx-1, by-1, barW+2, barH+2, ColorEnemyHPBack)

		ratio := sp.HPRatio()
		barColor := ColorHPFull
		if ratio < 0.3 {
			barColor = ColorHPLow
		} else if ratio < 0.6 {
			barColor = ColorHPMid
		}
		r.DrawRect(bx, by, barW*ratio, barH, barColor)
	}
}

func drawPauseMenu(r SceneRenderer, selected int) {
	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)

	r.DrawScreenRect(0, 0, sw, sh, color.NRGBA{R: 0, G: 0, B: 0, A: 160})

	panelW := 400.0
	panelH := 340.0
	px := (sw - panelW) / 2
	py := (sh - panelH) / 2

	border := color.NRGBA{R: 90, G: 140, B: 200, A: 200}
	bg := color.NRGBA{R: 16, G: 20, B: 30, A: 245}
	r.DrawScreenRect(px-2, py-2, panelW+4, panelH+4, border)
	r.DrawScreenRect(px, py, panelW, panelH, bg)

	r.DrawScreenTextScaled("PAUSED", px+40, py+24, 2.4,
		color.NRGBA{R: 120, G: 230, B: 255, A: 255})

	r.DrawScreenRect(px+40, py+86, panelW-80, 2,
		color.NRGBA{R: 60, G: 80, B: 110, A: 200})

	buttons := []string{"Settings", "Continue", "Exit to main menu"}
	btnW := panelW - 100
	btnX := px + 50
	btnH := 50.0
	gap := 16.0
	btnY := py + 120

	for i, label := range buttons {
		y := btnY + float64(i)*(btnH+gap)
		isSel := i == selected

		var bgBtn, borderBtn, textBtn color.NRGBA
		if isSel {
			bgBtn = color.NRGBA{R: 40, G: 70, B: 110, A: 255}
			borderBtn = color.NRGBA{R: 120, G: 230, B: 255, A: 255}
			textBtn = color.NRGBA{R: 240, G: 250, B: 255, A: 255}
		} else {
			bgBtn = color.NRGBA{R: 26, G: 32, B: 44, A: 255}
			borderBtn = color.NRGBA{R: 60, G: 80, B: 110, A: 255}
			textBtn = color.NRGBA{R: 180, G: 190, B: 210, A: 255}
		}
		r.DrawScreenRect(btnX-2, y-2, btnW+4, btnH+4, borderBtn)
		r.DrawScreenRect(btnX, y, btnW, btnH, bgBtn)
		r.DrawScreenTextScaled(label, btnX+20, y+10, 1.8, textBtn)
	}
}

func drawCoins(r SceneRenderer, w *world.World, ticks int) {
	for _, c := range w.Coins {
		pulse := 0.8 + 0.2*math.Sin(float64(ticks)*0.15)
		rad := world.CoinSize * 0.6 * pulse
		// Внешнее кольцо.
		r.DrawCircle(c.X, c.Y, rad*1.4, color.NRGBA{R: 200, G: 160, B: 40, A: 100})
		// Тело монеты.
		r.DrawCircle(c.X, c.Y, rad, color.NRGBA{R: 255, G: 210, B: 60, A: 255})
		r.DrawCircle(c.X, c.Y, rad*0.6, color.NRGBA{R: 255, G: 240, B: 140, A: 255})
	}
}
