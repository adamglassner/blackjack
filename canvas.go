// Canvas-style drawing on top of gg, so the card, chip and table art can be
// drawn with the same shapes, gradients and coordinates as the original
// artwork. Paths keep their canvas semantics: fill/stroke keep the path,
// Begin starts a new one, arcs join the current point with a line.
// (c) 2026 Ironite
package main

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"unsafe"

	"github.com/fogleman/gg"
)

/* ------------------------------------------------------------- colours */

var (
	colMu    sync.Mutex
	colCache = map[string]color.NRGBA{}
)

// col parses a CSS colour: #rgb, #rrggbb, rgb(), cssRGBA() or "transparent".
func cssCol(s string) color.NRGBA {
	colMu.Lock()
	c, ok := colCache[s]
	colMu.Unlock()
	if ok {
		return c
	}
	c = parseCol(s)
	colMu.Lock()
	colCache[s] = c
	colMu.Unlock()
	return c
}

func parseCol(s string) color.NRGBA {
	s = strings.TrimSpace(s)
	if strings.HasPrefix(s, "#") {
		h := s[1:]
		if len(h) == 3 {
			h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
		}
		v, _ := strconv.ParseUint(h, 16, 32)
		return color.NRGBA{uint8(v >> 16), uint8(v >> 8), uint8(v), 255}
	}
	if i := strings.Index(s, "("); i > 0 {
		parts := strings.Split(strings.TrimSuffix(s[i+1:], ")"), ",")
		f := func(k int) float64 {
			v, _ := strconv.ParseFloat(strings.TrimSpace(parts[k]), 64)
			return v
		}
		a := 1.0
		if len(parts) > 3 {
			a = f(3)
		}
		return color.NRGBA{u8(f(0)), u8(f(1)), u8(f(2)), u8(a * 255)}
	}
	if s == "transparent" {
		return color.NRGBA{}
	}
	return color.NRGBA{0, 0, 0, 255}
}

func u8(v float64) uint8 {
	if v <= 0 {
		return 0
	}
	if v >= 255 {
		return 255
	}
	return uint8(v + .5)
}

// rgba builds a colour like the CSS cssRGBA() function.
func cssRGBA(r, g, b int, a float64) color.NRGBA {
	return color.NRGBA{uint8(r), uint8(g), uint8(b), u8(a * 255)}
}

func cssAlpha(c color.NRGBA, a float64) color.NRGBA {
	c.A = u8(float64(c.A) * a)
	return c
}

/* ----------------------------------------------------------- gradients */

// Grad is a canvas gradient in user space (converted to device space when
// it is used, like a canvas gradient under the current transform).
type Grad struct {
	rad                    bool
	x0, y0, x1, y1, r0, r1 float64
	stops                  []color.NRGBA
}

func toStops(stops []any) []color.NRGBA {
	out := make([]color.NRGBA, len(stops))
	for i, s := range stops {
		switch v := s.(type) {
		case string:
			out[i] = cssCol(v)
		case color.NRGBA:
			out[i] = v
		}
	}
	return out
}

// lin is a linear gradient with evenly spaced stops (strings or colours).
func lin(x0, y0, x1, y1 float64, stops ...any) *Grad {
	return &Grad{x0: x0, y0: y0, x1: x1, y1: y1, stops: toStops(stops)}
}

// rad is a concentric radial gradient with evenly spaced stops.
func rad(cx, cy, r0, r1 float64, stops ...any) *Grad {
	return &Grad{rad: true, x0: cx, y0: cy, r0: r0, r1: r1, stops: toStops(stops)}
}

const lutN = 1024

type gradPat struct {
	rad                bool
	x0, y0, dx, dy, l2 float64
	r0, r1             float64
	lut                [lutN]color.Color
}

func (p *gradPat) ColorAt(x, y int) color.Color {
	fx, fy := float64(x)+.5, float64(y)+.5
	var t float64
	if p.rad {
		d := math.Hypot(fx-p.x0, fy-p.y0)
		if p.r1 != p.r0 {
			t = (d - p.r0) / (p.r1 - p.r0)
		}
	} else if p.l2 > 0 {
		t = ((fx-p.x0)*p.dx + (fy-p.y0)*p.dy) / p.l2
	}
	if t <= 0 {
		return p.lut[0]
	}
	if t >= 1 {
		return p.lut[lutN-1]
	}
	return p.lut[int(t*(lutN-1)+.5)]
}

func stopAt(stops []color.NRGBA, t float64) color.NRGBA {
	n := len(stops)
	if n == 1 {
		return stops[0]
	}
	f := t * float64(n-1)
	i := int(f)
	if i >= n-1 {
		return stops[n-1]
	}
	k := f - float64(i)
	a, b := stops[i], stops[i+1]
	l := func(p, q uint8) uint8 { return u8(float64(p) + (float64(q)-float64(p))*k) }
	return color.NRGBA{l(a.R, b.R), l(a.G, b.G), l(a.B, b.B), l(a.A, b.A)}
}

/* ----------------------------------------------------------------- Ctx */

// Ctx wraps a gg context with canvas semantics.
type Ctx struct {
	*gg.Context
	cur   bool
	masks []*image.Alpha
}

func newCtx(w, h int) *Ctx {
	if w < 1 {
		w = 1
	}
	if h < 1 {
		h = 1
	}
	dc := gg.NewContext(w, h)
	dc.SetLineCapButt()
	dc.SetLineJoinRound()
	return &Ctx{Context: dc}
}

func (c *Ctx) RGBA() *image.RGBA { return c.Image().(*image.RGBA) }

// mat returns the current affine transform (device = M * user).
func (c *Ctx) mat() (a, b, cc, d, e, f float64) {
	x0, y0 := c.TransformPoint(0, 0)
	x1, y1 := c.TransformPoint(1, 0)
	x2, y2 := c.TransformPoint(0, 1)
	return x1 - x0, y1 - y0, x2 - x0, y2 - y0, x0, y0
}

// S is the current uniform scale (device pixels per user unit).
func (c *Ctx) S() float64 {
	a, b, cc, d, _, _ := c.mat()
	return math.Sqrt(math.Abs(a*d - b*cc))
}

// Save/Restore behave like canvas save()/restore(): gg's Pop keeps the
// clip mask, so the mask is saved and put back here.
func (c *Ctx) Save() {
	c.masks = append(c.masks, c.curMask())
	c.Push()
}

func (c *Ctx) Restore() {
	c.Pop()
	m := c.masks[len(c.masks)-1]
	c.masks = c.masks[:len(c.masks)-1]
	if m == nil {
		c.ResetClip()
	} else {
		_ = c.SetMask(m)
	}
}

func (c *Ctx) curMask() *image.Alpha {
	v := reflect.ValueOf(c.Context).Elem().FieldByName("mask")
	return (*image.Alpha)(unsafe.Pointer(v.Pointer()))
}

func (c *Ctx) Begin() { c.ClearPath(); c.cur = false }
func (c *Ctx) M(x, y float64) {
	c.MoveTo(x, y)
	c.cur = true
}
func (c *Ctx) L(x, y float64) {
	if !c.cur {
		c.M(x, y)
		return
	}
	c.LineTo(x, y)
}
func (c *Ctx) Q(cx, cy, x, y float64) {
	if !c.cur {
		c.M(cx, cy)
	}
	c.QuadraticTo(cx, cy, x, y)
}
func (c *Ctx) Cu(c1x, c1y, c2x, c2y, x, y float64) {
	if !c.cur {
		c.M(c1x, c1y)
	}
	c.CubicTo(c1x, c1y, c2x, c2y, x, y)
}
func (c *Ctx) Z() { c.ClosePath() }

// EllArc adds an elliptical arc (canvas ellipse()/arc() semantics).
func (c *Ctx) EllArc(cx, cy, rx, ry, rot, a0, a1 float64, ccw bool) {
	const tau = 2 * math.Pi
	sweep := a1 - a0
	if !ccw {
		if sweep >= tau {
			sweep = tau
		} else {
			sweep = math.Mod(sweep, tau)
			if sweep < 0 {
				sweep += tau
			}
		}
	} else {
		if -sweep >= tau {
			sweep = -tau
		} else {
			sweep = math.Mod(sweep, tau)
			if sweep > 0 {
				sweep -= tau
			}
		}
	}
	cr, sr := math.Cos(rot), math.Sin(rot)
	pt := func(ux, uy float64) (float64, float64) {
		ux, uy = ux*rx, uy*ry
		return cx + ux*cr - uy*sr, cy + ux*sr + uy*cr
	}
	sx, sy := pt(math.Cos(a0), math.Sin(a0))
	if c.cur {
		c.LineTo(sx, sy)
	} else {
		c.M(sx, sy)
	}
	if sweep == 0 {
		return
	}
	n := int(math.Ceil(math.Abs(sweep) / (math.Pi / 4)))
	d := sweep / float64(n)
	k := 4.0 / 3 * math.Tan(d/4)
	t := a0
	for i := 0; i < n; i++ {
		t1 := t + d
		c0, s0, c1, s1 := math.Cos(t), math.Sin(t), math.Cos(t1), math.Sin(t1)
		p1x, p1y := pt(c0-k*s0, s0+k*c0)
		p2x, p2y := pt(c1+k*s1, s1-k*c1)
		ex, ey := pt(c1, s1)
		c.CubicTo(p1x, p1y, p2x, p2y, ex, ey)
		t = t1
	}
}

func (c *Ctx) Arc(cx, cy, r, a0, a1 float64, ccw bool) { c.EllArc(cx, cy, r, r, 0, a0, a1, ccw) }

func (c *Ctx) Rect(x, y, w, h float64) {
	c.M(x, y)
	c.LineTo(x+w, y)
	c.LineTo(x+w, y+h)
	c.LineTo(x, y+h)
	c.Z()
}

// RR is a rounded rectangle path (starts a new path, like D.rr).
func (c *Ctx) RR(x, y, w, h, r float64) {
	c.Begin()
	r = math.Max(0, math.Min(r, math.Min(w/2, h/2)))
	c.M(x+r, y)
	c.LineTo(x+w-r, y)
	c.Arc(x+w-r, y+r, r, -math.Pi/2, 0, false)
	c.LineTo(x+w, y+h-r)
	c.Arc(x+w-r, y+h-r, r, 0, math.Pi/2, false)
	c.LineTo(x+r, y+h)
	c.Arc(x+r, y+h-r, r, math.Pi/2, math.Pi, false)
	c.LineTo(x, y+r)
	c.Arc(x+r, y+r, r, math.Pi, 1.5*math.Pi, false)
	c.Z()
}

// paint applies a fill or stroke paint (CSS string, colour or *Grad).
func (c *Ctx) pattern(p any) gg.Pattern {
	switch v := p.(type) {
	case string:
		return gg.NewSolidPattern(cssCol(v))
	case color.NRGBA:
		return gg.NewSolidPattern(v)
	case *Grad:
		return c.devGrad(v)
	}
	return gg.NewSolidPattern(color.Black)
}

func (c *Ctx) devGrad(g *Grad) gg.Pattern {
	p := &gradPat{rad: g.rad}
	if g.rad {
		p.x0, p.y0 = c.TransformPoint(g.x0, g.y0)
		s := c.S()
		p.r0, p.r1 = g.r0*s, g.r1*s
	} else {
		x0, y0 := c.TransformPoint(g.x0, g.y0)
		x1, y1 := c.TransformPoint(g.x1, g.y1)
		p.x0, p.y0, p.dx, p.dy = x0, y0, x1-x0, y1-y0
		p.l2 = p.dx*p.dx + p.dy*p.dy
	}
	for i := 0; i < lutN; i++ {
		p.lut[i] = stopAt(g.stops, float64(i)/(lutN-1))
	}
	return p
}

// Fill fills the current path and keeps it (canvas semantics).
func (c *Ctx) Fill(p any) {
	c.SetFillStyle(c.pattern(p))
	c.FillPreserve()
}

// Stroke strokes the current path with a width in user units.
func (c *Ctx) Stroke(p any, w float64) {
	c.SetStrokeStyle(c.pattern(p))
	c.SetLineWidth(w * c.S())
	c.StrokePreserve()
}

// StrokeRound strokes with round caps (the D.line / D.curve style).
func (c *Ctx) StrokeRound(p any, w float64) {
	c.SetLineCapRound()
	c.Stroke(p, w)
	c.SetLineCapButt()
}

func (c *Ctx) ClipPath() { c.ClipPreserve() }

// Dash sets a canvas line dash in user units (nil clears it).
func (c *Ctx) Dash(d ...float64) {
	if len(d) == 0 {
		c.SetDash()
		return
	}
	s := c.S()
	out := make([]float64, len(d))
	for i, v := range d {
		out[i] = v * s
	}
	c.SetDash(out...)
}

func (c *Ctx) FillRect(x, y, w, h float64, p any) {
	c.Begin()
	c.Rect(x, y, w, h)
	c.Fill(p)
	c.Begin()
}

func (c *Ctx) StrokeRect(x, y, w, h float64, p any, lw float64) {
	c.Begin()
	c.Rect(x, y, w, h)
	c.Stroke(p, lw)
	c.Begin()
}

/* ---------------------------------------------------- D.* shape helpers */

// Ell is D.ell: a new elliptical path, filled when fill != nil.
func (c *Ctx) Ell(cx, cy, rx, ry float64, fill any, rot ...float64) {
	c.Begin()
	r := 0.0
	if len(rot) > 0 {
		r = rot[0]
	}
	c.EllArc(cx, cy, math.Abs(rx), math.Abs(ry), r, 0, 2*math.Pi, false)
	if fill != nil {
		c.Fill(fill)
	}
}

// Poly is D.poly.
func (c *Ctx) Poly(pts [][2]float64, fill any) {
	c.Begin()
	c.M(pts[0][0], pts[0][1])
	for _, p := range pts[1:] {
		c.LineTo(p[0], p[1])
	}
	c.Z()
	if fill != nil {
		c.Fill(fill)
	}
}

// Line is D.line: a round-capped segment.
func (c *Ctx) Line(x0, y0, x1, y1 float64, p any, w float64) {
	c.Begin()
	c.M(x0, y0)
	c.LineTo(x1, y1)
	c.StrokeRound(p, w)
}

/* -------------------------------------------------- shadows & blurring */

// Shadow draws a blurred, tinted copy of whatever fn draws, offset by
// (dx, dy) user units, like a CSS box-shadow / drop-shadow.
func (c *Ctx) Shadow(blur, dx, dy float64, sc color.NRGBA, fn func(l *Ctx)) {
	w, h := c.Width(), c.Height()
	l := newCtx(w, h)
	a, b, cc, d, e, f := c.mat()
	l.Translate(e, f)
	l.Rotate(math.Atan2(b, a))
	sx := math.Hypot(a, b)
	l.Scale(sx, (a*d-b*cc)/sx)
	fn(l)
	s := c.S()
	src := l.RGBA()
	alpha := make([]float32, w*h)
	for i := range alpha {
		alpha[i] = float32(src.Pix[i*4+3]) / 255
	}
	blurAlpha(alpha, w, h, blur*s/2)
	ox, oy := int(math.Round(dx*s)), int(math.Round(dy*s))
	dst := c.RGBA()
	ca := float32(sc.A) / 255
	for y := 0; y < h; y++ {
		ty := y + oy
		if ty < 0 || ty >= h {
			continue
		}
		for x := 0; x < w; x++ {
			tx := x + ox
			if tx < 0 || tx >= w {
				continue
			}
			al := alpha[y*w+x] * ca
			if al <= 0 {
				continue
			}
			o := dst.PixOffset(tx, ty)
			blendOver(dst.Pix[o:o+4], sc, al)
		}
	}
}

// blendOver blends a straight colour with coverage al over a premultiplied pixel.
func blendOver(px []uint8, c color.NRGBA, al float32) {
	if al > 1 {
		al = 1
	}
	inv := 1 - al
	px[0] = uint8(float32(c.R)*al + float32(px[0])*inv + .5)
	px[1] = uint8(float32(c.G)*al + float32(px[1])*inv + .5)
	px[2] = uint8(float32(c.B)*al + float32(px[2])*inv + .5)
	px[3] = uint8(255*al + float32(px[3])*inv + .5)
}

// boxBlur approximates a gaussian with standard deviation sigma (pixels)
// by three box blurs in each direction.
func blurAlpha(a []float32, w, h int, sigma float64) {
	if sigma < .3 {
		return
	}
	// three box passes of radius rr have variance close to sigma^2
	rad := int(math.Max(1, math.Round((math.Sqrt(4*sigma*sigma+1)-1)/2)))
	tmp := make([]float32, len(a))
	for pass := 0; pass < 3; pass++ {
		blur1(a, tmp, w, h, rad, true)
		blur1(tmp, a, w, h, rad, false)
	}
}

func blur1(src, dst []float32, w, h, r int, horiz bool) {
	n, m := w, h
	if !horiz {
		n, m = h, w
	}
	inv := 1 / float32(2*r+1)
	for j := 0; j < m; j++ {
		at := func(i int) float32 {
			if i < 0 || i >= n {
				return 0
			}
			if horiz {
				return src[j*w+i]
			}
			return src[i*w+j]
		}
		var sum float32
		for i := -r; i <= r; i++ {
			sum += at(i)
		}
		for i := 0; i < n; i++ {
			if horiz {
				dst[j*w+i] = sum * inv
			} else {
				dst[i*w+j] = sum * inv
			}
			sum += at(i+r+1) - at(i-r)
		}
	}
}

// drawImageAt composites src (premultiplied) onto the context at device
// pixel (x, y) ignoring the transform.
func (c *Ctx) drawImageAt(src *image.RGBA, x, y int) {
	draw.Draw(c.RGBA(), image.Rect(x, y, x+src.Bounds().Dx(), y+src.Bounds().Dy()), src, src.Bounds().Min, draw.Over)
}
