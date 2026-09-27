package world

// Bullet — визуальный след выстрела. Не летит, а сразу возникает
// на всей длине от Start до End. Живёт VisualLife тиков и гаснет.
type Bullet struct {
	StartX, StartY float64
	EndX, EndY     float64
	Life           int
	MaxLife        int
	Weapon         WeaponType
}
