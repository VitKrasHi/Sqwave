package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"Sqwave/internal/app"
	"Sqwave/internal/app/death"
	"Sqwave/internal/app/game"
	"Sqwave/internal/app/menu"
	"Sqwave/internal/app/setup"
	"Sqwave/internal/domain/systems"
	"Sqwave/internal/domain/world"
	"Sqwave/internal/embedassets"
	"Sqwave/internal/infra/audio"
	"Sqwave/internal/infra/ebiteninput"
	"Sqwave/internal/infra/ebitenrender"
)

type App struct {
	current app.Scene
	mapKind world.MapKind
	music   *audio.Player
	assets  fs.FS
}

func (a *App) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	altEnter := (ebiten.IsKeyPressed(ebiten.KeyAltLeft) || ebiten.IsKeyPressed(ebiten.KeyAltRight)) &&
		inpututil.IsKeyJustPressed(ebiten.KeyEnter)
	if altEnter {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	return a.current.Update()
}

func (a *App) Draw(screen *ebiten.Image) {
	a.current.Draw(ebitenrender.New(screen, a.assets))
}

func (a *App) Layout(_, _ int) (int, int) {
	return world.ScreenWidth, world.ScreenHeight
}

func (a *App) openMenu() {
	a.current = menu.New(
		a.music,
		func() { a.openSetup() },
		func() { os.Exit(0) },
	)
}

func (a *App) openSetup() {
	a.current = setup.New(
		func(s setup.Settings) { a.startGame(s) },
		func() { a.openMenu() },
	)
}

func (a *App) startGame(s setup.Settings) {
	w := world.NewWithOptions(time.Now().UnixNano(), a.mapKind, world.Options{
		BaseWaveSize: s.EnemiesPerWave,
		WavesLimit:   s.WavesLimit,
	})
	if a.mapKind != world.MapDefault {
		systems.StartTestRoom(w, a.mapKind)
	}

	g := game.New(w, ebiteninput.New(), a.music)

	g.SetOnPlayerDied(func() {
		a.current = death.New(g, a.music, func() {
			a.openMenu()
		})
	})
	g.SetOnExitToMenu(func() {
		a.openMenu()
	})

	a.current = g
}

func main() {
	mapName := flag.String("map", "default", "map: default|infantry|shooter|scout|medic|rammer|sniper|rocketeer|buffer|fleer|engineer|flamer|protector|tank|mg|summoner|harpooner")
	fullscreen := flag.Bool("fullscreen", false, "start in fullscreen")
	skipMenu := flag.Bool("skip-menu", false, "start directly in game (debug)")
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
	case "rocketeer":
		kind = world.MapTestRocketeer
	case "buffer":
		kind = world.MapTestBuffer
	case "fleer":
		kind = world.MapTestFleer
	case "engineer":
		kind = world.MapTestEngineer
	case "flamer":
		kind = world.MapTestFlamer
	case "protector":
		kind = world.MapTestProtector
	case "tank":
		kind = world.MapTestTank
	case "mg":
		kind = world.MapTestMachineGunner
	case "summoner":
		kind = world.MapTestSummoner
	case "harpooner":
		kind = world.MapTestHarpooner
	default:
		fmt.Printf("unknown map %q, using default\n", *mapName)
	}

	music, err := audio.New(embedassets.FS)
	if err != nil {
		fmt.Println("audio init failed:", err)
	}

	a := &App{mapKind: kind, music: music, assets: embedassets.FS}

	if *skipMenu {
		a.startGame(setup.Settings{EnemiesPerWave: 100, WavesLimit: 0})
	} else {
		a.openMenu()
	}

	ebiten.SetWindowSize(world.ScreenWidth, world.ScreenHeight)
	ebiten.SetWindowTitle("Sqwave")
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)

	if *fullscreen {
		ebiten.SetFullscreen(true)
	}

	if err := ebiten.RunGame(a); err != nil {
		panic(err)
	}
}
