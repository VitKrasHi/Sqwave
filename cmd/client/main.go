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
	"Sqwave/internal/app/card"
	"Sqwave/internal/app/death"
	"Sqwave/internal/app/game"
	"Sqwave/internal/app/logo"
	"Sqwave/internal/app/menu"
	"Sqwave/internal/app/setup"
	"Sqwave/internal/app/victory"
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

func (a *App) openMenu(glitch bool) {
	m := menu.New(a.music, func() { a.openSetup() }, func() { os.Exit(0) })
	m.SetGlitchMode(glitch)
	a.current = m
}

func (a *App) openSetup() {
	a.current = setup.New(
		func(s setup.Settings) { a.startGame(s) },
		func() { a.openMenu(false) },
	)
}

func (a *App) startGame(s setup.Settings) {
	w := world.NewWithOptions(time.Now().UnixNano(), a.mapKind, world.Options{
		WavesLimit: s.WavesLimit,
	})
	if a.mapKind != world.MapDefault {
		systems.StartTestRoom(w, a.mapKind)
	}
	g := game.New(w, ebiteninput.New(), a.music)

	g.SetOnPlayerDied(func() {
		a.current = death.New(g, a.music, func() {
			a.openMenu(true) // glitch-частицы
		})
	})

	g.SetOnVictory(func() {
		a.current = victory.New(g, a.music, func() {
			a.openMenu(false) // нормальные частицы
		})
	})

	g.SetOnExitToMenu(func() {
		a.openMenu(false)
	})

	a.current = g
}

func main() {
	mapName := flag.String("map", "default", "map: default|infantry|shooter|scout|medic|rammer|sniper|rocketeer|buffer|fleer|engineer|flamer|protector|tank|mg|summoner|harpooner")
	fullscreen := flag.Bool("fullscreen", false, "start in fullscreen")
	skipMenu := flag.Bool("skip-menu", false, "start directly in game (debug)")
	profiler := flag.Bool("prof", false, "enable pprof on localhost:6060")
	cardMode := flag.Bool("card", false, "show game card as splash")
	genBg := flag.Bool("gen-bg", false, "generate assets/background.png (1920x1080) and exit")
	genIcon := flag.Bool("gen-icon", false, "generate assets/icon.png (630x500) and exit")
	genLogo := flag.Bool("gen-logo", false, "generate logo assets (icon + wordmark) and exit")
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

	if *genIcon {
		if err := card.SaveIcon("assets/icon.png"); err != nil {
			fmt.Println("icon:", err)
			os.Exit(1)
		}
		fmt.Println("saved assets/icon.png (630x500)")
		return
	}

	if *genBg {
		if err := card.SaveBackground("assets/background.png"); err != nil {
			fmt.Println("background:", err)
			os.Exit(1)
		}
		fmt.Println("saved assets/background.png (1920x1080)")
		return
	}

	if *genLogo {
		if err := logo.SaveAvatar("assets/avatar.png"); err != nil {
			fmt.Println("avatar:", err)
			os.Exit(1)
		}
		if err := logo.SaveAvatarSquare("assets/avatar_square.png"); err != nil {
			fmt.Println("avatar square:", err)
			os.Exit(1)
		}
		if err := logo.SaveWordmark("assets/wordmark.png"); err != nil {
			fmt.Println("wordmark:", err)
			os.Exit(1)
		}
		fmt.Println("saved assets/logo/avatar.png (512x512)")
		fmt.Println("saved assets/logo/avatar_square.png (512x512)")
		fmt.Println("saved assets/logo/wordmark.png (1600x400)")
		return
	}

	a := &App{mapKind: kind, music: music, assets: embedassets.FS}

	if *cardMode {
		a.current = card.New(func() {
			a.openMenu(false)
		})
	} else if *skipMenu {
		a.startGame(setup.Settings{EnemiesPerWave: 100, WavesLimit: 0})
	} else {
		a.openMenu(false)
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
