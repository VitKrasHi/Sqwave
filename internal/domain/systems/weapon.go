package systems

import (
	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

// StepWeaponSelection обрабатывает только смену активного слота (Q).
// Выбор оружия через меню (E в зоне спавна) живёт в app-слое,
// потому что он доступен только в зоне спавна и только при открытом меню.
func StepWeaponSelection(w *world.World, in input.PlayerInput) {
	if in.SwapWeapon {
		before := w.Player.Weapon
		w.Player.SwapWeapon()
		if w.Player.Weapon != before {
			resetWeaponState(&w.Player)
		}
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
