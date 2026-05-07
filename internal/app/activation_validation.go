package app

import (
	"fmt"
	"strings"

	"github.com/LoriKarikari/pimctl/internal/domain"
)

func validateActivation(req domain.ActivationRequest) error {
	if strings.TrimSpace(req.Role) == "" {
		return ErrMissingRole
	}
	if strings.TrimSpace(req.Reason) == "" {
		return ErrMissingReason
	}
	if req.Duration <= 0 {
		return &Error{
			Code:    CodeInvalidDuration,
			Message: fmt.Sprintf("Invalid activation duration: %s.", req.Duration),
		}
	}
	return nil
}
