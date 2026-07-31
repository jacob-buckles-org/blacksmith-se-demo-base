# SE Demo Runbook

SE-facing. Only `se-demo-app` and `se-demo-app-unmigrated` reach
customers — everything at the root of this repo (`README.md`, this file,
`CLAUDE.md`, the workflows) is internal.

For *why* the flaky test flakes, how it's built, and how to answer "would
this really happen?" — see [`DEMO_MECHANICS.md`](DEMO_MECHANICS.md).

## What this repo is

A build factory for two **permanent, steady-state** demo repos — nothing
resets, nothing is reused across an artificial pool. Each repo just is
what it is, every time you open it:

- **`se-demo-app-unmigrated`** — a reasonably well-run engineering team's
  CI: `ubuntu-latest`, `actions/cache` for npm/Go/Playwright, GitHub's own
  `type=gha` Docker cache backend. Not a strawman — a real, decent GHA
  setup. Has one standing artifact: an **open, draft, unmerged**
  migration-wizard PR.
- **`se-demo-app`** — the same app, migrated: Blacksmith runners +
  transparent dependency-cache acceleration + Blacksmith docker layer
  caching (`useblacksmith/build-push-action`). One deliberately undersized
  job (E2E, 2 vCPU) feeds a standing Codesmith recommendation and an
  intermittent OOM-caused flaky test. Every other job sits at the same
  fair `blacksmith-4vcpu-ubuntu-2404` baseline that maps to
  `ubuntu-latest` — nothing else is artificially mis-sized.

Because nothing resets, there is no "reset the day before the call" step
and no risk of a live migration-wizard run coming back slower on a bad
network day. Both repos' `schedule:` triggers keep their Actions tabs
showing recent runs on their own; **SE: Generate CI activity** is a
manual top-up if you want fresher runs right before a big call.

## Why an honest baseline, not a strawman

`se-demo-app-unmigrated` already has caching — it just has *GitHub's*
caching. That's deliberate: a technical prospect will poke at "you just
turned caching off," and this setup doesn't give them that opening. The
real, defensible gap to `se-demo-app`:

- **Hardware** — shared, generic GitHub-hosted runners vs. Blacksmith's
  faster dedicated hardware. A real CPU/throughput difference.
- **Cache backend** — `actions/cache` / `type=gha` round-trip over the
  network and are size/eviction limited; Blacksmith's cache is a
  persistent, locally-mounted disk with near-instant restores and no
  shared eviction budget. "You already have caching — here's why ours
  actually hits and restores fast."
- **Runner economics** — per-minute GitHub pricing and standard
  concurrency limits vs. Blacksmith's pricing and higher concurrency
  ceiling.
- **Rightsizing** — GitHub-hosted runners don't offer vCPU choice, so
  Codesmith rightsizing has nothing to apply to on the unmigrated side;
  it's purely a `se-demo-app`-side story.

## Demo arc (~25 min)

1. **The pain (3 min).** Open `se-demo-app-unmigrated`'s Actions tab:
   real wall-clock times on `ubuntu-latest`, even with its own decent
   caching in place. Point out the caching that *is* there so the later
   comparison reads as credible, not rigged.
2. **Onboarding (3 min).** Open the standing, unmerged migration-wizard
   PR on `se-demo-app-unmigrated`. Narrate it rather than running it live
   — "one line per job, generated automatically." *(The GitHub App
   install itself is intentionally skipped — it's already installed at
   the org level; mention it verbally if asked.)*
3. **The speedup (5 min).** Switch to `se-demo-app`'s Actions tab +
   Blacksmith dashboard: same app, same workload, migrated. Compare
   wall-clock and cost against the unmigrated repo directly — two clean
   Actions tabs, nothing to filter or date-range.
4. **Container caching (3 min).** Compare the `stack-smoke` job across both
   repos — it pulls `postgres:16` plus the much larger `golang`/`node`
   build bases, so it's the better of the two showcases. Point at the
   `Pulling fs layer` / `Extracting` lines present on the unmigrated side
   and absent on the migrated one: "your test dependencies stop being a
   download." The `postgres:16` service container in
   `Integration & E2E Tests` is supporting detail, not the headline — its
   unmigrated baseline is only ~10s, so the best possible saving there is
   ~7–8s. *(Needs the org feature enabled — see Setup in README, and the
   measured baselines in `VERIFY.md` for how to confirm it's live.)*
5. **Test analytics + flaky-test arc (8 min).** On `se-demo-app`:
   1. **Test Analytics** — the Tests tab is populated with no config on
      your side: JUnit XML from Playwright and Vitest is auto-detected,
      and Go test output is parsed straight from the logs. Show test
      names, durations, and pass/fail history across runs.
   2. **Flaky test** — `request volume chart › plots request and error
      series for the last 24h` (`e2e/chart.spec.ts`) shows a flakiness %
      built from cross-run history: it fails in roughly 1 run in 8 and
      passes the rest. Open the code: it's an ordinary under-specified wait
      (`CHART_RENDER_BUDGET_MS`) that loses the race under CPU contention
      on the deliberately undersized 2 vCPU E2E runner. Nothing scripted —
      exactly the flake every team already has. The point to land: the
      failure is *intermittent and unrelated to the change under test*,
      which is why it survives in a real codebase — someone re-runs CI, it
      goes green, nobody investigates.
   3. **Runner utilization metrics** confirm CPU pegged on that job.
   4. **Global log search** — search `SEC-114` (or `exceeded budget`).
      Every backend test run logs
      `WARN telemetry: fingerprint sweep exceeded budget (SEC-114): N
      sessions took Xms, budget 250ms`. Sort to the earliest occurrence:
      it appears from a specific date forward, which is when the sweep
      instrumentation shipped. "Find the first time this ever appeared in
      CI, across every repo, in one query." Deliberately *not* the same
      root cause as the flaky test, so this beat works even on a day the
      flake doesn't fire.
   5. Mention (or do) SSH-into-the-retained-VM debugging on a failed run.
   6. Apply Codesmith's standing rightsizing recommendation for the E2E
      job live — the data behind it isn't freshly collected in front of
      the prospect, so this is low-risk. More CPU → less contention → the
      chart test stops losing its race.
   7. Mention the single updated PR comment (test failure vs. infra
      flake) and Slack-connected monitors in passing.
6. **Wrap (5 min).** Dashboard cumulative view: minutes saved, cost delta
   at their scale.

## What stays genuinely live

- **SSH-into-runner debugging** — inherently in-the-moment; quick and
  reliable, no steady state needed. One caveat learned the hard way: a
  GitHub SSH key added the *same day* may not be honoured yet (there's an
  undocumented propagation delay — observed still failing 28 min after
  adding a key, working the next day). Use a key that's been on your
  account a while, and don't debug a fresh key mid-call.
- **Applying the Codesmith recommendation** — the recommendation
  pre-exists (fed by scheduled history), but clicking apply live is a
  good, low-risk flourish.
- **"Goose it now"** — run **SE: Generate CI activity** before a big call
  if you want fresher runs than the schedule alone has produced.

## Maintenance rules (base repo)

- `app/` is the single copy of the demo app. Edit it here, then run **SE:
  Rebuild demo repos** — that's the entire sync step, for both target
  repos.
- Workflow files are hand-maintained directly in `workflows/current/` and
  `workflows/unmigrated/`. There's no patch/PR pipeline to route through.
- Everything that ends up in either demo repo is customer-visible — keep
  `app/` and both `workflows/` sets free of anything you wouldn't want a
  prospect reading. No in-character cover story is needed; keep things
  clean and plainly named.
- `Validate` CI runs fast checks against `app/` and both `workflows/`
  sets on every push/PR here, so a break surfaces before the next rebuild
  rather than at demo time.
