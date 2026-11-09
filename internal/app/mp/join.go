package mp

import (
	"fmt"
	"image/color"

	"Sqwave/internal/app/game"
	netpkg "Sqwave/internal/net"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type connectResult struct {
	client *netpkg.Client
	err    error
}

// JoinScene — поиск хоста в локальной сети, ручной ввод адреса,
// ожидание старта матча.
type JoinScene struct {
	listener *netpkg.Listener
	servers  []netpkg.ServerInfo
	selected int

	manual bool
	addr   string
	name   string

	client     *netpkg.Client
	connCh     chan connectResult
	connecting bool
	waitStart  bool
	errMsg     string

	onStart func(c *netpkg.Client, start netpkg.StartInfo)
	onBack  func()
}

func NewJoinScene(onStart func(c *netpkg.Client, start netpkg.StartInfo), onBack func()) *JoinScene {
	s := &JoinScene{
		name:    "Player",
		addr:    "",
		onStart: onStart,
		onBack:  onBack,
		connCh:  make(chan connectResult, 1),
	}
	if l, err := netpkg.NewListener(); err == nil {
		s.listener = l
	} else {
		s.errMsg = "Discovery недоступен: " + err.Error()
	}
	return s
}

func (s *JoinScene) Close() {
	if s.listener != nil {
		s.listener.Stop()
		s.listener = nil
	}
	if s.client != nil {
		s.client.Close()
		s.client = nil
	}
}

func (s *JoinScene) Update() error {
	// Уже подключились — ждём старт от хоста.
	if s.client != nil {
		for _, ev := range s.client.Poll() {
			switch ev.Kind {
			case netpkg.EvStart:
				if ev.Start != nil {
					c := s.client
					s.client = nil
					s.onStart(c, *ev.Start)
					return nil
				}
			case netpkg.EvDisconnected:
				s.client = nil
				s.waitStart = false
				s.errMsg = "Host closed the connection."
			case netpkg.EvError:
				s.errMsg = ev.Err
			}
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) && !s.waitStart {
			s.Close()
			s.onBack()
		}
		return nil
	}

	// Идёт подключение.
	if s.connecting {
		select {
		case res := <-s.connCh:
			s.connecting = false
			if res.err != nil {
				s.errMsg = "Connection failed: " + res.err.Error()
				return nil
			}
			s.client = res.client
			s.waitStart = true
			s.errMsg = ""
		default:
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
			s.connecting = false
		}
		return nil
	}

	// Обновляем список серверов.
	if s.listener != nil {
		s.servers = s.listener.Servers()
	}
	if s.selected >= len(s.servers) {
		s.selected = len(s.servers) - 1
	}
	if s.selected < 0 {
		s.selected = 0
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEscape) {
		s.Close()
		s.onBack()
		return nil
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyTab) {
		s.manual = !s.manual
		return nil
	}

	if s.manual {
		for _, r := range ebiten.AppendInputChars(nil) {
			s.addr += string(r)
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(s.addr) > 0 {
			s.addr = s.addr[:len(s.addr)-1]
		}
		if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
			s.connectManual()
		}
		return nil
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyArrowDown) || inpututil.IsKeyJustPressed(ebiten.KeyS) {
		if len(s.servers) > 0 {
			s.selected = (s.selected + 1) % len(s.servers)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyArrowUp) || inpututil.IsKeyJustPressed(ebiten.KeyW) {
		if len(s.servers) > 0 {
			s.selected = (s.selected - 1 + len(s.servers)) % len(s.servers)
		}
	}
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) || inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		if len(s.servers) > 0 {
			srv := s.servers[s.selected]
			s.connect(fmt.Sprintf("%s:%d", srv.Address, srv.Port))
		}
	}
	return nil
}

func (s *JoinScene) connectManual() {
	addr := s.addr
	if addr == "" {
		return
	}
	s.connect(addr)
}

// ConnectTo — начать подключение сразу (для флага -join).
func (s *JoinScene) ConnectTo(addr string) {
	s.addr = addr
	s.connect(addr)
}

func (s *JoinScene) connect(addr string) {
	if s.connecting {
		return
	}
	s.connecting = true
	s.errMsg = ""
	name := s.name
	if name == "" {
		name = "Player"
	}
	go func() {
		c, err := netpkg.Connect(netpkg.ClientConfig{Addr: addr, Name: name})
		s.connCh <- connectResult{client: c, err: err}
	}()
}

func (s *JoinScene) Draw(r game.SceneRenderer) {
	drawFrame(r, "JOIN GAME")

	if s.client != nil {
		r.DrawScreenText("Connected. Waiting for the host to start...", 60, 160, colGood)
		r.DrawScreenText(fmt.Sprintf("Player ID: %d", s.client.LocalID()), 60, 195, colText)
		r.DrawScreenText("Esc - cancel", 60, 240, colDim)
		return
	}
	if s.connecting {
		r.DrawScreenText("Connecting...", 60, 160, colAccent)
		r.DrawScreenText("Esc — cancel", 60, 200, colDim)
		return
	}

	if s.manual {
		r.DrawScreenText("Host address (IP or IP:port):", 60, 140, colDim)
		r.DrawScreenText(s.addr+"_", 60, 175, colAccent)
		r.DrawScreenText("Enter - connect   Tab - back to LAN list   Esc - back", 60, 230, colDim)
	} else {
		r.DrawScreenText("Hosts in local network:", 60, 140, colDim)
		if len(s.servers) == 0 {
			r.DrawScreenText("Not found. Press Tab and type the host IP.", 80, 180, colDim)
		}
		for i, srv := range s.servers {
			c := color.NRGBA{R: 180, G: 195, B: 215, A: 255}
			prefix := "  "
			if i == s.selected {
				c = colAccent
				prefix = "> "
			}
			r.DrawScreenText(
				fmt.Sprintf("%s%s  %s  (%d/%d)", prefix, srv.GameName, srv.Address,
					srv.CurrentPlayers, srv.MaxPlayers),
				80, float64(180+i*30), c)
		}
		r.DrawScreenText("Enter - join   Tab - enter IP manually   Esc - back", 60, 400, colDim)
	}

	r.DrawScreenText("Your name: "+s.name, 60, 460, colText)

	if s.errMsg != "" {
		r.DrawScreenText(s.errMsg, 60, 510, colWarning)
	}
}
