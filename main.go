package main

import (
	"os"

	"optimus/gui"
)

func main() {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh"
	}
	gui.CreateWindow(shell)
}
