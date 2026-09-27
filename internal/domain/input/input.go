package input

type PlayerInput struct {
	Up, Down, Left, Right bool
	Dash                  bool
	Fire                  bool
	FireReleased          bool
	SelectWeapon1         bool
	SelectWeapon2         bool
	SelectWeapon3         bool
	SelectWeapon4         bool
	SelectWeapon5         bool
	AimX, AimY            float64
}
