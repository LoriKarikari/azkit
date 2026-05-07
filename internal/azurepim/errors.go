package azurepim

import (
	"errors"
	"net/http"
	"strings"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/LoriKarikari/pimctl/internal/app"
)

func azurePIMOperationError(err error) error {
	if azurePIMPermissionDenied(err) {
		return app.PermissionDenied(err)
	}
	return app.AzureAPIError(err)
}

func azurePIMPermissionDenied(err error) bool {
	var responseErr *azcore.ResponseError
	if errors.As(err, &responseErr) {
		return responseErr.StatusCode == http.StatusForbidden || responseErr.ErrorCode == "AuthorizationFailed"
	}
	message := err.Error()
	return strings.Contains(message, "AuthorizationFailed") || strings.Contains(message, "403")
}
