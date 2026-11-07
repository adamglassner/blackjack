// Embedded fonts and audio, and the table, deck and chips, which are drawn in
// Go at startup by art.go (the same drawing as the original game).
// (c) 2026 Ironite
package main

import (
	"bytes"
	"embed"
	"encoding/json"
	"image"
	_ "image/png"
	"log"
	"math"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

//go:embed assets/audio/*.ogg assets/audio/tunes.json assets/img/*.png CHANGELOG.md
var files embed.FS

var (
	imgScene  *ebiten.Image
	imgCards  [52]*ebiten.Image
	imgBack   *ebiten.Image
	imgChip   = map[int]*ebiten.Image{}
	imgLogo   *ebiten.Image
	changelog string
	tuneNames []string
)

func mustImage(name string) image.Image {
	b, err := files.ReadFile("assets/" + name)
	if err != nil {
		log.Fatal(err)
	}
	img, _, err := image.Decode(bytes.NewReader(b))
	if err != nil {
		log.Fatalf("%s: %v", name, err)
	}
	return img
}

// artRes picks how many pixels per logical pixel to draw the art at, so it
// stays sharp full screen on this monitor (like the original's setRes).
func artRes() float64 {
	s := 1.0
	if mw, mh := ebiten.Monitor().Size(); mw > 0 && mh > 0 {
		s = min(float64(mw)/stageW, float64(mh)/stageH) * ebiten.Monitor().DeviceScaleFactor()
	}
	return clamp(math.Ceil(s*2)/2, 1.5, 3)
}

func loadAssets() {
	res := artRes()
	cres := max(2, res)
	// draw everything in parallel: the scene, 52 faces, the back and chips
	var (
		wg     sync.WaitGroup
		scene  *Ctx
		faces  [52]*Ctx
		back   *Ctx
		chips  = map[int]*Ctx{}
		chipMu sync.Mutex
	)
	wg.Add(1)
	go func() { defer wg.Done(); scene = renderScene(stageW, 844, res) }()
	for g := 0; g < 4; g++ {
		wg.Add(1)
		go func(g int) {
			defer wg.Done()
			for id := g; id < 52; id += 4 {
				faces[id] = renderFace(id, cres)
			}
		}(g)
	}
	wg.Add(1)
	go func() {
		defer wg.Done()
		back = renderBack(cres)
		for _, d := range denoms {
			n := int(math.Ceil(72 * cres))
			x := newCtx(n, n)
			x.Scale(cres, cres)
			x.drawChip(d, 72)
			chipMu.Lock()
			chips[d] = x
			chipMu.Unlock()
		}
	}()
	wg.Wait()
	imgScene = ebiten.NewImageFromImage(scene.RGBA())
	for i := range faces {
		imgCards[i] = ebiten.NewImageFromImage(faces[i].RGBA())
	}
	imgBack = ebiten.NewImageFromImage(back.RGBA())
	for d, c := range chips {
		imgChip[d] = ebiten.NewImageFromImage(c.RGBA())
	}
	imgLogo = ebiten.NewImageFromImage(mustImage("img/ironite-logo.png"))

	srcs := map[string]*text.GoTextFaceSource{}
	load := func(name string) *text.GoTextFaceSource {
		if s, ok := srcs[name]; ok {
			return s
		}
		b, err := assetFS.ReadFile("assets/fonts/" + name + ".ttf")
		if err != nil {
			log.Fatal(err)
		}
		s, err := text.NewGoTextFaceSource(bytes.NewReader(b))
		if err != nil {
			log.Fatal(err)
		}
		srcs[name] = s
		return s
	}
	for i, n := range []string{"Italiana-Regular", "Lora-Variable", "Lora-Variable", "Lora-Variable", "Lora-Italic-Variable", "Outfit-Regular", "Outfit-Bold"} {
		fontSrc[i] = load(n)
	}
	b, _ := files.ReadFile("CHANGELOG.md")
	changelog = string(b)
	b, _ = files.ReadFile("assets/audio/tunes.json")
	_ = json.Unmarshal(b, &tuneNames)
}

// windowIcons returns the game icon at the sizes Windows asks for.
func windowIcons() []image.Image {
	src := mustImage("img/icon.png")
	var out []image.Image
	for _, s := range []int{16, 32, 48, 64, 128, 256} {
		out = append(out, scaleDown(src, s))
	}
	return out
}

// scaleDown is a simple area-average resize for the window icons.
func scaleDown(src image.Image, size int) image.Image {
	b := src.Bounds()
	dst := image.NewNRGBA(image.Rect(0, 0, size, size))
	sx, sy := float64(b.Dx())/float64(size), float64(b.Dy())/float64(size)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			var r, g, bl, a, n float64
			for yy := int(float64(y) * sy); yy < max(int(float64(y)*sy)+1, int(float64(y+1)*sy)); yy++ {
				for xx := int(float64(x) * sx); xx < max(int(float64(x)*sx)+1, int(float64(x+1)*sx)); xx++ {
					cr, cg, cb, ca := src.At(b.Min.X+xx, b.Min.Y+yy).RGBA()
					r, g, bl, a, n = r+float64(cr), g+float64(cg), bl+float64(cb), a+float64(ca), n+1
				}
			}
			if n == 0 || a == 0 {
				continue
			}
			i := dst.PixOffset(x, y)
			dst.Pix[i], dst.Pix[i+1], dst.Pix[i+2], dst.Pix[i+3] = uint8(r/a*255), uint8(g/a*255), uint8(bl/a*255), uint8(a/n/257)
		}
	}
	return dst
}
