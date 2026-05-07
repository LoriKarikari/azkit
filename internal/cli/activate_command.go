package cli

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type ActivateCmd struct {
	Scope    string        `required:"" help:"Azure resource scope ID"`
	Role     string        `required:"" help:"Role display name or definition ID"`
	Reason   string        `required:"" help:"Justification for the activation"`
	Duration time.Duration `default:"2h" help:"How long the role stays active"`
	JSON     bool          `help:"Output as JSON"`
}

func (c *ActivateCmd) Run(ctx context.Context, services Services, streams *Streams) error {
	act, err := services.Activate()
	if err != nil {
		return err
	}
	result, err := act.Activate(ctx, domain.ActivationRequest{
		ScopeID:  c.Scope,
		Role:     c.Role,
		Reason:   c.Reason,
		Duration: c.Duration,
	})
	if err != nil {
		return err
	}
	if c.JSON {
		_, err = io.WriteString(streams.Stdout, renderActivationJSON(result))
	} else {
		_, _ = io.WriteString(streams.Stderr, fmt.Sprintf("Activating %s on %s for %s\n", result.Role, result.ScopeName, result.Duration))
		_, err = io.WriteString(streams.Stdout, renderActivationHuman(result))
	}
	return err
}
