package main

import (
	"fmt"
	"io"
	"net/http"
	"os"
)

func getRoot(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("got / request\n")
	io.WriteString(w, "This is my website!\n")
}

func main() {
	http.HandleFunc("/", getRoot)

	if err := http.ListenAndServe(":8080", nil); err != nil {
		os.Exit(1) // want "package main in func main contains expression os.Exit"
	}
}
