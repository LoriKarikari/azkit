package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"
	"time"

	"github.com/LoriKarikari/azkit/internal/domain"
)

type assignmentJSON struct {
	Role          string `json:"role"`
	ScopeType     string `json:"scope_type"`
	ScopeID       string `json:"scope_id"`
	ScopeName     string `json:"scope_name"`
	EligibleUntil string `json:"eligible_until"`
	AssignmentID  string `json:"assignment_id"`
}

func RenderJSON(as []domain.EligibleAssignment) string {
	if as == nil {
		as = []domain.EligibleAssignment{}
	}
	out := make([]assignmentJSON, len(as))
	for i, a := range as {
		out[i] = assignmentJSON{
			Role:          a.Role,
			ScopeType:     string(a.ScopeType),
			ScopeID:       a.ScopeID,
			ScopeName:     a.ScopeName,
			EligibleUntil: jsonTime(a.EligibleUntil),
			AssignmentID:  a.ID,
		}
	}
	return marshalJSON(out)
}

func marshalJSON(v any) string {
	b, _ := json.MarshalIndent(v, "", "  ")
	return string(b) + "\n"
}

func RenderHuman(as []domain.EligibleAssignment, verbose bool) string {
	if len(as) == 0 {
		return "No eligible assignments.\n"
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	if verbose {
		_, _ = fmt.Fprintln(w, "ROLE\tTYPE\tNAME\tELIGIBLE UNTIL\tASSIGNMENT ID\tSCOPE ID")
		for _, a := range as {
			_, _ = fmt.Fprintf(
				w,
				"%s\t%s\t%s\t%s\t%s\t%s\n",
				a.Role,
				a.ScopeType,
				a.ScopeName,
				humanTime(a.EligibleUntil),
				a.ID,
				a.ScopeID,
			)
		}
	} else {
		_, _ = fmt.Fprintln(w, "ROLE\tTYPE\tNAME\tELIGIBLE UNTIL")
		for _, a := range as {
			_, _ = fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.Role, a.ScopeType, a.ScopeName, humanTime(a.EligibleUntil))
		}
	}
	_ = w.Flush()
	return buf.String()
}

func jsonTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func humanTime(t time.Time) string {
	if t.IsZero() {
		return "-"
	}
	return t.UTC().Format("2006-01-02 15:04 UTC")
}
