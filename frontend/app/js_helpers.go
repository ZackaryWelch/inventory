//go:build js && wasm

package app

import (
	"syscall/js"
)

// redirectToURL navigates the browser to the specified URL
func redirectToURL(url string) {
	js.Global().Get("window").Get("location").Set("href", url)
}

// getCurrentURL returns the current browser URL
func getCurrentURL() string {
	return js.Global().Get("window").Get("location").Get("href").String()
}

// getCurrentPath returns just the path portion of the URL
func getCurrentPath() string {
	return js.Global().Get("window").Get("location").Get("pathname").String()
}

// getURLParam returns a URL parameter value
func getURLParam(param string) string {
	urlParams := js.Global().Get("URLSearchParams").New(
		js.Global().Get("window").Get("location").Get("search"),
	)
	return urlParams.Call("get", param).String()
}

// redirectToPath changes the URL path without reloading the page
func (ga *GioApp) redirectToPath(path string) {
	history := js.Global().Get("history")
	history.Call("pushState", nil, "", path)
}
