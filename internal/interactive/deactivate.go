package interactive

import (
	"context"
	"errors"
	"fmt"

	"github.com/charmbracelet/huh"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

func Deactivate(
	ctx context.Context,
	active []domain.ActiveAssignment,
	svc *app.DeactivationService,
	reason string,
	autoConfirm bool,
) (*domain.DeactivationResult, error) {
	if len(active) == 0 {
		return nil, app.ErrActiveAssignmentNotFound
	}

	selected := active[0]
	if len(active) > 1 {
		options := make([]huh.Option[domain.ActiveAssignment], len(active))
		for i, a := range active {
			label := fmtActiveAssignment(a)
			options[i] = huh.NewOption(label, a)
		}
		form := huh.NewForm(huh.NewGroup(
			huh.NewSelect[domain.ActiveAssignment]().
				Title("Select an active assignment to deactivate").
				Options(options...).
				Value(&selected),
		))
		if err := form.RunWithContext(ctx); err != nil {
			return nil, err
		}
	}

	if !autoConfirm {
		if err := confirmDeactivation(ctx, selected, reason); err != nil {
			return nil, err
		}
	}

	return svc.Deactivate(ctx, selected.ID, reason)
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
		return err
	}
	if !confirmed {
		return errors.New("deactivation canceled")
	}
	return nil
}

func fmtActiveAssignment(a domain.ActiveAssignment) string {
	if a.SubscriptionName != "" {
		return fmt.Sprintf("%s — %s (%s)", a.Role, a.ScopeName, a.SubscriptionName)
	}
	return fmt.Sprintf("%s — %s", a.Role, a.ScopeName)
}
