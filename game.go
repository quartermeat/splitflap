package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

// Game is the top-level Ebitengine game struct.
type Game struct {
	board *Board
}

func NewGame() *Game {
	g := &Game{
		board: NewBoard(6, 22),
	}
	g.board.SetText("HELLO WORLD")
	return g
}

func (g *Game) Update() error {
	g.board.Update()
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.Fill(color.RGBA{0x1a, 0x1a, 0x1a, 0xff})
	g.board.Draw(screen, screenWidth, screenHeight)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return screenWidth, screenHeight
}
