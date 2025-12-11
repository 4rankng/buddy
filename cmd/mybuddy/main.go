package main

import (
	"fmt"
	"os"

	"buddy/cmd"
	"buddy/internal/app"
)

func main() {
	ctx, err := app.NewContext("mybuddy")
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	if err := cmd.Execute(ctx); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}
