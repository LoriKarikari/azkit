package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/interactive"
)

type ActivateCmd struct {
	Scope         string        `help:"Azure resource scope ID (exact)"`
	Subscription  string        `help:"Subscription ID or exact name"`
	ResourceGroup string        `help:"Resource group name (requires --subscription)"`
	Role          string        `help:"Role display name or definition ID"`
	Reason        string        `help:"Justification for the activation"`
	Duration      time.Duration `help:"How long the role stays active (default from config)"`
	Wait          time.Duration `help:"Wait for Azure to report the role active"`
	JSON          bool          `help:"Output as JSON (alias for --output json)"`
}

type interactiveActivation struct {
	services Services
	streams  *Streams
	act      *app.ActivationService
}

func (c *ActivateCmd) jsonOutput() bool {
	return c.JSON
}

func (c *ActivateCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	needsInteractive := !c.JSON && c.needsInteractive(streams) && interactive.IsTerminalFn()
	if !needsInteractive {
		if err := c.validateNonInteractive(); err != nil {
			return err
		}
	}

	act, err := services.Activate(streams.Log)
	if err != nil {
		return err
	}

	if needsInteractive {
		return c.runInteractive(ctx, interactiveActivation{
			services: services,
			streams:  streams,
			act:      act,
		})
	}

	duration := streams.Config.ActivationDuration(c.Duration)
	subscription := streams.Config.ActivationSubscription(c.Subscription, c.Scope)

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
		if errors.Is(err, interactive.ErrCanceled) {
			_, _ = io.WriteString(streams.Stderr, "Activation canceled.\n")
			return err
		}
		return err
	}

	return c.finishActivation(ctx, services, streams, result, nil)
}

func (c *ActivateCmd) validateNonInteractive() error {
	if strings.TrimSpace(c.Role) == "" {
		return &app.Error{
			Code:    domain.CodeMissingRole,
			Message: "Activation role is required.",
		}
	}
	if strings.TrimSpace(c.Reason) == "" {
		return &app.Error{
			Code:    domain.CodeMissingReason,
			Message: "Activation reason is required.",
		}
	}
	return nil
}

func (c *ActivateCmd) needsInteractive(streams *Streams) bool {
	hasScopeSelector := strings.TrimSpace(c.Scope) != "" || strings.TrimSpace(c.Subscription) != ""
	if !hasScopeSelector && streams.Config != nil && streams.Config.SubscriptionID != "" {
		hasScopeSelector = true
	}
	hasRole := strings.TrimSpace(c.Role) != ""
	hasReason := strings.TrimSpace(c.Reason) != ""
	return !hasScopeSelector || !hasRole || !hasReason
}

func (c *ActivateCmd) runInteractive(ctx context.Context, flow interactiveActivation) error {
	if flow.streams.Config != nil && flow.streams.Config.TenantID != "" {
		_, _ = io.WriteString(flow.streams.Stderr, "Tenant: "+flow.streams.Config.TenantID+"\n")
	}

	listSvc, err := flow.services.List(flow.streams.Log)
	if err != nil {
		return err
	}

	var eligible []domain.EligibleAssignment
	{
		eligible, err = interactive.WithSpinner(
			ctx,
			flow.streams.Stderr,
			"Loading eligible assignments",
			!c.JSON,
			func(ctx context.Context) ([]domain.EligibleAssignment, error) {
				return listSvc.List(ctx)
			},
		)
		if err != nil {
			return err
		}
	}

	eligible = c.filterEligible(eligible)
	if len(eligible) == 0 {
		return app.ErrEligibleNotFound
	}

	input := interactive.ActivationInput{
		Reason:   c.Reason,
		Duration: c.Duration,
	}
	var activity *interactive.Activity
	if !c.JSON {
		activity = interactive.NewActivity(flow.streams.Stderr)
		input.Progress = activity
		input.KeepProgress = c.Wait > 0
	}

	result, err := flow.services.ActivateInteractive(
		ctx,
		eligible,
		flow.act,
		flow.streams.Config,
		input,
	)
	if err != nil {
		if errors.Is(err, interactive.ErrCanceled) {
			_, _ = io.WriteString(flow.streams.Stderr, "Activation canceled.\n")
			return err
		}
		return err
	}

	return c.finishActivation(ctx, flow.services, flow.streams, result, activity)
}

func (c *ActivateCmd) finishActivation(
	ctx context.Context,
	services Services,
	streams *Streams,
	result *domain.ActivationResult,
	activity *interactive.Activity,
) error {
	if c.Wait > 0 && result.Outcome != domain.ActivationAlreadyActive {
		confirmed := waitForActive(ctx, services, result, streams, c.Wait, activity)
		if confirmed != nil {
			result = confirmed
		} else {
			result.Outcome = domain.ActivationPending
		}
	} else if activity != nil {
		activity.Stop()
	}

	return renderActivationResult(streams, result, c.JSON)
}

func (c *ActivateCmd) filterEligible(eligible []domain.EligibleAssignment) []domain.EligibleAssignment {
	return app.FilterEligibleAssignments(eligible, domain.ActivationRequest{
		ScopeID:       c.Scope,
		Subscription:  c.Subscription,
		ResourceGroup: c.ResourceGroup,
		Role:          c.Role,
	})
}

func renderActivationResult(streams *Streams, result *domain.ActivationResult, asJSON bool) error {
	interactive.ClearProgress(streams.Stderr)
	if asJSON {
		_, err := io.WriteString(streams.Stdout, renderActivationJSON(result))
		return err
	}
	if result.Outcome != domain.ActivationAlreadyActive {
		activatingMsg := fmt.Sprintf(
			"Activating %s on %s for %s\n",
			result.Role,
			result.ScopeName,
			result.Duration,
		)
		if _, err := io.WriteString(streams.Stderr, activatingMsg); err != nil {
			streams.Log.Debug("failed to write to stderr", slog.Any("error", err))
		}
	}
	_, err := io.WriteString(streams.Stdout, renderActivationHuman(result))
	return err
}

func waitForActive(
	ctx context.Context,
	services Services,
	result *domain.ActivationResult,
	streams *Streams,
	wait time.Duration,
	activity *interactive.Activity,
) *domain.ActivationResult {
	statusSvc, err := services.Status(streams.Log)
	if err != nil {
		if activity != nil {
			activity.Stop()
		}
		return nil
	}

	deadline, cancel := context.WithTimeout(ctx, wait)
	defer cancel()

	if activity == nil {
		activity = interactive.NewActivity(streams.Stderr)
	}
	activity.Start(deadline, fmt.Sprintf("Activating %s on %s", result.Role, result.ScopeName))

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
			activity.Stop()
			message := fmt.Sprintf("Activation was accepted, but Azure did not report it active within %s. "+
				"Run azkit pim status to check again.\n", wait)
			if errors.Is(deadline.Err(), context.Canceled) {
				message = "Activation wait canceled. Run azkit pim status to check whether Azure finished it.\n"
			}
			if _, err := io.WriteString(streams.Stderr, message); err != nil {
				streams.Log.Debug("failed to write to stderr", slog.Any("error", err))
			}
			return nil
		case <-ticker.C:
			as, err := statusSvc.StatusForScope(deadline, result.ScopeID)
			if err != nil {
				streams.Log.Debug(
					"activation status poll failed",
					slog.String("scope", result.ScopeID),
					slog.Any("error", err),
				)
				continue
			}
			streams.Log.Debug(
				"activation poll returned",
				slog.Int("active_assignments", len(as)),
				slog.String("role", result.Role),
			)
			for _, a := range as {
				if domain.ActiveAssignmentConfirmsResult(a, result) {
					activity.Stop()
					confirmedMsg := fmt.Sprintf(
						"✓ %s is active on %s\n",
						result.Role,
						result.ScopeName,
					)
					if _, err := io.WriteString(streams.Stderr, confirmedMsg); err != nil {
						streams.Log.Debug("failed to write to stderr", slog.Any("error", err))
					}
					confirmed := domain.ActivationResultFromActive(a, domain.ActivationActive)
					confirmed.Duration = result.Duration
					confirmed.Reason = result.Reason
					return &confirmed
				}
			}
		}
	}
}
