package audio

import (
	"bytes"
	"fmt"
	"io"
	"os"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

type Player struct {
	ctx     *audio.Context
	current *audio.Player
	track   string
}

func New() (*Player, error) {
	return &Player{
		ctx: audio.NewContext(sampleRate),
	}, nil
}

// PlayLoop загружает WAV-файл и запускает его в цикле.
// Если этот трек уже играет — ничего не делает.
// Если играет другой — останавливает его и запускает новый.
func (p *Player) PlayLoop(path string) error {
	if p.current != nil && p.track == path {
		return nil
	}
	p.Stop()

	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("audio: read %s: %w", path, err)
	}

	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("audio: decode %s: %w", path, err)
	}

	// Бесконечный цикл.
	loop := audio.NewInfiniteLoop(stream, stream.Length())

	player, err := p.ctx.NewPlayer(loop)
	if err != nil {
		return fmt.Errorf("audio: new player: %w", err)
	}
	player.Play()

	p.current = player
	p.track = path
	return nil
}

// Stop останавливает текущий трек.
func (p *Player) Stop() {
	if p.current != nil {
		p.current.Pause()
		p.current.Close()
		p.current = nil
		p.track = ""
	}
}

// SetVolume громкость 0..1.
func (p *Player) SetVolume(v float64) {
	if p.current == nil {
		return
	}
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	p.current.SetVolume(v)
}

// ensure io не удалён линтером (используется косвенно).
var _ = io.EOF
