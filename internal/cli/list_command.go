package cli

import (
	"context"
	"io"
)

type ListCmd struct {
	JSON    bool `help:"Output as JSON"`
	Verbose bool `help:"Show resource IDs and assignment IDs"`
}

func (c *ListCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	service, err := services.List()
	if err != nil {
		return err
	}
	as, err := service.List(ctx)
	if err != nil {
		return err
	}

	if c.JSON {
		_, err = io.WriteString(streams.Stdout, RenderJSON(as))
		return err
	}
	_, err = io.WriteString(streams.Stdout, RenderHuman(as, c.Verbose))
	return err
}
