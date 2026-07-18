const UNITS = ['', 'K', 'M', 'B'] as const

/** Format a request count for display, e.g. 1_284_311 -> "1.3M". */
export function formatCount(value: number): string {
  if (!Number.isFinite(value) || value < 0) return '–'
  let scaled = value
  let unit = 0
  while (scaled >= 1000 && unit < UNITS.length - 1) {
    scaled /= 1000
    unit += 1
  }
  const digits = scaled >= 100 || unit === 0 ? 0 : 1
  return `${scaled.toFixed(digits)}${UNITS[unit]}`
}

/** Format a 0..1 rate as a percentage with sensible precision. */
export function formatRate(rate: number): string {
  if (!Number.isFinite(rate) || rate < 0) return '–'
  if (rate === 0) return '0%'
  if (rate < 0.0001) return '<0.01%'
  return `${(rate * 100).toFixed(2)}%`
}

/** Format milliseconds of latency, promoting to seconds above 1s. */
export function formatLatency(ms: number): string {
  if (!Number.isFinite(ms) || ms < 0) return '–'
  if (ms >= 1000) return `${(ms / 1000).toFixed(2)}s`
  return `${Math.round(ms)}ms`
}
