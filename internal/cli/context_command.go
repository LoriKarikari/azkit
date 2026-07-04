package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/charmbracelet/huh"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/config"
	"github.com/LoriKarikari/azkit/internal/contextstore"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/interactive"
)

const activeContextEnv = "AZKIT_CONTEXT"

type CtxCmd struct {
	List        bool       `short:"l" name:"list" help:"List saved contexts"`
	ListDefault CtxListCmd `cmd:"" default:"1" hidden:""`
	Add         CtxAddCmd  `cmd:"" help:"Add a tenant context"`
	Rm          CtxRmCmd   `cmd:"" name:"rm" help:"Remove a tenant context"`
}

type CtxListCmd struct{}

type CtxAddCmd struct {
	Name   string `arg:"" help:"Context name"`
	Tenant string `help:"Azure tenant ID"`
}

type CtxRmCmd struct {
	Name  string `arg:"" help:"Context name"`
	Force bool   `help:"Remove without confirmation and allow removing the active context"`
}

func (c *CtxListCmd) Run(ctx context.Context, streams *Streams) error {
	return renderContextList(ctx, streams)
}

func (c *CtxAddCmd) Run(ctx context.Context, streams *Streams) error {
	svc, err := contextService(streams)
	if err != nil {
		return err
	}
	tenantID := c.Tenant
	if tenantID == "" {
		tenantID = os.Getenv("AZURE_TENANT_ID")
	}
	added, err := svc.Add(ctx, c.Name, tenantID)
	if err != nil {
		return err
	}
	_, err = fmt.Fprintf(streams.Stdout, "Added context %s for tenant %s\n", added.Name, added.TenantID)
	return err
}

func (c *CtxRmCmd) Run(ctx context.Context, streams *Streams) error {
	svc, err := contextService(streams)
	if err != nil {
		return err
	}
	if !c.Force && os.Getenv(activeContextEnv) == c.Name {
		return svc.Remove(ctx, c.Name, false)
	}
	if !c.Force && !interactive.IsTerminalFn() {
		return app.ContextRemovalNeedsForce(c.Name)
	}
	if !c.Force {
		confirmed, err := confirmContextRemoval(ctx, c.Name)
		if err != nil {
			return err
		}
		if !confirmed {
			return interactive.ErrCanceled
		}
	}
	if err := svc.Remove(ctx, c.Name, c.Force); err != nil {
		return err
	}
	_, err = fmt.Fprintf(streams.Stdout, "Removed context %s\n", c.Name)
	return err
}

func renderContextList(ctx context.Context, streams *Streams) error {
	svc, err := contextService(streams)
	if err != nil {
		return err
	}
	contexts, err := svc.List(ctx)
	if err != nil {
		return err
	}
	_, err = io.WriteString(streams.Stdout, renderContextsHuman(contexts))
	return err
}

func renderContextsHuman(contexts []domain.TenantContext) string {
	if len(contexts) == 0 {
		return "No contexts.\n"
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "NAME\tTENANT\tSTATUS")
	for _, item := range contexts {
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\n", item.Name, item.TenantID, item.Status)
	}
	_ = w.Flush()
	return buf.String()
}

func contextService(streams *Streams) (*app.ContextService, error) {
	configDir, err := config.ConfigDir(streams.ConfigPath)
	if err != nil {
		return nil, err
	}
	stateDir, err := config.StateDir()
	if err != nil {
		return nil, err
	}
	store := contextstore.New(configDir, stateDir)
	return app.NewContextService(store, func() string { return os.Getenv(activeContextEnv) }), nil
}

func confirmContextRemoval(ctx context.Context, name string) (bool, error) {
	confirmed := false
	form := huh.NewForm(huh.NewGroup(
		huh.NewConfirm().
			Title("Remove this context?").
			Description(fmt.Sprintf("This deletes the saved context and its credential cache: %s", name)).
			Affirmative("Remove").
			Negative("Cancel").
			Value(&confirmed),
	))
	if err := form.RunWithContext(ctx); err != nil {
		if errors.Is(err, huh.ErrUserAborted) {
			return false, interactive.ErrCanceled
		}
		return false, err
	}
	return confirmed, nil
}
