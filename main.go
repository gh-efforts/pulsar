package main

import (
	"context"

	"github.com/bitrainforest/pulsar/commands"

	"github.com/urfave/cli/v2"

	"log"
	"os"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go func() {
		signal := make(chan os.Signal, 1)
		select {
		case <-signal:
			cancel()
		case <-ctx.Done():
		}

	}()

	app := &cli.App{
		Commands: []*cli.Command{
			commands.DaemonCmd, commands.InitCmd, commands.PulsarCommand},
	}
	if err := app.RunContext(ctx, os.Args); err != nil {
		log.Fatal(err)
	}
}
