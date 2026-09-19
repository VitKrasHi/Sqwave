package input

// PlayerInput — снимок намерений игрока за один тик.
// Ничего не знает про Ebiten: сюда попадают уже абстрактные флаги.
type PlayerInput struct {
	Up, Down, Left, Right bool
	Dash                  bool
	Fire                  bool
	SelectWeapon1         bool // только что нажат «1»
	SelectWeapon2         bool // только что нажат «2»
	AimX, AimY            float64
}
