package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/LoriKarikari/pimctl/internal/cli"
	"github.com/LoriKarikari/pimctl/internal/runtime"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	rt := runtime.New()
	runner := cli.NewRunner(rt.Services(), os.Stdout, os.Stderr)
	os.Exit(runner.Run(ctx, os.Args[1:]))
}
