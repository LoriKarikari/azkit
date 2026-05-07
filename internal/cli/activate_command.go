package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type ActivateCmd struct {
	Scope         string        `help:"Azure resource scope ID (exact)"`
	Subscription  string        `help:"Subscription ID or exact name"`
	ResourceGroup string        `help:"Resource group name (requires --subscription)"`
	Role          string        `required:"" help:"Role display name or definition ID"`
	Reason        string        `required:"" help:"Justification for the activation"`
	Duration      time.Duration `default:"2h" help:"How long the role stays active"`
	WaitTimeout   time.Duration `default:"60s" help:"How long to wait for activation to propagate"`
	JSON          bool          `help:"Output as JSON"`
}

func (c *ActivateCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	act, err := services.Activate()
	if err != nil {
		return err
	}
	result, err := act.Activate(ctx, domain.ActivationRequest{
		ScopeID:       c.Scope,
		Subscription:  c.Subscription,
		ResourceGroup: c.ResourceGroup,
		Role:          c.Role,
		Reason:        c.Reason,
		Duration:      c.Duration,
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
		_, _ = io.WriteString(streams.Stderr, fmt.Sprintf("Activating %s on %s for %s\n", result.Role, result.ScopeName, result.Duration))
		_, err = io.WriteString(streams.Stdout, renderActivationHuman(result))
	}
	return err
}

func (c *ActivateCmd) waitForActive(ctx context.Context, services Services, result *domain.ActivationResult, streams *Streams) *domain.ActivationResult {
	statusSvc, err := services.Status()
	if err != nil {
		return nil
	}

	timeout := c.WaitTimeout
	if timeout <= 0 {
		timeout = 60 * time.Second
	}
	deadline, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	_, _ = io.WriteString(streams.Stderr, "Waiting for activation to propagate...")

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-deadline.Done():
			_, _ = io.WriteString(streams.Stderr, " timeout.\n")
			return nil
		case <-ticker.C:
			as, err := statusSvc.Status(ctx)
			if err != nil {
				_, _ = io.WriteString(streams.Stderr, ".")
				continue
			}
			for _, a := range as {
				if a.Role == result.Role && a.ScopeID == result.ScopeID && a.Status == domain.ActiveAssignmentActive {
					_, _ = io.WriteString(streams.Stderr, " done.\n")
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
			_, _ = io.WriteString(streams.Stderr, ".")
		}
	}
}
