import { expect, test } from '@playwright/test'

// The chart only mounts once the metrics fetch resolves and the hourly
// rollup has run, so this budget has to cover fetch + aggregate + first
// paint.
const CHART_RENDER_BUDGET_MS = 400

// One hourly point per service across a full day — the shape /api/metrics
// returns for the dashboard's default "last 24 hours" window.
const services = ['ingest-api', 'query-api', 'billing', 'alerting', 'exporter']
const dayOfMetrics = services.flatMap((service, s) =>
  Array.from({ length: 24 }, (_, hour) => ({
    timestamp: `2026-07-15T${String(hour).padStart(2, '0')}:00:00Z`,
    service,
    requests: 30_000 + hour * 911 + s * 4_200,
    errors: hour * 7 + s,
    p95LatencyMs: 85 + hour + s * 3,
  })),
)

test.describe('request volume chart', () => {
  // No retries here. These are render-timing assertions, and retrying them
  // just papers over the regression they exist to catch: a retry runs against
  // a warm browser and a warm bundle cache, so it almost always passes and
  // the signal is lost. Fail loudly instead.
  test.describe.configure({ retries: 0 })

  test.beforeEach(async ({ page }) => {
    await page.route('**/api/metrics', (route) => route.fulfill({ json: dayOfMetrics }))
  })

  test('plots request and error series for the last 24h', async ({ page }) => {
    await page.goto('/')

    const chart = page.getByTestId('usage-chart')
    await expect(chart).toBeVisible({ timeout: CHART_RENDER_BUDGET_MS })

    // Recharts renders one <path class="recharts-area-area"> per <Area>:
    // one for requests, one for errors.
    await expect(chart.locator('path.recharts-area-area')).toHaveCount(2)
  })

  test('labels the chart with its reporting window', async ({ page }) => {
    await page.goto('/')
    await expect(page.getByTestId('usage-chart')).toContainText('Request volume by hour')
  })
})
