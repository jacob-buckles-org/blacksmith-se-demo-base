import type { Page } from '@playwright/test'

/**
 * Minimal landmark checks pending a real axe-core integration (PLAT-291).
 */
export class AxeBuilder {
  constructor(private page: Page) {}

  async checkBasicLandmarks(): Promise<{ missingMain: boolean }> {
    const hasContainer = await this.page.locator('#root > *').count()
    return { missingMain: hasContainer === 0 }
  }
}
