//go:build !js || !wasm

package app

// handleLogin initiates the desktop OAuth PKCE flow via the system browser.
func (ga *GioApp) handleLogin() {
	ga.logger.Info("Initiating desktop login")
	go func() {
		token, err := ga.authService.DesktopLogin()
		if err != nil {
			ga.logger.Error("Desktop login failed", "error", err)
			return
		}
		ga.logger.Info("Desktop login successful", "expires", token.Expiry)
		ga.isSignedIn = true
		ga.currentView = ViewDashboardGio
		ga.loadUserData()
		ga.window.Invalidate()
	}()
}
