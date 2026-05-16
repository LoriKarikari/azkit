package cli

import (
	"github.com/alecthomas/kong"
	"github.com/willabides/kongplete"
)

type CompletionCmd struct{}

func (c *CompletionCmd) BeforeApply(ctx *kong.Context) error {
	return (&kongplete.InstallCompletions{}).BeforeApply(ctx)
}
