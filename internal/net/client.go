package net

import (
	"errors"
	"fmt"
	"net"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	dialTimeout        = 5 * time.Second
	clientReadTimeout  = 10 * time.Second
	clientWriteTimeout = 10 * time.Second
	inputQueueLen      = 8
)

// ClientConfig — параметры подключения.
type ClientConfig struct {
	Addr string // "192.168.1.10:25000" или "192.168.1.10" (тогда DefaultTCPPort)
	Name string
}

// Client — подключение к хосту.
type Client struct {
	conn net.Conn

	mu      sync.Mutex
	localID uint8
	lobby   []LobbyPlayer
	events  []Event

	inCh      chan WireInput
	cmdCh     chan Command
	done      chan struct{}
	closeOnce sync.Once
}

// Connect: рукопожатие Join -> Welcome, таймаут 5 с.
func Connect(cfg ClientConfig) (*Client, error) {
	RegisterGobTypes()

	addr, err := tcpAddr(cfg.Addr)
	if err != nil {
		return nil, err
	}
	conn, err := net.DialTimeout("tcp", addr, dialTimeout)
	if err != nil {
		return nil, fmt.Errorf("net: connect to %s: %w", addr, err)
	}
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetNoDelay(true)
	}

	c := &Client{
		conn:  conn,
		inCh:  make(chan WireInput, inputQueueLen),
		cmdCh: make(chan Command, 8),
		done:  make(chan struct{}),
	}

	conn.SetDeadline(time.Now().Add(dialTimeout))
	join := &Message{Type: MsgJoin, Join: &JoinMsg{Name: cleanName(cfg.Name, "Player")}}
	err = WriteMessage(conn, join)
	var answer *Message
	if err == nil {
		answer, err = ReadMessage(conn)
	}
	if err != nil {
		conn.Close()
		return nil, fmt.Errorf("net: handshake with %s: %w", addr, err)
	}

	switch {
	case answer.Type == MsgWelcome && answer.Welcome != nil:
		c.localID = answer.Welcome.PlayerID
	case answer.Type == MsgKick && answer.Kick != nil:
		conn.Close()
		return nil, fmt.Errorf("net: host rejected: %s", answer.Kick.Reason)
	default:
		conn.Close()
		return nil, fmt.Errorf("net: unexpected host reply (type %d)", answer.Type)
	}
	conn.SetDeadline(time.Time{})

	go c.readLoop()
	go c.writeLoop()
	return c, nil
}

// tcpAddr добавляет порт по умолчанию, если его нет.
func tcpAddr(addr string) (string, error) {
	addr = strings.TrimSpace(addr)
	if addr == "" {
		return "", errors.New("net: empty address")
	}
	if host, port, err := net.SplitHostPort(addr); err == nil {
		if port == "" {
			return net.JoinHostPort(host, strconv.Itoa(DefaultTCPPort)), nil
		}
		return addr, nil
	}
	return net.JoinHostPort(addr, strconv.Itoa(DefaultTCPPort)), nil
}

// readLoop читает сообщения хоста; 10 с тишины — разрыв.
func (c *Client) readLoop() {
	for {
		c.conn.SetReadDeadline(time.Now().Add(clientReadTimeout))
		msg, err := ReadMessage(c.conn)
		if err != nil {
			if c.isClosed() {
				return
			}
			c.pushEvent(Event{Kind: EvDisconnected, Err: err.Error()})
			c.shutdown()
			return
		}
		c.handle(msg)
	}
}

func (c *Client) handle(msg *Message) {
	switch msg.Type {
	case MsgLobby:
		if msg.Lobby == nil {
			return
		}
		players := append([]LobbyPlayer(nil), msg.Lobby.Players...)
		c.mu.Lock()
		changed := !sameLobby(c.lobby, players)
		c.lobby = append([]LobbyPlayer(nil), players...)
		c.mu.Unlock()
		if changed {
			c.pushEvent(Event{Kind: EvLobby, Players: players})
		}
	case MsgStart:
		if msg.Start == nil {
			return
		}
		start := *msg.Start
		start.Players = append([]LobbyPlayer(nil), msg.Start.Players...)
		c.pushEvent(Event{Kind: EvStart, Start: &start})
	case MsgSnapshot:
		if msg.Snapshot == nil {
			return
		}
		c.pushEvent(Event{Kind: EvSnapshot, Snapshot: msg.Snapshot})
	case MsgGameOver:
		if msg.GameOver == nil {
			return
		}
		c.pushEvent(Event{Kind: EvGameOver, Over: msg.GameOver})
	case MsgKick:
		reason := "connection closed by host"
		if msg.Kick != nil && msg.Kick.Reason != "" {
			reason = "kicked: " + msg.Kick.Reason
		}
		c.pushEvent(Event{Kind: EvDisconnected, Err: reason})
		c.shutdown()
	}
}

// writeLoop пишет ввод из очереди; переполнение не блокирует игру.
func (c *Client) writeLoop() {
	for {
		select {
		case <-c.done:
			return
		case in := <-c.inCh:
			c.conn.SetWriteDeadline(time.Now().Add(clientWriteTimeout))
			msg := &Message{Type: MsgInput, Input: &InputMsg{Input: in}}
			if err := WriteMessage(c.conn, msg); err != nil {
				// Закрываем сокет: readLoop сообщит об обрыве.
				c.conn.Close()
				return
			}
		case cmd := <-c.cmdCh:
			c.conn.SetWriteDeadline(time.Now().Add(clientWriteTimeout))
			msg := &Message{Type: MsgCommand, Command: &CommandMsg{Command: cmd}}
			if err := WriteMessage(c.conn, msg); err != nil {
				c.conn.Close()
				return
			}
		}
	}
}

// SendCommand отправляет команду хосту (неблокирующе).
func (c *Client) SendCommand(cmd Command) {
	if c.isClosed() {
		return
	}
	select {
	case c.cmdCh <- cmd:
	default:
		// Команды редкие; при переполнении просто теряем старую.
		select {
		case <-c.cmdCh:
		default:
		}
		select {
		case c.cmdCh <- cmd:
		default:
		}
	}
}

// SendInput складывает ввод в очередь; при переполнении теряется самый старый.
func (c *Client) SendInput(in WireInput) {
	if c.isClosed() {
		return
	}
	select {
	case c.inCh <- in:
		return
	default:
	}
	select {
	case <-c.inCh:
	default:
	}
	select {
	case c.inCh <- in:
	default:
	}
}

func (c *Client) pushEvent(ev Event) {
	c.mu.Lock()
	defer c.mu.Unlock()
	// Если игра долго не забирает события (свёрнутое окно и т.п.),
	// снапшоты не копим: нужен только последний.
	if ev.Kind == EvSnapshot && len(c.events) > 128 {
		for i := len(c.events) - 1; i >= 0; i-- {
			if c.events[i].Kind == EvSnapshot {
				c.events[i] = ev
				return
			}
		}
	}
	c.events = append(c.events, ev)
}

// Poll отдаёт накопленные события, не блокируясь.
func (c *Client) Poll() []Event {
	c.mu.Lock()
	defer c.mu.Unlock()
	if len(c.events) == 0 {
		return nil
	}
	out := c.events
	c.events = nil
	return out
}

func (c *Client) LocalID() uint8 { return c.localID }

// Lobby — последний известный состав лобби.
func (c *Client) Lobby() []LobbyPlayer {
	c.mu.Lock()
	defer c.mu.Unlock()
	return append([]LobbyPlayer(nil), c.lobby...)
}

func (c *Client) isClosed() bool {
	select {
	case <-c.done:
		return true
	default:
		return false
	}
}

func (c *Client) shutdown() {
	c.closeOnce.Do(func() {
		close(c.done)
		c.conn.Close()
	})
}

// Close идемпотентен.
func (c *Client) Close() { c.shutdown() }

func sameLobby(a, b []LobbyPlayer) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
