package runtime

import (
	"log/slog"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/azurepim"
	"github.com/LoriKarikari/pimctl/internal/cli"
)

type Runtime struct{}

func New() *Runtime {
	return &Runtime{}
}

func (r *Runtime) credential() (azcore.TokenCredential, error) {
	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		return nil, app.AuthFailed(err)
	}
	return cred, nil
}

func (r *Runtime) Services() cli.Services {
	return cli.Services{
		List: func(log *slog.Logger) (*app.ListService, error) {
			cred, err := r.credential()
			if err != nil {
				return nil, err
			}
			store := azurepim.NewEligibleAssignmentsFromCred(cred, log)
			return app.NewListService(store), nil
		},
		Status: func(log *slog.Logger) (*app.StatusService, error) {
			cred, err := r.credential()
			if err != nil {
				return nil, err
			}
			store := azurepim.NewActiveAssignments(cred, log)
			return app.NewStatusService(store), nil
		},
		Activate: func(log *slog.Logger) (*app.ActivationService, error) {
			cred, err := r.credential()
			if err != nil {
				return nil, err
			}
			store := azurepim.NewEligibleAssignmentsFromCred(cred, log)
			active := azurepim.NewActiveAssignments(cred, log)
			activator := azurepim.NewActivationStore(cred, log)
			return app.NewActivationService(store, active, activator), nil
		},
		Deactivate: func(log *slog.Logger) (*app.DeactivationService, error) {
			cred, err := r.credential()
			if err != nil {
				return nil, err
			}
			active := azurepim.NewActiveAssignments(cred, log)
			deactivator := azurepim.NewDeactivationStore(cred, log)
			return app.NewDeactivationService(active, deactivator), nil
		},
	}
}
