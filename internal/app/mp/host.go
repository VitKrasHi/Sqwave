package mp

import (
	"fmt"
	"time"

	"Sqwave/internal/app/game"
	"Sqwave/internal/domain/world"
	netpkg "Sqwave/internal/net"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// HostLobby — экран хоста: поднимает сервер, ждёт игроков, стартует матч.
type HostLobby struct {
	host       *netpkg.Host
	players    []netpkg.LobbyPlayer
	wavesLimit int
	maxPlayers int
	errMsg     string

	onStart func(h *netpkg.Host, seed int64, wavesLimit int, players []netpkg.LobbyPlayer)
	onBack  func()
}

// NewHostLobby создаёт и сразу поднимает хост.
func NewHostLobby(onStart func(h *netpkg.Host, seed int64, wavesLimit int, players []netpkg.LobbyPlayer), onBack func()) *HostLobby {
	s := &HostLobby{
		wavesLimit: 0,
		maxPlayers: 4,
		onStart:    onStart,
		onBack:     onBack,
	}

	h := netpkg.NewHost(netpkg.HostConfig{
		Name:       "Host",
		GameName:   "Sqwave Co-op",
		MaxPlayers: s.maxPlayers,
	})
	if err := h.Start(); err != nil {
		s.errMsg = "Cannot start host: " + err.Error()
		return s
	}
	s.host = h
	s.players = h.Lobby()
	return s
}

func (s *HostLobby) Close() {
	if s.host != nil {
		s.host.Close()
		s.host = nil
	}
}

func (s *HostLobby) Update() error {
	if s.host == nil {
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) || inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			s.onBack()
		}
		return nil
	}

	// Лобби обновляем по событиям.
	for _, ev := range s.host.Poll() {
		switch ev.Kind {
		case netpkg.EvLobby:
			s.players = ev.Players
		case netpkg.EvError:
			s.errMsg = ev.Err
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.Close()
		s.onBack()
		return nil
	}

	// Лимит волн: ←/→.
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) || inpututil.IsKeyJustPressed(ebiten.KeyArrowRight) {
		if inpututil.IsKeyJustPressed(ebiten.KeyArrowLeft) {
			s.wavesLimit--
		} else {
			s.wavesLimit++
		}
		if s.wavesLimit < 0 {
			s.wavesLimit = 0
		}
		if s.wavesLimit > 99 {
			s.wavesLimit = 99
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		s.begin()
	}
	return nil
}

func (s *HostLobby) begin() {
	if s.host == nil {
		return
	}
	// Минимум двое: хост и хотя бы один гость.
	if len(s.players) < 2 {
		s.errMsg = "Need at least one guest."
		return
	}
	seed := time.Now().UnixNano()
	h := s.host
	s.host = nil // хост переходит игре, лобби его больше не закрывает
	players := h.Begin(seed, s.wavesLimit, int32(world.MapDefault))
	s.onStart(h, seed, s.wavesLimit, players)
}

func (s *HostLobby) Draw(r game.SceneRenderer) {
	drawFrame(r, "HOST GAME")

	if s.host == nil {
		r.DrawScreenText("Host is not running.", 60, 140, colWarning)
		if s.errMsg != "" {
			r.DrawScreenText(s.errMsg, 60, 170, colWarning)
		}
		r.DrawScreenText("Enter / Esc - back", 60, 220, colDim)
		return
	}

	r.DrawScreenText(fmt.Sprintf("Address: %s:%d", localIP(), netpkg.DefaultTCPPort),
		60, 130, colText)
	r.DrawScreenText(fmt.Sprintf("Players: %d / %d", len(s.players), s.maxPlayers),
		60, 160, colGood)

	// Список игроков.
	r.DrawScreenText("In lobby:", 60, 210, colDim)
	for i, p := range s.players {
		marker := "  "
		c := colText
		if i == 0 {
			marker = "* "
			c = colAccent
		}
		r.DrawScreenText(fmt.Sprintf("%s%d. %s", marker, i+1, p.Name), 80, float64(240+i*30), c)
	}
	for i := len(s.players); i < s.maxPlayers; i++ {
		r.DrawScreenText(fmt.Sprintf("   %d. —", i+1), 80, float64(240+i*30), colDim)
	}

	// Лимит волн.
	r.DrawScreenText(fmt.Sprintf("Waves limit: %s  (0 = endless, Left/Right to change)", wavesText(s.wavesLimit)),
		60, 400, colText)

	r.DrawScreenText("Enter - start   Esc - close lobby",
		60, 460, colDim)

	if s.errMsg != "" {
		r.DrawScreenText(s.errMsg, 60, 500, colWarning)
	}
}

func wavesText(n int) string {
	if n == 0 {
		return "endless"
	}
	return fmt.Sprintf("%d", n)
}
