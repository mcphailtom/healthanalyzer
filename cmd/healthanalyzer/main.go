package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 && os.Args[1] == "serve" {
		fmt.Println("Starting web server...")
		// TODO: start HTTP server
		return
	}

	fmt.Println("Starting TUI...")
	// TODO: start Bubbletea TUI
}
