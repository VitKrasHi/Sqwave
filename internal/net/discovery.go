package net

import (
	"encoding/json"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	DiscoveryPort = 25001
	bcInterval    = 1 * time.Second
	serverTTL     = 5 * time.Second
)

// ServerInfo — анонс хоста в дискавери.
type ServerInfo struct {
	App            string    `json:"app"`
	Version        int       `json:"v"`
	Port           int       `json:"port"`
	GameName       string    `json:"name"`
	MaxPlayers     int       `json:"max"`
	CurrentPlayers int       `json:"cur"`
	Address        string    `json:"-"`
	LastSeen       time.Time `json:"-"`
}

// Broadcaster (хост) шлёт UDP-анонсы на широковещательный адрес.
type Broadcaster struct {
	infoFn func() ServerInfo
	conn   *net.UDPConn
	done   chan struct{}
	once   sync.Once
}

func NewBroadcaster(infoFn func() ServerInfo) (*Broadcaster, error) {
	conn, err := net.DialUDP("udp4", nil, &net.UDPAddr{IP: net.IPv4bcast, Port: DiscoveryPort})
	if err != nil {
		return nil, err
	}
	conn.SetWriteBuffer(2048)
	return &Broadcaster{infoFn: infoFn, conn: conn, done: make(chan struct{})}, nil
}

func (b *Broadcaster) Start() {
	go func() {
		b.send()
		t := time.NewTicker(bcInterval)
		defer t.Stop()
		for {
			select {
			case <-b.done:
				return
			case <-t.C:
				b.send()
			}
		}
	}()
}

func (b *Broadcaster) send() {
	info := b.infoFn()
	info.App = AppMagic
	info.Version = ProtocolVersion
	data, err := json.Marshal(info)
	if err != nil {
		return
	}
	b.conn.Write(data)
}

func (b *Broadcaster) Stop() {
	b.once.Do(func() {
		close(b.done)
		b.conn.Close()
	})
}

// Listener (клиент) слушает порт дискавери на всех интерфейсах.
type Listener struct {
	conn    *net.UDPConn
	mu      sync.RWMutex
	servers map[string]*ServerInfo
	done    chan struct{}
	once    sync.Once
}

func NewListener() (*Listener, error) {
	// 0.0.0.0:25001 — ловим анонсы из LAN, ZeroTier, Radmin и т.п.
	conn, err := net.ListenUDP("udp4", &net.UDPAddr{IP: net.IPv4zero, Port: DiscoveryPort})
	if err != nil {
		return nil, err
	}
	conn.SetReadBuffer(64 * 1024)
	l := &Listener{conn: conn, servers: map[string]*ServerInfo{}, done: make(chan struct{})}
	go l.readLoop()
	go l.pruneLoop()
	return l, nil
}

func (l *Listener) readLoop() {
	buf := make([]byte, 4096)
	for {
		n, raddr, err := l.conn.ReadFromUDP(buf)
		if err != nil {
			select {
			case <-l.done:
				return
			default:
				time.Sleep(10 * time.Millisecond)
				continue
			}
		}
		var info ServerInfo
		if json.Unmarshal(buf[:n], &info) != nil {
			continue
		}
		if info.App != AppMagic || info.Version != ProtocolVersion {
			continue
		}
		info.Address = raddr.IP.String()
		info.LastSeen = time.Now()
		key := fmt.Sprintf("%s:%d", info.Address, info.Port)

		l.mu.Lock()
		l.servers[key] = &info
		l.mu.Unlock()
	}
}

func (l *Listener) pruneLoop() {
	t := time.NewTicker(2 * time.Second)
	defer t.Stop()
	for {
		select {
		case <-l.done:
			return
		case <-t.C:
			l.mu.Lock()
			for k, v := range l.servers {
				if time.Since(v.LastSeen) > serverTTL {
					delete(l.servers, k)
				}
			}
			l.mu.Unlock()
		}
	}
}

// Servers отдаёт найденные хосты в стабильном порядке.
func (l *Listener) Servers() []ServerInfo {
	l.mu.RLock()
	defer l.mu.RUnlock()
	out := make([]ServerInfo, 0, len(l.servers))
	for _, s := range l.servers {
		out = append(out, *s)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Address != out[j].Address {
			return out[i].Address < out[j].Address
		}
		return out[i].Port < out[j].Port
	})
	return out
}

func (l *Listener) Stop() {
	l.once.Do(func() {
		close(l.done)
		l.conn.Close()
	})
}

// ScanSubnet ищет хосты с открытым TCP-портом в локальных подсетях.
// timeout — таймаут одного подключения; <= 0 означает 200 мс.
func ScanSubnet(port int, timeout time.Duration) []string {
	if timeout <= 0 {
		timeout = 200 * time.Millisecond
	}
	var (
		mu  sync.Mutex
		ips []string
	)
	add := func(ip string) {
		mu.Lock()
		ips = append(ips, ip)
		mu.Unlock()
	}

	var wg sync.WaitGroup
	sem := make(chan struct{}, 100) // не больше 100 одновременных подключений

	probe := func(ip string) {
		defer wg.Done()
		sem <- struct{}{}
		defer func() { <-sem }()
		conn, err := net.DialTimeout("tcp", net.JoinHostPort(ip, strconv.Itoa(port)), timeout)
		if err == nil {
			conn.Close()
			add(ip)
		}
	}

	// 1. localhost — чтобы играть на одном ПК
	wg.Add(1)
	go probe("127.0.0.1")

	// 2. адреса активных интерфейсов (LAN, ZeroTier, Radmin)
	for _, subnet := range localSubnets() {
		for i := 1; i <= 254; i++ {
			wg.Add(1)
			go probe(fmt.Sprintf("%s%d", subnet, i))
		}
	}

	wg.Wait()
	sort.Strings(ips)
	return ips
}

// localSubnets возвращает уникальные префиксы IPv4-подсетей.
func localSubnets() []string {
	ifaces, _ := net.Interfaces()
	seen := make(map[string]bool)
	var subnets []string
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		name := strings.ToLower(iface.Name)
		// Пропускаем Docker, виртуалки и прочий мусор
		if strings.HasPrefix(name, "docker") || strings.HasPrefix(name, "br-") ||
			strings.HasPrefix(name, "veth") || strings.HasPrefix(name, "cni") ||
			strings.HasPrefix(name, "flannel") || strings.HasPrefix(name, "virbr") {
			continue
		}

		addrs, _ := iface.Addrs()
		for _, a := range addrs {
			ipnet, ok := a.(*net.IPNet)
			if !ok {
				continue
			}
			ipv4 := ipnet.IP.To4()
			if ipv4 == nil {
				continue
			}
			// Первые три октета: 192.168.1.
			sub := fmt.Sprintf("%d.%d.%d.", ipv4[0], ipv4[1], ipv4[2])
			if seen[sub] {
				continue
			}
			seen[sub] = true
			subnets = append(subnets, sub)
		}
	}
	return subnets
}
