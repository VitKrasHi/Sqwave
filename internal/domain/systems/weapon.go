package systems

import (
	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

func StepWeaponSelection(w *world.World, in input.PlayerInput) {
	if in.SelectWeapon1 {
		w.Player.Weapon = world.WeaponPistol
	}
	if in.SelectWeapon2 {
		w.Player.Weapon = world.WeaponSMG
	}
	if in.SelectWeapon3 {
		w.Player.Weapon = world.WeaponShotgun
	}
}
