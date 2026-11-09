package main

import (
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	_ "net/http/pprof"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"

	"Sqwave/internal/app"
	"Sqwave/internal/app/card"
	"Sqwave/internal/app/death"
	"Sqwave/internal/app/game"
	"Sqwave/internal/app/logo"
	"Sqwave/internal/app/menu"
	"Sqwave/internal/app/mp"
	"Sqwave/internal/app/setup"
	"Sqwave/internal/app/victory"
	"Sqwave/internal/domain/systems"
	"Sqwave/internal/domain/world"
	"Sqwave/internal/embedassets"
	"Sqwave/internal/infra/audio"
	"Sqwave/internal/infra/ebiteninput"
	"Sqwave/internal/infra/ebitenrender"
	netpkg "Sqwave/internal/net"
)

type App struct {
	current app.Scene
	mapKind world.MapKind
	music   *audio.Player
	assets  fs.FS

	// Сетевые ресурсы кооператива (закрываются при выходе в меню).
	netHost   *netpkg.Host
	netClient *netpkg.Client

	started bool
}

// closer — сцена, которой нужно освободить ресурсы при уходе.
type closer interface{ Close() }

// setScene переключает сцену, закрывая прежнюю, если та умеет.
func (a *App) setScene(s app.Scene) {
	if c, ok := a.current.(closer); ok {
		c.Close()
	}
	a.current = s
}

func (a *App) closeCoop() {
	if a.netHost != nil {
		a.netHost.Close()
		a.netHost = nil
	}
	if a.netClient != nil {
		a.netClient.Close()
		a.netClient = nil
	}
}

func (a *App) Update() error {
	if !a.started {
		a.started = true
		log.Printf("render loop started")
	}
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
	a.closeCoop()
	m := menu.New(a.music, func() { a.openSetup() }, func() { a.openCoop() }, func() { os.Exit(0) })
	m.SetGlitchMode(glitch)
	a.setScene(m)
}

// openCoop — экран выбора «хост / подключиться».
func (a *App) openCoop() {
	a.setScene(mp.NewCoopMenu(a.openHostLobby, func() { a.openJoin() }, func() { a.openMenu(false) }))
}

func (a *App) openHostLobby() {
	lobby := mp.NewHostLobby(func(h *netpkg.Host, seed int64, wavesLimit int, players []netpkg.LobbyPlayer) {
		a.startCoopHost(h, seed, wavesLimit, players)
	}, func() { a.openCoop() })
	a.setScene(lobby)
}

func (a *App) openJoin() *mp.JoinScene {
	join := mp.NewJoinScene(func(c *netpkg.Client, start netpkg.StartInfo) {
		a.startCoopGuest(c, start)
	}, func() { a.openCoop() })
	a.setScene(join)
	return join
}

// startCoopHost — мир симулируется здесь, клиенты только присылают ввод.
func (a *App) startCoopHost(h *netpkg.Host, seed int64, wavesLimit int, players []netpkg.LobbyPlayer) {
	w := world.NewWithOptions(seed, world.MapDefault, world.Options{
		WavesLimit: wavesLimit,
	})
	w.InitPlayers(len(players))

	g := game.New(w, ebiteninput.New(), a.music)
	g.EnableHost(h, players)
	a.netHost = h
	a.bindCoopCallbacks(g)
	a.setScene(g)
}

// startCoopGuest — клиент: только ввод и снапшоты.
func (a *App) startCoopGuest(c *netpkg.Client, start netpkg.StartInfo) {
	w := world.NewWithOptions(start.Seed, world.MapKind(start.MapKind), world.Options{
		WavesLimit: start.WavesLimit,
	})
	if n := len(start.Players); n > 0 {
		w.InitPlayers(n)
	} else {
		w.InitPlayers(1)
	}
	for i, p := range start.Players {
		if p.ID == c.LocalID() {
			w.Local = i
		}
	}
	g := game.New(w, ebiteninput.New(), a.music)
	g.EnableGuest(c, start)
	a.netClient = c
	a.bindCoopCallbacks(g)
	a.setScene(g)
}

// bindCoopCallbacks — смерть, победа, выход и обрыв связи в кооперативе.
func (a *App) bindCoopCallbacks(g *game.Game) {
	g.SetOnPlayerDied(func() {
		a.current = death.New(g, a.music, func() {
			a.openMenu(true)
		})
	})
	g.SetOnVictory(func() {
		a.current = victory.New(g, a.music, func() {
			a.openMenu(false)
		})
	})
	g.SetOnExitToMenu(func() {
		a.openMenu(false)
	})
	g.SetOnDisconnected(func() {
		a.openMenu(false)
	})
}

func (a *App) openSetup() {
	a.setScene(setup.New(
		func(s setup.Settings) { a.startGame(s) },
		func() { a.openMenu(false) },
	))
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

	a.setScene(g)
}

func main() {
	cleanup := setupLogging()
	defer cleanup()
	defer recoverPanic()
	run()
}

// retries — сколько раз родительские процессы уже перезапускали игру.
var retries = func() int {
	n, _ := strconv.Atoi(os.Getenv("SQWAVE_RETRY_COUNT"))
	return n
}()

func run() {
	mapName := flag.String("map", "default", "map: default|infantry|shooter|scout|medic|rammer|sniper|rocketeer|buffer|fleer|engineer|flamer|protector|tank|mg|summoner|harpooner")
	fullscreen := flag.Bool("fullscreen", false, "start in fullscreen")
	skipMenu := flag.Bool("skip-menu", false, "start directly in game (debug)")
	profiler := flag.Bool("prof", false, "enable pprof on localhost:6060")
	cardMode := flag.Bool("card", false, "show game card as splash")
	hostMode := flag.Bool("host", false, "start as co-op host (LAN)")
	joinAddr := flag.String("join", "", "join a co-op host at addr (LAN)")
	gfx := flag.String("gfx", "auto", "graphics library: auto|directx|opengl")
	soft := flag.Bool("soft", false, "software rendering (DirectX WARP on Windows)")
	noAudio := flag.Bool("noaudio", false, "disable audio")
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
	case "spinner":
		kind = world.MapTestSpinner
	default:
		fmt.Printf("unknown map %q, using default\n", *mapName)
	}

	var music *audio.Player
	if !*noAudio {
		var err error
		music, err = audio.New(embedassets.FS)
		if err != nil {
			fmt.Println("audio init failed:", err)
			music = nil
		}
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

	if *hostMode {
		a.openHostLobby()
	} else if *joinAddr != "" {
		a.openJoin().ConnectTo(*joinAddr)
	} else if *cardMode {
		a.setScene(card.New(func() {
			a.openMenu(false)
		}))
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

	if err := runGame(a, *gfx, *soft); err != nil {
		log.Printf("game failed: %v", err)

		// Аудиоустройства может не быть вовсе — тогда игра падает ещё до
		// первого кадра. Перезапускаемся без звука, это лучше, чем ничего.
		if retries < 2 {
			var extra []string
			msg := strings.ToLower(err.Error())
			switch {
			case strings.Contains(msg, "audio") && !*noAudio:
				extra = append(extra, "-noaudio")
			case strings.Contains(msg, "graphics") || strings.Contains(msg, "driver") || strings.Contains(msg, "directx") || strings.Contains(msg, "opengl"):
				if !*soft {
					extra = append(extra, "-soft")
				}
			}
			if len(extra) > 0 {
				if rerr := relaunchWith(extra...); rerr == nil {
					log.Printf("перезапуск с %v", extra)
					return
				} else {
					log.Printf("перезапуск не удался: %v", rerr)
				}
			}
		}

		reportFatal("не удалось запустить игру", err.Error())
		os.Exit(1)
	}
}

// runGame запускает игровой цикл. Если автоматический выбор графической
// библиотеки падает, под Windows пробуем программный DirectX (WARP) —
// это часто спасает на машинах со старым или битым драйвером.
// runGame запускает игровой цикл. Ровно один вызов RunGame на процесс:
// ebiten/vmguest паникует при повторной инициализации (MarkGuest дважды).
func runGame(a *App, gfx string, soft bool) error {
	if soft {
		_ = os.Setenv("EBITENGINE_DIRECTX", "warp")
	}

	lib := ebiten.GraphicsLibraryAuto
	switch gfx {
	case "directx":
		lib = ebiten.GraphicsLibraryDirectX
	case "opengl":
		lib = ebiten.GraphicsLibraryOpenGL
	}
	opts := &ebiten.RunGameOptions{GraphicsLibrary: lib}

	log.Printf("graphics: %s, soft=%v", gfx, soft)
	return ebiten.RunGameWithOptions(a, opts)
}

// relaunchWith перезапускает игру тем же exe с дополнительными флагами
// (мягкий рендер, без звука). Отдельный процесс нужен потому, что второй
// RunGame в том же процессе невозможен: ebiten/vmguest паникует.
func relaunchWith(extra ...string) error {
	exe, err := os.Executable()
	if err != nil {
		return err
	}
	args := append([]string{}, os.Args[1:]...)
	args = append(args, extra...)
	cmd := exec.Command(exe, args...)
	cmd.Env = append(os.Environ(),
		"SQWAVE_RETRY=1",
		fmt.Sprintf("SQWAVE_RETRY_COUNT=%d", retries+1))
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Start()
}
