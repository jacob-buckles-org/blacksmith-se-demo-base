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

Container caching support ticket filed 2026-07-31. Measured baselines that
same day confirm it was **not yet active**, and they double as the way to
detect when it lands (there's no dashboard toggle to check):

- `Initialize containers` (postgres:16 service), 10 runs each — se-demo-app
  ~10s vs unmigrated ~10s, i.e. **identical**, so no caching in effect.
- `Bring up the stack` (compose, also pulls golang:1.24 + node:22) —
  se-demo-app 34-44s vs unmigrated 66-80s.

Two things to keep straight about those numbers:

- The ~2x compose delta is **not** container caching (it wasn't on). Both
  runners are 4 vCPU / 16 GB — public-repo `ubuntu-latest` matches
  `blacksmith-4vcpu-ubuntu-2404` on paper — so that gap is CPU generation,
  disk I/O and network throughput. Which is a useful data point in its own
  right: it's clean evidence for the "hardware, not just core count" claim
  in the honest-baseline section.
- The postgres-only comparison can never be dramatic: unmigrated is already
  ~10s, so a perfect cache saves ~7-8s. The docs' "minutes to seconds"
  framing assumes much heavier service-container setups. The runbook was
  reworded to lead with the compose job instead.

## Still open

Is there a good way to track the evolution of the improvements as we
onboard onto the platform? Maybe Codesmith can tell me? Not a demo-repo
question anymore, but worth asking Blacksmith/Codesmith independently —
real prospects will ask this.

Could Codesmith guide me through onboarding new features?

~~E2E OOM-flake reliability is unvalidated live (2026-07-29)~~ —
**abandoned 2026-07-31 in favour of a timing race.** Reasons the OOM
approach was a bad bet: the runner has 8 GB (`blacksmith-2vcpu-ubuntu-2404`),
so 3 browsers at ~300-500MB each never got close to the ceiling without a
large synthetic allocation; the kernel OOM killer's trigger timing and
victim choice depend on cgroup/kernel details we don't control and that
can shift under us with no code change; and the failure signature is
inconsistent (killed worker / "Target closed" / segfault), which weakens
the "search the logs for the OOM" beat because there's no one greppable
string.

Replaced by a **timing race** in `app/frontend/e2e/chart.spec.ts`: assert
the chart is visible within `CHART_RENDER_BUDGET_MS` after `goto`. Reads
as an ordinary under-specified wait (the most common real-world flake),
one numeric knob, and it's fixed by *more CPU* — which is exactly what
Codesmith's rightsizing recommendation offers, so the causal
"rightsize → flake stops" beat is more direct than it was with memory.

Calibration data, 2026-07-31 (15 samples = 5 runs x 3 browsers, measuring
goto-return to chart-visible on the 2vCPU runner):

    chromium: 159 165 178 178 191
    firefox:  216 240 302 525 582
    webkit:   242 312 467 478 488

Bimodal: chromium is always fast; firefox/webkit split into a fast group
(216-312ms) and a contended group (467-582ms) depending on how much the
three browser workers collide that run. There's a clean 155ms gap between
312 and 467, so the budget is set to **400ms** — inside the gap, where the
per-attempt failure rate (~33%, i.e. 5/15) is stable against small
distribution shifts. Sitting at e.g. 500ms would be only 12ms above an
observed sample and therefore fragile.

**Verified 2026-07-31 over 8 live runs at 400ms: 1/8 runs (12%) recorded a
flaky test, 0/8 red.** The flake was
`[webkit] › chart.spec.ts › plots request and error series` — failed once,
passed on retry, job stayed green. Exactly the intended profile.

Two caveats worth knowing before touching this knob again:

1. **12% is below the 15-30% that was aimed for, and tightening the budget
   can't fix it.** The measured distribution has a 155ms gap (312→467), so
   every budget from ~320ms to ~460ms is behaviourally *identical* — same
   5/15 samples over the line. The next step down (below ~312ms) jumps to
   7/15 = 47% per attempt, which would start producing genuinely red runs.
   There's no setting that yields a modest increase; it's ~12% or ~47%.
   400ms sits mid-gap, which is why it's stable.
2. The verification runs came out **faster** than the calibration runs
   (per-attempt ~4% vs the 33% the calibration sample implied), i.e.
   host-level contention varies enough between bursts that 15 samples
   wasn't sufficient to pin the rate. Treat any single measurement as
   ±10pp. Re-read the true rate off accumulated scheduled history rather
   than another burst.

If a higher flakiness % is wanted later, the better lever is more render
work (a larger mocked payload / more services shifts the whole
distribution later and widens it) rather than a tighter timeout, which
just falls off the bimodal cliff.

**Retries disabled 2026-07-31 — the "flaky-but-green" idea did not work.**
Jacob noticed there were no signs of test failures anywhere, which turned
out to be a real gap, not just bad luck. Playwright's JUnit reporter writes
only a test's **final** status. Probed it directly with a test that fails
attempt 0 and passes on retry:

    <testsuites tests="1" failures="0" ...>
      <testcase name="probe: fails first attempt, passes on retry" time="0">
        <system-out>[[ATTACHMENT|...retry1/trace.zip]]</system-out>
      </testcase>

`failures="0"`, no `<failure>` element — the only residue is an attachment
filename. So anything reading the XML (i.e. Test Analytics) saw a clean
pass, and the whole flaky-test demo had nothing to show. The earlier
reasoning that "retries give us flaky-but-green, best of both worlds"
assumed the retry would be *recorded*; it isn't.

Fix: `test.describe.configure({ retries: 0 })` on that spec only. Verified
a real failure then yields `failures="1"` plus a full
`<failure message="expect(locator).toBeVisible() failed"
type="expect.toBeVisible">` with error, locator, call log and code frame.
Consequence: accepted deliberately — flakiness % now comes from cross-run
history rather than within-run retries. Measured rate: 1/8 runs with retries
on, 1/15 with retries off, i.e. ~9% across 23 runs (~1 red run in 10). The
confirmed red run is 30657538736 (webkit, `1 failed`, no `retry #` line).

Note the job log *did* always contain the evidence (`✘`, `retry #1`, the
assertion error, `1 flaky`), so log-parsing could in principle have caught
it — but that was never confirmed and depending on it was fragile. JUnit is
the documented path, so make JUnit correct.

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
sticky disk`), so this was a cold build as expected. **Confirmed working
on the very next run** (same day): the sticky disk parent snapshot
changed from a `__base__...` snapshot to a `commit-...` snapshot (i.e.,
committed from the prior run), and the `FROM golang:1.24` layer restored
in `0.0s` instead of the ~5-8s full extraction seen cold. Blacksmith
docker layer caching on `se-demo-app` is confirmed real and working as of
2026-07-29.

