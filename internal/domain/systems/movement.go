package systems

import (
	"math"

	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

func StepPlayer(w *world.World, in input.PlayerInput) {
	p := &w.Player
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

	if p.DashCooldownTimer > 0 {
		p.DashCooldownTimer--
	}
	if p.DashTimer > 0 {
		p.DashTimer--
	}

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

	// Рывок — только с щитом в руках.
	if in.Dash && p.DashCooldownTimer == 0 && p.Weapon.IsShield() && (dx != 0 || dy != 0) {
		p.DashTimer = world.DashDuration
		p.DashCooldownTimer = world.DashCooldown
	}

	speed := p.MoveSpeed()
	if p.Dashing() {
		speed = p.DashSpeed()
	} else if p.IsAiming() {
		speed = p.AimSpeed()
	}

	blocked := moveX(w, dx*speed)
	if moveY(w, dy*speed) {
		blocked = true
	}

	// Врезались в стену во время рывка — рывок прекращается.
	if p.Dashing() && blocked {
		p.DashTimer = 0
	}

	// Обновляем скорость — нужна снайперу для предсказания.
	p.VelX = p.X - p.PrevX
	p.VelY = p.Y - p.PrevY
	p.PrevX = p.X
	p.PrevY = p.Y
}

// moveX двигает игрока по X и возвращает true, если движение
// было ограничено стеной или границей мира.
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

// moveY — аналог для вертикали.
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
