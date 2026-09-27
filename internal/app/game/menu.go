package game

import (
	"fmt"

	"Sqwave/internal/domain/world"
)

type Menu struct {
	Open bool
}

var menuWeapons = []world.WeaponType{
	world.WeaponPistol,
	world.WeaponSMG,
	world.WeaponShotgun,
	world.WeaponSniper,
	world.WeaponRocket,
	world.WeaponSword,
	world.WeaponShield,
}

func drawMenu(r SceneRenderer, current world.WeaponType) {
	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)

	// Затемнение всего экрана.
	r.DrawScreenRect(0, 0, sw, sh, ColorMenuOverlay)

	// Панель по центру.
	panelW := 420.0
	panelH := 60.0 + float64(len(menuWeapons))*32.0 + 30
	px := (sw - panelW) / 2
	py := (sh - panelH) / 2

	r.DrawScreenRect(px-2, py-2, panelW+4, panelH+4, ColorMenuBorder)
	r.DrawScreenRect(px, py, panelW, panelH, ColorMenuPanel)

	r.DrawText("Select weapon  (E to close)", px+20, py+18)

	for i, wt := range menuWeapons {
		y := py + 52 + float64(i)*32
		marker := "  "
		text := ColorMenuText
		if wt == current {
			marker = "> "
			text = ColorMenuSelected
		}
		label := fmt.Sprintf("%s%d. %s", marker, i+1, wt.Stats().Name)
		r.DrawScreenText(label, px+20, y, text)
	}
}
