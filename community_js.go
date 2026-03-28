//go:build js && wasm

package main

import (
	"strings"
	"syscall/js"
)

var (
	pendingInitial string
	pendingEvents  []string
	advanceLock    bool
)

func isCommunityMode() bool {
	search := js.Global().Get("location").Get("search").String()
	return !strings.Contains(search, "solo")
}

func initCommunityJS() {
	js.Global().Set("splitflapOnInitialState", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			pendingInitial = args[0].String()
		}
		return nil
	}))

	js.Global().Set("splitflapOnCommunityEvent", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		if len(args) > 0 {
			pendingEvents = append(pendingEvents, args[0].String())
		}
		return nil
	}))

	js.Global().Set("splitflapAdvanceDone", js.FuncOf(func(_ js.Value, args []js.Value) interface{} {
		advanceLock = false
		return nil
	}))

	js.Global().Call("splitflapCommunityInit")
}

func syncCommunity() {
	if !communityMode || comm == nil {
		return
	}
	if pendingInitial != "" {
		comm.applyInitial(pendingInitial)
		pendingInitial = ""
	}
	for _, ev := range pendingEvents {
		comm.applyEvent(ev)
	}
	pendingEvents = nil
}

func maybeAdvanceQueue() {
	if !communityMode || comm == nil || !comm.loaded || advanceLock {
		return
	}
	if comm.active != nil && comm.active.isActive() {
		return
	}
	if len(comm.queue) == 0 {
		return
	}
	advanceLock = true
	nextID := comm.queue[0].ID
	currentID := ""
	if comm.active != nil {
		currentID = comm.active.ID
	}
	js.Global().Call("splitflapCommunityAdvance", nextID, currentID)
}

func submitToCommunity(text string) {
	js.Global().Call("splitflapCommunitySubmit", text)
}

func likeCurrentMessage() {
	if comm == nil {
		return
	}
	msg := comm.currentDisplay()
	if msg == nil {
		return
	}
	js.Global().Call("splitflapCommunityLike", msg.ID)
}

func navigateToMode(community bool) {
	loc := js.Global().Get("location")
	if community {
		loc.Set("href", loc.Get("origin").String()+loc.Get("pathname").String())
	} else {
		loc.Set("search", "?solo")
	}
}
