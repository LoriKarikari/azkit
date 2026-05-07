package main

import (
	"context"
	"fmt"
	"os"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/azurepim"
	"github.com/LoriKarikari/pimctl/internal/cli"
)

func main() {
	store, err := azurepim.NewEligibleAssignments()
	if err != nil {
		fmt.Fprintln(os.Stderr, "authentication failed:", err)
		os.Exit(1)
	}
	svc := app.NewListService(store)
	runner := cli.NewRunner(svc, os.Stdout, os.Stderr)
	if err := runner.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
