# Demo mechanics: what's deliberately engineered

Internal. Records the things in the demo repos that were **put there on
purpose** to make a Blacksmith feature demonstrable — what they are, the
exact levers that make them work, and the real-world story behind each so
you can defend them when a prospect pushes back.

Read this before changing any of it: several of these are load-bearing in
non-obvious ways, and the failure mode is silent (the demo stops working
without anything going red). Hard constraints are also mirrored in
`CLAUDE.md`.

Everything here lives in `app/` or `workflows/current/` and is therefore
**customer-visible**. That's deliberate — none of it should look
embarrassing if a prospect opens the file.

---

## 1. The flaky test

**Test:** `app/frontend/e2e/chart.spec.ts` →
`request volume chart › plots request and error series for the last 24h`

**What it asserts:** that the dashboard's Recharts volume chart becomes
visible, and renders both series (requests + errors), after the metrics
fetch resolves.

**What makes it flake:** a fixed render budget that's too tight for a
contended CI runner.

```ts
const CHART_RENDER_BUDGET_MS = 400
…
await expect(chart).toBeVisible({ timeout: CHART_RENDER_BUDGET_MS })
```

The chart is gated behind `loaded &&` in `App.tsx`, so it only mounts after
React mounts → `fetchMetrics()` resolves → `setLoaded(true)` → the
`useMemo` rollups (`bucketByHour`, `summarizeByService`, `computeOverview`)
run → Recharts lays out and paints. Under CPU contention that chain
sometimes takes longer than 400ms.

### The five levers that produce it

Change any one of these and the flake rate moves. All five are required.

| # | Lever | Where |
| --- | --- | --- |
| 1 | `CHART_RENDER_BUDGET_MS = 400` — the tuning knob | `app/frontend/e2e/chart.spec.ts` |
| 2 | `workers: 3` in CI — forces all 3 browsers to run concurrently | `app/frontend/playwright.config.ts` |
| 3 | 3 browser projects (chromium/firefox/webkit) | `app/frontend/playwright.config.ts` |
| 4 | `runs-on: blacksmith-2vcpu-ubuntu-2404` — 2 vCPU for 3 browsers | `workflows/current/test.yml` |
| 5 | `test.describe.configure({ retries: 0 })` — makes the failure *visible* | `app/frontend/e2e/chart.spec.ts` |

Lever 5 is the one people will want to "fix". Don't — see
[§1.4](#14-why-retries-are-off).

**Nothing random is involved.** No `Math.random()`, no `sleep`, no
injected failure. If a prospect reads the test, there is nothing to be
embarrassed about: it's a real assertion with a badly chosen wait.

### 1.1 Measured behaviour

Time from `page.goto('/')` returning to the chart being visible:

| Environment | chromium | firefox | webkit | overall |
| --- | --- | --- | --- | --- |
| Local (M-series Mac, 3 projects concurrent) | 42–76ms | 42–54ms | 29–57ms | **29–76ms, median ~46ms** |
| CI (`blacksmith-2vcpu`, 3 projects concurrent) | 159–191ms | 216–582ms | 242–488ms | **159–582ms, median ~242ms** |

So CI is **~5x slower at the median and ~13x slower at the worst case** than
a developer's laptop. The 400ms budget is ~9x the local median — which is
why it feels absurdly safe when you write it, and still fails in CI.

The CI distribution is **bimodal**: chromium is always fast, while
firefox/webkit split into a fast group (216–312ms) and a contended group
(467–582ms) depending on how hard the three workers collide that run.
webkit is the usual victim.

**Rate:** ~9% of runs (1/8 with retries on, 1/15 with retries off, 23 runs
total) → roughly **1 red run in 10**. It will *not* fail on demand.

### 1.2 The realistic story (this is the bit that sells it)

This is the single most common flaky-test pattern in the industry, and the
mechanism here is exactly how it happens in real codebases:

1. A developer adds a test for the chart. On their laptop the chart appears
   in ~46ms. They add an explicit `timeout: 400` — nearly **9x** the time
   they just observed. That feels like a generous safety margin, and it's a
   completely reasonable thing to write.
2. The margin was measured against the wrong baseline. CI runs on slower
   shared hardware, with three browser projects competing for 2 vCPU. The
   same operation takes 159–582ms there.
3. So the 9x margin silently becomes ~0.7x in the worst case, and the test
   starts failing roughly 1 run in 10.
4. Someone clicks **re-run failed jobs**. It goes green. The PR merges.
   No bug is filed, because from the developer's point of view nothing was
   wrong — and they're not entirely mistaken, the product code is fine.
5. Repeat across a year and a team. Now CI is "flaky" as a general vibe,
   people re-run red builds reflexively instead of reading them, and real
   regressions get lost in the noise.

Details that make it hold up under scrutiny:

- **The test's intent is correct.** It's checking something real and worth
  checking. The defect is in the *waiting strategy*, not the assertion.
- **Explicit `timeout:` on an assertion is idiomatic**, not exotic. Teams
  add them to fail fast, or because some global default was too slow.
- **It's browser-specific.** "It only fails on Safari, only in CI" is a
  sentence every frontend team has said out loud.
- **It's load-dependent, not deterministic**, which is precisely why it
  survives: it passes on the retry, so it never looks like a real bug.

### 1.3 If a prospect challenges you

**"Isn't the real fix just to write the test properly?"** Yes — and say so.
The correct engineering fix is to drop the arbitrary timeout and let
Playwright's auto-retrying assertion use the default, or wait on a
deterministic signal. Blacksmith's claim is *not* "we fix your bad tests."
It's three narrower, defensible things:

1. **Surfacing.** You cannot fix what you can't see. This test failed
   intermittently for weeks in a green-looking repo; test analytics is what
   makes it a known item instead of ambient noise.
2. **Reducing the trigger.** A large class of these flakes is contention,
   not logic. Right-sizing the runner removes the contention that trips
   them, so the same suite gets less flaky without touching test code.
3. **Diagnosis time.** Utilization metrics (CPU pegged), global log search,
   and SSH into the retained VM turn "why did this fail?" from an
   afternoon into a few minutes.

**"So Blacksmith just hides the symptom?"** Be straight: right-sizing makes
*this* test stop failing without fixing the underlying fragility. That's a
genuine trade-off, and the honest pitch is that it buys the team time and
visibility — the flake goes from "fails randomly, nobody knows why" to
"known, attributed to contention, and quantified."

### 1.4 Why retries are off

Originally this spec kept Playwright's `retries: 2`, so a failure would be
recorded as *flaky* while the job stayed green — seemingly the best of both
worlds. **That didn't work, and it's worth understanding why before anyone
re-enables it.**

Playwright's JUnit reporter writes only a test's **final** status. Verified
with a probe that deliberately fails attempt 0 and passes on retry:

```xml
<testsuites tests="1" failures="0" ...>
  <testcase name="probe: fails first attempt, passes on retry" time="0">
    <system-out>[[ATTACHMENT|...retry1/trace.zip]]</system-out>
  </testcase>
```

`failures="0"`, no `<failure>` element — the only residue is an attachment
filename. Since Blacksmith's Test Analytics ingests JUnit XML, it saw a
clean pass and **the flaky-test demo had nothing to show at all.**

With `retries: 0`, the same failure produces `failures="1"` and a full
`<failure message="expect(locator).toBeVisible() failed"
type="expect.toBeVisible">` element carrying the error, locator, call log
and code frame. Flakiness % then comes from cross-run history (fails some
runs, passes others), which is how it's normally computed anyway.

The cost — ~1 red run in 10 — is accepted deliberately. **Re-enabling
retries to "clean up" the red runs would silently break this demo.**

---

## 2. The searchable log line (global log search)

**Where:** `app/backend/internal/telemetry/fingerprint.go` →
`FingerprintSweep`

```go
log.Printf("WARN telemetry: fingerprint sweep exceeded budget (SEC-114): %d sessions took %s, budget %s", …)
```

**Mechanism:** the sweep logs whenever it exceeds a 250ms budget. At
`METRICS_WORKLOAD: 200` the sweep processes 24,000 sessions and takes
~2m17s, so it trips **every run** — the string is always findable.

**Two things that are load-bearing:**

- It lives in **production code, not a test**, so it reads honestly if a
  prospect follows it.
- The demo workflows run `go test -v`. Without `-v`, `go test` discards
  passing packages' output and **this line never reaches the logs.** That's
  also why Go test names appear for analytics. Don't remove `-v`.

**Realistic story:** a threshold warning nobody triaged. Real teams have
dozens of these — a warning gets added with a plausible budget, the budget
is exceeded routinely under real load, and because it's a `WARN` and not a
failure it scrolls past forever. The demo beat is "when did this *start*?",
which is a genuinely hard question without searchable logs and a trivial
one with them. The measured duration in the message varies run to run,
which is a nice tell that it's real telemetry rather than a fixed string.

**Deliberately decoupled from the flaky test.** An earlier design had one
root cause feed both the flake and the log line. That's more elegant but
fragile: on a day the flake didn't fire, *both* beats would break. Keep
them independent.

---

## 3. The undersized E2E runner

**Where:** `workflows/current/test.yml` → `e2e` job,
`runs-on: blacksmith-2vcpu-ubuntu-2404`

Every other job in `workflows/current/` sits at
`blacksmith-4vcpu-ubuntu-2404`, which is the fair baseline matching
`ubuntu-latest`'s 4 vCPU / 16 GB on the unmigrated side. The E2E job is the
**one deliberate exception**: 2 vCPU for three concurrent browsers.

It does double duty — it creates the contention behind the flaky test, and
it gives Codesmith something real to recommend right-sizing.

**Realistic story:** nobody sized this job on purpose. Someone set it once
when the suite had one browser, projects got added later, and the runner
size was never revisited. That's how essentially all runner mis-sizing
happens in the wild — it's drift, not a decision.

An earlier version of the demo also oversized the frontend matrix to
manufacture a "scale down" recommendation. That was removed on 2026-07-29
because it undercut the honest-baseline premise: it's rigging the setup
rather than showing a real one. **Don't reintroduce artificially oversized
jobs** — any Codesmith recommendation should be organic.

---

## 4. Scenario PRs (`se-demo-app-prs`)

**Where:** `scenarios/NN-<name>.patch`, opened by
`SE: Simulate PR activity` against `se-demo-app-prs`.

Each scenario is a single-commit `git format-patch` touching `app/` paths.
The commit **subject becomes the PR title and the body becomes the PR
body**, so both are customer-visible — write them as a real engineer would.
Applied with `git am -p2` to strip the leading `a/app/`, because `app/`
contents sit at the repo root in the target.

`Validate` re-applies every scenario patch to current `app/` on each push,
so an `app/` change that breaks one fails there rather than during demo prep.

### Regenerating a scenario patch

```sh
git checkout -b tmp-scenario main
git am scenarios/01-single-pass-rollup.patch   # patch -> commit
# ...edit, then: git commit -a --amend
git format-patch -1 --stdout > scenarios/01-single-pass-rollup.patch
git checkout main && git branch -D tmp-scenario
```

### 4.1 `01-single-pass-rollup`

**PR title:** *perf(aggregate): roll up service metrics in a single pass*

**What it claims to do:** replace the three `reduce()` walks in
`summarizeByService` with one loop, framed as a profiling win now that large
tenants send hundreds of thousands of points per window.

**The planted bug:** the rewrite changes the `errorRate` denominator from
`totalRequests` to `bucket.length` — i.e. errors per *data point* instead of
errors per *request*.

```ts
errorRate: bucket.length === 0 ? 0 : totalErrors / bucket.length,
//                                                 ^^^^^^^^^^^^^ should be totalRequests
```

**What CI does:** exactly one test fails —
`aggregate.test.ts › summarizeByService › rolls points up per service,
sorted by volume` — with `expected 1 to be close to 0.005` (2/2 instead of
2/400). Lint and typecheck **pass**, so the unit test is the sole signal.
That single unambiguous failure is deliberate: it gives Codesmith a clean
target and gives you a clean thing to point at.

**Why it's realistic:** a performance refactor that silently changes
semantics is one of the most common ways real bugs ship. The diff *looks*
like a mechanical optimisation, the claim "no behaviour change intended" is
sincere, and a reviewer skimming it would plausibly approve. The bug is also
the classic rate-vs-average denominator confusion, which is easy to make and
easy to miss because the resulting number is still a plausible-looking
small float.

**Autofix mechanics worth knowing:** Codesmith enables autofix by default
only on PRs **it** created. A PR opened by this workflow therefore sits
broken until you either click **"Autofix with [code]smith"** on it or enable
autofix for all PRs in settings. That's the desired shape: a standing broken
PR you can open cold, plus a live moment to trigger the fix. The docs also
warn that interdependent autofix PRs can cascade into long sessions, so
avoid stacking many open scenarios at once — `cleanup` exists for that.

**Verified 2026-08-04.** Autofix produced the *right* fix on PR #1:
`fix(aggregate): restore totalRequests as errorRate denominator`. It kept
the performance refactor and corrected only the denominator, rather than
reverting the PR wholesale — i.e. it understood the intent of the change,
which is a much better story than "it undid it."

### 4.2 The long-term cycle (important)

**You never need to revert or undo a merged scenario PR.** Merging is the
point — it's what gives this repo real history. But merging does mean main
now contains the fix, so the same patch won't apply again:

```
main after merge:  single-pass loop + CORRECT denominator   ← Codesmith's fix
the patch expects: three reduce() calls                      ← canonical app/
→ git am fails
```

**The reset is `SE: Rebuild demo repos` with `target: churn`.** It pushes the
canonical `app/` back over main as an ordinary commit, so the patch applies
again. Verified end to end: patch failed against the post-merge main, then
applied cleanly after a rebuild.

Crucially the rebuild **only replaces content on main — it does not rewrite
history**. After the reset, the log still reads:

```
Sync from blacksmith-se-demo-base (churn)     ← the reset
Merge pull request #1 from …scenario/01-…      ← your merge, preserved
fix(aggregate): restore totalRequests …        ← Codesmith's fix, preserved
perf(aggregate): roll up service metrics …     ← the scenario PR
Sync from blacksmith-se-demo-base (churn)
```

So every demo run leaves a permanent, real-looking merged-PR trail while the
scenario stays replayable. The repo gets *more* realistic over time, which is
the opposite of the other two.

**Per-demo loop:**

1. `SE: Simulate PR activity` → `open` (a fresh timestamped branch, so it
   never collides with previous runs).
2. Demo: CI goes red → click **Autofix with [code]smith** → merge if you want
   the history.
3. Only if you merged: `SE: Rebuild demo repos` → `target: churn` to reset
   main. Skip it if you closed the PR instead.

`cleanup` closes any open scenario PRs and deletes their branches — worth
running if several are stacked, since the docs warn interdependent autofix
PRs can cascade into long sessions.

### 4.3 Why this isn't just done in `se-demo-app`

Reasonable question, and worth writing down because it'll come up again. Four
reasons, roughly in order of how much they'd actually hurt:

1. **`se-demo-app` and `se-demo-app-unmigrated` must stay content-identical.**
   The entire before/after comparison rests on "same app, same workload, only
   the CI config differs." Merging PRs into one and not the other makes them
   diverge, and the comparison stops being honest. This is the reason that
   really settles it.
2. **The deliberately flaky E2E test would poison the loop.** It reddens ~9%
   of runs for reasons unrelated to the change under review — and with
   autofix enabled, Codesmith could go off and "fix" the flaky test instead
   of the regression. Actively harmful to the demo you're trying to run.
3. **`se-demo-app`'s CI is deliberately slow** (`METRICS_WORKLOAD: 200`,
   3-browser E2E matrix, ~6-7 min). A PR/autofix loop wants fast feedback;
   the churn repo does the whole PR run in **52s**.
4. **Churn breaks the steady-state guarantee.** The value of those two repos
   is that they're always exactly what you expect when you open them cold.
   A stream of merged PRs and periodic main resets is the opposite.

**The narrow version that *would* work in `se-demo-app`:** open a scenario
PR and always **close** it rather than merging. Main never changes, so
divergence and resets both go away. You'd still inherit the slow CI and the
flaky test (1 and 4 solved, 2 and 3 not). Worth knowing as a fallback if you
ever need to demo autofix against the migrated repo specifically — but note
that *merging* is what makes the history realistic, and merging is precisely
what you can't safely do there.

## 5. What is *not* engineered

Worth being clear on, because the credibility of the above depends on it:

- **`se-demo-app-unmigrated` is not a strawman.** It has real
  `actions/cache` for npm/Playwright, setup-go's built-in module cache, and
  GitHub's own `type=gha` Docker layer cache. The gap to `se-demo-app` is
  hardware, cache backend, economics and right-sizing — not "we turned
  caching off."
- **The ~2x compose-job delta is hardware, not container caching.**
  Measured at ~38s vs ~72s while container caching was still *off*. Both
  runners are 4 vCPU / 16 GB, so that gap is CPU generation, disk I/O and
  network. Attribute it correctly if asked — and note it's clean evidence
  for the "faster hardware, not just more cores" claim.
- **Codesmith recommendations are unseeded** (see §3).
- **`METRICS_WORKLOAD` is real CPU work**, not sleeps — a CPU-bound
  fingerprint sweep. Faster runners genuinely shrink it, which is the whole
  point.
