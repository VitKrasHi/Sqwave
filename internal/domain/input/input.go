package input

type PlayerInput struct {
	Up, Down, Left, Right bool
	Dash                  bool
	Fire                  bool // ЛКМ удерживается
	FireReleased          bool // ЛКМ только что отпущена
	SelectWeapon1         bool
	SelectWeapon2         bool
	SelectWeapon3         bool
	SelectWeapon4         bool
	AimX, AimY            float64
}
