package main

import (
	"github.com/hajimehoshi/ebiten/v2"

	"Sqwave/internal/app/game"
	"Sqwave/internal/domain/world"
	"Sqwave/internal/infra/ebiteninput"
	"Sqwave/internal/infra/ebitenrender"
)

// adapter превращает app.Game в ebiten.Game.
// Единственное место, где Draw получает *ebiten.Image и оборачивает
// его в SceneRenderer для app-слоя.
type adapter struct {
	g *game.Game
}

func (a *adapter) Update() error {
	return a.g.Update()
}

func (a *adapter) Draw(screen *ebiten.Image) {
	a.g.Draw(ebitenrender.New(screen))
}

func (a *adapter) Layout(_, _ int) (int, int) {
	return world.ScreenWidth, world.ScreenHeight
}

func main() {
	w := world.New()
	in := ebiteninput.New()
	g := game.New(w, in)

	ebiten.SetWindowSize(world.ScreenWidth, world.ScreenHeight)
	ebiten.SetWindowTitle("Sqwave")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(&adapter{g: g}); err != nil {
		panic(err)
	}
}
