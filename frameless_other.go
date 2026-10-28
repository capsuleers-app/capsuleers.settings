//go:build !(linux && gtk3)

package main

func init() { framelessFix = nil }
