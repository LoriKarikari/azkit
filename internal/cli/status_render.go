package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type statusEntryJSON struct {
	Role         string `json:"role"`
	ScopeType    string `json:"scope_type"`
	ScopeName    string `json:"scope_name"`
	Status       string `json:"status"`
	StartedAt    string `json:"started_at"`
	ExpiresAt    string `json:"expires_at"`
	AssignmentID string `json:"assignment_id"`
}

func RenderStatusJSON(as []domain.ActiveAssignment) string {
	if as == nil {
		as = []domain.ActiveAssignment{}
	}
	out := make([]statusEntryJSON, len(as))
	for i, a := range as {
		out[i] = statusEntryJSON{
			Role:         a.Role,
			ScopeType:    string(a.ScopeType),
			ScopeName:    a.ScopeName,
			Status:       string(a.Status),
			StartedAt:    jsonTime(a.StartTime),
			ExpiresAt:    jsonTime(a.EndTime),
			AssignmentID: a.ID,
		}
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return string(b) + "\n"
}

func RenderStatusHuman(as []domain.ActiveAssignment, verbose bool) string {
	if len(as) == 0 {
		return "No active assignments.\n"
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	if verbose {
		fmt.Fprintln(w, "ROLE\tTYPE\tNAME\tSTATUS\tSTARTED\tEXPIRES\tASSIGNMENT ID")
		for _, a := range as {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				a.Role, a.ScopeType, a.ScopeName,
				a.Status,
				humanTime(a.StartTime), humanTime(a.EndTime),
				a.ID)
		}
	} else {
		fmt.Fprintln(w, "ROLE\tTYPE\tNAME\tSTATUS\tEXPIRES")
		for _, a := range as {
			status := statusSymbol(a.Status)
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				a.Role, a.ScopeType, a.ScopeName,
				status,
				expiresStatus(a.EndTime))
		}
	}
	w.Flush()
	return buf.String()
}

func statusSymbol(s domain.ActiveAssignmentStatus) string {
	switch s {
	case domain.ActiveAssignmentGranted, domain.ActiveAssignmentProvisioned:
		return "Active"
	case domain.ActiveAssignmentRequested:
		return "Requested"
	case domain.ActiveAssignmentRevoked:
		return "Revoked"
	default:
		return string(s)
	}
}

func expiresStatus(end time.Time) string {
	if end.IsZero() {
		return "Permanent"
	}
	return humanTime(end)
}
