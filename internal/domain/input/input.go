package input

// PlayerInput — снимок намерений игрока за один тик.
// Ничего не знает про Ebiten: сюда попадают уже абстрактные флаги.
type PlayerInput struct {
	Up, Down, Left, Right bool
	Dash                  bool // «только что нажат»
	Fire                  bool // «удерживается»
	AimX, AimY            float64
}
