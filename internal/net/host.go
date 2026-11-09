package net

import (
	"errors"
	"fmt"
	"log"
	"net"
	"sort"
	"sync"
	"time"
)

const (
	joinTimeout      = 10 * time.Second
	hostWriteTimeout = 10 * time.Second
	lobbyInterval    = 3 * time.Second
	ctrlQueueLen     = 32
	maxPendingInputs = 8192
)

// HostConfig — настройки хоста.
type HostConfig struct {
	Name       string // имя хоста в лобби
	GameName   string // имя игры (для дискавери)
	MaxPlayers int    // 2..4, по умолчанию 4
	Port       int    // 0 => DefaultTCPPort
}

// hostPlayer — подключённый клиент.
type hostPlayer struct {
	ID   uint8
	Name string
	conn net.Conn
	ctrl chan outFrame // контрольные сообщения (надёжные)
	snap chan []byte   // только последний снапшот, ёмкость 1
	done chan struct{}
	once sync.Once
}

// outFrame — готовый кадр и признак «закрыть после отправки».
type outFrame struct {
	data  []byte
	close bool
}

func (p *hostPlayer) close() {
	p.once.Do(func() { close(p.done) })
	p.conn.Close()
}

// Host — сервер LAN-кооператива.
type Host struct {
	cfg HostConfig

	done      chan struct{}
	closeOnce sync.Once

	mu          sync.Mutex
	started     bool
	closed      bool
	listener    net.Listener
	broadcaster *Broadcaster
	players     map[uint8]*hostPlayer
	nextID      uint8
	gameStarted bool
	events      []Event
	inputs      []PlayerInput
	cmds        []Command
}

func NewHost(cfg HostConfig) *Host {
	RegisterGobTypes()
	if cfg.MaxPlayers < 2 || cfg.MaxPlayers > 4 {
		cfg.MaxPlayers = 4
	}
	if cfg.Port <= 0 || cfg.Port > 65535 {
		cfg.Port = DefaultTCPPort
	}
	cfg.Name = cleanName(cfg.Name, "Host")
	return &Host{
		cfg:     cfg,
		done:    make(chan struct{}),
		players: make(map[uint8]*hostPlayer),
	}
}

var (
	errHostClosed  = errors.New("net: хост уже закрыт")
	errHostStarted = errors.New("net: хост уже запущен")
)

// Start слушает TCP и запускает broadcast-дискавери.
func (h *Host) Start() error {
	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return errHostClosed
	}
	if h.started {
		h.mu.Unlock()
		return errHostStarted
	}
	h.started = true
	h.mu.Unlock()

	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", h.cfg.Port))
	if err != nil {
		h.mu.Lock()
		h.started = false
		h.mu.Unlock()
		return fmt.Errorf("net: listen on port %d: %w", h.cfg.Port, err)
	}

	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		ln.Close()
		return errHostClosed
	}
	h.listener = ln
	h.mu.Unlock()

	// Дискавери не критичен: без него хост работает по прямому адресу.
	bc, err := NewBroadcaster(h.discoveryInfo)
	if err != nil {
		log.Printf("net: broadcast-дискавери недоступен: %v", err)
	} else {
		bc.Start()
		h.mu.Lock()
		if h.closed {
			h.mu.Unlock()
			bc.Stop()
			ln.Close()
			return errHostClosed
		}
		h.broadcaster = bc
		h.mu.Unlock()
	}

	select {
	case <-h.done:
		return errHostClosed
	default:
	}

	go h.acceptLoop(ln)
	go h.heartbeatLoop()
	return nil
}

// discoveryInfo — актуальные данные для анонса.
func (h *Host) discoveryInfo() ServerInfo {
	return ServerInfo{
		App:            AppMagic,
		Version:        ProtocolVersion,
		Port:           h.cfg.Port,
		GameName:       h.cfg.GameName,
		MaxPlayers:     h.cfg.MaxPlayers,
		CurrentPlayers: len(h.Lobby()),
	}
}

func (h *Host) acceptLoop(ln net.Listener) {
	for {
		conn, err := ln.Accept()
		if err != nil {
			return
		}
		go h.handleConn(conn)
	}
}

// heartbeatLoop шлёт текущее лобби, чтобы клиенты не рвали связь по таймауту.
func (h *Host) heartbeatLoop() {
	t := time.NewTicker(lobbyInterval)
	defer t.Stop()
	for {
		select {
		case <-h.done:
			return
		case <-t.C:
			if h.hasClients() {
				h.broadcastLobby()
			}
		}
	}
}

func (h *Host) hasClients() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.players) > 0
}

func (h *Host) handleConn(conn net.Conn) {
	defer conn.Close()
	if tc, ok := conn.(*net.TCPConn); ok {
		tc.SetNoDelay(true)
	}

	conn.SetReadDeadline(time.Now().Add(joinTimeout))
	msg, err := ReadMessage(conn)
	if err != nil || msg.Type != MsgJoin || msg.Join == nil {
		return
	}
	conn.SetReadDeadline(time.Time{})

	name := cleanName(msg.Join.Name, "Player")

	h.mu.Lock()
	if h.closed {
		h.mu.Unlock()
		return
	}
	if h.gameStarted {
		h.mu.Unlock()
		h.reject(conn, "the game already started")
		return
	}
	// 0-й ID занят хостом, поэтому +1.
	if len(h.players)+1 >= h.cfg.MaxPlayers {
		h.mu.Unlock()
		h.reject(conn, "lobby is full")
		return
	}
	id := h.nextID
	for {
		if id == 0 {
			id = 1
		}
		if _, busy := h.players[id]; !busy {
			break
		}
		id++
	}
	h.nextID = id + 1
	p := &hostPlayer{
		ID:   id,
		Name: name,
		conn: conn,
		ctrl: make(chan outFrame, ctrlQueueLen),
		snap: make(chan []byte, 1),
		done: make(chan struct{}),
	}
	h.players[id] = p
	h.mu.Unlock()

	welcome := &Message{Type: MsgWelcome, Welcome: &WelcomeMsg{
		PlayerID:   id,
		MaxPlayers: h.cfg.MaxPlayers,
		GameName:   h.cfg.GameName,
	}}
	if err := WriteMessage(conn, welcome); err != nil {
		// Игрок ещё не был анонсирован — убираем молча.
		h.mu.Lock()
		delete(h.players, id)
		h.mu.Unlock()
		p.close()
		return
	}

	go h.sender(p)
	h.broadcastLobby()
	h.pushEvent(Event{Kind: EvLobby, Players: h.Lobby()})

	for {
		msg, err := ReadMessage(conn)
		if err != nil {
			break
		}
		switch msg.Type {
		case MsgInput:
			if msg.Input != nil {
				h.pushInput(PlayerInput{PlayerID: id, Input: msg.Input.Input})
			}
		case MsgCommand:
			if msg.Command != nil {
				cmd := msg.Command.Command
				cmd.PlayerID = id
				h.pushCommand(cmd)
			}
		}
	}
	h.dropPlayer(p)
}

// reject отправляет Kick и закрывает соединение.
func (h *Host) reject(conn net.Conn, reason string) {
	conn.SetWriteDeadline(time.Now().Add(hostWriteTimeout))
	WriteMessage(conn, &Message{Type: MsgKick, Kick: &KickMsg{Reason: reason}})
	conn.Close()
}

func (h *Host) dropPlayer(p *hostPlayer) {
	h.mu.Lock()
	_, ok := h.players[p.ID]
	if ok {
		delete(h.players, p.ID)
	}
	h.mu.Unlock()
	if !ok {
		return
	}
	p.close()
	h.broadcastLobby()
	h.pushEvent(Event{Kind: EvLobby, Players: h.Lobby()})
}

// sender — единственный писатель в сокет клиента.
func (h *Host) sender(p *hostPlayer) {
	// На выходе всегда закрываем сокет, чтобы читающая горутина проснулась.
	defer p.conn.Close()
	for {
		// Контрольные сообщения в приоритете, снапшоты — по остаточному принципу.
		select {
		case <-p.done:
			return
		case <-h.done:
			return
		case f := <-p.ctrl:
			if !h.writeFrame(p, f) {
				return
			}
		default:
		}
		select {
		case <-p.done:
			return
		case <-h.done:
			return
		case f := <-p.ctrl:
			if !h.writeFrame(p, f) {
				return
			}
		case data := <-p.snap:
			if !h.writeFrame(p, outFrame{data: data}) {
				return
			}
		}
	}
}

// writeFrame возвращает false, если соединение пора закрывать.
func (h *Host) writeFrame(p *hostPlayer, f outFrame) bool {
	p.conn.SetWriteDeadline(time.Now().Add(hostWriteTimeout))
	if err := writeFrame(p.conn, f.data); err != nil {
		p.conn.Close() // разбудить читающую горутину
		return false
	}
	if f.close {
		p.conn.Close()
		return false
	}
	return true
}

// enqueueAll кладёт кадр всем клиентам; отстающего клиента отключаем.
func (h *Host) enqueueAll(f outFrame) {
	h.mu.Lock()
	players := make([]*hostPlayer, 0, len(h.players))
	for _, p := range h.players {
		players = append(players, p)
	}
	h.mu.Unlock()

	for _, p := range players {
		select {
		case p.ctrl <- f:
		default:
			p.conn.Close()
		}
	}
}

func (h *Host) broadcastLobby() {
	frame, err := encodeMessage(&Message{Type: MsgLobby, Lobby: &LobbyMsg{Players: h.Lobby()}})
	if err != nil {
		log.Printf("net: лобби не закодировано: %v", err)
		return
	}
	h.enqueueAll(outFrame{data: frame})
}

func (h *Host) pushEvent(ev Event) {
	h.mu.Lock()
	h.events = append(h.events, ev)
	h.mu.Unlock()
}

func (h *Host) pushInput(in PlayerInput) {
	h.mu.Lock()
	if len(h.inputs) < maxPendingInputs {
		h.inputs = append(h.inputs, in)
	}
	h.mu.Unlock()
}

// pushCommand складывает команду клиента в очередь хоста.
func (h *Host) pushCommand(cmd Command) {
	h.mu.Lock()
	if len(h.cmds) < maxPendingInputs {
		h.cmds = append(h.cmds, cmd)
	}
	h.mu.Unlock()
}

// Commands — забрать накопленные команды клиентов (и очистить буфер).
func (h *Host) Commands() []Command {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.cmds) == 0 {
		return nil
	}
	out := h.cmds
	h.cmds = nil
	return out
}

// LocalID — ID локального (хост-)игрока, всегда 0.
func (h *Host) LocalID() uint8 { return 0 }

// Lobby — состав лобби в стабильном порядке по ID, включая хоста.
func (h *Host) Lobby() []LobbyPlayer {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.lobbyLocked()
}

func (h *Host) lobbyLocked() []LobbyPlayer {
	out := make([]LobbyPlayer, 0, len(h.players)+1)
	out = append(out, LobbyPlayer{ID: 0, Name: h.cfg.Name})
	ids := make([]uint8, 0, len(h.players))
	for id := range h.players {
		ids = append(ids, id)
	}
	sort.Slice(ids, func(i, j int) bool { return ids[i] < ids[j] })
	for _, id := range ids {
		out = append(out, LobbyPlayer{ID: id, Name: h.players[id].Name})
	}
	return out
}

// Poll отдаёт накопленные события, не блокируясь.
func (h *Host) Poll() []Event {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.events) == 0 {
		return nil
	}
	out := h.events
	h.events = nil
	return out
}

// Inputs отдаёт и очищает ввод клиентов, накопленный с прошлого вызова.
// Ввод локального игрока сюда не попадает.
func (h *Host) Inputs() []PlayerInput {
	h.mu.Lock()
	defer h.mu.Unlock()
	if len(h.inputs) == 0 {
		return nil
	}
	out := h.inputs
	h.inputs = nil
	return out
}

// Begin рассылает StartInfo всем клиентам и закрывает лобби
// (подключиться после старта нельзя). Возвращает состав, который ушёл в Start.
func (h *Host) Begin(seed int64, wavesLimit int, mapKind int32) []LobbyPlayer {
	h.mu.Lock()
	h.gameStarted = true
	h.mu.Unlock()

	players := h.Lobby()
	start := &StartInfo{
		Seed:       seed,
		WavesLimit: wavesLimit,
		MapKind:    mapKind,
		Players:    players,
	}
	frame, err := encodeMessage(&Message{Type: MsgStart, Start: start})
	if err != nil {
		log.Printf("net: старт не закодирован: %v", err)
		return players
	}
	h.enqueueAll(outFrame{data: frame})
	return players
}

// BroadcastSnapshot кодирует снапшот один раз на всех и кладёт каждому
// только последний вариант (latest wins).
func (h *Host) BroadcastSnapshot(tick uint32, data []byte) {
	frame, err := encodeMessage(&Message{Type: MsgSnapshot, Snapshot: &SnapshotMsg{Tick: tick, Data: data}})
	if err != nil {
		log.Printf("net: снапшот не закодирован: %v", err)
		return
	}

	h.mu.Lock()
	players := make([]*hostPlayer, 0, len(h.players))
	for _, p := range h.players {
		players = append(players, p)
	}
	h.mu.Unlock()

	for _, p := range players {
		// Вытесняем устаревший снапшот, если он ещё не ушёл.
		select {
		case <-p.snap:
		default:
		}
		select {
		case p.snap <- frame:
		default:
		}
	}
}

func (h *Host) SendGameOver(victory bool, data []byte) {
	frame, err := encodeMessage(&Message{Type: MsgGameOver, GameOver: &GameOverMsg{Victory: victory, Data: data}})
	if err != nil {
		log.Printf("net: GameOver не закодирован: %v", err)
		return
	}
	h.enqueueAll(outFrame{data: frame})
}

// Kick отключает клиента и сообщает причину.
func (h *Host) Kick(id uint8, reason string) {
	if id == 0 {
		return
	}
	h.mu.Lock()
	p, ok := h.players[id]
	if ok {
		delete(h.players, id)
	}
	h.mu.Unlock()
	if !ok {
		return
	}

	h.broadcastLobby()
	h.pushEvent(Event{Kind: EvLobby, Players: h.Lobby()})

	frame, err := encodeMessage(&Message{Type: MsgKick, Kick: &KickMsg{Reason: reason}})
	if err == nil {
		select {
		case p.ctrl <- outFrame{data: frame, close: true}:
			return
		default:
		}
	}
	p.close()
}

// Close идемпотентен.
func (h *Host) Close() {
	h.closeOnce.Do(func() {
		h.mu.Lock()
		h.closed = true
		ln := h.listener
		bc := h.broadcaster
		players := make([]*hostPlayer, 0, len(h.players))
		for _, p := range h.players {
			players = append(players, p)
		}
		h.players = make(map[uint8]*hostPlayer)
		h.mu.Unlock()

		close(h.done)
		if bc != nil {
			bc.Stop()
		}
		if ln != nil {
			ln.Close()
		}
		for _, p := range players {
			p.close()
		}
	})
}
