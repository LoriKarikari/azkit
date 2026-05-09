package cli

import (
	"context"
	"io"
)

type StatusCmd struct {
	JSON     bool `help:"Output as JSON"`
	Extended bool `help:"Show more details"`
}

func (c *StatusCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	service, err := services.Status(streams.Log)
	if err != nil {
		return err
	}
	as, err := service.Status(ctx)
	if err != nil {
		return err
	}
	if c.JSON {
		_, err = io.WriteString(streams.Stdout, RenderStatusJSON(as))
		return err
	}
	_, err = io.WriteString(streams.Stdout, RenderStatusHuman(as, c.Extended))
	return err
}
