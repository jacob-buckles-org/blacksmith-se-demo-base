# SE Demo Runbook

SE-facing. Only `template/` reaches customers — it is the entire demo
repo (the demo app plus its slow workflows). Everything at the root of
this repo is internal.

## What this repo is

A reset factory: `template/` holds the demo-customer app (Node/TS
dashboard + Go API + Postgres, branded in-character as "Usage Analytics"
so prospects browsing the demo repo see a realistic codebase) with
deliberately slow "before" CI. Instead of a new repo per demo, the base
repo maintains a **fixed pool of 3 long-lived repos** and `reset.yml`
restores one to its "before" state before each call. Blacksmith features
live as patch files in `features/` that become branches + draft PRs in
the demo repo — you choose features at demo time by which PRs you open.

Why a pool, not fresh repos: Blacksmith has no way to remove a repo's
runs from the dashboard (even after the GitHub repo is deleted), so
per-demo repos pile up dead entries and filter rows forever. Three reused
repos keep the dashboard bounded.

The pool (topic `blacksmith-se-demo-pool`):
- `se-demo-app` — **primary**, shown in most demos
- `se-demo-app-backup-2`, `se-demo-app-backup-3` — standbys (parallel or
  back-to-back demos)

## Reset a demo repo

**Reset the day before the call** — baseline seeding takes real
wall-clock time on GitHub's queue.

1. One-time setup: org secret `DEMO_PROVISION_TOKEN` (PAT with `repo`,
   `workflow`, `delete_repo` on `jacob-buckles-org`); run reset once per
   pool repo to create them; install the Blacksmith GitHub App on the 3
   repos (stays installed across resets).
2. Actions → **SE: Reset demo repo** → Run workflow. Inputs:
   - `target` — `primary` (usual), `backup-2`, `backup-3`, or `all`
   - `baseline_pushes` (default 4; `0` skips seeding) — each triggers CI +
     Integration + Docker, so 4 pushes ≈ 12 baseline runs
   - `push_spacing_seconds` (default 90) — spreads runs so history looks
     organic
   - `open_draft_prs` (default false) — open the feature draft PRs, or
     leave the branches unopened and generate PRs live
3. The job summary links the repo and any draft PRs.
4. The app is already installed, so **don't** demo app-install — start at
   the migration wizard on the call. That's the show.

End of day / between customers: reset again (use `baseline_pushes: 0` for
a quick clean-slate), or reset `all` to wipe every pool repo.

Full teardown (rare): **SE: Decommission demo pool** (dry-run by default)
deletes every repo with the `blacksmith-se-demo-pool` topic.

## The "before" state (what the customer sees)

All workflows on `ubuntu-latest`, zero caching. Measured on GitHub-hosted
runners at `METRICS_WORKLOAD: 200`:

| Job | Duration | Deliberate inefficiency |
| --- | --- | --- |
| Frontend (Node 20/22/24) | ~4–6 min × 3 legs | matrix fan-out, npm ci every run, CPU-bound test sweep |
| Backend unit tests | ~4 min | module download every run, CPU-bound sweep |
| Backend build | ~1 min | no Go build cache, 140-file generated package |
| Backend integration | ~1.5 min | Postgres service, row-at-a-time retention sweep |
| Dashboard E2E | ~2.5 min | Playwright browsers re-downloaded every run, 3 browsers |
| Docker backend/frontend | ~1.5–2.5 min each | single-stage builds, full base images, no layer cache |

CI wall-clock ≈ 6–7 min; ~20–25 billable job-minutes per push.

**Timing knob:** `METRICS_WORKLOAD` (env at the top of the template's
`ci.yml` and `test.yml`) scales the CPU-bound "session fingerprint
sweep" and the
integration row count. It's real compute, not sleeps — faster runners
genuinely shrink it (which is the point). 120 ≈ 4-min CI wall; 200 ≈
6–7 min; raise it for a slower "before", lower it for tight calls.

## Demo arc (~30 min)

1. **The pain (3 min).** Open the demo repo's Actions tab: slow runs,
   the matrix burning triple minutes, browser downloads on every E2E
   run, docker rebuilding from scratch. Open a Dockerfile if they're
   docker-heavy.
2. **Onboarding (5 min).** The Blacksmith GitHub App is already installed
   on the pool repo, so start at the **migration wizard**: run it live and
   merge its `runs-on` PR. "One line per job." *(Fallback: draft PR
   "Migrate CI to Blacksmith runners".)* (If a prospect specifically wants
   to see the app-install step, that's the one beat the pool model can't
   show fresh — mention it verbally or use a throwaway repo.)
3. **The speedup (5 min).** Push a trivial commit (or re-run a workflow).
   The "before" is the aged GitHub-hosted baseline runs in the Actions tab
   (slow wall-clock); the "after" is the new Blacksmith run — compare
   wall-clock side by side, then open the Blacksmith dashboard for per-job
   timings and cost. **Date-filter the dashboard to today** — a reused
   pool repo carries prior demos' Blacksmith runs in its history (they
   can't be deleted), so filtering keeps the after view clean.
4. **Codesmith rightsizing (3 min).** The migration PR deliberately
   oversizes the frontend matrix (8 vCPU for lint/unit) and undersizes
   E2E (2 vCPU). Apply Codesmith's recommendation live.
5. **Adoption ladder (10 min).** One feature per persona (below), live
   with Codesmith where possible, otherwise its draft PR. Each PR is one
   feature with one measurable dashboard delta.
6. **Wrap (2 min).** Dashboard cumulative view: minutes saved, cost
   delta at their scale.

## Feature PR toolkit

After a reset the demo repo has these six feature branches, in this order,
built from the patches in `features/` (opened as draft PRs #1–#6 when you
run reset with `open_draft_prs: true`). Preferred path: generate live with
the wizard/Codesmith; the pre-built PR is the zero-dead-air fallback.

| PR | Patch | What it shows | Notes |
| --- | --- | --- | --- |
| #1 | `01-migrate-runners` | runs-on swap, 2x CPU at half price | Wizard generates this live — prefer that |
| #2 | `02-dependency-caching` | npm/Go/Playwright caches via standard actions | Blacksmith transparently accelerates `actions/cache` & `setup-*` — no forked actions |
| #3 | `03-docker-caching` | `useblacksmith/build-push-action` layer cache | 2nd run is the payoff — build twice on the call |
| #4 | `04-sticky-disks` | hot-mounted cache disks, ~3s regardless of size | Overlaps dependency-caching — pick ONE per demo |
| #5 | `05-git-checkout-caching` | `useblacksmith/checkout` drop-in | Repo has 22MB fixtures + chunky history on purpose |
| #6 | `06-static-ip` | CI → IP-allowlisted prod replica | Skips cleanly until `REPLICA_DATABASE_URL` secret set |

Merge-order caveats:

- Mark the sticky-disks / docker-caching PRs ready **after** the runner
  migration merges — they need Blacksmith runners for checks to pass
  (PR checks run against the merge ref, so they go green once main has
  the migration).
- dependency-caching and sticky-disks touch the same lines; merging both
  will conflict. They solve the same problem — pick per persona.

## Persona presets (which PRs to open)

- **Docker-heavy startup:** migrate → docker-caching → sticky-disks.
  Build an image twice; 2nd build's layer-cache hit is the money shot.
- **Monorepo enterprise:** migrate → dependency-caching →
  git-checkout-caching.
- **Security-sensitive fintech:** migrate → static-ip (have a tiny
  hosted Postgres allowlisting only the org's static IP; set
  `REPLICA_DATABASE_URL` in the demo repo) → dependency-caching. Mention
  SSH-into-runner debugging.
- **Quick 15-min call:** migrate only, dashboard before/after, cost
  view. Lower `METRICS_WORKLOAD` to ~80 first so runs fit.

## Extras in the toolkit

- Live runner SSH debugging, ARM (`-arm` labels) and macOS runners, test
  analytics, automatic free tier upgrades — all demoable from the
  dashboard without repo changes.
- If a customer browses code: `data/` is "sample telemetry exports";
  the slow KDF tests are "SEC-114 compliance". It all reads
  in-character.

## Maintenance rules (base repo)

- **Features are patch files** (`features/NN-<name>.patch`): one commit
  each, touching only `template/` paths. The base repo has no feature
  branches — main is the only branch. Editing recipe is in README.md;
  `Validate template` CI fails if a template change breaks a patch.
- Patch commit messages are customer-visible: subject = draft PR title,
  body = PR body. Write them in-character.
- Everything inside `template/` is customer-visible: keep it
  in-character, and keep it free of Blacksmith references (Blacksmith
  only appears inside the feature patches). Root-level files never
  reach demo repos.
