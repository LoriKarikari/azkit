package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type activationJSON struct {
	Role      string          `json:"role"`
	Scope     activationScope `json:"scope"`
	Duration  string          `json:"duration"`
	StartedAt string          `json:"started_at"`
	ExpiresAt string          `json:"expires_at"`
	Reason    string          `json:"reason"`
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
		Duration:  result.Duration.String(),
		StartedAt: renderTime(result.StartedAt),
		ExpiresAt: renderTime(result.ExpiresAt),
		Reason:    result.Reason,
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return string(b) + "\n"
}

func renderActivationHuman(result *domain.ActivationResult) string {
	var buf bytes.Buffer
	_, _ = fmt.Fprintf(&buf, "✓ Activated %s on %s\n", result.Role, result.ScopeName)
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "Duration:\t%s\n", result.Duration)
	_, _ = fmt.Fprintf(w, "Expires:\t%s\n", result.ExpiresAt.UTC().Format("2006-01-02 15:04 UTC"))
	_, _ = fmt.Fprintf(w, "Reason:\t%s\n", result.Reason)
	_ = w.Flush()
	return buf.String()
}

func renderTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}
