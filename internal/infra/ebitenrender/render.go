package ebitenrender

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

type Renderer struct {
	screen     *ebiten.Image
	camX, camY float64
}

func New(screen *ebiten.Image) *Renderer {
	return &Renderer{screen: screen}
}

func (r *Renderer) Clear(c color.NRGBA) {
	r.screen.Fill(c)
}

func (r *Renderer) SetCamera(x, y float64) {
	r.camX, r.camY = x, y
}

func (r *Renderer) DrawRect(x, y, w, h float64, c color.NRGBA) {
	ebitenutil.DrawRect(r.screen, x-r.camX, y-r.camY, w, h, c)
}

func (r *Renderer) DrawLine(x1, y1, x2, y2, thickness float64, c color.NRGBA) {
	vector.StrokeLine(
		r.screen,
		float32(x1-r.camX), float32(y1-r.camY),
		float32(x2-r.camX), float32(y2-r.camY),
		float32(thickness),
		c,
		false,
	)
}

func (r *Renderer) DrawText(text string, x, y float64) {
	ebitenutil.DebugPrintAt(r.screen, text, int(x), int(y))
}
