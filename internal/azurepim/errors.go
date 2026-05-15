package azurepim

import (
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"

	"github.com/LoriKarikari/pimctl/internal/app"
	"github.com/LoriKarikari/pimctl/internal/domain"
)

func azurePIMOperationError(err error) error {
	if azurePIMPermissionDenied(err) {
		return app.PermissionDenied(err)
	}
	if duration := activeDurationTooShort(err); duration != 0 {
		return &app.Error{
			Code:    domain.CodeAzureAPIError,
			Message: fmt.Sprintf("Role must be active for at least %s before deactivating.", duration),
			Cause:   err,
		}
	}
	return app.AzureAPIError(err)
}

func activeDurationTooShort(err error) time.Duration {
	var responseErr *azcore.ResponseError
	if errors.As(err, &responseErr) && responseErr.ErrorCode == "ActiveDurationTooShort" {
		// Extract minimum duration from message like "The Active duration is too short. Minimum Required is 5 minutes."
		msg := responseErr.Error()
		if parts := strings.Split(msg, "Minimum Required is "); len(parts) == 2 {
			if d, err := parseDurationString(strings.TrimSuffix(parts[1], ".")); err == nil {
				return d
			}
		}
		return 5 * time.Minute // fallback
	}
	return 0
}

func parseDurationString(s string) (time.Duration, error) {
	var b strings.Builder
	for _, c := range s {
		if c >= '0' && c <= '9' {
			b.WriteRune(c)
		}
	}
	digits := b.String()
	if d, err := strconv.Atoi(digits); err == nil {
		if strings.Contains(s, "hour") {
			return time.Duration(d) * time.Hour, nil
		}
		if strings.Contains(s, "minute") {
			return time.Duration(d) * time.Minute, nil
		}
	}
	return 0, fmt.Errorf("cannot parse duration from %q", s)
}

func azurePIMPermissionDenied(err error) bool {
	var responseErr *azcore.ResponseError
	if errors.As(err, &responseErr) {
		return responseErr.StatusCode == http.StatusForbidden || responseErr.ErrorCode == "AuthorizationFailed"
	}
	message := err.Error()
	return strings.Contains(message, "AuthorizationFailed") || strings.Contains(message, "403")
}
