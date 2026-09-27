package systems

import (
	"math"

	"Sqwave/internal/domain/input"
	"Sqwave/internal/domain/world"
)

func StepPlayer(w *world.World, in input.PlayerInput) {
	p := &w.Player

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

	if in.Dash && p.DashCooldownTimer == 0 && (dx != 0 || dy != 0) {
		p.DashTimer = world.DashDuration
		p.DashCooldownTimer = world.DashCooldown
	}

	speed := world.PlayerSpeed
	if p.Dashing() {
		speed = world.DashSpeed
	} else if p.IsAiming() {
		speed = world.PlayerAimSpeed
	}

	moveX(w, dx*speed)
	moveY(w, dy*speed)
}

func moveX(w *world.World, dx float64) {
	if dx == 0 {
		return
	}
	p := &w.Player
	targetX := p.X + dx

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
			}
		} else {
			if wall.X+wall.W > p.X {
				continue
			}
			if limit := wall.X + wall.W; limit > targetX {
				targetX = limit
			}
		}
	}

	if targetX < 0 {
		targetX = 0
	}
	if targetX+world.PlayerSize > world.WorldWidth {
		targetX = world.WorldWidth - world.PlayerSize
	}
	p.X = targetX
}

func moveY(w *world.World, dy float64) {
	if dy == 0 {
		return
	}
	p := &w.Player
	targetY := p.Y + dy

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
			}
		} else {
			if wall.Y+wall.H > p.Y {
				continue
			}
			if limit := wall.Y + wall.H; limit > targetY {
				targetY = limit
			}
		}
	}

	if targetY < 0 {
		targetY = 0
	}
	if targetY+world.PlayerSize > world.WorldHeight {
		targetY = world.WorldHeight - world.PlayerSize
	}
	p.Y = targetY
}
