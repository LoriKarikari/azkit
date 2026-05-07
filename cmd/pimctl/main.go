package main

import (
	"context"
	"fmt"
	"os"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/cli"
	"github.com/LoriKarikari/pimctl/internal/domain"
	"github.com/LoriKarikari/pimctl/internal/inmemory"
)

func main() {
	store := &inmemory.EligibleAssignments{Assignments: fakeAssignments()}
	svc := app.NewListService(store)
	runner := cli.NewRunner(svc, os.Stdout, os.Stderr)
	if err := runner.Run(context.Background(), os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func fakeAssignments() []domain.EligibleAssignment {
	return []domain.EligibleAssignment{
		{
			ID:          "a1",
			Role:        "Contributor",
			ScopeType:   domain.ScopeSubscription,
			ScopeID:     "/subscriptions/00000000-0000-0000-0000-000000000000",
			ScopeName:   "sub-prod",
			MaxDuration: "8h",
		},
		{
			ID:          "a2",
			Role:        "Reader",
			ScopeType:   domain.ScopeResourceGroup,
			ScopeID:     "/subscriptions/00000000-0000-0000-0000-000000000000/resourceGroups/rg-dev-app",
			ScopeName:   "rg-dev-app",
			MaxDuration: "2h",
		},
	}
}
