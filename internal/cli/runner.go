package cli

import (
	"context"
	"errors"
	"io"

	"github.com/alecthomas/kong"

	"github.com/LoriKarikari/pimctl/internal/app"
)

type Streams struct {
	Stdout io.Writer
	Stderr io.Writer
}

type Services struct {
	List     func() (*app.ListService, error)
	Activate func() (*app.ActivationService, error)
}

type Runner struct {
	services Services
	streams  *Streams
}

func NewRunner(services Services, stdout io.Writer, stderr io.Writer) *Runner {
	return &Runner{
		services: services,
		streams:  &Streams{Stdout: stdout, Stderr: stderr},
	}
}

type kongExit int

func (r *Runner) Run(ctx context.Context, args []string) (code int) {
	defer func() {
		if value := recover(); value != nil {
			if exit, ok := value.(kongExit); ok {
				code = int(exit)
				return
			}
			panic(value)
		}
	}()

	model := CLI{}
	parser, err := kong.New(
		&model,
		kong.Name("pimctl"),
		kong.Writers(r.streams.Stdout, r.streams.Stderr),
		kong.Exit(func(code int) { panic(kongExit(code)) }),
		kong.BindTo(ctx, (*context.Context)(nil)),
		kong.Bind(r.services),
		kong.Bind(r.streams),
	)
	if err != nil {
		_, _ = io.WriteString(r.streams.Stderr, RenderError(err, false))
		return 1
	}
	parsed, err := parser.Parse(args)
	if err != nil {
		return r.handleParseError(err)
	}
	if err := parsed.Run(); err != nil {
		_, _ = io.WriteString(r.streams.Stderr, RenderError(err, wantsJSON(model, parsed)))
		return 1
	}
	return 0
}

func (r *Runner) handleParseError(err error) int {
	var exitCoder kong.ExitCoder
	if errors.As(err, &exitCoder) {
		code := exitCoder.ExitCode()
		if code != 0 {
			_, _ = io.WriteString(r.streams.Stderr, err.Error()+"\n")
		}
		return code
	}
	_, _ = io.WriteString(r.streams.Stderr, err.Error()+"\n")
	return 1
}

func wantsJSON(model CLI, parsed *kong.Context) bool {
	switch parsed.Command() {
	case "list":
		return model.List.JSON
	case "activate":
		return model.Activate.JSON
	}
	return false
}
