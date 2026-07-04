package runtime

import (
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/azurepim"
	"github.com/LoriKarikari/azkit/internal/cli"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/subscriptionstore"
)

type Runtime struct {
	credentialOnce  sync.Once
	credentialValue azcore.TokenCredential
	credentialErr   error
}

func New() *Runtime {
	return &Runtime{}
}

func (r *Runtime) credential() (azcore.TokenCredential, error) {
	r.credentialOnce.Do(func() {
		r.credentialValue, r.credentialErr = azidentity.NewDefaultAzureCredential(nil)
		if r.credentialErr != nil {
			r.credentialErr = app.AuthFailed(r.credentialErr)
		}
	})
	return r.credentialValue, r.credentialErr
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
		Subscriptions: func(*slog.Logger) (*app.SubscriptionService, error) {
			return app.NewSubscriptionService(
				subscriptionstore.New(),
				subscriptionstore.NewAliases(),
				func(active domain.TenantContext) (app.SubscriptionSource, error) {
					if !strings.EqualFold(os.Getenv("AZURE_TENANT_ID"), active.TenantID) {
						return nil, app.ContextEnvironmentMismatch(active)
					}
					cred, err := r.credential()
					if err != nil {
						return nil, err
					}
					return azurepim.NewSubscriptionSourceFromCred(cred), nil
				},
				time.Now,
			), nil
		},
	}
}
