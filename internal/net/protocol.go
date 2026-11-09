package net

import (
	"bytes"
	"encoding/binary"
	"encoding/gob"
	"fmt"
	"io"
	"strings"
)

const (
	DefaultTCPPort  = 25000
	AppMagic        = "SQWAVE"
	ProtocolVersion = 1
)

// maxMsgSize — лимит одного сообщения (8 МБ).
const maxMsgSize = 8 << 20

type MsgType uint8

const (
	MsgJoin MsgType = iota
	MsgWelcome
	MsgLobby
	MsgStart
	MsgInput
	MsgSnapshot
	MsgGameOver
	MsgKick
	MsgCommand
)

// Message — конверт всех сообщений протокола.
type Message struct {
	Type     MsgType
	Join     *JoinMsg
	Welcome  *WelcomeMsg
	Lobby    *LobbyMsg
	Start    *StartInfo
	Input    *InputMsg
	Snapshot *SnapshotMsg
	GameOver *GameOverMsg
	Kick     *KickMsg
	Command  *CommandMsg
}

type JoinMsg struct{ Name string }

type WelcomeMsg struct {
	PlayerID   uint8
	MaxPlayers int
	GameName   string
}

type LobbyMsg struct{ Players []LobbyPlayer }

type InputMsg struct{ Input WireInput }

type KickMsg struct{ Reason string }

// CommandMsg — команда клиента хосту (смена снаряжения, покупка лечения).
type CommandMsg struct{ Command Command }

func RegisterGobTypes() {
	gob.Register(&Message{})
	gob.Register(&JoinMsg{})
	gob.Register(&WelcomeMsg{})
	gob.Register(&LobbyMsg{})
	gob.Register(&StartInfo{})
	gob.Register(&InputMsg{})
	gob.Register(&SnapshotMsg{})
	gob.Register(&GameOverMsg{})
	gob.Register(&KickMsg{})
	gob.Register(&CommandMsg{})
}

// cleanName обрезает имя и подставляет запасное, если пусто.
func cleanName(name, fallback string) string {
	name = strings.TrimSpace(name)
	if name == "" {
		return fallback
	}
	r := []rune(name)
	if len(r) > 24 {
		r = r[:24]
	}
	return string(r)
}

// encodeMessage собирает готовый кадр (длина + gob) один раз.
func encodeMessage(m *Message) ([]byte, error) {
	var buf bytes.Buffer
	if err := gob.NewEncoder(&buf).Encode(m); err != nil {
		return nil, err
	}
	if buf.Len() > maxMsgSize {
		return nil, fmt.Errorf("net: message too large (%d bytes)", buf.Len())
	}
	frame := make([]byte, 4+buf.Len())
	binary.LittleEndian.PutUint32(frame[:4], uint32(buf.Len()))
	copy(frame[4:], buf.Bytes())
	return frame, nil
}

func WriteMessage(w io.Writer, m *Message) error {
	frame, err := encodeMessage(m)
	if err != nil {
		return err
	}
	_, err = w.Write(frame)
	return err
}

// writeFrame пишет уже собранный кадр.
func writeFrame(w io.Writer, frame []byte) error {
	_, err := w.Write(frame)
	return err
}

func ReadMessage(r io.Reader) (*Message, error) {
	var lb [4]byte
	if _, err := io.ReadFull(r, lb[:]); err != nil {
		return nil, err
	}
	size := binary.LittleEndian.Uint32(lb[:])
	if size > maxMsgSize {
		return nil, fmt.Errorf("net: message too large (%d bytes)", size)
	}
	buf := make([]byte, size)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, err
	}
	var m Message
	if err := gob.NewDecoder(bytes.NewReader(buf)).Decode(&m); err != nil {
		return nil, err
	}
	return &m, nil
}
