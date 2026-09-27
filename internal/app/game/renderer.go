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

	ColorPistolBullet = color.NRGBA{R: 255, G: 220, B: 60, A: 255}
	ColorPistolTrail  = color.NRGBA{R: 255, G: 200, B: 40, A: 120}

	ColorSMGBullet = color.NRGBA{R: 120, G: 230, B: 255, A: 255}
	ColorSMGTrail  = color.NRGBA{R: 80, G: 200, B: 255, A: 120}

	ColorShotgunBullet = color.NRGBA{R: 255, G: 140, B: 40, A: 255}
	ColorShotgunTrail  = color.NRGBA{R: 255, G: 100, B: 20, A: 100}

	ColorSniperBullet = color.NRGBA{R: 220, G: 220, B: 255, A: 255}
	ColorSniperTrail  = color.NRGBA{R: 180, G: 180, B: 255, A: 200}
)

func bulletColor(w world.WeaponType) color.NRGBA {
	switch w {
	case world.WeaponSMG:
		return ColorSMGBullet
	case world.WeaponShotgun:
		return ColorShotgunBullet
	case world.WeaponSniper:
		return ColorSniperBullet
	default:
		return ColorPistolBullet
	}
}

func bulletTrailColor(w world.WeaponType) color.NRGBA {
	switch w {
	case world.WeaponSMG:
		return ColorSMGTrail
	case world.WeaponShotgun:
		return ColorShotgunTrail
	case world.WeaponSniper:
		return ColorSniperTrail
	default:
		return ColorPistolTrail
	}
}

type SceneRenderer interface {
	Clear(c color.NRGBA)

	// SetCamera задаёт мировую точку в центре экрана и масштаб.
	// Влияет только на DrawRect и DrawLine.
	SetCamera(x, y, zoom float64)

	// DrawRect и DrawLine принимают мировые координаты.
	DrawRect(x, y, w, h float64, c color.NRGBA)
	DrawLine(x1, y1, x2, y2, thickness float64, c color.NRGBA)

	// Screen-space — для оверлеев и HUD (не зависят от камеры и зума).
	DrawScreenRect(x, y, w, h float64, c color.NRGBA)
	DrawScreenLine(x1, y1, x2, y2, thickness float64, c color.NRGBA)
	DrawText(text string, x, y float64)
}
