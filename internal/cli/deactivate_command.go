package cli

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
	"github.com/LoriKarikari/pimctl/internal/interactive"
)

type DeactivateCmd struct {
	AssignmentID string `arg:"" name:"assignment-id" help:"Active assignment ID to deactivate" optional:""`
	Reason       string `help:"Justification for the deactivation"`
	JSON         bool   `help:"Output as JSON"`
}

func (c *DeactivateCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	if c.AssignmentID == "" && interactive.IsTerminalFn() {
		return c.runInteractive(ctx, services, streams)
	}

	if c.AssignmentID == "" {
		return &app.Error{
			Code:    domain.CodeActiveAssignmentNotFound,
			Message: "Assignment ID is required. Run pimctl status --verbose to find it.",
		}
	}

	svc, err := services.Deactivate(streams.Log)
	if err != nil {
		return err
	}

	deactivateCtx, cancel := context.WithTimeout(ctx, 60*time.Second)
	defer cancel()

	result, err := svc.Deactivate(deactivateCtx, c.AssignmentID, c.Reason)
	if err != nil {
		return err
	}

	if c.JSON {
		_, err = io.WriteString(streams.Stdout, renderDeactivationJSON(result))
		return err
	}

	_, err = io.WriteString(streams.Stdout, renderDeactivationHuman(result))
	return err
}

func (c *DeactivateCmd) runInteractive(ctx context.Context, services Services, streams *Streams) error {
	if streams.Config != nil && streams.Config.TenantID != "" {
		_, _ = io.WriteString(streams.Stderr, "Tenant: "+streams.Config.TenantID+"\n")
	}

	statusSvc, err := services.Status(streams.Log)
	if err != nil {
		return err
	}

	var active []domain.ActiveAssignment
	{
		sp := interactive.NewSpinner(streams.Stderr, "Loading active assignments")
		if !c.JSON {
			sp.Start()
		}
		active, err = statusSvc.Status(ctx)
		if !c.JSON {
			sp.Stop()
		}
		if err != nil {
			return err
		}
	}

	deactivator, err := services.Deactivate(streams.Log)
	if err != nil {
		return err
	}

	input := interactive.DeactivationInput{
		Reason: c.Reason,
	}
	if !c.JSON {
		input.Progress = streams.Stderr
	}

	result, err := services.DeactivateInteractive(ctx, active, deactivator, input)
	if err != nil {
		if errors.Is(err, interactive.ErrCanceled) {
			_, _ = io.WriteString(streams.Stderr, "Deactivation canceled.\n")
			return nil
		}
		return err
	}

	_, _ = io.WriteString(streams.Stderr, "\r\033[K")

	confirmed := waitForDeactivation(ctx, services, result, streams)
	if confirmed != nil {
		result = confirmed
	}

	if c.JSON {
		_, err = io.WriteString(streams.Stdout, renderDeactivationJSON(result))
		return err
	}

	_, err = io.WriteString(streams.Stdout, renderDeactivationHuman(result))
	return err
}

func waitForDeactivation(
	ctx context.Context,
	services Services,
	result *domain.DeactivationResult,
	streams *Streams,
) *domain.DeactivationResult {
	statusSvc, err := services.Status(streams.Log)
	if err != nil {
		return nil
	}

	deadline, cancel := context.WithTimeout(ctx, defaultWaitTimeout)
	defer cancel()

	sp := interactive.NewSpinner(streams.Stderr, fmt.Sprintf("Deactivating %s on %s", result.Role, result.ScopeName))
	sp.Start()

	streams.Log.Debug(
		"waiting for deactivation to propagate",
		slog.String("role", result.Role),
		slog.String("scope", result.ScopeID),
		slog.String("assignment", result.AssignmentID),
	)

	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-deadline.Done():
			sp.Stop()
			message := "Deactivation was accepted, but Azure did not report it completed within 60s. " +
				"Run pimctl status to check again.\n"
			if errors.Is(deadline.Err(), context.Canceled) {
				message = "Deactivation wait canceled. Run pimctl status to check whether Azure finished it.\n"
			}
			_, _ = io.WriteString(streams.Stderr, message)
			return nil
		case <-ticker.C:
			as, err := statusSvc.StatusForScope(deadline, result.ScopeID)
			if err != nil {
				streams.Log.Debug(
					"deactivation status poll failed",
					slog.String("scope", result.ScopeID),
					slog.Any("error", err),
				)
				continue
			}
			streams.Log.Debug(
				"deactivation poll returned",
				slog.Int("active_assignments", len(as)),
				slog.String("target_assignment", result.AssignmentID),
			)
			found := false
			for _, a := range as {
				if a.ID == result.AssignmentID {
					found = true
					break
				}
			}
			if !found {
				sp.Stop()
				confirmedMsg := fmt.Sprintf(
					"✓ Deactivated %s on %s\n",
					result.Role,
					result.ScopeName,
				)
				_, _ = io.WriteString(streams.Stderr, confirmedMsg)
				confirmed := *result
				confirmed.Status = domain.DeactivationConfirmed
				return &confirmed
			}
		}
	}
}
