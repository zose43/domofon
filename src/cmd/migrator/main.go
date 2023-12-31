package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"os"
)

func main() {
	var migrationsPath, migrationsTable, action string

	flag.StringVar(&action, "action", "up", "up or down")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to db migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "migrations", "name of migrations table")
	flag.Parse()

	if migrationsPath == "" {
		panic("migrations path is required")
	}

	psqlUrl := os.Getenv("POSTGRESQL_URL")
	if psqlUrl == "" {
		panic("empty env POSTGRESQL_URL")
	}

	m, err := migrate.New(
		"file://"+migrationsPath,
		fmt.Sprintf(
			"%s&x-migrations-path-table=%s&x-multi-statement=true",
			psqlUrl,
			migrationsTable,
		),
	)
	if err != nil {
		panic(err)
	}

	result := "migrations applied successfully"
	switch action {
	case "up":
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return
			}

			panic(err)
		}
	case "down":
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("no migrations to apply")
				return
			}

			panic(err)
		}
	default:
		result = "unsupported action"
	}

	fmt.Println(result)
}
