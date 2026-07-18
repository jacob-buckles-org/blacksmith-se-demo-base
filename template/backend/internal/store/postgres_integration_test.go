//go:build integration

package store

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/anvil-analytics/anvil/backend/internal/telemetry"
)

// Integration tests require a live postgres; run with:
//
//	DATABASE_URL=postgres://... go test -tags integration ./internal/store/
func connectTestDB(t *testing.T) *Postgres {
	t.Helper()
	dsn := os.Getenv("DATABASE_URL")
	if dsn == "" {
		t.Skip("DATABASE_URL not set; skipping integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	pg, err := Connect(ctx, dsn)
	if err != nil {
		t.Fatalf("connect: %v", err)
	}
	t.Cleanup(pg.Close)

	if _, err := pg.pool.Exec(ctx, `DROP TABLE IF EXISTS metric_points`); err != nil {
		t.Fatalf("reset schema: %v", err)
	}
	if err := pg.Migrate(ctx); err != nil {
		t.Fatalf("migrate: %v", err)
	}
	return pg
}

func TestMigrateAndSeed(t *testing.T) {
	pg := connectTestDB(t)
	ctx := context.Background()

	if err := pg.SeedIfEmpty(ctx); err != nil {
		t.Fatalf("seed: %v", err)
	}

	points, err := pg.Metrics(ctx)
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if len(points) != 120 {
		t.Errorf("expected 120 seeded points, got %d", len(points))
	}

	// Seeding must be idempotent.
	if err := pg.SeedIfEmpty(ctx); err != nil {
		t.Fatalf("re-seed: %v", err)
	}
	points, err = pg.Metrics(ctx)
	if err != nil {
		t.Fatalf("metrics after re-seed: %v", err)
	}
	if len(points) != 120 {
		t.Errorf("re-seed must be a no-op, got %d points", len(points))
	}
}

func TestSummariesFromPostgres(t *testing.T) {
	pg := connectTestDB(t)
	ctx := context.Background()

	if err := pg.SeedIfEmpty(ctx); err != nil {
		t.Fatalf("seed: %v", err)
	}
	points, err := pg.Metrics(ctx)
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}

	summaries := telemetry.SummarizeByService(points)
	if len(summaries) != 5 {
		t.Fatalf("expected 5 services, got %d", len(summaries))
	}
	if summaries[0].Service != "ingest-api" {
		t.Errorf("expected ingest-api to have the highest volume, got %s", summaries[0].Service)
	}
	for _, s := range summaries {
		if s.ErrorRate < 0 || s.ErrorRate > 0.05 {
			t.Errorf("service %s error rate %v outside sane bounds", s.Service, s.ErrorRate)
		}
	}
}
