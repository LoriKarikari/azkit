package cli

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"

	"github.com/alecthomas/kong"
	"github.com/willabides/kongplete"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/config"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/interactive"
)

type Streams struct {
	Stdout io.Writer
	Stderr io.Writer
	Log    *slog.Logger
	Config *config.Config
}

type activateInteractiveFunc func(context.Context, []domain.EligibleAssignment, *app.ActivationService, *config.Config, interactive.ActivationInput) (*domain.ActivationResult, error)

type deactivateInteractiveFunc func(context.Context, []domain.ActiveAssignment, *app.DeactivationService, interactive.DeactivationInput) (*domain.DeactivationResult, error)

type Services struct {
	List                  func(*slog.Logger) (*app.ListService, error)
	Status                func(*slog.Logger) (*app.StatusService, error)
	Activate              func(*slog.Logger) (*app.ActivationService, error)
	Deactivate            func(*slog.Logger) (*app.DeactivationService, error)
	ActivateInteractive   activateInteractiveFunc
	DeactivateInteractive deactivateInteractiveFunc
}

type Runner struct {
	services Services
	streams  *Streams
	log      *slog.Logger
}

func NewRunner(services Services, stdout io.Writer, stderr io.Writer) *Runner {
	if services.ActivateInteractive == nil {
		services.ActivateInteractive = interactive.Activate
	}
	if services.DeactivateInteractive == nil {
		services.DeactivateInteractive = interactive.Deactivate
	}
	return &Runner{
		services: services,
		streams:  &Streams{Stdout: stdout, Stderr: stderr},
	}
}

type CLI struct {
	Verbose    bool          `short:"v" help:"Enable debug logging to stderr"`
	ConfigPath string        `name:"config" help:"Path to config file"`
	Pim        pimCmd        `cmd:"" help:"Manage Azure resource-role PIM workflows"`
	Completion CompletionCmd `cmd:"" help:"Generate shell completion script"`
	Version    VersionCmd    `cmd:"" help:"Show version information"`
}

type pimCmd struct {
	List       ListCmd       `cmd:"" help:"List eligible PIM role assignments"`
	Status     StatusCmd     `cmd:"" help:"List active PIM role assignments"`
	Activate   ActivateCmd   `cmd:"" help:"Activate an eligible PIM role assignment"`
	Deactivate DeactivateCmd `cmd:"" help:"Deactivate an active PIM role assignment"`
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
		kong.Name("azkit"),
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
	if os.Getenv("COMP_LINE") != "" {
		kongplete.Complete(parser, kongplete.WithExitFunc(func(code int) { panic(kongExit(code)) }))
	}

	parsed, err := parser.Parse(args)
	if err != nil {
		return r.handleParseError(err)
	}

	level := slog.LevelWarn
	if model.Verbose {
		level = slog.LevelDebug
	}
	r.log = slog.New(slog.NewTextHandler(r.streams.Stderr, &slog.HandlerOptions{Level: level}))
	r.streams.Log = r.log

	if commandNeedsConfig(parsed) {
		cfg, err := config.Load(model.ConfigPath)
		if err != nil {
			_, _ = io.WriteString(r.streams.Stderr, RenderError(err, wantsJSON(model, parsed)))
			return 1
		}
		r.streams.Config = cfg
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
		if code == 0 {
			return 0
		}
		_, _ = io.WriteString(r.streams.Stderr, err.Error()+"\n")
		return 2
	}
	_, _ = io.WriteString(r.streams.Stderr, err.Error()+"\n")
	return 2
}

func commandNeedsConfig(parsed *kong.Context) bool {
	command := strings.Fields(parsed.Command())
	if len(command) == 0 {
		return true
	}
	switch command[0] {
	case "completion", "version":
		return false
	}
	return true
}

func wantsJSON(model CLI, parsed *kong.Context) bool {
	switch parsed.Command() {
	case "pim list":
		return model.Pim.List.JSON
	case "pim status":
		return model.Pim.Status.JSON
	case "pim activate":
		return model.Pim.Activate.JSON
	case "pim deactivate":
		return model.Pim.Deactivate.JSON
	}
	return false
}
