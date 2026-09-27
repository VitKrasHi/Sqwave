package game

import (
	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
	"fmt"
)

type Menu struct {
	Open      bool
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

// weaponTakenByOtherSlot — оружие уже стоит в другом слоте.
// Если picked совпадает с тем слотом, который мы НЕ заменяем — брать нельзя.
func weaponTakenByOtherSlot(p *world.Player, picked world.WeaponType) bool {
	switch {
	case p.Weapon == world.WeaponNone:
		// Активный слот пуст — любой выбор идёт туда, конфликтов нет.
		return false
	case p.Secondary == world.WeaponNone:
		// Заполняем второй слот — нельзя то же, что в активном.
		return picked == p.Weapon
	default:
		// Заменяем активный — нельзя то же, что во втором.
		return picked == p.Secondary
	}
}

func handleMenuInput(w *world.World, m *Menu, in input.PlayerInput) bool {
	if in.Interact {
		return true
	}

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

	if !in.MenuConfirm {
		return false
	}

	p := &w.Player
	picked := menuWeapons[m.Selection]

	// Занято в другом слоте — игнорируем подтверждение.
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

func drawMenu(r SceneRenderer, m *Menu, p *world.Player) {
	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)

	const (
		rowH      = 30.0
		topPad    = 16.0
		headerH   = 36.0
		slotH     = 30.0
		headerGap = 20.0
		listGap   = 24.0
		bottomPad = 30.0
		panelW    = 620.0

		// Колонки от правого края панели.
		spColFromRight = 80.0
		hpColFromRight = 170.0
	)

	contentH := topPad +
		headerH +
		slotH*3 +
		headerGap +
		listGap +
		float64(len(menuWeapons))*rowH
	panelH := contentH + bottomPad

	px := (sw - panelW) / 2
	py := (sh - panelH) / 2

	r.DrawScreenRect(px-2, py-2, panelW+4, panelH+4, ColorMenuBorder)
	r.DrawScreenRect(px, py, panelW, panelH, ColorMenuPanel)

	r.DrawScreenText("Loadout  (↑/↓ select, Enter confirm, E close)",
		px+20, py+topPad, ColorMenuText)

	slotY := py + topPad + headerH
	r.DrawScreenText("Active:  "+weaponLabel(p.Weapon), px+20, slotY, ColorMenuSelected)
	r.DrawScreenText("Backup:  "+weaponLabel(p.Secondary), px+20, slotY+slotH, ColorMenuText)

	// Итоги: HP слева, SP справа — тоже привязаны к колонкам, чтобы не слипались.
	r.DrawScreenText(fmt.Sprintf("Total HP: %d", p.MaxHP()),
		px+20, slotY+slotH*2, ColorMenuSelected)
	r.DrawScreenText(fmt.Sprintf("Total SP: %d  (speed %.1f)", p.TotalSP(), p.MoveSpeed()),
		px+20+240, slotY+slotH*2, ColorMenuSelected)

	listY := slotY + slotH*3 + headerGap + listGap
	for i, wt := range menuWeapons {
		y := listY + float64(i)*rowH

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

		hpText := fmt.Sprintf("+%d HP", wt.Stats().HPBonus)
		r.DrawScreenText(hpText, px+panelW-hpColFromRight, y, c)

		spText := fmt.Sprintf("%d SP", wt.Stats().SpeedBonus)
		r.DrawScreenText(spText, px+panelW-spColFromRight, y, c)
	}
}

func weaponLabel(w world.WeaponType) string {
	if w == world.WeaponNone {
		return "-"
	}
	return w.Stats().Name
}
