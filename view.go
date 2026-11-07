// Drawing the toolbar, the table, the console and the panels.
// Positions, sizes, fonts and colours match the original layout.
// (c) 2026 Ironite
package main

import (
	"fmt"
	"image"
	"math"
	"sort"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

var (
	cInk     = hex("#241a10")
	cAccent  = hex("#6c1520")
	cGold    = hex("#e2c172")
	cBarText = hex("#eadfc6")
	cBtnInk  = hex("#f1dfae")
	cEdge    = hex("#a8822e")
	cPale    = hex("#ead6a2")
	cCream   = hex("#ffe7a6")
)

// zoom scales everything drawn until the returned func is called, around the
// logical point (cx, cy). Used for the pop-in animations.
func zoom(cx, cy, s float64) func() {
	old := cv
	cv.ox += cx * cv.k * (1 - s)
	cv.oy += cy * cv.k * (1 - s)
	cv.k *= s
	return func() { cv = old }
}

// clip limits drawing to a logical box until the returned func is called.
func clip(x, y, w, h float64) func() {
	old := cv
	r := cv.dst.Bounds()
	x0, y0 := max(r.Min.X, int(math.Floor(float64(X(x))))), max(r.Min.Y, int(math.Floor(float64(Y(y)))))
	x1, y1 := min(r.Max.X, int(math.Ceil(float64(X(x+w))))), min(r.Max.Y, int(math.Ceil(float64(Y(y+h)))))
	if x1 <= x0 || y1 <= y0 {
		x1, y1 = x0+1, y0+1
	}
	cv.dst = cv.dst.SubImage(image.Rect(x0, y0, x1, y1)).(*ebiten.Image)
	return func() { cv = old }
}

func ascent(f fontID, size float64) float64 { return face(f, size).Metrics().HAscent / cv.k }

// txtBase draws text with its baseline at y.
func txtBase(s string, x, y float64, t T) float64 {
	return txt(s, x, y-ascent(t.F, t.Size)-lineOffset(t.F, t.Size, t.LH), t)
}

// baseOf is the baseline of a line box of height lh (0: normal) at top y.
func baseOf(f fontID, size, y, lh float64) float64 {
	return y + lineOffset(f, size, lh) + ascent(f, size)
}

// drawScaled draws img into the logical box through a colour matrix.
func drawScaled(img *ebiten.Image, x, y, w, h float64, cm colorm.ColorM) {
	b := img.Bounds()
	op := &colorm.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(w*cv.k/float64(b.Dx()), h*cv.k/float64(b.Dy()))
	op.GeoM.Translate(float64(X(x)), float64(Y(y)))
	colorm.DrawImage(cv.dst, img, cm, op)
}

/* =============================================================== toolbar */

var deckOpts = []int{1, 2, 6}

func drawBar() {
	shadow(0, 0, stageW, 56, 0, 0, 3, 12, 0, rgba(0, 0, 0, .55))
	fillRectV(0, 0, stageW, 56, hex("#2a1a12"), hex("#120b07"))
	fillRect(0, 55, stageW, 1, cEdge)

	ty := (56 - normalLH(fDisplay, 30)) / 2
	tt := T{F: fDisplay, Size: 30, LS: 3, Col: cGold, Shad: hex("#000"), ShadY: 1}
	w := txt("Blackjack", 16, ty, tt)
	base := baseOf(fDisplay, 30, ty, 0)
	sw := txtBase("THE GOLD ROOM", 16+w+8, base, T{F: fBody, Size: 12, LS: 3, Col: cBarText.A(.75), Shad: hex("#000"), ShadY: 1})
	bx := 16 + w + 8 + sw + 14 + 8

	const by, bh = 10.5, 35.0
	// shoe size select
	mw := 0.0
	for _, d := range deckOpts {
		mw = max(mw, textW(fBody, 15, deckName(d), 0))
	}
	sw2 := mw + 20 + 2 + 22
	hov, _ := hit("decks", bx, by, sw2, bh, true)
	f := func(c col) col {
		if hov {
			c = c.bright(1.08)
		}
		return c
	}
	boxRR(bx, by, sw2, bh, 7, 1, f(hex("#3a2a1c")), f(hex("#1c130c")), f(cEdge))
	txt(deckName(save.Settings.Decks), bx+11, by+(bh-normalLH(fBody, 15))/2, T{F: fBody, Size: 15, Col: cBtnInk})
	ax, ay := bx+sw2-15, by+bh/2
	fillTri([3][2]float64{{ax - 4.5, ay - 2}, {ax + 4.5, ay - 2}, {ax, ay + 3}}, cBtnInk)
	deckBox = [4]float64{bx, by + bh + 2, sw2, bh}
	bx += sw2 + 8

	btn := func(id, label string, off, icon bool) {
		w := textW(fBody, 15, label, 0) + 28
		if icon {
			w += 17
		}
		barButton(id, label, bx, by, w, bh, off, icon)
		bx += w + 8
	}
	btn("rules", "How to Play", false, false)
	bx += 10 + 8 - 8
	btn("music", "Music", !save.Settings.Music, true)
	slider("vol", bx+2, by+(bh-16)/2, 90, 16, float64(save.Settings.Vol), 0, 100, true)
	bx += 94 + 8
	btn("sfx", "Sounds", !save.Settings.Sfx, false)
	btn("credits", "Credits", false, false)

	// stats, right-aligned
	type stat struct{ k, v string }
	st := []stat{{"Hands", fmtN(save.Stats.Hands)}, {"Won", fmtN(save.Stats.Won)}, {"Blackjacks", fmtN(save.Stats.Blackjacks)}}
	x := 1584.0
	sy := (56 - normalLH(fBody, 16)) / 2
	for i := len(st) - 1; i >= 0; i-- {
		vw := textW(fBodyBold, 16, st[i].v, 0)
		kw := textW(fBody, 16, st[i].k+" ", 0)
		txt(st[i].v, x-vw, sy, T{F: fBodyBold, Size: 16, Col: cGold})
		txt(st[i].k+" ", x-vw-kw, sy, T{F: fBody, Size: 16, Col: cBarText})
		x -= vw + kw + 18
	}
}

var deckBox [4]float64

// drawDeckList is the open shoe-size list under the toolbar select.
func drawDeckList() {
	if !tb.deckOpen {
		return
	}
	// a click anywhere else closes it
	hit("deckclose", 0, 0, stageW, stageH, true)
	onClick("deckclose", func() { tb.deckOpen = false })
	x, y, w := deckBox[0], deckBox[1], deckBox[2]
	rh := 30.0
	h := rh*float64(len(deckOpts)) + 8
	shadow(x, y, w, h, 7, 0, 6, 16, 0, rgba(0, 0, 0, .6))
	boxRR(x, y, w, h, 7, 1, hex("#1c130c"), hex("#1c130c"), cEdge)
	for i, d := range deckOpts {
		ry := y + 4 + float64(i)*rh
		id := fmt.Sprintf("deck:%d", d)
		hov, _ := hit(id, x+1, ry, w-2, rh, true)
		c := cBtnInk
		if hov {
			fillRect(x+1, ry, w-2, rh, cEdge.A(.55))
		}
		if d == save.Settings.Decks {
			c = cCream
			fillCircle(x+8, ry+rh/2, 2.5, cGold)
		}
		txt(deckName(d), x+16, ry+(rh-normalLH(fBody, 15))/2, T{F: fBody, Size: 15, Col: c})
	}
}

/* ================================================================= table */

func drawTable() {
	defer shift(0, TY)()
	drawImg(imgScene, 0, 0, stageW, 844, plain())
	drawPlaque()
	drawShoeLabel()

	cs := append([]*Card(nil), tb.cards.cards...)
	sort.SliceStable(cs, func(a, b int) bool { return cs[a].z < cs[b].z })
	for _, c := range cs {
		drawCard(c)
	}
	drawLabels()
	drawConsole()
	drawFooter()
}

const TY = 56.0

func drawPlaque() {
	const x, y, w = 34.0, 222.0, 250.0
	lhK, lhD, lhR := normalLH(fBody, 12), normalLH(fDisplay, 30), 22.5
	h := 1 + 14 + lhK + 2 + lhD + 6 + 3*lhR + 14 + 1
	fillRRV(x, y, w, h, 10, rgba(10, 30, 20, .55), rgba(5, 20, 12, .55))
	borderRR(x+1, y+1, w-2, h-2, 9, 4, rgba(226, 193, 114, .08))
	borderRR(x, y, w, h, 10, 1, rgba(226, 193, 114, .55))
	cx := x + w/2
	yy := y + 15
	txt("THIS TABLE", cx+2, yy, T{F: fBody, Size: 12, LS: 4, Col: cPale.A(.7), Align: alignC})
	yy += lhK + 2
	txt(deckName(shoeDecksShown()), cx, yy, T{F: fDisplay, Size: 30, Col: hex("#f3dc9a"), Align: alignC})
	yy += lhD + 6
	for i, l := range []string{"Blackjack pays 3:2", "Dealer stands on 17", "Bets 5 – 5,000"} {
		txt(l, cx, yy+float64(i)*lhR, T{F: fBody, Size: 15, Col: cPale.A(.9), Align: alignC, LH: lhR})
	}
}

func shoeDecksShown() int {
	if tb.shoeDecks > 0 {
		return tb.shoeDecks
	}
	return save.Settings.Decks
}

func drawShoeLabel() {
	cx := 1330 + 90.0
	txt(fmtN(len(tb.shoe)), cx, 168, T{F: fNum, Size: 26, Col: cPale, Align: alignC, LH: 26})
	txt("CARDS IN THE SHOE", cx+1, 168+26, T{F: fBody, Size: 12, LS: 2, Col: cPale.A(.75), Align: alignC})
}

func drawCard(c *Card) {
	x, y, r, s, ang := c.x.val(), c.y.val(), c.r.val(), c.s.val(), c.ang.val()
	cx, cy := x+cardW/2, y+cardH/2
	cosA := math.Cos(ang * math.Pi / 180)
	sx := math.Max(math.Abs(cosA), .002)
	img := imgBack
	if cosA >= 0 && c.face >= 0 {
		img = imgCards[c.face]
	}
	drawMask := func(m *ebiten.Image, pad, dy float64, cc col) {
		op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
		op.GeoM.Translate(-pad-cardW/2, -pad-cardH/2+dy)
		op.GeoM.Scale(sx*s, s)
		op.GeoM.Rotate(r * math.Pi / 180)
		op.GeoM.Scale(cv.k, cv.k)
		op.GeoM.Translate(float64(X(cx)), float64(Y(cy)))
		op.ColorScale = cc.scale()
		cv.dst.DrawImage(m, op)
	}
	// .face box-shadow: 0 3px 8px rgba(0,0,0,.4), 0 1px 2px rgba(0,0,0,.3)
	drawMask(softMask(cardW, cardH, 8, 8, 0), 8, 3, rgba(0, 0, 0, .4))
	drawMask(softMask(cardW, cardH, 8, 2, 0), 2, 1, rgba(0, 0, 0, .3))
	b := img.Bounds()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Translate(-float64(b.Dx())/2, -float64(b.Dy())/2)
	k := cardW / float64(b.Dx())
	op.GeoM.Scale(sx*s*k, s*k)
	op.GeoM.Rotate(r * math.Pi / 180)
	op.GeoM.Scale(cv.k, cv.k)
	op.GeoM.Translate(float64(X(cx)), float64(Y(cy)))
	cv.dst.DrawImage(img, op)
}

/* ---------------------------------------------------------------- labels */

func drawLabels() {
	t := &tb
	n := len(t.hands)
	if t.phase == "bet" || t.phase == "title" {
		drawStack(800, 646, t.bet)
	}
	for i, h := range t.hands {
		cx := handCx(i, n)
		if len(h.cards) > 0 {
			drawTot("", h.cards, false, cx, PY+CH+10)
		}
		drawStack(cx, 646, h.bet)
		if t.phase == "play" && i == t.active && n > 1 {
			drawMark(cx, PY-44)
		}
		if h.result != nil {
			drawRes(h.result, cx, PY-46)
		}
	}
	if len(t.dealer.cards) > 0 {
		drawTot("Dealer", t.dealer.cards, !t.holeUp, 800, DY+CH+9)
	}
}

func drawStack(cx, cy, amount float64) {
	all := breakdown(int(amount), 40)
	// the original reverses the list (largest on top) and keeps the top 12
	for i, j := 0, len(all)-1; i < j; i, j = i+1, j-1 {
		all[i], all[j] = all[j], all[i]
	}
	if len(all) > 12 {
		all = all[len(all)-12:]
	}
	for k, v := range all {
		img := imgChip[v]
		y := cy - 36 - float64(k)*5
		var sh ebiten.ColorScale
		sh.Scale(0, 0, 0, .5)
		drawImg(img, cx-36, y+2, 72, 72, sh)
		drawImg(img, cx-36, y, 72, 72, plain())
	}
	if amount > 0 {
		txt(money(amount), cx, cy-36+76, T{F: fNum, Size: 17, Col: hex("#fff0c8"), Align: alignC, Shad: hex("#000"), ShadY: 1, ShadB: 3})
	}
}

func drawTot(label string, ids []int, hideHole bool, cx, top float64) {
	var num, small string
	bust := false
	if hideHole {
		num = fmt.Sprint(tot(ids[:1]))
	} else {
		pt, soft := total(ids)
		num = fmt.Sprint(pt)
		if pt > 21 {
			small, bust = "Bust", true
		} else if soft && pt < 21 {
			small = "Soft"
		}
	}
	if label != "" {
		small = label
		bust = !hideHole && tot(ids) > 21
	}
	nw := textW(fNum, 19, num, 0)
	smw := 0.0
	if small != "" {
		smw = textW(fBody, 12, strings.ToUpper(small), 1) + 6
	}
	lh := normalLH(fNum, 19)
	w, h := 28+smw+nw+2, lh+6+2
	x := cx - w/2
	shadow(x, top, w, h, h/2, 0, 3, 10, 0, rgba(0, 0, 0, .45))
	bg := rgba(8, 6, 4, .82)
	if bust {
		bg = rgba(110, 18, 18, .9)
	}
	fillRR(x, top, w, h, 18, bg)
	borderRR(x, top, w, h, 18, 1, rgba(226, 193, 114, .7))
	base := baseOf(fNum, 19, top+4, 0)
	if small != "" {
		txtBase(strings.ToUpper(small), x+15, base, T{F: fBody, Size: 12, LS: 1, Col: hex("#fff3d6").A(.7)})
	}
	txtBase(num, x+15+smw, base, T{F: fNum, Size: 19, Col: hex("#fff3d6")})
}

func drawRes(r *result, cx, top float64) {
	label := strings.ToUpper(r.text)
	lw := textW(fBodyBold, 15, label, 2)
	aw := 0.0
	if r.amt != "" {
		aw = 8 + textW(fNum, 18, r.amt, 0)
	}
	lh := normalLH(fNum, 18)
	w, h := 32+lw+aw, lh+12
	x := cx - w/2
	var c0, c1, ink col
	switch r.cls {
	case "win":
		c0, c1, ink = hex("#f6dc8a"), hex("#c99a3a"), hex("#2a1a06")
	case "bj":
		c0, c1, ink = hex("#fff2c0"), hex("#e2b440"), hex("#2a1a06")
	case "push":
		c0, c1, ink = hex("#d9d2c2"), hex("#d9d2c2"), hex("#2a2a2a")
	default:
		c0, c1, ink = hex("#5a1414"), hex("#5a1414"), hex("#f5d9d0")
	}
	if r.cls == "bj" {
		shadow(x, top, w, h, 8, 0, 0, 24, 0, rgba(255, 214, 110, .8))
	} else {
		shadow(x, top, w, h, 8, 0, 4, 14, 0, rgba(0, 0, 0, .5))
	}
	fillRRV(x, top, w, h, 8, c0, c1)
	base := baseOf(fNum, 18, top+6, 0)
	txtBase(label, x+16, base, T{F: fBodyBold, Size: 15, LS: 2, Col: ink})
	if r.amt != "" {
		txtBase(r.amt, x+16+lw+8, base, T{F: fNum, Size: 18, Col: ink})
	}
}

func drawMark(cx, top float64) {
	// ▼ bobs 4px once a second (ease-in-out)
	p := math.Mod(now(), 1)
	d := 4 * (1 - math.Cos(p*2*math.Pi)) / 2
	ty := top + d
	sh := rgba(0, 0, 0, .7)
	tri := func(dx, dy float64, c col) {
		fillTri([3][2]float64{{cx - 6.5 + dx, ty + 4 + dy}, {cx + 6.5 + dx, ty + 4 + dy}, {cx + dx, ty + 15 + dy}}, c)
	}
	tri(0, 1, sh)
	tri(0, 0, cGold)
	txt("PLAYING", cx+1.5, top+normalLH(fBody, 14), T{F: fBody, Size: 13, LS: 3, Col: cGold, Align: alignC, Shad: hex("#000"), ShadY: 1, ShadB: 3})
}

/* ---------------------------------------------------------------- console */

const conY = 742.0

func consoleBox(x, y float64) {
	shadow(x, y, 300, 82, 10, 0, 4, 12, 0, rgba(0, 0, 0, .5))
	boxRR(x, y, 300, 82, 10, 1, hex("#1d130c"), hex("#0e0906"), rgba(226, 193, 114, .55))
	insetHi(x, y, 300, 82, 10, rgba(255, 255, 255, .06))
}

func drawConsole() {
	t := &tb
	// bank
	bx, by := 26.0, conY+10
	consoleBox(bx, by)
	yy := by + 1 + 8
	txt("BALANCE", bx+17, yy, T{F: fBody, Size: 11, LS: 4, Col: cPale.A(.65)})
	yy += normalLH(fBody, 11)
	txt(money(save.Balance), bx+17, yy, T{F: fNum, Size: 32, Col: cCream, LH: 33.6})
	yy += 33.6
	betNow := t.bet
	if t.phase != "bet" && t.phase != "title" {
		betNow = 0
		for _, h := range t.hands {
			betNow += h.bet
		}
	}
	base := baseOf(fBody, 14, yy, 0)
	w := txtBase("Bet ", bx+17, base, T{F: fBody, Size: 14, Col: cPale.A(.85)})
	txtBase(money(betNow), bx+17+w, base, T{F: fNum7, Size: 14, Col: hex("#fff3d6").A(.85)})

	// status
	sx := 1600 - 26 - 300.0
	consoleBox(sx, by)
	txt("TABLE", sx+17, by+9, T{F: fBody, Size: 11, LS: 4, Col: cPale.A(.65)})
	lines := wrapRich(t.msg, fBody, fBodyBold, 17, 266)
	drawRich(lines, sx+17, by+9+normalLH(fBody, 11)+3, 21.25, fBody, fBodyBold, 17, hex("#f6ead0"), cCream, 1)

	// centre: chips and buttons, or the play buttons
	cy := conY + 51
	betting := t.phase == "bet"
	if betting || t.phase == "title" {
		if t.broke() {
			drawRestock(cy)
			return
		}
		x := 484.0
		for _, v := range []int{5, 25, 100, 500, 1000} {
			chipButton(v, x, cy-35)
			x += 78
		}
		x += 14 - 8 + 10
		actButton("clear", "Clear", "", x, cy-29, 96, 58, false, t.canClear(), fBody, 18)
		actButton("deal", "Deal", "", x+106, cy-29, 120, 58, true, t.canDeal(), fBodyBold, 18)
		return
	}
	h := t.cur()
	x := 503.0
	for _, b := range []struct {
		id, label, key string
		on             bool
	}{
		{"hit", "Hit", "H", t.playing() && h != nil},
		{"stand", "Stand", "S", t.playing() && h != nil},
		{"double", "Double", "D", t.playing() && t.canDouble(h)},
		{"split", "Split", "P", t.playing() && t.canSplit(h)},
	} {
		actButton(b.id, b.label, b.key, x, cy-29, 138, 58, false, b.on, fBodyBold, 19)
		x += 152
	}
}

var chipLift = map[int]*tween{}

func chipButton(v int, x, y float64) {
	on := tb.canChip(v)
	hov, _ := hit(fmt.Sprintf("chip:%d", v), x, y, 70, 70, on)
	lt := chipLift[v]
	if lt == nil {
		lt = &tween{}
		chipLift[v] = lt
	}
	target := 0.0
	if hov {
		target = -5
	}
	if lt.to != target {
		lt.set(target, .12, easeStd)
	}
	y += lt.val()
	img := imgChip[v]
	var sh ebiten.ColorScale
	sh.Scale(0, 0, 0, .6)
	if on {
		drawImg(img, x, y+3, 70, 70, sh)
		drawImg(img, x, y, 70, 70, plain())
		return
	}
	// grayscale(.9) brightness(.45)
	var cm colorm.ColorM
	cm.ChangeHSV(0, .1, .45)
	var shm colorm.ColorM
	shm.Scale(0, 0, 0, .6*.45)
	drawScaled(img, x, y+3, 70, 70, shm)
	drawScaled(img, x, y, 70, 70, cm)
}

// actButton is the console's .act button; key adds the small shortcut line.
func actButton(id, label, key string, x, y, w, h float64, gold, enabled bool, f fontID, size float64) {
	hov, down := hit(id, x, y, w, h, enabled)
	if down {
		y++
	}
	a := float32(1)
	if !enabled {
		a = .32
	}
	fc := func(c col) col {
		if hov {
			c = c.bright(1.25)
		}
		return c.A(a)
	}
	shadow(x, y, w, h, 10, 0, 3, 8, 0, rgba(0, 0, 0, .55*a))
	ink := cBtnInk
	if gold {
		boxRR(x, y, w, h, 10, 1, fc(hex("#f3d98c")), fc(hex("#c4953a")), fc(hex("#f6e2a4")))
		ink = hex("#2a1a06")
	} else {
		boxRR(x, y, w, h, 10, 1, fc(hex("#3b2a1c")), fc(hex("#1a110b")), fc(cEdge))
	}
	insetHi(x, y, w, h, 10, fc(rgba(255, 255, 255, .1)))
	ls := 0.0
	if key != "" {
		ls = 1
	}
	lh := normalLH(f, size)
	if key == "" {
		txt(label, x+w/2, y+(h-lh)/2, T{F: f, Size: size, LS: ls, Col: fc(ink), Align: alignC})
		return
	}
	kh := normalLH(fBody, 11)
	top := y + (h-(lh+1+kh))/2
	txt(label, x+w/2+ls/2, top, T{F: f, Size: size, LS: ls, Col: fc(ink), Align: alignC})
	txt(key, x+w/2, top+lh+1, T{F: fBody, Size: 11, Col: fc(ink).A(.55), Align: alignC})
}

func drawRestock(cy float64) {
	label := "Restock Chips"
	bw := max(120, textW(fBodyBold, 18, label, 0)+14)
	maxSpan := 560 - 14 - bw
	lines := wrapRich("Out of chips? The house will stake you another <b>1,000</b> play chips.", fBody, fBodyBold, 15, maxSpan)
	sw := 0.0
	for _, l := range lines {
		sw = max(sw, l.w)
	}
	lh := normalLH(fBody, 15)
	total := sw + 14 + bw
	x := 348 + (904-total)/2
	drawRich(lines, x, cy-float64(len(lines))*lh/2, lh, fBody, fBodyBold, 15, cPale, cCream, 1)
	actButton("restock", label, "", x+sw+14, cy-29, bw, 58, true, true, fBodyBold, 18)
}

func drawFooter() {
	// now playing (fades in for five seconds when a new tune starts)
	if snd.nowName != "" {
		e := now() - snd.nowSince
		a := 0.0
		switch {
		case e < 1:
			a = e
		case e < 5:
			a = 1
		case e < 6:
			a = 6 - e
		}
		if a > 0 {
			lh := normalLH(fBody, 13)
			y := 844 - 3 - lh
			c := hex("#d9c69a").A(float32(.75 * a))
			noteIcon(16+6, y+lh/2, c)
			txt(snd.nowName, 16+15, y, T{F: fBody, Size: 13, Col: c})
		}
	}
	v := strings.ToUpper("v" + version + " · © 2026 Ironite · play money only")
	lh := normalLH(fBody, 12)
	txt(v, 1600-14, 844-3-lh, T{F: fBody, Size: 12, LS: 2, Col: hex("#d9c69a").A(.55), Align: alignR})
}

/* ------------------------------------------------------------ fx, banner */

func drawParticles() {
	for _, p := range tb.fx.p {
		a := float32(max(0, 1-p.t/p.life))
		c := p.c.A(a)
		if p.circle {
			fillCircle(p.x, p.y, p.s/2, c)
			continue
		}
		cs, sn := math.Cos(p.r), math.Sin(p.r)
		hw, hh := p.s/2, p.s/4
		pt := func(dx, dy float64) (float32, float32) {
			return X(p.x + dx*cs - dy*sn), Y(p.y + dx*sn + dy*cs)
		}
		var path vector.Path
		x0, y0 := pt(-hw, -hh)
		path.MoveTo(x0, y0)
		x1, y1 := pt(hw, -hh)
		path.LineTo(x1, y1)
		x2, y2 := pt(hw, hh)
		path.LineTo(x2, y2)
		x3, y3 := pt(-hw, hh)
		path.LineTo(x3, y3)
		path.Close()
		fillPathV(&path, 0, 1, c, c)
	}
}

func drawBanner() {
	b := &tb.banner
	p := b.show.val()
	if p <= .01 {
		return
	}
	a := float32(p)
	lh1 := normalLH(fDisplay, 62)
	lh2 := normalLH(fBody, 22)
	w := textW(fDisplay, 62, b.title, 3)
	h := 10 + lh1 + 14
	if b.sub != "" {
		w = max(w, textW(fBody, 22, b.sub, 1))
		h += 4 + lh2
	}
	w += 92 + 2
	h += 2
	cx, cy := 800.0, 352.0
	defer zoom(cx, cy, .8+.2*p)()
	x, y := cx-w/2, cy-h/2
	shadow(x, y, w, h, 16, 0, 10, 40, 0, rgba(0, 0, 0, .6*a))
	fillRR(x, y, w, h, 16, rgba(6, 16, 10, .84*a))
	borderRR(x, y, w, h, 16, 1, rgba(226, 193, 114, .6*a))
	sh := rgba(0, 0, 0, .4*a)
	txt(b.title, cx+1.5, y+11, T{F: fDisplay, Size: 62, LS: 3, Col: cCream.A(a), Align: alignC, Shad: sh, ShadY: 3})
	if b.sub != "" {
		txt(b.sub, cx+.5, y+11+lh1+4, T{F: fBody, Size: 22, LS: 1, Col: hex("#ffffff").A(a), Align: alignC, Shad: sh, ShadY: 3})
	}
}

/* ================================================================ panels */

var veilImg *ebiten.Image

func drawVeil() {
	if veilImg == nil {
		const w, h = 160, 90
		pix := make([]byte, w*h*4)
		for y := 0; y < h; y++ {
			for x := 0; x < w; x++ {
				// radial-gradient(ellipse at center, rgba(10,6,3,.35), rgba(5,3,2,.82))
				dx, dy := (float64(x)+.5)/w*2-1, (float64(y)+.5)/h*2-1
				t := min(1, math.Hypot(dx, dy)/math.Sqrt2)
				r, g, b := 10+(5-10)*t, 6+(3-6)*t, 3+(2-3)*t
				a := .35 + (.82-.35)*t
				i := (y*w + x) * 4
				pix[i], pix[i+1], pix[i+2], pix[i+3] = byte(r*a), byte(g*a), byte(b*a), byte(a*255)
			}
		}
		veilImg = ebiten.NewImage(w, h)
		veilImg.WritePixels(pix)
	}
	drawImg(veilImg, 0, 0, stageW, stageH, plain())
	blockArea(0, 0, stageW, stageH)
}

// paperBox is the .box panel: cream paper, gold edge, a dark frame ring and
// a thin gold outer line.
func paperBox(x, y, w, h float64) {
	shadow(x, y, w, h, 14, 0, 30, 80, 0, rgba(0, 0, 0, .8))
	fillRR(x-7, y-7, w+14, h+14, 21, cEdge)
	fillRR(x-6, y-6, w+12, h+12, 20, rgba(20, 12, 6, 1))
	fillRR(x-6, y-6, w+12, h+12, 20, rgba(20, 12, 6, .9))
	boxRR(x, y, w, h, 14, 1, hex("#f8f2e4"), hex("#e8dcc0"), cEdge)
}

/* --------------------------------------------------------- title panel */

func drawTitle() {
	drawVeil()
	const w = 820.0
	ink := cInk
	lhKick, lhH1, lhH2 := normalLH(fBody, 14), 100.8, normalLH(fBody, 17)
	optH := normalLH(fBody, 15) + 14 + 2
	purseH := max(normalLH(fBody, 18), normalLH(fNum, 22))
	bigH := normalLH(fDisplay, 22) + 24 + 2
	var notice []richLine
	noticeH := 0.0
	if tb.notice != "" {
		notice = wrapRich(tb.notice, fBody, fBodyBold, 15, 600-30)
		noticeH = float64(len(notice))*normalLH(fBody, 15) + 18
	}
	h := 1 + 36 + lhKick + lhH1 + 6 + lhH2 + 18 + 54 + 16 + optH*2 + 10 + 20 + purseH + 16 +
		bigH + 16 + normalLH(fBody, 15) + 12 + normalLH(fBody, 13) + 20 + 26 + 28 + 1
	if notice != nil {
		h += -6 + noticeH + 14
	}
	x0, y0 := (stageW-w)/2, math.Round((stageH-h)/2)
	paperBox(x0, y0, w, h)
	cx := 800.0
	y := y0 + 1 + 36
	txt("WELCOME TO THE GOLD ROOM", cx+2, y, T{F: fBody, Size: 14, LS: 4, Col: ink.A(.7), Align: alignC})
	y += lhKick
	txt("Blackjack", cx+3, y, T{F: fDisplay, Size: 96, LS: 6, Col: cAccent, Align: alignC, LH: lhH1})
	y += lhH1 + 6
	txt("TWENTY-ONE · PLAYER VS. DEALER", cx+2, y, T{F: fBody, Size: 17, LS: 4, Col: ink.A(.8), Align: alignC})
	y += lhH2 + 18 // h2's 18px margin and the chips' 4px collapse
	chx := cx - (5*54+4*6)/2
	for _, v := range []int{5, 25, 100, 500, 1000} {
		var sh ebiten.ColorScale
		sh.Scale(0, 0, 0, .3)
		drawImg(imgChip[v], chx, y+2, 54, 54, sh)
		drawImg(imgChip[v], chx, y, 54, 54, plain())
		chx += 60
	}
	y += 54 + 16 // 6px and 16px margins collapse
	optRow("Shoe", y, []string{"1", "2", "6"}, []string{"Single Deck", "Double Deck", "6-Deck Shoe"}, fmt.Sprint(save.Settings.Decks), "decks:")
	y += optH + 10
	optRow("Dealing", y, []string{"relaxed", "steady", "quick"}, []string{"Relaxed", "Steady", "Quick"}, save.Settings.Pace, "pace:")
	y += optH + 20

	// Your balance: 1,000 chips
	v := money(save.Balance)
	a, b := "Your balance: ", " chips"
	aw, vw, bw := textW(fBody, 18, a, 0), textW(fNum7, 22, v, 0), textW(fBody, 18, b, 0)
	x := cx - (aw+vw+bw)/2
	base := baseOf(fNum, 22, y, 0)
	txtBase(a, x, base, T{F: fBody, Size: 18, Col: ink})
	txtBase(v, x+aw, base, T{F: fNum7, Size: 22, Col: cAccent})
	txtBase(b, x+aw+vw, base, T{F: fBody, Size: 18, Col: ink})
	y += purseH + 16

	if notice != nil {
		y -= 6
		fillRR(cx-300, y, 600, noticeH, 8, rgba(108, 21, 32, .1))
		borderRR(cx-300, y, 600, noticeH, 8, 1, rgba(108, 21, 32, .35))
		for j, l := range notice {
			txt(joinAll(l), cx, y+9+float64(j)*normalLH(fBody, 15), T{F: fBody, Size: 15, Col: ink, Align: alignC})
		}
		y += noticeH + 14
	}
	bw2 := bigButtonW("Take a Seat", 22, 34)
	bigButton("start", "Take a Seat", cx-bw2/2, y, bw2, bigH, 22, true, false, col{})
	y += bigH + 16

	links := []struct{ id, label string }{{"how", "How to Play"}, {"credits2", "Credits & Changelog"}}
	sep := " · "
	sw := textW(fBody, 15, sep, 0)
	total := -sw
	for _, l := range links {
		total += textW(fBody, 15, l.label, 0) + sw + 12
	}
	lx := cx - total/2
	for i, l := range links {
		lw := textW(fBody, 15, l.label, 0) + 12
		hov, _ := hit(l.id, lx, y, lw, 20, true)
		c := cAccent
		if hov {
			c = c.bright(1.4)
		}
		txt(l.label, lx+6, y, T{F: fBody, Size: 15, Col: c})
		fillRect(lx+6, y+18.5, lw-12, 1, c)
		lx += lw
		if i < len(links)-1 {
			txt(sep, lx, y, T{F: fBody, Size: 15, Col: ink})
			lx += sw
		}
	}
	y += normalLH(fBody, 15) + 12
	txt("Play money only. No real money, no purchases, no accounts.", cx, y, T{F: fBody, Size: 13, Col: ink.A(.65), Align: alignC})
	y += normalLH(fBody, 13) + 20
	// An [Ironite] game: the logo is 26px tall including its 6px/11px padding
	lb := imgLogo.Bounds()
	logoW := 14 * float64(lb.Dx()) / float64(lb.Dy())
	t := T{F: fBody, Size: 12, LS: 3, Col: ink.A(.9)}
	aw2, gw := textW(fBody, 12, "AN", 3), textW(fBody, 12, "GAME", 3)
	tw := aw2 + 8 + logoW + 22 + 8 + gw
	bx := cx - tw/2
	ty := y + (26-normalLH(fBody, 12))/2
	txt("AN", bx, ty, t)
	fillRR(bx+aw2+8, y, logoW+22, 26, 6, hex("#003b2f"))
	drawImg(imgLogo, bx+aw2+8+11, y+6, logoW, 14, plain())
	txt("GAME", bx+aw2+8+logoW+22+8, ty, t)
}

func joinAll(l richLine) string {
	var b strings.Builder
	for i, p := range l.pieces {
		if i > 0 {
			b.WriteString(" ")
		}
		b.WriteString(p.s)
	}
	return b.String()
}

// optRow draws "Label [a|b|c]" segmented choices, centred.
func optRow(label string, y float64, vals, labels []string, cur, idPrefix string) {
	ws := make([]float64, len(labels))
	gw := 2.0
	for i, l := range labels {
		ws[i] = textW(fBody, 15, l, 0) + 28
		gw += ws[i]
		if i < len(labels)-1 {
			gw++
		}
	}
	bh := normalLH(fBody, 15) + 14 + 2
	x := 800 - (140+12+gw)/2
	txt(label, x+140, y+(bh-normalLH(fBody, 16))/2, T{F: fBody, Size: 16, Col: cInk.A(.8), Align: alignR})
	gx := x + 152
	fillRR(gx, y, gw, bh, 8, cEdge)
	defer clip(gx+1, y+1, gw-2, bh-2)()
	bx := gx + 1
	for i, l := range labels {
		on := vals[i] == cur
		hov, _ := hit(idPrefix+vals[i], bx, y+1, ws[i], bh-2, true)
		bg := rgba(255, 255, 255, .45)
		fg := cInk
		if on {
			bg, fg = cAccent, hex("#ffffff")
		}
		fillRect(bx, y+1, ws[i], bh-2, hex("#efe6d2"))
		fillRect(bx, y+1, ws[i], bh-2, bg)
		if hov && !on {
			fillRect(bx, y+1, ws[i], bh-2, rgba(255, 255, 255, .25))
		}
		txt(l, bx+ws[i]/2, y+1+(bh-2-normalLH(fBody, 15))/2, T{F: fBody, Size: 15, Col: fg, Align: alignC})
		bx += ws[i]
		if i < len(labels)-1 {
			fillRect(bx, y+1, 1, bh-2, rgba(0, 0, 0, .15))
			bx++
		}
	}
}

/* ---------------------------------------------------------- text panels */

// A text panel is a list of blocks laid out like the original HTML.
type pblock struct {
	kind   string // h3, p, li, muted, cred, tunes, h4, h5, bullet, button
	text   string
	a      string
	n      int
	mt, mb float64
}

func rulesBlocks() []pblock {
	return []pblock{
		{kind: "h3", text: "How to Play Blackjack", mb: 12},
		{kind: "p", text: "Get closer to <b>21</b> than the dealer without going over. Number cards count their number, J, Q and K count 10, and an Ace counts 1 or 11, whichever helps. A hand with an Ace counted as 11 is called <b>soft</b>.", mt: 17, mb: 17},
		{kind: "li", n: 1, text: "<b>Bet.</b> Click chips to build your bet in the circle, then press <b>Deal</b>. Your bet stays up for the next hand; press Clear to change it.", mt: 17, mb: 6},
		{kind: "li", n: 2, text: "<b>The deal.</b> You get two cards face up. The dealer gets one up and one face down. If the dealer shows an Ace or a ten, they check for Blackjack first.", mt: 6, mb: 6},
		{kind: "li", n: 3, text: "<b>Hit</b> takes another card. <b>Stand</b> keeps your total. Go over 21 and you <b>bust</b>, losing the bet.", mt: 6, mb: 6},
		{kind: "li", n: 4, text: "<b>Double Down</b> on your first two cards: double your bet and take exactly one more card.", mt: 6, mb: 6},
		{kind: "li", n: 5, text: "<b>Split</b> two cards of the same value into two hands, each with its own bet equal to the first. You can split up to four hands. Split Aces get one card each, and 21 after a split counts as 21, not Blackjack.", mt: 6, mb: 6},
		{kind: "li", n: 6, text: "<b>The dealer</b> turns over the hidden card, draws to 16 and <b>stands on all 17s</b>.", mt: 6, mb: 6},
		{kind: "li", n: 7, text: "<b>Payouts.</b> A win pays <b>1 to 1</b>. A Blackjack (Ace and a ten-card on your first two cards) pays <b>3 to 2</b>. A tie is a <b>push</b> and your bet comes back.", mt: 6, mb: 17},
		{kind: "p", text: "You start with <b>1,000</b> play chips and your balance is saved between sessions. If you run out, the house restocks you for free. These chips have no cash value.", mt: 17, mb: 17},
		{kind: "muted", text: "Keys: Enter deal · H hit · S stand · D double · P split · Backspace clear bet · M music · F11 full screen", mt: 17, mb: 0},
		{kind: "button", text: "Back to the Table", mt: 14},
	}
}

func creditsBlocks() []pblock {
	bs := []pblock{{kind: "h3", text: "Credits", mb: 12}}
	cred := [][2]string{{"Developed & Published by", "Ironite"}, {"Game Design", "Ironite"}, {"Card Art & Illustration", "Ironite"}, {"Music", "Ironite"}}
	for i, c := range cred {
		mt := 0.0
		if i == 0 {
			mt = 12
		}
		bs = append(bs, pblock{kind: "cred", text: c[0], a: c[1], mt: mt})
	}
	var tn []string
	for _, n := range tuneNames {
		tn = append(tn, "\""+n+"\"")
	}
	bs = append(bs, pblock{kind: "tunes", text: strings.Join(tn, ", ")})
	for _, c := range [][2]string{{"Sound Effects", "Ironite"}, {"Programming", "Ironite"}, {"Typefaces", "Italiana, Lora & Outfit (SIL Open Font License)"}} {
		bs = append(bs, pblock{kind: "cred", text: c[0], a: c[1]})
	}
	bs = append(bs,
		pblock{kind: "muted", text: "Copyright © 2026 Ironite. All rights reserved. Play money only: chips have no cash value.", mt: 14, mb: 14},
		pblock{kind: "h3", text: "Changelog", mt: 14, mb: 12})
	inList := false
	for _, l := range strings.Split(strings.ReplaceAll(changelog, "\r", ""), "\n") {
		switch {
		case strings.HasPrefix(l, "## "):
			ver, date := l[3:], ""
			if i := strings.Index(l, "]"); strings.HasPrefix(l, "## [") && i > 0 {
				ver = "Version " + l[4:i]
				date = strings.TrimSpace(strings.TrimLeft(l[i+1:], " -"))
			}
			bs = append(bs, pblock{kind: "h4", text: ver, a: date, mt: 16, mb: 2})
			inList = false
		case strings.HasPrefix(l, "### "):
			bs = append(bs, pblock{kind: "h5", text: strings.ToUpper(l[4:]), mt: 8, mb: 2})
			inList = false
		case strings.HasPrefix(l, "- "):
			mt := 6.0
			if !inList {
				mt = 2
			}
			inList = true
			item := l[2:]
			for strings.Contains(item, "**") {
				item = strings.Replace(item, "**", "<b>", 1)
				item = strings.Replace(item, "**", "</b>", 1)
			}
			bs = append(bs, pblock{kind: "bullet", text: item, mt: mt, mb: 6})
		}
	}
	return append(bs, pblock{kind: "button", text: "Close", mt: 14})
}

var (
	scrollable string
	scrollMax  float64
)

// drawTextPanel lays out and draws a scrollable text panel.
func drawTextPanel(id string, blocks []pblock) {
	drawVeil()
	const w, padX, padY = 860.0, 46.0, 34.0
	const inner = w - 2*padX - 2
	h3LH := normalLH(fDisplay, 38)
	btnH := normalLH(fDisplay, 18) + 18 + 2
	type laid struct {
		b     pblock
		y, h  float64
		lines []richLine
	}
	var ls []laid
	y, prevMB := 0.0, 0.0
	for i, b := range blocks {
		gap := max(prevMB, b.mt)
		if i == 0 {
			gap = 0
		}
		y += gap
		l := laid{b: b, y: y}
		switch b.kind {
		case "h3":
			l.h = h3LH
		case "p":
			l.lines = wrapRich(b.text, fBody, fBodyBold, 17, inner)
			l.h = float64(len(l.lines)) * 25.5
		case "li", "bullet":
			l.lines = wrapRich(b.text, fBody, fBodyBold, 17, inner-24)
			l.h = float64(len(l.lines)) * 25.5
		case "muted":
			l.lines = wrapRich(b.text, fBody, fBodyBold, 14, inner)
			l.h = float64(len(l.lines)) * 21
		case "cred":
			l.h = 25.5 + 12 + 1
			if textW(fBodyBold, 17, b.a, 0) > inner-textW(fBody, 17, b.text, 0)-20 {
				l.lines = wrapRich("<b>"+b.a+"</b>", fBody, fBodyBold, 17, inner-textW(fBody, 17, b.text, 0)-20)
				l.h = 13 + float64(len(l.lines))*25.5
			}
		case "tunes":
			l.h = 22.5 + 6
		case "h4":
			l.h = normalLH(fDisplay, 21)
		case "h5":
			l.h = 21
		case "button":
			l.h = btnH
		}
		y += l.h
		prevMB = b.mb
		ls = append(ls, l)
	}
	contentH := y
	h := min(800, contentH+2*padY+2)
	x0, y0 := (stageW-w)/2, math.Round((stageH-h)/2)
	paperBox(x0, y0, w, h)
	blockArea(x0, y0, w, h)

	viewH := h - 2
	maxScroll := max(0, contentH+2*padY-viewH)
	sc := clamp(panelScroll[id], 0, maxScroll)
	panelScroll[id] = sc
	restore := clip(x0+1, y0+1, w-2, viewH)
	cx := x0 + 1 + padX
	top := y0 + 1 + padY - sc
	ink := cInk
	for _, l := range ls {
		yy := top + l.y
		if (yy > y0+h+50 || yy+l.h < y0-50) && l.b.kind != "button" {
			continue
		}
		b := l.b
		switch b.kind {
		case "h3":
			txt(b.text, cx, yy, T{F: fDisplay, Size: 38, Col: cAccent})
		case "p":
			drawRich(l.lines, cx, yy, 25.5, fBody, fBodyBold, 17, ink, ink, 1)
		case "li":
			txt(fmt.Sprintf("%d.", b.n), cx+24-6, yy, T{F: fBody, Size: 17, Col: ink, Align: alignR, LH: 25.5})
			drawRich(l.lines, cx+24, yy, 25.5, fBody, fBodyBold, 17, ink, ink, 1)
		case "bullet":
			fillCircle(cx+24-14, yy+13.5, 2.6, ink)
			drawRich(l.lines, cx+24, yy, 25.5, fBody, fBodyBold, 17, ink, ink, 1)
		case "muted":
			drawRich(l.lines, cx, yy, 21, fBody, fBodyBold, 14, ink, ink, .7)
		case "cred":
			txt(b.text, cx, yy+6, T{F: fBody, Size: 17, Col: ink.A(.75), LH: 25.5})
			if l.lines != nil {
				for j, rl := range l.lines {
					txt(joinAll(rl), cx+inner, yy+6+float64(j)*25.5, T{F: fBodyBold, Size: 17, Col: ink, Align: alignR, LH: 25.5})
				}
			} else {
				txt(b.a, cx+inner, yy+6, T{F: fBodyBold, Size: 17, Col: ink, Align: alignR, LH: 25.5})
			}
			dashed(cx, yy+l.h-1, inner)
		case "tunes":
			txt(b.text, cx+18, yy, T{F: fBody, Size: 15, Col: ink.A(.85), LH: 22.5})
		case "h4":
			tw := txt(b.text, cx, yy, T{F: fDisplay, Size: 21, Col: cAccent})
			if b.a != "" {
				txtBase(b.a, cx+tw+6, baseOf(fDisplay, 21, yy, 0), T{F: fBody, Size: 13, Col: ink.A(.7)})
			}
		case "h5":
			txt(b.text, cx, yy, T{F: fBody, Size: 14, LS: 1, Col: ink.A(.7), LH: 21})
		case "button":
			bw := bigButtonW(b.text, 18, 26)
			bigButton("close:"+id, b.text, x0+w/2-bw/2, yy, bw, btnH, 18, true, false, col{})
		}
	}
	restore()
	if maxScroll > 0 {
		th := viewH * viewH / (contentH + 2*padY)
		ty := y0 + 1 + (viewH-th)*sc/maxScroll
		fillRR(x0+w-14, ty+4, 7, th-8, 3.5, rgba(60, 36, 16, .35))
	}
	scrollable = id
	scrollMax = maxScroll
}

func dashed(x, y, w float64) {
	for xx := x; xx < x+w; xx += 6 {
		fillRect(xx, y, min(3, x+w-xx), 1, rgba(0, 0, 0, .18))
	}
}
