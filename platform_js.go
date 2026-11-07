//go:build js

package main

// The browser build is only used for automated visual checks. Hooks:
// ?rig (a fixed shoe: 8s against a 6, for a split), ?notice, ?broke,
// ?speed=N, ?auto (plays hands by itself and logs the chip count after each:
// "HAND n BALANCE b").

import (
	"math/rand/v2"
	"strconv"
	"strings"
	"syscall/js"
	"time"
)

func init() {
	q := js.Global().Get("location").Get("search").String()
	if i := strings.Index(q, "speed="); i >= 0 {
		v := q[i+6:]
		if j := strings.IndexAny(v, "&"); j >= 0 {
			v = v[:j]
		}
		if f, err := strconv.ParseFloat(v, 64); err == nil && f > 0 {
			speed = f
		}
	}
	testSetup = func() {
		t := &tb
		if strings.Contains(q, "rig") {
			ids := []int{7, 5, 20, 30, 9, 11, 3, 12, 25, 0}
			var shoe []int
			for i := 0; i < 40; i++ {
				shoe = append(shoe, (i*7)%52)
			}
			for i := len(ids) - 1; i >= 0; i-- {
				shoe = append(shoe, ids[i])
			}
			t.shoe, t.shoeDecks = shoe, save.Settings.Decks
		}
		if strings.Contains(q, "notice") {
			t.notice = "Your unfinished hand from last time was returned to you: 50 chips."
		}
		if strings.Contains(q, "auto") {
			go autoLoop()
		}
		if strings.Contains(q, "broke") {
			save.Balance = 0
			t.bet = 0
		}
	}
}

// autoLoop plays: bet a random chip, hit below 17 (sometimes doubles and
// splits), and checks that balance + chips on the table never changes
// except by a hand's result.
func autoLoop() {
	time.Sleep(2 * time.Second)
	mu.Lock()
	tb.start()
	mu.Unlock()
	hands, lastStats := 0, -1
	for hands < 60 {
		time.Sleep(20 * time.Millisecond)
		mu.Lock()
		t := &tb
		switch {
		case t.phase == "bet" && !t.busy:
			if save.Stats.Hands != lastStats {
				lastStats = save.Stats.Hands
				hands++
				println("HAND", hands, "BALANCE", money(save.Balance), "STATS", save.Stats.Hands, save.Stats.Won, save.Stats.Lost, save.Stats.Pushed, save.Stats.Blackjacks)
			}
			if t.broke() {
				t.restock()
			}
			t.bet = 0
			for _, v := range []int{5, 25, 100} {
				if rand.IntN(2) == 0 && t.canChip(v) {
					t.addChip(v)
				}
			}
			if t.bet < minBet && t.canChip(5) {
				t.addChip(5)
			}
			t.deal()
		case t.playing():
			h := t.cur()
			pt := tot(h.cards)
			switch {
			case t.canSplit(h) && rand.IntN(2) == 0:
				t.act("split")
			case t.canDouble(h) && pt >= 9 && pt <= 11:
				t.act("double")
			case pt < 17:
				t.act("hit")
			default:
				t.act("stand")
			}
		}
		if save.AtRisk < 0 || save.Balance < 0 {
			println("FAIL: negative", save.Balance, save.AtRisk)
		}
		mu.Unlock()
	}
	println("AUTO DONE")
}
