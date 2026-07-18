import { expect, test } from '@playwright/test'

const apiPoints = [
  { timestamp: '2026-07-15T09:00:00Z', service: 'ingest-api', requests: 120_000, errors: 240, p95LatencyMs: 95 },
  { timestamp: '2026-07-15T09:00:00Z', service: 'query-api', requests: 48_000, errors: 12, p95LatencyMs: 130 },
  { timestamp: '2026-07-15T10:00:00Z', service: 'ingest-api', requests: 131_000, errors: 190, p95LatencyMs: 101 },
]

test.describe('dashboard with live API', () => {
  test.beforeEach(async ({ page }) => {
    await page.route('**/api/metrics', (route) => route.fulfill({ json: apiPoints }))
  })

  test('renders headline metrics from the API', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByTestId('data-source')).toHaveText('live data')
    await expect(page.getByTestId('metric-requests')).toHaveText('299K')
    await expect(page.getByTestId('metric-services')).toHaveText('2')
  })

  test('lists services ordered by volume', async ({ page }) => {
    await page.goto('/')
    const rows = page.getByTestId('service-table').locator('tbody tr')
    await expect(rows).toHaveCount(2)
    await expect(rows.first()).toContainText('ingest-api')
  })
})

test.describe('dashboard when the API is down', () => {
  test('falls back to sample data instead of breaking', async ({ page }) => {
    await page.route('**/api/metrics', (route) => route.abort())
    await page.goto('/')
    await expect(page.getByTestId('data-source')).toHaveText('sample data')
    const rows = page.getByTestId('service-table').locator('tbody tr')
    await expect(rows).toHaveCount(5)
  })
})
