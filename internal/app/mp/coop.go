package mp

import (
	"image/color"

	"Sqwave/internal/app/game"
	"Sqwave/internal/app/menu"
	"Sqwave/internal/domain/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// CoopMenu — выбор режима: создать игру в локальной сети или подключиться.
type CoopMenu struct {
	buttons  []*menu.Button
	selected int
	onHost   func()
	onJoin   func()
	onBack   func()
}

func NewCoopMenu(onHost, onJoin, onBack func()) *CoopMenu {
	s := &CoopMenu{onHost: onHost, onJoin: onJoin, onBack: onBack}
	const (
		btnW = 320.0
		btnH = 54.0
		btnX = 120.0
		btnY = 240.0
		gap  = 18.0
	)
	s.buttons = []*menu.Button{
		{Label: "Host game", X: btnX, Y: btnY, W: btnW, H: btnH, Action: onHost},
		{Label: "Join game", X: btnX, Y: btnY + btnH + gap, W: btnW, H: btnH, Action: onJoin},
		{Label: "Back", X: btnX, Y: btnY + (btnH+gap)*2, W: btnW, H: btnH, Action: onBack},
	}
	return s
}

func (s *CoopMenu) Update() error {
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	hovered := -1
	for i, b := range s.buttons {
		if b.Contains(fx, fy) {
			hovered = i
		}
	}
	if hovered >= 0 {
		s.selected = hovered
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) && hovered >= 0 {
		s.buttons[hovered].Action()
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.selected = (s.selected + 1) % len(s.buttons)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.selected = (s.selected - 1 + len(s.buttons)) % len(s.buttons)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.buttons[s.selected].Action()
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.onBack()
	}
	return nil
}

func (s *CoopMenu) Draw(r game.SceneRenderer) {
	drawFrame(r, "CO-OP")
	sh := float64(world.ScreenHeight)

	r.DrawScreenText("Local network co-op: shared world, host simulates, guests play.",
		60, 140, colDim)
	r.DrawScreenText("2-4 players. Same network (LAN, ZeroTier, Radmin).",
		60, 166, colDim)

	for i, b := range s.buttons {
		b.Draw(r, i == s.selected)
	}

	r.DrawScreenText("Arrows UP/DOWN W S select   Enter confirm   Esc back",
		60, sh-60, color.NRGBA{R: 110, G: 130, B: 160, A: 255})
}
