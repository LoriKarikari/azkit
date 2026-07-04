package azurepim

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/LoriKarikari/azkit/internal/app"
	"github.com/LoriKarikari/azkit/internal/domain"
)

func azurePIMOperationError(err error) error {
	if azurePIMPermissionDenied(err) {
		return app.PermissionDenied(err, "PIM")
	}
	if activeDurationTooShort(err) {
		return &app.Error{
			Code:    domain.CodeAzureAPIError,
			Message: "Role must be active for at least 5 minutes before deactivating.",
			Cause:   err,
		}
	}
	return app.AzureAPIError(err)
}

func activeDurationTooShort(err error) bool {
	var responseErr *azcore.ResponseError
	if errors.As(err, &responseErr) {
		return responseErr.ErrorCode == "ActiveDurationTooShort"
	}
	return strings.Contains(err.Error(), "ActiveDurationTooShort")
}

func azurePIMPermissionDenied(err error) bool {
	var responseErr *azcore.ResponseError
	if errors.As(err, &responseErr) {
		return responseErr.StatusCode == http.StatusForbidden || responseErr.ErrorCode == "AuthorizationFailed"
	}
	return strings.Contains(err.Error(), "AuthorizationFailed")
}
