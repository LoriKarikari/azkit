package app

import (
	"fmt"

	"github.com/LoriKarikari/azkit/internal/domain"
)

type Error struct {
	Code    domain.Code
	Message string
	Cause   error
}

func (e *Error) Error() string {
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.Cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

func (e *Error) Unwrap() error {
	return e.Cause
}

func AuthFailed(err error) *Error {
	return &Error{
		Code:    domain.CodeAuthFailed,
		Message: "Could not authenticate with Azure.",
		Cause:   err,
	}
}

func AzureAPIError(err error) *Error {
	return &Error{
		Code:    domain.CodeAzureAPIError,
		Message: "Azure API call failed.",
		Cause:   err,
	}
}

func PermissionDenied(err error, operation string) *Error {
	return &Error{
		Code:    domain.CodePermissionDenied,
		Message: fmt.Sprintf("Insufficient permissions for %s operations.", operation),
		Cause:   err,
	}
}

func ShellIntegrationRequired(command string) *Error {
	return &Error{
		Code: domain.CodeShellIntegrationRequired,
		Message: fmt.Sprintf(
			"%s changes your current shell. Enable shell integration with `azkit shell-init <shell>` first.",
			command,
		),
	}
}

func UnknownSubscription(selector string) *Error {
	return &Error{
		Code:    domain.CodeUnknownSubscription,
		Message: fmt.Sprintf("Subscription %q was not found. Use `azkit sub -l` to list subscriptions or check your alias.", selector),
	}
}

func AmbiguousSubscription(selector string) *Error {
	return &Error{
		Code:    domain.CodeAmbiguousSubscription,
		Message: fmt.Sprintf("Subscription %q matches more than one subscription. Use the subscription ID.", selector),
	}
}

func InvalidAliasName(name string) *Error {
	return &Error{
		Code:    domain.CodeInvalidAliasName,
		Message: fmt.Sprintf("Invalid alias name %q. Use a portable name starting with a letter and containing only letters, numbers, hyphens, or underscores.", name),
	}
}

func AliasNameCollision(alias string, sub domain.Subscription) *Error {
	return &Error{
		Code:    domain.CodeAliasNameCollision,
		Message: fmt.Sprintf("Alias %q matches the existing subscription %q (%s). Choose a different alias.", alias, sub.Name, sub.ID),
	}
}

func AliasAlreadyExists(alias string) *Error {
	return &Error{
		Code:    domain.CodeAliasAlreadyExists,
		Message: fmt.Sprintf("Alias %q already exists. Remove it first or choose a different name.", alias),
	}
}

func AliasNotFound(alias string) *Error {
	return &Error{
		Code:    domain.CodeAliasNotFound,
		Message: fmt.Sprintf("Alias %q was not found.", alias),
	}
}

func MissingAliasName() *Error {
	return &Error{
		Code:    domain.CodeMissingAlias,
		Message: "Alias name is required.",
	}
}

func MissingAliasSelector() *Error {
	return &Error{
		Code:    domain.CodeMissingAliasSelector,
		Message: "Subscription selector (ID or exact name) is required.",
	}
}

func PreviousSubscriptionNotFound() *Error {
	return &Error{
		Code:    domain.CodePreviousSubscriptionNotFound,
		Message: "Previous subscription is not set in this shell.",
	}
}

func ConflictingSubscriptionSelectors() *Error {
	return &Error{
		Code:    domain.CodeConflictingSelectors,
		Message: "Use either a subscription target or --list/--refresh, not both.",
	}
}
