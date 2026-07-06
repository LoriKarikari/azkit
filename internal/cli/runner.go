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
	Stdout     io.Writer
	Stderr     io.Writer
	Log        *slog.Logger
	Config     *config.Config
	ConfigPath string
	ShellEnv   bool
	Shell      string
}

type activateInteractiveFunc func(context.Context, []domain.EligibleAssignment, *app.ActivationService, *config.Config, interactive.ActivationInput) (*domain.ActivationResult, error)

type deactivateInteractiveFunc func(context.Context, []domain.ActiveAssignment, *app.DeactivationService, interactive.DeactivationInput) (*domain.DeactivationResult, error)

type subscriptionPickerFunc func(context.Context, []domain.Subscription) (domain.Subscription, error)

type jsonCommand interface {
	jsonOutput() bool
}

type Services struct {
	List                  func(*slog.Logger) (*app.ListService, error)
	Status                func(*slog.Logger) (*app.StatusService, error)
	Activate              func(*slog.Logger) (*app.ActivationService, error)
	Deactivate            func(*slog.Logger) (*app.DeactivationService, error)
	Subscriptions         func(*slog.Logger) (*app.SubscriptionService, error)
	ActivateInteractive   activateInteractiveFunc
	DeactivateInteractive deactivateInteractiveFunc
	PickContext           contextPickerFunc
	PickSubscription      subscriptionPickerFunc
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
	if services.PickContext == nil {
		services.PickContext = interactive.PickContext
	}
	if services.PickSubscription == nil {
		services.PickSubscription = interactive.PickSubscription
	}
	return &Runner{
		services: services,
		streams:  &Streams{Stdout: stdout, Stderr: stderr},
	}
}

type CLI struct {
	Verbose     bool             `short:"v" help:"Enable debug logging to stderr"`
	Output      string           `short:"o" enum:"table,json" default:"table" help:"Output format: table or json"`
	VersionFlag kong.VersionFlag `name:"version" help:"Show version information and exit"`
	ShellEnv    bool             `name:"shell-env" hidden:"" help:"Emit shell environment changes for shell integration"`
	ConfigPath  string           `name:"config" help:"Path to config file"`
	Pim         pimCmd           `cmd:"" help:"Manage Azure resource-role PIM workflows"`
	Ctx         CtxCmd           `cmd:"" name:"ctx" help:"Manage tenant contexts"`
	Sub         SubCmd           `cmd:"" name:"sub" help:"Manage subscriptions in the active context"`
	ShellInit   ShellInitCmd     `cmd:"" name:"shell-init" help:"Print shell integration script"`
	Completion  CompletionCmd    `cmd:"" help:"Generate shell completion script"`
	Version     VersionCmd       `cmd:"" help:"Show version information"`
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
		kong.Vars{"version": versionLine()},
		kong.Help(azkitHelp),
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
	r.streams.ConfigPath = model.ConfigPath
	r.streams.ShellEnv = model.ShellEnv
	r.streams.Shell = os.Getenv("AZKIT_SHELL")
	if model.Output == "json" {
		forceJSONOutput(&model)
	}

	if commandNeedsConfig(parsed) {
		cfg, err := config.Load(model.ConfigPath)
		if err != nil {
			_, _ = io.WriteString(r.streams.Stderr, RenderError(err, wantsJSON(parsed)))
			return 1
		}
		r.streams.Config = cfg
	}

	if err := parsed.Run(); err != nil {
		if errors.Is(err, interactive.ErrCanceled) {
			return 130
		}
		_, _ = io.WriteString(r.streams.Stderr, RenderError(err, wantsJSON(parsed)))
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

func azkitHelp(options kong.HelpOptions, ctx *kong.Context) error {
	if ctx.Selected() == nil {
		options.Compact = true
		options.NoExpandSubcommands = true
		return kong.DefaultHelpPrinter(options, ctx)
	}

	options.Tree = true
	return kong.DefaultHelpPrinter(options, ctx)
}

func commandNeedsConfig(parsed *kong.Context) bool {
	command := strings.Fields(parsed.Command())
	if len(command) == 0 {
		return true
	}
	return command[0] == "pim"
}

func forceJSONOutput(model *CLI) {
	model.Pim.List.JSON = true
	model.Pim.Status.JSON = true
	model.Pim.Activate.JSON = true
	model.Pim.Deactivate.JSON = true
	model.Ctx.Switch.JSON = true
	model.Ctx.Current.JSON = true
	model.Sub.Switch.JSON = true
	model.Sub.Current.JSON = true
}

func wantsJSON(parsed *kong.Context) bool {
	selected := parsed.Selected()
	if selected == nil || !selected.Target.IsValid() {
		return false
	}
	if selected.Target.CanInterface() {
		cmd, ok := selected.Target.Interface().(jsonCommand)
		if ok {
			return cmd.jsonOutput()
		}
	}
	if selected.Target.CanAddr() {
		cmd, ok := selected.Target.Addr().Interface().(jsonCommand)
		return ok && cmd.jsonOutput()
	}
	return false
}
