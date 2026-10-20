package audio

import (
	"bytes"
	"fmt"
	"io/fs"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/wav"
)

const sampleRate = 44100

type Player struct {
	ctx     *audio.Context
	current *audio.Player
	track   string
	volume  float64
	fs      fs.FS
}

func New(fsys fs.FS) (*Player, error) {
	return &Player{
		ctx:    audio.NewContext(sampleRate),
		volume: 0.7,
		fs:     fsys,
	}, nil
}

func (p *Player) PlayLoop(path string) error {
	if p.current != nil {
		return nil
	}
	return p.playInternal(path)
}

func (p *Player) PlayLoopForce(path string) error {
	if p.current != nil && p.track == path {
		return nil
	}
	p.Stop()
	return p.playInternal(path)
}

func (p *Player) playInternal(path string) error {
	data, err := fs.ReadFile(p.fs, path)
	if err != nil {
		return fmt.Errorf("audio: read %s: %w", path, err)
	}

	stream, err := wav.DecodeWithSampleRate(sampleRate, bytes.NewReader(data))
	if err != nil {
		return fmt.Errorf("audio: decode %s: %w", path, err)
	}

	loop := audio.NewInfiniteLoop(stream, stream.Length())
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
