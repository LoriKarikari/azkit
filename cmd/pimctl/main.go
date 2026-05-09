package main

import (
	"context"
	"os"

	"github.com/LoriKarikari/pimctl/internal/cli"
	"github.com/LoriKarikari/pimctl/internal/runtime"
)

func main() {
	rt := runtime.New()
	runner := cli.NewRunner(rt.Services(), os.Stdout, os.Stderr)
	os.Exit(runner.Run(context.Background(), os.Args[1:]))
}
