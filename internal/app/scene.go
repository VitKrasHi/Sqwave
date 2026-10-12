package app

import "Sqwave/internal/app/game"

// Scene — общий интерфейс сцен приложения. Game и Menu его реализуют.
type Scene interface {
	Update() error
	Draw(r game.SceneRenderer)
}
