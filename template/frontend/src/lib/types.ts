export interface MetricPoint {
  timestamp: string
  service: string
  requests: number
  errors: number
  p95LatencyMs: number
}

export interface ServiceSummary {
  service: string
  totalRequests: number
  errorRate: number
  avgP95LatencyMs: number
}

export interface OverviewStats {
  totalRequests: number
  overallErrorRate: number
  worstP95Ms: number
  serviceCount: number
}
