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
	if !p.HasWeapon() {
		return
	}

	if p.FireCooldownTimer > 0 {
		p.FireCooldownTimer--
	}

	if p.Weapon.IsMelee() {
		stepMelee(w, in)
		return
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
		angle := baseAngle + spreadOffset(w.Rng, weapon.Spread)
		if weapon.ProjectileSpeed > 0 {
			spawnRocket(w, cx, cy, angle, weapon)
		} else {
			fireRay(w, cx, cy, angle)
		}
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

func stepMelee(w *world.World, in input.PlayerInput) {
	p := &w.Player
	weapon := p.Weapon.Stats()

	if p.Swing.Active {
		p.Swing.Timer++
		if p.Swing.Timer >= p.Swing.Duration {
			p.Swing.Active = false
			p.FireCooldownTimer = weapon.FireCooldown
		}
		return
	}

	if !in.Fire || p.FireCooldownTimer > 0 {
		return
	}

	cx, cy := p.Center()
	aimDX := in.AimX - cx
	aimDY := in.AimY - cy
	if aimDX == 0 && aimDY == 0 {
		return
	}

	baseAngle := math.Atan2(aimDY, aimDX)
	arcRad := weapon.MeleeArc * math.Pi / 180

	p.Swing = world.SwingState{
		Active:     true,
		Timer:      0,
		Duration:   weapon.SwingDuration,
		StartAngle: baseAngle - arcRad/2, // проворот начинается «левее» курсора
		ArcRadians: arcRad,
		Range:      weapon.MeleeRange,
		Weapon:     p.Weapon,
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

// spawnRocket создаёт движущийся снаряд. Life вычисляется так,
// чтобы запас хода соответствовал weapon.Range.
func spawnRocket(w *world.World, cx, cy, angle float64, weapon world.WeaponStats) {
	life := int(weapon.Range / weapon.ProjectileSpeed)
	if life <= 0 {
		life = 60
	}
	w.Rockets = append(w.Rockets, world.Rocket{
		X:               cx,
		Y:               cy,
		VX:              math.Cos(angle) * weapon.ProjectileSpeed,
		VY:              math.Sin(angle) * weapon.ProjectileSpeed,
		Size:            weapon.ProjectileSize,
		Life:            life,
		MaxLife:         life,
		ExplosionRadius: weapon.ExplosionRadius,
		ExplosionLife:   weapon.ExplosionLife,
	})
}
