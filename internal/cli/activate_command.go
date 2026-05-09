package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type ActivateCmd struct {
	Scope         string        `help:"Azure resource scope ID (exact)"`
	Subscription  string        `help:"Subscription ID or exact name"`
	ResourceGroup string        `help:"Resource group name (requires --subscription)"`
	Role          string        `required:"" help:"Role display name or definition ID"`
	Reason        string        `required:"" help:"Justification for the activation"`
	Duration      time.Duration `help:"How long the role stays active (default from config)"`
	JSON          bool          `help:"Output as JSON"`
}

func (c *ActivateCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	act, err := services.Activate(streams.Log)
	if err != nil {
		return err
	}
	duration := c.Duration
	subscription := c.Subscription
	if streams.Config != nil {
		if duration == 0 {
			duration = streams.Config.DefaultDuration
		}
		if subscription == "" && c.Scope == "" {
			subscription = streams.Config.SubscriptionID
		}
	}
	if duration == 0 {
		duration = 2 * time.Hour
	}
	result, err := act.Activate(ctx, domain.ActivationRequest{
		ScopeID:       c.Scope,
		Subscription:  subscription,
		ResourceGroup: c.ResourceGroup,
		Role:          c.Role,
		Reason:        c.Reason,
		Duration:      duration,
	})
	if err != nil {
		return err
	}

	confirmed := c.waitForActive(ctx, services, result, streams)
	if confirmed != nil {
		result = confirmed
	}

	if c.JSON {
		_, err = io.WriteString(streams.Stdout, renderActivationJSON(result))
	} else {
		_, _ = io.WriteString(
			streams.Stderr,
			fmt.Sprintf(
				"Activating %s on %s for %s\n",
				result.Role,
				result.ScopeName,
				result.Duration,
			),
		)
		_, err = io.WriteString(streams.Stdout, renderActivationHuman(result))
	}
	return err
}

func (c *ActivateCmd) waitForActive(
	ctx context.Context,
	services Services,
	result *domain.ActivationResult,
	streams *Streams,
) *domain.ActivationResult {
	statusSvc, err := services.Status(streams.Log)
	if err != nil {
		return nil
	}

	const defaultWaitTimeout = 60 * time.Second
	deadline, cancel := context.WithTimeout(ctx, defaultWaitTimeout)
	defer cancel()

	waitingMessage := fmt.Sprintf("Waiting for %s on %s...\n", result.Role, result.ScopeName)
	_, _ = io.WriteString(streams.Stderr, waitingMessage)

	streams.Log.Debug(
		"waiting for activation to propagate",
		slog.String("role", result.Role),
		slog.String("scope", result.ScopeID),
	)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-deadline.Done():
			_, _ = io.WriteString(streams.Stderr, "Timeout waiting for activation.\n")
			return nil
		case <-ticker.C:
			as, err := statusSvc.Status(deadline)
			if err != nil {
				continue
			}
			for _, a := range as {
				isMatchingAssignment := a.Role == result.Role && a.ScopeID == result.ScopeID
				if isMatchingAssignment && a.Status == domain.ActiveAssignmentActive {
					activeMessage := fmt.Sprintf("✓ %s is active on %s\n", result.Role, result.ScopeName)
					_, _ = io.WriteString(streams.Stderr, activeMessage)
					return &domain.ActivationResult{
						Role:      a.Role,
						ScopeID:   a.ScopeID,
						ScopeName: a.ScopeName,
						Duration:  result.Duration,
						StartedAt: a.StartTime,
						ExpiresAt: a.EndTime,
						Reason:    result.Reason,
					}
				}
			}
		}
	}
}
