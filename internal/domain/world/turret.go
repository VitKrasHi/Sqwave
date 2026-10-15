package world

import "Sqwave/internal/domain/geometry"

type Turret struct {
	X, Y  float64
	HP    int
	MaxHP int
	Size  float64

	Life int

	// Один счётчик кулдауна + счётчик выстрелов в очереди.
	Cooldown     int // тиков до следующего выстрела
	CooldownMax  int // кулдаун между очередями
	ShotGap      int // тиков между выстрелами внутри очереди
	BurstSize    int // сколько выстрелов в очереди
	ShotsInBurst int // сколько уже сделано

	DetectRange float64

	ProjSpeed  float64
	ProjSize   float64
	ProjLife   int
	ProjDamage int

	LastSwingHitID int
}

func (t *Turret) Rect() geometry.Rect {
	return geometry.Rect{X: t.X - t.Size/2, Y: t.Y - t.Size/2, W: t.Size, H: t.Size}
}

func (t *Turret) Damage(n int) {
	if n <= 0 {
		return
	}
	t.HP -= n
	if t.HP < 0 {
		t.HP = 0
	}
}

func (t *Turret) IsDead() bool { return t.HP <= 0 }
