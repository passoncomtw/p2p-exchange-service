// cmd/migrate/main.go — 獨立執行的 DB migration 工具
// 執行：DATABASE_URL=postgres://... ./migrate
package main

import (
	"context"
	"fmt"
	"io/fs"
	"log"
	"os"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"
	"p2p-exchange/migrations"
)

func main() {
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		log.Fatal("DATABASE_URL environment variable is required")
	}

	ctx := context.Background()
	conn, err := pgx.Connect(ctx, dsn)
	if err != nil {
		log.Fatalf("connect to database: %v", err)
	}
	defer conn.Close(ctx)

	if err := ensureMigrationsTable(ctx, conn); err != nil {
		log.Fatalf("ensure schema_migrations: %v", err)
	}

	applied, err := loadApplied(ctx, conn)
	if err != nil {
		log.Fatalf("load applied migrations: %v", err)
	}

	files, err := listMigrationFiles()
	if err != nil {
		log.Fatalf("list migration files: %v", err)
	}

	count := 0
	for _, name := range files {
		if applied[name] {
			fmt.Printf("[skip]  %s\n", name)
			continue
		}

		sql, err := fs.ReadFile(migrations.FS, name)
		if err != nil {
			log.Fatalf("read %s: %v", name, err)
		}

		fmt.Printf("[apply] %s ...\n", name)
		if _, err := conn.Exec(ctx, string(sql)); err != nil {
			log.Fatalf("execute %s: %v", name, err)
		}

		if _, err := conn.Exec(ctx,
			"INSERT INTO schema_migrations (filename) VALUES ($1)", name,
		); err != nil {
			log.Fatalf("record %s: %v", name, err)
		}

		fmt.Printf("[done]  %s\n", name)
		count++
	}

	if count == 0 {
		fmt.Println("No pending migrations.")
	} else {
		fmt.Printf("Applied %d migration(s).\n", count)
	}
}

func ensureMigrationsTable(ctx context.Context, conn *pgx.Conn) error {
	_, err := conn.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id          SERIAL PRIMARY KEY,
			filename    VARCHAR(255) NOT NULL UNIQUE,
			applied_at  TIMESTAMPTZ  NOT NULL DEFAULT NOW()
		)
	`)
	return err
}

func loadApplied(ctx context.Context, conn *pgx.Conn) (map[string]bool, error) {
	rows, err := conn.Query(ctx, "SELECT filename FROM schema_migrations")
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	applied := make(map[string]bool)
	for rows.Next() {
		var name string
		if err := rows.Scan(&name); err != nil {
			return nil, err
		}
		applied[name] = true
	}
	return applied, rows.Err()
}

func listMigrationFiles() ([]string, error) {
	entries, err := fs.ReadDir(migrations.FS, ".")
	if err != nil {
		return nil, err
	}

	var files []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasSuffix(e.Name(), ".sql") {
			files = append(files, e.Name())
		}
	}
	sort.Strings(files)
	return files, nil
}
