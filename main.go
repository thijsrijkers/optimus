package main

import (
	"os"

	"optimus/windowing"
)

func main() {
	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "/bin/zsh"
	}
	windowing.CreateWindow(shell)
}
