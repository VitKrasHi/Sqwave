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

	if w.Player.Weapon != old {
		w.Player.AimCharge = 0
	}
}
