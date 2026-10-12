package menu

import (
	"image/color"

	"Sqwave/internal/app/game"
)

type Button struct {
	Label  string
	X, Y   float64
	W, H   float64
	Action func()
}

func (b *Button) Contains(x, y float64) bool {
	return x >= b.X && x <= b.X+b.W && y >= b.Y && y <= b.Y+b.H
}

func (b *Button) Draw(r game.SceneRenderer, selected bool) {
	var bg, border, text color.NRGBA
	if selected {
		bg = color.NRGBA{R: 40, G: 70, B: 110, A: 255}
		border = color.NRGBA{R: 120, G: 230, B: 255, A: 255}
		text = color.NRGBA{R: 240, G: 250, B: 255, A: 255}
	} else {
		bg = color.NRGBA{R: 26, G: 32, B: 44, A: 255}
		border = color.NRGBA{R: 60, G: 80, B: 110, A: 255}
		text = color.NRGBA{R: 180, G: 190, B: 210, A: 255}
	}

	// Рамка.
	r.DrawScreenRect(b.X-2, b.Y-2, b.W+4, b.H+4, border)
	// Фон.
	r.DrawScreenRect(b.X, b.Y, b.W, b.H, bg)

	// Метка по центру кнопки (примерно).
	textScale := 1.8
	r.DrawScreenTextScaled(b.Label, b.X+18, b.Y+b.H/2-24, textScale, text)

	// Маркер выбора слева от кнопки.
	if selected {
		const markerSize = 10.0
		my := b.Y + b.H/2
		marker := color.NRGBA{R: 120, G: 230, B: 255, A: 255}
		r.DrawScreenRect(b.X-24, my-markerSize/2, markerSize, markerSize, marker)
	}
}
