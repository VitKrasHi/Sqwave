package setup

import (
	"fmt"
	"image/color"
	"math/rand"
	"time"

	"Sqwave/internal/app/game"
	"Sqwave/internal/app/particles"
	"Sqwave/internal/domain/world"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Settings struct {
	EnemiesPerWave int
	WavesLimit     int
}

const (
	minWaves = 0
	maxWaves = 99

	panelW = 520.0
	panelH = 500.0
)

type row int

const (
	rowWaves row = iota
	rowStart
	rowBack
	rowCount
)

type Scene struct {
	settings Settings
	onStart  func(Settings)
	onBack   func()

	selected int
	ps       []particles.Particle

	repeatDelay int
	repeatRate  int
	heldDir     int
}

func New(onStart func(Settings), onBack func()) *Scene {
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := &Scene{
		settings: Settings{
			EnemiesPerWave: 100, // фиксированное значение
			WavesLimit:     0,
		},
		onStart: onStart,
		onBack:  onBack,
	}
	s.ps = particles.New(40, float64(world.ScreenWidth)-40,
		40, float64(world.ScreenHeight)-40, rng)
	return s
}

func panelOrigin() (float64, float64) {
	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)
	return (sw - panelW) / 2, (sh - panelH) / 2
}

// rowRect — координаты строки для hit-test и отрисовки.
func (s *Scene) rowRect(idx row) (x, y, w, h float64) {
	px, py := panelOrigin()
	switch idx {
	case rowWaves:
		return px + 30, py + 160, panelW - 60, 60
	case rowStart:
		return px + 60, py + 320, panelW - 120, 50
	case rowBack:
		return px + 60, py + 390, panelW - 120, 50
	}
	return 0, 0, 0, 0
}

// Кнопки +/- по бокам от строки Waves to win.
func (s *Scene) waveMinusButtonRect() (x, y, w, h float64) {
	rx, ry, rw, _ := s.rowRect(rowWaves)
	return rx + rw - 190, ry + 10, 50, 40
}

func (s *Scene) wavePlusButtonRect() (x, y, w, h float64) {
	rx, ry, rw, _ := s.rowRect(rowWaves)
	return rx + rw - 60, ry + 10, 50, 40
}

func pointInRect(px, py, x, y, w, h float64) bool {
	return px >= x && px <= x+w && py >= y && py <= y+h
}

func (s *Scene) Update() error {
	particles.Update(s.ps, 40, float64(world.ScreenWidth)-40,
		40, float64(world.ScreenHeight)-40, false)

	// Навигация по строкам.
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		s.selected = (s.selected + 1) % int(rowCount)
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		s.selected = (s.selected - 1 + int(rowCount)) % int(rowCount)
	}

	// Мышь.
	mx, my := ebiten.CursorPosition()
	fx, fy := float64(mx), float64(my)

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		// Кнопки +/-.
		bx, by, bw, bh := s.waveMinusButtonRect()
		if pointInRect(fx, fy, bx, by, bw, bh) {
			s.selected = int(rowWaves)
			s.settings.WavesLimit--
			if s.settings.WavesLimit < minWaves {
				s.settings.WavesLimit = maxWaves
			}
			return nil
		}
		bx, by, bw, bh = s.wavePlusButtonRect()
		if pointInRect(fx, fy, bx, by, bw, bh) {
			s.selected = int(rowWaves)
			s.settings.WavesLimit++
			if s.settings.WavesLimit > maxWaves {
				s.settings.WavesLimit = minWaves
			}
			return nil
		}

		// Остальные строки.
		for i := row(0); i < rowCount; i++ {
			x, y, w, h := s.rowRect(i)
			if pointInRect(fx, fy, x, y, w, h) {
				s.selected = int(i)
				if i == rowStart {
					s.onStart(s.settings)
					return nil
				}
				if i == rowBack {
					s.onBack()
					return nil
				}
				break
			}
		}
	}

	// Клавиатура для строк.
	switch row(s.selected) {
	case rowWaves:
		s.handleWavesInput()
	case rowStart:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			s.onStart(s.settings)
		}
	case rowBack:
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
			s.onBack()
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.onBack()
	}
	return nil
}

func (s *Scene) handleWavesInput() {
	leftHeld := ebiten.IsKeyPressed(ebiten.KeyArrowLeft) || ebiten.IsKeyPressed(ebiten.KeyA)
	rightHeld := ebiten.IsKeyPressed(ebiten.KeyArrowRight) || ebiten.IsKeyPressed(ebiten.KeyD)

	dir := 0
	if leftHeld && !rightHeld {
		dir = -1
	} else if rightHeld && !leftHeld {
		dir = 1
	}

	if dir == 0 {
		s.heldDir = 0
		s.repeatDelay = 0
		s.repeatRate = 0
		return
	}

	if dir != s.heldDir {
		s.heldDir = dir
		s.repeatDelay = 15
		s.repeatRate = 4
		s.settings.WavesLimit += dir
		if s.settings.WavesLimit < minWaves {
			s.settings.WavesLimit = maxWaves
		}
		if s.settings.WavesLimit > maxWaves {
			s.settings.WavesLimit = minWaves
		}
		return
	}

	if s.repeatDelay > 0 {
		s.repeatDelay--
		return
	}
	if s.repeatRate > 0 {
		s.repeatRate--
		return
	}
	s.settings.WavesLimit += dir
	if s.settings.WavesLimit < minWaves {
		s.settings.WavesLimit = maxWaves
	}
	if s.settings.WavesLimit > maxWaves {
		s.settings.WavesLimit = minWaves
	}
	s.repeatRate = 4
}

func (s *Scene) Draw(r game.SceneRenderer) {
	r.Clear(color.NRGBA{R: 10, G: 12, B: 20, A: 255})
	particles.Draw(r, s.ps)

	px, py := panelOrigin()

	panelBg := color.NRGBA{R: 16, G: 20, B: 30, A: 240}
	border := color.NRGBA{R: 90, G: 140, B: 200, A: 200}
	r.DrawScreenRect(px-2, py-2, panelW+4, panelH+4, border)
	r.DrawScreenRect(px, py, panelW, panelH, panelBg)

	r.DrawScreenTextScaled("ARENA SETUP", px+40, py+24, 2.4,
		color.NRGBA{R: 120, G: 230, B: 255, A: 255})

	r.DrawScreenRect(px+40, py+86, panelW-80, 2,
		color.NRGBA{R: 60, G: 80, B: 110, A: 200})

	// Waves to win — с кнопками +/-.
	wavesValue := "0 (endless)"
	if s.settings.WavesLimit > 0 {
		wavesValue = fmt.Sprintf("%d", s.settings.WavesLimit)
	}
	s.drawWavesRow(r, wavesValue)

	s.drawButton(r, rowStart, "Start")
	s.drawButton(r, rowBack, "Back")

	r.DrawScreenText("UP/DOWN select   LEFT/RIGHT change   Enter confirm   Esc back",
		px+40, py+panelH-30, color.NRGBA{R: 110, G: 130, B: 160, A: 255})
}

func (s *Scene) drawWavesRow(r game.SceneRenderer, value string) {
	rx, ry, rw, _ := s.rowRect(rowWaves)
	selected := s.selected == int(rowWaves)

	if selected {
		hl := color.NRGBA{R: 40, G: 70, B: 110, A: 120}
		r.DrawScreenRect(rx, ry, rw, 60, hl)
		marker := color.NRGBA{R: 120, G: 230, B: 255, A: 255}
		r.DrawScreenRect(rx-10, ry+25, 6, 10, marker)
	}

	textCol := color.NRGBA{R: 200, G: 210, B: 230, A: 255}
	r.DrawScreenText("Waves to win", rx+10, ry, textCol)
	r.DrawScreenText(value, rx+10, ry+22, color.NRGBA{R: 120, G: 230, B: 255, A: 255})

	mx, my, mw, mh := s.waveMinusButtonRect()
	s.drawRoundButton(r, mx, my, mw, mh, "-")

	px2, py2, pw, ph := s.wavePlusButtonRect()
	s.drawRoundButton(r, px2, py2, pw, ph, "+")
}

func (s *Scene) drawRoundButton(r game.SceneRenderer, x, y, w, h float64, label string) {
	bg := color.NRGBA{R: 26, G: 32, B: 44, A: 255}
	border := color.NRGBA{R: 60, G: 80, B: 110, A: 255}

	if s.selected == int(rowWaves) {
		bg = color.NRGBA{R: 40, G: 70, B: 110, A: 255}
		border = color.NRGBA{R: 120, G: 230, B: 255, A: 255}
	}

	r.DrawScreenRect(x-2, y-2, w+4, h+4, border)
	r.DrawScreenRect(x, y, w, h, bg)

	// Центрируем символ.
	scale := 3.0
	textW := float64(len(label)) * 7.0 * scale
	r.DrawScreenTextScaled(label, x+(w-textW)/2, y+h/2-50, scale,
		color.NRGBA{R: 240, G: 250, B: 255, A: 255})
}

func (s *Scene) drawButton(r game.SceneRenderer, idx row, label string) {
	x, y, w, h := s.rowRect(idx)
	selected := int(idx) == s.selected

	var bg, border, text color.NRGBA
	if selected {
		bg = color.NRGBA{R: 40, G: 70, B: 110, A: 255}
		border = color.NRGBA{R: 120, G: 230, B: 255, A: 255}
		text = color.NRGBA{R: 240, G: 250, B: 255, A: 255}
	} else {
		bg = color.NRGBA{R: 26, G: 32, B: 44, A: 255}
		border = color.NRGBA{R: 60, G: 80, B: 110, A: 255}
		text = color.NRGBA{R: 180, G: 190, B: 210, A: 255}
	}

	r.DrawScreenRect(x-2, y-2, w+4, h+4, border)
	r.DrawScreenRect(x, y, w, h, bg)
	r.DrawScreenTextScaled(label, x+20, y, 1.8, text)
}
