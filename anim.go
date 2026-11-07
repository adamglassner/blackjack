// Animation: easing curves, tweens and the moving things on the table
// (cards, chip stacks, speech bubbles, hand names, confetti).
// (c) 2026 Ironite
package main

import (
	"math"
	"math/rand/v2"
	"time"
)

var epoch = time.Now()

// now is the animation clock in seconds.
func now() float64 { return time.Since(epoch).Seconds() }

/* ---------------------------------------------------------------- easing */

// bezier is a CSS cubic-bezier timing function.
type bezier struct{ x1, y1, x2, y2 float64 }

var (
	easeCard   = &bezier{.2, .75, .25, 1}
	easeIn     = &bezier{.42, 0, 1, 1}
	easeStd    = &bezier{.25, .1, .25, 1}
	easePop    = &bezier{.2, 1.4, .4, 1}
	easeLinear = (*bezier)(nil)
)

func (b *bezier) at(t float64) float64 {
	if b == nil || t <= 0 || t >= 1 {
		return clamp(t, 0, 1)
	}
	// solve x(u) = t for u with Newton's method, then return y(u)
	cx, bx := 3*b.x1, 3*(b.x2-b.x1)-3*b.x1
	ax := 1 - cx - bx
	cy, by := 3*b.y1, 3*(b.y2-b.y1)-3*b.y1
	ay := 1 - cy - by
	u := t
	for i := 0; i < 8; i++ {
		x := ((ax*u+bx)*u+cx)*u - t
		d := (3*ax*u+2*bx)*u + cx
		if math.Abs(x) < 1e-6 || d == 0 {
			break
		}
		u -= x / d
	}
	u = clamp(u, 0, 1)
	return ((ay*u+by)*u + cy) * u
}

// tween is one animated number, like a CSS transition on one property.
type tween struct {
	from, to, t0, dur float64
	ease              *bezier
}

func still(v float64) tween { return tween{from: v, to: v} }

func (t *tween) val() float64 {
	if t.dur <= 0 {
		return t.to
	}
	p := (now() - t.t0) / t.dur
	if p >= 1 {
		return t.to
	}
	if p <= 0 {
		return t.from
	}
	return t.from + (t.to-t.from)*t.ease.at(p)
}

// set starts a transition from wherever the value is now.
func (t *tween) set(v, dur float64, e *bezier) {
	if dur <= 0 {
		*t = still(v)
		return
	}
	cur := t.val()
	*t = tween{from: cur, to: v, t0: now(), dur: dur, ease: e}
}

func (t *tween) done() bool { return t.dur <= 0 || now()-t.t0 >= t.dur }

/* ----------------------------------------------------------------- cards */

const cardW, cardH = 112.0, 157.0

// Card is one card on the table. (x, y) is the unscaled top-left corner;
// the card scales and rotates around its centre, as in the original.
type Card struct {
	id               int   // token
	face             int   // card id shown on the front (-1: unknown)
	x, y, r, s, ang  tween // ang: 0 face up, 180 face down
	z                int
	win, dim, mucked bool
}

type cardLayer struct {
	cards []*Card
	z     int
}

type cardOpts struct {
	x, y, r, s float64
	up         bool
}

func (l *cardLayer) make(id int, o cardOpts) *Card {
	ang := 180.0
	if o.up {
		ang = 0
	}
	l.z++
	c := &Card{id: id, face: -1, x: still(o.x), y: still(o.y), r: still(o.r), s: still(o.s), ang: still(ang), z: l.z}
	l.cards = append(l.cards, c)
	return c
}

type moveOpts struct {
	r, s  *float64
	up    *bool
	dur   float64
	noTop bool
	z     *int
}

func (l *cardLayer) move(c *Card, x, y float64, o moveOpts) {
	d := o.dur
	if d == 0 {
		d = 420
	}
	d /= 1000 * speed
	c.x.set(x, d, easeCard)
	c.y.set(y, d, easeCard)
	if o.r != nil {
		c.r.set(*o.r, d, easeCard)
	} else {
		c.r.set(c.r.to, d, easeCard)
	}
	if o.s != nil {
		c.s.set(*o.s, d, easeCard)
	} else {
		c.s.set(c.s.to, d, easeCard)
	}
	if o.up != nil {
		a := 180.0
		if *o.up {
			a = 0
		}
		c.ang.set(a, d, easeCard)
	} else {
		c.ang.set(c.ang.to, d, easeCard)
	}
	if o.z != nil {
		c.z = *o.z
	} else if !o.noTop {
		l.z++
		c.z = l.z
	}
}

func (l *cardLayer) flip(c *Card, up bool) {
	l.move(c, c.x.to, c.y.to, moveOpts{up: &up, dur: 360, noTop: true})
}

func (l *cardLayer) clear() { l.cards = nil; l.z = 10 }

func fp(v float64) *float64 { return &v }
func bp(v bool) *bool       { return &v }
func ip(v int) *int         { return &v }

func (l *cardLayer) remove(c *Card) {
	for i, x := range l.cards {
		if x == c {
			l.cards = append(l.cards[:i], l.cards[i+1:]...)
			return
		}
	}
}

/* ---------------------------------------------------------- chip stacks */

var denoms = []int{1000, 500, 100, 25, 5}

func breakdown(amount, max int) []int {
	var out []int
	for _, d := range denoms {
		for amount >= d && len(out) < max {
			out = append(out, d)
			amount -= d
		}
	}
	return out
}

/* ------------------------------------------------------------ banner, fx */

type banner struct {
	title, sub string
	show       tween
	hideAt     float64
}

type particle struct {
	x, y, vx, vy, g, life, t, s, r, vr float64
	c                                  col
	circle                             bool
}

type particles struct{ p []particle }

func (ps *particles) burst(x, y float64, n int, colors []col, speedV, size float64, circle bool) {
	for i := 0; i < n; i++ {
		a := -math.Pi/2 + (rand.Float64()-.5)*math.Pi*1.6
		v := speedV * (.4 + rand.Float64()*.8)
		ps.p = append(ps.p, particle{x: x, y: y, vx: math.Cos(a) * v, vy: math.Sin(a) * v, g: 900, life: 2.2,
			c: colors[i%len(colors)], s: size * (.6 + rand.Float64()*.8), r: rand.Float64() * 6, vr: (rand.Float64() - .5) * 12, circle: circle})
	}
}

func (ps *particles) step(dt float64) {
	dt = min(dt, .05)
	keep := ps.p[:0]
	for _, p := range ps.p {
		p.t += dt
		p.vy += p.g * dt
		p.vx *= .992
		p.x += p.vx * dt
		p.y += p.vy * dt
		p.r += p.vr * dt
		if p.t < p.life {
			keep = append(keep, p)
		}
	}
	ps.p = keep
}
