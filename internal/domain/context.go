package domain

type ContextStatus string

const (
	ContextReady      ContextStatus = "ready"
	ContextNeedsLogin ContextStatus = "needs_login"
	ContextMissingDir ContextStatus = "missing_dir"
)

type TenantContext struct {
	Name     string
	TenantID string
	Dir      string
	Status   ContextStatus
}
