package interactive

import (
	"context"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/charmbracelet/huh"
	"github.com/samber/lo"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/config"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

type ActivationInput struct {
	Reason      string
	Duration    time.Duration
	AutoConfirm bool
}

func Activate(
	ctx context.Context,
	eligible []domain.EligibleAssignment,
	svc *app.ActivationService,
	cfg *config.Config,
	input ActivationInput,
) (*domain.ActivationResult, error) {
	if len(eligible) == 0 {
		return nil, app.ErrEligibleNotFound
	}

	defaultDuration := 2 * time.Hour
	if cfg != nil && cfg.DefaultDuration > 0 {
		defaultDuration = cfg.DefaultDuration
	}
	if input.Duration != 0 {
		defaultDuration = input.Duration
	}

	selected := eligible[0]

	if len(eligible) == 1 && !input.AutoConfirm {
		if err := confirmTarget(ctx, selected); err != nil {
			return nil, err
		}
	}

	reason := strings.TrimSpace(input.Reason)
	durationText := defaultDuration.String()
	groups := make([]*huh.Group, 0, 2)

	if len(eligible) > 1 {
		options := lo.Map(eligible, func(a domain.EligibleAssignment, _ int) huh.Option[domain.EligibleAssignment] {
			return huh.NewOption(fmtAssignment(a), a)
		})
		groups = append(groups, huh.NewGroup(
			huh.NewSelect[domain.EligibleAssignment]().
				Title("Select an eligible assignment").
				Options(options...).
				Value(&selected),
		))
	}

	fields := make([]huh.Field, 0, 2)
	if reason == "" {
		fields = append(fields, huh.NewInput().
			Title("Reason for activation").
			Value(&reason).
			Validate(func(s string) error {
				if strings.TrimSpace(s) == "" {
					return errors.New("reason is required")
				}
				return nil
			}))
	}
	if input.Duration <= 0 {
		fields = append(fields, huh.NewInput().
			Title("Duration").
			Value(&durationText).
			Validate(func(s string) error {
				d, err := time.ParseDuration(s)
				if err != nil {
					return fmt.Errorf("invalid duration: %w", err)
				}
				if d <= 0 {
					return errors.New("duration must be positive")
				}
				return nil
			}))
	}
	if len(fields) > 0 {
		groups = append(groups, huh.NewGroup(fields...))
	}

	if len(groups) > 0 {
		form := huh.NewForm(groups...)
		if err := form.RunWithContext(ctx); err != nil {
			if errors.Is(err, huh.ErrUserAborted) {
				return nil, ErrCanceled
			}
			return nil, err
		}
	}

	duration, err := time.ParseDuration(durationText)
	if err != nil {
		return nil, fmt.Errorf("parsing activation duration: %w", err)
	}

	if !input.AutoConfirm {
		if err := confirmActivation(ctx, selected, reason, durationText); err != nil {
			return nil, err
		}
	}

	target := domain.ActivationTarget{
		Assignment: selected,
		Reason:     strings.TrimSpace(reason),
		Duration:   duration,
	}
	if !input.AutoConfirm {
		_, _ = os.Stderr.WriteString("Activating...\n")
	}
	return svc.ActivateResolved(ctx, target)
}

func fmtAssignment(a domain.EligibleAssignment) string {
	if a.SubscriptionName != "" {
		return fmt.Sprintf("%s — %s (%s)", a.Role, a.ScopeName, a.SubscriptionName)
	}
	return fmt.Sprintf("%s — %s", a.Role, a.ScopeName)
}

func confirmTarget(ctx context.Context, a domain.EligibleAssignment) error {
	confirmed := false
	desc := fmt.Sprintf("Role: %s\nScope: %s", a.Role, a.ScopeName)
	if a.SubscriptionName != "" {
		desc += fmt.Sprintf("\nSubscription: %s", a.SubscriptionName)
	}
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Activate this role?").
			Description(desc).
			Affirmative("Continue").
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

func confirmActivation(ctx context.Context, selected domain.EligibleAssignment, reason string, duration string) error {
	confirmed := false
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Activate this role?").
			Description(confirmDescription(selected, reason, duration)).
			Affirmative("Activate").
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

func confirmDescription(a domain.EligibleAssignment, reason string, duration string) string {
	return fmt.Sprintf(
		"Role: %s\nScope: %s\nDuration: %s\nReason: %s",
		a.Role,
		a.ScopeName,
		duration,
		strings.TrimSpace(reason),
	)
}
