//go:build js && wasm

package app

import (
	"syscall/js"
)

// getCurrentPath returns just the path portion of the URL
func getCurrentPath() string {
	return js.Global().Get("window").Get("location").Get("pathname").String()
}

// redirectToPath changes the URL path without reloading the page
func (ga *GioApp) redirectToPath(path string) {
	history := js.Global().Get("history")
	history.Call("pushState", nil, "", path)
}
