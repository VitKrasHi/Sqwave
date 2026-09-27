package input

type PlayerInput struct {
	Up, Down, Left, Right bool
	Dash                  bool
	Fire                  bool
	SelectWeapon1         bool
	SelectWeapon2         bool
	SelectWeapon3         bool
	AimX, AimY            float64
}
