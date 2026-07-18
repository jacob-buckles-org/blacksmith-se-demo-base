import type { MetricPoint } from './types'
import { sampleMetrics } from './sampleData'

/**
 * Fetch raw metric points from the backend. Falls back to bundled sample
 * data when the API is unreachable so the dashboard renders standalone
 * (local dev without the Go service, docker previews, demos).
 */
export async function fetchMetrics(): Promise<{ points: MetricPoint[]; live: boolean }> {
  try {
    const res = await fetch('/api/metrics')
    if (!res.ok) throw new Error(`api responded ${res.status}`)
    const points = (await res.json()) as MetricPoint[]
    return { points, live: true }
  } catch {
    return { points: sampleMetrics(), live: false }
  }
}
