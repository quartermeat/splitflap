package main

import (
	"image/color"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type gameState int

const (
	stateWaiting gameState = iota
	stateRunning
)

type Game struct {
	board         *Board
	ui            *UI
	state         gameState
	screenW       int
	screenH       int
	boardBottomY  float64
	likeBarY      float64
	lastDisplayID string
}

func NewGame() *Game {
	g := &Game{
		board: NewBoard(6, 22),
		ui:    NewUI(),
		state: stateWaiting,
	}
	initAudio()
	initCommunity()
	registerJS(g)
	return g
}

func (g *Game) Update() error {
	checkPending(g)
	syncCommunity()

	// Community: flip board when displayed message changes.
	if communityMode && comm != nil && comm.loaded {
		maybeAdvanceQueue()
		if msg := comm.currentDisplay(); msg != nil {
			if msg.ID != g.lastDisplayID {
				g.lastDisplayID = msg.ID
				g.state = stateRunning
				g.board.SetText(msg.Text)
			}
		}
	}

	if g.ui.Update(g.screenW, g.boardBottomY) {
		text := g.ui.TakeText()
		if text != "" {
			g.state = stateRunning
			g.board.SetText(text)
			if communityMode {
				submitToCommunity(text)
			}
		}
	}

	if g.state == stateRunning {
		g.board.Update()
	}

	// Like button — mouse and touch.
	if communityMode {
		likeRect := func() (x, y, w, h float64) {
			return float64(uiPadding), g.likeBarY, float64(g.screenW)/2, likeBarH
		}
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			lx, ly, lw, lh := likeRect()
			if float64(mx) >= lx && float64(mx) <= lx+lw && float64(my) >= ly && float64(my) <= ly+lh {
				likeCurrentMessage()
			}
		}
		for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
			tx, ty := ebiten.TouchPosition(id)
			lx, ly, lw, lh := likeRect()
			if float64(tx) >= lx && float64(tx) <= lx+lw && float64(ty) >= ly && float64(ty) <= ly+lh {
				likeCurrentMessage()
			}
		}
	}

	// Mode link — mouse and touch.
	{
		mlx, mly, mlw, mlh := modeLinkRect(g.screenW, g.screenH)
		if inpututil.IsMouseButtonJustPressed(ebiten.MouseButtonLeft) {
			mx, my := ebiten.CursorPosition()
			if float64(mx) >= mlx && float64(mx) <= mlx+mlw && float64(my) >= mly && float64(my) <= mly+mlh {
				navigateToMode(!communityMode)
			}
		}
		for _, id := range inpututil.AppendJustPressedTouchIDs(nil) {
			tx, ty := ebiten.TouchPosition(id)
			if float64(tx) >= mlx && float64(tx) <= mlx+mlw && float64(ty) >= mly && float64(ty) <= mly+mlh {
				navigateToMode(!communityMode)
			}
		}
	}

	return nil
}

func (g *Game) Draw(screen *ebiten.Image) {
	w, h := screen.Bounds().Dx(), screen.Bounds().Dy()
	g.screenW, g.screenH = w, h
	screen.Fill(color.RGBA{0x1a, 0x1a, 0x1a, 0xff})

	// Reserve space below the board for UI elements.
	belowBoard := uiBarHeight + uiPadding*2
	if communityMode {
		belowBoard += int(likeBarH) + uiPadding + int(leaderboardAreaH) + uiPadding
	}
	belowBoard += int(modeLinkAreaH)
	boardAreaH := h - 16 - belowBoard
	if boardAreaH < 80 {
		boardAreaH = 80
	}

	g.boardBottomY = g.board.Draw(screen, w, boardAreaH)

	g.ui.Draw(screen, g.boardBottomY)

	if communityMode && comm != nil {
		g.likeBarY = g.boardBottomY + float64(uiBarHeight)
		drawLikeBar(screen, comm, float64(uiPadding), g.likeBarY, float64(w)-float64(uiPadding)*2)
		lbY := g.likeBarY + likeBarH + float64(uiPadding)
		drawLeaderboard(screen, comm, float64(uiPadding), lbY, float64(w)-float64(uiPadding)*2)
	}

	drawModeLink(screen, w, h, communityMode)
	drawVersion(screen, w, h)
}

func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return outsideWidth, outsideHeight
}
