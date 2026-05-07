package app

import "fmt"

type Code string

const (
	CodeAuthFailed       Code = "authentication_failed"
	CodeAzureAPIError    Code = "azure_api_error"
	CodePermissionDenied Code = "permission_denied"
)

type Error struct {
	Code    Code
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
		Code:    CodeAuthFailed,
		Message: "Could not authenticate with Azure.",
		Cause:   err,
	}
}

func AzureAPIError(err error) *Error {
	return &Error{
		Code:    CodeAzureAPIError,
		Message: "Azure API call failed.",
		Cause:   err,
	}
}

func PermissionDenied(err error) *Error {
	return &Error{
		Code:    CodePermissionDenied,
		Message: "Insufficient permissions for PIM operations.",
		Cause:   err,
	}
}
