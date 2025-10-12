package main

import (
	"context"
	"log/slog"
	"os"
	"syscall"

	"github.com/posener/ctxutil"
	"github.com/nishiki/frontend/app"
)

func main() {
	ctx, cancel := context.WithCancel(ctxutil.WithSignal(context.Background(), syscall.SIGTERM, syscall.SIGINT))
	defer cancel()

	logger := slog.New(slog.NewJSONHandler(os.Stdout, nil))

	// Create the app
	application := app.NewApp()
	
	// Run the desktop application
	application.RunDesktop(ctx, logger)
}