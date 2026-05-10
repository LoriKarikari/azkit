package interactive

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/huh"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/config"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

func Activate(
	ctx context.Context,
	eligible []domain.EligibleAssignment,
	svc *app.ActivationService,
	cfg *config.Config,
) (*domain.ActivationResult, error) {
	defaultDuration := 2 * time.Hour
	if cfg != nil && cfg.DefaultDuration > 0 {
		defaultDuration = cfg.DefaultDuration
	}

	options := make([]huh.Option[domain.EligibleAssignment], len(eligible))
	for i, a := range eligible {
		label := fmtAssignment(a)
		options[i] = huh.NewOption(label, a)
	}

	var selected domain.EligibleAssignment
	var reason string
	durationStr := defaultDuration.String()

	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[domain.EligibleAssignment]().
				Title("Select an eligible assignment").
				Options(options...).
				Value(&selected),
		),
		huh.NewGroup(
			huh.NewInput().
				Title("Reason for activation").
				Value(&reason).
				Validate(func(s string) error {
					if strings.TrimSpace(s) == "" {
						return fmt.Errorf("reason is required")
					}
					return nil
				}),
			huh.NewInput().
				Title("Duration").
				Value(&durationStr).
				Validate(func(s string) error {
					d, err := time.ParseDuration(s)
					if err != nil {
						return fmt.Errorf("invalid duration: %w", err)
					}
					if d <= 0 {
						return fmt.Errorf("duration must be positive")
					}
					return nil
				}),
		),
	)

	if err := form.Run(); err != nil {
		return nil, err
	}

	duration, _ := time.ParseDuration(durationStr)

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
