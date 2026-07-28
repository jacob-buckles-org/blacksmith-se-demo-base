package telemetry

import (
	"testing"
	"time"
)

func point(service string, requests, errors int64, p95 float64) MetricPoint {
	return MetricPoint{
		Timestamp:    time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC),
		Service:      service,
		Requests:     requests,
		Errors:       errors,
		P95LatencyMs: p95,
	}
}

func TestSummarizeByService(t *testing.T) {
	points := []MetricPoint{
		point("ingest-api", 100, 2, 80),
		point("ingest-api", 300, 0, 120),
		point("billing", 50, 5, 400),
	}
	summaries := SummarizeByService(points)

	if len(summaries) != 2 {
		t.Fatalf("expected 2 summaries, got %d", len(summaries))
	}
	if summaries[0].Service != "ingest-api" {
		t.Errorf("expected ingest-api first (highest volume), got %s", summaries[0].Service)
	}
	if summaries[0].TotalRequests != 400 {
		t.Errorf("expected 400 requests, got %d", summaries[0].TotalRequests)
	}
	if got, want := summaries[0].ErrorRate, 2.0/400.0; got != want {
		t.Errorf("expected error rate %v, got %v", want, got)
	}
	if summaries[0].AvgP95LatencyMs != 100 {
		t.Errorf("expected avg p95 100, got %v", summaries[0].AvgP95LatencyMs)
	}
}

func TestSummarizeByServiceZeroRequests(t *testing.T) {
	summaries := SummarizeByService([]MetricPoint{point("idle", 0, 0, 10)})
	if summaries[0].ErrorRate != 0 {
		t.Errorf("zero-request service must not divide by zero, got %v", summaries[0].ErrorRate)
	}
}

func TestSummarizeByServiceEmpty(t *testing.T) {
	if got := SummarizeByService(nil); len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestSampleMetricsDeterministic(t *testing.T) {
	a := SampleMetrics(7)
	b := SampleMetrics(7)
	if len(a) != 24*5 {
		t.Fatalf("expected 120 points, got %d", len(a))
	}
	for i := range a {
		if a[i] != b[i] {
			t.Fatalf("sample metrics not deterministic at index %d: %+v != %+v", i, a[i], b[i])
		}
	}
}
