package systems

import (
	"math"
	"math/rand"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

func StepShooting(w *world.World, in input.PlayerInput) {
	p := &w.Player
	if p.FireCooldownTimer > 0 {
		p.FireCooldownTimer--
	}
	if !in.Fire || p.FireCooldownTimer > 0 {
		return
	}

	weapon := p.Weapon.Stats()
	cx, cy := p.Center()
	aimDX := in.AimX - cx
	aimDY := in.AimY - cy
	if aimDX == 0 && aimDY == 0 {
		return
	}
	baseAngle := math.Atan2(aimDY, aimDX)

	for i := 0; i < weapon.Pellets; i++ {
		angle := baseAngle + spreadOffset(w.Rng, weapon.Spread)
		cos := math.Cos(angle)
		sin := math.Sin(angle)

		// Луч начинается на краю игрока, а не в центре — чтобы
		// визуально не проходил сквозь собственный квадрат.
		startX := cx + cos*world.PlayerSize/2
		startY := cy + sin*world.PlayerSize/2

		endX := cx + cos*weapon.Range
		endY := cy + sin*weapon.Range

		// Ищем ближайшее пересечение со стенами по параметру t.
		hitT := 1.0
		for _, wall := range w.Walls {
			if t, ok := geometry.RaySegmentIntersectsRect(startX, startY, endX, endY, wall); ok {
				if t < hitT {
					hitT = t
				}
			}
		}

		w.Bullets = append(w.Bullets, world.Bullet{
			StartX:  startX,
			StartY:  startY,
			EndX:    startX + (endX-startX)*hitT,
			EndY:    startY + (endY-startY)*hitT,
			Life:    weapon.VisualLife,
			MaxLife: weapon.VisualLife,
			Weapon:  p.Weapon,
		})
	}

	p.FireCooldownTimer = weapon.FireCooldown
}

// spreadOffset возвращает случайное отклонение в радианах
// в диапазоне [-spreadDeg/2, +spreadDeg/2].
func spreadOffset(rng *rand.Rand, spreadDeg float64) float64 {
	if spreadDeg <= 0 {
		return 0
	}
	half := spreadDeg * math.Pi / 360
	return (rng.Float64()*2 - 1) * half
}
