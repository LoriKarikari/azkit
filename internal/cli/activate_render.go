package cli

import (
	"bytes"
	"fmt"
	"text/tabwriter"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type activationJSON struct {
	Role          string          `json:"role"`
	Scope         activationScope `json:"scope"`
	Duration      string          `json:"duration"`
	StartedAt     string          `json:"started_at"`
	ExpiresAt     string          `json:"expires_at"`
	Reason        string          `json:"reason"`
	AlreadyActive bool            `json:"already_active"`
	Pending       bool            `json:"pending"`
}

type activationScope struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func renderActivationJSON(result *domain.ActivationResult) string {
	out := activationJSON{
		Role: result.Role,
		Scope: activationScope{
			ID:   result.ScopeID,
			Name: result.ScopeName,
		},
		Duration:      result.Duration.String(),
		StartedAt:     jsonTime(result.StartedAt),
		ExpiresAt:     jsonTime(result.ExpiresAt),
		Reason:        result.Reason,
		AlreadyActive: result.AlreadyActive,
		Pending:       result.Pending,
	}
	return marshalJSON(out)
}

func renderActivationHuman(result *domain.ActivationResult) string {
	var buf bytes.Buffer
	message := "✓ Activated %s on %s\n"
	if result.Pending {
		message = "✓ Activation requested for %s on %s\n"
	}
	if result.AlreadyActive {
		message = "✓ Already active: %s on %s\n"
	}
	_, _ = fmt.Fprintf(&buf, message, result.Role, result.ScopeName)
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "Duration:\t%s\n", result.Duration)
	_, _ = fmt.Fprintf(w, "Expires:\t%s\n", result.ExpiresAt.UTC().Format("2006-01-02 15:04 UTC"))
	if result.Reason != "" {
		_, _ = fmt.Fprintf(w, "Reason:\t%s\n", result.Reason)
	}
	if result.Pending {
		_, _ = fmt.Fprintln(w, "Status:\tWaiting for Azure to report this assignment as active")
	}
	_ = w.Flush()
	return buf.String()
}
