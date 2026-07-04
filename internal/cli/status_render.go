package cli

import (
	"bytes"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/LoriKarikari/azkit/internal/domain"
)

type statusEntryJSON struct {
	Role         string `json:"role"`
	ScopeType    string `json:"scope_type"`
	ScopeID      string `json:"scope_id"`
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
			ScopeID:      a.ScopeID,
			ScopeName:    a.ScopeName,
			Status:       string(a.Status),
			StartedAt:    jsonTime(a.StartTime),
			ExpiresAt:    jsonTime(a.EndTime),
			AssignmentID: a.ID,
		}
	}
	return marshalJSON(out)
}

func RenderStatusHuman(as []domain.ActiveAssignment, extended bool) string {
	if len(as) == 0 {
		return "No active assignments.\n"
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	if extended {
		_, _ = fmt.Fprintln(w, "ROLE\tTYPE\tNAME\tSTATUS\tSTARTED\tEXPIRES\tASSIGNMENT ID\tSCOPE ID")
		for _, a := range as {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\t%s\t%s\n",
				a.Role, a.ScopeType, a.ScopeName,
				a.Status,
				humanTime(a.StartTime), humanTime(a.EndTime),
				a.ID, a.ScopeID)
		}
	} else {
		_, _ = fmt.Fprintln(w, "ROLE\tTYPE\tNAME\tSTATUS\tEXPIRES")
		for _, a := range as {
			status := statusSymbol(a.Status)
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\n",
				a.Role, a.ScopeType, a.ScopeName,
				status,
				expiresStatus(a.EndTime))
		}
	}
	_ = w.Flush()
	return buf.String()
}

func statusSymbol(s domain.ActiveAssignmentStatus) string {
	if s == domain.ActiveAssignmentActive {
		return "Active"
	}
	return string(s)
}

func expiresStatus(end time.Time) string {
	if end.IsZero() {
		return "Permanent"
	}
	return humanTime(end)
}
