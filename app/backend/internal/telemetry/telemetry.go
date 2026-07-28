// Package telemetry holds the shared domain types and aggregation logic
// for the metrics pipeline. The JSON field names are the public API
// contract with the dashboard.
package telemetry

import (
	"sort"
	"time"
)

type MetricPoint struct {
	Timestamp    time.Time `json:"timestamp"`
	Service      string    `json:"service"`
	Requests     int64     `json:"requests"`
	Errors       int64     `json:"errors"`
	P95LatencyMs float64   `json:"p95LatencyMs"`
}

type ServiceSummary struct {
	Service         string  `json:"service"`
	TotalRequests   int64   `json:"totalRequests"`
	ErrorRate       float64 `json:"errorRate"`
	AvgP95LatencyMs float64 `json:"avgP95LatencyMs"`
}

// SummarizeByService rolls raw points up into one row per service,
// sorted by request volume descending.
func SummarizeByService(points []MetricPoint) []ServiceSummary {
	type acc struct {
		requests int64
		errors   int64
		p95Sum   float64
		count    int64
	}
	byService := map[string]*acc{}
	for _, p := range points {
		a, ok := byService[p.Service]
		if !ok {
			a = &acc{}
			byService[p.Service] = a
		}
		a.requests += p.Requests
		a.errors += p.Errors
		a.p95Sum += p.P95LatencyMs
		a.count++
	}

	summaries := make([]ServiceSummary, 0, len(byService))
	for service, a := range byService {
		errorRate := 0.0
		if a.requests > 0 {
			errorRate = float64(a.errors) / float64(a.requests)
		}
		summaries = append(summaries, ServiceSummary{
			Service:         service,
			TotalRequests:   a.requests,
			ErrorRate:       errorRate,
			AvgP95LatencyMs: a.p95Sum / float64(a.count),
		})
	}
	sort.Slice(summaries, func(i, j int) bool {
		if summaries[i].TotalRequests != summaries[j].TotalRequests {
			return summaries[i].TotalRequests > summaries[j].TotalRequests
		}
		return summaries[i].Service < summaries[j].Service
	})
	return summaries
}

// SampleMetrics synthesizes a deterministic 24h of per-service metrics.
// Mirrors the frontend's sample generator so standalone deployments of
// either half agree with each other.
func SampleMetrics(seed uint32) []MetricPoint {
	services := []string{"ingest-api", "query-api", "billing", "alerting", "exporter"}
	state := seed
	rand := func() float64 {
		state = state*1664525 + 1013904223
		return float64(state) / 4294967296.0
	}

	day := time.Date(2026, 7, 15, 0, 0, 0, 0, time.UTC)
	var points []MetricPoint
	for hour := 0; hour < 24; hour++ {
		for _, service := range services {
			base := int64(8_000)
			switch service {
			case "ingest-api":
				base = 90_000
			case "query-api":
				base = 40_000
			}
			requests := int64(float64(base) * (0.7 + rand()*0.6))
			errors := int64(float64(requests) * rand() * 0.012)
			latBase := 80.0
			if service == "billing" {
				latBase = 220.0
			}
			spread := 0.5
			if hour >= 18 {
				spread = 1.4
			}
			points = append(points, MetricPoint{
				Timestamp:    day.Add(time.Duration(hour) * time.Hour),
				Service:      service,
				Requests:     requests,
				Errors:       errors,
				P95LatencyMs: float64(int(latBase * (0.8 + rand()*spread))),
			})
		}
	}
	return points
}
