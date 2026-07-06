package cli

import (
	"errors"
	"runtime"

	"github.com/alecthomas/kong"
	"github.com/willabides/kongplete"
)

type CompletionCmd struct{}

func (c *CompletionCmd) BeforeApply(ctx *kong.Context) error {
	if err := completionPlatformError(runtime.GOOS); err != nil {
		return err
	}
	return (&kongplete.InstallCompletions{}).BeforeApply(ctx)
}

func completionPlatformError(goos string) error {
	if goos == "windows" {
		return errors.New("shell completion generation is not supported on Windows yet; use `azkit shell-init pwsh` for PowerShell shell integration")
	}
	return nil
}
