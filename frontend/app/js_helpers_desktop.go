//go:build !js || !wasm

package app

// Desktop stubs for browser URL helpers. On desktop, navigation is handled
// by the Gio windowing system rather than browser URL changes.

func getCurrentPath() string { return "" }

func (ga *GioApp) redirectToPath(_ string) {}
