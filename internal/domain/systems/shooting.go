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
			fireRay(w, cx, cy, angle, 0)
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

		chargeRatio := float64(p.AimCharge) / float64(weapon.ChargeTime)
		if chargeRatio > 1 {
			chargeRatio = 1
		}
		fireRay(w, cx, cy, baseAngle, chargeRatio)

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
		applyMeleeHit(w, p, weapon)

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

	p.SwingID++
	p.Swing = world.SwingState{
		Active:     true,
		Timer:      0,
		Duration:   weapon.SwingDuration,
		StartAngle: baseAngle - arcRad/2,
		ArcRadians: arcRad,
		Range:      weapon.MeleeRange,
		Weapon:     p.Weapon,
	}
}

// applyMeleeHit проверяет текущее положение клинка против всех врагов.
// Один замах может задеть каждого врага только один раз — за это
// отвечает LastSwingHitID. Клинок обрезается о стены: врагов за
// стеной удар не достаёт.
func applyMeleeHit(w *world.World, p *world.Player, weapon world.WeaponStats) {
	cx, cy := p.Center()
	progress := p.SwingProgress()

	var bladeAngle, reach float64
	if weapon.IsThrust {
		reach = p.Swing.Range * math.Sin(progress*math.Pi)
		bladeAngle = p.Swing.StartAngle
	} else {
		bladeAngle = p.Swing.StartAngle + p.Swing.ArcRadians*progress
		reach = p.Swing.Range
	}

	maxEndX := cx + math.Cos(bladeAngle)*reach
	maxEndY := cy + math.Sin(bladeAngle)*reach

	// Ищем ближайшее пересечение со стенами — клинок обрезается по нему.
	// Тот же приём, что и в fireRay.
	wallT := 1.0
	for _, wall := range w.Walls {
		if t, ok := geometry.RaySegmentIntersectsRect(cx, cy, maxEndX, maxEndY, wall); ok {
			if t < wallT {
				wallT = t
			}
		}
	}

	endX := cx + (maxEndX-cx)*wallT
	endY := cy + (maxEndY-cy)*wallT

	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() {
			continue
		}
		if e.LastSwingHitID == p.SwingID {
			continue
		}
		if _, ok := geometry.RaySegmentIntersectsRect(cx, cy, endX, endY, e.Rect()); !ok {
			continue
		}

		e.LastSwingHitID = p.SwingID

		if weapon.KnockbackForce > 0 {
			dx := e.X - cx
			dy := e.Y - cy
			l := math.Hypot(dx, dy)
			if l > 0.01 {
				e.KnockbackVX = dx / l * weapon.KnockbackForce
				e.KnockbackVY = dy / l * weapon.KnockbackForce
				e.KnockbackTimer = 8
			}
		} else {
			dmg := p.Weapon.DamageAt(0, 1)
			e.TakeDamage(dmg)
		}
	}
}

// fireRay пускает один hitscan-луч под заданным углом и создаёт Bullet-отрезок.
// fireRay пускает луч, ищет ближайшее попадание среди стен и врагов
// и наносит урон, если попал во врага.
func fireRay(w *world.World, cx, cy, angle, chargeRatio float64) {
	weapon := w.Player.Weapon.Stats()
	cos := math.Cos(angle)
	sin := math.Sin(angle)

	startX := cx + cos*world.PlayerSize/2
	startY := cy + sin*world.PlayerSize/2
	maxEndX := cx + cos*weapon.Range
	maxEndY := cy + sin*weapon.Range

	// Ближайшая стена.
	wallT := 1.0
	for _, wall := range w.Walls {
		if t, ok := geometry.RaySegmentIntersectsRect(startX, startY, maxEndX, maxEndY, wall); ok {
			if t < wallT {
				wallT = t
			}
		}
	}

	// Ближайший враг до стены.
	enemyT := wallT
	var hitEnemy *world.Enemy
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() {
			continue
		}
		if t, ok := geometry.RaySegmentIntersectsRect(startX, startY, maxEndX, maxEndY, e.Rect()); ok {
			if t < enemyT {
				enemyT = t
				hitEnemy = e
			}
		}
	}

	endX := startX + (maxEndX-startX)*enemyT
	endY := startY + (maxEndY-startY)*enemyT

	if hitEnemy != nil {
		distance := enemyT * weapon.Range
		dmg := w.Player.Weapon.DamageAt(distance, chargeRatio)
		hitEnemy.TakeDamage(dmg)
	}

	w.Bullets = append(w.Bullets, world.Bullet{
		StartX:  startX,
		StartY:  startY,
		EndX:    endX,
		EndY:    endY,
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
		Weapon:          w.Player.Weapon,
		ExplosionRadius: weapon.ExplosionRadius,
		ExplosionLife:   weapon.ExplosionLife,
	})
}
