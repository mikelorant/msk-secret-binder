package main

import (
	"fmt"
	"os"

	"github.com/mikelorant/msk-secret-binder/internal/app"
)

func main() {
	if err := app.Run(); err != nil {
		fmt.Fprintf(os.Stdout, "error: %v", err)
		os.Exit(1)
	}
}
