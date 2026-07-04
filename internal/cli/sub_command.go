package cli

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"text/tabwriter"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
	"github.com/LoriKarikari/azkit/internal/subscriptionstore"
)

type SubCmd struct {
	List    bool `short:"l" name:"list" help:"List subscriptions for the active context"`
	Refresh bool `help:"Refresh the active context subscription cache"`
}

func (c *SubCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	active, err := activeTenantContext(ctx, streams)
	if err != nil {
		return err
	}
	svc := app.NewSubscriptionService(
		subscriptionstore.New(),
		func() (app.SubscriptionSource, error) {
			if services.SubscriptionSource == nil {
				return nil, app.AuthFailed(errors.New("subscription source is not configured"))
			}
			return services.SubscriptionSource(streams.Log)
		},
		services.Now,
	)
	subscriptions, err := svc.List(ctx, active, c.Refresh)
	if err != nil {
		return err
	}
	_, err = io.WriteString(streams.Stdout, renderSubscriptionsHuman(subscriptions))
	return err
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
