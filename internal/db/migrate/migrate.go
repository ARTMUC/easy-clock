// Package migrate provides a reusable migration runner backed by golang-migrate.
// SQL files are embedded directly into the binary so no external migration files
// need to be present at runtime.
package migrate

import (
	"database/sql"
	"embed"
	"errors"
	"fmt"
	"log"

	"github.com/golang-migrate/migrate/v4"
	migratemysql "github.com/golang-migrate/migrate/v4/database/mysql"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	_ "github.com/go-sql-driver/mysql"
)

//go:embed *.sql
var migrationsFS embed.FS

// Run opens a MySQL connection with the given DSN and runs the requested
// migration command.
//
// Supported commands:
//
//	up      – apply all pending migrations
//	down    – roll back all applied migrations
//	steps N – apply (+N) or roll back (-N) exactly N migrations
//	version – print the current migration version and dirty flag, then exit
func Run(dsn, command string, steps, forceVersion int) error {
	sqlDB, err := sql.Open("mysql", dsn)
	if err != nil {
		return fmt.Errorf("open db: %w", err)
	}
	defer sqlDB.Close()

	src, err := iofs.New(migrationsFS, ".")
	if err != nil {
		return fmt.Errorf("create migration source: %w", err)
	}

	driver, err := migratemysql.WithInstance(sqlDB, &migratemysql.Config{})
	if err != nil {
		return fmt.Errorf("create mysql driver: %w", err)
	}

	m, err := migrate.NewWithInstance("iofs", src, "mysql", driver)
	if err != nil {
		return fmt.Errorf("create migrator: %w", err)
	}

	switch command {
	case "up":
		if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate up: %w", err)
		}
		log.Println("migrations applied successfully")

	case "down":
		if err := m.Down(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate down: %w", err)
		}
		log.Println("migrations rolled back successfully")

	case "steps":
		if steps == 0 {
			return fmt.Errorf("steps command requires a non-zero -steps value")
		}
		if err := m.Steps(steps); err != nil && !errors.Is(err, migrate.ErrNoChange) {
			return fmt.Errorf("migrate steps %d: %w", steps, err)
		}
		log.Printf("migrated %d step(s) successfully", steps)

	case "version":
		version, dirty, err := m.Version()
		if err != nil && !errors.Is(err, migrate.ErrNilVersion) {
			return fmt.Errorf("get version: %w", err)
		}
		log.Printf("migration version: %d, dirty: %v", version, dirty)

	case "force":
		// Force sets the migration version without running any SQL.
		// Use this ONLY after you have manually inspected and fixed the database state.
		// Typical flow when dirty:
		//   1. make migrate-version        — see which version is dirty (e.g. 8)
		//   2. inspect 000008_*.up.sql     — check what SQL ran before the failure
		//   3. fix the DB manually if needed
		//   4. make migrate-force N=7      — set version to last known-good (N-1)
		//   5. make migrate-up             — re-apply the fixed migration
		if forceVersion <= 0 {
			return fmt.Errorf("force command requires -force <version> (must be > 0)")
		}
		if err := m.Force(forceVersion); err != nil {
			return fmt.Errorf("migrate force %d: %w", forceVersion, err)
		}
		log.Printf("forced migration version to %d (dirty flag cleared)", forceVersion)

	default:
		return fmt.Errorf("unknown command %q – supported: up, down, steps, version", command)
	}

	return nil
}
