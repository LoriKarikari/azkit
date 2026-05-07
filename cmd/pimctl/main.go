package main

import (
	"context"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/azurepim"
	"github.com/LoriKarikari/pimctl/internal/cli"
)

func main() {
	runner := cli.NewRunner(cli.Services{
		List:     listService,
		Status:   statusService,
		Activate: activateService,
	}, os.Stdout, os.Stderr)
	os.Exit(runner.Run(context.Background(), os.Args[1:]))
}

func listService() (*app.ListService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, app.AuthFailed(err)
	}
	store := azurepim.NewEligibleAssignmentsFromCred(cred)
	return app.NewListService(store), nil
}

func statusService() (*app.StatusService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, app.AuthFailed(err)
	}
	store := azurepim.NewActiveAssignments(cred)
	return app.NewStatusService(store), nil
}

func activateService() (*app.ActivationService, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, app.AuthFailed(err)
	}
	store := azurepim.NewEligibleAssignmentsFromCred(cred)
	activator := azurepim.NewActivationStore(cred)
	return app.NewActivationService(store, activator), nil
}
