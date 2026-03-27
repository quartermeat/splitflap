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

// charIndex returns the position of a rune in the charset, or 0 (blank) if not found.
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

	// Animation: how many ticks a single flap step takes.
	// At 60fps, 4 ticks = ~67ms per character flip — snappy.
	flipTicksPerStep = 4
)

// Tile represents a single split-flap character tile.
type Tile struct {
	// Current displayed character index in charset.
	current int
	// Target character index to flip to.
	target int

	// Animation state.
	flipping bool
	flipTick int // ticks into current single-step flip (0..flipTicksPerStep)

	// The character we're flipping away from (top of old flap).
	flipFrom int
}

func NewTile() *Tile {
	return &Tile{}
}

// SetTarget sets the target character. The tile will flip forward through the
// charset until it reaches this character, just like a real split-flap.
func (t *Tile) SetTarget(r rune) {
	idx := charIndex(r)
	if idx == t.current && !t.flipping {
		return
	}
	t.target = idx
	if !t.flipping {
		t.startFlip()
	}
}

func (t *Tile) startFlip() {
	t.flipping = true
	t.flipTick = 0
	t.flipFrom = t.current
}

func (t *Tile) Update() {
	if !t.flipping {
		return
	}

	t.flipTick++

	if t.flipTick >= flipTicksPerStep {
		// Advance to next character in the drum.
		t.current = (t.current + 1) % len(charset)
		t.flipTick = 0

		if t.current == t.target {
			t.flipping = false
			return
		}
		// Keep flipping — start next step.
		t.flipFrom = t.current
	}
}

// IsFlipping returns true if the tile is currently animating.
func (t *Tile) IsFlipping() bool {
	return t.flipping
}

// Draw renders the tile at the given position.
func (t *Tile) Draw(screen *ebiten.Image, x, y float64) {
	// Background rectangle.
	bgColor := color.RGBA{0x22, 0x22, 0x22, 0xff}
	vector.DrawFilledRect(screen, float32(x), float32(y), tileWidth, tileHeight, bgColor, false)

	if !t.flipping {
		// Static: just draw the current character centered.
		drawChar(screen, charset[t.current], x, y, tileWidth, tileHeight, 1.0)
		// Draw the horizontal split line.
		drawSplitLine(screen, x, y)
		return
	}

	// Animated flip.
	progress := float64(t.flipTick) / float64(flipTicksPerStep)

	nextChar := (t.flipFrom + 1) % len(charset)

	if progress < 0.5 {
		// First half: top flap falls forward.
		// Bottom half shows the NEXT character (revealed underneath).
		// Top half shows old character, being folded down.

		// Draw bottom half — next char (static, revealed).
		drawCharBottomHalf(screen, charset[nextChar], x, y, tileWidth, tileHeight)

		// Draw top flap — old char, scaling vertically to simulate rotation.
		flapProgress := progress * 2 // 0..1 over first half
		scaleY := math.Cos(flapProgress * math.Pi / 2)
		if scaleY > 0.01 {
			drawCharTopHalf(screen, charset[t.flipFrom], x, y, tileWidth, tileHeight, scaleY)
		}

		// Top half background (static, old char still visible above flap).
		// Actually for the top static portion, keep showing old char.
	} else {
		// Second half: bottom flap swings up into place.
		// Top half shows the NEXT character (already settled).
		// Bottom flap carries the next character, swinging down.

		drawCharTopHalf(screen, charset[nextChar], x, y, tileWidth, tileHeight, 1.0)

		flapProgress := (progress - 0.5) * 2 // 0..1 over second half
		scaleY := math.Sin(flapProgress * math.Pi / 2)
		if scaleY > 0.01 {
			drawCharBottomHalf(screen, charset[nextChar], x, y+tileHeight/2*(1-scaleY), tileWidth, tileHeight*scaleY)
		}
	}

	drawSplitLine(screen, x, y)
}

// drawSplitLine draws the dark horizontal line across the middle of a tile.
func drawSplitLine(screen *ebiten.Image, x, y float64) {
	lineColor := color.RGBA{0x10, 0x10, 0x10, 0xff}
	vector.DrawFilledRect(screen, float32(x), float32(y+tileHeight/2-1), tileWidth, 2, lineColor, false)
}
