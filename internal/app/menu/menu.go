package menu

import (
	"image/color"
	"math/rand"
	"time"

	"Sqwave/internal/app/game"
	"Sqwave/internal/app/particles"
	"Sqwave/internal/domain/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const panelW = 380.0

type Menu struct {
	onStart func()
	onExit  func()

	buttons  []*Button
	selected int
	ps       []particles.Particle
}

func New(onStart, onExit func()) *Menu {
	m := &Menu{
		onStart: onStart,
		onExit:  onExit,
	}
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	m.buildButtons()
	m.ps = particles.New(panelW+20, float64(world.ScreenWidth)-20,
		20, float64(world.ScreenHeight)-20, rng)
	return m
}

func (m *Menu) buildButtons() {
	const (
		btnW = 260.0
		btnH = 54.0
		btnX = 60.0
		btnY = 320.0
		gap  = 20.0
	)
	m.buttons = []*Button{
		{Label: "Create game", X: btnX, Y: btnY, W: btnW, H: btnH, Action: m.onStart},
		{Label: "Exit", X: btnX, Y: btnY + btnH + gap, W: btnW, H: btnH, Action: m.onExit},
	}
}

func (m *Menu) Update() error {
	particles.Update(m.ps, panelW+20, float64(world.ScreenWidth)-20,
		20, float64(world.ScreenHeight)-20)

	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	hovered := -1
	for i, b := range m.buttons {
		if b.Contains(fx, fy) {
			hovered = i
		}
	}
	if hovered >= 0 {
		m.selected = hovered
	}

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && hovered >= 0 {
		m.buttons[hovered].Action()
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		m.selected = (m.selected + 1) % len(m.buttons)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		m.selected = (m.selected - 1 + len(m.buttons)) % len(m.buttons)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		m.buttons[m.selected].Action()
	}
	return nil
}

func (m *Menu) Draw(r game.SceneRenderer) {
	sh := float64(world.ScreenHeight)

	r.Clear(color.NRGBA{R: 10, G: 12, B: 20, A: 255})
	particles.Draw(r, m.ps)

	r.DrawScreenRect(0, 0, panelW, sh, color.NRGBA{R: 16, G: 20, B: 30, A: 245})
	r.DrawScreenRect(panelW, 0, 2, sh, color.NRGBA{R: 90, G: 140, B: 200, A: 180})

	r.DrawScreenTextScaled("SQWAVE", 60, 120, 5.0,
		color.NRGBA{R: 120, G: 230, B: 255, A: 255})

	for i, b := range m.buttons {
		b.Draw(r, i == m.selected)
	}

	r.DrawScreenText("Arrows / W S select   Enter confirm",
		60, sh-60, color.NRGBA{R: 110, G: 130, B: 160, A: 255})
	r.DrawScreenText("F11 fullscreen",
		60, sh-40, color.NRGBA{R: 90, G: 110, B: 140, A: 255})
}
