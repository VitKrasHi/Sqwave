package world

import "Sqwave/internal/domain/geometry"

const (
	SpawnFactionPlayer = 0
	SpawnFactionEnemy  = 1
)

type SpawnPoint struct {
	X, Y  float64
	Size  float64
	HP    int
	MaxHP int
	Index int

	// Faction — кому принадлежит спавн: 0=игрок (Defend), 1=враг (Attack).
	Faction int
}

func (s *SpawnPoint) Rect() geometry.Rect {
	return geometry.Rect{X: s.X - s.Size/2, Y: s.Y - s.Size/2, W: s.Size, H: s.Size}
}

func (s *SpawnPoint) Center() (float64, float64) {
	return s.X, s.Y
}

func (s *SpawnPoint) IsDead() bool { return s.HP <= 0 }

func (s *SpawnPoint) Damage(n int) {
	if n <= 0 {
		return
	}
	s.HP -= n
	if s.HP < 0 {
		s.HP = 0
	}
}

func (s *SpawnPoint) HPRatio() float64 {
	if s.MaxHP <= 0 {
		return 0
	}
	r := float64(s.HP) / float64(s.MaxHP)
	if r < 0 {
		return 0
	}
	if r > 1 {
		return 1
	}
	return r
}
