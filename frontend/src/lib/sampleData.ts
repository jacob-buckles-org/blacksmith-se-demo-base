import type { MetricPoint } from './types'

const SERVICES = ['ingest-api', 'query-api', 'billing', 'alerting', 'exporter'] as const

/**
 * Deterministic 24h of per-service sample metrics. Seeded PRNG so unit
 * tests, e2e snapshots, and the offline dashboard all agree.
 */
export function sampleMetrics(seed = 42): MetricPoint[] {
  let state = seed
  const rand = () => {
    state = (state * 1664525 + 1013904223) % 4294967296
    return state / 4294967296
  }

  const points: MetricPoint[] = []
  for (let hour = 0; hour < 24; hour++) {
    for (const service of SERVICES) {
      const base = service === 'ingest-api' ? 90_000 : service === 'query-api' ? 40_000 : 8_000
      const requests = Math.floor(base * (0.7 + rand() * 0.6))
      const errors = Math.floor(requests * rand() * 0.012)
      const p95LatencyMs = Math.round(
        (service === 'billing' ? 220 : 80) * (0.8 + rand() * (hour >= 18 ? 1.4 : 0.5)),
      )
      points.push({
        timestamp: `2026-07-15T${String(hour).padStart(2, '0')}:00:00Z`,
        service,
        requests,
        errors,
        p95LatencyMs,
      })
    }
  }
  return points
}
