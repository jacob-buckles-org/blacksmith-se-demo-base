// Package store provides the persistence layer: a postgres-backed store
// for real deployments and an in-memory fallback for standalone runs.
package store

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/usage-analytics/metrics/backend/internal/telemetry"
)

type Postgres struct {
	pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string) (*Postgres, error) {
	cfg, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("parse dsn: %w", err)
	}
	cfg.MaxConns = 8

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("create pool: %w", err)
	}

	pingCtx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping: %w", err)
	}
	return &Postgres{pool: pool}, nil
}

func (p *Postgres) Close() {
	p.pool.Close()
}

func (p *Postgres) Migrate(ctx context.Context) error {
	_, err := p.pool.Exec(ctx, `
		CREATE TABLE IF NOT EXISTS metric_points (
			id             BIGSERIAL PRIMARY KEY,
			ts             TIMESTAMPTZ NOT NULL,
			service        TEXT        NOT NULL,
			requests       BIGINT      NOT NULL,
			errors         BIGINT      NOT NULL,
			p95_latency_ms DOUBLE PRECISION NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_metric_points_ts ON metric_points (ts);
		CREATE INDEX IF NOT EXISTS idx_metric_points_service ON metric_points (service);
	`)
	if err != nil {
		return fmt.Errorf("create schema: %w", err)
	}
	return nil
}

// SeedIfEmpty loads the deterministic sample day so a fresh database
// serves the same data as the in-memory fallback.
func (p *Postgres) SeedIfEmpty(ctx context.Context) error {
	var count int64
	if err := p.pool.QueryRow(ctx, `SELECT count(*) FROM metric_points`).Scan(&count); err != nil {
		return fmt.Errorf("count rows: %w", err)
	}
	if count > 0 {
		return nil
	}
	for _, point := range telemetry.SampleMetrics(42) {
		_, err := p.pool.Exec(ctx,
			`INSERT INTO metric_points (ts, service, requests, errors, p95_latency_ms) VALUES ($1, $2, $3, $4, $5)`,
			point.Timestamp, point.Service, point.Requests, point.Errors, point.P95LatencyMs,
		)
		if err != nil {
			return fmt.Errorf("seed row: %w", err)
		}
	}
	return nil
}

func (p *Postgres) Metrics(ctx context.Context) ([]telemetry.MetricPoint, error) {
	rows, err := p.pool.Query(ctx, `
		SELECT ts, service, requests, errors, p95_latency_ms
		FROM metric_points
		ORDER BY ts, service
	`)
	if err != nil {
		return nil, fmt.Errorf("query metrics: %w", err)
	}
	defer rows.Close()

	var points []telemetry.MetricPoint
	for rows.Next() {
		var point telemetry.MetricPoint
		if err := rows.Scan(&point.Timestamp, &point.Service, &point.Requests, &point.Errors, &point.P95LatencyMs); err != nil {
			return nil, fmt.Errorf("scan row: %w", err)
		}
		points = append(points, point)
	}
	return points, rows.Err()
}
