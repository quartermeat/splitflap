package main

import (
	"image/color"
	"math"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/vector"
)

// Character set matching a real split-flap display.
// Blank + A-Z + 0-9 + common punctuation.
var charset = []rune{
	' ',
	'A', 'B', 'C', 'D', 'E', 'F', 'G', 'H', 'I', 'J',
	'K', 'L', 'M', 'N', 'O', 'P', 'Q', 'R', 'S', 'T',
	'U', 'V', 'W', 'X', 'Y', 'Z',
	'0', '1', '2', '3', '4', '5', '6', '7', '8', '9',
	'.', ',', '\'', '!', '?', '-', ':', '/',
}

func charIndex(r rune) int {
	for i, c := range charset {
		if c == r {
			return i
		}
	}
	return 0
}

const (
	tileWidth  = 40
	tileHeight = 60
	tileGap    = 3

	// Ticks for a single flap step when spinning fast (intermediate chars).
	flipTicksFast = 2
	// Ticks for the final flap — slower, dramatic landing.
	flipTicksFinal = 6
	// Extra ticks for the bounce/settle at the very end.
	bounceSettleTicks = 8
)

type flipPhase int

const (
	phaseIdle    flipPhase = iota
	phaseFlip              // top flap falling, bottom revealed
	phaseBounce            // final flap just landed, small bounce
)

type Tile struct {
	current  int
	target   int
	flipFrom int

	phase    flipPhase
	tick     int
	maxTick  int // ticks for current step
	isLast   bool // true when this is the final flip to target

	// Bounce state.
	bounceAngle float64 // current bounce deflection in radians
}

func NewTile() *Tile {
	return &Tile{phase: phaseIdle}
}

func (t *Tile) SetTarget(r rune) {
	idx := charIndex(r)
	if idx == t.current && t.phase == phaseIdle {
		return
	}
	t.target = idx
	if t.phase == phaseIdle {
		t.startNextFlip()
	}
}

func (t *Tile) stepsRemaining() int {
	if t.current == t.target {
		return 0
	}
	steps := t.target - t.current
	if steps <= 0 {
		steps += len(charset)
	}
	return steps
}

func (t *Tile) startNextFlip() {
	t.flipFrom = t.current
	remaining := t.stepsRemaining()
	t.isLast = remaining <= 1
	if t.isLast {
		t.maxTick = flipTicksFinal
	} else {
		t.maxTick = flipTicksFast
	}
	t.phase = phaseFlip
	t.tick = 0
}

func (t *Tile) Update() {
	switch t.phase {
	case phaseIdle:
		return

	case phaseFlip:
		t.tick++
		if t.tick >= t.maxTick {
			t.current = (t.current + 1) % len(charset)
			if t.current == t.target {
				// Landing — hard clack, concurrent.
				go PlayClack()
				t.phase = phaseBounce
				t.tick = 0
				t.bounceAngle = 0.12
				return
			}
			// Intermediate flip — light flicker, concurrent.
			go PlayFlicker()
			t.startNextFlip()
		}

	case phaseBounce:
		t.tick++
		// Damped oscillation.
		decay := math.Exp(-float64(t.tick) * 0.5)
		t.bounceAngle = 0.12 * decay * math.Cos(float64(t.tick)*1.8)
		if t.tick >= bounceSettleTicks {
			t.phase = phaseIdle
			t.bounceAngle = 0
		}
	}
}

func (t *Tile) IsFlipping() bool {
	return t.phase != phaseIdle
}

func (t *Tile) DrawScaled(screen *ebiten.Image, x, y, w, h float64) {
	bgColor := color.RGBA{0x2e, 0x2e, 0x2e, 0xff}
	vector.DrawFilledRect(screen, float32(x), float32(y), float32(w), float32(h), bgColor, false)

	switch t.phase {
	case phaseIdle:
		drawChar(screen, charset[t.current], x, y, w, h, 1.0)
		drawSplitLine(screen, x, y, w, h)
	case phaseFlip:
		t.drawFlipping(screen, x, y, w, h)
	case phaseBounce:
		t.drawBounce(screen, x, y, w, h)
	}
}

func (t *Tile) drawFlipping(screen *ebiten.Image, x, y, w, h float64) {
	progress := float64(t.tick) / float64(t.maxTick)
	nextChar := (t.flipFrom + 1) % len(charset)

	if progress < 0.5 {
		drawCharBottomHalf(screen, charset[nextChar], x, y, w, h)
		flapT := progress * 2
		eased := flapT * flapT
		scaleY := math.Cos(eased * math.Pi / 2)
		if scaleY > 0.01 {
			drawCharTopHalf(screen, charset[t.flipFrom], x, y, w, h, scaleY)
		}
		drawCharTopHalfStatic(screen, charset[t.flipFrom], x, y, w, h)
	} else {
		drawCharTopHalf(screen, charset[nextChar], x, y, w, h, 1.0)
		flapT := (progress - 0.5) * 2
		eased := 1 - (1-flapT)*(1-flapT)
		if eased > 0.01 {
			drawCharBottomHalfScaled(screen, charset[nextChar], x, y, w, h, eased)
		}
	}

	drawSplitLine(screen, x, y, w, h)
}

func (t *Tile) drawBounce(screen *ebiten.Image, x, y, w, h float64) {
	drawChar(screen, charset[t.current], x, y, w, h, 1.0)

	lift := t.bounceAngle * 3
	if lift > 0 {
		liftColor := color.RGBA{0x18, 0x18, 0x18, 0xff}
		liftH := float32(lift * h * 0.15)
		if liftH > 0.5 {
			vector.DrawFilledRect(screen, float32(x), float32(y+h/2), float32(w), liftH, liftColor, false)
		}
	}

	drawSplitLine(screen, x, y, w, h)
}

func drawSplitLine(screen *ebiten.Image, x, y, w, h float64) {
	lineColor := color.RGBA{0x10, 0x10, 0x10, 0xff}
	vector.DrawFilledRect(screen, float32(x), float32(y+h/2-1), float32(w), 2, lineColor, false)
}
