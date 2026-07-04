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
	CodeInvalidDuration          Code = "invalid_duration"
	CodeConflictingSelectors     Code = "conflicting_selectors"
	CodeShellIntegrationRequired Code = "shell_integration_required"
	CodeUnknownSubscription      Code = "unknown_subscription"
	CodeAmbiguousSubscription    Code = "ambiguous_subscription"
	CodeUnknownResourceGroup     Code = "unknown_resource_group"
)
