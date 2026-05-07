package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"text/tabwriter"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

type assignmentJSON struct {
	Role         string `json:"role"`
	ScopeType    string `json:"scope_type"`
	ScopeID      string `json:"scope_id"`
	ScopeName    string `json:"scope_name"`
	MaxDuration  string `json:"max_duration"`
	AssignmentID string `json:"assignment_id"`
}

func RenderJSON(as []domain.EligibleAssignment) string {
	if as == nil {
		as = []domain.EligibleAssignment{}
	}
	out := make([]assignmentJSON, len(as))
	for i, a := range as {
		out[i] = assignmentJSON{
			Role:         a.Role,
			ScopeType:    string(a.ScopeType),
			ScopeID:      a.ScopeID,
			ScopeName:    a.ScopeName,
			MaxDuration:  a.MaxDuration,
			AssignmentID: a.ID,
		}
	}
	b, _ := json.MarshalIndent(out, "", "  ")
	return string(b) + "\n"
}

func RenderHuman(as []domain.EligibleAssignment, verbose bool) string {
	if len(as) == 0 {
		return "No eligible assignments.\n"
	}
	var buf bytes.Buffer
	w := tabwriter.NewWriter(&buf, 0, 0, 2, ' ', 0)

	if verbose {
		fmt.Fprintln(w, "ROLE\tTYPE\tNAME\tMAX DURATION\tASSIGNMENT ID\tSCOPE ID")
		for _, a := range as {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\t%s\t%s\n", a.Role, a.ScopeType, a.ScopeName, a.MaxDuration, a.ID, a.ScopeID)
		}
	} else {
		fmt.Fprintln(w, "ROLE\tTYPE\tNAME\tMAX DURATION")
		for _, a := range as {
			fmt.Fprintf(w, "%s\t%s\t%s\t%s\n", a.Role, a.ScopeType, a.ScopeName, a.MaxDuration)
		}
	}
	w.Flush()
	return buf.String()
}
