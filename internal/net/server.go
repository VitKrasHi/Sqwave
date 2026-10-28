package net

import (
	"encoding/json"
	"fmt"
	"net"
	"sync"
)

type Server struct {
	listener net.Listener
	mu       sync.Mutex
	clients  map[net.Conn]*clientConn
	stopChan chan struct{}
	port     int
}

type clientConn struct {
	conn    net.Conn
	decoder *json.Decoder
	encoder *json.Encoder
	state   PlayerState
	mu      sync.Mutex
}

func StartServer(port int) (*Server, error) {
	ln, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return nil, err
	}
	s := &Server{
		listener: ln,
		clients:  make(map[net.Conn]*clientConn),
		stopChan: make(chan struct{}),
		port:     port,
	}
	go s.acceptLoop()
	return s, nil
}

func (s *Server) Port() int { return s.port }

func (s *Server) acceptLoop() {
	for {
		select {
		case <-s.stopChan:
			return
		default:
		}
		conn, err := s.listener.Accept()
		if err != nil {
			return
		}
		go s.handleConn(conn)
	}
}

func (s *Server) handleConn(conn net.Conn) {
	cc := &clientConn{
		conn:    conn,
		decoder: json.NewDecoder(conn),
		encoder: json.NewEncoder(conn),
	}
	s.mu.Lock()
	s.clients[conn] = cc
	s.mu.Unlock()

	defer func() {
		s.mu.Lock()
		delete(s.clients, conn)
		s.mu.Unlock()
		conn.Close()
	}()

	for {
		var msg Message
		if err := cc.decoder.Decode(&msg); err != nil {
			return
		}
		if msg.Type == "state" && msg.Player != nil {
			cc.mu.Lock()
			cc.state = *msg.Player
			cc.mu.Unlock()
		}
	}
}

func (s *Server) Broadcast(host PlayerState) {
	s.mu.Lock()
	defer s.mu.Unlock()

	all := []PlayerState{host}
	for _, cc := range s.clients {
		cc.mu.Lock()
		if cc.state.Name != "" {
			all = append(all, cc.state)
		}
		cc.mu.Unlock()
	}

	msg := Message{Type: "state", Players: all}
	for _, cc := range s.clients {
		cc.mu.Lock()
		_ = cc.encoder.Encode(msg)
		cc.mu.Unlock()
	}
}

func (s *Server) GetRemotePlayers() []PlayerState {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]PlayerState, 0, len(s.clients))
	for _, cc := range s.clients {
		cc.mu.Lock()
		if cc.state.Name != "" {
			out = append(out, cc.state)
		}
		cc.mu.Unlock()
	}
	return out
}

func (s *Server) Stop() {
	select {
	case <-s.stopChan:
	default:
		close(s.stopChan)
	}
	s.listener.Close()
	s.mu.Lock()
	for _, cc := range s.clients {
		cc.conn.Close()
	}
	s.mu.Unlock()
}
