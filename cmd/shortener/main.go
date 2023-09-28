package main

import (
	"go-shortener-url/internal/app"
)

func main() {
	//ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGKILL)
	//defer stop()

	app.Start()
}
