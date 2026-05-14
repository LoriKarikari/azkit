package cli

import (
	"context"
	"io"
	"time"
)

type DeactivateCmd struct {
	AssignmentID string `arg:"" name:"assignment-id" help:"Active assignment ID to deactivate"`
	Reason       string `help:"Justification for the deactivation"`
	JSON         bool   `help:"Output as JSON"`
}

func (c *DeactivateCmd) Run(ctx context.Context, services Services, streams *Streams) error {
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
