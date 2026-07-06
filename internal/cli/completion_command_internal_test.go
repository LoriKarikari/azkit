package cli

import (
	"strings"
	"testing"
)

func TestCompletionPlatformError_WindowsIsFriendly(t *testing.T) {
	err := completionPlatformError("windows")
	if err == nil {
		t.Fatal("want Windows completion error")
	}
	got := err.Error()
	if !strings.Contains(got, "not supported on Windows") {
		t.Fatalf("want Windows-specific error, got: %s", got)
	}
	if !strings.Contains(got, "azkit shell-init pwsh") {
		t.Fatalf("want PowerShell shell-init guidance, got: %s", got)
	}
	if strings.Contains(strings.ToLower(got), "cmd.exe") {
		t.Fatalf("raw cmd.exe shell error should not leak, got: %s", got)
	}
}

func TestCompletionPlatformError_NonWindowsAllowed(t *testing.T) {
	if err := completionPlatformError("linux"); err != nil {
		t.Fatalf("linux should allow completion generation: %v", err)
	}
}
