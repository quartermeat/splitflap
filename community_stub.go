//go:build !js

package main

func isCommunityMode() bool      { return false }
func initCommunityJS()           {}
func syncCommunity()             {}
func maybeAdvanceQueue()         {}
func submitToCommunity(_ string) {}
func likeCurrentMessage()        {}
func navigateToMode(_ bool)      {}
