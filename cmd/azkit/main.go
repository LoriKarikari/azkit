package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/LoriKarikari/azkit/internal/cli"
	"github.com/LoriKarikari/azkit/internal/runtime"
)

func main() {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)

	rt := runtime.New()
	runner := cli.NewRunner(rt.Services(), os.Stdout, os.Stderr)
	code := runner.Run(ctx, os.Args[1:])
	stop()
	os.Exit(code)
}
