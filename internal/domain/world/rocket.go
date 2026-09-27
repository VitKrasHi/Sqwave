package world

import "Sqwave/internal/domain/geometry"

// Rocket — движущийся снаряд. В отличие от Bullet (hitscan-отрезок),
// летит тик за тиком, пока не столкнётся со стеной, не выйдет
// за пределы мира или не истечёт Life.
type Rocket struct {
	X, Y    float64
	VX, VY  float64
	Size    float64
	Life    int
	MaxLife int
	Weapon  WeaponType // NEW

	ExplosionRadius float64
	ExplosionLife   int
}

func (r *Rocket) Rect() geometry.Rect {
	return geometry.Rect{X: r.X - r.Size/2, Y: r.Y - r.Size/2, W: r.Size, H: r.Size}
}

// Explosion — визуальная зона взрыва. Не имеет логики урона,
// только радиус и таймер жизни.
type Explosion struct {
	X, Y    float64
	Radius  float64
	Life    int
	MaxLife int
}
