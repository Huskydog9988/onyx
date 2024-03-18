package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"

	"github.com/huskydog9988/onyx/discordbot"
)

var logger = slog.New(slog.NewTextHandler(os.Stdout, nil))

func main() {
	// Handle SIGINT (CTRL+C) gracefully.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt)
	defer stop()

	onyx := discordbot.New(ctx, logger)
	defer onyx.Close(ctx)

	// wait for context to be done
	<-ctx.Done()
}
