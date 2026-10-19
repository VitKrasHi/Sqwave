package systems

import (
	"math"

	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

func StepPlayer(w *world.World, in input.PlayerInput) {
	p := &w.Player

	// --- Прицел ---
	if p.HasPrevAim {
		dx := in.AimX - p.PrevAimX
		dy := in.AimY - p.PrevAimY
		p.AimPlayerSpeed = math.Hypot(dx, dy)
	} else {
		p.AimPlayerSpeed = 0
		p.HasPrevAim = true
	}
	p.PrevAimX = in.AimX
	p.PrevAimY = in.AimY
	p.AimX = in.AimX
	p.AimY = in.AimY
	p.IsFiring = in.Fire

	// --- Таймеры ---
	if p.ParryWindow > 0 {
		p.ParryWindow--
	}
	if p.ParryCooldown > 0 {
		p.ParryCooldown--
	}

	// --- Захват/гарпун ---
	if p.Held {
		valid := false
		if p.HeldBy >= 0 && p.HeldBy < len(w.Enemies) {
			other := &w.Enemies[p.HeldBy]
			if !other.IsDead() && other.Latched {
				valid = true
			}
		}
		if !valid {
			p.Held = false
			p.HeldBy = -1
		}
	}
	if p.Harpooned && p.HarpoonID != -1 {
		valid := false
		for i := range w.Harpoons {
			if w.Harpoons[i].ID == p.HarpoonID && !w.Harpoons[i].IsDead() {
				valid = true
				break
			}
		}
		if !valid {
			p.Harpooned = false
			p.HarpoonID = -1
		}
	}

	// --- Обездвиживание ---
	if p.DashActive {
		p.Held = false
		p.HeldBy = -1
		p.Harpooned = false
		p.HarpoonID = -1
		for i := range w.Enemies {
			if w.Enemies[i].Latched {
				w.Enemies[i].Latched = false
			}
		}
	} else if p.Held || p.Harpooned {
		p.DashActive = false
		p.DashActiveTimer = 0
		p.DashChargeTicks = 0
		p.VelX = 0
		p.VelY = 0
		p.PrevX = p.X
		p.PrevY = p.Y
		return
	}

	// --- Кулдаун щит-рывка ---
	if p.DashCooldown > 0 {
		p.DashCooldown--
	}

	// --- Активный щит-рывок ---
	if p.DashActive {
		blocked := moveX(w, p.DashActiveDirX*p.DashActiveSpeed)
		if moveY(w, p.DashActiveDirY*p.DashActiveSpeed) {
			blocked = true
		}
		taranCollisions(w, p)

		if blocked || p.DashActiveTimer <= 0 {
			p.DashActive = false
			p.DashActiveTimer = 0
			p.DashCooldown = world.DashCdBaseTicks +
				int(p.DashActiveStrength*float64(world.DashCdBonusTicks))
			return
		}
		p.DashActiveTimer--
		return
	}

	// --- Обычное движение ---
	var dx, dy float64
	if in.Up {
		dy--
	}
	if in.Down {
		dy++
	}
	if in.Left {
		dx--
	}
	if in.Right {
		dx++
	}
	if l := math.Hypot(dx, dy); l > 0 {
		dx /= l
		dy /= l
	}

	// Зарядка щит-рывка (ПКМ).
	if p.Weapon.IsShield() && p.DashCooldown == 0 {
		if in.DashChargeHeld {
			p.DashChargeTicks++
			if p.DashChargeTicks > world.DashChargeMaxTicks {
				p.DashChargeTicks = world.DashChargeMaxTicks
			}
		} else if p.DashChargeTicks >= world.DashChargeMinTicks {
			launchDash(w, p)
			return
		} else {
			p.DashChargeTicks = 0
		}
	} else {
		p.DashChargeTicks = 0
	}

	// Скорость.
	speed := p.MoveSpeed()
	if p.IsAiming() {
		speed = p.AimSpeed()
	}

	moveX(w, dx*speed)
	moveY(w, dy*speed)

	p.VelX = p.X - p.PrevX
	p.VelY = p.Y - p.PrevY
	p.PrevX = p.X
	p.PrevY = p.Y
}

func launchDash(w *world.World, p *world.Player) {
	pcx, pcy := p.Center()
	aimX := p.AimX - pcx
	aimY := p.AimY - pcy
	l := math.Hypot(aimX, aimY)
	if l < 0.01 {
		aimX, aimY, l = 1, 0, 1
	}
	dx := aimX / l
	dy := aimY / l

	ratio := float64(p.DashChargeTicks-world.DashChargeMinTicks) /
		float64(world.DashChargeMaxTicks-world.DashChargeMinTicks)
	if ratio < 0 {
		ratio = 0
	}
	if ratio > 1 {
		ratio = 1
	}

	p.DashActive = true
	p.DashActiveDur = world.DashDurBase + int(ratio*float64(world.DashDurBonus))
	p.DashActiveTimer = p.DashActiveDur
	p.DashActiveDirX = dx
	p.DashActiveDirY = dy
	p.DashActiveSpeed = world.DashSpeedBase + ratio*world.DashSpeedBonus
	p.DashActiveStrength = ratio
	p.DashActiveStun = world.DashStunBaseTicks +
		int(ratio*float64(world.DashStunBonusTicks))
	p.DashInstanceID++
	p.DashChargeTicks = 0
}

func taranCollisions(w *world.World, p *world.Player) {
	pRect := p.Rect()
	for i := range w.Enemies {
		e := &w.Enemies[i]
		if e.IsDead() {
			continue
		}
		if e.Type == world.EnemyTank || e.Type == world.EnemyMachineGunner {
			continue
		}
		if e.LastDashHitID == p.DashInstanceID {
			continue
		}
		if pRect.Intersects(e.Rect()) {
			e.LastDashHitID = p.DashInstanceID
			e.StunTimer = p.DashActiveStun
		}
	}
}

// ==================== Движение с коллизиями ====================

func moveX(w *world.World, dx float64) bool {
	if dx == 0 {
		return false
	}
	p := &w.Player
	targetX := p.X + dx
	blocked := false

	for _, wall := range w.Walls {
		if p.Y >= wall.Y+wall.H || p.Y+world.PlayerSize <= wall.Y {
			continue
		}
		if dx > 0 {
			if wall.X < p.X+world.PlayerSize {
				continue
			}
			if limit := wall.X - world.PlayerSize; limit < targetX {
				targetX = limit
				blocked = true
			}
		} else {
			if wall.X+wall.W > p.X {
				continue
			}
			if limit := wall.X + wall.W; limit > targetX {
				targetX = limit
				blocked = true
			}
		}
	}

	if targetX < 0 {
		targetX = 0
		blocked = true
	}
	if targetX+world.PlayerSize > world.WorldWidth {
		targetX = world.WorldWidth - world.PlayerSize
		blocked = true
	}
	p.X = targetX
	return blocked
}

func moveY(w *world.World, dy float64) bool {
	if dy == 0 {
		return false
	}
	p := &w.Player
	targetY := p.Y + dy
	blocked := false

	for _, wall := range w.Walls {
		if p.X >= wall.X+wall.W || p.X+world.PlayerSize <= wall.X {
			continue
		}
		if dy > 0 {
			if wall.Y < p.Y+world.PlayerSize {
				continue
			}
			if limit := wall.Y - world.PlayerSize; limit < targetY {
				targetY = limit
				blocked = true
			}
		} else {
			if wall.Y+wall.H > p.Y {
				continue
			}
			if limit := wall.Y + wall.H; limit > targetY {
				targetY = limit
				blocked = true
			}
		}
	}

	if targetY < 0 {
		targetY = 0
		blocked = true
	}
	if targetY+world.PlayerSize > world.WorldHeight {
		targetY = world.WorldHeight - world.PlayerSize
		blocked = true
	}
	p.Y = targetY
	return blocked
}
