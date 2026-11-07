// Blackjack artwork: The Gold Room table, classic court deck, casino chips.
// Everything is drawn in code, shape for shape the same as the original
// table art. (c) 2026 Ironite
package main

import (
	"fmt"
	"math"
)

const (
	PI  = math.Pi
	TAU = 2 * math.Pi
	RH  = 1.62 // portrait unit height (width 1)
)

/* ---------------------------------------------------------------- utils */

// rng is the original xorshift generator (bit-for-bit, including its
// signed shift), so speckles and felt nap land in the same places.
func rng(seed uint32) func() float64 {
	s := seed
	if s == 0 {
		s = 1
	}
	return func() float64 {
		x := int32(s)
		x ^= x << 13
		x ^= x >> 17
		x ^= x << 5
		s = uint32(x)
		return float64(s) / 4294967296
	}
}

func suitOf(id int) int { return (id / 13) % 4 }
func rankOf(id int) int { return id%13 + 1 }
func isRed(id int) bool { s := suitOf(id); return s == 1 || s == 2 }
func cardVal(id int) int {
	r := rankOf(id)
	if r == 1 {
		return 11
	}
	return min(10, r)
}

var RANKS = []string{"A", "2", "3", "4", "5", "6", "7", "8", "9", "10", "J", "Q", "K"}

/* ---------------------------------------------------------------- suits */

func (c *Ctx) suitPath(s int) {
	c.Begin()
	switch s {
	case 1:
		c.M(0, .42)
		c.CubicTo(-.15, .28, -.5, .05, -.5, -.18)
		c.CubicTo(-.5, -.38, -.32, -.48, -.2, -.48)
		c.CubicTo(-.08, -.48, 0, -.38, 0, -.28)
		c.CubicTo(0, -.38, .08, -.48, .2, -.48)
		c.CubicTo(.32, -.48, .5, -.38, .5, -.18)
		c.CubicTo(.5, .05, .15, .28, 0, .42)
	case 2:
		c.M(0, -.5)
		c.QuadraticTo(.16, -.2, .38, 0)
		c.QuadraticTo(.16, .2, 0, .5)
		c.QuadraticTo(-.16, .2, -.38, 0)
		c.QuadraticTo(-.16, -.2, 0, -.5)
	case 0:
		c.M(0, -.5)
		c.CubicTo(-.15, -.33, -.5, -.12, -.5, .1)
		c.CubicTo(-.5, .28, -.36, .36, -.24, .36)
		c.CubicTo(-.12, .36, -.05, .3, -.03, .22)
		c.CubicTo(-.05, .34, -.12, .44, -.22, .5)
		c.LineTo(.22, .5)
		c.CubicTo(.12, .44, .05, .34, .03, .22)
		c.CubicTo(.05, .3, .12, .36, .24, .36)
		c.CubicTo(.36, .36, .5, .28, .5, .1)
		c.CubicTo(.5, -.12, .15, -.33, 0, -.5)
	default:
		c.Arc(0, -.24, .21, 0, TAU, false)
		c.M(-.03, .08)
		c.Arc(-.24, .08, .21, 0, TAU, false)
		c.M(.45, .08)
		c.Arc(.24, .08, .21, 0, TAU, false)
		c.M(-.1, -.1)
		c.LineTo(.1, -.1)
		c.LineTo(.1, .1)
		c.LineTo(-.1, .1)
		c.Z()
		c.M(-.04, .1)
		c.CubicTo(-.05, .3, -.12, .44, -.22, .5)
		c.LineTo(.22, .5)
		c.CubicTo(.12, .44, .05, .3, .04, .1)
		c.Z()
	}
}

type suitColors struct {
	red, black []any
	edge       string
}

func (c *Ctx) drawSuit(s int, cx, cy, size float64, flip bool, sc *suitColors) {
	c.Save()
	c.Translate(cx, cy)
	if flip {
		c.Rotate(PI)
	}
	c.Scale(size, size)
	c.suitPath(s)
	red := s == 1 || s == 2
	st := []any{"#3d3832", "#110f0d"}
	if red {
		st = []any{"#d23a2f", "#8a1913"}
	}
	if sc != nil {
		if red && sc.red != nil {
			st = sc.red
		} else if !red && sc.black != nil {
			st = sc.black
		}
	}
	c.Fill(lin(0, -.5, 0, .5, st...))
	if sc != nil && sc.edge != "" {
		c.Stroke(sc.edge, .025)
	}
	c.Restore()
	c.Begin()
}

/* ------------------------------------------------------------ portraits */

type hatOpt struct {
	typ                         string
	color, band, jewel, feather string
}

type portraitOpt struct {
	skin, skinShade, hair, beard, mustache, brows, lips, earring, eyes string
	coat                                                               []any
	beardLen                                                           float64
	longHair, lashes, blush, curls                                     bool
	hat                                                                *hatOpt
	bg, body, collar, front                                            func(c *Ctx)
}

func or(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

func (x *Ctx) portrait(o portraitOpt) {
	skin, shade := or(o.skin, "#e6b892"), or(o.skinShade, "#c99470")
	hair := or(o.hair, "#6b4426")
	coat := o.coat
	if coat == nil {
		coat = []any{"#3a4f6a", "#1f2b3c"}
	}
	if o.bg != nil {
		o.bg(x)
	} else {
		x.FillRect(0, 0, 1, RH, lin(0, 0, 0, RH, "#f4e8cf", "#d9c6a0"))
	}
	if o.longHair {
		x.Ell(.5, .9, .25, .3, hair)
		x.FillRect(.25, .9, .5, .38, hair)
	}
	// body / coat
	x.Begin()
	x.M(-.05, RH)
	x.LineTo(0, 1.3)
	x.QuadraticTo(.08, 1.12, .3, 1.09)
	x.LineTo(.7, 1.09)
	x.QuadraticTo(.92, 1.12, 1, 1.3)
	x.LineTo(1.05, RH)
	x.Z()
	x.Fill(lin(0, 1.1, 1, RH, coat...))
	if o.body != nil {
		o.body(x)
	}
	// neck
	x.FillRect(.42, .95, .16, .17, shade)
	if o.collar != nil {
		o.collar(x)
	}
	// ears + head
	x.Ell(.35, .86, .033, .05, shade)
	x.Ell(.65, .86, .033, .05, shade)
	x.Ell(.5, .85, .15, .185, skin)
	x.Ell(.5, .85, .15, .185, rad(.45, .8, .02, .22, "cssRGBA(255,255,255,.16)", "cssRGBA(120,60,20,.16)"))
	if o.blush {
		x.Ell(.42, .9, .03, .02, "cssRGBA(220,90,90,.3)")
		x.Ell(.58, .9, .03, .02, "cssRGBA(220,90,90,.3)")
	}
	if o.beard != "" {
		bl := o.beardLen
		x.Begin()
		x.M(.35, .85)
		x.QuadraticTo(.32, 1.18+bl, .5, 1.25+bl)
		x.QuadraticTo(.68, 1.18+bl, .65, .85)
		x.QuadraticTo(.6, .97, .5, .965)
		x.QuadraticTo(.4, .97, .35, .85)
		x.Fill(o.beard)
		for i := 0; i < 10; i++ {
			bx := .38 + float64(i)*.026
			x.Line(bx, .96, bx+(.5-bx)*.2, 1.12+math.Sin(float64(i)*1.7)*.04+bl*.8, "cssRGBA(0,0,0,.12)", .006)
		}
	}
	if o.mustache != "" {
		x.Begin()
		x.M(.5, .925)
		x.QuadraticTo(.42, .885, .36, .97)
		x.QuadraticTo(.44, .945, .5, .965)
		x.QuadraticTo(.56, .945, .64, .97)
		x.QuadraticTo(.58, .885, .5, .925)
		x.Fill(o.mustache)
	}
	// nose
	x.Line(.5, .85, .492, .9, "cssRGBA(150,85,55,.6)", .009)
	x.Ell(.5, .9, .022, .014, "cssRGBA(160,90,60,.25)")
	// eyes
	ey, ec := .82, or(o.eyes, "#3b2a1a")
	x.Ell(.445, ey, .02, .014, "#fff")
	x.Ell(.555, ey, .02, .014, "#fff")
	x.Ell(.447, ey+.002, .011, .012, ec)
	x.Ell(.557, ey+.002, .011, .012, ec)
	x.Ell(.443, ey-.003, .004, .004, "#fff")
	x.Ell(.553, ey-.003, .004, .004, "#fff")
	if o.lashes {
		x.Line(.42, ey-.012, .465, ey-.016, "#2a1a10", .008)
		x.Line(.535, ey-.016, .58, ey-.012, "#2a1a10", .008)
	}
	bc := or(o.brows, hair)
	x.Line(.415, ey-.04, .475, ey-.045, bc, .014)
	x.Line(.585, ey-.04, .525, ey-.045, bc, .014)
	// mouth
	if o.mustache == "" || o.lips != "" {
		if o.lips != "" {
			x.Ell(.5, .955, .036, .015, o.lips)
			x.Line(.468, .955, .532, .955, "cssRGBA(90,20,20,.5)", .004)
		} else {
			x.Begin()
			x.Arc(.5, .925, .045, .35, PI-.35, false)
			x.Stroke("cssRGBA(120,50,40,.75)", .01)
		}
	}
	if o.earring != "" {
		x.Ell(.352, .92, .013, .013, o.earring)
		x.Ell(.648, .92, .013, .013, o.earring)
	}
	// hair on top
	x.Begin()
	x.M(.35, .84)
	x.QuadraticTo(.36, .64, .5, .645)
	x.QuadraticTo(.64, .64, .65, .84)
	x.QuadraticTo(.6, .72, .5, .73)
	x.QuadraticTo(.41, .72, .35, .84)
	x.Fill(hair)
	if o.curls {
		for _, p := range [][2]float64{{.31, .97}, {.69, .97}, {.3, 1.07}, {.7, 1.07}} {
			x.Ell(p[0], p[1], .052, .052, hair)
			x.Ell(p[0]-.012, p[1]-.012, .018, .018, "cssRGBA(255,220,180,.18)")
		}
	}
	if o.hat != nil {
		switch o.hat.typ {
		case "crown":
			x.hatCrown(o.hat)
		case "tiara":
			x.hatTiara(o.hat)
		case "beret":
			x.hatBeret(o.hat)
		}
	}
	if o.front != nil {
		o.front(x)
	}
	x.Begin()
}

func (x *Ctx) hatCrown(h *hatOpt) {
	g := or(h.color, "#d9b04a")
	x.Poly([][2]float64{{.33, .69}, {.33, .55}, {.39, .61}, {.44, .52}, {.5, .6}, {.56, .52}, {.61, .61}, {.67, .55}, {.67, .69}},
		lin(.33, 0, .67, 0, "#9c7524", g, "#fff0b0", g, "#9c7524"))
	x.FillRect(.33, .655, .34, .035, or(h.band, "#7a1a1a"))
	for i, cx := range []float64{.39, .5, .61} {
		f := "#e5e5e5"
		if i == 1 {
			f = or(h.jewel, "#2f6bd1")
		}
		x.Ell(cx, .672, .014, .012, f)
	}
	for _, p := range [][2]float64{{.33, .55}, {.44, .52}, {.56, .52}, {.67, .55}} {
		x.Ell(p[0], p[1], .012, .012, "#fff3c0")
	}
}

func (x *Ctx) hatTiara(h *hatOpt) {
	c := or(h.color, "#e3c45e")
	x.Begin()
	x.M(.37, .7)
	x.QuadraticTo(.5, .6, .63, .7)
	x.LineTo(.6, .7)
	x.QuadraticTo(.5, .64, .4, .7)
	x.Z()
	x.Fill(c)
	x.Poly([][2]float64{{.47, .64}, {.5, .58}, {.53, .64}}, c)
	x.Ell(.5, .63, .012, .012, or(h.jewel, "#c22"))
}

func (x *Ctx) hatBeret(h *hatOpt) {
	x.Ell(.52, .64, .21, .09, h.color, -.12)
	x.Ell(.52, .62, .17, .07, rad(.47, .6, .01, .2, "cssRGBA(255,255,255,.2)", "cssRGBA(0,0,0,0)"), -.12)
	x.Begin()
	x.M(.62, .6)
	x.QuadraticTo(.8, .4, .9, .3)
	x.QuadraticTo(.82, .46, .66, .62)
	x.Z()
	x.Fill(h.feather)
	for i := 1; i < 9; i++ {
		t := float64(i) / 9
		x.Line(.64+t*.25, .6-t*.29, .7+t*.25, .6-t*.26, "cssRGBA(255,255,255,.5)", .004)
	}
	x.FillRect(.34, .66, .32, .03, h.band)
	x.Ell(.4, .675, .016, .016, "#e8c66a")
}

/* ------------------------------------------------------------- the deck */

var ROYAL = []string{"#1f3a5f", "#7d1c22", "#7d1c22", "#1f3a5f"}

func damask(c1, c2 string) func(*Ctx) {
	return func(x *Ctx) {
		x.FillRect(0, 0, 1, RH, lin(0, 0, 0, RH, c1, c2))
		for i := 0; i < 6; i++ {
			for j := 0; j < 9; j++ {
				cx, cy := (float64(i)+float64(j%2)*.5)/5, float64(j)/5.4
				x.Save()
				x.Translate(cx, cy)
				x.Begin()
				x.M(0, -.06)
				x.QuadraticTo(.05, 0, 0, .06)
				x.QuadraticTo(-.05, 0, 0, -.06)
				x.Fill("cssRGBA(232,200,120,.13)")
				x.Restore()
			}
		}
		x.Begin()
	}
}

func ermine(x *Ctx) {
	x.Begin()
	x.M(.08, RH)
	x.LineTo(.12, 1.2)
	x.QuadraticTo(.3, 1.1, .5, 1.13)
	x.QuadraticTo(.7, 1.1, .88, 1.2)
	x.LineTo(.92, RH)
	x.Z()
	x.Fill("#f6f1e6")
	for i := 0; i < 14; i++ {
		px := .15 + math.Mod(float64(i)*.37, .7)
		py := 1.2 + math.Mod(float64(i)*.23, .38)
		x.Ell(px, py, .012, .02, "#1a1a1a")
	}
}

func courtBg(s int) func(*Ctx) {
	if s%3 != 0 {
		return damask("#5a1a1e", "#2e0c0f")
	}
	return damask("#1d2f4a", "#0f1828")
}

func (x *Ctx) court(s, r int) {
	jewel := func(a, b string) string {
		if s%3 != 0 {
			return a
		}
		return b
	}
	switch r {
	case 13:
		x.portrait(portraitOpt{bg: courtBg(s), coat: []any{ROYAL[s], "#121212"}, hair: "#8a7a6a", beard: "#9c8c78", beardLen: .04,
			mustache: "#8d7d6a", brows: "#6e604f", skin: "#e6b892",
			hat: &hatOpt{typ: "crown", color: "#dcb24c", band: ROYAL[s], jewel: jewel("#2f6bd1", "#c4302b")},
			body: func(x *Ctx) {
				ermine(x)
				x.FillRect(.36, 1.2, .28, .45, ROYAL[s])
				x.Ell(.5, 1.3, .03, .03, "#e8c66a")
			},
			front: func(x *Ctx) {
				x.Line(.82, RH, .86, 1.0, "#c9a24e", .03)
				x.Ell(.86, .98, .04, .04, "#e8c66a")
				x.Ell(.86, .98, .02, .02, "#c4302b")
			}})
	case 12:
		x.portrait(portraitOpt{bg: courtBg(s), coat: []any{ROYAL[s], "#141414"}, hair: "#3b2416", longHair: true, lips: "#b8322e",
			lashes: true, blush: true, skin: "#f0c9a6", earring: "#f3e7c9",
			hat: &hatOpt{typ: "tiara", color: "#e3c45e", jewel: jewel("#c22", "#2f6bd1")},
			collar: func(x *Ctx) {
				for i := 0; i < 9; i++ {
					x.Ell(.4+float64(i)*.025, 1.08+math.Sin(float64(i)/8*PI)*.04, .01, .01, "#f6efe0")
				}
			},
			front: func(x *Ctx) {
				x.Line(.72, RH, .76, 1.18, "#3d6b3a", .012)
				for i := 0; i < 6; i++ {
					a := float64(i) / 6 * TAU
					x.Ell(.76+math.Cos(a)*.03, 1.15+math.Sin(a)*.03, .03, .022, "#b3222a", a)
				}
				x.Ell(.76, 1.15, .022, .022, "#d93a3a")
			}})
	case 11:
		x.portrait(portraitOpt{bg: courtBg(s), coat: []any{ROYAL[s], "#141414"}, hair: "#b07a3a", blush: true, skin: "#efc29c",
			hat: &hatOpt{typ: "beret", color: "#1a1a1a", feather: "#f3ead6", band: ROYAL[s]},
			body: func(x *Ctx) {
				for i := 0; i < 4; i++ {
					x.Ell(.5, 1.24+float64(i)*.1, .018, .018, "#e8c66a")
				}
				x.Line(.3, 1.12, .7, 1.12, "#e8c66a", .02)
			},
			collar: func(x *Ctx) {
				x.Poly([][2]float64{{.36, 1.06}, {.5, 1.16}, {.64, 1.06}, {.62, 1.12}, {.5, 1.2}, {.38, 1.12}}, "#f6efe0")
			}})
	}
}

var (
	pipColors   = &suitColors{red: []any{"#c8262a", "#7e1014"}, black: []any{"#2b2b2b", "#050505"}}
	frameColors = []string{"#3a2a10", "#f6dc8a", "#a8822e", "#f2d37a"}
)

func (x *Ctx) ace(s int, cw, ch float64) {
	cx, cy := cw/2, ch/2
	if s == 0 { // ornate ace of spades
		for i := 0; i < 24; i++ {
			a := float64(i) / 24 * TAU
			f := "#e8c66a"
			if i%2 == 1 {
				f = "#c9a24e"
			}
			x.Ell(cx+math.Cos(a)*cw*.33, cy+math.Sin(a)*cw*.33, cw*.014, cw*.014, f)
		}
		x.Ell(cx, cy, cw*.3, cw*.3, nil)
		x.Stroke("#c9a24e", cw*.006)
		x.drawSuit(0, cx, cy, cw*.42, false, &suitColors{black: []any{"#2b2b2b", "#050505"}, edge: "#c9a24e"})
		t := loraI(cw*.07, 600)
		t.Align = AlignC
		x.Text("The Gold Room", cx, ch*.87, t, "#8a6a32")
	} else {
		x.drawSuit(s, cx, cy, cw*.36, false, pipColors)
	}
}

func (x *Ctx) cardBack(cw, ch float64) {
	x.FillRect(0, 0, cw, ch, lin(0, 0, cw, ch, "#6c1520", "#3e0910"))
	m := cw * .065
	x.Save()
	x.Begin()
	x.Rect(m, m, cw-2*m, ch-2*m)
	x.ClipPath()
	st := cw * .09
	for i := -ch; i < cw+ch; i += st {
		x.Line(i, 0, i+ch, ch, "cssRGBA(232,200,120,.3)", cw*.005)
		x.Line(i, ch, i+ch, 0, "cssRGBA(232,200,120,.3)", cw*.005)
	}
	for i := 0; i < 14; i++ {
		for j := 0; j < 20; j++ {
			x.Ell(float64(i)*st, float64(j)*st*.5+float64(i%2)*st*.25, cw*.008, cw*.008, "cssRGBA(232,200,120,.45)")
		}
	}
	x.Restore()
	x.StrokeRect(m, m, cw-2*m, ch-2*m, "#d9b45c", cw*.016)
	x.StrokeRect(m+cw*.028, m+cw*.028, cw-2*m-cw*.056, ch-2*m-cw*.056, "cssRGBA(217,180,92,.55)", cw*.006)
	cx, cy, r := cw/2, ch/2, cw*.2
	x.Ell(cx, cy, r*1.25, r*1.25, "#4a0c14")
	x.Ell(cx, cy, r*1.25, r*1.25, nil)
	x.Stroke("#d9b45c", cw*.012)
	gold := &suitColors{red: []any{"#f0d27c", "#b8892e"}, black: []any{"#f0d27c", "#b8892e"}}
	for i := 0; i < 4; i++ {
		a := float64(i)*PI/2 - PI/2
		x.drawSuit(i, cx+math.Cos(a)*r*.62, cy+math.Sin(a)*r*.62, r*.55, false, gold)
	}
	x.Ell(cx, cy, r*.2, r*.2, "#e8c66a")
	x.Begin()
}

var PIPS = map[int][][2]float64{
	2: {{.5, .2}, {.5, .8}}, 3: {{.5, .2}, {.5, .5}, {.5, .8}}, 4: {{.3, .2}, {.7, .2}, {.3, .8}, {.7, .8}},
	5:  {{.3, .2}, {.7, .2}, {.5, .5}, {.3, .8}, {.7, .8}},
	6:  {{.3, .2}, {.7, .2}, {.3, .5}, {.7, .5}, {.3, .8}, {.7, .8}},
	7:  {{.3, .2}, {.7, .2}, {.5, .35}, {.3, .5}, {.7, .5}, {.3, .8}, {.7, .8}},
	8:  {{.3, .2}, {.7, .2}, {.5, .35}, {.3, .5}, {.7, .5}, {.5, .65}, {.3, .8}, {.7, .8}},
	9:  {{.3, .2}, {.7, .2}, {.3, .4}, {.7, .4}, {.5, .5}, {.3, .6}, {.7, .6}, {.3, .8}, {.7, .8}},
	10: {{.3, .2}, {.7, .2}, {.5, .3}, {.3, .4}, {.7, .4}, {.3, .6}, {.7, .6}, {.5, .7}, {.3, .8}, {.7, .8}},
}

const (
	CW, CH     = 112.0, 157.0
	cardRadius = .07
)

// cardBody is the paper of a face card: gradient, speckles, vignette.
func (x *Ctx) cardBody(cw, ch float64) {
	x.RR(0, 0, cw, ch, cw*cardRadius)
	x.Save()
	x.ClipPath()
	x.FillRect(0, 0, cw, ch, lin(0, 0, cw*.6, ch, "#fffdf7", "#f1ebdc"))
	R := rng(uint32(cw*977 + ch))
	for i := 0; i < 160; i++ {
		f := cssRGBA(110, 80, 40, R()*.05)
		s := R()*cw*.012 + .3
		px, py := R()*cw, R()*ch
		x.FillRect(px, py, s, s, f)
	}
	x.FillRect(0, 0, cw, ch, rad(cw/2, ch/2, cw*.3, ch*.75, "cssRGBA(0,0,0,0)", "cssRGBA(90,70,30,.1)"))
	x.Restore()
	x.Begin()
}

func (x *Ctx) cardIndex(id int, cw, ch float64) {
	s, r := suitOf(id), rankOf(id)
	fill := "#141414"
	if isRed(id) {
		fill = "#a51e22"
	}
	pass := func() {
		t := lora(cw*.2, 700)
		t.Align = AlignC
		x.Save()
		x.Translate(cw*.118, cw*.235)
		if r == 10 {
			x.Scale(.76, 1)
		}
		x.Text(RANKS[r-1], 0, 0, t, fill)
		x.Restore()
		x.drawSuit(s, cw*.118, cw*.34, cw*.13, false, pipColors)
	}
	pass()
	x.Save()
	x.Translate(cw, ch)
	x.Rotate(PI)
	pass()
	x.Restore()
}

// renderFace draws card id (0..51) at res device pixels per unit.
func renderFace(id int, res float64) *Ctx {
	cw, ch := CW, CH
	x := newCtx(int(math.Ceil(cw*res)), int(math.Ceil(ch*res)))
	x.Scale(res, res)
	x.cardBody(cw, ch)
	s, r := suitOf(id), rankOf(id)
	switch {
	case r >= 11:
		fw := cw * .64
		fh := fw * RH
		fx, fy := (cw-fw)/2, (ch-fh)/2
		x.Save()
		x.Begin()
		x.Rect(fx, fy, fw, fh)
		x.ClipPath()
		x.Translate(fx, fy)
		x.Scale(fw, fw)
		x.court(s, r)
		x.Restore()
		x.StrokeRect(fx, fy, fw, fh, frameColors[0], cw*.014)
		x.StrokeRect(fx+cw*.012, fy+cw*.012, fw-cw*.024, fh-cw*.024, lin(fx, fy, fx+fw, fy+fh, "#f6dc8a", "#a8822e", "#f2d37a"), cw*.009)
		badge := func(bx, by float64) {
			x.Ell(bx, by, fw*.1, fw*.1, "cssRGBA(253,248,236,.93)")
			x.Stroke(frameColors[2], cw*.006)
			x.drawSuit(s, bx, by+fw*.005, fw*.12, false, pipColors)
		}
		badge(fx+fw*.13, fy+fw*.13)
		x.Save()
		x.Translate(cw, ch)
		x.Rotate(PI)
		badge(fx+fw*.13, fy+fw*.13)
		x.Restore()
	case r == 1:
		x.ace(s, cw, ch)
	default:
		sz := cw * .19
		if r >= 9 {
			sz = cw * .17
		}
		for _, p := range PIPS[r] {
			x.drawSuit(s, p[0]*cw, p[1]*ch, sz, p[1] > .5, pipColors)
		}
	}
	x.cardIndex(id, cw, ch)
	x.RR(cw*.006, cw*.006, cw-cw*.012, ch-cw*.012, cw*cardRadius)
	x.Stroke("#8a6a32", cw*.012)
	return x
}

func renderBack(res float64) *Ctx {
	cw, ch := CW, CH
	x := newCtx(int(math.Ceil(cw*res)), int(math.Ceil(ch*res)))
	x.Scale(res, res)
	x.RR(0, 0, cw, ch, cw*cardRadius)
	x.Save()
	x.ClipPath()
	x.cardBack(cw, ch)
	x.Restore()
	x.RR(cw*.006, cw*.006, cw-cw*.012, ch-cw*.012, cw*cardRadius)
	x.Stroke("#1c0606", cw*.012)
	return x
}

/* ---------------------------------------------------------------- chips */

var chipCols = map[int][3]string{
	5: {"#b3221d", "#f4ead8", "#ffffff"}, 25: {"#1e7a46", "#f4ead8", "#ffffff"}, 100: {"#1b1b1b", "#e5d3a5", "#e5d3a5"},
	500: {"#5a2a86", "#f4ead8", "#ffffff"}, 1000: {"#d6a536", "#5a3a10", "#3e2808"},
}

func lighten(hex string) string {
	c := cssCol(hex)
	f := func(v uint8) int { return int(math.Min(255, math.Round(float64(v)+(255-float64(v))*.22))) }
	return fmt.Sprintf("rgb(%d,%d,%d)", f(c.R), f(c.G), f(c.B))
}

// drawChip draws a casino chip of diameter size with its top-left at (0,0).
func (x *Ctx) drawChip(v int, size float64) {
	r := size / 2
	cc := chipCols[v]
	base, ins, txt := cc[0], cc[1], cc[2]
	x.Ell(r, r, r-1, r-1, base)
	x.Save()
	x.Begin()
	x.Arc(r, r, r-1, 0, TAU, false)
	x.ClipPath()
	for i := 0; i < 6; i++ {
		a := float64(i) / 6 * TAU
		x.Save()
		x.Translate(r, r)
		x.Rotate(a)
		x.FillRect(-r*.13, -r, r*.26, r*.3, ins)
		x.Restore()
	}
	x.Restore()
	x.Ell(r, r, r*.68, r*.68, nil)
	x.Dash(r*.09, r*.07)
	x.Stroke(ins, r*.045)
	x.Dash()
	x.Ell(r, r, r*.6, r*.6, rad(r*.85, r*.8, r*.05, r*.6, lighten(base), base))
	label, fs := fmt.Sprint(v), r*.5
	if v >= 1000 {
		label, fs = "1K", r*.42
	}
	t := lora(fs, 700)
	t.Align, t.Base = AlignC, BaseMid
	x.Text(label, r, r+r*.03, t, txt)
	x.Ell(r, r, r-1.5, r-1.5, nil)
	x.Stroke("cssRGBA(0,0,0,.45)", 1.5)
	x.Ell(r, r, r-1.5, r-1.5, rad(r*.7, r*.55, 0, r, "cssRGBA(255,255,255,.22)", "cssRGBA(255,255,255,0)", "cssRGBA(0,0,0,.18)"))
	x.Begin()
}

/* ---------------------------------------------------------------- table */

var RAIL = struct{ cx, cy, r float64 }{800, -1180, 1880}

func (x *Ctx) arcText(text string, cx, cy, r, mid float64, t TS, color string, spacing float64) {
	t.Align, t.Base = AlignC, BaseMid
	var ws []float64
	total := 0.0
	rs := []rune(text)
	for _, ch := range rs {
		w := t.Width(string(ch)) * spacing
		ws = append(ws, w)
		total += w
	}
	a := mid + total/r/2
	for i, ch := range rs {
		w := ws[i]
		a -= w / 2 / r
		x.Save()
		x.Translate(cx+math.Cos(a)*r, cy+math.Sin(a)*r)
		x.Rotate(a - PI/2)
		x.Text(string(ch), 0, 0, t, color)
		x.Restore()
		a -= w / 2 / r
	}
}

// renderScene draws the whole table (W x H user units) at res.
func renderScene(W, H, res float64) *Ctx {
	x := newCtx(int(math.Ceil(W*res)), int(math.Ceil(H*res)))
	x.Scale(res, res)
	x.FillRect(0, 0, W, H, lin(0, 0, 0, H, "#120c09", "#1c120c"))
	// felt
	x.Save()
	x.Begin()
	x.Arc(RAIL.cx, RAIL.cy, RAIL.r, 0, PI, false)
	x.LineTo(0, 0)
	x.Z()
	x.ClipPath()
	x.FillRect(0, 0, W, H, rad(800, 330, 60, 1000, "#1f6a49", "#17573b", "#0f3f2a", "#0a2d1e"))
	R := rng(77)
	x.napSpecks(R, W, H)
	x.Begin()
	x.Arc(RAIL.cx, RAIL.cy, RAIL.r-28, .1, PI-.1, false)
	x.Stroke("cssRGBA(226,190,110,.55)", 2.5)
	x.Begin()
	x.Arc(RAIL.cx, RAIL.cy, RAIL.r-36, .1, PI-.1, false)
	x.Stroke("cssRGBA(226,190,110,.3)", 1)
	x.arcText("BLACKJACK PAYS 3 TO 2", 800, -492, 860, PI/2, lora(30, 600), "cssRGBA(232,200,120,.82)", 1.25)
	x.arcText("Dealer must stand on all 17s and draw to 16", 800, -492, 824, PI/2, loraI(19, 400), "cssRGBA(232,200,120,.6)", 1.05)
	mono := loraI(22, 600)
	mono.Align = AlignC
	x.Text("The Gold Room", 800, 112, mono, "cssRGBA(232,200,120,.28)")
	x.ring(800, 646, 54)
	x.FillRect(0, 0, W, H, rad(800, 360, 40, 900, "cssRGBA(255,230,170,.16)", "cssRGBA(255,230,170,.05)", "cssRGBA(0,0,0,.35)"))
	x.Restore()
	// padded leather rail
	x.Save()
	x.Begin()
	x.Arc(RAIL.cx, RAIL.cy, RAIL.r+46, 0, PI, false)
	x.Arc(RAIL.cx, RAIL.cy, RAIL.r, PI, 0, true)
	x.Z()
	x.Fill(rad(RAIL.cx, RAIL.cy, RAIL.r, RAIL.r+46, "#2a140c", "#5a2c18", "#6e3a20", "#3a1a0e", "#1a0a05"))
	x.ClipPath()
	for a := .2; a < PI-.2; a += .006 {
		px, py := RAIL.cx+math.Cos(a)*(RAIL.r+23), RAIL.cy+math.Sin(a)*(RAIL.r+23)
		x.Ell(px, py, 1.2, 1.2, "cssRGBA(255,200,150,.05)")
	}
	x.Begin()
	x.Arc(RAIL.cx, RAIL.cy, RAIL.r+16, 0, PI, false)
	x.Stroke("cssRGBA(255,220,180,.12)", 3)
	x.Restore()
	x.Begin()
	x.Arc(RAIL.cx, RAIL.cy, RAIL.r+1, .05, PI-.05, false)
	x.Stroke("#c9a24e", 2)
	x.Begin()
	x.Arc(RAIL.cx, RAIL.cy, RAIL.r+47, .05, PI-.05, false)
	x.Stroke("cssRGBA(201,162,78,.6)", 2)
	// wood ledge below the rail
	x.Begin()
	x.Arc(RAIL.cx, RAIL.cy, RAIL.r+48, 0, PI, false)
	x.LineTo(0, H)
	x.LineTo(W, H)
	x.Z()
	x.Fill(lin(0, 700, 0, H, "#24140c", "#140a06"))
	x.Begin()
	x.tray(560, 4, 480, 70)
	x.shoe(1330, 34)
	x.discard(150, 34)
	return x
}

// napSpecks scatters 9000 tiny felt fibres. They are 1.2 unit squares, so
// they are blended straight into the pixels (inside the felt clip) rather
// than rasterized one path at a time.
func (x *Ctx) napSpecks(R func() float64, W, H float64) {
	img := x.RGBA()
	s := x.S()
	sz := 1.2 * s
	inFelt := func(px, py float64) bool { // user coords
		dx, dy := px-RAIL.cx, py-RAIL.cy
		return dy >= RAIL.cy-RAIL.cy && dx*dx+dy*dy <= RAIL.r*RAIL.r && py >= 0
	}
	b := img.Bounds()
	for i := 0; i < 9000; i++ {
		var c [4]float64
		if R() < .5 {
			c = [4]float64{0, 0, 0, R() * .07}
		} else {
			c = [4]float64{140, 220, 170, R() * .035}
		}
		ux, uy := R()*W, R()*H
		if !inFelt(ux, uy) {
			continue
		}
		x0, y0 := ux*s, uy*s
		for py := int(math.Floor(y0)); py < int(math.Ceil(y0+sz)); py++ {
			for px := int(math.Floor(x0)); px < int(math.Ceil(x0+sz)); px++ {
				if px < b.Min.X || py < b.Min.Y || px >= b.Max.X || py >= b.Max.Y {
					continue
				}
				cov := (math.Min(float64(px+1), x0+sz) - math.Max(float64(px), x0)) * (math.Min(float64(py+1), y0+sz) - math.Max(float64(py), y0))
				if cov <= 0 {
					continue
				}
				a := c[3] * cov
				o := img.PixOffset(px, py)
				for k := 0; k < 3; k++ {
					img.Pix[o+k] = uint8(math.Round(c[k]*a + float64(img.Pix[o+k])*(1-a)))
				}
			}
		}
	}
}

func (x *Ctx) ring(cx, cy, r float64) {
	x.Ell(cx, cy, r, r, "cssRGBA(0,0,0,.12)")
	x.Ell(cx, cy, r, r, nil)
	x.Stroke("cssRGBA(232,200,120,.75)", 3)
	x.Ell(cx, cy, r-7, r-7, nil)
	x.Stroke("cssRGBA(232,200,120,.35)", 1.2)
	x.Begin()
}

func (x *Ctx) tray(X, Y, w, h float64) {
	x.RR(X-6, Y-6, w+12, h+12, 10)
	x.Fill("#1b0e08")
	x.RR(X, Y, w, h, 7)
	x.Fill(lin(0, Y, 0, Y+h, "#3a2416", "#120a06"))
	cols := [][2]string{{"#b8241f", "#f2e7d6"}, {"#1f7a46", "#f2e7d6"}, {"#1d1d1d", "#e8d8b0"}, {"#5b2d86", "#f2e7d6"}, {"#d9a93a", "#5a3a10"}}
	slots := 10
	sw := w / float64(slots)
	for i := 0; i < slots; i++ {
		c1, c2 := cols[i%5][0], cols[i%5][1]
		fi := float64(i)
		x.Save()
		x.Begin()
		x.Rect(X+fi*sw+3, Y+4, sw-6, h-8)
		x.ClipPath()
		for k := 0; k < 9; k++ {
			yy := Y + 6 + float64(k)*7
			x.RR(X+fi*sw+5, yy, sw-10, 6, 2)
			x.Fill(c1)
			for s := 0; s < 3; s++ {
				x.FillRect(X+fi*sw+8+float64(s)*(sw-16)/2.5, yy+1, 4, 4, c2)
			}
		}
		x.Restore()
		x.Line(X+fi*sw, Y+3, X+fi*sw, Y+h-3, "cssRGBA(0,0,0,.6)", 2)
	}
	x.FillRect(X, Y, w, 20, lin(0, Y, 0, Y+20, "cssRGBA(255,255,255,.12)", "cssRGBA(255,255,255,0)"))
}

func (x *Ctx) shoe(X, Y float64) {
	x.Save()
	x.Translate(X, Y)
	x.Rotate(-.06)
	x.RR(0, 0, 170, 120, 10)
	x.Fill(lin(0, 0, 0, 120, "#2d2a28", "#0e0d0c"))
	x.RR(8, 8, 154, 104, 7)
	x.Fill(lin(0, 0, 170, 0, "#3b2414", "#5c3a20", "#3b2414"))
	x.RR(14, 20, 90, 80, 5)
	x.Fill("#5e1717")
	for i := 0; i < 6; i++ {
		fi := float64(i)
		x.Line(18+fi*14, 24, 18+fi*14, 96, "cssRGBA(255,215,140,.25)", 1)
	}
	x.RR(110, 30, 46, 60, 6)
	x.Fill("#111")
	x.Stroke("#c9a24e", 2)
	t := lora(11, 600)
	t.Align = AlignC
	x.Text("SHOE", 85, 115, t, "#c9a24e")
	x.Restore()
	x.Begin()
}

func (x *Ctx) discard(X, Y float64) {
	x.Save()
	x.Translate(X, Y)
	x.Rotate(.05)
	x.RR(0, 0, 140, 150, 9)
	x.Fill("cssRGBA(10,8,6,.55)")
	x.Stroke("cssRGBA(201,162,78,.5)", 2)
	t := lora(11, 600)
	t.Align = AlignC
	x.Text("DISCARD", 70, 166, t, "cssRGBA(201,162,78,.5)")
	x.Restore()
	x.Begin()
}
