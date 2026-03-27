//go:build !js

package main

func registerJS(g *Game)    {}
func checkPending(g *Game)  {}
func syncMobileInput(u *UI) {}
func focusMobileKeyboard()  {}
func clearMobileKeyboard()  {}
