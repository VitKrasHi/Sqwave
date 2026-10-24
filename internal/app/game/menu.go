package game

import (
	"fmt"
	"image/color"

	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

const (
	TabWeapons = 0
	TabFullHP  = 1
)

type Menu struct {
	Open      bool
	Tab       int
	Selection int
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

func (m *Menu) Reset(p *world.Player) {
	m.Tab = TabWeapons
	m.Selection = 0
	if p.Weapon != world.WeaponNone {
		for i, wt := range menuWeapons {
			if wt == p.Weapon {
				m.Selection = i
				break
			}
		}
	}
}

func weaponTakenByOtherSlot(p *world.Player, picked world.WeaponType) bool {
	switch {
	case p.Weapon == world.WeaponNone:
		return false
	case p.Secondary == world.WeaponNone:
		return picked == p.Weapon
	default:
		return picked == p.Secondary
	}
}

func handleMenuInput(w *world.World, m *Menu, in input.PlayerInput) bool {
	if in.Interact {
		return true
	}

	if in.MenuLeft {
		m.Tab = (m.Tab - 1 + 2) % 2
		m.Selection = 0
	}
	if in.MenuRight {
		m.Tab = (m.Tab + 1) % 2
		m.Selection = 0
	}

	switch m.Tab {
	case TabWeapons:
		if in.MenuUp {
			m.Selection--
			if m.Selection < 0 {
				m.Selection = len(menuWeapons) - 1
			}
		}
		if in.MenuDown {
			m.Selection++
			if m.Selection >= len(menuWeapons) {
				m.Selection = 0
			}
		}
		if in.MenuConfirm {
			p := &w.Player
			picked := menuWeapons[m.Selection]
			if weaponTakenByOtherSlot(p, picked) {
				return false
			}
			switch {
			case p.Weapon == world.WeaponNone:
				p.EquipPrimary(picked)
			case p.Secondary == world.WeaponNone:
				p.EquipSecondary(picked)
			default:
				p.EquipPrimary(picked)
			}
			return p.Weapon != world.WeaponNone && p.Secondary != world.WeaponNone
		}

	case TabFullHP:
		if in.MenuConfirm {
			p := &w.Player
			if p.ActiveModule == world.ModuleBloodExchange {
				return false
			}
			if p.Damage == 0 {
				return false
			}
			if p.Coins < world.FullHPPrice {
				return false
			}
			p.Coins -= world.FullHPPrice
			p.Damage = 0
		}
	}

	return false
}

func drawMenu(r SceneRenderer, m *Menu, p *world.Player) {
	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)

	r.DrawScreenRect(0, 0, sw, sh, ColorMenuOverlay)

	panelW := 520.0
	panelH := 460.0
	px := (sw - panelW) / 2
	py := (sh - panelH) / 2

	r.DrawScreenRect(px-2, py-2, panelW+4, panelH+4, ColorMenuBorder)
	r.DrawScreenRect(px, py, panelW, panelH, ColorMenuPanel)

	tabY := py + 16
	tabH := 32.0
	tabs := []struct {
		idx   int
		label string
	}{
		{TabWeapons, "Weapons"},
		{TabFullHP, "Full HP"},
	}

	const (
		tabPad = 20.0
		tabGap = 10.0
		tabN   = 2
	)
	tabW := (panelW - 2*tabPad - float64(tabN-1)*tabGap) / float64(tabN)
	for i, t := range tabs {
		tx := px + tabPad + float64(i)*(tabW+tabGap)
		var bg, border, text color.NRGBA
		if m.Tab == t.idx {
			bg = color.NRGBA{R: 40, G: 70, B: 110, A: 255}
			border = color.NRGBA{R: 120, G: 230, B: 255, A: 255}
			text = color.NRGBA{R: 240, G: 250, B: 255, A: 255}
		} else {
			bg = color.NRGBA{R: 26, G: 32, B: 44, A: 255}
			border = color.NRGBA{R: 60, G: 80, B: 110, A: 255}
			text = color.NRGBA{R: 180, G: 190, B: 210, A: 255}
		}
		r.DrawScreenRect(tx-2, tabY-2, tabW+4, tabH+4, border)
		r.DrawScreenRect(tx, tabY, tabW, tabH, bg)
		r.DrawScreenTextScaled(t.label, tx+14, tabY-8, 1.6, text)
	}

	r.DrawScreenText("LEFT/RIGHT switch tab   UP/DOWN select   Enter confirm   E close",
		px+20, py+58, ColorMenuText)

	contentY := py + 96

	switch m.Tab {
	case TabWeapons:
		drawWeaponsTab(r, m, p, px, panelW, contentY)
	case TabFullHP:
		drawFullHPTab(r, p, px, panelW, contentY)
	}

	r.DrawScreenText(fmt.Sprintf("Coins: %d", p.Coins),
		px+20, py+panelH-30, ColorMenuSelected)
}

func drawWeaponsTab(r SceneRenderer, m *Menu, p *world.Player, px, panelW, topY float64) {
	r.DrawScreenText("Active:  "+weaponLabel(p.Weapon), px+20, topY, ColorMenuSelected)
	r.DrawScreenText("Backup:  "+weaponLabel(p.Secondary), px+20, topY+22, ColorMenuText)

	listY := topY + 60
	for i, wt := range menuWeapons {
		y := listY + float64(i)*24

		marker := "  "
		c := ColorMenuText

		taken := weaponTakenByOtherSlot(p, wt)
		switch {
		case i == m.Selection && taken:
			marker = "> "
			c = ColorMenuDim
		case i == m.Selection:
			marker = "> "
			c = ColorMenuSelected
		case taken:
			c = ColorMenuDim
		}

		label := marker + wt.Stats().Name
		if taken {
			label += "  (taken)"
		}
		r.DrawScreenText(label, px+40, y, c)

		r.DrawScreenText(fmt.Sprintf("+%d HP", wt.Stats().HPBonus), px+panelW-190, y, c)
		r.DrawScreenText(fmt.Sprintf("%d SP", wt.Stats().SpeedBonus), px+panelW-90, y, c)
	}
}

func drawFullHPTab(r SceneRenderer, p *world.Player, px, panelW, topY float64) {
	curHP := p.CurrentHP()
	maxHP := p.MaxHP()
	ratio := 0.0
	if maxHP > 0 {
		ratio = float64(curHP) / float64(maxHP)
	}

	r.DrawScreenText("Full Heal", px+20, topY, ColorMenuSelected)
	r.DrawScreenText(fmt.Sprintf("Your HP: %d / %d", curHP, maxHP),
		px+20, topY+32, ColorMenuText)

	barW := panelW - 80
	barH := 20.0
	r.DrawScreenRect(px+20, topY+60, barW, barH, ColorHPBack)
	barColor := ColorHPFull
	switch {
	case ratio < 0.3:
		barColor = ColorHPLow
	case ratio < 0.6:
		barColor = ColorHPMid
	}
	if ratio > 0 {
		r.DrawScreenRect(px+20, topY+60, barW*ratio, barH, barColor)
	}

	priceY := topY + 110
	r.DrawScreenText(fmt.Sprintf("Price: %d coins", world.FullHPPrice),
		px+20, priceY, ColorMenuText)
	r.DrawScreenText(fmt.Sprintf("You have: %d coins", p.Coins),
		px+20, priceY+24, ColorMenuText)

	statusY := priceY + 70
	switch {
	case p.ActiveModule == world.ModuleBloodExchange:
		r.DrawScreenText("Cannot heal — Blood Exchange active",
			px+20, statusY, ColorHPLow)
	case p.Damage == 0:
		r.DrawScreenText("HP is full", px+20, statusY, ColorMenuDim)
	case p.Coins < world.FullHPPrice:
		need := world.FullHPPrice - p.Coins
		r.DrawScreenText(fmt.Sprintf("Need %d more coins", need),
			px+20, statusY, ColorHPLow)
	default:
		r.DrawScreenText("[Enter] Buy Full Heal",
			px+20, statusY, ColorMenuSelected)
	}
}

func weaponLabel(w world.WeaponType) string {
	if w == world.WeaponNone {
		return "-"
	}
	return w.Stats().Name
}
