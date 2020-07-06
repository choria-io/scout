package main

//go:generate go run plugin_generate.go

import (
	"fmt"
	"os"

	"github.com/choria-io/scout/cmd"
)

func main() {
	err := cmd.Run()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Choria Scout failed to run: %s", err)
		os.Exit(1)
	}

	os.Exit(0)
}
