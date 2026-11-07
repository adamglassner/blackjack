// Blackjack: the rules and the flow of a round. A round runs as a script on
// its own goroutine holding mu, and lets go of it while it waits, so the
// window keeps drawing (like the original's async/await).
// (c) 2026 Ironite
package main

import (
	"fmt"
	"math"
	"math/rand/v2"
	"strings"
	"sync"
	"time"
)

const (
	startChips = 1000
	minBet     = 5
	maxBet     = 5000
	maxHands   = 4
	PY, DY     = 392.0, 112.0 // player and dealer card rows (table space)
)

var (
	mu       sync.Mutex
	speed    = 1.0 // test builds run faster
	autoPlay bool
)

var (
	shoeXY = [2]float64{1388, 62}
	discXY = [2]float64{166, 30}
)

type result struct{ cls, text, amt string }

type hand struct {
	cards                []int
	els                  []*Card
	bet                  float64
	doubled, done, split bool
	splitAces            bool
	result               *result
}

func newHand(b float64, split bool) *hand { return &hand{bet: b, split: split} }

type table struct {
	shoe      []int
	shoeDecks int
	discards  []*Card
	hands     []*hand
	dealer    hand
	active    int
	phase     string // title, bet, deal, play, dealer, settle
	bet       float64
	busy      bool
	holeUp    bool
	msg       string
	notice    string
	panels    []string
	banner    banner
	fx        particles
	cards     cardLayer
	removals  []removal
	deckOpen  bool // the toolbar's deck list is open
}

type removal struct {
	c  *Card
	at float64
}

var tb table

/* ---------------------------------------------------------------- rules */

func val(id int) int { return cardVal(id) }

func total(ids []int) (int, bool) {
	t, aces := 0, 0
	for _, id := range ids {
		if rankOf(id) == 1 {
			aces++
			t += 11
		} else {
			t += min(10, rankOf(id))
		}
	}
	for t > 21 && aces > 0 {
		t -= 10
		aces--
	}
	return t, aces > 0
}

func tot(ids []int) int { t, _ := total(ids); return t }

func isBJ(h *hand) bool { return len(h.cards) == 2 && tot(h.cards) == 21 && !h.split }

func cutFor(d int) int {
	switch d {
	case 1:
		return 18
	case 2:
		return 32
	}
	return int(math.Round(float64(d) * 52 * .25))
}

func deckName(n int) string {
	switch n {
	case 1:
		return "Single Deck"
	case 2:
		return "Double Deck"
	case 6:
		return "6-Deck Shoe"
	}
	return fmt.Sprintf("%d Decks", n)
}

// money shows whole chips, or two decimals when a 3:2 payout left a half.
func money(n float64) string {
	if int(math.Round(n*100))%100 != 0 {
		neg := n < 0
		n = math.Abs(n)
		w := math.Floor(n)
		s := fmtN(int(w)) + fmt.Sprintf(".%02d", int(math.Round((n-w)*100)))
		if neg {
			return "-" + s
		}
		return s
	}
	return fmtN(int(math.Round(n)))
}

/* --------------------------------------------------------------- script */

func (t *table) pace() float64 {
	m := map[string]float64{"relaxed": 1.35, "steady": 1, "quick": .6}[save.Settings.Pace]
	if m == 0 {
		m = 1
	}
	return m
}

// wait pauses the script for ms (scaled by the pace), letting the window run.
func (t *table) wait(ms float64) {
	d := time.Duration(ms * t.pace() / speed * float64(time.Millisecond))
	mu.Unlock()
	time.Sleep(d)
	mu.Lock()
}

// run starts a script on its own goroutine.
func (t *table) run(f func()) {
	go func() {
		mu.Lock()
		defer mu.Unlock()
		f()
	}()
}

/* --------------------------------------------------------------- layout */

func handCx(i, n int) float64 {
	if n == 1 {
		return 800
	}
	return [][]float64{{560, 1040}, {400, 800, 1200}, {290, 620, 980, 1310}}[n-2][i]
}

func (t *table) layout(dur float64) {
	n := len(t.hands)
	step := 24.0
	if n == 1 {
		step = 34
	}
	for i, h := range t.hands {
		cx := handCx(i, n)
		w := CW + step*float64(len(h.cards)-1)
		for k, c := range h.els {
			dbl := h.doubled && k == len(h.cards)-1
			x, y, r := cx-w/2+float64(k)*step, PY, 0.0
			if dbl {
				x, y, r = x+18, y+14, 90
			}
			t.cards.move(c, x, y, moveOpts{up: bp(true), r: fp(r), dur: dur, z: ip(100 + i*20 + k)})
		}
	}
	dw := CW + 40*float64(len(t.dealer.cards)-1)
	for k, c := range t.dealer.els {
		t.cards.move(c, 800-dw/2+float64(k)*40, DY, moveOpts{up: bp(k != 1 || t.holeUp), r: fp(0), dur: dur, z: ip(50 + k)})
	}
}

func (t *table) say(s string) { t.msg = s }

/* ----------------------------------------------------------------- shoe */

func (t *table) shuffle() {
	t.shoeDecks = save.Settings.Decks
	t.showBanner("Shuffling", deckName(t.shoeDecks))
	t.say("Shuffling a fresh " + strings.ToLower(deckName(t.shoeDecks)) + "...")
	play("shuffle")
	for k, c := range t.discards {
		t.cards.move(c, shoeXY[0], shoeXY[1], moveOpts{up: bp(false), r: fp(-8), dur: 500 + float64(k)*2})
		t.removals = append(t.removals, removal{c, now() + .9/speed})
	}
	t.discards = nil
	t.shoe = make([]int, 52*t.shoeDecks)
	for i := range t.shoe {
		t.shoe[i] = i % 52
	}
	rand.Shuffle(len(t.shoe), func(i, j int) { t.shoe[i], t.shoe[j] = t.shoe[j], t.shoe[i] })
	t.wait(1700)
}

func (t *table) give(h *hand, dur float64) int {
	id := t.shoe[len(t.shoe)-1]
	t.shoe = t.shoe[:len(t.shoe)-1]
	c := t.cards.make(id, cardOpts{x: shoeXY[0], y: shoeXY[1], r: -8, s: 1, up: false})
	c.face = id
	h.cards = append(h.cards, id)
	h.els = append(h.els, c)
	play("deal")
	t.layout(dur)
	t.wait(dur + 120)
	return id
}

func (t *table) sweep() {
	var all []*Card
	for _, h := range t.hands {
		all = append(all, h.els...)
	}
	all = append(all, t.dealer.els...)
	for _, c := range all {
		k := len(t.discards)
		t.discards = append(t.discards, c)
		t.cards.move(c, discXY[0]+(rand.Float64()-.5)*8, discXY[1]+8-min(40, float64(k)*.35),
			moveOpts{up: bp(false), r: fp((rand.Float64() - .5) * 8), dur: 520, z: ip(20 + k)})
	}
	t.hands = nil
	t.dealer = hand{}
	t.holeUp = false
}

/* -------------------------------------------------------------- betting */

func (t *table) betCap() float64 { return min(save.Balance, maxBet) }

func (t *table) canChip(v int) bool {
	return t.phase == "bet" && !t.busy && t.bet+float64(v) <= t.betCap()
}

func (t *table) addChip(v int) {
	if !t.canChip(v) {
		return
	}
	t.bet += float64(v)
	play("chip1")
	if t.bet >= minBet {
		t.say("Bet <b>" + money(t.bet) + "</b>. Press <b>Deal</b> when you're ready.")
	} else {
		t.say("Place your bet.")
	}
}

func (t *table) canClear() bool { return t.phase == "bet" && !t.busy && t.bet > 0 }

func (t *table) clearBet() {
	if t.phase != "bet" || t.busy {
		return
	}
	t.bet = 0
	play("chip2")
	t.say("Place your bet.")
}

func (t *table) broke() bool { return t.phase == "bet" && save.Balance < minBet && t.bet == 0 }

func (t *table) restock() {
	save.Balance += startChips
	save.Stats.Restocks++
	writeSave()
	play("chip6")
	t.bet = min(lastBet(), save.Balance)
	t.say("The house restocked you with <b>" + fmtN(startChips) + "</b> play chips.")
}

func lastBet() float64 {
	if save.LastBet > 0 {
		return save.LastBet
	}
	return 25
}

func (t *table) toBetting() {
	t.phase, t.busy = "bet", false
	t.bet = 0
	if save.Balance >= minBet {
		t.bet = max(minBet, min(lastBet(), save.Balance, maxBet))
	}
	t.bet -= math.Mod(t.bet, 5)
	if save.Balance < minBet {
		t.say("You’re out of chips.")
	} else {
		t.say("Same bet, <b>" + money(t.bet) + "</b>? Press <b>Deal</b>, or change it with the chips.")
	}
}

/* ---------------------------------------------------------------- round */

func (t *table) canDeal() bool {
	return t.phase == "bet" && !t.busy && t.bet >= minBet && t.bet <= save.Balance
}

func (t *table) deal() {
	if !t.canDeal() {
		return
	}
	t.busy, t.phase = true, "deal"
	t.run(t.dealRound)
}

func (t *table) dealRound() {
	save.LastBet = t.bet
	save.Balance -= t.bet
	save.AtRisk = t.bet
	writeSave()
	play("chip3")
	if len(t.shoe) == 0 || len(t.shoe) < cutFor(t.shoeDecks) || t.shoeDecks != save.Settings.Decks {
		t.shuffle()
	}
	t.hands = []*hand{newHand(t.bet, false)}
	t.dealer = hand{}
	t.active, t.holeUp = 0, false
	t.say("Dealing...")
	t.give(t.hands[0], 420)
	t.give(&t.dealer, 420)
	t.give(t.hands[0], 420)
	t.give(&t.dealer, 420)
	upV := val(t.dealer.cards[0])
	if upV == 11 || upV == 10 {
		what := "a ten"
		if upV == 11 {
			what = "an Ace"
		}
		t.say("Dealer shows " + what + " and checks for Blackjack...")
		t.wait(1100)
		if tot(t.dealer.cards) == 21 {
			t.revealHole()
			t.settle()
			return
		}
		t.say("No Blackjack for the dealer. Play on.")
		t.wait(500)
	}
	if isBJ(t.hands[0]) {
		t.hands[0].done = true
		t.revealHole()
		t.settle()
		return
	}
	t.phase, t.busy = "play", false
	t.prompt()
}

func (t *table) cur() *hand {
	if t.active < len(t.hands) {
		return t.hands[t.active]
	}
	return nil
}

func (t *table) canDouble(h *hand) bool {
	return h != nil && !h.done && len(h.cards) == 2 && !h.splitAces && save.Balance >= h.bet
}

func (t *table) canSplit(h *hand) bool {
	return h != nil && !h.done && len(h.cards) == 2 && val(h.cards[0]) == val(h.cards[1]) &&
		len(t.hands) < maxHands && save.Balance >= h.bet && !h.splitAces
}

func (t *table) playing() bool { return t.phase == "play" && !t.busy }

func (t *table) prompt() {
	h := t.cur()
	if h == nil {
		return
	}
	pt, soft := total(h.cards)
	up := t.dealer.cards[0]
	opts := []string{"<b>Hit</b>", "<b>Stand</b>"}
	if t.canDouble(h) {
		opts = append(opts, "<b>Double</b>")
	}
	if t.canSplit(h) {
		opts = append(opts, "<b>Split</b>")
	}
	list := strings.Join(opts[:len(opts)-1], ", ") + " or " + opts[len(opts)-1]
	pre := ""
	if len(t.hands) > 1 {
		pre = fmt.Sprintf("Hand %d: ", t.active+1)
	}
	sf := ""
	if soft && pt < 21 {
		sf = "soft "
	}
	t.say(fmt.Sprintf("%sYou have %s<b>%d</b> against the dealer's <b>%s</b>. %s?", pre, sf, pt, RANKS[rankOf(up)-1], list))
}

// act runs one of the player's moves.
func (t *table) act(kind string) {
	if !t.playing() {
		return
	}
	h := t.cur()
	switch kind {
	case "double":
		if !t.canDouble(h) {
			return
		}
	case "split":
		if !t.canSplit(h) {
			return
		}
	}
	t.busy = true
	t.run(func() {
		switch kind {
		case "hit":
			t.hit(h)
		case "stand":
			h.done = true
			play("card")
			t.next()
		case "double":
			t.double(h)
		case "split":
			t.split(h)
		}
	})
}

func (t *table) hit(h *hand) {
	t.give(h, 420)
	pt := tot(h.cards)
	if pt > 21 {
		h.done = true
		h.result = &result{cls: "bust", text: "Bust"}
		play("bust")
		t.wait(600)
	} else if pt == 21 {
		h.done = true
	}
	t.next()
}

func (t *table) double(h *hand) {
	save.Balance -= h.bet
	save.AtRisk += h.bet
	h.bet *= 2
	h.doubled = true
	writeSave()
	play("chip3")
	t.say("Doubled to <b>" + money(h.bet) + "</b>. One card...")
	t.wait(300)
	t.give(h, 420)
	h.done = true
	if tot(h.cards) > 21 {
		h.result = &result{cls: "bust", text: "Bust"}
		play("bust")
		t.wait(600)
	}
	t.next()
}

func (t *table) split(h *hand) {
	save.Balance -= h.bet
	save.AtRisk += h.bet
	writeSave()
	play("chip3")
	nh := newHand(h.bet, true)
	nh.cards = []int{h.cards[1]}
	nh.els = []*Card{h.els[1]}
	h.cards, h.els = h.cards[:1], h.els[:1]
	h.split = true
	aces := rankOf(h.cards[0]) == 1
	if aces {
		h.splitAces, nh.splitAces = true, true
	}
	t.hands = append(t.hands[:t.active+1], append([]*hand{nh}, t.hands[t.active+1:]...)...)
	t.layout(420)
	if aces {
		t.say("Split Aces: one card each.")
	} else {
		t.say("Split into two hands.")
	}
	t.wait(500)
	t.give(h, 420)
	if aces {
		t.give(nh, 420)
		h.done, nh.done = true, true
	} else if tot(h.cards) == 21 {
		h.done = true
	}
	t.next()
}

func (t *table) next() {
	if !t.hands[t.active].done {
		t.busy = false
		t.prompt()
		return
	}
	for i := t.active + 1; i < len(t.hands); i++ {
		h := t.hands[i]
		if h.done {
			continue
		}
		t.active = i
		t.wait(250)
		if len(h.cards) == 1 {
			t.give(h, 420)
			if tot(h.cards) == 21 {
				h.done = true
				continue
			}
		}
		t.busy = false
		t.prompt()
		return
	}
	t.dealerPlay()
}

func (t *table) revealHole() {
	t.holeUp = true
	play("flip")
	t.layout(380)
	t.wait(650)
}

func (t *table) dealerPlay() {
	t.phase, t.busy = "dealer", true
	t.say("Dealer turns over the hole card...")
	t.revealHole()
	live := false
	for _, h := range t.hands {
		if h.result == nil || h.result.cls != "bust" {
			live = true
		}
	}
	if live {
		for tot(t.dealer.cards) < 17 {
			t.say(fmt.Sprintf("Dealer has <b>%d</b> and draws.", tot(t.dealer.cards)))
			t.wait(350)
			t.give(&t.dealer, 460)
		}
	}
	dt := tot(t.dealer.cards)
	switch {
	case !live:
		t.say("All hands bust.")
	case dt > 21:
		t.say(fmt.Sprintf("Dealer busts with <b>%d</b>!", dt))
	default:
		t.say(fmt.Sprintf("Dealer stands on <b>%d</b>.", dt))
	}
	t.wait(500)
	t.settle()
}

func (t *table) settle() {
	t.phase, t.busy = "settle", true
	dt := tot(t.dealer.cards)
	dBJ := len(t.dealer.cards) == 2 && dt == 21
	net, bjHit := 0.0, false
	var outcomes []string
	for _, h := range t.hands {
		pt := tot(h.cards)
		bj := isBJ(h) && len(t.hands) == 1
		var r result
		switch {
		case pt > 21:
			r = result{cls: "bust", text: "Bust"}
		case dBJ:
			if bj {
				r = result{cls: "push", text: "Push"}
			} else {
				r = result{cls: "lose", text: "Dealer Blackjack"}
			}
		case bj:
			r = result{cls: "bj", text: "Blackjack"}
			bjHit = true
		case dt > 21 || pt > dt:
			r = result{cls: "win", text: "Win"}
		case pt == dt:
			r = result{cls: "push", text: "Push"}
		default:
			r = result{cls: "lose", text: "Lose"}
		}
		pay := 0.0
		switch r.cls {
		case "bj":
			pay = h.bet * 2.5
		case "win":
			pay = h.bet * 2
		case "push":
			pay = h.bet
		}
		d := pay - h.bet
		net += d
		if d > 0 {
			r.amt = "+" + money(d)
		} else if d < 0 {
			r.amt = "−" + money(-d)
		}
		h.result = &r
		save.Balance += pay
		outcomes = append(outcomes, r.cls)
		save.Stats.Hands++
		switch r.cls {
		case "win", "bj":
			save.Stats.Won++
		case "push":
			save.Stats.Pushed++
		default:
			save.Stats.Lost++
		}
		if r.cls == "bj" {
			save.Stats.Blackjacks++
		}
	}
	save.AtRisk = 0
	save.Stats.Peak = max(save.Stats.Peak, save.Balance)
	if net > 0 {
		save.Stats.BiggestWin = max(save.Stats.BiggestWin, net)
	}
	writeSave()
	every := func(c string) bool {
		for _, o := range outcomes {
			if o != c {
				return false
			}
		}
		return true
	}
	why := fmt.Sprintf("Dealer has %d", dt)
	if dBJ {
		why = "Dealer has Blackjack"
	} else if dt > 21 {
		why = fmt.Sprintf("Dealer busts with %d", dt)
	}
	gold := []col{hex("#ffe7a6"), hex("#f6dc8a"), hex("#ffffff"), hex("#c99a3a")}
	switch {
	case bjHit:
		t.showBanner("Blackjack!", "Paid 3 to 2: +"+money(net)+" chips")
		play("blackjack")
		t.fx.burst(800, 360, 150, gold, 760, 8, false)
	case net > 0:
		t.showBanner("You Win", "+"+money(net)+" chips · "+why)
		play("win")
		t.fx.burst(800, 380, 80, []col{hex("#ffe7a6"), hex("#ffffff"), hex("#c99a3a")}, 600, 7, true)
	case net == 0:
		if every("push") {
			t.showBanner("Push", "Your bet is returned")
		} else {
			t.showBanner("Even", "You broke even this hand")
		}
		play("push")
	default:
		if every("bust") {
			t.showBanner("Bust", "−"+money(-net)+" chips · Over 21")
		} else {
			t.showBanner("Dealer Wins", "−"+money(-net)+" chips · "+why)
			play("lose")
		}
	}
	switch {
	case net > 0:
		t.say("You won <b>" + money(net) + "</b>. " + why + ".")
	case net < 0:
		t.say("You lost <b>" + money(-net) + "</b>. " + why + ".")
	default:
		t.say("No chips changed hands. " + why + ".")
	}
	t.wait(3200)
	t.sweep()
	play("card")
	t.wait(450)
	t.toBetting()
}

func (t *table) showBanner(title, sub string) {
	t.banner.title, t.banner.sub = title, sub
	t.banner.show.set(1, .35, easeStd)
	t.banner.hideAt = now() + 1.5*t.pace()/speed
}

/* --------------------------------------------------------------- menus */

func (t *table) setDecks(n int) {
	save.Settings.Decks = n
	writeSave()
	if t.phase == "bet" && !t.busy {
		t.busy = true
		t.run(func() {
			t.shuffle()
			t.toBetting()
		})
	} else if t.phase != "title" {
		t.say("The table switches to a <b>" + strings.ToLower(deckName(n)) + "</b> after this hand.")
	}
}

func (t *table) start() {
	t.closePanel("title")
	if len(t.shoe) == 0 || t.shoeDecks != save.Settings.Decks {
		t.phase, t.busy = "deal", true
		t.run(func() {
			t.shuffle()
			t.toBetting()
		})
		return
	}
	t.toBetting()
}

func newTable() {
	tb = table{phase: "title", msg: "Place your bet."}
	tb.cards.clear()
	// a bet left on the table when the game was closed mid-hand is returned
	if save.AtRisk > 0 {
		save.Balance += save.AtRisk
		tb.notice = "Your unfinished hand from last time was returned to you: " + money(save.AtRisk) + " chips."
		save.AtRisk = 0
		writeSave()
	}
	tb.bet = min(lastBet(), max(0, save.Balance))
	tb.panels = []string{"title"}
}

func (t *table) topPanel() string {
	if len(t.panels) == 0 {
		return ""
	}
	return t.panels[len(t.panels)-1]
}

func (t *table) showPanel(id string) {
	t.closePanel(id)
	t.panels = append(t.panels, id)
	panelScroll[id] = 0
	t.deckOpen = false
}

func (t *table) closePanel(id string) {
	keep := t.panels[:0]
	for _, p := range t.panels {
		if p != id {
			keep = append(keep, p)
		}
	}
	t.panels = keep
}

func (t *table) closeTop() {
	if t.deckOpen {
		t.deckOpen = false
		return
	}
	switch p := t.topPanel(); p {
	case "rules", "credits":
		t.closePanel(p)
	}
}

var panelScroll = map[string]float64{}

// step runs once a frame: finishes removals and fades the banner.
func (t *table) step(dt float64) {
	keep := t.removals[:0]
	for _, r := range t.removals {
		if now() >= r.at {
			t.cards.remove(r.c)
		} else {
			keep = append(keep, r)
		}
	}
	t.removals = keep
	if t.banner.hideAt > 0 && now() > t.banner.hideAt && t.banner.show.to == 1 {
		t.banner.show.set(0, .35, easeStd)
	}
	t.fx.step(dt)
}
