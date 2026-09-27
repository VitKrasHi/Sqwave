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

	ColorRocketBullet = color.NRGBA{R: 255, G: 80, B: 60, A: 255}
	ColorRocketTrail  = color.NRGBA{R: 255, G: 120, B: 40, A: 180}
	ColorExplosion    = color.NRGBA{R: 255, G: 180, B: 60, A: 255}

	ColorSword      = color.NRGBA{R: 220, G: 240, B: 255, A: 255}
	ColorSwordTrail = color.NRGBA{R: 140, G: 200, B: 255, A: 180}
)

func bulletColor(w world.WeaponType) color.NRGBA {
	switch w {
	case world.WeaponSMG:
		return ColorSMGBullet
	case world.WeaponShotgun:
		return ColorShotgunBullet
	case world.WeaponSniper:
		return ColorSniperBullet
	case world.WeaponRocket:
		return ColorRocketBullet
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
	SetCamera(x, y, zoom float64)

	DrawRect(x, y, w, h float64, c color.NRGBA)
	DrawLine(x1, y1, x2, y2, thickness float64, c color.NRGBA)
	DrawCircle(x, y, r float64, c color.NRGBA)

	DrawScreenRect(x, y, w, h float64, c color.NRGBA)
	DrawScreenLine(x1, y1, x2, y2, thickness float64, c color.NRGBA)
	DrawText(text string, x, y float64)
}
