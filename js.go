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
	jsCallbackReg []js.Func
)

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
	js.Global().Set("splitflapAllowed", js.ValueOf(allowedChars()))
}

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

// syncMobileInput reads from the hidden HTML input and syncs to the UI text buffer.
func syncMobileInput(u *UI) {
	jsGlobal := js.Global()

	val := jsGlobal.Call("splitflapGetKbValue").String()
	// Rebuild text buffer from hidden input value.
	u.text = u.text[:0]
	for _, r := range []rune(strings.ToUpper(val)) {
		if r == '\n' {
			u.text = append(u.text, '\n')
		} else if isAllowed(unicode.ToUpper(r)) {
			u.text = append(u.text, unicode.ToUpper(r))
		}
	}

	// Check for special keys.
	key := jsGlobal.Call("splitflapGetLastKey").String()
	if key == "enter" {
		u.mobileSend = true
	}
}

func focusMobileKeyboard() {
	js.Global().Call("splitflapFocusKeyboard")
}

func clearMobileKeyboard() {
	js.Global().Call("splitflapClearKb")
}

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
