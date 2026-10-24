package systems

import (
	"math"

	"Sqwave/internal/domain/world"
)

// StepCoins — обработка монет: притягивание к игроку, подбор, таймер жизни.
func StepCoins(w *world.World) {
	p := &w.Player
	pcx, pcy := p.Center()

	alive := w.Coins[:0]
	for _, c := range w.Coins {
		c.Life--
		if c.Life <= 0 {
			continue
		}

		dx := pcx - c.X
		dy := pcy - c.Y
		d := math.Hypot(dx, dy)

		// Подбор.
		if d <= world.CoinPickupRadius {
			p.Coins += c.Value
			w.Stats.CoinsEarned += c.Value
			continue
		}

		// Притягивание.
		if d <= world.CoinAttractRange && d > 0.01 {
			speed := world.CoinAttractSpeed
			// Чем ближе — тем быстрее летит.
			scale := 1.0 + (1.0-d/world.CoinAttractRange)*1.5
			c.X += dx / d * speed * scale
			c.Y += dy / d * speed * scale
		}

		alive = append(alive, c)
	}
	w.Coins = alive
}

// SpawnCoinAt — создаёт монету в заданной точке.
func SpawnCoinAt(w *world.World, x, y float64, value int) {
	if value <= 0 {
		value = 1
	}
	w.Coins = append(w.Coins, world.Coin{
		X:     x,
		Y:     y,
		Life:  world.CoinLifetime,
		Value: value,
	})
}
