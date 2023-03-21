package main

import (
	"go-shortener-url/internal/app"

	"fmt"
	"net/http"
)

func main() {
	r := app.NewRouter()
	if err := http.ListenAndServe(":8080", r); err != nil {
		fmt.Println(err)
	}
}
