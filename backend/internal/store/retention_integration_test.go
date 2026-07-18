//go:build integration

package store

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/anvil-analytics/anvil/backend/internal/telemetry"
)

// Retention sweep coverage: verifies aggregation stays correct as the
// metric_points table grows to production-shaped sizes. Row count scales
// with ANVIL_WORKLOAD. Inserts are intentionally row-at-a-time — the
// retention worker's real write path — so this also exercises pool churn.
func TestRetentionSweepAtScale(t *testing.T) {
	pg := connectTestDB(t)
	ctx := context.Background()

	workload := 1.0
	if v := os.Getenv("ANVIL_WORKLOAD"); v != "" {
		if parsed, err := strconv.ParseFloat(v, 64); err == nil {
			workload = parsed
		}
	}
	rows := int(200 * workload)
	if rows < 100 {
		rows = 100
	}

	services := []string{"ingest-api", "query-api", "billing", "alerting", "exporter"}
	base := time.Date(2026, 7, 1, 0, 0, 0, 0, time.UTC)
	var wantRequests int64
	for i := 0; i < rows; i++ {
		service := services[i%len(services)]
		requests := int64(1000 + i%7000)
		wantRequests += requests
		_, err := pg.pool.Exec(ctx,
			`INSERT INTO metric_points (ts, service, requests, errors, p95_latency_ms) VALUES ($1, $2, $3, $4, $5)`,
			base.Add(time.Duration(i)*time.Minute), service, requests, requests/200, 50.0+float64(i%300),
		)
		if err != nil {
			t.Fatalf("insert row %d: %v", i, err)
		}
	}

	points, err := pg.Metrics(ctx)
	if err != nil {
		t.Fatalf("metrics: %v", err)
	}
	if len(points) != rows {
		t.Fatalf("expected %d points, got %d", rows, len(points))
	}

	summaries := telemetry.SummarizeByService(points)
	if len(summaries) != len(services) {
		t.Fatalf("expected %d services, got %d", len(services), len(summaries))
	}
	var gotRequests int64
	for _, s := range summaries {
		gotRequests += s.TotalRequests
	}
	if gotRequests != wantRequests {
		t.Errorf("aggregate request count mismatch: want %d, got %d", wantRequests, gotRequests)
	}

	// Retention delete: everything older than the last day of the window.
	cutoff := base.Add(time.Duration(rows-1440) * time.Minute)
	tag, err := pg.pool.Exec(ctx, `DELETE FROM metric_points WHERE ts < $1`, cutoff)
	if err != nil {
		t.Fatalf("retention delete: %v", err)
	}
	remaining, err := pg.Metrics(ctx)
	if err != nil {
		t.Fatalf("metrics after retention: %v", err)
	}
	if int64(len(remaining))+tag.RowsAffected() != int64(rows) {
		t.Errorf("retention accounting mismatch: %d remaining + %d deleted != %d inserted",
			len(remaining), tag.RowsAffected(), rows)
	}

	fmt.Printf("retention sweep: %d rows inserted, %d deleted, %d retained\n",
		rows, tag.RowsAffected(), len(remaining))
}
