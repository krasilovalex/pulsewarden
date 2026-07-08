package main

import (
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/krasilovalex/pulsewarden/internal/platform/config"
)

const defaultMigrationsPath = "file://migrations"

func main() {
	os.Exit(run())
}

func run() int {
	var (
		direction      string
		migrationsPath string
	)

	flag.StringVar(&direction, "direction", "up", "migration direction: up or down")

	flag.StringVar(&migrationsPath, "path", defaultMigrationsPath, "path to migration files")

	flag.Parse()

	if direction != "up" && direction != "down" {
		fmt.Fprintf(os.Stderr, "invalid migration direction %q: expected up or down\n", direction)
		return 2
	}

	cfg, err := config.Load()
	if err != nil {
		fmt.Fprintf(os.Stderr, "load config: %v\n", err)
		return 1
	}

	migrator, err := migrate.New(
		migrationsPath,
		cfg.Postgres.DSN,
	)
	if err != nil {
		fmt.Fprintf(os.Stderr, "create migrator: %v\n", err)
		return 1
	}

	defer func() {
		sourceErr, databaseErr := migrator.Close()

		if sourceErr != nil {
			fmt.Fprintf(os.Stderr, "close migration source: %v\n", sourceErr)
		}

		if databaseErr != nil {
			fmt.Fprintf(os.Stderr, "close migration database: %v\n", databaseErr)
		}
	}()

	switch direction {
	case "up":
		err = migrator.Up()
	case "down":
		err = migrator.Steps(-1)
	}

	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		fmt.Fprintf(
			os.Stderr,
			"run %s migrations: %v\n",
			direction,
			err,
		)
		return 1
	}

	if errors.Is(err, migrate.ErrNoChange) {
		fmt.Println("no migration changes")
		return 0
	}

	fmt.Printf("migrations %s completed\n", direction)

	return 0
}
