# Blacksmith SE Demo Base

Internal demo factory for Blacksmith SE demos. This repo builds and keeps
in sync **two permanent, steady-state demo repos** — one shown as-is
("the pain"), one shown as the payoff — instead of resetting a repo live
during a call.

**Start here:** [`DEMO_RUNBOOK.md`](DEMO_RUNBOOK.md) — talk track and
demo arc.

## How it works

```
this repo                              permanent demo repos (never reset)
─────────                              ────────────────────────────────────
app/                    ── rebuild ──▶  se-demo-app              (migrated)
  (shared app source,                     app/ + workflows/current/
   no workflows)                       se-demo-app-unmigrated    (before)
workflows/current/      ─────────────▶    app/ + workflows/unmigrated/
workflows/unmigrated/
```

- **`app/`** is the single shared source of truth for the demo app
  (React/TS dashboard, Go API, Postgres fixtures, `docker-compose.yml`).
  It holds no workflow files and is edited once, ever — there's no
  duplicated tree to keep in sync because there's only one copy.
- **`workflows/current/`** — `ci.yml` / `test.yml` / `docker.yml` for
  `se-demo-app`: Blacksmith runners with the canonical feature combo
  (runner migration, dependency-cache acceleration, Blacksmith docker
  layer caching) baked in by hand, and one deliberately undersized job
  (E2E) that feeds a standing Codesmith rightsizing recommendation and an
  intermittent OOM flake. Every other job sits at the same
  `blacksmith-4vcpu-ubuntu-2404` fair baseline that maps to
  `ubuntu-latest` — no artificially oversized jobs.
- **`workflows/unmigrated/`** — the same three workflows for
  `se-demo-app-unmigrated`: `ubuntu-latest`, with a reasonable, already-decent
  GHA setup (`actions/cache` for npm/Go/Playwright, `type=gha` Docker
  caching) — an honest baseline, not a strawman. See "An honest baseline"
  in the design doc for why the remaining gap to `se-demo-app` is still
  real (hardware, cache backend, runner economics, rightsizing).
- **`SE: Rebuild demo repos`** pushes `app/` + the matching `workflows/`
  set to each target repo's `main` as one normal commit (no force-push,
  no history wipe) — so Codesmith and test-analytics history keeps
  accumulating across rebuilds, and the standing migration-wizard PR
  branch on `se-demo-app-unmigrated` is untouched. Run it only when you
  deliberately change `app/` or a workflow file.
- **Root** (this level) is SE-facing only and never reaches customers:
  this README, the runbook, `CLAUDE.md`, and the workflows below.

## The three workflow pairs

Each pair is the same workflow file, once per repo — same jobs, same
workload, different runner/caching. **Keep this table in sync**: if you
change a job's runner size, jobs, or display name in either `workflows/`
set, update this table in the same commit (see also the matching
invariant in `CLAUDE.md`).

| Pair | `se-demo-app` (current) | `se-demo-app-unmigrated` |
| --- | --- | --- |
| **CI** — lint/typecheck/unit tests | 3-way Node matrix + backend build/vet/unit, all on `blacksmith-4vcpu-ubuntu-2404` | Same jobs on `ubuntu-latest` |
| **Integration & E2E Tests** — backend Postgres integration + Playwright E2E | Backend integration on `blacksmith-4vcpu`; E2E on `blacksmith-2vcpu` (deliberately undersized — 3 browsers forced parallel, feeds the OOM-flake + Codesmith rightsizing arc) | Same two jobs, both on `ubuntu-latest` |
| **Docker: Images & Stack Smoke** — image builds (never pushed) + full-stack compose smoke | 3 jobs: backend/frontend images via Blacksmith layer cache, plus `stack-smoke` (plain `docker compose up --build`) | 3 jobs: backend/frontend images via GitHub's `type=gha` cache, plus the same `stack-smoke` |

## Workflows in this repo

| Workflow | What it does |
| --- | --- |
| `SE: Rebuild demo repos` | Pushes current `app/` + `workflows/{current,unmigrated}` to `se-demo-app` / `se-demo-app-unmigrated` (`current` / `unmigrated` / `both`). Creates the repo on first use. Normal commit, not a reset. |
| `SE: Generate CI activity` | "Goose it now" button: triggers `ci`/`test`/`docker` via `workflow_dispatch` on a target repo, on top of its own `schedule:` triggers. No commits. |
| `Validate` | Base-repo CI: fast correctness checks on `app/` and `workflows/` changes, plus a mirror of the demo repos' `stack-smoke` so a broken compose stack fails here rather than at demo time |
| `SE: SSH sandbox` | No demo purpose — spins up a Blacksmith VM and holds it open (`minutes` input) so you can SSH in and poke around. `workflow_dispatch` only. |

## Setup (one-time)

1. Create the org secret `DEMO_PROVISION_TOKEN`: a PAT with `repo` and
   `workflow` scopes on `jacob-buckles-org`.
2. Run **SE: Rebuild demo repos** with `target: both` to create
   `se-demo-app` and `se-demo-app-unmigrated`.
3. Install the Blacksmith GitHub App at the org level (or on both repos).
   The install is org-scoped regardless, so this costs nothing extra over
   covering one repo.
4. Run the Blacksmith migration wizard once against
   `se-demo-app-unmigrated`; leave the generated PR open, draft, unmerged
   — it's the standing onboarding artifact for the demo.
5. Enable **Docker container caching** for the org by filing a support
   issue from the Blacksmith dashboard (there's no workflow-side setting).
   It's what makes the `Initialize Containers` / image-pull comparison in
   the demo arc land. Available in US West and EU West only — the pool's
   runners report `us-west`.

## Keeping both repos looking alive

Each workflow in `workflows/current/` and `workflows/unmigrated/` has its
own `schedule:` trigger, so both repos' Actions tabs show recent runs
between demos without new commits — nothing to drift or conflict with the
standing migration PR. Need more activity right before a call? Run **SE:
Generate CI activity**.

## Changing the demo

Edit `app/` or a file under `workflows/current/` / `workflows/unmigrated/`
on `main` here, then run **SE: Rebuild demo repos**. That's the whole
sync step — there's no per-repo edit to remember. If the edit touches a
job's runner size, jobs, or name, update the workflow-pairs table above
in the same commit.

- `app/backend/internal/events/` is generated — edit
  `app/backend/tools/genevents` and re-run `go run ./tools/genevents`
  instead. `Validate` diffs it.
- `app/data/*.csv` are generated by `node app/data/generate.mjs`.
- The "before" CI (`workflows/unmigrated/`) must stay slow *authentically*
  — the timing knob is `METRICS_WORKLOAD` (CPU-bound test sweeps, no
  sleeps), so faster runners genuinely speed it up. `Validate` pins it to
  `1` for speed.
