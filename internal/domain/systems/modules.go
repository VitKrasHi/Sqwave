package systems

import "Sqwave/internal/domain/world"

// DealDamageToEnemy — наносит урон врагу с учётом активного модуля.
//
//	Blood Exchange: игрок лечится на 1/4 от нанесённого урона.
//	Vulnerability: первый удар ставит метку, второй — x4 урона.
func DealDamageToEnemy(w *world.World, e *world.Enemy, dmg int) {
	if dmg <= 0 {
		return
	}

	p := &w.Player

	switch p.ActiveModule {
	case world.ModuleVulnerability:
		if p.VulnArmed {
			dmg *= world.VulnDamageMul
			p.VulnArmed = false
			p.VulnTimer = 0
		} else {
			p.VulnArmed = true
			p.VulnTimer = world.VulnMarkTicks
		}
	case world.ModuleBloodExchange:
		heal := dmg / 15
		if heal > 0 {
			p.Heal(heal)
		}
	}

	e.TakeDamage(dmg)
}

// StepModules — тикает таймеры активного модуля.
// Vulnerability: если метка не была снята за 5 сек — штраф -25% текущего HP.
func StepModules(w *world.World) {
	p := &w.Player

	if p.ActiveModule != world.ModuleVulnerability {
		return
	}

	if !p.VulnArmed {
		return
	}

	p.VulnTimer--
	if p.VulnTimer <= 0 {
		penalty := p.CurrentHP() / world.VulnPenaltyDiv
		if penalty < 1 {
			penalty = 1
		}
		p.Damage += penalty
		p.VulnArmed = false
		p.VulnTimer = 0
	}
}

// effectiveFireCooldown — кулдаун выстрела с учётом модуля Rate of Fire.
//
//	Без модуля: base.
//	С модулем: base × (2 - stacks × 0.15), clamp [base × 0.5, base × 2].
func effectiveFireCooldown(p *world.Player, base int) int {
	if p.ActiveModule != world.ModuleRateOfFire {
		return base
	}
	mult := world.RoFBaseMultiplier - float64(p.RoFStacks)*world.RoFDecayPerStack
	if mult < world.RoFMinMultiplier {
		mult = world.RoFMinMultiplier
	}
	if mult > world.RoFBaseMultiplier {
		mult = world.RoFBaseMultiplier
	}
	result := int(float64(base) * mult)
	if result < 1 {
		result = 1
	}
	return result
}

// updateRoFStacks — обновляет стек Rate of Fire после выстрела.
func updateRoFStacks(p *world.Player, hit bool) {
	if p.ActiveModule != world.ModuleRateOfFire {
		return
	}
	if hit {
		p.RoFStacks++
		if p.RoFStacks > world.RoFMaxStacks {
			p.RoFStacks = world.RoFMaxStacks
		}
	} else {
		p.RoFStacks = 0
	}
}
