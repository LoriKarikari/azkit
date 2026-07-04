package cli

import (
	"bytes"
	"context"
	"encoding/json"
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

const (
	activeContextEnv          = "AZKIT_CONTEXT"
	previousContextEnv        = "AZKIT_PREVIOUS_CONTEXT"
	activeSubscriptionEnv     = "AZKIT_SUBSCRIPTION_ID"
	previousSubscriptionEnv   = "AZKIT_PREVIOUS_SUBSCRIPTION_ID"
	azureTenantEnv            = "AZURE_TENANT_ID"
	azureConfigDirEnv         = "AZURE_CONFIG_DIR"
	azureSubscriptionEnv      = "AZURE_SUBSCRIPTION_ID"
	terraformTenantEnv        = "ARM_TENANT_ID"
	terraformSubscriptionEnv  = "ARM_SUBSCRIPTION_ID"
	terraformSubscriptionName = "ARM_SUBSCRIPTION_NAME"
)

type contextPickerFunc func(context.Context, []domain.TenantContext) (domain.TenantContext, error)

type CtxCmd struct {
	Switch  CtxSwitchCmd  `cmd:"" default:"withargs" hidden:""`
	Add     CtxAddCmd     `cmd:"" help:"Add a tenant context"`
	Rm      CtxRmCmd      `cmd:"" name:"rm" help:"Remove a tenant context"`
	Current CtxCurrentCmd `cmd:"" help:"Show the active tenant context"`
}

type CtxSwitchCmd struct {
	List bool   `short:"l" name:"list" help:"List saved contexts"`
	Name string `arg:"" optional:"" help:"Context name, or '-' for the previous context"`
}

type CtxAddCmd struct {
	Name   string `arg:"" help:"Context name"`
	Tenant string `help:"Azure tenant ID"`
}

type CtxRmCmd struct {
	Name  string `arg:"" help:"Context name"`
	Force bool   `help:"Remove without confirmation and allow removing the active context"`
}

type CtxCurrentCmd struct {
	JSON bool `help:"Output as JSON"`
}

func (c *CtxCurrentCmd) jsonOutput() bool {
	return c.JSON
}

func (c *CtxSwitchCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	if c.List {
		return renderContextList(ctx, streams)
	}
	svc, err := contextService(streams)
	if err != nil {
		return err
	}
	name := c.Name
	if name == "" {
		if !interactive.IsTerminalFn() {
			return app.MissingContextName()
		}
		contexts, err := svc.List(ctx)
		if err != nil {
			return err
		}
		selected, err := services.PickContext(ctx, contexts)
		if err != nil {
			return err
		}
		name = selected.Name
	}
	if name == "-" {
		previous := os.Getenv(previousContextEnv)
		if previous == "" {
			return app.PreviousContextNotFound()
		}
		name = previous
	}
	target, err := svc.Get(ctx, name)
	if err != nil {
		return err
	}
	return switchContext(streams, target)
}

func (c *CtxCurrentCmd) Run(ctx context.Context, streams *Streams) error {
	svc, err := contextService(streams)
	if err != nil {
		return err
	}
	current, ok, err := svc.Current(ctx)
	if err != nil {
		return err
	}
	if c.JSON {
		_, err = io.WriteString(streams.Stdout, renderCurrentContextJSON(current, ok))
		return err
	}
	_, err = io.WriteString(streams.Stdout, renderCurrentContextHuman(current, ok))
	return err
}

func (c *CtxAddCmd) Run(ctx context.Context, streams *Streams) error {
	svc, err := contextService(streams)
	if err != nil {
		return err
	}
	tenantID := c.Tenant
	if tenantID == "" {
		tenantID = os.Getenv(azureTenantEnv)
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
		return app.ActiveContextRemoval(c.Name)
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
	_, err = io.WriteString(streams.Stdout, renderContextsHuman(contexts, os.Getenv(activeContextEnv)))
	return err
}

func switchContext(streams *Streams, target domain.TenantContext) error {
	if err := streams.RequireShellIntegration("azkit ctx " + target.Name); err != nil {
		return err
	}
	current := os.Getenv(activeContextEnv)
	changes := []ShellEnvChange{
		{Name: activeContextEnv, Value: target.Name},
		{Name: azureTenantEnv, Value: target.TenantID},
		{Name: terraformTenantEnv, Value: target.TenantID},
		{Name: azureConfigDirEnv, Value: target.Dir},
		{Name: activeSubscriptionEnv, Unset: true},
		{Name: previousSubscriptionEnv, Unset: true},
		{Name: azureSubscriptionEnv, Unset: true},
		{Name: terraformSubscriptionEnv, Unset: true},
		{Name: terraformSubscriptionName, Unset: true},
	}
	if current != "" && current != target.Name {
		changes = append([]ShellEnvChange{{Name: previousContextEnv, Value: current}}, changes...)
	}
	script, err := streams.RenderShellEnv(changes)
	if err != nil {
		return err
	}
	if target.Status != domain.ContextReady {
		_, _ = fmt.Fprintf(streams.Stderr, "Run: az login --tenant %s\n", target.TenantID)
	}
	_, err = io.WriteString(streams.Stdout, script)
	return err
}

func renderContextsHuman(contexts []domain.TenantContext, active string) string {
	if len(contexts) == 0 {
		return "No contexts.\n"
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "CURRENT\tNAME\tTENANT\tSTATUS")
	for _, item := range contexts {
		marker := ""
		if item.Name == active {
			marker = "*"
		}
		_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", marker, item.Name, item.TenantID, item.Status)
	}
	_ = w.Flush()
	return buf.String()
}

type currentContextJSON struct {
	Context   string `json:"context"`
	TenantID  string `json:"tenant_id"`
	ConfigDir string `json:"config_dir"`
	Status    string `json:"status"`
}

func renderCurrentContextHuman(current domain.TenantContext, ok bool) string {
	if !ok {
		return "No active context.\n"
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "Context:\t%s\n", current.Name)
	_, _ = fmt.Fprintf(w, "Tenant:\t%s\n", current.TenantID)
	_, _ = fmt.Fprintf(w, "Status:\t%s\n", current.Status)
	_, _ = fmt.Fprintf(w, "Config dir:\t%s\n", current.Dir)
	_ = w.Flush()
	return buf.String()
}

func renderCurrentContextJSON(current domain.TenantContext, ok bool) string {
	out := currentContextJSON{}
	if ok {
		out = currentContextJSON{
			Context:   current.Name,
			TenantID:  current.TenantID,
			ConfigDir: current.Dir,
			Status:    string(current.Status),
		}
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return string(b) + "\n"
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
