package card

import (
	"image/color"
	"math"

	"Sqwave/internal/app/game"
	"Sqwave/internal/domain/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Дизайн карточки: 630×500.
const (
	cardDesignW = 630.0
	cardDesignH = 500.0

	cardTitle      = "Sqwave"
	cardTitleScale = 6.0

	cardPlayerCX   = 315.0
	cardPlayerCY   = 330.0
	cardPlayerSize = 140.0
)

var (
	colorTitle  = color.NRGBA{R: 120, G: 230, B: 255, A: 255}
	colorBg     = color.NRGBA{R: 10, G: 12, B: 20, A: 255}
	colorBorder = color.NRGBA{R: 90, G: 140, B: 200, A: 180}
	colorPlayer = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
)

type Scene struct {
	onExit func()
	ticks  int
}

func New(onExit func()) *Scene {
	return &Scene{onExit: onExit}
}

func (s *Scene) Update() error {
	s.ticks++

	for k := ebiten.Key(0); k < 256; k++ {
		if inpututil.IsKeyJustPressed(k) {
			if s.onExit != nil {
				s.onExit()
			}
			return nil
		}
	}
	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if s.onExit != nil {
			s.onExit()
		}
	}
	return nil
}

func (s *Scene) Draw(r game.SceneRenderer) {
	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)

	r.Clear(color.NRGBA{R: 0, G: 0, B: 0, A: 255})

	scale := math.Min(sw/cardDesignW, sh/cardDesignH)
	cardW := cardDesignW * scale
	cardH := cardDesignH * scale
	cardX := (sw - cardW) / 2
	cardY := (sh - cardH) / 2

	drawCardInto(r, cardX, cardY, cardW, cardH)

	// Подсказка внизу.
	hintPulse := uint8(120 + 100*math.Abs(math.Sin(float64(s.ticks)*0.06)))
	hint := color.NRGBA{R: 200, G: 200, B: 200, A: hintPulse}
	r.DrawScreenTextScaled("PRESS ANY KEY", 20, sh-40, 1.6, hint)
}

// drawCardInto рисует композицию карточки в прямоугольник (x, y, w, h).
func drawCardInto(r game.SceneRenderer, x, y, w, h float64) {
	scale := w / cardDesignW

	// Фон.
	r.DrawScreenRect(x, y, w, h, colorBg)

	// Рамка.
	bw := 2.0 * scale
	r.DrawScreenRect(x, y, w, bw, colorBorder)
	r.DrawScreenRect(x, y+h-bw, w, bw, colorBorder)
	r.DrawScreenRect(x, y, bw, h, colorBorder)
	r.DrawScreenRect(x+w-bw, y, bw, h, colorBorder)

	// Заголовок сверху по центру.
	titleScale := cardTitleScale * scale
	titleW := float64(len(cardTitle)) * 7.0 * titleScale
	titleX := x + (w-titleW)/2
	titleY := y + 55*scale

	shadowCol := color.NRGBA{R: 0, G: 60, B: 100, A: 200}
	r.DrawScreenTextScaled(cardTitle, titleX+3*scale, titleY+3*scale, titleScale, shadowCol)
	r.DrawScreenTextScaled(cardTitle, titleX, titleY, titleScale, colorTitle)

	// Черточки: подчёркивание + линии по бокам.
	underlineY := titleY + 13*titleScale + 8*scale
	underlineW := titleW * 1.15
	underlineX := x + (w-underlineW)/2
	r.DrawScreenRect(underlineX, underlineY, underlineW, 2*scale, colorTitle)

	// Боковые линии (тире) слева и справа от заголовка.
	dashY := titleY + 13*titleScale/2 - scale
	dashW := 30.0 * scale
	dashH := 3.0 * scale
	gap := 16.0 * scale

	leftDashX := titleX - dashW - gap
	rightDashX := titleX + titleW + gap
	r.DrawScreenRect(leftDashX, dashY, dashW, dashH, colorTitle)
	r.DrawScreenRect(rightDashX, dashY, dashW, dashH, colorTitle)

	// Игрок — белый квадрат.
	pcx := x + cardPlayerCX*scale
	pcy := y + cardPlayerCY*scale
	ps := cardPlayerSize * scale

	r.DrawScreenRect(pcx-ps/2, pcy-ps/2, ps, ps, colorPlayer)
}

// ---------- Иконка 630×500 ----------

// SaveIcon рендерит карточку в PNG 630×500 и сохраняет по пути path.
// Импортирует только image/*, без Ebiten, чтобы можно было вызвать
// до инициализации окна.
func SaveIcon(path string) error {
	return saveIconInternal(path)
}
