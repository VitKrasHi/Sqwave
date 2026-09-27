package input

type PlayerInput struct {
	Up, Down, Left, Right bool
	Dash                  bool
	Fire                  bool
	FireReleased          bool
	SwapWeapon            bool
	Interact              bool
	MenuUp                bool // стрелка вверх / W
	MenuDown              bool // стрелка вниз / S
	MenuConfirm           bool // Enter / Space
	AimX, AimY            float64
}
