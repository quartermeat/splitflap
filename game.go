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
	board *Board
	ui   *UI
	state gameState
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

	if g.ui.Update(ebiten.Monitor().Size()) {
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
	screen.Fill(color.RGBA{0x1a, 0x1a, 0x1a, 0xff})

	// Board draws in the area above the UI bar.
	boardH := h - uiBarHeight
	g.board.Draw(screen, w, boardH)

	if g.state == stateWaiting {
		drawPrompt(screen, w, boardH)
	}

	g.ui.Draw(screen)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
