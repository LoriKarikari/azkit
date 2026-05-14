package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type deactivationJSON struct {
	Role         string `json:"role"`
	ScopeType    string `json:"scope_type"`
	ScopeID      string `json:"scope_id"`
	ScopeName    string `json:"scope_name"`
	AssignmentID string `json:"assignment_id"`
	Status       string `json:"status"`
}

func renderDeactivationJSON(result *domain.DeactivationResult) string {
	out := deactivationJSON{
		Role:         result.Role,
		ScopeType:    string(result.ScopeType),
		ScopeID:      result.ScopeID,
		ScopeName:    result.ScopeName,
		AssignmentID: result.AssignmentID,
		Status:       string(result.Status),
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return string(b) + "\n"
}

func renderDeactivationHuman(result *domain.DeactivationResult) string {
	var buf bytes.Buffer
	_, _ = fmt.Fprintf(&buf, "Deactivation requested for %s on %s\n", result.Role, result.ScopeName)
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)
	_, _ = fmt.Fprintf(w, "Assignment ID:\t%s\n", result.AssignmentID)
	_ = w.Flush()
	return buf.String()
}
