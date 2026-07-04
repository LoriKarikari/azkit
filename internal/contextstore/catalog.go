package contextstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

const (
	catalogFileName = "contexts.json"
	azureProfile    = "azureProfile.json"
)

type Catalog struct {
	configDir string
	stateDir  string
}

type catalogFile struct {
	Contexts []catalogRecord `json:"contexts"`
}

type catalogRecord struct {
	Name     string `json:"name"`
	TenantID string `json:"tenant_id"`
}

var _ app.ContextCatalog = (*Catalog)(nil)

func New(configDir string, stateDir string) *Catalog {
	return &Catalog{configDir: configDir, stateDir: stateDir}
}

func (c *Catalog) Save(ctx context.Context, item domain.TenantContext) (domain.TenantContext, error) {
	if err := ctx.Err(); err != nil {
		return domain.TenantContext{}, err
	}
	data, err := c.load()
	if err != nil {
		return domain.TenantContext{}, err
	}

	updated := false
	for i := range data.Contexts {
		if data.Contexts[i].Name == item.Name {
			data.Contexts[i].TenantID = item.TenantID
			updated = true
			break
		}
	}
	if !updated {
		data.Contexts = append(data.Contexts, catalogRecord{Name: item.Name, TenantID: item.TenantID})
	}
	if err := c.save(data); err != nil {
		return domain.TenantContext{}, err
	}
	contextDir := c.ContextDir(item.Name)
	if err := os.MkdirAll(contextDir, 0700); err != nil {
		return domain.TenantContext{}, fmt.Errorf("creating context cache dir: %w", err)
	}
	return c.enrich(catalogRecord{Name: item.Name, TenantID: item.TenantID}), nil
}

func (c *Catalog) Get(ctx context.Context, name string) (domain.TenantContext, bool, error) {
	if err := ctx.Err(); err != nil {
		return domain.TenantContext{}, false, err
	}
	data, err := c.load()
	if err != nil {
		return domain.TenantContext{}, false, err
	}
	for _, record := range data.Contexts {
		if record.Name == name {
			return c.enrich(record), true, nil
		}
	}
	return domain.TenantContext{}, false, nil
}

func (c *Catalog) List(ctx context.Context) ([]domain.TenantContext, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	data, err := c.load()
	if err != nil {
		return nil, err
	}
	contexts := make([]domain.TenantContext, 0, len(data.Contexts))
	for _, record := range data.Contexts {
		contexts = append(contexts, c.enrich(record))
	}
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Name < contexts[j].Name
	})
	return contexts, nil
}

func (c *Catalog) Remove(ctx context.Context, name string) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	data, err := c.load()
	if err != nil {
		return err
	}

	kept := make([]catalogRecord, 0, len(data.Contexts))
	removed := false
	for _, record := range data.Contexts {
		if record.Name == name {
			removed = true
			continue
		}
		kept = append(kept, record)
	}
	if !removed {
		return app.ContextNotFound(name)
	}
	data.Contexts = kept
	if err := c.save(data); err != nil {
		return err
	}
	if err := os.RemoveAll(c.ContextDir(name)); err != nil {
		return fmt.Errorf("removing context cache dir: %w", err)
	}
	return nil
}

func (c *Catalog) ContextDir(name string) string {
	return filepath.Join(c.stateDir, "contexts", name)
}

func (c *Catalog) load() (catalogFile, error) {
	path := c.catalogPath()
	contents, err := os.ReadFile(path) // #nosec G304 -- fixed catalog file under the user-controlled azkit config root.
	if errors.Is(err, os.ErrNotExist) {
		return catalogFile{Contexts: []catalogRecord{}}, nil
	}
	if err != nil {
		return catalogFile{}, fmt.Errorf("reading context catalog: %w", err)
	}
	if len(contents) == 0 {
		return catalogFile{Contexts: []catalogRecord{}}, nil
	}
	var data catalogFile
	if err := json.Unmarshal(contents, &data); err != nil {
		return catalogFile{}, fmt.Errorf("parsing context catalog: %w", err)
	}
	if data.Contexts == nil {
		data.Contexts = []catalogRecord{}
	}
	return data, nil
}

func (c *Catalog) save(data catalogFile) error {
	if err := os.MkdirAll(c.configDir, 0700); err != nil {
		return fmt.Errorf("creating config dir: %w", err)
	}
	sort.Slice(data.Contexts, func(i, j int) bool {
		return data.Contexts[i].Name < data.Contexts[j].Name
	})
	contents, err := json.MarshalIndent(data, "", "  ")
	if err != nil {
		return fmt.Errorf("encoding context catalog: %w", err)
	}
	contents = append(contents, '\n')
	if err := os.WriteFile(c.catalogPath(), contents, 0600); err != nil {
		return fmt.Errorf("writing context catalog: %w", err)
	}
	return nil
}

func (c *Catalog) catalogPath() string {
	return filepath.Join(c.configDir, catalogFileName)
}

func (c *Catalog) enrich(record catalogRecord) domain.TenantContext {
	dir := c.ContextDir(record.Name)
	return domain.TenantContext{
		Name:     record.Name,
		TenantID: record.TenantID,
		Dir:      dir,
		Status:   contextStatus(dir),
	}
}

func contextStatus(dir string) domain.ContextStatus {
	info, err := os.Stat(dir)
	if errors.Is(err, os.ErrNotExist) || (err == nil && !info.IsDir()) {
		return domain.ContextMissingDir
	}
	if err != nil {
		return domain.ContextMissingDir
	}
	if _, err := os.Stat(filepath.Join(dir, azureProfile)); err == nil {
		return domain.ContextReady
	}
	return domain.ContextNeedsLogin
}
