package cli

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
	"github.com/LoriKarikari/pimctl/internal/interactive"
)

type ActivateCmd struct {
	Scope         string        `help:"Azure resource scope ID (exact)"`
	Subscription  string        `help:"Subscription ID or exact name"`
	ResourceGroup string        `help:"Resource group name (requires --subscription)"`
	Role          string        `help:"Role display name or definition ID"`
	Reason        string        `help:"Justification for the activation"`
	Duration      time.Duration `help:"How long the role stays active (default from config)"`
	JSON          bool          `help:"Output as JSON"`
}

type interactiveActivation struct {
	services Services
	streams  *Streams
	act      *app.ActivationService
}

func (c *ActivateCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	act, err := services.Activate(streams.Log)
	if err != nil {
		return err
	}

	if c.needsInteractive() && interactive.IsTerminal() {
		return c.runInteractive(ctx, interactiveActivation{
			services: services,
			streams:  streams,
			act:      act,
		})
	}

	if c.Role == "" {
		return &app.Error{Code: app.CodeMissingRole, Message: "Activation role is required."}
	}
	if c.Reason == "" {
		return &app.Error{Code: app.CodeMissingReason, Message: "Activation reason is required."}
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

	activateCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	result, err := act.Activate(activateCtx, domain.ActivationRequest{
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

	confirmed := waitForActive(ctx, services, result, streams)
	if confirmed != nil {
		result = confirmed
	}

	return renderActivationResult(streams, result, c.JSON)
}

func (c *ActivateCmd) needsInteractive() bool {
	hasScopeSelector := c.Scope != "" || c.Subscription != ""
	hasRole := c.Role != ""
	hasReason := c.Reason != ""
	return !hasScopeSelector || !hasRole || !hasReason
}

func (c *ActivateCmd) runInteractive(ctx context.Context, flow interactiveActivation) error {
	listSvc, err := flow.services.List(flow.streams.Log)
	if err != nil {
		return err
	}
	eligible, err := listSvc.List(ctx)
	if err != nil {
		return err
	}
	if c.Role != "" {
		eligible = filterEligibleByRole(eligible, c.Role)
	}
	if len(eligible) == 0 {
		return app.ErrEligibleNotFound
	}

	result, err := interactive.Activate(
		ctx,
		eligible,
		flow.act,
		flow.streams.Config,
	)
	if err != nil {
		return err
	}

	confirmed := waitForActive(ctx, flow.services, result, flow.streams)
	if confirmed != nil {
		result = confirmed
	}

	return renderActivationResult(flow.streams, result, c.JSON)
}

func filterEligibleByRole(
	eligible []domain.EligibleAssignment,
	role string,
) []domain.EligibleAssignment {
	filtered := make([]domain.EligibleAssignment, 0, len(eligible))
	for _, a := range eligible {
		if a.Role == role || a.RoleDefID == role {
			filtered = append(filtered, a)
		}
	}
	return filtered
}

func renderActivationResult(streams *Streams, result *domain.ActivationResult, asJSON bool) error {
	if asJSON {
		_, err := io.WriteString(streams.Stdout, renderActivationJSON(result))
		return err
	}
	activatingMsg := fmt.Sprintf(
		"Activating %s on %s for %s\n",
		result.Role,
		result.ScopeName,
		result.Duration,
	)
	if _, err := io.WriteString(streams.Stderr, activatingMsg); err != nil {
		streams.Log.Debug("failed to write to stderr", slog.Any("error", err))
	}
	_, err := io.WriteString(streams.Stdout, renderActivationHuman(result))
	return err
}

func waitForActive(
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

	waitingMsg := fmt.Sprintf("Waiting for %s on %s...\n", result.Role, result.ScopeName)
	if _, err := io.WriteString(streams.Stderr, waitingMsg); err != nil {
		streams.Log.Debug("failed to write to stderr", slog.Any("error", err))
	}

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
			if _, err := io.WriteString(streams.Stderr, "Timeout waiting for activation.\n"); err != nil {
				streams.Log.Debug("failed to write to stderr", slog.Any("error", err))
			}
			return nil
		case <-ticker.C:
			as, err := statusSvc.Status(deadline)
			if err != nil {
				continue
			}
			for _, a := range as {
				isMatch := a.Role == result.Role && a.ScopeID == result.ScopeID
				if isMatch && a.Status == domain.ActiveAssignmentActive {
					confirmedMsg := fmt.Sprintf(
						"✓ %s is active on %s\n",
						result.Role,
						result.ScopeName,
					)
					if _, err := io.WriteString(streams.Stderr, confirmedMsg); err != nil {
						streams.Log.Debug("failed to write to stderr", slog.Any("error", err))
					}
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
