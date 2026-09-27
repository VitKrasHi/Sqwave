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

	if p.Weapon == world.WeaponSniper {
		stepSniper(w, in)
		return
	}

	// Автоматическое оружие: сброс заряда не нужен, он и так нулевой.
	p.AimCharge = 0

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
		fireRay(w, cx, cy, baseAngle+spreadOffset(w.Rng, weapon.Spread))
	}

	p.FireCooldownTimer = weapon.FireCooldown
}

// stepSniper: зажал → копится заряд, отпустил → выстрел.
// Если отпустил без накопления (быстрый клик) — всё равно стреляет,
// но с минимальным зарядом.
func stepSniper(w *world.World, in input.PlayerInput) {
	p := &w.Player
	weapon := p.Weapon.Stats()

	switch {
	case in.Fire && p.FireCooldownTimer == 0:
		p.AimCharge++
		if p.AimCharge > weapon.ChargeTime {
			p.AimCharge = weapon.ChargeTime
		}

	case in.FireReleased && p.AimCharge > 0 && p.FireCooldownTimer == 0:
		cx, cy := p.Center()
		aimDX := in.AimX - cx
		aimDY := in.AimY - cy
		if aimDX == 0 && aimDY == 0 {
			p.AimCharge = 0
			return
		}
		baseAngle := math.Atan2(aimDY, aimDX)
		fireRay(w, cx, cy, baseAngle)

		p.AimCharge = 0
		p.FireCooldownTimer = weapon.FireCooldown

	default:
		p.AimCharge = 0
	}
}

// fireRay пускает один hitscan-луч под заданным углом и создаёт Bullet-отрезок.
func fireRay(w *world.World, cx, cy, angle float64) {
	weapon := w.Player.Weapon.Stats()
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	startX := cx + cos*world.PlayerSize/2
	startY := cy + sin*world.PlayerSize/2
	endX := cx + cos*weapon.Range
	endY := cy + sin*weapon.Range

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
		Weapon:  w.Player.Weapon,
	})
}

func spreadOffset(rng *rand.Rand, spreadDeg float64) float64 {
	if spreadDeg <= 0 {
		return 0
	}
	half := spreadDeg * math.Pi / 360
	return (rng.Float64()*2 - 1) * half
}
