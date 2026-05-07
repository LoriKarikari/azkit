package main

import (
	"context"
	"os"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/azurepim"
	"github.com/LoriKarikari/pimctl/internal/cli"
)

func main() {
	runner := cli.NewRunner(cli.Services{List: listService}, os.Stdout, os.Stderr)
	os.Exit(runner.Run(context.Background(), os.Args[1:]))
}

func listService() (*app.ListService, error) {
	store, err := azurepim.NewEligibleAssignments()
	if err != nil {
		return nil, app.AuthFailed(err)
	}
	return app.NewListService(store), nil
}
