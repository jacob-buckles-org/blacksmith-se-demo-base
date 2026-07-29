## Notes

Codesmith needs ~10+ workflow runs before it is ok making a recommendation
— this is why `se-demo-app`'s workflows carry `schedule:` triggers now:
history accumulates on its own between demos instead of needing a live
warmup burst.

Historical Codesmith recommendations seen on the old pool setup (for
reference — the mis-sized jobs in `workflows/current/` were picked to
reproduce this shape):

- Integration Tests / Backend integration: 4vCPU -> 8vCPU
- Docker / Build backend image: 4vCPU -> 8vCPU
- CI / Backend unit tests: 4vCPU -> 8vCPU

Low Confidence recs:
- CI / Frontend: 4vCPU -> 2 vCPU (×3 — matrix legs, not distinct recs)

## Resolved by the two-permanent-repos redesign

See `/Users/jacobbuckles/.claude/plans/cheeky-tumbling-bear.md` for the
full rationale.

- "Should reset also clear Blacksmith's cache" — moot, nothing resets.
- "cache: false feels fake" — moot, `workflows/unmigrated/` now has a
  real, honest `actions/cache`/`type=gha` baseline instead of no caching.
- July 22: migration wizard ran slower on a live test, chalked up to
  transient network — moot, the wizard is no longer run live in front of
  a prospect (narrated from the standing draft PR instead).

## Still open

Is there a good way to track the evolution of the improvements as we
onboard onto the platform? Maybe Codesmith can tell me? Not a demo-repo
question anymore, but worth asking Blacksmith/Codesmith independently —
real prospects will ask this.

Could Codesmith guide me through onboarding new features?

E2E OOM-flake reliability is unvalidated live (2026-07-29) — the design
(3 browser projects forced parallel via `workers: 3` in
`app/frontend/playwright.config.ts`, on a `blacksmith-2vcpu-ubuntu-2404`
runner) is a real resource constraint, not scripted, but hasn't actually
been observed failing yet. Watch the next several scheduled/triggered
runs. If it doesn't reproduce reliably, next iteration: add a heavier
in-memory fixture to the E2E test to increase memory pressure
predictably rather than relying on incidental browser-process overhead.

`frontend-checks` was normalized from `blacksmith-8vcpu` down to the fair
`blacksmith-4vcpu-ubuntu-2404` baseline on 2026-07-29 (was deliberately
oversized to manufacture a "scale down" Codesmith rec — Jacob's call:
that undercut the honest-baseline premise). Every CI/Docker/Integration
job now sits at that same 4 vCPU baseline except `e2e` (2 vCPU,
intentionally undersized). Worth watching whether Codesmith organically
flags `backend-test`/`docker-build`/`backend-integration` for scale-up
once real history accumulates there — the historical recs logged above
already showed exactly that shape at this same 4 vCPU baseline.

Found and fixed a real bug 2026-07-29: `workflows/current/docker.yml`'s
`build-backend`/`build-frontend` jobs were missing
`useblacksmith/setup-docker-builder@v1` (present in the original feature
patch, dropped when this file got rewritten during the redesign) —
confirmed live via a "Not using a Blacksmith builder ... Build metrics
will not be reported" warning, meaning zero Blacksmith cache benefit was
active in any se-demo-app Docker run before the fix. Fixed; the first
post-fix run correctly showed a fresh sticky disk being created (`Getting
sticky disk for jacob-buckles-org/se-demo-app` → `Successfully obtained
sticky disk`), so this was a cold build as expected. Watch the *next*
scheduled/triggered Docker run — it should reuse that sticky disk and
show a real speedup (faster base-image layer restore, faster `go
build`/`npm ci`). If it doesn't, the sticky-disk caching itself needs
investigating, not just the builder wiring.

