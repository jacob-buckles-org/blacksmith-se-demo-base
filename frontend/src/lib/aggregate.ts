import type { MetricPoint, OverviewStats, ServiceSummary } from './types'

/** Roll raw metric points up into one row per service. */
export function summarizeByService(points: MetricPoint[]): ServiceSummary[] {
  const byService = new Map<string, MetricPoint[]>()
  for (const point of points) {
    const bucket = byService.get(point.service)
    if (bucket) {
      bucket.push(point)
    } else {
      byService.set(point.service, [point])
    }
  }

  const summaries: ServiceSummary[] = []
  for (const [service, bucket] of byService) {
    const totalRequests = bucket.reduce((sum, p) => sum + p.requests, 0)
    const totalErrors = bucket.reduce((sum, p) => sum + p.errors, 0)
    const avgP95 = bucket.reduce((sum, p) => sum + p.p95LatencyMs, 0) / bucket.length
    summaries.push({
      service,
      totalRequests,
      errorRate: totalRequests === 0 ? 0 : totalErrors / totalRequests,
      avgP95LatencyMs: avgP95,
    })
  }

  return summaries.sort((a, b) => b.totalRequests - a.totalRequests)
}

/** Headline numbers for the top of the dashboard. */
export function computeOverview(points: MetricPoint[]): OverviewStats {
  const summaries = summarizeByService(points)
  const totalRequests = summaries.reduce((sum, s) => sum + s.totalRequests, 0)
  const totalErrors = points.reduce((sum, p) => sum + p.errors, 0)
  return {
    totalRequests,
    overallErrorRate: totalRequests === 0 ? 0 : totalErrors / totalRequests,
    worstP95Ms: points.reduce((worst, p) => Math.max(worst, p.p95LatencyMs), 0),
    serviceCount: summaries.length,
  }
}

/** Bucket points into fixed-width time windows for charting. */
export function bucketByHour(points: MetricPoint[]): { hour: string; requests: number; errors: number }[] {
  const byHour = new Map<string, { requests: number; errors: number }>()
  for (const point of points) {
    const hour = point.timestamp.slice(0, 13)
    const bucket = byHour.get(hour) ?? { requests: 0, errors: 0 }
    bucket.requests += point.requests
    bucket.errors += point.errors
    byHour.set(hour, bucket)
  }
  return [...byHour.entries()]
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([hour, agg]) => ({ hour: `${hour}:00`, ...agg }))
}
