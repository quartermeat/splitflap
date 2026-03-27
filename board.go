package main

import (
	"strings"
	"unicode"

	"github.com/hajimehoshi/ebiten/v2"
)

// Board is a grid of split-flap tiles.
type Board struct {
	rows, cols int
	tiles      [][]*Tile

	// Stagger: pending targets that haven't been applied yet.
	pending      [][]rune
	staggerTick  int
	staggerDelay int // ticks between each column starting
}

func NewBoard(rows, cols int) *Board {
	b := &Board{
		rows:         rows,
		cols:         cols,
		tiles:        make([][]*Tile, rows),
		staggerDelay: 2, // 2 ticks (~33ms) between columns
	}
	for r := 0; r < rows; r++ {
		b.tiles[r] = make([]*Tile, cols)
		for c := 0; c < cols; c++ {
			b.tiles[r][c] = NewTile()
		}
	}
	return b
}

// SetText sets the board to display the given text with staggered column starts.
// Text is word-wrapped to fit the board width, then centered per row.
func (b *Board) SetText(s string) {
	s = strings.ToUpper(s)

	// Split explicit newlines first, then word-wrap each segment.
	var lines []string
	for _, seg := range strings.Split(s, "\n") {
		lines = append(lines, wordWrap(seg, b.cols)...)
	}

	b.pending = make([][]rune, b.rows)
	for r := 0; r < b.rows; r++ {
		b.pending[r] = make([]rune, b.cols)
		var line string
		if r < len(lines) {
			line = lines[r]
		}

		runes := []rune(strings.TrimSpace(line))
		padded := make([]rune, b.cols)
		pad := (b.cols - len(runes)) / 2
		for c := 0; c < b.cols; c++ {
			idx := c - pad
			if idx >= 0 && idx < len(runes) {
				padded[c] = unicode.ToUpper(runes[idx])
			} else {
				padded[c] = ' '
			}
		}
		b.pending[r] = padded
	}
	b.staggerTick = 0
}

// wordWrap breaks s into lines of at most width runes, breaking on word boundaries.
func wordWrap(s string, width int) []string {
	words := strings.Fields(s)
	if len(words) == 0 {
		return []string{""}
	}

	var lines []string
	line := ""
	for _, w := range words {
		if len(line) == 0 {
			line = w
		} else if len(line)+1+len(w) <= width {
			line += " " + w
		} else {
			lines = append(lines, line)
			line = w
		}
	}
	if line != "" {
		lines = append(lines, line)
	}
	return lines
}

func (b *Board) Update() {
	// Apply staggered targets.
	if b.pending != nil {
		col := b.staggerTick / b.staggerDelay
		if col < b.cols {
			// Start this column flipping.
			for r := 0; r < b.rows; r++ {
				b.tiles[r][col].SetTarget(b.pending[r][col])
			}
		}
		b.staggerTick++
		if col >= b.cols {
			b.pending = nil
		}
	}

	for r := 0; r < b.rows; r++ {
		for c := 0; c < b.cols; c++ {
			b.tiles[r][c].Update()
		}
	}
}

// Draw renders the board and returns the Y coordinate of the board's bottom edge.
func (b *Board) Draw(screen *ebiten.Image, screenW, screenH int) float64 {
	// Scale tiles to fill 97% of the available area, maintaining aspect ratio.
	margin := 0.97
	scaleX := float64(screenW) * margin / (float64(b.cols) * (tileWidth + tileGap))
	scaleY := float64(screenH) * margin / (float64(b.rows) * (tileHeight + tileGap))
	scale := scaleX
	if scaleY < scale {
		scale = scaleY
	}
	if scale > 2.5 {
		scale = 2.5 // cap so it doesn't get absurdly large on big screens
	}

	tw := float64(tileWidth) * scale
	th := float64(tileHeight) * scale
	tg := float64(tileGap) * scale

	totalW := float64(b.cols)*(tw+tg) - tg
	totalH := float64(b.rows)*(th+tg) - tg

	offsetX := (float64(screenW) - totalW) / 2
	offsetY := 16.0 // top-aligned so keyboard open on mobile doesn't push board off screen

	for r := 0; r < b.rows; r++ {
		for c := 0; c < b.cols; c++ {
			x := offsetX + float64(c)*(tw+tg)
			y := offsetY + float64(r)*(th+tg)
			b.tiles[r][c].DrawScaled(screen, x, y, tw, th)
		}
	}

	return offsetY + totalH
}
