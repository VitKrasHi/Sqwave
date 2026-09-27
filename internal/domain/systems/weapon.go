package systems

import (
	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

func StepWeaponSelection(w *world.World, in input.PlayerInput) {
	old := w.Player.Weapon

	if in.SelectWeapon1 {
		w.Player.Weapon = world.WeaponPistol
	}
	if in.SelectWeapon2 {
		w.Player.Weapon = world.WeaponSMG
	}
	if in.SelectWeapon3 {
		w.Player.Weapon = world.WeaponShotgun
	}
	if in.SelectWeapon4 {
		w.Player.Weapon = world.WeaponSniper
	}
	if in.SelectWeapon5 {
		w.Player.Weapon = world.WeaponRocket
	}
	if in.SelectWeapon6 {
		w.Player.Weapon = world.WeaponSword
	}
	if in.SelectWeapon7 {
		w.Player.Weapon = world.WeaponShield
	}

	if w.Player.Weapon != old {
		w.Player.AimCharge = 0
		w.Player.Swing.Active = false

		// Если ушли с щита — активный рывок обрывается.
		if !w.Player.Weapon.IsShield() {
			w.Player.DashTimer = 0
		}
	}
}
