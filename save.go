// The save file: balance, stats and settings in
// %APPDATA%\Ironite\Blackjack\save.json, outside the install folder, in the
// same format as version 1, so nobody loses their chips.
// (c) 2026 Ironite
package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
)

type Settings struct {
	Music bool   `json:"music"`
	Sfx   bool   `json:"sfx"`
	Vol   int    `json:"vol"`
	Decks int    `json:"decks"`
	Pace  string `json:"pace"`
}

type Stats struct {
	Hands      int     `json:"hands"`
	Won        int     `json:"won"`
	Lost       int     `json:"lost"`
	Pushed     int     `json:"pushed"`
	Blackjacks int     `json:"blackjacks"`
	BiggestWin float64 `json:"biggestWin"`
	Peak       float64 `json:"peak"`
	Restocks   int     `json:"restocks"`
}

// Money is a float: Blackjack pays 3 to 2, so an odd bet can win half a chip.
type SaveData struct {
	Settings Settings `json:"settings"`
	Balance  float64  `json:"balance"`
	AtRisk   float64  `json:"atRisk"`
	LastBet  float64  `json:"lastBet"`
	Stats    Stats    `json:"stats"`
}

const gameID = "Blackjack"

var (
	save   = defaults()
	saveMu sync.Mutex
)

func defaults() SaveData {
	return SaveData{
		Settings: Settings{Music: true, Sfx: true, Vol: 60, Decks: 1, Pace: "steady"},
		Balance:  startChips, LastBet: 25,
		Stats: Stats{Peak: startChips},
	}
}

func saveDir() string {
	d, err := os.UserConfigDir() // %APPDATA% on Windows
	if err != nil {
		d = os.TempDir()
	}
	return filepath.Join(d, "Ironite", gameID)
}

func loadSave() {
	b, err := os.ReadFile(filepath.Join(saveDir(), "save.json"))
	if err != nil {
		return
	}
	d := defaults()
	if json.Unmarshal(b, &d) != nil {
		return
	}
	switch d.Settings.Decks {
	case 1, 2, 6:
	default:
		d.Settings.Decks = 1
	}
	if d.Settings.Pace == "" {
		d.Settings.Pace = "steady"
	}
	save = d
}

var saveCh = make(chan []byte, 1)

// writeSave hands the latest save to a background writer so the game never
// stutters on disk. Only the newest pending save is kept.
func writeSave() {
	b, err := json.Marshal(save)
	if err != nil {
		return
	}
	select {
	case <-saveCh:
	default:
	}
	saveCh <- b
}

func saveLoop() {
	for b := range saveCh {
		writeFile(b)
	}
}

// flushSave writes the save right now (used when the window closes).
func flushSave() {
	if b, err := json.Marshal(save); err == nil {
		writeFile(b)
	}
}

func writeFile(b []byte) {
	saveMu.Lock()
	defer saveMu.Unlock()
	dir := saveDir()
	if os.MkdirAll(dir, 0o755) != nil {
		return
	}
	tmp := filepath.Join(dir, "save.json.tmp")
	if os.WriteFile(tmp, b, 0o644) == nil {
		_ = os.Rename(tmp, filepath.Join(dir, "save.json"))
	}
}
