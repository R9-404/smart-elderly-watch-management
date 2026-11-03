//go:build !linux || !amd64 || !cgo || purego

package handlers

func unpackBMPNative(_, _ []byte, _, _, _, _ int, _ bool) bool { return false }
