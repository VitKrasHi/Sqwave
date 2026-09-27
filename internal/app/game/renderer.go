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

	ColorShield      = color.NRGBA{R: 170, G: 190, B: 220, A: 255}
	ColorShieldTrail = color.NRGBA{R: 110, G: 130, B: 165, A: 170}

	// HUD
	ColorHPFull = color.NRGBA{R: 60, G: 200, B: 80, A: 255}
	ColorHPMid  = color.NRGBA{R: 230, G: 180, B: 60, A: 255}
	ColorHPLow  = color.NRGBA{R: 220, G: 60, B: 60, A: 255}
	ColorHPBack = color.NRGBA{R: 25, G: 25, B: 30, A: 230}

	ColorSpawnZone = color.NRGBA{R: 90, G: 220, B: 150, A: 255}

	// Меню
	ColorMenuOverlay  = color.NRGBA{A: 180}
	ColorMenuPanel    = color.NRGBA{R: 30, G: 32, B: 42, A: 245}
	ColorMenuBorder   = color.NRGBA{R: 90, G: 140, B: 200, A: 255}
	ColorMenuText     = color.NRGBA{R: 220, G: 220, B: 220, A: 255}
	ColorMenuSelected = color.NRGBA{R: 120, G: 230, B: 255, A: 255}

	ColorMenuDim = color.NRGBA{R: 130, G: 130, B: 140, A: 255}

	// Враги
	ColorEnemy       = color.NRGBA{R: 200, G: 60, B: 60, A: 255}
	ColorEnemySwing  = color.NRGBA{R: 255, G: 110, B: 90, A: 255}
	ColorEnemySword  = color.NRGBA{R: 255, G: 220, B: 220, A: 255}
	ColorEnemyFacing = color.NRGBA{R: 255, G: 160, B: 160, A: 200}
	ColorEnemyHPBack = color.NRGBA{R: 30, G: 15, B: 15, A: 200}
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

func meleeColor(w world.WeaponType) color.NRGBA {
	if w.IsShield() {
		return ColorShield
	}
	return ColorSword
}

func meleeTrailColor(w world.WeaponType) color.NRGBA {
	if w.IsShield() {
		return ColorShieldTrail
	}
	return ColorSwordTrail
}

type SceneRenderer interface {
	Clear(c color.NRGBA)
	SetCamera(x, y, zoom float64)

	DrawRect(x, y, w, h float64, c color.NRGBA)
	DrawLine(x1, y1, x2, y2, thickness float64, c color.NRGBA)
	DrawCircle(x, y, r float64, c color.NRGBA)

	DrawScreenRect(x, y, w, h float64, c color.NRGBA)
	DrawScreenLine(x1, y1, x2, y2, thickness float64, c color.NRGBA)
	DrawScreenText(text string, x, y float64, c color.NRGBA)

	DrawText(text string, x, y float64) // оставляем для отладочного HUD
}
