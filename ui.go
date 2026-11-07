// Clickable regions, buttons and sliders.
// Draw registers every clickable thing; Update hit-tests the mouse against
// what was drawn last frame (topmost wins).
// (c) 2026 Ironite
package main

import (
	"github.com/hajimehoshi/ebiten/v2"
)

type region struct {
	x0, y0, x1, y1 float32
	id             string
	on             bool
	fn             func()
}

var (
	regs, lastRegs []region
	hoverID        string
	pressID        string // mouse is held down on this
	dragID         string // slider being dragged
	mouseX, mouseY float32
)

// hit registers a clickable box (logical coords) and reports hover/press.
func hit(id string, x, y, w, h float64, on bool) (hover, down bool) {
	regs = append(regs, region{X(x), Y(y), X(x + w), Y(y + h), id, on, nil})
	return on && hoverID == id, on && pressID == id
}

// onClick attaches an action to the region just registered under id.
func onClick(id string, fn func()) {
	for i := len(regs) - 1; i >= 0; i-- {
		if regs[i].id == id {
			regs[i].fn = fn
			return
		}
	}
}

// block makes everything under a box unclickable (used by the veils).
func blockArea(x, y, w, h float64) {
	regs = append(regs, region{X(x), Y(y), X(x + w), Y(y + h), "", false, nil})
}

func topRegion(x, y float32) (region, bool) {
	for i := len(lastRegs) - 1; i >= 0; i-- {
		r := lastRegs[i]
		if x >= r.x0 && x < r.x1 && y >= r.y0 && y < r.y1 {
			return r, true
		}
	}
	return region{}, false
}

func regionByID(id string) (region, bool) {
	for _, r := range lastRegs {
		if r.id == id {
			return r, true
		}
	}
	return region{}, false
}

func updateCursor() {
	r, ok := topRegion(mouseX, mouseY)
	if ok && r.on {
		hoverID = r.id
		ebiten.SetCursorShape(ebiten.CursorShapePointer)
	} else {
		hoverID = ""
		ebiten.SetCursorShape(ebiten.CursorShapeDefault)
	}
}

/* ------------------------------------------------------------- widgets */

// barButton is a toolbar button (dark leather, gold edge, 15px Outfit).
func barButton(id, label string, x, y, w, h float64, off bool, icon bool) {
	hov, down := hit(id, x, y, w, h, true)
	if down {
		y++
	}
	f := func(c col) col {
		if off {
			c = c.filt(.8, .75)
		}
		if hov {
			c = c.bright(1.08)
		}
		return c
	}
	shadow(x, y, w, h, 7, 0, 1, 3, 0, rgba(0, 0, 0, .5))
	boxRR(x, y, w, h, 7, 1, f(hex("#3a2a1c")), f(hex("#1c130c")), f(hex("#a8822e")))
	insetHi(x, y, w, h, 7, f(rgba(255, 255, 255, .5)))
	ink := f(hex("#f1dfae"))
	tx := x + w/2
	if icon {
		lw := textW(fBody, 15, label, 0)
		total := 12 + 5 + lw
		noteIcon(tx-total/2+5, y+h/2, ink)
		txt(label, tx-total/2+17, y+7, T{F: fBody, Size: 15, Col: ink})
		return
	}
	txt(label, tx, y+7, T{F: fBody, Size: 15, Col: ink, Align: alignC})
}

// noteIcon draws a small beamed music note (the fonts have no ♫).
func noteIcon(cx, cy float64, c col) {
	fillCircle(cx-3.6, cy+4, 2.4, c)
	fillCircle(cx+3.6, cy+3, 2.4, c)
	strokeLine(cx-1.5, cy+4, cx-1.5, cy-5, 1.2, c)
	strokeLine(cx+5.7, cy+3, cx+5.7, cy-6, 1.2, c)
	var p = rrPath(cx-1.5-.6, cy-6.6, 7.8, 2.2, 0)
	fillPathV(p, cy-7, cy-4, c, c)
}

// bigButton is the large rounded button (Take a Seat, Deal Next Hand...).
func bigButton(id, label string, x, y, w, h, size float64, enabled, ghost bool, ghostCol col) {
	hov, down := hit(id, x, y, w, h, enabled)
	if down {
		y++
	}
	f := func(c col) col {
		if !enabled {
			c = c.filt(.6, 1).A(.45)
		} else if hov {
			c = c.bright(1.07)
		}
		return c
	}
	ink := f(hex("#f1dfae"))
	if ghost {
		ink = f(ghostCol)
		borderRR(x, y, w, h, 10, 1, ink)
	} else {
		if enabled {
			shadow(x, y, w, h, 10, 0, 3, 8, 0, rgba(0, 0, 0, .4))
		}
		boxRR(x, y, w, h, 10, 1, f(hex("#3a2a1c")), f(hex("#1c130c")), f(hex("#a8822e")))
		insetHi(x, y, w, h, 10, f(rgba(255, 255, 255, .55)))
	}
	lh := normalLH(fDisplay, size)
	txt(label, x+w/2, y+(h-lh)/2, T{F: fDisplay, Size: size, Col: ink, Align: alignC})
}

func bigButtonW(label string, size, padX float64) float64 {
	return textW(fDisplay, size, label, 0) + 2*padX + 2
}

// slider is a range input styled like Chrome's with a gold accent colour.
func slider(id string, x, y, w, h float64, v, lo, hi float64, enabled bool) {
	hit(id, x-4, y-4, w+8, h+8, enabled)
	a := float32(1)
	if !enabled {
		a = .45
	}
	cy := y + h/2
	th := 8.0
	frac := 0.0
	if hi > lo {
		frac = clamp((v-lo)/(hi-lo), 0, 1)
	}
	tx := x + 8 + (w-16)*frac
	fillRR(x, cy-th/2, w, th, th/2, hex("#3b3b3b").A(a))
	borderRR(x, cy-th/2, w, th, th/2, 1, hex("#7a7a7a").A(a))
	if frac > 0 {
		fillRR(x, cy-th/2, tx-x, th, th/2, hex("#e2c172").A(a))
	}
	fillCircle(tx, cy, 8, hex("#e2c172").A(a))
}

// sliderValue converts the mouse position into a value for a slider.
func sliderValue(id string, lo, hi, step float64) (float64, bool) {
	r, ok := regionByID(id)
	if !ok {
		return 0, false
	}
	x0, x1 := float64(r.x0)+(4+8)*cv.k, float64(r.x1)-(4+8)*cv.k
	f := clamp((float64(mouseX)-x0)/(x1-x0), 0, 1)
	v := lo + (hi-lo)*f
	if step > 0 {
		v = lo + float64(int((v-lo)/step+.5))*step
	}
	return clamp(v, lo, hi), true
}
