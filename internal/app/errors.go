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
