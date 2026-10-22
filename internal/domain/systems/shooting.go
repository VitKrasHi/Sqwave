package systems

import (
	"math"
	"math/rand"
	"sort"

	"Sqwave/internal/domain/geometry"
	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

// StepShooting — диспетчер стрельбы: melee, снайперка с зарядкой,
// автоматическое оружие (пистолет/автомат/дробовик/ракетница).
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
		stepPlayerSniper(w, in)
		return
	}

	// Автоматическое оружие.
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

// stepPlayerSniper — зажал ЛКМ → копится заряд, отпустил → выстрел.
func stepPlayerSniper(w *world.World, in input.PlayerInput) {
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

		chargeRatio := 0.0
		if weapon.ChargeTime > 0 {
			chargeRatio = float64(p.AimCharge) / float64(weapon.ChargeTime)
		}
		if chargeRatio > 1 {
			chargeRatio = 1
		}
		if weapon.MaxPierce > 1 {
			fireSniperPierce(w, cx, cy, baseAngle, chargeRatio)
		} else {
			fireRay(w, cx, cy, baseAngle, chargeRatio)
		}

		p.AimCharge = 0
		p.FireCooldownTimer = weapon.FireCooldown

	default:
		p.AimCharge = 0
	}
}

// stepMelee — замах меча/щита. Урон наносится каждый тик через applyMeleeHit,
// один и тот же враг получает удар не больше одного раза за замах (SwingID).
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

// fireRay пускает hitscan-луч. Ищет ближайшее попадание среди стен,
// врагов, турелей, зацепленных гарпунов и вражеских спавнов.
// Урон наносится в зависимости от того, во что попал.
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

	enemyT := wallT
	var hitEnemy *world.Enemy
	var hitTurret *world.Turret
	var hitHarpoon *world.Harpoon
	var hitSpawn *world.SpawnPoint

	// Враги.
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() {
			continue
		}
		if t, ok := geometry.RaySegmentIntersectsRect(startX, startY, maxEndX, maxEndY, e.Rect()); ok {
			if t < enemyT {
				enemyT = t
				hitEnemy = e
				hitTurret = nil
				hitHarpoon = nil
				hitSpawn = nil
			}
		}
	}

	// Турели (неуязвимы, но блокируют выстрел).
	for i := range w.Turrets {
		t := &w.Turrets[i]
		if t.IsDead() {
			continue
		}
		if tt, ok := geometry.RaySegmentIntersectsRect(startX, startY, maxEndX, maxEndY, t.Rect()); ok {
			if tt < enemyT {
				enemyT = tt
				hitTurret = t
				hitEnemy = nil
				hitHarpoon = nil
				hitSpawn = nil
			}
		}
	}

	// Зацепленные гарпуны.
	for i := range w.Harpoons {
		hp := &w.Harpoons[i]
		if hp.IsDead() || !hp.Attached {
			continue
		}
		if tt, ok := geometry.RaySegmentIntersectsRect(startX, startY, maxEndX, maxEndY, hp.Rect()); ok {
			if tt < enemyT {
				enemyT = tt
				hitHarpoon = hp
				hitEnemy = nil
				hitTurret = nil
				hitSpawn = nil
			}
		}
	}

	// Вражеские спавны.
	for i := range w.SpawnPoints {
		sp := &w.SpawnPoints[i]
		if sp.IsDead() || sp.Faction != world.SpawnFactionEnemy {
			continue
		}
		if tt, ok := geometry.RaySegmentIntersectsRect(startX, startY, maxEndX, maxEndY, sp.Rect()); ok {
			if tt < enemyT {
				enemyT = tt
				hitSpawn = sp
				hitEnemy = nil
				hitTurret = nil
				hitHarpoon = nil
			}
		}
	}

	endX := startX + (maxEndX-startX)*enemyT
	endY := startY + (maxEndY-startY)*enemyT

	if hitEnemy != nil {
		if !hitEnemy.Invulnerable && !BlocksDamageFrom(hitEnemy, startX, startY) {
			distance := enemyT * weapon.Range
			dmg := w.Player.Weapon.DamageAt(distance, chargeRatio)
			hitEnemy.TakeDamage(dmg)
		}
	} else if hitTurret != nil {
		// Турели неуязвимы до конца волны.
	} else if hitHarpoon != nil {
		dmg := w.Player.Weapon.DamageAt(0, 1)
		hitHarpoon.Damage(dmg)
		if hitHarpoon.IsDead() {
			p := &w.Player
			if p.Harpooned && p.HarpoonID == hitHarpoon.ID {
				p.Harpooned = false
				p.HarpoonID = -1
			}
		}
	} else if hitSpawn != nil {
		distance := enemyT * weapon.Range
		dmg := w.Player.Weapon.DamageAt(distance, chargeRatio)
		hitSpawn.Damage(dmg)
	}

	// Визуальный след выстрела.
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

// applyMeleeHit проверяет текущее положение клинка против врагов,
// турелей и вражеских спавнов. Клинок обрезается о стены.
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

	// Обрезка о стены.
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

	// Враги.
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() {
			continue
		}
		if e.LastSwingHitID == p.SwingID {
			continue
		}
		if e.Invulnerable {
			continue
		}
		if BlocksDamageFrom(e, cx, cy) {
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

			// Оглушение мечом — работает на всех врагов.
			if weapon.StunOnHit > 0 {
				if e.StunTimer < weapon.StunOnHit {
					e.StunTimer = weapon.StunOnHit
				}
			}

			// i-frames после попадания мечом.
			if weapon.StunOnHit > 0 {
				if p.MeleeIFrames < world.MeleeHitIFrames {
					p.MeleeIFrames = world.MeleeHitIFrames
				}
			}
		}
	}

	// Турели (неуязвимы, но проверка на всякий случай если когда-нибудь
	// станут уязвимыми — просто раскомментировать t.Damage(dmg)).
	for i := range w.Turrets {
		t := &w.Turrets[i]
		if t.IsDead() {
			continue
		}
		if t.LastSwingHitID == p.SwingID {
			continue
		}
		if _, ok := geometry.RaySegmentIntersectsRect(cx, cy, endX, endY, t.Rect()); ok {
			t.LastSwingHitID = p.SwingID
			// t.Damage(weapon.DamageAt(0, 1))
		}
	}

	// Вражеские спавны.
	for i := range w.SpawnPoints {
		sp := &w.SpawnPoints[i]
		if sp.IsDead() || sp.Faction != world.SpawnFactionEnemy {
			continue
		}
		if _, ok := geometry.RaySegmentIntersectsRect(cx, cy, endX, endY, sp.Rect()); ok {
			dmg := p.Weapon.DamageAt(0, 1)
			sp.Damage(dmg)
		}
	}
}

// spawnRocket создаёт движущийся снаряд ракетницы.
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

// spreadOffset — случайный угол в диапазоне ±spreadDeg/2.
func spreadOffset(rng *rand.Rand, spreadDeg float64) float64 {
	if spreadDeg <= 0 {
		return 0
	}
	half := spreadDeg * math.Pi / 360
	return (rng.Float64()*2 - 1) * half
}

// fireSniperPierce — выстрел снайперки, пробивающий несколько врагов.
// Количество пробитых = 1 + AimChargeRatio * (MaxPierce - 1).
// Каждый следующий враг получает меньше урона (-PierceDamageFalloff от базы).
func fireSniperPierce(w *world.World, cx, cy, angle, chargeRatio float64) {
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

	// Препятствия, которые блокируют прошив (спавны, турели).
	obstacleT := wallT
	var hitSpawn *world.SpawnPoint
	for i := range w.SpawnPoints {
		sp := &w.SpawnPoints[i]
		if sp.IsDead() || sp.Faction != world.SpawnFactionEnemy {
			continue
		}
		if t, ok := geometry.RaySegmentIntersectsRect(startX, startY, maxEndX, maxEndY, sp.Rect()); ok {
			if t < obstacleT {
				obstacleT = t
				hitSpawn = sp
			}
		}
	}
	for i := range w.Turrets {
		t := &w.Turrets[i]
		if t.IsDead() {
			continue
		}
		if tt, ok := geometry.RaySegmentIntersectsRect(startX, startY, maxEndX, maxEndY, t.Rect()); ok {
			if tt < obstacleT {
				obstacleT = tt
				hitSpawn = nil // турель блокирует, спавн за ней не достать
			}
		}
	}

	// Собираем всех врагов на луче до препятствия.
	type hitTarget struct {
		e *world.Enemy
		t float64
	}
	var hits []hitTarget
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() {
			continue
		}
		if t, ok := geometry.RaySegmentIntersectsRect(startX, startY, maxEndX, maxEndY, e.Rect()); ok {
			if t < obstacleT {
				hits = append(hits, hitTarget{e, t})
			}
		}
	}

	sort.Slice(hits, func(i, j int) bool { return hits[i].t < hits[j].t })

	// Сколько врагов пробьёт выстрел.
	pierceCount := 1
	if weapon.MaxPierce > 1 {
		pierceCount = 1 + int(chargeRatio*float64(weapon.MaxPierce-1))
		if pierceCount > weapon.MaxPierce {
			pierceCount = weapon.MaxPierce
		}
		if pierceCount < 1 {
			pierceCount = 1
		}
	}

	// Наносим урон по цепочке.
	endT := wallT
	pierced := 0
	for i := 0; i < len(hits) && pierced < pierceCount; i++ {
		h := hits[i]

		falloff := 1.0 - weapon.PierceDamageFalloff*float64(pierced)
		if falloff < 0.15 {
			falloff = 0.15
		}

		if !h.e.Invulnerable && !BlocksDamageFrom(h.e, startX, startY) {
			distance := h.t * weapon.Range
			baseDmg := w.Player.Weapon.DamageAt(distance, chargeRatio)
			dmg := int(float64(baseDmg) * falloff)
			if dmg < 1 {
				dmg = 1
			}
			h.e.TakeDamage(dmg)
		}

		endT = h.t
		pierced++
	}

	// Если прошив не исчерпан и есть вражеский спавн — бьём его.
	if pierced < pierceCount && hitSpawn != nil {
		distance := obstacleT * weapon.Range
		baseDmg := w.Player.Weapon.DamageAt(distance, chargeRatio)
		falloff := 1.0 - weapon.PierceDamageFalloff*float64(pierced)
		if falloff < 0.15 {
			falloff = 0.15
		}
		dmg := int(float64(baseDmg) * falloff)
		if dmg < 1 {
			dmg = 1
		}
		hitSpawn.Damage(dmg)
		endT = obstacleT
	}

	endX := startX + (maxEndX-startX)*endT
	endY := startY + (maxEndY-startY)*endT

	// Визуальный след.
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
