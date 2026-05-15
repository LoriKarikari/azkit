package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"time"

	"github.com/samber/lo"

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
	needsInteractive := c.needsInteractive(streams) && interactive.IsTerminalFn()
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
		if errors.Is(err, interactive.ErrCanceled) {
			_, _ = io.WriteString(streams.Stderr, "Activation canceled.\n")
			return nil
		}
		return err
	}

	if result.Outcome != domain.ActivationAlreadyActive {
		confirmed := waitForActive(ctx, services, result, streams)
		if confirmed != nil {
			result = confirmed
		} else {
			result.Outcome = domain.ActivationPending
		}
	}

	return renderActivationResult(streams, result, c.JSON)
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
	listSvc, err := flow.services.List(flow.streams.Log)
	if err != nil {
		return err
	}

	var eligible []domain.EligibleAssignment
	{
		sp := NewSpinner(flow.streams.Stderr, "Loading eligible assignments...")
		if !c.JSON {
			sp.Start()
		}
		eligible, err = listSvc.List(ctx)
		if !c.JSON {
			sp.Stop()
		}
		if err != nil {
			return err
		}
	}

	eligible = c.filterEligible(eligible)
	if len(eligible) == 0 {
		return app.ErrEligibleNotFound
	}

	result, err := flow.services.ActivateInteractive(
		ctx,
		eligible,
		flow.act,
		flow.streams.Config,
		interactive.ActivationInput{
			Reason:   c.Reason,
			Duration: c.Duration,
		},
	)
	if err != nil {
		if errors.Is(err, interactive.ErrCanceled) {
			_, _ = io.WriteString(flow.streams.Stderr, "Activation canceled.\n")
			return nil
		}
		return err
	}

	if result.Outcome != domain.ActivationAlreadyActive {
		confirmed := waitForActive(ctx, flow.services, result, flow.streams)
		if confirmed != nil {
			result = confirmed
		} else {
			result.Outcome = domain.ActivationPending
		}
	}

	return renderActivationResult(flow.streams, result, c.JSON)
}

func (c *ActivateCmd) filterEligible(eligible []domain.EligibleAssignment) []domain.EligibleAssignment {
	return lo.Filter(eligible, func(a domain.EligibleAssignment, _ int) bool {
		return c.matchesEligible(a)
	})
}

func (c *ActivateCmd) matchesEligible(a domain.EligibleAssignment) bool {
	role := strings.TrimSpace(c.Role)
	if role != "" && a.Role != role && a.RoleDefID != role {
		return false
	}
	if c.Scope != "" && a.ScopeID != c.Scope {
		return false
	}
	if c.Subscription != "" && !matchesSubscription(a, c.Subscription) {
		return false
	}
	if c.ResourceGroup != "" && !matchesResourceGroup(a, c.ResourceGroup) {
		return false
	}
	return true
}

func matchesSubscription(a domain.EligibleAssignment, input string) bool {
	input = strings.TrimSpace(input)
	if input == "" {
		return true
	}
	lower := strings.ToLower(input)
	if a.SubscriptionID == input || strings.EqualFold(a.SubscriptionName, input) {
		return true
	}
	if a.ScopeType == domain.ScopeSubscription && strings.EqualFold(a.ScopeName, input) {
		return true
	}
	return strings.HasPrefix(strings.ToLower(a.ScopeID), "/subscriptions/"+lower)
}

func matchesResourceGroup(a domain.EligibleAssignment, input string) bool {
	input = strings.TrimSpace(input)
	if input == "" {
		return true
	}
	return a.ScopeType == domain.ScopeResourceGroup && strings.EqualFold(a.ScopeName, input)
}

func renderActivationResult(streams *Streams, result *domain.ActivationResult, asJSON bool) error {
	_, _ = io.WriteString(streams.Stderr, "\r\033[K")
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

const defaultWaitTimeout = 60 * time.Second

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

	deadline, cancel := context.WithTimeout(ctx, defaultWaitTimeout)
	defer cancel()

	sp := NewSpinner(streams.Stderr, fmt.Sprintf("Waiting for %s on %s", result.Role, result.ScopeName))
	sp.Start()

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
			sp.Stop()
			message := "Activation was accepted, but Azure did not report it active within 60s. " +
				"Run pimctl status to check again.\n"
			if errors.Is(deadline.Err(), context.Canceled) {
				message = "Activation wait canceled. Run pimctl status to check whether Azure finished it.\n"
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
			for _, a := range as {
				if domain.ActiveAssignmentConfirmsResult(a, result) {
					sp.Stop()
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
