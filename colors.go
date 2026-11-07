// Colours and small helpers shared by the drawing and the game logic.
// (c) 2026 Ironite
package main

import (
	"image/color"
	"strconv"
	"strings"
)

/* ---------------------------------------------------------------- colours */

// col is a straight (non-premultiplied) RGBA colour, 0..1.
type col struct{ r, g, b, a float32 }

func hex(s string) col {
	s = strings.TrimPrefix(s, "#")
	v, _ := strconv.ParseUint(s, 16, 32)
	if len(s) == 3 {
		r, g, b := v>>8&15, v>>4&15, v&15
		return col{float32(r*17) / 255, float32(g*17) / 255, float32(b*17) / 255, 1}
	}
	return col{float32(v>>16&255) / 255, float32(v>>8&255) / 255, float32(v&255) / 255, 1}
}
func rgba(r, g, b int, a float32) col {
	return col{float32(r) / 255, float32(g) / 255, float32(b) / 255, a}
}
func (c col) A(a float32) col { c.a *= a; return c }
func (c col) lerp(d col, t float32) col {
	return col{c.r + (d.r-c.r)*t, c.g + (d.g-c.g)*t, c.b + (d.b-c.b)*t, c.a + (d.a-c.a)*t}
}

// filt applies CSS grayscale(g) then brightness(b).
func (c col) filt(g, b float32) col {
	l := .2126*c.r + .7152*c.g + .0722*c.b
	c.r = (c.r + (l-c.r)*g) * b
	c.g = (c.g + (l-c.g)*g) * b
	c.b = (c.b + (l-c.b)*g) * b
	return c
}
func (c col) bright(b float32) col {
	return col{min(1, c.r*b), min(1, c.g*b), min(1, c.b*b), c.a}
}
func (c col) std() color.Color {
	return color.NRGBA{uint8(c.r*255 + .5), uint8(c.g*255 + .5), uint8(c.b*255 + .5), uint8(c.a*255 + .5)}
}

/* ------------------------------------------------------------------ misc */

func fmtN(n int) string {
	neg := n < 0
	if neg {
		n = -n
	}
	s := strconv.Itoa(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	if neg {
		return "-" + s
	}
	return s
}

func clamp(v, a, b float64) float64 { return max(a, min(b, v)) }
