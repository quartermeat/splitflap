package main

import (
	"image/color"
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

const (
	uiBarHeight  = 80
	uiPadding    = 16
	uiBtnWidth   = 100
	uiCursorRate = 30 // ticks per blink cycle half
)

type UI struct {
	text       []rune
	cursorTick int
	hoverBtn   bool
	clickBtn   bool
}

func NewUI() *UI {
	return &UI{}
}

// Update handles keyboard input and button clicks. Returns true when text should be sent.
func (u *UI) Update(screenW, screenH int) bool {
	send := false

	// Accept printable chars from keyboard.
	for _, r := range ebiten.AppendInputChars(nil) {
		r = unicode.ToUpper(r)
		if isAllowed(r) && len(u.text) < 132 { // 6 rows × 22 cols
			u.text = append(u.text, r)
		}
	}

	// Backspace.
	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(u.text) > 0 {
		u.text = u.text[:len(u.text)-1]
	}

	// Shift+Enter = newline, Enter alone = send.
	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			if len(u.text) < 132 {
				u.text = append(u.text, '\n')
			}
		} else {
			send = u.doSend()
		}
	}

	// Button hit test.
	mx, my := ebiten.CursorPosition()
	btnX, btnY, btnW, btnH := u.btnRect(screenW, screenH)
	u.hoverBtn = float64(mx) >= btnX && float64(mx) <= btnX+btnW &&
		float64(my) >= btnY && float64(my) <= btnY+btnH

	if u.hoverBtn && inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
		send = u.doSend()
	}

	u.cursorTick++

	return send
}

func (u *UI) doSend() bool {
	if len(u.text) == 0 {
		return false
	}
	return true
}

func (u *UI) TakeText() string {
	s := strings.ToUpper(string(u.text))
	u.text = u.text[:0] // clear after send
	return s
}

func (u *UI) btnRect(screenW, screenH int) (x, y, w, h float64) {
	barY := float64(screenH) - uiBarHeight
	w = uiBtnWidth
	h = float64(uiBarHeight) - uiPadding*2
	x = float64(screenW) - uiPadding - w
	y = barY + uiPadding
	return
}

func (u *UI) Draw(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	barY := float64(h) - uiBarHeight

	// Bar background.
	vector.DrawFilledRect(screen, 0, float32(barY), float32(w), uiBarHeight,
		color.RGBA{0x15, 0x15, 0x15, 0xff}, false)
	// Top border line.
	vector.DrawFilledRect(screen, 0, float32(barY), float32(w), 1,
		color.RGBA{0x30, 0x30, 0x30, 0xff}, false)

	// Input field background.
	btnX, btnY, btnW, btnH := u.btnRect(w, h)
	inputX := float64(uiPadding)
	inputY := barY + uiPadding
	inputW := btnX - float64(uiPadding)*2
	inputH := btnH

	vector.DrawFilledRect(screen, float32(inputX), float32(inputY), float32(inputW), float32(inputH),
		color.RGBA{0x22, 0x22, 0x22, 0xff}, false)
	vector.StrokeRect(screen, float32(inputX), float32(inputY), float32(inputW), float32(inputH),
		1, color.RGBA{0x40, 0x40, 0x40, 0xff}, false)

	// Input text — show newlines as ↵ symbol.
	face := newFace(18)
	displayText := strings.ReplaceAll(string(u.text), "\n", " ↵ ")
	// Cursor blink.
	if (u.cursorTick/uiCursorRate)%2 == 0 {
		displayText += "_"
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(inputX+10, inputY+(inputH-36)/2)
	op.ColorScale.ScaleWithColor(color.RGBA{0xcc, 0xcc, 0xbb, 0xff})
	text.Draw(screen, displayText, face, op)

	// Send button.
	btnColor := color.RGBA{0x30, 0x30, 0x30, 0xff}
	if u.hoverBtn {
		btnColor = color.RGBA{0x44, 0x44, 0x44, 0xff}
	}
	vector.DrawFilledRect(screen, float32(btnX), float32(btnY), float32(btnW), float32(btnH),
		btnColor, false)
	vector.StrokeRect(screen, float32(btnX), float32(btnY), float32(btnW), float32(btnH),
		1, color.RGBA{0x50, 0x50, 0x50, 0xff}, false)

	btnLabel := "SEND"
	lw, lh := text.Measure(btnLabel, face, 0)
	bop := &text.DrawOptions{}
	bop.GeoM.Translate(btnX+(btnW-lw)/2, btnY+(btnH-lh)/2)
	bop.ColorScale.ScaleWithColor(color.RGBA{0xaa, 0xaa, 0xaa, 0xff})
	text.Draw(screen, btnLabel, face, bop)
}

func isAllowed(r rune) bool {
	for _, c := range charset {
		if c == r {
			return true
		}
	}
	return false
}
