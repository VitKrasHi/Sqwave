package audio

import (
	"bytes"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

type Player struct {
	ctx     *audio.Context
	current *audio.Player
	track   string
	volume  float64
	assets  fs.FS
}

func New(assets fs.FS) (*Player, error) {
	return &Player{
		ctx:    audio.NewContext(sampleRate),
		volume: 0.7,
		assets: assets,
	}, nil
}

// PlayLoop — играет трек в цикле, если НИЧЕГО не играет.
func (p *Player) PlayLoop(path string) error {
	if p.current != nil {
		return nil
	}
	return p.playInternal(path)
}

// PlayLoopForce — принудительно переключает на трек.
func (p *Player) PlayLoopForce(path string) error {
	if p.current != nil && p.track == path {
		return nil
	}
	p.Stop()
	return p.playInternal(path)
}

func (p *Player) playInternal(path string) error {
	data, err := fs.ReadFile(p.assets, path)
	if err != nil {
		return fmt.Errorf("audio: read %s: %w", path, err)
	}

	var src io.ReadSeeker
	var length int64

	if strings.HasSuffix(path, ".ogg") {
		s, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("audio: decode %s: %w", path, err)
		}
		src = s
		length = s.Length()
	} else {
		s, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
		if err != nil {
			return fmt.Errorf("audio: decode %s: %w", path, err)
		}
		src = s
		length = s.Length()
	}

	loop := audio.NewInfiniteLoop(src, length)

	player, err := p.ctx.NewPlayer(loop)
	if err != nil {
		return fmt.Errorf("audio: new player: %w", err)
	}
	player.SetVolume(p.volume)
	player.Play()

	p.current = player
	p.track = path
	return nil
}

func (p *Player) Stop() {
	if p.current != nil {
		p.current.Pause()
		p.current.Close()
		p.current = nil
		p.track = ""
	}
}

func (p *Player) SetVolume(v float64) {
	if v < 0 {
		v = 0
	}
	if v > 1 {
		v = 1
	}
	p.volume = v
	if p.current != nil {
		p.current.SetVolume(v)
	}
}

func (p *Player) Volume() float64 {
	return p.volume
}
