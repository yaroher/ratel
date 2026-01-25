package main

import (
	"os"

	"github.com/yaroher/ratel/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}
