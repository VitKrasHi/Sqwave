package world

import "Sqwave/internal/domain/geometry"

type Harpoon struct {
	ID     int
	X, Y   float64
	VX, VY float64
	Size   float64
	Life   int

	HP    int
	MaxHP int

	OwnerUID int // UniqueID владельца
	Attached bool
}

func (h *Harpoon) Rect() geometry.Rect {
	return geometry.Rect{X: h.X - h.Size/2, Y: h.Y - h.Size/2, W: h.Size, H: h.Size}
}

func (h *Harpoon) Damage(n int) {
	if n <= 0 {
		return
	}
	h.HP -= n
	if h.HP < 0 {
		h.HP = 0
	}
}

func (h *Harpoon) IsDead() bool { return h.HP <= 0 }
