package victory

import (
	"image/color"
	"math"

	"Sqwave/internal/app/game"
	"Sqwave/internal/domain/world"
	"Sqwave/internal/infra/audio"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

const fadeTicks = 60
const victoryTrackPath = "assets/music/nfx.wav"

type Scene struct {
	game   *game.Game
	music  *audio.Player
	onExit func()
	timer  int
}

func New(g *game.Game, music *audio.Player, onExit func()) *Scene {
	if music != nil {
		_ = music.PlayLoopForce(victoryTrackPath)
	}
	return &Scene{game: g, music: music, onExit: onExit}
}

func (s *Scene) Update() error {
	s.timer++
	if s.timer < fadeTicks {
		return nil
	}
	if anyKeyJustPressed() ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
		inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
		if s.onExit != nil {
			s.onExit()
		}
	}
	return nil
}

func (s *Scene) Draw(r game.SceneRenderer) {
	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)

	if s.timer < fadeTicks {
		s.game.DrawFrozen(r)
		alpha := uint8(255 * s.timer / fadeTicks)
		r.DrawScreenRect(0, 0, sw, sh, color.NRGBA{R: 0, G: 0, B: 0, A: alpha})
		return
	}

	r.Clear(color.NRGBA{R: 5, G: 10, B: 20, A: 255})

	// Мягкое свечение снизу.
	glowPulse := 0.9 + 0.1*math.Sin(float64(s.timer)*0.05)
	glowA := uint8(40 * glowPulse)
	r.DrawScreenRect(0, sh*0.5, sw, sh*0.5,
		color.NRGBA{R: 30, G: 60, B: 120, A: glowA})

	// VICTORY.
	const scale = 6.0
	title := "VICTORY"
	tw := float64(len(title)) * 7.0 * scale
	tx := (sw - tw) / 2
	ty := 50.0

	pulse := 0.7 + 0.3*math.Sin(float64(s.timer)*0.08)
	gold := color.NRGBA{R: 255, G: 220, B: 80, A: uint8(200 + 55*pulse)}
	r.DrawScreenTextScaled(title, tx+4, ty+4, scale,
		color.NRGBA{R: 0, G: 0, B: 0, A: 200})
	r.DrawScreenTextScaled(title, tx, ty, scale, gold)

	// Stats.
	game.DrawStatsOverlay(r, s.game.Stats(), sh*0.5)

	// Hint.
	hintPulse := uint8(120 + 100*math.Abs(math.Sin(float64(s.timer)*0.06)))
	hint := color.NRGBA{R: 200, G: 200, B: 100, A: hintPulse}
	r.DrawScreenTextScaled("PRESS ANY KEY", 20, sh-60, 1.6, hint)
}

func anyKeyJustPressed() bool {
	for k := ebiten.Key(0); k < 256; k++ {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}
