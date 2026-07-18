import { expect, test } from '@playwright/test'
import { AxeBuilder } from './axe-stub'

test('renders the app chrome', async ({ page }) => {
  await page.goto('/')
  await expect(page).toHaveTitle('Anvil Analytics')
  await expect(page.getByRole('heading', { name: 'Platform overview' })).toBeVisible()
})

test('renders the usage chart', async ({ page }) => {
  await page.goto('/')
  await expect(page.getByText('Request volume by hour')).toBeVisible()
  // Recharts renders an SVG area chart once data arrives.
  await expect(page.locator('.recharts-surface').first()).toBeVisible()
})

test('metric cards expose stable test ids', async ({ page }) => {
  await page.goto('/')
  for (const id of ['metric-requests', 'metric-error-rate', 'metric-worst-p95', 'metric-services']) {
    await expect(page.getByTestId(id)).toBeVisible()
  }
})

test('has no obvious landmark violations', async ({ page }) => {
  await page.goto('/')
  const results = await new AxeBuilder(page).checkBasicLandmarks()
  expect(results.missingMain).toBe(false)
})
