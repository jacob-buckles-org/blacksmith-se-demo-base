// Package api exposes the HTTP surface consumed by the dashboard.
package api

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"github.com/anvil-analytics/anvil/backend/internal/telemetry"
)

// MetricsSource is the read-side contract the API serves from; satisfied
// by both the postgres store and the in-memory fallback.
type MetricsSource interface {
	Metrics(ctx context.Context) ([]telemetry.MetricPoint, error)
}

func NewRouter(source MetricsSource) http.Handler {
	r := chi.NewRouter()
	r.Use(middleware.RequestID)
	r.Use(middleware.Recoverer)

	r.Get("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})

	r.Get("/api/metrics", func(w http.ResponseWriter, req *http.Request) {
		points, err := source.Metrics(req.Context())
		if err != nil {
			log.Printf("metrics query failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "metrics unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, points)
	})

	r.Get("/api/services", func(w http.ResponseWriter, req *http.Request) {
		points, err := source.Metrics(req.Context())
		if err != nil {
			log.Printf("metrics query failed: %v", err)
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "metrics unavailable"})
			return
		}
		writeJSON(w, http.StatusOK, telemetry.SummarizeByService(points))
	})

	return r
}

func writeJSON(w http.ResponseWriter, status int, body any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(body)
}
