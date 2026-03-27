package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
)

type gameState int

const (
	stateWaiting gameState = iota
	stateRunning
)

type Game struct {
	board        *Board
	ui           *UI
	state        gameState
	screenW      int
	screenH      int
	boardBottomY float64 // updated each Draw, used in Update for UI hit testing
}

func NewGame() *Game {
	g := &Game{
		board: NewBoard(6, 22),
		ui:    NewUI(),
		state: stateWaiting,
	}
	initAudio()
	registerJS(g)
	return g
}

func (g *Game) Update() error {
	checkPending(g)

	if g.ui.Update(g.screenW, g.boardBottomY) {
		text := g.ui.TakeText()
		if text != "" {
			g.state = stateRunning
			g.board.SetText(text)
		}
	}

	if g.state == stateRunning {
		g.board.Update()
	}
	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	g.screenW, g.screenH = w, h
	screen.Fill(color.RGBA{0x1a, 0x1a, 0x1a, 0xff})

	g.boardBottomY = g.board.Draw(screen, w, h)

	if g.state == stateWaiting {
		drawPrompt(screen, w, h)
	}

	g.ui.Draw(screen, g.boardBottomY)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
