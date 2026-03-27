//go:build js && wasm

package main

import (
	"strings"
	"sync"
	"syscall/js"
	"unicode"
)

var (
	pendingText   string
	pendingMu     sync.Mutex
	hasPending    bool
	jsCallbackReg []js.Func // keep refs to prevent GC
)

// registerJS exposes Go functions to JavaScript.
func registerJS(g *Game) {
	sendFn := js.FuncOf(func(this js.Value, args []js.Value) any {
		if len(args) == 0 {
			return nil
		}
		raw := args[0].String()
		filtered := filterInput(raw)

		pendingMu.Lock()
		pendingText = filtered
		hasPending = true
		pendingMu.Unlock()
		return nil
	})
	js.Global().Set("splitflapSend", sendFn)
	jsCallbackReg = append(jsCallbackReg, sendFn)

	// Expose the allowed charset so JS can filter the input live.
	allowed := allowedChars()
	js.Global().Set("splitflapAllowed", js.ValueOf(allowed))
}

// checkPending is called each Update — applies any pending JS-submitted text.
func checkPending(g *Game) {
	pendingMu.Lock()
	defer pendingMu.Unlock()
	if !hasPending {
		return
	}
	text := pendingText
	hasPending = false

	g.state = stateRunning
	g.board.SetText(text)
}

// filterInput strips any characters not in the display charset.
func filterInput(s string) string {
	allowed := allowedChars()
	var b strings.Builder
	for _, r := range strings.ToUpper(s) {
		if strings.ContainsRune(allowed, r) || r == '\n' {
			b.WriteRune(unicode.ToUpper(r))
		}
	}
	return b.String()
}

func allowedChars() string {
	var b strings.Builder
	for _, r := range charset {
		b.WriteRune(r)
	}
	return b.String()
}
