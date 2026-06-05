package cli

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/LoriKarikari/azkit/internal/app"
)

type errorBody struct {
	Error errorPayload `json:"error"`
}

type errorPayload struct {
	Code    string `json:"code"`
	Message string `json:"message"`
}

func RenderError(err error, jsonOutput bool) string {
	code, message := errorDetails(err)
	if jsonOutput {
		b, _ := json.MarshalIndent(errorBody{Error: errorPayload{Code: code, Message: message}}, "", "  ")
		return string(b) + "\n"
	}
	return fmt.Sprintf("Error: %s\n", message)
}

func errorDetails(err error) (string, string) {
	var appErr *app.Error
	if errors.As(err, &appErr) {
		return string(appErr.Code), appErr.Message
	}
	return "command_failed", err.Error()
}
