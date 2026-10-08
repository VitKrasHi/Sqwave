package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	_ "net/http/pprof"

	"github.com/hajimehoshi/ebiten/v2"

	"Sqwave/internal/app/game"
	"Sqwave/internal/domain/systems"
	"Sqwave/internal/domain/world"
	"Sqwave/internal/infra/ebiteninput"
	"Sqwave/internal/infra/ebitenrender"
)

type adapter struct {
	g *game.Game
}

func (a *adapter) Update() error { return a.g.Update() }
func (a *adapter) Draw(screen *ebiten.Image) {
	a.g.Draw(ebitenrender.New(screen))
}
func (a *adapter) Layout(_, _ int) (int, int) {
	return world.ScreenWidth, world.ScreenHeight
}

func main() {
	mapName := flag.String("map", "default", "map: default|infantry|shooter|scout")
	profiler := flag.Bool("prof", false, "enable pprof on localhost:6060")
	flag.Parse()

	if *profiler {
		go func() {
			log.Println("pprof: http://localhost:6060/debug/pprof/")
			_ = http.ListenAndServe("localhost:6060", nil)
		}()
	}

	kind := world.MapDefault
	switch *mapName {
	case "default":
		kind = world.MapDefault
	case "infantry":
		kind = world.MapTestInfantry
	case "shooter":
		kind = world.MapTestShooter
	case "scout":
		kind = world.MapTestScout
	case "medic":
		kind = world.MapTestMedic
	case "rammer":
		kind = world.MapTestRammer
	case "sniper":
		kind = world.MapTestSniper
	default:
		fmt.Printf("unknown map %q, using default\n", *mapName)
	}

	w := world.NewWithMap(0, kind)

	if kind != world.MapDefault {
		systems.StartTestRoom(w, kind)
	}

	g := game.New(w, ebiteninput.New())

	ebiten.SetWindowSize(world.ScreenWidth, world.ScreenHeight)
	ebiten.SetWindowTitle("Sqwave — " + *mapName)
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if err := ebiten.RunGame(&adapter{g: g}); err != nil {
		panic(err)
	}
}
