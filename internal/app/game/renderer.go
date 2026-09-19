package game

import (
	"image/color"

	"Sqwave/internal/domain/world"
)

var (
	ColorBackground = color.NRGBA{A: 255}
	ColorWall       = color.NRGBA{R: 90, G: 90, B: 90, A: 255}
	ColorPlayer     = color.NRGBA{R: 255, G: 255, B: 255, A: 255}
	ColorPlayerDash = color.NRGBA{R: 180, G: 220, B: 255, A: 255}
	ColorAim        = color.NRGBA{R: 0, G: 255, B: 0, A: 255}

	// Пистолет — жёлтый.
	ColorPistolBullet = color.NRGBA{R: 255, G: 220, B: 60, A: 255}
	ColorPistolTrail  = color.NRGBA{R: 255, G: 200, B: 40, A: 120}

	// SMG — голубой.
	ColorSMGBullet = color.NRGBA{R: 120, G: 230, B: 255, A: 255}
	ColorSMGTrail  = color.NRGBA{R: 80, G: 200, B: 255, A: 120}
)

func bulletColor(w world.WeaponType) color.NRGBA {
	if w == world.WeaponSMG {
		return ColorSMGBullet
	}
	return ColorPistolBullet
}

func bulletTrailColor(w world.WeaponType) color.NRGBA {
	if w == world.WeaponSMG {
		return ColorSMGTrail
	}
	return ColorPistolTrail
}

type SceneRenderer interface {
	Clear(c color.NRGBA)
	SetCamera(x, y float64)
	DrawRect(x, y, w, h float64, c color.NRGBA)
	DrawLine(x1, y1, x2, y2, thickness float64, c color.NRGBA)
	DrawText(text string, x, y float64)
}
