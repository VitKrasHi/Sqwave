package ebitenrender

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/ebitenutil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
	"golang.org/x/image/font/basicfont"
)

type Renderer struct {
	screen     *ebiten.Image
	camX, camY float64
	zoom       float64
}

func New(screen *ebiten.Image) *Renderer {
	return &Renderer{screen: screen, zoom: 1.0}
}

func (r *Renderer) Clear(c color.NRGBA) {
	r.screen.Fill(c)
}

func (r *Renderer) SetCamera(x, y, zoom float64) {
	r.camX, r.camY, r.zoom = x, y, zoom
}

func (r *Renderer) w2s(x, y float64) (float64, float64) {
	b := r.screen.Bounds()
	sw := float64(b.Dx())
	sh := float64(b.Dy())
	return (x-r.camX)*r.zoom + sw/2, (y-r.camY)*r.zoom + sh/2
}

func (r *Renderer) DrawRect(x, y, w, h float64, c color.NRGBA) {
	sx, sy := r.w2s(x, y)
	ebitenutil.DrawRect(r.screen, sx, sy, w*r.zoom, h*r.zoom, c)
}

func (r *Renderer) DrawLine(x1, y1, x2, y2, thickness float64, c color.NRGBA) {
	sx1, sy1 := r.w2s(x1, y1)
	sx2, sy2 := r.w2s(x2, y2)
	vector.StrokeLine(
		r.screen,
		float32(sx1), float32(sy1),
		float32(sx2), float32(sy2),
		float32(thickness*r.zoom),
		c,
		false,
	)
}

func (r *Renderer) DrawScreenRect(x, y, w, h float64, c color.NRGBA) {
	ebitenutil.DrawRect(r.screen, x, y, w, h, c)
}

func (r *Renderer) DrawScreenLine(x1, y1, x2, y2, thickness float64, c color.NRGBA) {
	vector.StrokeLine(
		r.screen,
		float32(x1), float32(y1),
		float32(x2), float32(y2),
		float32(thickness),
		c,
		false,
	)
}

func (r *Renderer) DrawText(text string, x, y float64) {
	ebitenutil.DebugPrintAt(r.screen, text, int(x), int(y))
}

func (r *Renderer) DrawCircle(x, y, radius float64, c color.NRGBA) {
	sx, sy := r.w2s(x, y)
	vector.DrawFilledCircle(r.screen, float32(sx), float32(sy), float32(radius*r.zoom), c, true)
}

var screenTextFace = text.NewGoXFace(basicfont.Face7x13)

func (r *Renderer) DrawScreenText(s string, x, y float64, c color.NRGBA) {
	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(c)
	text.Draw(r.screen, s, screenTextFace, op)
}

func (r *Renderer) DrawWorldCell(x, y, size float64, c color.NRGBA) {
	sx, sy := r.w2s(x, y)
	ebitenutil.DrawRect(r.screen, sx, sy, size*r.zoom, size*r.zoom, c)
}
