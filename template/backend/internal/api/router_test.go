package api

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/usage-analytics/metrics/backend/internal/telemetry"
)

type stubSource struct {
	points []telemetry.MetricPoint
	err    error
}

func (s stubSource) Metrics(context.Context) ([]telemetry.MetricPoint, error) {
	return s.points, s.err
}

func TestHealthz(t *testing.T) {
	srv := httptest.NewServer(NewRouter(stubSource{}))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/healthz")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Errorf("expected 200, got %d", res.StatusCode)
	}
}

func TestMetricsEndpoint(t *testing.T) {
	srv := httptest.NewServer(NewRouter(stubSource{points: telemetry.SampleMetrics(1)}))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	var points []telemetry.MetricPoint
	if err := json.NewDecoder(res.Body).Decode(&points); err != nil {
		t.Fatal(err)
	}
	if len(points) != 120 {
		t.Errorf("expected 120 points, got %d", len(points))
	}
}

func TestServicesEndpoint(t *testing.T) {
	srv := httptest.NewServer(NewRouter(stubSource{points: telemetry.SampleMetrics(1)}))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/services")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()

	var summaries []telemetry.ServiceSummary
	if err := json.NewDecoder(res.Body).Decode(&summaries); err != nil {
		t.Fatal(err)
	}
	if len(summaries) != 5 {
		t.Errorf("expected 5 services, got %d", len(summaries))
	}
	if summaries[0].Service != "ingest-api" {
		t.Errorf("expected ingest-api first, got %s", summaries[0].Service)
	}
}

func TestMetricsEndpointSurfacesErrors(t *testing.T) {
	srv := httptest.NewServer(NewRouter(stubSource{err: errors.New("pool exhausted")}))
	defer srv.Close()

	res, err := http.Get(srv.URL + "/api/metrics")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusInternalServerError {
		t.Errorf("expected 500, got %d", res.StatusCode)
	}
}
