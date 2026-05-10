package interactive

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/config"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

type ActivationInput struct {
	Reason   string
	Duration time.Duration
}

func Activate(
	ctx context.Context,
	eligible []domain.EligibleAssignment,
	svc *app.ActivationService,
	cfg *config.Config,
	input ActivationInput,
) (*domain.ActivationResult, error) {
	defaultDuration := 2 * time.Hour
	if cfg != nil && cfg.DefaultDuration > 0 {
		defaultDuration = cfg.DefaultDuration
	}
	if input.Duration > 0 {
		defaultDuration = input.Duration
	}

	selected := eligible[0]
	reason := strings.TrimSpace(input.Reason)
	durationText := defaultDuration.String()
	groups := make([]*huh.Group, 0, 2)

	if len(eligible) > 1 {
		options := make([]huh.Option[domain.EligibleAssignment], len(eligible))
		for i, a := range eligible {
			label := fmtAssignment(a)
			options[i] = huh.NewOption(label, a)
		}
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
	if input.Duration == 0 {
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
			return nil, err
		}
	}

	duration, err := time.ParseDuration(durationText)
	if err != nil {
		return nil, fmt.Errorf("parsing activation duration: %w", err)
	}

	target := domain.ActivationTarget{
		Assignment: selected,
		Reason:     strings.TrimSpace(reason),
		Duration:   duration,
	}
	return svc.ActivateResolved(ctx, target)
}

func fmtAssignment(a domain.EligibleAssignment) string {
	if a.SubscriptionName != "" {
		return fmt.Sprintf("%s — %s (%s)", a.Role, a.ScopeName, a.SubscriptionName)
	}
	return fmt.Sprintf("%s — %s", a.Role, a.ScopeName)
}
