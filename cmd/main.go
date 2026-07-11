package main

import (
	"fmt"
	"os"

	"github.com/egomes/mdo/internal/app"
)

func main() {
	if err := app.Run(os.Args[1:]); err != nil {
		fmt.Fprintf(os.Stderr, "mdo: %v\n", err)
		os.Exit(1)
	}
}
