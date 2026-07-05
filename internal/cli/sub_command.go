package cli

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"os"
	"text/tabwriter"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/interactive"
)

type SubCmd struct {
	Switch  SubSwitchCmd  `cmd:"" default:"withargs" hidden:""`
	Alias   SubAliasCmd   `cmd:"" help:"Create a subscription alias"`
	Unalias SubUnaliasCmd `cmd:"" name:"unalias" help:"Remove a subscription alias"`
	Current SubCurrentCmd `cmd:"" help:"Show the active subscription"`
}

type SubSwitchCmd struct {
	List    bool   `short:"l" name:"list" help:"List subscriptions for the active context"`
	Refresh bool   `help:"Refresh the active context subscription cache"`
	JSON    bool   `help:"Output as JSON"`
	Target  string `arg:"" optional:"" help:"Alias, subscription ID, or exact subscription name; '-' for the previous subscription"`
}

type SubAliasCmd struct {
	Alias    string `arg:"" help:"Alias name"`
	Selector string `arg:"" help:"Subscription ID or exact subscription name"`
}

type SubUnaliasCmd struct {
	Alias string `arg:"" help:"Alias name"`
}

type SubCurrentCmd struct {
	JSON bool `help:"Output as JSON"`
}

func (c *SubCurrentCmd) jsonOutput() bool {
	return c.JSON
}

func (c *SubSwitchCmd) jsonOutput() bool {
	return c.JSON
}

func (c *SubSwitchCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	if c.Target == "" && !c.List && !c.Refresh && !interactive.IsTerminalFn() {
		return app.MissingSubscriptionCommand()
	}
	if c.Target != "" && (c.List || c.Refresh) {
		return app.ConflictingSubscriptionSelectors()
	}
	if c.JSON && !c.List && !c.Refresh {
		return app.JSONOutputNotSupported("azkit sub")
	}
	if c.List || c.Refresh {
		return runSubscriptionList(ctx, services, streams, c.Refresh, c.JSON)
	}
	active, err := activeTenantContext(ctx, streams)
	if err != nil {
		return err
	}
	svc, err := services.Subscriptions(streams.Log)
	if err != nil {
		return err
	}
	if c.Target == "" {
		if err := streams.RequireShellIntegration("azkit sub"); err != nil {
			return err
		}
		subscriptions, err := svc.List(ctx, active, false)
		if err != nil {
			return err
		}
		selected, err := services.PickSubscription(ctx, subscriptions)
		if err != nil {
			return err
		}
		return switchSubscription(streams, selected)
	}
	target := c.Target
	if target == "-" {
		previous := os.Getenv(previousSubscriptionEnv)
		if previous == "" {
			return app.PreviousSubscriptionNotFound()
		}
		target = previous
	}
	sub, err := svc.Resolve(ctx, active, target)
	if err != nil {
		return err
	}
	return switchSubscription(streams, sub)
}

func (c *SubAliasCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	active, err := activeTenantContext(ctx, streams)
	if err != nil {
		return err
	}
	svc, err := services.Subscriptions(streams.Log)
	if err != nil {
		return err
	}
	if c.Alias == "" {
		return app.MissingAliasName()
	}
	if c.Selector == "" {
		return app.MissingAliasSelector()
	}
	if err := svc.CreateAlias(ctx, active, c.Alias, c.Selector); err != nil {
		return err
	}
	_, err = fmt.Fprintf(streams.Stdout, "Added alias %s\n", c.Alias)
	return err
}

func (c *SubUnaliasCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	active, err := activeTenantContext(ctx, streams)
	if err != nil {
		return err
	}
	svc, err := services.Subscriptions(streams.Log)
	if err != nil {
		return err
	}
	if c.Alias == "" {
		return app.MissingAliasName()
	}
	if err := svc.RemoveAlias(ctx, active, c.Alias); err != nil {
		return err
	}
	_, err = fmt.Fprintf(streams.Stdout, "Removed alias %s\n", c.Alias)
	return err
}

func (c *SubCurrentCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	active, err := activeTenantContext(ctx, streams)
	if err != nil {
		return err
	}
	id := os.Getenv(activeSubscriptionEnv)
	name := ""
	if id != "" {
		svc, err := services.Subscriptions(streams.Log)
		if err != nil {
			return err
		}
		subs, err := svc.List(ctx, active, false)
		if err == nil {
			for _, sub := range subs {
				if sub.ID == id {
					name = sub.Name
					break
				}
			}
		}
	}
	if c.JSON {
		_, err = io.WriteString(streams.Stdout, renderCurrentSubscriptionJSON(active, id, name))
		return err
	}
	_, err = io.WriteString(streams.Stdout, renderCurrentSubscriptionHuman(id, name))
	return err
}

func runSubscriptionList(
	ctx context.Context,
	services Services,
	streams *Streams,
	refresh bool,
	asJSON bool,
) error {
	active, err := activeTenantContext(ctx, streams)
	if err != nil {
		return err
	}
	svc, err := services.Subscriptions(streams.Log)
	if err != nil {
		return err
	}
	subscriptions, err := svc.List(ctx, active, refresh)
	if err != nil {
		return err
	}
	if asJSON {
		_, err = io.WriteString(streams.Stdout, renderSubscriptionsJSON(active, subscriptions, os.Getenv(activeSubscriptionEnv)))
		return err
	}
	_, err = io.WriteString(streams.Stdout, renderSubscriptionsHuman(subscriptions))
	return err
}

func switchSubscription(streams *Streams, target domain.Subscription) error {
	if err := streams.RequireShellIntegration("azkit sub " + target.ID); err != nil {
		return err
	}
	current := os.Getenv(activeSubscriptionEnv)
	changes := []ShellEnvChange{
		{Name: activeSubscriptionEnv, Value: target.ID},
		{Name: azureSubscriptionEnv, Value: target.ID},
		{Name: terraformSubscriptionEnv, Value: target.ID},
		{Name: terraformSubscriptionName, Value: target.Name},
	}
	if current != "" && current != target.ID {
		changes = append([]ShellEnvChange{{Name: previousSubscriptionEnv, Value: current}}, changes...)
	}
	script, err := streams.RenderShellEnv(changes)
	if err != nil {
		return err
	}
	_, err = io.WriteString(streams.Stdout, script)
	return err
}

func renderSubscriptionsHuman(subscriptions []domain.Subscription) string {
	if len(subscriptions) == 0 {
		return "No subscriptions.\n"
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintln(w, "ID\tNAME")
	for _, sub := range subscriptions {
		_, _ = fmt.Fprintf(w, "%s\t%s\n", sub.ID, sub.Name)
	}
	_ = w.Flush()
	return buf.String()
}

type subscriptionJSON struct {
	SubscriptionID   string `json:"subscription_id"`
	SubscriptionName string `json:"subscription_name"`
}

type subscriptionListJSON struct {
	Context       string             `json:"context"`
	TenantID      string             `json:"tenant_id"`
	Current       subscriptionJSON   `json:"current"`
	Subscriptions []subscriptionJSON `json:"subscriptions"`
}

type currentSubscriptionJSON struct {
	Context          string `json:"context"`
	TenantID         string `json:"tenant_id"`
	SubscriptionID   string `json:"subscription_id"`
	SubscriptionName string `json:"subscription_name"`
}

func renderCurrentSubscriptionHuman(id string, name string) string {
	if id == "" {
		return "No active subscription.\n"
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "Subscription:\t%s\n", id)
	if name != "" {
		_, _ = fmt.Fprintf(w, "Name:\t%s\n", name)
	}
	_ = w.Flush()
	return buf.String()
}

func renderSubscriptionsJSON(active domain.TenantContext, subscriptions []domain.Subscription, currentID string) string {
	entries := make([]subscriptionJSON, 0, len(subscriptions))
	current := subscriptionJSON{SubscriptionID: currentID}
	for _, sub := range subscriptions {
		entry := subscriptionJSON{SubscriptionID: sub.ID, SubscriptionName: sub.Name}
		entries = append(entries, entry)
		if sub.ID == currentID {
			current = entry
		}
	}
	return marshalJSON(subscriptionListJSON{
		Context:       active.Name,
		TenantID:      active.TenantID,
		Current:       current,
		Subscriptions: entries,
	})
}

func renderCurrentSubscriptionJSON(active domain.TenantContext, id string, name string) string {
	return marshalJSON(currentSubscriptionJSON{
		Context:          active.Name,
		TenantID:         active.TenantID,
		SubscriptionID:   id,
		SubscriptionName: name,
	})
}

func activeTenantContext(ctx context.Context, streams *Streams) (domain.TenantContext, error) {
	svc, err := contextService(streams)
	if err != nil {
		return domain.TenantContext{}, err
	}
	active, ok, err := svc.Current(ctx)
	if err != nil {
		return domain.TenantContext{}, err
	}
	if !ok {
		return domain.TenantContext{}, app.MissingActiveContext()
	}
	return active, nil
}
