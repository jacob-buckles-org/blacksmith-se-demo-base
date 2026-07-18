// Regenerates the sample telemetry exports in this directory.
// Deterministic so re-runs produce identical files:
//
//   node data/generate.mjs
//
// These exports back the load-test suite and the docs' notebook examples.
import { writeFileSync } from 'node:fs'
import { dirname, join } from 'node:path'
import { fileURLToPath } from 'node:url'

const here = dirname(fileURLToPath(import.meta.url))
const SERVICES = ['ingest-api', 'query-api', 'billing', 'alerting', 'exporter']
const REGIONS = ['us-east-1', 'us-west-2', 'eu-central-1', 'ap-southeast-2']
const DAYS = ['2026-07-08', '2026-07-09', '2026-07-10', '2026-07-11', '2026-07-12', '2026-07-13']
const ROWS_PER_DAY = 40_000

let state = 20260708
const rand = () => {
  state = (state * 1664525 + 1013904223) % 4294967296
  return state / 4294967296
}

for (const day of DAYS) {
  const lines = ['ts,service,region,tenant,requests,errors,p50_ms,p95_ms,p99_ms,bytes_in,bytes_out']
  for (let i = 0; i < ROWS_PER_DAY; i++) {
    const hour = String(Math.floor((i / ROWS_PER_DAY) * 24)).padStart(2, '0')
    const minute = String(Math.floor(rand() * 60)).padStart(2, '0')
    const service = SERVICES[Math.floor(rand() * SERVICES.length)]
    const region = REGIONS[Math.floor(rand() * REGIONS.length)]
    const tenant = `t-${String(Math.floor(rand() * 4000)).padStart(5, '0')}`
    const requests = Math.floor(rand() * 12_000)
    const errors = Math.floor(requests * rand() * 0.015)
    const p50 = (20 + rand() * 60).toFixed(2)
    const p95 = (Number(p50) * (1.8 + rand())).toFixed(2)
    const p99 = (Number(p95) * (1.3 + rand() * 0.8)).toFixed(2)
    const bytesIn = Math.floor(requests * (400 + rand() * 2200))
    const bytesOut = Math.floor(requests * (900 + rand() * 6800))
    lines.push(
      `${day}T${hour}:${minute}:00Z,${service},${region},${tenant},${requests},${errors},${p50},${p95},${p99},${bytesIn},${bytesOut}`,
    )
  }
  writeFileSync(join(here, `telemetry-${day}.csv`), lines.join('\n') + '\n')
  console.log(`wrote telemetry-${day}.csv (${lines.length - 1} rows)`)
}
