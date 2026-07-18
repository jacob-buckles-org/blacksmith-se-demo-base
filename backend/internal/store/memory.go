package store

import (
	"context"

	"github.com/anvil-analytics/anvil/backend/internal/telemetry"
)

// MemoryStore serves the deterministic sample day; used when no
// DATABASE_URL is configured (local dev, docker previews).
type MemoryStore struct {
	points []telemetry.MetricPoint
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{points: telemetry.SampleMetrics(42)}
}

func (m *MemoryStore) Metrics(context.Context) ([]telemetry.MetricPoint, error) {
	return m.points, nil
}
