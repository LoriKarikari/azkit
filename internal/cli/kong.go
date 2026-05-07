package cli

import (
	"context"
	"io"

	"github.com/alecthomas/kong"

	"github.com/LoriKarikari/pimctl/internal/app"
)

type Streams struct {
	Stdout io.Writer
	Stderr io.Writer
}

type Runner struct {
	service *app.ListService
	streams *Streams
}

func NewRunner(service *app.ListService, stdout io.Writer, stderr io.Writer) *Runner {
	return &Runner{
		service: service,
		streams: &Streams{Stdout: stdout, Stderr: stderr},
	}
}

func (r *Runner) Run(ctx context.Context, args []string) int {
	model := CLI{}
	parser, err := kong.New(
		&model,
		kong.Name("pimctl"),
		kong.Writers(r.streams.Stdout, r.streams.Stderr),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.Bind(r.service),
		kong.Bind(r.streams),
	)
	if err != nil {
		_, _ = io.WriteString(r.streams.Stderr, RenderError(err, false))
		return 1
	}
	parsed, err := parser.Parse(args)
	if err != nil {
		_, _ = io.WriteString(r.streams.Stderr, err.Error()+"\n")
		return 1
	}
	if err := parsed.Run(); err != nil {
		_, _ = io.WriteString(r.streams.Stderr, RenderError(err, wantsJSON(model, parsed)))
		return 1
	}
	return 0
}

func wantsJSON(model CLI, parsed *kong.Context) bool {
	return parsed.Command() == "list" && model.List.JSON
}

type ListCmd struct {
	JSON    bool `help:"Output as JSON"`
	Verbose bool `help:"Show resource IDs and assignment IDs"`
}

func (c *ListCmd) Run(ctx context.Context, service *app.ListService, streams *Streams) error {
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

type CLI struct {
	List ListCmd `cmd:"" help:"List eligible PIM role assignments"`
}
