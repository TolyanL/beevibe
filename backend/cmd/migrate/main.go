package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/TolyanL/beevibe/config"
	"github.com/TolyanL/beevibe/migration"
)

func main() {
	command := flag.String("command", "up", "migration command: up, down, status")
	flag.Parse()

	c, err := config.NewMigrate()
	if err != nil {
		fmt.Fprintf(os.Stderr, "config.NewMigrate: %v\n", err)
		os.Exit(1)
	}

	ctx := context.Background()

	switch *command {
	case "up":
		err = migration.Up(ctx, c.DSN)
	case "down":
		err = migration.Down(ctx, c.DSN)
	case "status":
		err = migration.Status(ctx, c.DSN)
	default:
		err = fmt.Errorf("unknown command %q", *command)
	}

	if err != nil {
		fmt.Fprintf(os.Stderr, "migrate %s: %v\n", *command, err)
		os.Exit(1)
	}
}
