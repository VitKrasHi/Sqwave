package world

import "Sqwave/internal/domain/geometry"

const (
	CoinSize         = 8.0
	CoinLifetime     = 1200 // 20 сек
	CoinPickupRadius = 20.0 // радиус автоподбора
	CoinAttractRange = 140.0
	CoinAttractSpeed = 6.0
	CoinBossBonus    = 3 // монет за боссов (Tank, MG, Summoner, Flamer), обычные дают 1
)

type Coin struct {
	X, Y float64
	Life int
	// Value — сколько монет даст при подборе (1 обычный, 3 за босса).
	Value int
}

func (c *Coin) Rect() geometry.Rect {
	return geometry.Rect{
		X: c.X - CoinSize/2,
		Y: c.Y - CoinSize/2,
		W: CoinSize,
		H: CoinSize,
	}
}
