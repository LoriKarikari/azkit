package main

import (
	"context"
	"io"
	"os"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/azurepim"
	"github.com/LoriKarikari/pimctl/internal/cli"
)

func main() {
	store, err := azurepim.NewEligibleAssignments()
	if err != nil {
		_, _ = io.WriteString(os.Stderr, cli.RenderError(app.AuthFailed(err), false))
		os.Exit(1)
	}
	svc := app.NewListService(store)
	runner := cli.NewRunner(svc, os.Stdout, os.Stderr)
	os.Exit(runner.Run(context.Background(), os.Args[1:]))
}
