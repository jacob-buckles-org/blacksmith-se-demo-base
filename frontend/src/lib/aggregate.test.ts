import { describe, expect, it } from 'vitest'
import { bucketByHour, computeOverview, summarizeByService } from './aggregate'
import { sampleMetrics } from './sampleData'
import type { MetricPoint } from './types'

const points: MetricPoint[] = [
  { timestamp: '2026-07-15T00:00:00Z', service: 'ingest-api', requests: 100, errors: 2, p95LatencyMs: 80 },
  { timestamp: '2026-07-15T01:00:00Z', service: 'ingest-api', requests: 300, errors: 0, p95LatencyMs: 120 },
  { timestamp: '2026-07-15T00:00:00Z', service: 'billing', requests: 50, errors: 5, p95LatencyMs: 400 },
]

describe('summarizeByService', () => {
  it('rolls points up per service, sorted by volume', () => {
    const summaries = summarizeByService(points)
    expect(summaries.map((s) => s.service)).toEqual(['ingest-api', 'billing'])
    expect(summaries[0].totalRequests).toBe(400)
    expect(summaries[0].errorRate).toBeCloseTo(2 / 400)
    expect(summaries[0].avgP95LatencyMs).toBe(100)
  })

  it('handles zero-request services without dividing by zero', () => {
    const summaries = summarizeByService([
      { timestamp: '2026-07-15T00:00:00Z', service: 'idle', requests: 0, errors: 0, p95LatencyMs: 10 },
    ])
    expect(summaries[0].errorRate).toBe(0)
  })

  it('returns empty for no data', () => {
    expect(summarizeByService([])).toEqual([])
  })
})

describe('computeOverview', () => {
  it('computes headline stats', () => {
    const overview = computeOverview(points)
    expect(overview.totalRequests).toBe(450)
    expect(overview.overallErrorRate).toBeCloseTo(7 / 450)
    expect(overview.worstP95Ms).toBe(400)
    expect(overview.serviceCount).toBe(2)
  })
})

describe('bucketByHour', () => {
  it('merges services within an hour and sorts chronologically', () => {
    const buckets = bucketByHour(points)
    expect(buckets).toHaveLength(2)
    expect(buckets[0]).toEqual({ hour: '2026-07-15T00:00', requests: 150, errors: 7 })
  })
})

describe('sampleMetrics', () => {
  it('is deterministic for a given seed', () => {
    expect(sampleMetrics(7)).toEqual(sampleMetrics(7))
  })

  it('covers 24 hours of five services', () => {
    expect(sampleMetrics()).toHaveLength(24 * 5)
  })
})
