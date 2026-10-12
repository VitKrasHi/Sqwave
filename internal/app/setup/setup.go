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
	EnemiesPerWave int // минимум 10, шаг 5
	MaxWaveLevel   int // 1..10
	WavesLimit     int // 0 = бесконечные
}

const (
	minEnemies  = 10
	maxEnemies  = 300
	stepEnemies = 5

	minLevel = 1
	maxLevel = 10

	minWaves = 0
	maxWaves = 99

	panelW = 520.0
	panelH = 560.0
)

type row int

const (
	rowEnemies row = iota
	rowLevel
	rowWaves
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
			EnemiesPerWave: 100,
			MaxWaveLevel:   5,
			WavesLimit:     0,
		},
		onStart: onStart,
		onBack:  onBack,
	}
	s.ps = particles.New(40, float64(world.ScreenWidth)-40,
		40, float64(world.ScreenHeight)-40, rng)
	return s
}

// rowRect — единый источник координат: используется и в Draw, и в hit-test.
func (s *Scene) rowRect(idx row) (x, y, w, h float64) {
	px, py := panelOrigin()
	switch idx {
	case rowEnemies:
		return px + 30, py + 120, panelW - 60, 40
	case rowLevel:
		return px + 30, py + 195, panelW - 60, 40
	case rowWaves:
		return px + 30, py + 270, panelW - 60, 40
	case rowStart:
		return px + 60, py + 390, panelW - 120, 50
	case rowBack:
		return px + 60, py + 464, panelW - 120, 50
	}
	return 0, 0, 0, 0
}

func panelOrigin() (float64, float64) {
	sw := float64(world.ScreenWidth)
	sh := float64(world.ScreenHeight)
	return (sw - panelW) / 2, (sh - panelH) / 2
}

func (s *Scene) Update() error {
	particles.Update(s.ps, 40, float64(world.ScreenWidth)-40,
		40, float64(world.ScreenHeight)-40)

	// Навигация.
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
		for i := row(0); i < rowCount; i++ {
			x, y, w, h := s.rowRect(i)
			if fx >= x && fx <= x+w && fy >= y && fy <= y+h {
				s.selected = int(i)
				if i == rowStart {
					s.onStart(s.settings)
					return nil
				}
				if i == rowBack {
					s.onBack()
					return nil
				}
				// Для строк-настроек — сразу переключаем значение
				// при клике, чтобы клик был осмысленным.
				switch i {
				case rowLevel:
					s.settings.MaxWaveLevel++
					if s.settings.MaxWaveLevel > maxLevel {
						s.settings.MaxWaveLevel = minLevel
					}
				case rowWaves:
					if s.settings.WavesLimit == 0 {
						s.settings.WavesLimit = 5
					} else {
						s.settings.WavesLimit++
						if s.settings.WavesLimit > maxWaves {
							s.settings.WavesLimit = 0
						}
					}
				}
				break
			}
		}
	}

	// Клавиатура для строк.
	switch row(s.selected) {
	case rowEnemies:
		s.handleNumericInput(&s.settings.EnemiesPerWave, stepEnemies, minEnemies, maxEnemies)
	case rowLevel:
		s.handleNumericInput(&s.settings.MaxWaveLevel, 1, minLevel, maxLevel)
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

// handleNumericInput — общий обработчик для числовых настроек с автоповтором.
func (s *Scene) handleNumericInput(val *int, step, minVal, maxVal int) {
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
		*val += dir * step
		s.clamp(val, minVal, maxVal)
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
	*val += dir * step
	s.clamp(val, minVal, maxVal)
	s.repeatRate = 4
}

// handleWavesInput — особый случай: 0 = бесконечные.
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
		if s.settings.WavesLimit < 0 {
			s.settings.WavesLimit = 0
		}
		if s.settings.WavesLimit > maxWaves {
			s.settings.WavesLimit = maxWaves
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
	if s.settings.WavesLimit < 0 {
		s.settings.WavesLimit = 0
	}
	if s.settings.WavesLimit > maxWaves {
		s.settings.WavesLimit = maxWaves
	}
	s.repeatRate = 4
}

func (s *Scene) clamp(val *int, minVal, maxVal int) {
	if *val < minVal {
		*val = minVal
	}
	if *val > maxVal {
		*val = maxVal
	}
}

func (s *Scene) Draw(r game.SceneRenderer) {
	sh := float64(world.ScreenHeight)
	_ = sh

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

	wavesValue := "∞ (endless)"
	if s.settings.WavesLimit > 0 {
		wavesValue = fmt.Sprintf("%d", s.settings.WavesLimit)
	}

	s.drawNumericRow(r, rowEnemies, "Enemies per wave",
		fmt.Sprintf("%d", s.settings.EnemiesPerWave))
	s.drawNumericRow(r, rowLevel, "Max wave level",
		fmt.Sprintf("%d / %d", s.settings.MaxWaveLevel, maxLevel))
	s.drawNumericRow(r, rowWaves, "Waves to win", wavesValue)

	s.drawButton(r, rowStart, "Start")
	s.drawButton(r, rowBack, "Back")

	r.DrawScreenText("↑/↓ select   ←/→ change   Enter confirm   Esc back",
		px+40, py+panelH-30, color.NRGBA{R: 110, G: 130, B: 160, A: 255})
}

func (s *Scene) drawNumericRow(r game.SceneRenderer, idx row, label, value string) {
	x, y, _, _ := s.rowRect(idx)
	selected := int(idx) == s.selected

	// Подсветка строки.
	if selected {
		hl := color.NRGBA{R: 40, G: 70, B: 110, A: 120}
		r.DrawScreenRect(x, y, panelW-60, 40, hl)
		marker := color.NRGBA{R: 120, G: 230, B: 255, A: 255}
		r.DrawScreenRect(x-10, y+15, 6, 10, marker)
	}

	textCol := color.NRGBA{R: 200, G: 210, B: 230, A: 255}
	valueCol := color.NRGBA{R: 120, G: 230, B: 255, A: 255}
	if !selected {
		valueCol = color.NRGBA{R: 180, G: 190, B: 210, A: 255}
	}

	r.DrawScreenText(label, x+10, y+12, textCol)
	r.DrawScreenText(value, x+330, y+12, valueCol)
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
	r.DrawScreenTextScaled(label, x+20, y+12, 1.8, text)
}
