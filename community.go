package main

import (
	"encoding/json"
	"fmt"
	"image/color"
	"strings"
	"time"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/text/v2"
)

const (
	displayDuration  = 30 * time.Second
	leaderboardMax   = 6
	leaderboardLineH = 18.0
	leaderboardAreaH = leaderboardMax*leaderboardLineH + 8
	likeBarH         = 32.0
	modeLinkAreaH    = 20.0
)

type Message struct {
	ID           string  `json:"id"`
	Text         string  `json:"text"`
	Likes        int     `json:"likes"`
	SubmittedAt  string  `json:"submitted_at"`
	DisplayUntil *string `json:"display_until"`
	IsChampion   bool    `json:"is_champion"`
}

func (m *Message) displayUntilTime() time.Time {
	if m.DisplayUntil == nil {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, *m.DisplayUntil)
	if err != nil {
		// try without nanoseconds
		t, _ = time.Parse("2006-01-02T15:04:05Z", *m.DisplayUntil)
	}
	return t
}

func (m *Message) isActive() bool {
	if m.DisplayUntil == nil {
		return false
	}
	return time.Now().UTC().Before(m.displayUntilTime())
}

type communityInitJSON struct {
	Active      *Message  `json:"active"`
	Champion    *Message  `json:"champion"`
	Queue       []Message `json:"queue"`
	Leaderboard []Message `json:"leaderboard"`
}

type communityEventJSON struct {
	Type    string   `json:"type"`
	Message *Message `json:"message"`
}

type CommunityState struct {
	active      *Message
	champion    *Message
	queue       []Message
	leaderboard []Message
	loaded      bool
}

var (
	comm          *CommunityState
	communityMode bool
)

func initCommunity() {
	comm = &CommunityState{}
	communityMode = isCommunityMode()
	if communityMode {
		initCommunityJS()
	}
}

func (c *CommunityState) currentDisplay() *Message {
	if c.active != nil && c.active.isActive() {
		return c.active
	}
	return c.champion
}

func (c *CommunityState) applyInitial(jsonStr string) {
	var s communityInitJSON
	if err := json.Unmarshal([]byte(jsonStr), &s); err != nil {
		return
	}
	c.active = s.Active
	c.champion = s.Champion
	c.queue = s.Queue
	c.leaderboard = s.Leaderboard
	c.loaded = true
}

func (c *CommunityState) applyEvent(jsonStr string) {
	var ev communityEventJSON
	if err := json.Unmarshal([]byte(jsonStr), &ev); err != nil || ev.Message == nil {
		return
	}
	m := ev.Message

	if ev.Type == "DELETE" {
		if c.active != nil && c.active.ID == m.ID {
			c.active = nil
		}
		if c.champion != nil && c.champion.ID == m.ID {
			c.champion = nil
		}
		c.removeFromLeaderboard(m.ID)
		c.removeFromQueue(m.ID)
		return
	}

	if m.IsChampion {
		c.champion = m
	}
	if m.isActive() {
		c.active = m
	} else if c.active != nil && c.active.ID == m.ID {
		c.active = nil
	}
	c.upsertLeaderboard(m)
	if m.DisplayUntil == nil && !m.IsChampion {
		c.upsertQueue(m)
	} else {
		c.removeFromQueue(m.ID)
	}
}

func (c *CommunityState) upsertLeaderboard(m *Message) {
	for i, l := range c.leaderboard {
		if l.ID == m.ID {
			c.leaderboard[i] = *m
			sortByLikes(c.leaderboard)
			return
		}
	}
	c.leaderboard = append(c.leaderboard, *m)
	sortByLikes(c.leaderboard)
	if len(c.leaderboard) > 20 {
		c.leaderboard = c.leaderboard[:20]
	}
}

func (c *CommunityState) removeFromLeaderboard(id string) {
	for i, l := range c.leaderboard {
		if l.ID == id {
			c.leaderboard = append(c.leaderboard[:i], c.leaderboard[i+1:]...)
			return
		}
	}
}

func (c *CommunityState) upsertQueue(m *Message) {
	for i, q := range c.queue {
		if q.ID == m.ID {
			c.queue[i] = *m
			return
		}
	}
	c.queue = append(c.queue, *m)
}

func (c *CommunityState) removeFromQueue(id string) {
	for i, q := range c.queue {
		if q.ID == id {
			c.queue = append(c.queue[:i], c.queue[i+1:]...)
			return
		}
	}
}

func sortByLikes(msgs []Message) {
	for i := 1; i < len(msgs); i++ {
		for j := i; j > 0 && msgs[j].Likes > msgs[j-1].Likes; j-- {
			msgs[j], msgs[j-1] = msgs[j-1], msgs[j]
		}
	}
}

func drawLikeBar(screen *ebiten.Image, c *CommunityState, x, y, w float64) {
	face := newFace(13)
	var label string
	if c == nil || !c.loaded {
		label = "♥  connecting..."
	} else {
		msg := c.currentDisplay()
		if msg == nil {
			label = "♥  0  —  tap to like"
		} else {
			label = fmt.Sprintf("♥  %d  —  tap to like", msg.Likes)
		}
	}
	op := &text.DrawOptions{}
	op.GeoM.Translate(x+10, y+(likeBarH-15)/2)
	op.ColorScale.ScaleWithColor(color.RGBA{0x55, 0x33, 0x33, 0xff})
	text.Draw(screen, label, face, op)
}

func drawLeaderboard(screen *ebiten.Image, c *CommunityState, x, y, w float64) {
	if c == nil || !c.loaded || len(c.leaderboard) == 0 {
		return
	}
	face := newFace(11)
	count := leaderboardMax
	if len(c.leaderboard) < count {
		count = len(c.leaderboard)
	}
	activeID := ""
	if c.active != nil {
		activeID = c.active.ID
	}
	champID := ""
	if c.champion != nil {
		champID = c.champion.ID
	}
	for i := 0; i < count; i++ {
		msg := c.leaderboard[i]
		preview := strings.ReplaceAll(strings.ToUpper(msg.Text), "\n", " ")
		maxRunes := int(w/7) - 8
		if maxRunes < 5 {
			maxRunes = 5
		}
		if len([]rune(preview)) > maxRunes {
			preview = string([]rune(preview)[:maxRunes]) + "..."
		}
		line := fmt.Sprintf("♥%-3d  %s", msg.Likes, preview)

		col := color.RGBA{0x2a, 0x2a, 0x2a, 0xff}
		if msg.ID == activeID {
			col = color.RGBA{0x55, 0x44, 0x22, 0xff}
		} else if msg.ID == champID {
			col = color.RGBA{0x33, 0x44, 0x22, 0xff}
		}

		op := &text.DrawOptions{}
		op.GeoM.Translate(x, y+float64(i)*leaderboardLineH)
		op.ColorScale.ScaleWithColor(col)
		text.Draw(screen, line, face, op)
	}
}

func drawModeLink(screen *ebiten.Image, screenW, screenH int, isCommunity bool) {
	face := newFace(11)
	label := "SOLO →"
	if !isCommunity {
		label = "← COMMUNITY"
	}
	tw, th := text.Measure(label, face, 0)
	op := &text.DrawOptions{}
	op.GeoM.Translate(float64(screenW)-tw-8, float64(screenH)-th-4)
	op.ColorScale.ScaleWithColor(color.RGBA{0x2e, 0x2e, 0x2e, 0xff})
	text.Draw(screen, label, face, op)
}

func modeLinkRect(screenW, screenH int) (x, y, w, h float64) {
	w, h = 100, 20
	x = float64(screenW) - w - 4
	y = float64(screenH) - h - 4
	return
}
