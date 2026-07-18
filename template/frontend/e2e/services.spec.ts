import { expect, test } from '@playwright/test'

const apiPoints = [
  { timestamp: '2026-07-15T09:00:00Z', service: 'ingest-api', requests: 200_000, errors: 400, p95LatencyMs: 95 },
  { timestamp: '2026-07-15T09:00:00Z', service: 'query-api', requests: 90_000, errors: 45, p95LatencyMs: 130 },
  { timestamp: '2026-07-15T09:00:00Z', service: 'billing', requests: 8_000, errors: 160, p95LatencyMs: 420 },
  { timestamp: '2026-07-15T10:00:00Z', service: 'billing', requests: 7_500, errors: 12, p95LatencyMs: 380 },
]

test.beforeEach(async ({ page }) => {
  await page.route('**/api/metrics', (route) => route.fulfill({ json: apiPoints }))
})

test('formats request volumes in compact units', async ({ page }) => {
  await page.goto('/')
  const rows = page.getByTestId('service-table').locator('tbody tr')
  await expect(rows.filter({ hasText: 'ingest-api' })).toContainText('200K')
})

test('shows per-service error rates', async ({ page }) => {
  await page.goto('/')
  const billing = page.getByTestId('service-table').locator('tbody tr').filter({ hasText: 'billing' })
  // (160 + 12) / (8000 + 7500) = 1.11%
  await expect(billing).toContainText('1.11%')
})

test('promotes sub-second latencies to ms and above to seconds', async ({ page }) => {
  await page.goto('/')
  const billing = page.getByTestId('service-table').locator('tbody tr').filter({ hasText: 'billing' })
  await expect(billing).toContainText('400ms')
})

test('headline error rate aggregates all services', async ({ page }) => {
  await page.goto('/')
  // 617 errors / 305500 requests = 0.20%
  await expect(page.getByTestId('metric-error-rate')).toHaveText('0.20%')
})

test('worst p95 reflects the slowest service', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByTestId('metric-worst-p95')).toHaveText('420ms')
})
