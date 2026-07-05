package subscriptionstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/LoriKarikari/azkit/internal/domain"
)

const aliasesFileName = "aliases.json"

type Aliases struct{}

func NewAliases() *Aliases {
	return &Aliases{}
}

func (a *Aliases) Load(ctx context.Context, active domain.TenantContext) (map[string]domain.Subscription, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	path := aliasesPath(active)
	contents, err := os.ReadFile(path) // #nosec G304 -- fixed aliases file under the user-controlled context dir.
	if errors.Is(err, os.ErrNotExist) {
		return map[string]domain.Subscription{}, nil
	}
	if err != nil {
		return nil, fmt.Errorf("reading subscription aliases %s: %w", path, err)
	}
	var records map[string]aliasRecord
	if err := json.Unmarshal(contents, &records); err != nil {
		return nil, fmt.Errorf("parsing subscription aliases %s: %w", path, err)
	}
	out := make(map[string]domain.Subscription, len(records))
	for k, v := range records {
		out[k] = domain.Subscription{ID: v.ID, Name: v.Name}
	}
	return out, nil
}

func (a *Aliases) Save(
	ctx context.Context,
	active domain.TenantContext,
	aliases map[string]domain.Subscription,
) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if err := os.MkdirAll(active.Dir, 0700); err != nil {
		return fmt.Errorf("creating context dir: %w", err)
	}
	records := make(map[string]aliasRecord, len(aliases))
	for k, v := range aliases {
		records[k] = aliasRecord{ID: v.ID, Name: v.Name}
	}
	contents, err := json.MarshalIndent(records, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding subscription aliases: %w", err)
	}
	contents = append(contents, '\n')
	path := aliasesPath(active)
	if err := writeFileAtomic(path, contents); err != nil {
		return fmt.Errorf("writing subscription aliases %s: %w", path, err)
	}
	return nil
}

type aliasRecord struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func aliasesPath(active domain.TenantContext) string {
	return filepath.Join(active.Dir, aliasesFileName)
}
