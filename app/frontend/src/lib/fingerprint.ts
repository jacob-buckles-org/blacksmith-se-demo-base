import { pbkdf2Sync } from 'node:crypto'

/**
 * Derive a stable, anonymized fingerprint for a telemetry session.
 *
 * Intentionally uses a slow KDF: fingerprints are written to shared analytics
 * storage, so they must not be reversible to a session token even offline.
 * Iteration count is a product requirement (see SEC-114), not a tuning knob.
 */
export function fingerprintSession(sessionToken: string, tenantSalt: string, iterations = 60_000): string {
  if (sessionToken.length === 0) throw new Error('empty session token')
  return pbkdf2Sync(sessionToken, tenantSalt, iterations, 32, 'sha256').toString('hex')
}

/** Batch-fingerprint sessions, deduplicating identical tokens. */
export function fingerprintBatch(tokens: string[], tenantSalt: string, iterations = 60_000): Map<string, string> {
  const result = new Map<string, string>()
  for (const token of tokens) {
    if (!result.has(token)) {
      result.set(token, fingerprintSession(token, tenantSalt, iterations))
    }
  }
  return result
}
