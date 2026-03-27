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
	uiBarHeight  = 56
	uiPadding    = 12
	uiCursorRate = 30
)

type UI struct {
	text       []rune
	cursorTick int
	mobileSend bool
}

func NewUI() *UI {
	return &UI{}
}

func (u *UI) inputRect(screenW int, boardBottomY float64) (x, y, w, h float64) {
	x = float64(uiPadding)
	y = boardBottomY + float64(uiPadding)
	w = float64(screenW) - float64(uiPadding)*2
	h = float64(uiBarHeight) - float64(uiPadding)*2
	return
}

func (u *UI) Update(screenW int, boardBottomY float64) bool {
	send := false

	syncMobileInput(u)

	if u.mobileSend {
		u.mobileSend = false
		send = u.doSend()
	}

	for _, r := range ebiten.AppendInputChars(nil) {
		r = unicode.ToUpper(r)
		if isAllowed(r) && len(u.text) < 132 {
			u.text = append(u.text, r)
		}
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyBackspace) && len(u.text) > 0 {
		u.text = u.text[:len(u.text)-1]
	}

	if inpututil.IsKeyJustPressed(ebiten.KeyEnter) {
		if ebiten.IsKeyPressed(ebiten.KeyShift) {
			if len(u.text) < 132 {
				u.text = append(u.text, '\n')
			}
		} else {
			send = u.doSend()
		}
	}

	// Focus mobile keyboard on tap anywhere in the input rect.
	ix, iy, iw, ih := u.inputRect(screenW, boardBottomY)
	for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
		tx, ty := ebiten.TouchPosition(id)
		if float64(tx) >= ix && float64(tx) <= ix+iw &&
			float64(ty) >= iy && float64(ty) <= iy+ih {
			focusMobileKeyboard()
		}
	}

	u.cursorTick++
	return send
}

func (u *UI) doSend() bool {
	return len(u.text) > 0
}

func (u *UI) TakeText() string {
	s := strings.ToUpper(string(u.text))
	u.text = u.text[:0]
	clearMobileKeyboard()
	return s
}

func (u *UI) Draw(screen *ebiten.Image, boardBottomY float64) {
	w := screen.Bounds().Dx()

	// Input field directly below the board.
	ix, iy, iw, ih := u.inputRect(w, boardBottomY)
	vector.DrawFilledRect(screen, float32(ix), float32(iy), float32(iw), float32(ih),
		color.RGBA{0x1e, 0x1e, 0x1e, 0xff}, false)
	vector.StrokeRect(screen, float32(ix), float32(iy), float32(iw), float32(ih),
		1, color.RGBA{0x38, 0x38, 0x38, 0xff}, false)

	// Text + cursor.
	face := newFace(16)
	displayText := strings.ReplaceAll(string(u.text), "\n", " ↵ ")
	if (u.cursorTick/uiCursorRate)%2 == 0 {
		displayText += "_"
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(ix+10, iy+(ih-20)/2)
	op.ColorScale.ScaleWithColor(color.RGBA{0xbb, 0xbb, 0xaa, 0xff})
	text.Draw(screen, displayText, face, op)
}

func isAllowed(r rune) bool {
	for _, c := range charset {
		if c == r {
			return true
		}
	}
	return false
}
