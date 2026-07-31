# Verification walkthrough

Hands-on checklist for confirming the demo material actually works before
you rely on it in front of a prospect. Talk track lives in
`[DEMO_RUNBOOK.md](DEMO_RUNBOOK.md)`; this file is just "prove it works."

The two demo repos (both public, so Actions minutes are free):

- `se-demo-app` — migrated → [https://github.com/jacob-buckles-org/se-demo-app](https://github.com/jacob-buckles-org/se-demo-app)
- `se-demo-app-unmigrated` — before state → [https://github.com/jacob-buckles-org/se-demo-app-unmigrated](https://github.com/jacob-buckles-org/se-demo-app-unmigrated)
- Blacksmith dashboard → [https://app.blacksmith.sh/jacob-buckles-org](https://app.blacksmith.sh/jacob-buckles-org)

Both repos run the same three workflows, named identically on purpose so
you can compare them side by side: `CI`, `Integration & E2E Tests`,
`Docker: Images & Stack Smoke`.

## What's already proven vs. what needs your eyes

I verified everything reachable from logs and the GitHub API. I could not
see the Blacksmith dashboard, so every dashboard-rendering claim below is
unverified and is what you're actually checking.


| Already verified (live)                              | Needs your eyes (dashboard UI)                         |
| ---------------------------------------------------- | ------------------------------------------------------ |
| JUnit XML is written by Playwright + Vitest          | Tests tab is actually populated from it                |
| `SEC-114` warning appears in CI logs every run       | Global log search finds it + sorts to first occurrence |
| Go test names/durations appear in logs (`-v`)        | Go tests show up in Test Analytics                     |
| Chart test flakes ~12% of runs, 0 red (8-run sample) | A flakiness % is displayed for it                      |
| `stack-smoke` green in both repos + base `Validate`  | Container-cache effect on Initialize Containers        |
| Docker sticky-disk layer cache reused across runs    | Runner utilization (CPU) metrics on the E2E job        |




## 0. Blockers to clear first

- [x] **Enable Docker container caching for the org.** File a support issue
  ```
  from the Blacksmith dashboard — there is no workflow-side setting, so
  until this is on, section 4 shows no delta. US West / EU West only;
  the runners report `us-west`, so you're eligible.
  ```
- [ ] **Reinstate the migration-wizard PR.** `se-demo-app-unmigrated` PR #1
  ```
  was closed (by you, 2026-07-29 — not by the rebuild), so the
  onboarding beat currently has nothing to open. Two options:
  - **Recommended: re-run the wizard** against `se-demo-app-unmigrated`
    to get a fresh PR. The old one predates the `stack-smoke` job, so
    reopening it would migrate the two image-build jobs and silently
    leave `stack-smoke` on `ubuntu-latest` — a sharp prospect may ask
    why one job wasn't migrated.
  - Or just reopen [PR #1](https://github.com/jacob-buckles-org/se-demo-app-unmigrated/pull/1).
    It still merges cleanly and the diff is otherwise correct
    (`runs-on` swaps + `useblacksmith/setup-docker-builder` +
    `useblacksmith/build-push-action`, dropping `type=gha`).
  - Either way leave it **open, draft, unmerged**.
  ```
- [x] **SSH into a runner — resolved 2026-07-31.** It started working with
  no config change on either side, so the earlier
  `Permission denied (publickey)` was almost certainly a propagation delay
  between adding the key to GitHub and Blacksmith honouring it (key
  registered 07-30 18:51 UTC, still failing at T+28 min, fine by the next
  day). Caveat for demos: **don't rely on a live SSH beat using a key added
  that same day.** `SE: SSH sandbox` in the base repo spins up a VM to
  retest whenever you want. Full writeup in the vault:
  `3 - Product/Features/SSH Access.md`.



## 1. Test Analytics

Nothing to configure on Blacksmith's side — it auto-detects JUnit XML
written anywhere during a job, and parses Go test output from logs.

- [ ] Open a recent `Integration & E2E Tests` run on `se-demo-app` in
  ```
  the Blacksmith dashboard → **Tests** tab.
  ```
- [ ] Confirm you see individual Playwright test names (not just job
  ```
  names), with per-test durations, e.g. `request volume chart › plots
  request and error series for the last 24h`.
  ```
- [ ] Open a recent `CI` run → confirm Vitest unit tests appear
  ```
  (21 of them, from `frontend/test-results/junit-unit.xml`) **and** Go
  tests appear (`TestHealthz`, `TestFingerprintSweepNoCollisions`, …).
  ```
- [ ] Sanity check the history depth: `se-demo-app` has ~91 runs, ~39 of
  ```
  them `Integration & E2E Tests`. That should be enough history for
  trends to render.
  ```

**If Go tests are missing:** they only appear because the demo workflows run
`go test -v`. Without `-v`, `go test` discards passing packages' output
entirely. Don't "tidy" that flag away.

## 2. Flaky test finding

The flaky test is `frontend/e2e/chart.spec.ts` →
`request volume chart › plots request and error series for the last 24h`.
It asserts the chart renders within `CHART_RENDER_BUDGET_MS` (400ms) and
loses that race under CPU contention on the deliberately undersized
2 vCPU E2E runner. **webkit** is the usual victim.

- [ ] In the dashboard, find that test and confirm a **non-zero flakiness %**
  ```
  with run history.
  ```
- [ ] Open the code in the repo and confirm it reads like an ordinary
  ```
  under-specified wait — no `Math.random()`, nothing scripted. This is
  the point: it's the same flake every team already has.
  ```
- [ ] Confirm the job is **green** despite the flake. Playwright retries
  ```
  twice with warm caches, so a failure normally becomes
  *flaky-but-green* — which is exactly why nobody notices it without
  analytics.
  ```

**Expectations, measured over 8 live runs:** 1/8 runs (12%) recorded the
flake, 0/8 went red. So **it will not flake on demand** — do not plan to
trigger it live. Point at accumulated history instead.

- [ ] Want more samples before a demo? Run `SE: Generate CI activity`
  ```
  in the base repo (`target: current`) a few times.
  ```

If flakiness reads 0%, the history window may predate the test (added
2026-07-31). Trigger several runs and re-check. Before retuning the 400ms
budget, read the bimodal-distribution note in `notepad.md` — tightening it
to anything between 320–460ms provably changes nothing.

## 3. Global log search

Every backend test run logs a line like:

```
WARN telemetry: fingerprint sweep exceeded budget (SEC-114): 24000 sessions took 2m16.76s, budget 250ms
```

It comes from production code (`telemetry.FingerprintSweep`), not a test,
so it reads honestly if a prospect goes looking.

- [ ] In the dashboard, search `SEC-114` (or `exceeded budget`).
- [ ] Confirm hits across many runs and **both** repos.
- [ ] Sort/scan to the **earliest** occurrence — it should start
  ```
  **2026-07-31**, when the instrumentation shipped. That's the demo
  beat: "find the first time this ever appeared in CI, in one query."
  All run history before that date is a clean "before."
  ```
- [ ] Confirm the duration in the message varies run to run (it's measured,
  ```
  not hardcoded) — good for showing this is real telemetry.
  ```

This is deliberately **not** the same root cause as the flaky test, so this
beat still works on a day the flake doesn't fire. Keep them independent.

## 4. Docker container caching

Support ticket filed 2026-07-31. **Measured that same day, it was not yet
active** — see baselines below, which are also how you tell when it lands.

### Baselines measured before caching took effect (2026-07-31)

`Initialize containers` — the `postgres:16` service in
`Integration & E2E Tests`, sampled over 10 runs each:

| Repo | Initialize containers |
| --- | --- |
| `se-demo-app` (Blacksmith) | 6–14s, typically **~10s** |
| `se-demo-app-unmigrated` (`ubuntu-latest`) | 8–12s, typically **~10s** |

`Bring up the stack` — the compose job, which also pulls the much larger
`golang:1.24` and `node:22` build bases:

| Repo | Bring up the stack |
| --- | --- |
| `se-demo-app` (Blacksmith) | 34–44s (**~38s**) |
| `se-demo-app-unmigrated` (`ubuntu-latest`) | 66–80s (**~72s**) |

### Two honest expectations to set

1. **The postgres-only comparison will never be dramatic.** The unmigrated
   baseline is already only ~10s, so even a perfect cache saves ~7–8s. The
   docs' "minutes to seconds" claim assumes far heavier service-container
   setups than one Postgres. Don't build the beat around this number — use
   it as supporting detail.
2. **The ~38s vs ~72s compose delta is real but is *not* container caching**
   — caching wasn't active when it was measured. Both runners are 4 vCPU /
   16 GB (public-repo `ubuntu-latest` matches
   `blacksmith-4vcpu-ubuntu-2404` on paper), so that ~2x is CPU
   generation, disk I/O and network throughput — i.e. the hardware story
   from the honest-baseline section. Attribute it correctly if asked.

### Checks

- [ ] Re-run the numbers above and confirm `se-demo-app`'s figures have
      dropped relative to these baselines. That's the only reliable signal
      that the ticket has been actioned — there's no dashboard toggle to
      look at.
- [ ] In the unmigrated run's log, confirm the visible `Pulling fs layer` /
      `Extracting` progress lines, then confirm they're absent (or much
      shorter) on the migrated side. This is more legible than the raw
      durations.
- [ ] Once active, the compose job is the better showcase of the two —
      bigger images, bigger delta.

That job deliberately uses **plain** `docker compose`**, no Blacksmith
builder**, so its cost is dominated by image *pulls* (container caching)
rather than build layers (already covered by the per-service image jobs).
Don't "improve" it by adding the Blacksmith builder — it would confuse the
two stories.

## 5. Adjacent things worth a glance

- [ ] **Docker layer caching** (separate from container caching, already
  ```
  working): in `se-demo-app`'s `Build backend image` job, the
  `useblacksmith/setup-docker-builder` step logs `Getting sticky disk
  for jacob-buckles-org/se-demo-app` → `Successfully obtained sticky
  disk`, and a warm run restores the base image layer in `0.0s` vs
  ~5–8s cold.
  ```
- [ ] **Codesmith** recommendations — check whether it now organically
  ```
  flags anything. Every job sits at the fair 4 vCPU baseline except the
  2 vCPU E2E job, so any recommendation is unseeded. Historical recs on
  the old setup are logged in `notepad.md`.
  ```
- [ ] **Runner utilization metrics** on the E2E job — CPU should be pegged,
  ```
  which is the corroborating evidence for the flaky test's cause and
  the rightsizing recommendation.
  ```



## If something looks wrong

- Both repos are rebuilt from this repo — never edit them directly. Change
`app/` or `workflows/{current,unmigrated}/` here, then run
`SE: Rebuild demo repos` (`target: both`).
- The rebuild pushes a normal commit and **no-ops when content is
identical**, so if you see "no new runs," that's expected, not a failure.
- It also leaves other branches alone, so the wizard PR branch survives
rebuilds (confirmed: PR #1 was closed by a human, not by the sync).
- Base-repo `Validate` mirrors the `stack-smoke` job, so a broken compose
stack fails there rather than at demo time.

