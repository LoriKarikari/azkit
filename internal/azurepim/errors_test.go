package azurepim

import (
	"errors"
	"net/http"
	"testing"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

func TestAzurePIMOperationError_mapsForbiddenResponse(t *testing.T) {
	got := azurePIMOperationError(&azcore.ResponseError{StatusCode: http.StatusForbidden})

	var appErr *app.Error
	if !errors.As(got, &appErr) || appErr.Code != domain.CodePermissionDenied {
		t.Fatalf("want permission denied, got %v", got)
	}
}

func TestAzurePIMOperationError_mapsAuthorizationFailedText(t *testing.T) {
	got := azurePIMOperationError(errors.New("AuthorizationFailed: denied"))

	var appErr *app.Error
	if !errors.As(got, &appErr) || appErr.Code != domain.CodePermissionDenied {
		t.Fatalf("want permission denied, got %v", got)
	}
}

func TestAzurePIMOperationError_mapsOtherErrors(t *testing.T) {
	got := azurePIMOperationError(errors.New("boom"))

	var appErr *app.Error
	if !errors.As(got, &appErr) || appErr.Code != domain.CodeAzureAPIError {
		t.Fatalf("want azure api error, got %v", got)
	}
}
