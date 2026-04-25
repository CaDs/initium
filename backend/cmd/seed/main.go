// Seed binary for dev data. Runs via `make db:seed`.
//
// This template ships empty — there's nothing to seed in the starter.
// Add idempotent seed logic here as your schema grows:
//
//  1. Load config via internal/infra/config.Load().
//  2. Open the database via internal/infra/database.NewPostgresDB.
//  3. Insert-or-update fixtures. Use ON CONFLICT DO NOTHING / UPSERT so
//     `make db:seed` stays safe to run repeatedly.
//
// Keep seeds in version control. Production should never run this binary.
package main

import (
	"log/slog"
	"os"
)

func main() {
	slog.SetDefault(slog.New(slog.NewTextHandler(os.Stderr, nil)))
	slog.Info("no seed data yet — add fixtures to backend/cmd/seed/main.go as schema grows")
}
