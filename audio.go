//go:build !nosound

// Music and sound effects. The tunes and sounds are the game's original
// synthesized audio, rendered to Ogg Vorbis by tools/export.
// (c) 2026 Ironite
package main

import (
	"bytes"
	"fmt"
	"io"
	"math"
	"math/rand/v2"

	"github.com/hajimehoshi/ebiten/v2/audio"
	"github.com/hajimehoshi/ebiten/v2/audio/vorbis"
)

const sampleRate = 44100

type sounds struct {
	ctx      *audio.Context
	sfx      map[string][][]byte // name -> variants (decoded PCM)
	tunes    [][]byte            // encoded Ogg
	music    *audio.Player
	idx      int
	playing  bool
	nowName  string
	nowSince float64
	queue    []string
}

var snd sounds

func (s *sounds) init() {
	s.ctx = audio.NewContext(sampleRate)
	s.sfx = map[string][][]byte{}
	for _, v := range []struct {
		name string
		n    int
	}{{"card", 4}, {"deal", 4}, {"flip", 3}, {"shuffle", 2}, {"chip1", 3}, {"chip2", 3}, {"chip3", 3}, {"chip6", 2},
		{"win", 1}, {"blackjack", 1}, {"push", 1}, {"lose", 1}, {"bust", 1}} {
		for i := 0; i < v.n; i++ {
			b, err := files.ReadFile(fmt.Sprintf("assets/audio/%s-%d.ogg", v.name, i))
			if err != nil {
				continue
			}
			st, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(b))
			if err != nil {
				continue
			}
			pcm, err := io.ReadAll(st)
			if err != nil {
				continue
			}
			s.sfx[v.name] = append(s.sfx[v.name], pcm)
		}
	}
	for i := range tuneNames {
		b, err := files.ReadFile(fmt.Sprintf("assets/audio/tune-%d.ogg", i))
		if err == nil {
			s.tunes = append(s.tunes, b)
		}
	}
	if len(s.tunes) > 0 {
		s.idx = rand.IntN(len(s.tunes))
	}
}

// musicVol and sfxVol follow the original's volume curves.
func musicVol() float64 { return math.Pow(float64(save.Settings.Vol)/100, 1.6) }
func sfxVol() float64   { return max(.35, math.Pow(float64(save.Settings.Vol)/100, .6)) }

// play queues a sound effect; sounds start on the next frame.
func play(name string) {
	if save.Settings.Sfx {
		snd.queue = append(snd.queue, name)
	}
}

func (s *sounds) startMusic() {
	if s.ctx == nil || s.playing || len(s.tunes) == 0 {
		return
	}
	s.playing = true
	s.load(false)
}

func (s *sounds) load(announce bool) {
	st, err := vorbis.DecodeWithSampleRate(sampleRate, bytes.NewReader(s.tunes[s.idx]))
	if err != nil {
		s.playing = false
		return
	}
	p, err := s.ctx.NewPlayer(st)
	if err != nil {
		s.playing = false
		return
	}
	p.SetVolume(musicVol())
	p.Play()
	s.music = p
	if announce && s.idx < len(tuneNames) {
		s.nowName, s.nowSince = tuneNames[s.idx], now()
	}
}

// stopMusic stops the music; the next tune in the rotation plays next time.
func (s *sounds) stopMusic() {
	if !s.playing {
		return
	}
	s.playing = false
	if s.music != nil {
		s.music.Pause()
		_ = s.music.Close()
		s.music = nil
	}
	s.idx = (s.idx + 1) % len(s.tunes)
}

// update runs once a frame: starts queued sounds and moves to the next tune.
func (s *sounds) update() {
	if s.ctx == nil {
		return
	}
	for _, name := range s.queue {
		vs := s.sfx[name]
		if len(vs) == 0 {
			continue
		}
		p := s.ctx.NewPlayerFromBytes(vs[rand.IntN(len(vs))])
		p.SetVolume(sfxVol())
		p.Play()
	}
	s.queue = s.queue[:0]
	if s.playing && s.music != nil {
		s.music.SetVolume(musicVol())
		if !s.music.IsPlaying() {
			_ = s.music.Close()
			s.idx = (s.idx + 1) % len(s.tunes)
			s.load(true)
		}
	}
}

// chip sounds come in 1, 2, 3 and 6-chip clinks.
func chip(n int) { play(fmt.Sprintf("chip%d", n)) }
