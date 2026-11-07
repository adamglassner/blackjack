// Drawing helpers. Everything is laid out on the original 1600x900 stage in
// "logical" pixels; the canvas maps those to real pixels so the game stays
// sharp at any window size or display scale.
// (c) 2026 Ironite
package main

import (
	"image"
	"image/color"
	"math"
	"strings"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/colorm"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const stageW, stageH = 1600.0, 900.0

// canvas maps logical stage coordinates to pixels on dst.
type canvas struct {
	dst    *ebiten.Image
	k      float64 // pixels per logical pixel
	ox, oy float64 // pixel offset of the logical origin
}

var cv canvas

func X(x float64) float32  { return float32(cv.ox + x*cv.k) }
func Y(y float64) float32  { return float32(cv.oy + y*cv.k) }
func Sc(v float64) float32 { return float32(v * cv.k) }

// shift moves the logical origin (e.g. into table space) and returns a func
// that undoes it.
func shift(dx, dy float64) func() {
	old := cv
	cv.ox += dx * cv.k
	cv.oy += dy * cv.k
	return func() { cv = old }
}

func (c col) scale() ebiten.ColorScale {
	var s ebiten.ColorScale
	s.SetR(c.r * c.a)
	s.SetG(c.g * c.a)
	s.SetB(c.b * c.a)
	s.SetA(c.a)
	return s
}

/* ----------------------------------------------------------------- shapes */

var (
	whiteImg *ebiten.Image
	whiteSub *ebiten.Image
	vbuf     []ebiten.Vertex
	ibuf     []uint16
)

func initGfx() {
	whiteImg = ebiten.NewImage(3, 3)
	whiteImg.Fill(color.White)
	whiteSub = whiteImg.SubImage(image.Rect(1, 1, 2, 2)).(*ebiten.Image)
}

// rrPath is a rounded rectangle in logical coordinates.
func rrPath(x, y, w, h, r float64) *vector.Path {
	var p vector.Path
	x0, y0, x1, y1 := X(x), Y(y), X(x+w), Y(y+h)
	R := min(Sc(r), (x1-x0)/2, (y1-y0)/2)
	if R <= 0 {
		p.MoveTo(x0, y0)
		p.LineTo(x1, y0)
		p.LineTo(x1, y1)
		p.LineTo(x0, y1)
		p.Close()
		return &p
	}
	p.MoveTo(x0+R, y0)
	p.LineTo(x1-R, y0)
	p.Arc(x1-R, y0+R, R, -math.Pi/2, 0, vector.Clockwise)
	p.LineTo(x1, y1-R)
	p.Arc(x1-R, y1-R, R, 0, math.Pi/2, vector.Clockwise)
	p.LineTo(x0+R, y1)
	p.Arc(x0+R, y1-R, R, math.Pi/2, math.Pi, vector.Clockwise)
	p.LineTo(x0, y0+R)
	p.Arc(x0+R, y0+R, R, math.Pi, math.Pi*3/2, vector.Clockwise)
	p.Close()
	return &p
}

func drawVerts(vs []ebiten.Vertex, is []uint16, colAt func(y float32) col) {
	for i := range vs {
		c := colAt(vs[i].DstY)
		vs[i].SrcX, vs[i].SrcY = 1.5, 1.5
		vs[i].ColorR, vs[i].ColorG, vs[i].ColorB, vs[i].ColorA = c.r*c.a, c.g*c.a, c.b*c.a, c.a
	}
	cv.dst.DrawTriangles(vs, is, whiteSub, &ebiten.DrawTrianglesOptions{FillRule: ebiten.FillRuleFillAll, ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha})
}

// fillPathV fills p (small shapes: triangles, icons, confetti) with a
// vertical gradient from c0 at logical y0 to c1 at y1.
func fillPathV(p *vector.Path, y0, y1 float64, c0, c1 col) {
	vbuf, ibuf = p.AppendVerticesAndIndicesForFilling(vbuf[:0], ibuf[:0])
	py0, py1 := Y(y0), Y(y1)
	drawVerts(vbuf, ibuf, func(y float32) col {
		if py1 == py0 {
			return c0
		}
		return c0.lerp(c1, min(1, max(0, (y-py0)/(py1-py0))))
	})
}

/* Rounded boxes and circles are drawn as nine pieces: four corners cut from a
   small anti-aliased disc (or ring) image made once per radius, and plain
   strips in between. One cheap draw call per shape, crisp at any scale. */

type discKey struct{ r, bw int } // in quarter pixels

var discs = map[discKey]*ebiten.Image{}

const discPad = 4

// disc returns a white disc (or ring of width bw) of radius r pixels. Pixels
// (0,0)-(2,2) are solid white, for the strips.
func disc(r, bw float64) (*ebiten.Image, float64) {
	k := discKey{int(r*4 + .5), int(bw*4 + .5)}
	R := float64(k.r) / 4
	B := float64(k.bw) / 4
	c := math.Ceil(R) + discPad // centre
	if img, ok := discs[k]; ok {
		return img, c
	}
	n := int(2 * c)
	pix := make([]byte, n*n*4)
	for y := 0; y < n; y++ {
		for x := 0; x < n; x++ {
			d := math.Hypot(float64(x)+.5-c, float64(y)+.5-c)
			a := clamp(R-d+.5, 0, 1)
			if B > 0 && R-B > 0 {
				a -= clamp(R-B-d+.5, 0, 1)
			}
			if x < 3 && y < 3 {
				a = 1
			}
			b := byte(a*255 + .5)
			i := (y*n + x) * 4
			pix[i], pix[i+1], pix[i+2], pix[i+3] = b, b, b, b
		}
	}
	img := ebiten.NewImage(n, n)
	img.WritePixels(pix)
	discs[k] = img
	return img, c
}

// rrDraw draws a rounded box (logical coords) with a vertical gradient;
// bw > 0 draws only a border of that width, inside the box.
func rrDraw(x, y, w, h, r, bw float64, c0, c1 col) {
	X0, Y0, X1, Y1 := float64(X(x)), float64(Y(y)), float64(X(x+w)), float64(Y(y+h))
	if X1-X0 <= 0 || Y1-Y0 <= 0 {
		return
	}
	R := min(r*cv.k, (X1-X0)/2, (Y1-Y0)/2)
	B := bw * cv.k
	if B > 0 && R > 0 {
		R = max(R, B)
	}
	img, c := disc(R, B)
	R = float64(int(R*4+.5)) / 4
	vs := vbuf[:0]
	is := ibuf[:0]
	colAt := func(py float64) col {
		if Y1 == Y0 {
			return c0
		}
		return c0.lerp(c1, float32(clamp((py-Y0)/(Y1-Y0), 0, 1)))
	}
	quad := func(dx0, dy0, dx1, dy1, sx0, sy0, sx1, sy1 float64) {
		if dx1-dx0 <= 0 || dy1-dy0 <= 0 {
			return
		}
		base := uint16(len(vs))
		for _, v := range [4][4]float64{{dx0, dy0, sx0, sy0}, {dx1, dy0, sx1, sy0}, {dx0, dy1, sx0, sy1}, {dx1, dy1, sx1, sy1}} {
			cc := colAt(v[1])
			vs = append(vs, ebiten.Vertex{DstX: float32(v[0]), DstY: float32(v[1]), SrcX: float32(v[2]), SrcY: float32(v[3]),
				ColorR: cc.r * cc.a, ColorG: cc.g * cc.a, ColorB: cc.b * cc.a, ColorA: cc.a})
		}
		is = append(is, base, base+1, base+2, base+1, base+3, base+2)
	}
	solid := func(dx0, dy0, dx1, dy1 float64) { quad(dx0, dy0, dx1, dy1, 1, 1, 2, 2) }
	if R > 0 {
		quad(X0, Y0, X0+R, Y0+R, c-R, c-R, c, c)
		quad(X1-R, Y0, X1, Y0+R, c, c-R, c+R, c)
		quad(X0, Y1-R, X0+R, Y1, c-R, c, c, c+R)
		quad(X1-R, Y1-R, X1, Y1, c, c, c+R, c+R)
	}
	if B <= 0 {
		solid(X0+R, Y0, X1-R, Y1)
		solid(X0, Y0+R, X0+R, Y1-R)
		solid(X1-R, Y0+R, X1, Y1-R)
	} else if R > 0 {
		solid(X0+R, Y0, X1-R, Y0+B)
		solid(X0+R, Y1-B, X1-R, Y1)
		solid(X0, Y0+R, X0+B, Y1-R)
		solid(X1-B, Y0+R, X1, Y1-R)
	} else {
		solid(X0, Y0, X1, Y0+B)
		solid(X0, Y1-B, X1, Y1)
		solid(X0, Y0+B, X0+B, Y1-B)
		solid(X1-B, Y0+B, X1, Y1-B)
	}
	vbuf, ibuf = vs, is
	cv.dst.DrawTriangles(vs, is, img, &ebiten.DrawTrianglesOptions{Filter: ebiten.FilterLinear, ColorScaleMode: ebiten.ColorScaleModePremultipliedAlpha})
}

func fillRR(x, y, w, h, r float64, c col)       { rrDraw(x, y, w, h, r, 0, c, c) }
func fillRRV(x, y, w, h, r float64, c0, c1 col) { rrDraw(x, y, w, h, r, 0, c0, c1) }

// borderRR draws a CSS-style border of width bw inside the box.
func borderRR(x, y, w, h, r, bw float64, c col) { rrDraw(x, y, w, h, r, bw, c, c) }

// boxRR draws a box with a vertical gradient background and a border.
func boxRR(x, y, w, h, r, bw float64, c0, c1, border col) {
	fillRRV(x, y, w, h, r, c0, c1)
	if bw > 0 {
		borderRR(x, y, w, h, r, bw, border)
	}
}

// insetHi draws the 1px inner top highlight of `inset 0 1px 0 c`.
func insetHi(x, y, w, h, r float64, c col) {
	fillRect(x+r*.6, y+1, w-r*1.2, 1, c)
}

func fillCircle(cx, cy, r float64, c col) { rrDraw(cx-r, cy-r, 2*r, 2*r, r, 0, c, c) }

// strokeCircle strokes a circle of radius r with a line w wide (centred on r).
func strokeCircle(cx, cy, r, w float64, c col) {
	ro := r + w/2
	rrDraw(cx-ro, cy-ro, 2*ro, 2*ro, ro, w, c, c)
}
func fillRect(x, y, w, h float64, c col)       { rrDraw(x, y, w, h, 0, 0, c, c) }
func fillRectV(x, y, w, h float64, c0, c1 col) { rrDraw(x, y, w, h, 0, 0, c0, c1) }

func strokeLine(x0, y0, x1, y1, w float64, c col) {
	vector.StrokeLine(cv.dst, X(x0), Y(y0), X(x1), Y(y1), Sc(w), c.std(), false)
}

func fillTri(pts [3][2]float64, c col) {
	var p vector.Path
	p.MoveTo(X(pts[0][0]), Y(pts[0][1]))
	p.LineTo(X(pts[1][0]), Y(pts[1][1]))
	p.LineTo(X(pts[2][0]), Y(pts[2][1]))
	p.Close()
	fillPathV(&p, 0, 1, c, c)
}

/* --------------------------------------------------- soft shadows & glows */

type maskKey struct{ w, h, r, blur, spread int }

var masks = map[maskKey]*ebiten.Image{}

// softMask is a white rounded rectangle (grown by spread) blurred like a CSS
// box-shadow with the given blur radius, at 1 mask pixel per logical pixel.
// The image is padded by blur on every side.
func softMask(w, h, r, blur, spread float64) *ebiten.Image {
	k := maskKey{int(w + .5), int(h + .5), int(r + .5), int(blur + .5), int(spread + .5)}
	if m, ok := masks[k]; ok {
		return m
	}
	pad := k.blur
	W, H := k.w+2*k.spread+2*pad, k.h+2*k.spread+2*pad
	a := make([]float32, W*H)
	rr := float64(k.r + k.spread)
	if k.r == 0 {
		rr = 0
	}
	bw, bh := float64(k.w+2*k.spread), float64(k.h+2*k.spread)
	for y := 0; y < H; y++ {
		for x := 0; x < W; x++ {
			// signed distance to the rounded box, for a smooth edge
			px, py := float64(x-pad)+.5-bw/2, float64(y-pad)+.5-bh/2
			qx, qy := math.Abs(px)-(bw/2-rr), math.Abs(py)-(bh/2-rr)
			d := math.Hypot(max(qx, 0), max(qy, 0)) + min(max(qx, qy), 0) - rr
			a[y*W+x] = float32(min(1, max(0, .5-d)))
		}
	}
	if k.blur > 0 {
		// three box blurs approximate a gaussian with sigma = blur/2
		sigma := float64(k.blur) / 2
		box := int(math.Sqrt(12*sigma*sigma/3+1) + .5)
		if box%2 == 0 {
			box++
		}
		tmp := make([]float32, len(a))
		for pass := 0; pass < 3; pass++ {
			boxBlur(a, tmp, W, H, box/2, true)
			boxBlur(tmp, a, W, H, box/2, false)
		}
	}
	pix := make([]byte, W*H*4)
	for i, v := range a {
		b := byte(min(255, max(0, v*255+.5)))
		pix[i*4], pix[i*4+1], pix[i*4+2], pix[i*4+3] = b, b, b, b
	}
	img := ebiten.NewImage(W, H)
	img.WritePixels(pix)
	masks[k] = img
	return img
}

func boxBlur(src, dst []float32, w, h, r int, horiz bool) {
	n := float32(2*r + 1)
	if horiz {
		for y := 0; y < h; y++ {
			row := src[y*w : (y+1)*w]
			var s float32
			for i := -r; i <= r; i++ {
				s += row[min(w-1, max(0, i))]
			}
			for x := 0; x < w; x++ {
				dst[y*w+x] = s / n
				s += row[min(w-1, x+r+1)] - row[max(0, x-r)]
			}
		}
		return
	}
	for x := 0; x < w; x++ {
		var s float32
		for i := -r; i <= r; i++ {
			s += src[min(h-1, max(0, i))*w+x]
		}
		for y := 0; y < h; y++ {
			dst[y*w+x] = s / n
			s += src[min(h-1, y+r+1)*w+x] - src[max(0, y-r)*w+x]
		}
	}
}

// shadow draws a CSS box-shadow: offset (dx, dy), blur radius, spread.
func shadow(x, y, w, h, r, dx, dy, blur, spread float64, c col) {
	if c.a <= 0 {
		return
	}
	m := softMask(w, h, r, blur, spread)
	pad := float64(int(blur+.5)) + float64(int(spread+.5))
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(cv.k, cv.k)
	op.GeoM.Translate(float64(X(x+dx-pad)), float64(Y(y+dy-pad)))
	op.ColorScale = c.scale()
	cv.dst.DrawImage(m, op)
}

/* ----------------------------------------------------------------- images */

// drawImg draws img to fill the logical box (x, y, w, h).
func drawImg(img *ebiten.Image, x, y, w, h float64, cs ebiten.ColorScale) {
	b := img.Bounds()
	op := &ebiten.DrawImageOptions{Filter: ebiten.FilterLinear}
	op.GeoM.Scale(w*cv.k/float64(b.Dx()), h*cv.k/float64(b.Dy()))
	op.GeoM.Translate(float64(X(x)), float64(Y(y)))
	op.ColorScale = cs
	cv.dst.DrawImage(img, op)
}

func plain() ebiten.ColorScale { return ebiten.ColorScale{} }
func alphaCS(a float32) ebiten.ColorScale {
	var s ebiten.ColorScale
	s.ScaleAlpha(a)
	return s
}

// circleImg draws img clipped... the avatars are already round, so this just
// adds the CSS border ring around them.
func avatar(img *ebiten.Image, x, y, size, border float64, ring col, cs ebiten.ColorScale) {
	drawImg(img, x+border, y+border, size-2*border, size-2*border, cs)
	if border > 0 {
		strokeCircle(x+size/2, y+size/2, size/2-border/2, border, ring)
	}
}

// drawFiltered draws img through a colour matrix (for grayscale/dim effects).
func drawFiltered(img *ebiten.Image, px, py float64, cm colorm.ColorM) {
	op := &colorm.DrawImageOptions{}
	op.GeoM.Translate(px, py)
	colorm.DrawImage(cv.dst, img, cm, op)
}

/* ------------------------------------------------------------------- text */

type fontID int

const (
	fDisplay  fontID = iota // Italiana
	fNum                    // Lora
	fNum6                   // Lora semibold (600)
	fNum7                   // Lora bold (700)
	fNumIt                  // Lora italic
	fBody                   // Outfit
	fBodyBold               // Outfit bold
)

// fontSrc holds the font files; Lora is a variable font, so the weights
// share one source.
var fontSrc [7]*text.GoTextFaceSource

var fontWeight = [7]float32{0, 400, 600, 700, 400, 0, 0}

type faceKey struct {
	f    fontID
	size float64
}

var faces = map[faceKey]*text.GoTextFace{}

var wght = text.MustParseTag("wght")

func face(f fontID, size float64) *text.GoTextFace {
	k := faceKey{f, math.Round(size*cv.k*4) / 4}
	if fc, ok := faces[k]; ok {
		return fc
	}
	fc := &text.GoTextFace{Source: fontSrc[f], Size: k.size}
	if fontWeight[f] > 0 {
		fc.SetVariation(wght, fontWeight[f])
	}
	faces[k] = fc
	return fc
}

// textW is the logical width of s, including CSS letter-spacing.
func textW(f fontID, size float64, s string, ls float64) float64 {
	if s == "" {
		return 0
	}
	fc := face(f, size)
	if ls == 0 {
		return text.Advance(s, fc) / cv.k
	}
	w := 0.0
	for _, r := range s {
		w += text.Advance(string(r), fc)/cv.k + ls
	}
	return w
}

const (
	alignL = iota
	alignC
	alignR
)

// T describes how to draw one run of text.
type T struct {
	F     fontID
	Size  float64
	Col   col
	Align int
	LS    float64 // letter-spacing
	LH    float64 // line-height in logical px (0 = the font's normal)
	Shad  col     // text-shadow colour (alpha 0 = none)
	ShadY float64
	ShadB float64 // text-shadow blur
}

// lineOffset returns how far below the line box top the glyph box starts.
func lineOffset(f fontID, size, lh float64) float64 {
	m := face(f, size).Metrics()
	asc, desc := m.HAscent/cv.k, m.HDescent/cv.k
	if lh <= 0 {
		lh = asc + desc + m.HLineGap/cv.k
	}
	return (lh - asc - desc) / 2
}

// normalLH is the font's normal line height.
func normalLH(f fontID, size float64) float64 {
	m := face(f, size).Metrics()
	return (m.HAscent + m.HDescent + m.HLineGap) / cv.k
}

// txt draws s with its CSS line box top at y. x is the left edge, centre or
// right edge depending on t.Align. Returns the width drawn.
func txt(s string, x, y float64, t T) float64 {
	if s == "" {
		return 0
	}
	w := textW(t.F, t.Size, s, t.LS)
	switch t.Align {
	case alignC:
		x -= w / 2
	case alignR:
		x -= w
	}
	y += lineOffset(t.F, t.Size, t.LH)
	if t.Shad.a > 0 {
		sh := t.Shad
		if t.ShadB > 0 {
			// a soft text-shadow: a few faint copies around the offset
			sh.a *= .35
			for _, d := range [][2]float64{{0, 0}, {-.7, 0}, {.7, 0}, {0, -.7}, {0, .7}, {-1.2, 1}, {1.2, 1}} {
				s2 := t.ShadB * .5
				rawText(s, x+d[0]*s2, y+t.ShadY+d[1]*s2, t, sh)
			}
		} else {
			rawText(s, x, y+t.ShadY, t, sh)
		}
	}
	rawText(s, x, y, t, t.Col)
	return w
}

func rawText(s string, x, y float64, t T, c col) {
	fc := face(t.F, t.Size)
	op := &text.DrawOptions{}
	op.ColorScale = c.scale()
	op.Filter = ebiten.FilterLinear
	if t.LS == 0 {
		op.GeoM.Translate(float64(X(x)), float64(Y(y)))
		text.Draw(cv.dst, s, fc, op)
		return
	}
	px := float64(X(x))
	for _, r := range s {
		ch := string(r)
		op.GeoM.Reset()
		op.GeoM.Translate(px, float64(Y(y)))
		text.Draw(cv.dst, ch, fc, op)
		px += text.Advance(ch, fc) + t.LS*cv.k
	}
}

/* -------------------------------------------------------------- rich text */

// A rich string is plain text with <b>bold</b> runs, like the original log.
type piece struct {
	s    string
	bold bool
	x    float64
}

type richLine struct {
	pieces []piece
	w      float64
}

func parseRich(s string) []piece {
	var out []piece
	bold := false
	for s != "" {
		tag := "<b>"
		if bold {
			tag = "</b>"
		}
		i := strings.Index(s, tag)
		if i < 0 {
			out = append(out, piece{s: s, bold: bold})
			break
		}
		if i > 0 {
			out = append(out, piece{s: s[:i], bold: bold})
		}
		s = s[i+len(tag):]
		bold = !bold
	}
	return out
}

// wrapRich lays rich text out into lines no wider than maxW.
func wrapRich(s string, reg, bold fontID, size, maxW float64) []richLine {
	type word struct {
		s    string
		bold bool
		gap  bool // a space comes before this word
	}
	var words []word
	space := true
	for _, p := range parseRich(s) {
		parts := strings.Split(p.s, " ")
		for i, w := range parts {
			if i > 0 {
				space = true
			}
			if w == "" {
				continue
			}
			words = append(words, word{w, p.bold, space && len(words) > 0})
			space = false
		}
	}
	var lines []richLine
	var cur richLine
	x := 0.0
	for _, w := range words {
		f := reg
		if w.bold {
			f = bold
		}
		ww := textW(f, size, w.s, 0)
		sp := 0.0
		if w.gap {
			sp = textW(f, size, " ", 0)
		}
		if len(cur.pieces) > 0 && x+sp+ww > maxW+.5 {
			cur.w = x
			lines = append(lines, cur)
			cur, x, sp = richLine{}, 0, 0
		}
		x += sp
		cur.pieces = append(cur.pieces, piece{s: w.s, bold: w.bold, x: x})
		x += ww
	}
	if len(cur.pieces) > 0 {
		cur.w = x
		lines = append(lines, cur)
	}
	return lines
}

// drawRich draws wrapped rich text from (x, y); returns the height used.
func drawRich(lines []richLine, x, y, lh float64, reg, bold fontID, size float64, c, bc col, alpha float32) float64 {
	for i, l := range lines {
		for _, p := range l.pieces {
			f, cc := reg, c
			if p.bold {
				f, cc = bold, bc
			}
			txt(p.s, x+p.x, y+float64(i)*lh, T{F: f, Size: size, Col: cc.A(alpha), LH: lh})
		}
	}
	return float64(len(lines)) * lh
}
