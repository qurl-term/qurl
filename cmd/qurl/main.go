package main

import (
	"context"
	"os"
	"os/signal"

	"github.com/qurl-term/qurl/internal/adapter/cli"
)

func main() {
	os.Exit(run())
}

func run() int {
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()
	return cli.Execute(ctx, os.Args[1:], cli.IO{In: os.Stdin, Out: os.Stdout, Err: os.Stderr})
}
