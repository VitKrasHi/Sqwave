package game

import "image/color"

// Палитра — решение представления, а не игровой логики,
// поэтому живёт в app, а не в domain.
var (
	ColorBackground = color.NRGBA{A: 255}
	ColorWall       = color.NRGBA{R: 90, G: 90, B: 90, A: 255}
	ColorPlayer     = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	ColorPlayerDash = color.NRGBA{R: 180, G: 220, B: 255, A: 255}
	ColorAim        = color.NRGBA{R: 0, G: 255, B: 0, A: 255}
	ColorBullet     = color.NRGBA{R: 255, G: 220, B: 60, A: 255}
	ColorTrail      = color.NRGBA{R: 255, G: 200, B: 40, A: 120}
)

// SceneRenderer — абстракция над движком рисования.
// App не знает, что под ней Ebiten.
type SceneRenderer interface {
	Clear(c color.NRGBA)
	DrawRect(x, y, w, h float64, c color.NRGBA)
	DrawLine(x1, y1, x2, y2, thickness float64, c color.NRGBA)
	DrawText(text string, x, y float64)
}
