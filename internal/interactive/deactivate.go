package interactive

import (
	"context"
	"errors"
	"fmt"
	"io"

	"github.com/charmbracelet/huh"
	"github.com/samber/lo"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

type DeactivationInput struct {
	Reason      string
	AutoConfirm bool
	Progress    io.Writer
}

func Deactivate(
	ctx context.Context,
	active []domain.ActiveAssignment,
	svc *app.DeactivationService,
	input DeactivationInput,
) (*domain.DeactivationResult, error) {
	if len(active) == 0 {
		return nil, app.ErrActiveAssignmentNotFound
	}

	selected := active[0]
	if len(active) > 1 {
		options := lo.Map(active, func(a domain.ActiveAssignment, _ int) huh.Option[domain.ActiveAssignment] {
			return huh.NewOption(fmtActiveAssignment(a), a)
		})
		form := huh.NewForm(huh.NewGroup(
			huh.NewSelect[domain.ActiveAssignment]().
				Title("Select an active assignment to deactivate").
				Options(options...).
				Value(&selected),
		))
		if err := form.RunWithContext(ctx); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil, ErrCanceled
			}
			return nil, err
		}
	}

	if !input.AutoConfirm {
		if err := confirmDeactivation(ctx, selected, input.Reason); err != nil {
			return nil, err
		}
	}

	if input.Progress != nil {
		sp := NewSpinner(input.Progress, fmt.Sprintf("Deactivating %s on %s", selected.Role, selected.ScopeName))
		sp.Start()
		result, err := svc.Deactivate(ctx, selected.ID, input.Reason)
		sp.Stop()
		return result, err
	}

	return svc.Deactivate(ctx, selected.ID, input.Reason)
}

func confirmDeactivation(ctx context.Context, selected domain.ActiveAssignment, reason string) error {
	confirmed := false
	desc := fmt.Sprintf("Role: %s\nScope: %s", selected.Role, selected.ScopeName)
	if reason != "" {
		desc += fmt.Sprintf("\nReason: %s", reason)
	}
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Deactivate this role?").
			Description(desc).
			Affirmative("Deactivate").
			Negative("Cancel").
			Value(&confirmed),
	))
	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return ErrCanceled
		}
		return err
	}
	if !confirmed {
		return ErrCanceled
	}
	return nil
}

func fmtActiveAssignment(a domain.ActiveAssignment) string {
	if a.SubscriptionName != "" {
		return fmt.Sprintf("%s — %s (%s)", a.Role, a.ScopeName, a.SubscriptionName)
	}
	return fmt.Sprintf("%s — %s", a.Role, a.ScopeName)
}
