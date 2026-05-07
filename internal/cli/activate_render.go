package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

func renderActivationJSON(result *domain.ActivationResult) string {
	b, _ := json.MarshalIndent(result, "", "  ")
	return string(b) + "\n"
}

func renderActivationHuman(result *domain.ActivationResult) string {
	var buf bytes.Buffer
	fmt.Fprintf(&buf, "✓ Activated %s on %s\n", result.Role, result.ScopeName)
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	fmt.Fprintf(w, "Duration:\t%s\n", result.Duration)
	fmt.Fprintf(w, "Expires:\t%s\n", result.ExpiresAt.UTC().Format("2006-01-02 15:04 UTC"))
	fmt.Fprintf(w, "Reason:\t%s\n", result.Reason)
	w.Flush()
	return buf.String()
}
