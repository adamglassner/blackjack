// Fonts: the bundled TrueType faces are drawn as vector outlines on the
// canvas (see Ctx.Text), so text stays sharp at any scale and can be filled
// with gradients or emboldened like a CSS font-weight.
// (c) 2026 Ironite
package main

import (
	"embed"
	"math"
	"strings"
	"sync"
	"unicode/utf8"

	"golang.org/x/image/font"
	"golang.org/x/image/font/sfnt"
	"golang.org/x/image/math/fixed"
)

//go:embed assets/fonts/*.ttf
var assetFS embed.FS

type Font struct {
	f              *sfnt.Font
	upem           float64
	asc, desc, gap float64 // in em
	mu             sync.Mutex
	buf            sfnt.Buffer
	glyphs         map[rune]*glyph
}

type glyph struct {
	segs sfnt.Segments
	adv  float64 // em
	ok   bool
}

func loadFont(name string) *Font {
	b, err := assetFS.ReadFile("assets/fonts/" + name)
	if err != nil {
		panic(err)
	}
	f, err := sfnt.Parse(b)
	if err != nil {
		panic(err)
	}
	fo := &Font{f: f, upem: float64(f.UnitsPerEm()), glyphs: map[rune]*glyph{}}
	m, err := f.Metrics(&fo.buf, fixed.I(int(f.UnitsPerEm())), font.HintingNone)
	if err == nil {
		fo.asc = float64(m.Ascent) / 64 / fo.upem
		fo.desc = float64(m.Descent) / 64 / fo.upem
		fo.gap = float64(m.Height)/64/fo.upem - fo.asc - fo.desc
		if fo.gap < 0 {
			fo.gap = 0
		}
	}
	return fo
}

var (
	fItaliana = loadFont("Italiana-Regular.ttf")
	fLora     = loadFont("Lora-Variable.ttf")
	fLoraI    = loadFont("Lora-Italic-Variable.ttf")
	fOutfit   = loadFont("Outfit-Regular.ttf")
	fOutfitB  = loadFont("Outfit-Bold.ttf")
)

func (fo *Font) glyph(r rune) *glyph {
	fo.mu.Lock()
	defer fo.mu.Unlock()
	if g, ok := fo.glyphs[r]; ok {
		return g
	}
	g := &glyph{}
	ppem := fixed.I(int(fo.upem))
	if idx, err := fo.f.GlyphIndex(&fo.buf, r); err == nil && idx != 0 {
		if segs, err := fo.f.LoadGlyph(&fo.buf, idx, ppem, nil); err == nil {
			g.segs = append(sfnt.Segments(nil), segs...)
			g.ok = true
		}
		if a, err := fo.f.GlyphAdvance(&fo.buf, idx, ppem, font.HintingNone); err == nil {
			g.adv = float64(a) / 64 / fo.upem
		}
	}
	fo.glyphs[r] = g
	return g
}

// lineHeight is CSS "line-height: normal" in em.
func (fo *Font) lineHeight() float64 { return fo.asc + fo.desc + fo.gap }

// TS is a text style. Size is in user units; Bold adds a stroke of
// Bold*Size around each glyph (a stand-in for heavier weights of the
// variable fonts); Track is CSS letter-spacing in user units.
type TS struct {
	F      *Font
	Size   float64
	Bold   float64
	Track  float64
	Align  int // 0 left, 1 center, 2 right
	Base   int // 0 alphabetic, 1 middle, 2 top (em box top)
	Upper  bool
	ScaleX float64
}

const (
	AlignL = iota
	AlignC
	AlignR
)
const (
	BaseAlpha = iota
	BaseMid
	BaseTop
)

// CSS-like named styles.
func lora(size float64, weight int) TS {
	return TS{F: fLora, Size: size, Bold: emb(weight)}
}
func loraI(size float64, weight int) TS {
	return TS{F: fLoraI, Size: size, Bold: emb(weight)}
}
func outfit(size float64) TS     { return TS{F: fOutfit, Size: size} }
func outfitB(size float64) TS    { return TS{F: fOutfitB, Size: size} }
func outfitSemi(size float64) TS { return TS{F: fOutfitB, Size: size} }
func italiana(size float64) TS   { return TS{F: fItaliana, Size: size} }

func emb(weight int) float64 {
	if weight <= 400 {
		return 0
	}
	return float64(weight-400) / 300 * .05
}

func (t TS) str(s string) string {
	if t.Upper {
		return strings.ToUpper(s)
	}
	return s
}

func (t TS) sx() float64 {
	if t.ScaleX == 0 {
		return 1
	}
	return t.ScaleX
}

// Width is the advance width of s (CSS: includes letter-spacing after
// every character).
func (t TS) Width(s string) float64 {
	s = t.str(s)
	w := 0.0
	for _, r := range s {
		w += t.adv(r) + t.Track
	}
	return w * t.sx()
}

func (t TS) adv(r rune) float64 {
	if fb, ok := fallbacks[r]; ok {
		if g := t.F.glyph(r); !g.ok {
			return fb.adv * t.Size
		}
	}
	return t.F.glyph(r).adv * t.Size
}

func (t TS) Asc() float64  { return t.F.asc * t.Size }
func (t TS) Desc() float64 { return t.F.desc * t.Size }
func (t TS) LH() float64   { return t.F.lineHeight() * t.Size }

// fallback glyphs the bundled fonts lack, drawn in em units around the
// baseline (y up is negative).
type fbGlyph struct {
	adv  float64
	draw func(c *Ctx, x, y, size float64)
}

var fallbacks = map[rune]fbGlyph{
	'♫': {.62, func(c *Ctx, x, y, s float64) {
		c.Begin()
		c.EllArc(x+.16*s, y-.06*s, .11*s, .08*s, -.35, 0, 2*math.Pi, false)
		c.Z()
		c.M(x+.5*s, y-.12*s)
		c.EllArc(x+.44*s, y-.12*s, .11*s, .08*s, -.35, 0, 2*math.Pi, false)
		c.Z()
		c.M(x+.25*s, y-.07*s)
		c.L(x+.25*s, y-.72*s)
		c.L(x+.55*s, y-.8*s)
		c.L(x+.55*s, y-.13*s)
		c.L(x+.5*s, y-.13*s)
		c.L(x+.5*s, y-.66*s)
		c.L(x+.3*s, y-.6*s)
		c.L(x+.3*s, y-.07*s)
		c.Z()
	}},
	'▼': {.78, func(c *Ctx, x, y, s float64) {
		c.Begin()
		c.M(x+.06*s, y-.68*s)
		c.L(x+.72*s, y-.68*s)
		c.L(x+.39*s, y-.05*s)
		c.Z()
	}},
}

// textPath adds the outline of s to the current path, starting at pen
// position (x, y) on the baseline. Fallback glyphs are drawn by fill
// callback because they build their own paths.
func (c *Ctx) textRun(s string, x, y float64, t TS, paint any) {
	k := t.Size / 1 // em → user units
	sx := t.sx()
	pen := 0.0
	c.Begin()
	var deferred []func()
	for _, r := range s {
		g := t.F.glyph(r)
		if !g.ok {
			if fb, ok := fallbacks[r]; ok {
				px := x + pen*sx
				deferred = append(deferred, func() { fb.draw(c, px, y, t.Size); c.fillOrBold(paint, t) })
				pen += fb.adv*t.Size + t.Track
				continue
			}
		}
		ox := x + pen*sx
		u := k / t.F.upem
		for _, sg := range g.segs {
			a := sg.Args
			px := func(i int) float64 { return ox + float64(a[i].X)/64*u*sx }
			py := func(i int) float64 { return y + float64(a[i].Y)/64*u }
			switch sg.Op {
			case sfnt.SegmentOpMoveTo:
				c.ClosePath()
				c.M(px(0), py(0))
			case sfnt.SegmentOpLineTo:
				c.LineTo(px(0), py(0))
			case sfnt.SegmentOpQuadTo:
				c.QuadraticTo(px(0), py(0), px(1), py(1))
			case sfnt.SegmentOpCubeTo:
				c.CubicTo(px(0), py(0), px(1), py(1), px(2), py(2))
			}
		}
		c.ClosePath()
		pen += g.adv*t.Size + t.Track
	}
	c.fillOrBold(paint, t)
	for _, d := range deferred {
		d()
	}
	c.Begin()
}

func (c *Ctx) fillOrBold(paint any, t TS) {
	c.Fill(paint)
	if t.Bold > 0 {
		c.SetLineJoinRound()
		c.Stroke(paint, t.Bold*t.Size)
	}
}

// Text draws s at (x, y) with canvas-style align and baseline.
func (c *Ctx) Text(s string, x, y float64, t TS, paint any) {
	s = t.str(s)
	if s == "" {
		return
	}
	w := t.Width(s)
	switch t.Align {
	case AlignC:
		x -= w / 2
	case AlignR:
		x -= w
	}
	switch t.Base {
	case BaseMid:
		y += (t.Asc() - t.Desc()) / 2
	case BaseTop:
		y += t.Asc()
	}
	c.textRun(s, x, y, t, paint)
}

func runeCount(s string) int { return utf8.RuneCountInString(s) }
