package interactive

import (
	"context"
	"os"

	"github.com/mattn/go-isatty"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/config"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

func IsTerminal() bool {
	return isatty.IsTerminal(os.Stdin.Fd()) || isatty.IsCygwinTerminal(os.Stdin.Fd())
}

var IsTerminalFn = IsTerminal

// ActivateInteractive is called by the CLI for interactive activation.
var ActivateInteractive = func(
	ctx context.Context,
	eligible []domain.EligibleAssignment,
	svc *app.ActivationService,
	cfg *config.Config,
	input ActivationInput,
) (*domain.ActivationResult, error) {
	return Activate(ctx, eligible, svc, cfg, input)
}

// DeactivateInteractive is the function called by the CLI for interactive deactivation.
var DeactivateInteractive = func(
	ctx context.Context,
	active []domain.ActiveAssignment,
	svc *app.DeactivationService,
	reason string,
	autoConfirm bool,
) (*domain.DeactivationResult, error) {
	return Deactivate(ctx, active, svc, reason, autoConfirm)
}
