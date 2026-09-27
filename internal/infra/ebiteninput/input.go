package ebiteninput

import (
	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"Sqwave/internal/domain/input"
)

type Source struct{}

func New() *Source { return &Source{} }

func (s *Source) Poll() input.PlayerInput {
	mx, my := ebiten.CursorPosition()
	return input.PlayerInput{
		Up:    ebiten.IsKeyPressed(ebiten.KeyW),
		Down:  ebiten.IsKeyPressed(ebiten.KeyS),
		Left:  ebiten.IsKeyPressed(ebiten.KeyA),
		Right: ebiten.IsKeyPressed(ebiten.KeyD),

		Dash:         inpututil.IsKeyJustPressed(ebiten.KeySpace),
		Fire:         ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft),
		FireReleased: inpututil.IsMouseButtonJustReleased(ebiten.MouseButtonLeft),

		Interact:   inpututil.IsKeyJustPressed(ebiten.KeyE),
		SwapWeapon: inpututil.IsKeyJustPressed(ebiten.KeyQ),

		MenuUp:      inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW),
		MenuDown:    inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS),
		MenuConfirm: inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace),

		AimX: float64(mx),
		AimY: float64(my),
	}
}
