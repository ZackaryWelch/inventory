//go:build js && wasm

package app

// handleLogin initiates the browser-based OAuth login flow.
func (ga *GioApp) handleLogin() {
	ga.logger.Info("Initiating login")
	if err := ga.authService.InitiateLogin(); err != nil {
		ga.logger.Error("Failed to initiate login", "error", err)
	}
}
