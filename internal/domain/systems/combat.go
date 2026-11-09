package systems

import (
	"Sqwave/internal/domain/world"
)

// TryMeleeDamage — обёртка для melee-атаки.
// Возвращает true, если урон был заблокирован неуязвимостью.
func TryMeleeDamage(w *world.World, p *world.Player, source *world.Enemy, damage int) bool {
	invuln := p.DashActive
	if damage > 0 {
		p.TakeDamage(damage)
	}
	return invuln
}
