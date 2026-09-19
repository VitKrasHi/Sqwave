package game

import "image/color"

var (
	ColorBackground = color.NRGBA{A: 255}
	ColorWall       = color.NRGBA{R: 90, G: 90, B: 90, A: 255}
	ColorPlayer     = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	ColorPlayerDash = color.NRGBA{R: 180, G: 220, B: 255, A: 255}
	ColorAim        = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	ColorBullet     = color.NRGBA{R: 255, G: 220, B: 60, A: 255}
	ColorTrail      = color.NRGBA{R: 255, G: 200, B: 40, A: 120}
)

type SceneRenderer interface {
	Clear(c color.NRGBA)

	// SetCamera задаёт смещение вьюпорта в мировых координатах.
	// Влияет только на DrawRect и DrawLine.
	SetCamera(x, y float64)

	// DrawRect и DrawLine принимают мировые координаты.
	DrawRect(x, y, w, h float64, c color.NRGBA)
	DrawLine(x1, y1, x2, y2, thickness float64, c color.NRGBA)

	// DrawText принимает экранные координаты (HUD).
	DrawText(text string, x, y float64)
}
