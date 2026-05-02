// Command migrate is a standalone CLI for running database migrations.
//
// Usage:
//
//	migrate -cmd up                  # apply all pending migrations
//	migrate -cmd down                # roll back all applied migrations
//	migrate -cmd steps -steps 2      # apply 2 migrations forward
//	migrate -cmd steps -steps -1     # roll back 1 migration
//	migrate -cmd version             # show current version
//
// Connection parameters are read from the environment (same .env as the main
// server), so the command can be run locally or in a CI/CD pipeline:
//
//	GONE_DB_HOST=localhost GONE_DB_PORT=3306 ... migrate -cmd up
//
// or simply:
//
//	make migrate-up   (see Makefile)
package main

import (
	"flag"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"

	dbmigrate "starter/internal/db/migrate"
)

func main() {
	if err := godotenv.Overload(); err != nil {
		log.Println("no .env file, using system environment")
	}

	cmd   := flag.String("cmd", "up", "Migration command: up | down | steps | version | force")
	steps := flag.Int("steps", 0, "Steps to migrate (+N forward, -N backward) — only for -cmd steps")
	force := flag.Int("force", 0, "Version to force-set (clears dirty flag, does NOT touch schema) — only for -cmd force")
	flag.Parse()

	// MySQL DSN: user:password@tcp(host:port)/dbname?params
	dsn := fmt.Sprintf(
		"%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=UTC&multiStatements=true",
		os.Getenv("GONE_DB_USERNAME"),
		os.Getenv("GONE_DB_PASSWORD"),
		os.Getenv("GONE_DB_HOST"),
		os.Getenv("GONE_DB_PORT"),
		os.Getenv("GONE_DB_DATABASE"),
	)

	if err := dbmigrate.Run(dsn, *cmd, *steps, *force); err != nil {
		log.Fatalf("migration error: %v", err)
	}
}
