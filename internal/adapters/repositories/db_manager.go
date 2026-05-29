package repositories

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DatabaseManager struct {
	Pool *pgxpool.Pool
}

func NewDatabaseManager(ctx context.Context, databaseURL, user, pass, host, port, dbname string) *DatabaseManager {
	var connStr string
	if databaseURL != "" {
		connStr = databaseURL
	} else {
		connStr = fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=require", user, pass, host, port, dbname)
	}

	poolConfig, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		log.Fatalf("Unable to parse config: %v", err)
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolConfig)
	if err != nil {
		log.Printf("⚠️ Warning: Unable to connect to database: %v", err)
		return &DatabaseManager{Pool: nil}
	}

	if err := pool.Ping(ctx); err != nil {
		log.Printf("⚠️ Warning: Database ping failed: %v", err)
	} else {
		log.Println("✅ Successfully connected to PostgreSQL")

		// Run migrations automatically (each file runs at most once)
		runMigrations(ctx, pool)
	}

	return &DatabaseManager{
		Pool: pool,
	}
}

func runMigrations(ctx context.Context, pool *pgxpool.Pool) {
	log.Println("Running database migrations...")

	if _, err := pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			filename TEXT PRIMARY KEY,
			applied_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
		)
	`); err != nil {
		log.Printf("⚠️ Could not ensure schema_migrations table: %v", err)
		return
	}

	migrationsDir := "db/migrations"
	if _, err := os.Stat(migrationsDir); os.IsNotExist(err) {
		migrationsDir = "/app/db/migrations"
	}

	files, err := os.ReadDir(migrationsDir)
	if err != nil {
		log.Printf("⚠️ Could not read migrations directory: %v", err)
		return
	}

	var sqlFiles []string
	for _, f := range files {
		if !f.IsDir() && strings.HasSuffix(f.Name(), ".sql") {
			sqlFiles = append(sqlFiles, f.Name())
		}
	}
	sort.Strings(sqlFiles)

	bootstrapMigrationLedger(ctx, pool, sqlFiles)

	for _, fileName := range sqlFiles {
		var alreadyApplied bool
		if err := pool.QueryRow(ctx,
			`SELECT EXISTS(SELECT 1 FROM schema_migrations WHERE filename = $1)`,
			fileName,
		).Scan(&alreadyApplied); err != nil {
			log.Printf("⚠️ Could not check migration %s: %v", fileName, err)
			continue
		}
		if alreadyApplied {
			log.Printf("Skipping already applied migration: %s", fileName)
			continue
		}

		log.Printf("Executing migration: %s", fileName)
		filePath := filepath.Join(migrationsDir, fileName)
		content, err := os.ReadFile(filePath)
		if err != nil {
			log.Printf("⚠️ Could not read migration file %s: %v", fileName, err)
			continue
		}

		tx, err := pool.Begin(ctx)
		if err != nil {
			log.Printf("⚠️ Could not begin transaction for %s: %v", fileName, err)
			continue
		}

		if _, err = tx.Exec(ctx, string(content)); err != nil {
			_ = tx.Rollback(ctx)
			log.Printf("❌ Migration %s failed and was rolled back: %v", fileName, err)
			continue
		}

		if _, err = tx.Exec(ctx,
			`INSERT INTO schema_migrations (filename) VALUES ($1)`,
			fileName,
		); err != nil {
			_ = tx.Rollback(ctx)
			log.Printf("❌ Could not record migration %s: %v", fileName, err)
			continue
		}

		if err = tx.Commit(ctx); err != nil {
			log.Printf("❌ Could not commit migration %s: %v", fileName, err)
		} else {
			log.Printf("✅ Migration %s completed successfully", fileName)
		}
	}
}

// bootstrapMigrationLedger marks all migrations as applied when the DB already has the
// redesigned pricing_plans schema (prevents re-running DROP migrations on restart).
func bootstrapMigrationLedger(ctx context.Context, pool *pgxpool.Pool, sqlFiles []string) {
	var recorded int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM schema_migrations`).Scan(&recorded); err != nil || recorded > 0 {
		return
	}

	var hasRedesignedPlans bool
	err := pool.QueryRow(ctx, `
		SELECT EXISTS (
			SELECT 1 FROM information_schema.columns
			WHERE table_schema = 'public'
			  AND table_name = 'pricing_plans'
			  AND column_name = 'pricing_plan_id'
		)
	`).Scan(&hasRedesignedPlans)
	if err != nil || !hasRedesignedPlans {
		return
	}

	for _, fileName := range sqlFiles {
		if _, err := pool.Exec(ctx,
			`INSERT INTO schema_migrations (filename) VALUES ($1) ON CONFLICT DO NOTHING`,
			fileName,
		); err != nil {
			log.Printf("⚠️ Could not bootstrap migration record %s: %v", fileName, err)
			return
		}
	}
	log.Println("✅ Bootstrapped schema_migrations for existing database (skipped re-applying destructive migrations)")
}

func (m *DatabaseManager) Close() {
	if m.Pool != nil {
		m.Pool.Close()
	}
}
