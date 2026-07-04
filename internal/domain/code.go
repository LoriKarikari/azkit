package domain

type Code string

const (
	CodeAuthFailed               Code = "authentication_failed"
	CodeAzureAPIError            Code = "azure_api_error"
	CodePermissionDenied         Code = "permission_denied"
	CodeEligibleNotFound         Code = "eligible_not_found"
	CodeActiveAssignmentNotFound Code = "active_assignment_not_found"
	CodeMissingScope             Code = "missing_scope"
	CodeMissingRole              Code = "missing_role"
	CodeMissingReason            Code = "missing_reason"
	CodeMissingTenant            Code = "missing_tenant"
	CodeMissingContext           Code = "missing_context"
	CodeMissingActiveContext     Code = "missing_active_context"
	CodeMissingSubscriptionCmd   Code = "missing_subscription_command"
	CodeContextNeedsLogin        Code = "context_needs_login"
	CodeContextEnvMismatch       Code = "context_environment_mismatch"
	CodeInvalidDuration          Code = "invalid_duration"
	CodeInvalidContextName       Code = "invalid_context_name"
	CodeConflictingSelectors     Code = "conflicting_selectors"
	CodeConfirmationRequired     Code = "confirmation_required"
	CodeShellIntegrationRequired Code = "shell_integration_required"
	CodeActiveContextRemoval     Code = "active_context_removal"
	CodeContextNotFound          Code = "context_not_found"
	CodePreviousContextNotFound  Code = "previous_context_not_found"
	CodeUnknownSubscription      Code = "unknown_subscription"
	CodeAmbiguousSubscription    Code = "ambiguous_subscription"
	CodeUnknownResourceGroup     Code = "unknown_resource_group"
)
