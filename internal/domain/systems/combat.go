package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

// TryMeleeDamage наносит урон от melee-атаки с учётом парирования
// и рывка-неуязвимости. Возвращает true, если урон не прошёл.
func TryMeleeDamage(w *world.World, source *world.Enemy, damage int) bool {
	p := &w.Player
	if p.DashActive {
		return true
	}
	if damage <= 0 {
		return false
	}
	if p.ParryWindow > 0 && source != nil {
		// Парирование.
		// Танк и пулемётчик не оглушаются и не отталкиваются,
		// но урон всё равно блокируется.
		if source.Type == world.EnemyTank || source.Type == world.EnemyMachineGunner {
			p.ParryWindow = 0
			return true
		}

		source.StunTimer = world.ParryStunTicks

		ecx, ecy := source.Center()
		pcx, pcy := p.Center()
		dx := ecx - pcx
		dy := ecy - pcy
		l := math.Hypot(dx, dy)
		if l > 0.01 {
			source.KnockbackVX = dx / l * world.ParryKnockback
			source.KnockbackVY = dy / l * world.ParryKnockback
			source.KnockbackTimer = world.ParryKnockbackTicks
		}

		// Для тарана — сброс рывка.
		if source.Type == world.EnemyRammer {
			source.DashActive = false
			source.DashHit = false
		}

		p.ParryWindow = 0
		return true
	}
	p.TakeDamage(damage)
	return false
}
