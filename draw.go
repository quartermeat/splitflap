package main

import (
	"image"
	"image/color"
	"strings"
	"sync"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"golang.org/x/image/font/gofont/gomonobold"
)

var (
	fontSrc     *text.GoTextFaceSource
	fontSrcOnce sync.Once
)

func getFontSrc() *text.GoTextFaceSource {
	fontSrcOnce.Do(func() {
		src, err := text.NewGoTextFaceSource(strings.NewReader(string(gomonobold.TTF)))
		if err != nil {
			panic(err)
		}
		fontSrc = src
	})
	return fontSrc
}

func newFace(size float64) *text.GoTextFace {
	return &text.GoTextFace{Source: getFontSrc(), Size: size}
}

// getFontFace returns the tile character face at a given size.
func getFontFace() text.Face { return newFace(36) }

var (
	charColor   = color.RGBA{0xdd, 0xdd, 0xcc, 0xff}
	promptColor = color.RGBA{0x55, 0x55, 0x55, 0xff}
)

// drawPrompt draws the "type a message below" hint centered on screen.
func drawPrompt(screen *ebiten.Image, screenW, screenH int) {
	face := newFace(22)
	msg := "TYPE A MESSAGE BELOW"
	tw, th := text.Measure(msg, face, 0)
	x := (float64(screenW) - tw) / 2
	y := (float64(screenH) - th) / 2

	op := &text.DrawOptions{}
	op.GeoM.Translate(x, y)
	op.ColorScale.ScaleWithColor(promptColor)
	text.Draw(screen, msg, face, op)
}

// drawChar draws a character centered in a tile rect at a given font size.
func drawChar(screen *ebiten.Image, r rune, x, y, w, h, alpha float64) {
	if r == ' ' {
		return
	}
	// Scale font to fit tile height.
	fontSize := h * 0.55
	face := newFace(fontSize)
	s := string(r)

	tw, th := text.Measure(s, face, 0)
	tx := x + (w-tw)/2
	ty := y + (h-th)/2

	op := &text.DrawOptions{}
	op.GeoM.Translate(tx, ty)
	op.ColorScale.ScaleWithColor(charColor)
	if alpha < 1.0 {
		op.ColorScale.ScaleAlpha(float32(alpha))
	}
	text.Draw(screen, s, face, op)
}

// drawCharTopHalf draws only the top half of a character with vertical scaling.
func drawCharTopHalf(screen *ebiten.Image, r rune, x, y, w, h, scaleY float64) {
	if r == ' ' {
		return
	}

	offscreen := ebiten.NewImage(int(w), int(h))
	drawChar(offscreen, r, 0, 0, w, h, 1.0)

	halfH := int(h / 2)
	topHalf := offscreen.SubImage(image.Rect(0, 0, int(w), halfH)).(*ebiten.Image)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(0, -float64(halfH))
	op.GeoM.Scale(1, scaleY)
	op.GeoM.Translate(x, y+float64(halfH))
	screen.DrawImage(topHalf, op)
}

func drawCharTopHalfStatic(screen *ebiten.Image, r rune, x, y, w, h float64) {
	drawCharTopHalf(screen, r, x, y, w, h, 1.0)
}

// drawCharBottomHalf draws the bottom half at full size.
func drawCharBottomHalf(screen *ebiten.Image, r rune, x, y, w, h float64) {
	if r == ' ' {
		return
	}

	offscreen := ebiten.NewImage(int(w), int(h))
	drawChar(offscreen, r, 0, 0, w, h, 1.0)

	halfH := int(h / 2)
	bottomHalf := offscreen.SubImage(image.Rect(0, halfH, int(w), int(h))).(*ebiten.Image)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Translate(x, y+float64(halfH))
	screen.DrawImage(bottomHalf, op)
}

// drawCharBottomHalfScaled draws the bottom half with vertical scaling from the hinge.
func drawCharBottomHalfScaled(screen *ebiten.Image, r rune, x, y, w, h, scaleY float64) {
	if r == ' ' {
		return
	}

	offscreen := ebiten.NewImage(int(w), int(h))
	drawChar(offscreen, r, 0, 0, w, h, 1.0)

	halfH := int(h / 2)
	bottomHalf := offscreen.SubImage(image.Rect(0, halfH, int(w), int(h))).(*ebiten.Image)

	op := &ebiten.DrawImageOptions{}
	op.GeoM.Scale(1, scaleY)
	op.GeoM.Translate(x, y+float64(halfH))
	screen.DrawImage(bottomHalf, op)
}
