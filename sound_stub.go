//go:build nosound

// Test builds without audio.
package main

type sounds struct {
	nowName  string
	nowSince float64
}

var snd sounds

func (s *sounds) init()       {}
func (s *sounds) startMusic() {}
func (s *sounds) stopMusic()  {}
func (s *sounds) update()     {}
func play(string)             {}
func chip(int)                {}
