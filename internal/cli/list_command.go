package cli

import (
	"context"
	"io"
	"log/slog"
)

type ListCmd struct {
	JSON     bool `help:"Output as JSON"`
	Extended bool `help:"Show resource IDs and assignment IDs"`
}

func (c *ListCmd) jsonOutput() bool {
	return c.JSON
}

func (c *ListCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	service, err := services.List(streams.Log)
	if err != nil {
		return err
	}
	as, err := service.List(ctx)
	if err != nil {
		return err
	}
	streams.Log.Debug("listed eligible assignments", slog.Int("count", len(as)))

	if c.JSON {
		_, err = io.WriteString(streams.Stdout, RenderJSON(as))
		return err
	}
	_, err = io.WriteString(streams.Stdout, RenderHuman(as, c.Extended))
	return err
}
