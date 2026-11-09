package systems

import (
	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

// StepWeaponSelection обрабатывает только смену активного слота (Q).
// Выбор оружия через меню (E в зоне спавна) живёт в app-слое,
// потому что он доступен только в зоне спавна и только при открытом меню.
func StepWeaponSelection(w *world.World, idx int, in input.PlayerInput) {
	if !in.SwapWeapon || idx < 0 || idx >= len(w.Players) {
		return
	}
	p := &w.Players[idx]
	before := p.Weapon
	p.SwapWeapon()
	if p.Weapon != before {
		resetWeaponState(p)
	}
}

// resetWeaponState сбрасывает всё, что привязано к конкретному оружию:
// заряд снайперки, активный свинг меча/щита, заряд и активный рывок щита.
// Кулдаун рывка НЕ сбрасывается — он честно дотикает.
func resetWeaponState(p *world.Player) {
	p.AimCharge = 0
	p.Swing.Active = false
	p.DashActive = false
	p.DashActiveTimer = 0
	p.DashChargeTicks = 0
}
