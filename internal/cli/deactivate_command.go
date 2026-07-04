package cli

import (
	"context"
	"errors"
	"io"
	"time"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/interactive"
)

type DeactivateCmd struct {
	AssignmentID string `arg:"" name:"assignment-id" help:"Active assignment ID to deactivate" optional:""`
	Reason       string `help:"Justification for the deactivation"`
	JSON         bool   `help:"Output as JSON"`
}

func (c *DeactivateCmd) jsonOutput() bool {
	return c.JSON
}

func (c *DeactivateCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	if c.AssignmentID == "" && !c.JSON && interactive.IsTerminalFn() {
		return c.runInteractive(ctx, services, streams)
	}

	if c.AssignmentID == "" {
		return &app.Error{
			Code:    domain.CodeActiveAssignmentNotFound,
			Message: "Assignment ID is required. Run azkit pim status --extended to find it.",
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
		active, err = interactive.WithSpinner(
			ctx,
			streams.Stderr,
			"Loading active assignments",
			!c.JSON,
			func(ctx context.Context) ([]domain.ActiveAssignment, error) {
				return statusSvc.Status(ctx)
			},
		)
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

	interactive.ClearProgress(streams.Stderr)

	if c.JSON {
		_, err = io.WriteString(streams.Stdout, renderDeactivationJSON(result))
		return err
	}

	_, err = io.WriteString(streams.Stdout, renderDeactivationHuman(result))
	return err
}
