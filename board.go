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
}

func NewBoard(rows, cols int) *Board {
	b := &Board{
		rows: rows,
		cols: cols,
		tiles: make([][]*Tile, rows),
	}
	for r := 0; r < rows; r++ {
		b.tiles[r] = make([]*Tile, cols)
		for c := 0; c < cols; c++ {
			b.tiles[r][c] = NewTile()
		}
	}
	return b
}

// SetText sets the board to display the given text.
// Lines are split by newline. Text is centered on each row.
func (b *Board) SetText(s string) {
	s = strings.ToUpper(s)
	lines := strings.Split(s, "\n")

	for r := 0; r < b.rows; r++ {
		var line string
		if r < len(lines) {
			line = lines[r]
		}

		// Pad/center the line.
		if len(line) < b.cols {
			pad := (b.cols - len([]rune(line))) / 2
			line = strings.Repeat(" ", pad) + line
		}

		runes := []rune(line)
		for c := 0; c < b.cols; c++ {
			var ch rune = ' '
			if c < len(runes) {
				ch = unicode.ToUpper(runes[c])
			}
			b.tiles[r][c].SetTarget(ch)
		}
	}
}

func (b *Board) Update() {
	for r := 0; r < b.rows; r++ {
		for c := 0; c < b.cols; c++ {
			b.tiles[r][c].Update()
		}
	}
}

func (b *Board) Draw(screen *ebiten.Image, screenW, screenH int) {
	totalW := float64(b.cols) * (tileWidth + tileGap)
	totalH := float64(b.rows) * (tileHeight + tileGap)

	// Center the board on screen.
	offsetX := (float64(screenW) - totalW) / 2
	offsetY := (float64(screenH) - totalH) / 2

	for r := 0; r < b.rows; r++ {
		for c := 0; c < b.cols; c++ {
			x := offsetX + float64(c)*(tileWidth+tileGap)
			y := offsetY + float64(r)*(tileHeight+tileGap)
			b.tiles[r][c].Draw(screen, x, y)
		}
	}
}
