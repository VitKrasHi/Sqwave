package death

import (
	"image/color"
	"math"
	"math/rand"
	"time"

	"Sqwave/internal/app/game"
	"Sqwave/internal/domain/world"
	"Sqwave/internal/infra/audio"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Phase int

const (
	PhaseFreeze Phase = iota
	PhaseFadeToBlack
	PhaseBlackHold
	PhaseReveal
)

const (
	freezeTicks    = 90
	fadeTicks      = 60
	blackHoldTicks = 30
)

const deathImagePath = "assets/images/monster.jpg"
const deathTrackPath = "assets/music/monster.wav"

const failedText = "FAILED"
const failedLetterCount = len("FAILED")

type Scene struct {
	game   *game.Game
	music  *audio.Player
	onExit func()

	phase Phase
	timer int

	rng *rand.Rand

	// Телепорт букв — мгновенные прыжки.
	letterOffsets    [failedLetterCount][2]float64
	letterActive     [failedLetterCount]bool
	teleportActive   bool
	teleportTimer    int
	teleportCooldown int
	nextJumpTick     int // тик, когда буквы прыгнут снова

	// Внезапное исчезновение-появление.
	blinkActive    bool
	blinkTimer     int
	blinkNextCheck int
}

func New(g *game.Game, music *audio.Player, onExit func()) *Scene {
	return &Scene{
		game:             g,
		music:            music,
		onExit:           onExit,
		phase:            PhaseFreeze,
		timer:            0,
		rng:              rand.New(rand.NewSource(time.Now().UnixNano())),
		teleportCooldown: 120,
		blinkActive:      true,
		blinkTimer:       0,
		blinkNextCheck:   90,
	}
}

func (s *Scene) Update() error {
	s.timer++

	switch s.phase {
	case PhaseFreeze:
		if s.timer >= freezeTicks {
			s.phase = PhaseFadeToBlack
			s.timer = 0
		}

	case PhaseFadeToBlack:
		if s.timer >= fadeTicks {
			s.phase = PhaseBlackHold
			s.timer = 0
		}

	case PhaseBlackHold:
		if s.timer >= blackHoldTicks {
			s.phase = PhaseReveal
			s.timer = 0
			s.teleportCooldown = 120
			s.blinkActive = true
			s.blinkTimer = 0
			s.blinkNextCheck = 90
			if s.music != nil {
				_ = s.music.PlayLoopForce(deathTrackPath)
			}
		}

	case PhaseReveal:
		s.tickTeleport()
		s.tickBlink()

		if anyKeyJustPressed() ||
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) ||
			inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonRight) {
			if s.onExit != nil {
				s.onExit()
			}
		}
	}
	return nil
}

// tickBlink — внезапное исчезновение-появление всей надписи.
func (s *Scene) tickBlink() {
	s.blinkTimer++
	if s.blinkTimer < s.blinkNextCheck {
		return
	}

	if s.blinkActive {
		s.blinkActive = false
		s.blinkTimer = 0
		s.blinkNextCheck = 6 + s.rng.Intn(8)
	} else {
		s.blinkActive = true
		s.blinkTimer = 0
		s.blinkNextCheck = 60 + s.rng.Intn(120)
	}
}

// tickTeleport — буквы телепортируются в случайные точки,
// несколько раз дёргаются туда-сюда, потом резко возвращаются.
func (s *Scene) tickTeleport() {
	if s.teleportActive {
		s.teleportTimer++

		if s.teleportTimer >= s.nextJumpTick {
			// Новый прыжок — буквы перемещаются в свежие случайные точки.
			s.jumpLetters()

			// Интервал между прыжками: 2..6 тиков — как «лаги».
			s.nextJumpTick = s.teleportTimer + 2 + s.rng.Intn(5)
		}

		// Всего длится ~40 тиков, потом резкий возврат.
		if s.teleportTimer >= 40 {
			for i := range s.letterOffsets {
				s.letterOffsets[i][0] = 0
				s.letterOffsets[i][1] = 0
				s.letterActive[i] = false
			}
			s.teleportActive = false
			s.teleportTimer = 0
			s.teleportCooldown = 180 + s.rng.Intn(180)
		}
		return
	}

	if s.teleportCooldown > 0 {
		s.teleportCooldown--
		return
	}

	// Начинаем серию телепортов — 2 или 3 буквы.
	count := 2 + s.rng.Intn(2)
	perm := s.rng.Perm(failedLetterCount)
	for i := range s.letterActive {
		s.letterActive[i] = false
	}
	for i := 0; i < count; i++ {
		s.letterActive[perm[i]] = true
	}

	s.teleportActive = true
	s.teleportTimer = 0
	s.nextJumpTick = 0
	s.jumpLetters()
}

// jumpLetters задаёт активным буквам новые случайные смещения.
func (s *Scene) jumpLetters() {
	for i := range s.letterOffsets {
		if !s.letterActive[i] {
			continue
		}
		// Резкий разлёт: 40..150 px, угол случайный.
		angle := s.rng.Float64() * math.Pi * 2
		dist := 40 + s.rng.Float64()*110
		s.letterOffsets[i][0] = math.Cos(angle) * dist
		s.letterOffsets[i][1] = math.Sin(angle) * dist
	}
}

func (s *Scene) Draw(r game.SceneRenderer) {
	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)

	switch s.phase {
	case PhaseFreeze:
		s.game.DrawFrozen(r)

	case PhaseFadeToBlack:
		s.game.DrawFrozen(r)
		alpha := uint8(255 * s.timer / fadeTicks)
		r.DrawScreenRect(0, 0, sw, sh, color.NRGBA{R: 0, G: 0, B: 0, A: alpha})

	case PhaseBlackHold:
		r.Clear(color.NRGBA{R: 0, G: 0, B: 0, A: 255})

	case PhaseReveal:
		r.Clear(color.NRGBA{R: 0, G: 0, B: 0, A: 255})

		r.DrawScreenImage(deathImagePath, 0, 0, sw, sh)
		r.DrawScreenRect(0, 0, sw, sh, color.NRGBA{R: 0, G: 0, B: 0, A: 220})

		if s.blinkActive {
			const scale = 3.0
			charW := 7.0 * scale
			textW := float64(failedLetterCount) * charW
			baseX := (sw - textW) / 2
			baseY := 40.0

			failPulse := 0.85 + 0.15*math.Sin(float64(s.timer)*0.12)
			failAlpha := uint8(255 * failPulse)

			for i := 0; i < failedLetterCount; i++ {
				ch := string(failedText[i])
				lx := baseX + float64(i)*charW + s.letterOffsets[i][0]
				ly := baseY + s.letterOffsets[i][1]

				r.DrawScreenTextScaled(ch, lx+2, ly+2, scale,
					color.NRGBA{R: 0, G: 0, B: 0, A: failAlpha})
				r.DrawScreenTextScaled(ch, lx, ly, scale,
					color.NRGBA{R: 220, G: 20, B: 20, A: failAlpha})
			}
		}

		game.DrawStatsOverlay(r, s.game.Stats(), sh*0.5)

		hintPulse := uint8(120 + 100*math.Abs(math.Sin(float64(s.timer)*0.06)))
		hint := color.NRGBA{R: 200, G: 30, B: 30, A: hintPulse}
		r.DrawScreenTextScaled("PRESS ANY KEY", 20, sh-60, 1.6, hint)
	}
}

func anyKeyJustPressed() bool {
	for k := ebiten.Key(0); k < 256; k++ {
		if inpututil.IsKeyJustPressed(k) {
			return true
		}
	}
	return false
}
