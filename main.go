// Blackjack: The Gold Room. Twenty-one against the dealer at a casino table
// with dark green felt and gold trim. Play money only.
// (c) 2026 Ironite
package main

import (
	"fmt"
	"log"
	"math"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

// Set at build time with -ldflags "-X main.version=..."; otherwise read from
// the embedded CHANGELOG.md.
var (
	title   = "Blackjack"
	version = ""
)

// testSetup lets the browser test build open a screen straight away.
var testSetup func()

func readVersion() {
	if version != "" {
		return
	}
	b, _ := files.ReadFile("CHANGELOG.md")
	if m := regexp.MustCompile(`(?m)^## \[([0-9.]+)\]`).FindSubmatch(b); m != nil {
		version = string(m[1])
	} else {
		version = "0.0.0"
	}
}

func main() {
	readVersion()
	if len(os.Args) > 1 && (os.Args[1] == "--version" || os.Args[1] == "/version") {
		fmt.Println(version)
		return
	}
	if !singleInstance() {
		return
	}
	if d, err := os.UserCacheDir(); err == nil {
		dir := filepath.Join(d, "Ironite", gameID)
		_ = os.MkdirAll(dir, 0o755)
		if f, err := os.OpenFile(filepath.Join(dir, "log.txt"), os.O_CREATE|os.O_TRUNC|os.O_WRONLY, 0o644); err == nil {
			log.SetOutput(f)
			defer f.Close()
		}
	}
	log.Printf("%s %s starting", title, version)
	loadSave()
	go saveLoop()
	loadAssets()
	snd.init()
	newTable()
	if testSetup != nil {
		testSetup()
	}
	if save.Settings.Music {
		snd.startMusic()
	}

	ebiten.SetWindowTitle(title)
	ebiten.SetWindowIcon(windowIcons())
	ebiten.SetWindowResizingMode(ebiten.WindowResizingModeEnabled)
	ebiten.SetWindowSizeLimits(800, 450, -1, -1)
	if mw, mh := ebiten.Monitor().Size(); mw > 0 && mh > 0 {
		s := min(float64(mw)*.88/stageW, float64(mh)*.85/stageH, 1.25)
		ebiten.SetWindowSize(int(stageW*s), int(stageH*s))
	} else {
		ebiten.SetWindowSize(1280, 720)
	}
	ebiten.SetRunnableOnUnfocused(true)
	ebiten.SetTPS(60)
	if err := ebiten.RunGame(&game{last: time.Now()}); err != nil {
		log.Print(err)
	}
	mu.Lock()
	// a bet still on the table is returned next time (save.AtRisk)
	flushSave()
	mu.Unlock()
	log.Print("bye")
}

type game struct {
	last time.Time
	init bool
}

func (g *game) Layout(ow, oh int) (int, int) {
	s := ebiten.Monitor().DeviceScaleFactor()
	return int(math.Ceil(float64(ow) * s)), int(math.Ceil(float64(oh) * s))
}

// view sets the canvas so the 1600x900 stage fits the window, letterboxed.
func view(w, h int) {
	k := math.Min(float64(w)/stageW, float64(h)/stageH)
	cv.k = k
	cv.ox = math.Round((float64(w) - stageW*k) / 2)
	cv.oy = math.Round((float64(h) - stageH*k) / 2)
}

func (g *game) Update() error {
	mu.Lock()
	defer mu.Unlock()
	t := time.Now()
	dt := t.Sub(g.last).Seconds()
	g.last = t

	mx, my := ebiten.CursorPosition()
	mouseX, mouseY = float32(mx), float32(my)
	keys()
	updateCursor()

	if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		if r, ok := topRegion(mouseX, mouseY); ok && r.on {
			pressID = r.id
			if r.id == "vol" {
				dragID = r.id
			} else if r.fn != nil {
				r.fn()
			} else {
				click(r.id)
			}
		}
	}
	if !ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		pressID, dragID = "", ""
	}
	if dragID == "vol" {
		if v, ok := sliderValue("vol", 0, 100, 1); ok && int(v) != save.Settings.Vol {
			save.Settings.Vol = int(v)
			writeSave()
		}
	}
	if _, wy := ebiten.Wheel(); wy != 0 && scrollable != "" && tb.topPanel() == scrollable {
		panelScroll[scrollable] = clamp(panelScroll[scrollable]-wy*60, 0, scrollMax)
	}
	tb.step(dt)
	snd.update()
	return nil
}

func click(id string) {
	t := &tb
	if id != "decks" {
		t.deckOpen = false
	}
	switch {
	case id == "rules", id == "how":
		t.showPanel("rules")
	case id == "credits", id == "credits2":
		t.showPanel("credits")
	case id == "music":
		toggleMusic()
	case id == "sfx":
		save.Settings.Sfx = !save.Settings.Sfx
		writeSave()
	case id == "decks":
		t.deckOpen = !t.deckOpen
	case strings.HasPrefix(id, "deck:"), strings.HasPrefix(id, "decks:"):
		var n int
		fmt.Sscan(id[strings.Index(id, ":")+1:], &n)
		if n != save.Settings.Decks {
			t.setDecks(n)
		}
	case strings.HasPrefix(id, "pace:"):
		save.Settings.Pace = id[5:]
		writeSave()
	case id == "start":
		t.start()
	case strings.HasPrefix(id, "chip:"):
		var v int
		fmt.Sscan(id[5:], &v)
		t.addChip(v)
	case id == "clear":
		t.clearBet()
	case id == "deal":
		t.deal()
	case id == "restock":
		t.restock()
	case id == "hit", id == "stand", id == "double", id == "split":
		t.act(id)
	case strings.HasPrefix(id, "close:"):
		t.closePanel(id[6:])
	}
}

func toggleMusic() {
	save.Settings.Music = !save.Settings.Music
	writeSave()
	if save.Settings.Music {
		snd.startMusic()
	} else {
		snd.stopMusic()
	}
}

func keys() {
	just := inpututil.IsKeyJustPressed
	if just(ebiten.KeyF11) {
		ebiten.SetFullscreen(!ebiten.IsFullscreen())
	}
	if just(ebiten.KeyEscape) {
		tb.closeTop()
		return
	}
	if just(ebiten.KeyM) {
		toggleMusic()
	}
	if tb.topPanel() != "" || tb.deckOpen {
		return
	}
	for k, a := range map[ebiten.Key]string{ebiten.KeyH: "hit", ebiten.KeyS: "stand", ebiten.KeyD: "double", ebiten.KeyP: "split"} {
		if just(k) {
			tb.act(a)
		}
	}
	if (just(ebiten.KeyEnter) || just(ebiten.KeyNumpadEnter) || just(ebiten.KeySpace)) && tb.phase == "bet" {
		tb.deal()
	}
	if just(ebiten.KeyBackspace) && tb.phase == "bet" {
		tb.clearBet()
	}
}

func (g *game) Draw(screen *ebiten.Image) {
	mu.Lock()
	defer mu.Unlock()
	if !g.init {
		initGfx()
		g.init = true
	}
	screen.Fill(hex("#0b0806").std())
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	view(w, h)
	cv.dst = screen
	regs = make([]region, 0, len(lastRegs)+16)
	scrollable = ""

	drawTable()
	drawBar()
	func() {
		defer shift(0, TY)()
		drawParticles()
		drawBanner()
	}()
	for _, id := range append([]string(nil), tb.panels...) {
		switch id {
		case "title":
			drawTitle()
		case "rules":
			drawTextPanel("rules", rulesBlocks())
		case "credits":
			drawTextPanel("credits", creditsBlocks())
		}
	}
	if len(tb.panels) == 0 {
		drawDeckList()
	} else {
		tb.deckOpen = false
	}
	lastRegs = regs
	// letterbox bars
	if cv.ox > 0 {
		fillScreenRect(screen, 0, 0, cv.ox, float64(h))
		fillScreenRect(screen, cv.ox+stageW*cv.k, 0, float64(w), float64(h))
	}
	if cv.oy > 0 {
		fillScreenRect(screen, 0, 0, float64(w), cv.oy)
		fillScreenRect(screen, 0, cv.oy+stageH*cv.k, float64(w), float64(h))
	}
}

func fillScreenRect(dst *ebiten.Image, x0, y0, x1, y1 float64) {
	old := cv
	cv = canvas{dst: dst, k: 1}
	fillRect(x0, y0, x1-x0, y1-y0, hex("#0b0806"))
	cv = old
}
