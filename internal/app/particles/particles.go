package particles

import (
	"image/color"
	"math"
	"math/rand"

	"Sqwave/internal/app/game"
)

const Count = 45

type Particle struct {
	X, Y   float64
	VX, VY float64
	Size   float64
	Phase  float64
	Speed  float64
	Color  color.NRGBA

	GlitchTimer int // тиков до следующего телепорта (в glitch-режиме)
}

var palette = []color.NRGBA{
	game.ColorEnemy,
	game.ColorShooter,
	game.ColorScout,
	game.ColorMedic,
	game.ColorRammer,
	game.ColorSniper,
	game.ColorRocketeer,
}

func New(minX, maxX, minY, maxY float64, rng *rand.Rand) []Particle {
	ps := make([]Particle, Count)
	for i := range ps {
		ps[i] = Particle{
			X:           minX + rng.Float64()*(maxX-minX),
			Y:           minY + rng.Float64()*(maxY-minY),
			VX:          (rng.Float64()*2 - 1) * 1.2,
			VY:          (rng.Float64()*2 - 1) * 1.2,
			Size:        4 + rng.Float64()*6,
			Phase:       rng.Float64() * math.Pi * 2,
			Speed:       0.04 + rng.Float64()*0.04,
			Color:       palette[rng.Intn(len(palette))],
			GlitchTimer: rng.Intn(12),
		}
	}
	return ps
}

func Update(ps []Particle, minX, maxX, minY, maxY float64, glitch bool) {
	for i := range ps {
		p := &ps[i]

		if glitch {
			// Дёрганые частицы: стоят на месте, потом телепорт.
			p.GlitchTimer--
			if p.GlitchTimer <= 0 {
				p.X = minX + rand.Float64()*(maxX-minX)
				p.Y = minY + rand.Float64()*(maxY-minY)
				p.GlitchTimer = 2 + rand.Intn(12)
			}
			continue
		}

		// Плавное движение.
		p.X += p.VX
		p.Y += p.VY
		p.Phase += p.Speed

		if p.X < minX {
			p.X = minX
			p.VX = -p.VX
		}
		if p.X > maxX {
			p.X = maxX
			p.VX = -p.VX
		}
		if p.Y < minY {
			p.Y = minY
			p.VY = -p.VY
		}
		if p.Y > maxY {
			p.Y = maxY
			p.VY = -p.VY
		}
	}
}

func Draw(r game.SceneRenderer, ps []Particle) {
	const linkDist = 130.0
	for i := 0; i < len(ps); i++ {
		for j := i + 1; j < len(ps); j++ {
			dx := ps[i].X - ps[j].X
			dy := ps[i].Y - ps[j].Y
			d2 := dx*dx + dy*dy
			if d2 > linkDist*linkDist {
				continue
			}
			d := math.Sqrt(d2)
			alpha := uint8(120 * (1 - d/linkDist))
			line := color.NRGBA{R: 120, G: 180, B: 220, A: alpha}
			r.DrawScreenLine(ps[i].X, ps[i].Y, ps[j].X, ps[j].Y, 1, line)
		}
	}
	for _, p := range ps {
		pulse := 1.0 + 0.3*math.Sin(p.Phase)
		s := p.Size * pulse
		c := p.Color
		c.A = 200
		r.DrawScreenRect(p.X-s/2, p.Y-s/2, s, s, c)
	}
}
