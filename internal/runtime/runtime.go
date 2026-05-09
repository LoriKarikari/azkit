package runtime

import (
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/azurepim"
	"github.com/LoriKarikari/pimctl/internal/cli"
)

type Runtime struct{}

func New() *Runtime {
	return &Runtime{}
}

func (r *Runtime) Services() cli.Services {
	return cli.Services{
		List: func(log *slog.Logger) (*app.ListService, error) {
			cred, err := azidentity.NewDefaultAzureCredential(nil)
			if err != nil {
				return nil, app.AuthFailed(err)
			}
			store := azurepim.NewEligibleAssignmentsFromCred(cred, log)
			return app.NewListService(store), nil
		},
		Status: func(log *slog.Logger) (*app.StatusService, error) {
			cred, err := azidentity.NewDefaultAzureCredential(nil)
			if err != nil {
				return nil, app.AuthFailed(err)
			}
			store := azurepim.NewActiveAssignments(cred, log)
			return app.NewStatusService(store), nil
		},
		Activate: func(log *slog.Logger) (*app.ActivationService, error) {
			cred, err := azidentity.NewDefaultAzureCredential(nil)
			if err != nil {
				return nil, app.AuthFailed(err)
			}
			store := azurepim.NewEligibleAssignmentsFromCred(cred, log)
			activator := azurepim.NewActivationStore(cred, log)
			return app.NewActivationService(store, activator), nil
		},
	}
}
