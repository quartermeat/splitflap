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
	fontFace     text.Face
	fontFaceOnce sync.Once
)

func getFontFace() text.Face {
	fontFaceOnce.Do(func() {
		src, err := text.NewGoTextFaceSource(strings.NewReader(string(gomonobold.TTF)))
		if err != nil {
			panic(err)
		}
		fontFace = &text.GoTextFace{
			Source: src,
			Size:   36,
		}
	})
	return fontFace
}

var charColor = color.RGBA{0xdd, 0xdd, 0xcc, 0xff}

// drawChar draws a character centered in the given rect.
func drawChar(screen *ebiten.Image, r rune, x, y, w, h, alpha float64) {
	if r == ' ' {
		return
	}
	face := getFontFace()
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

// drawCharTopHalf draws only the top half of a character, with vertical scaling.
func drawCharTopHalf(screen *ebiten.Image, r rune, x, y, w, h, scaleY float64) {
	if r == ' ' {
		return
	}

	// Render full char to offscreen image, then draw top half.
	offscreen := ebiten.NewImage(int(w), int(h))
	drawChar(offscreen, r, 0, 0, w, h, 1.0)

	halfH := int(h / 2)
	topHalf := offscreen.SubImage(image.Rect(0, 0, int(w), halfH)).(*ebiten.Image)

	op := &ebiten.DrawImageOptions{}
	// Scale vertically from the bottom of the top half.
	op.GeoM.Translate(0, -float64(halfH))
	op.GeoM.Scale(1, scaleY)
	op.GeoM.Translate(x, y+float64(halfH)*scaleY)
	screen.DrawImage(topHalf, op)
}

// drawCharBottomHalf draws only the bottom half of a character.
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
