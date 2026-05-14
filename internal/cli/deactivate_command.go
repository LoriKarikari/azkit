package cli

import (
	"context"
	"io"
	"time"

	"github.com/LoriKarikari/pimctl/internal/app"
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
			Code:    app.CodeActiveAssignmentNotFound,
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
	statusSvc, err := services.Status(streams.Log)
	if err != nil {
		return err
	}
	active, err := statusSvc.Status(ctx)
	if err != nil {
		return err
	}

	deactivator, err := services.Deactivate(streams.Log)
	if err != nil {
		return err
	}

	result, err := interactive.Deactivate(ctx, active, deactivator, c.Reason, false)
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
