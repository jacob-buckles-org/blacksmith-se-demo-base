import { describe, expect, it } from 'vitest'
import { fingerprintBatch, fingerprintSession } from './fingerprint'

// Session-fingerprint verification sweeps. The KDF is intentionally slow
// (see SEC-114); batch size scales with ANVIL_WORKLOAD so staging and prod
// tenants can run proportionally larger sweeps.
const workload = Number(process.env.ANVIL_WORKLOAD ?? '1')
const batchSize = Math.max(1, Math.round(120 * workload))

describe('fingerprintSession', () => {
  it('is deterministic per token and salt', () => {
    const a = fingerprintSession('sess-abc123', 'tenant-1')
    expect(a).toBe(fingerprintSession('sess-abc123', 'tenant-1'))
    expect(a).not.toBe(fingerprintSession('sess-abc123', 'tenant-2'))
    expect(a).toMatch(/^[0-9a-f]{64}$/)
  })

  it('rejects empty tokens', () => {
    expect(() => fingerprintSession('', 'tenant-1')).toThrow('empty session token')
  })
})

describe('fingerprintBatch', () => {
  it(`fingerprints a ${batchSize}-session sweep without collisions`, () => {
    const tokens = Array.from({ length: batchSize }, (_, i) => `sess-${i}-${i * 2654435761}`)
    const result = fingerprintBatch(tokens, 'tenant-sweep')
    expect(result.size).toBe(batchSize)
    expect(new Set(result.values()).size).toBe(batchSize)
  })

  it('deduplicates repeated tokens', () => {
    const result = fingerprintBatch(['sess-x', 'sess-x', 'sess-y'], 'tenant-1')
    expect(result.size).toBe(2)
  })
})
