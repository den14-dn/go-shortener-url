package main

import (
	"context"
	"go-shortener-url/internal/app"
	"os/signal"
	"syscall"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL)
	defer stop()

	app.Start(ctx)
}
