package systems

import (
	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

// StepWeaponSelection обрабатывает только смену активного слота (Q).
// Выбор оружия через цифры 1–7 живёт в app-слое (меню), потому что
// он доступен только в зоне спавна и только при открытом меню.
func StepWeaponSelection(w *world.World, in input.PlayerInput) {
	if in.SwapWeapon {
		w.Player.SwapWeapon()
	}
}
