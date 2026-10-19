package settings

import (
	"fmt"
	"image/color"

	"Sqwave/internal/domain/world"
	"Sqwave/internal/infra/audio"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Renderer — минимальный интерфейс для отрисовки панели.
// SceneRenderer из game пакета удовлетворяет ему структурно.
type Renderer interface {
	DrawScreenRect(x, y, w, h float64, c color.NRGBA)
	DrawScreenText(text string, x, y float64, c color.NRGBA)
	DrawScreenTextScaled(text string, x, y, scale float64, c color.NRGBA)
}

const (
	SelVolume = 0
	SelBack   = 1
)

type Panel struct {
	Open     bool
	Selected int
	music    *audio.Player
}

func New(music *audio.Player) *Panel {
	return &Panel{
		music:    music,
		Selected: SelVolume,
	}
}

// Update обрабатывает ввод. Возвращает true, если панель нужно закрыть.
func (p *Panel) Update() bool {
	if !p.Open {
		return false
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		return true
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		p.Selected = (p.Selected - 1 + 2) % 2
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		p.Selected = (p.Selected + 1) % 2
	}

	if p.Selected == SelVolume && p.music != nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyA) {
			p.music.SetVolume(p.music.Volume() - 0.05)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) || inpututil.IsKeyJustPressed(ebiten.KeyD) {
			p.music.SetVolume(p.music.Volume() + 0.05)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if p.Selected == SelBack {
			return true
		}
	}
	return false
}

func (p *Panel) Draw(r Renderer) {
	if !p.Open {
		return
	}

	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)

	r.DrawScreenRect(0, 0, sw, sh, color.NRGBA{R: 0, G: 0, B: 0, A: 180})

	panelW := 500.0
	panelH := 320.0
	px := (sw - panelW) / 2
	py := (sh - panelH) / 2

	border := color.NRGBA{R: 90, G: 140, B: 200, A: 200}
	bg := color.NRGBA{R: 16, G: 20, B: 30, A: 245}
	r.DrawScreenRect(px-2, py-2, panelW+4, panelH+4, border)
	r.DrawScreenRect(px, py, panelW, panelH, bg)

	r.DrawScreenTextScaled("SETTINGS", px+40, py+24, 2.4,
		color.NRGBA{R: 120, G: 230, B: 255, A: 255})

	r.DrawScreenRect(px+40, py+86, panelW-80, 2,
		color.NRGBA{R: 60, G: 80, B: 110, A: 200})

	// Volume row.
	volY := py + 120
	volCol := color.NRGBA{R: 200, G: 210, B: 230, A: 255}
	volValCol := color.NRGBA{R: 120, G: 230, B: 255, A: 255}
	if p.Selected != SelVolume {
		volValCol = color.NRGBA{R: 180, G: 190, B: 210, A: 255}
	}

	if p.Selected == SelVolume {
		r.DrawScreenRect(px+20, volY+6, 10, 10,
			color.NRGBA{R: 120, G: 230, B: 255, A: 255})
	}
	r.DrawScreenText("Music volume", px+50, volY, volCol)

	barW := 160.0
	barH := 14.0
	barX := px + 260.0
	barY := volY + 3
	r.DrawScreenRect(barX-1, barY-1, barW+2, barH+2,
		color.NRGBA{R: 25, G: 25, B: 30, A: 230})

	vol := 0.0
	if p.music != nil {
		vol = p.music.Volume()
	}
	r.DrawScreenRect(barX, barY, barW*vol, barH, volValCol)
	r.DrawScreenText(fmt.Sprintf("%d%%", int(vol*100)),
		barX+barW+10, volY, volValCol)

	// Back button.
	backY := py + 240
	backW := 200.0
	backX := px + (panelW-backW)/2
	backH := 44.0

	var bgBtn, borderBtn, textBtn color.NRGBA
	if p.Selected == SelBack {
		bgBtn = color.NRGBA{R: 40, G: 70, B: 110, A: 255}
		borderBtn = color.NRGBA{R: 120, G: 230, B: 255, A: 255}
		textBtn = color.NRGBA{R: 240, G: 250, B: 255, A: 255}
	} else {
		bgBtn = color.NRGBA{R: 26, G: 32, B: 44, A: 255}
		borderBtn = color.NRGBA{R: 60, G: 80, B: 110, A: 255}
		textBtn = color.NRGBA{R: 180, G: 190, B: 210, A: 255}
	}
	r.DrawScreenRect(backX-2, backY-2, backW+4, backH+4, borderBtn)
	r.DrawScreenRect(backX, backY, backW, backH, bgBtn)
	r.DrawScreenTextScaled("Back", backX+20, backY+10, 1.6, textBtn)

	r.DrawScreenText("↑/↓ select   ←/→ change   Enter confirm   Esc back",
		px+40, py+panelH-30,
		color.NRGBA{R: 110, G: 130, B: 160, A: 255})
}
