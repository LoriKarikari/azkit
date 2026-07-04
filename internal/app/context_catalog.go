package app

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/LoriKarikari/azkit/internal/domain"
)

var contextNameRE = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]{0,62}$`)

var reservedContextNames = map[string]struct{}{
	"-":       {},
	"add":     {},
	"current": {},
	"help":    {},
	"list":    {},
	"rm":      {},
}

type ContextCatalog interface {
	Save(context.Context, domain.TenantContext) (domain.TenantContext, error)
	List(context.Context) ([]domain.TenantContext, error)
	Remove(context.Context, string) error
}

type ContextService struct {
	catalog    ContextCatalog
	activeName func() string
}

func NewContextService(catalog ContextCatalog, activeName func() string) *ContextService {
	if activeName == nil {
		activeName = func() string { return "" }
	}
	return &ContextService{catalog: catalog, activeName: activeName}
}

func (s *ContextService) Add(ctx context.Context, name string, tenantID string) (domain.TenantContext, error) {
	name = strings.TrimSpace(name)
	tenantID = strings.TrimSpace(tenantID)
	if err := validateContextName(name); err != nil {
		return domain.TenantContext{}, err
	}
	if tenantID == "" {
		return domain.TenantContext{}, MissingTenantID()
	}
	return s.catalog.Save(ctx, domain.TenantContext{Name: name, TenantID: tenantID})
}

func (s *ContextService) List(ctx context.Context) ([]domain.TenantContext, error) {
	contexts, err := s.catalog.List(ctx)
	if err != nil {
		return nil, err
	}
	sort.Slice(contexts, func(i, j int) bool {
		return contexts[i].Name < contexts[j].Name
	})
	return contexts, nil
}

func (s *ContextService) Get(ctx context.Context, name string) (domain.TenantContext, error) {
	name = strings.TrimSpace(name)
	if err := validateContextName(name); err != nil {
		return domain.TenantContext{}, err
	}
	contexts, err := s.List(ctx)
	if err != nil {
		return domain.TenantContext{}, err
	}
	for _, item := range contexts {
		if item.Name == name {
			return item, nil
		}
	}
	return domain.TenantContext{}, ContextNotFound(name)
}

func (s *ContextService) Current(ctx context.Context) (domain.TenantContext, bool, error) {
	name := strings.TrimSpace(s.activeName())
	if name == "" {
		return domain.TenantContext{}, false, nil
	}
	item, err := s.Get(ctx, name)
	if err != nil {
		return domain.TenantContext{}, false, err
	}
	return item, true, nil
}

func (s *ContextService) Remove(ctx context.Context, name string, force bool) error {
	name = strings.TrimSpace(name)
	if err := validateContextName(name); err != nil {
		return err
	}
	if !force && name == s.activeName() {
		return ActiveContextRemoval(name)
	}
	return s.catalog.Remove(ctx, name)
}

func validateContextName(name string) error {
	if !contextNameRE.MatchString(name) {
		return InvalidContextName(name)
	}
	if _, ok := reservedContextNames[strings.ToLower(name)]; ok {
		return InvalidContextName(name)
	}
	return nil
}

func InvalidContextName(name string) *Error {
	return &Error{
		Code: domain.CodeInvalidContextName,
		Message: fmt.Sprintf(
			"Invalid context name %q. Use a portable name starting with a letter and containing only letters, numbers, hyphens, or underscores.",
			name,
		),
	}
}

func MissingTenantID() *Error {
	return &Error{
		Code:    domain.CodeMissingTenant,
		Message: "Tenant ID is required. Pass --tenant or set AZURE_TENANT_ID.",
	}
}

func MissingContextName() *Error {
	return &Error{
		Code:    domain.CodeMissingContext,
		Message: "Context name is required outside an interactive terminal.",
	}
}

func MissingActiveContext() *Error {
	return &Error{
		Code:    domain.CodeMissingActiveContext,
		Message: "No active context. Run `azkit ctx <name>` first.",
	}
}

func ContextNeedsLogin(ctx domain.TenantContext) *Error {
	return &Error{
		Code: domain.CodeContextNeedsLogin,
		Message: fmt.Sprintf(
			"Context %q needs Azure login. Run `az login --tenant %s`.",
			ctx.Name,
			ctx.TenantID,
		),
	}
}

func ContextNotFound(name string) *Error {
	return &Error{
		Code:    domain.CodeContextNotFound,
		Message: fmt.Sprintf("Context %q was not found.", name),
	}
}

func PreviousContextNotFound() *Error {
	return &Error{
		Code:    domain.CodePreviousContextNotFound,
		Message: "Previous context is not set in this shell.",
	}
}

func ActiveContextRemoval(name string) *Error {
	return &Error{
		Code:    domain.CodeActiveContextRemoval,
		Message: fmt.Sprintf("Context %q is active. Switch away first, or pass --force to remove it anyway.", name),
	}
}

func ContextRemovalNeedsForce(name string) *Error {
	return &Error{
		Code:    domain.CodeConfirmationRequired,
		Message: fmt.Sprintf("Removing context %q requires confirmation. Re-run with --force outside an interactive terminal.", name),
	}
}
