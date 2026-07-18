# SE Demo Runbook

SE-facing. Only `template/` reaches customers — it is the entire demo
repo (the demo app plus its slow workflows). Everything at the root of
this repo is internal.

## What this repo is

A provisioning factory: `template/` holds the demo-customer app (Node/TS
dashboard + Go API + Postgres, branded in-character as "Anvil Analytics"
so prospects browsing the demo repo see a realistic codebase) with
deliberately slow "before" CI. `provision.yml` stamps out one fresh demo
repo per demo from that directory. Blacksmith features live as patch
files in `features/` that become branches + draft PRs in each demo repo
— you choose features at demo time by which PRs you open, not at
provision time.

## Provisioning a demo repo

**Provision the day before the call** — baseline seeding takes real
wall-clock time on GitHub's queue. Or keep 2–3 pre-seeded repos warm.

1. One-time setup: org secret `DEMO_PROVISION_TOKEN` — a PAT with
   `repo`, `workflow`, and `delete_repo` scopes on `jacob-buckles-org`.
2. Actions → **SE: Provision demo repo** → Run workflow. Inputs:
   - `customer_name` — becomes `demo-<slug>-<yyyymmdd>`
   - `baseline_pushes` (default 8) — each triggers CI + Integration +
     Docker, so 8 pushes ≈ 24 baseline runs for the before/after chart
   - `push_spacing_seconds` (default 90) — spreads runs so history looks
     organic
3. The job summary links the new repo and its draft PRs.
4. Afterwards: install the Blacksmith GitHub App on the demo repo (or
   confirm org-wide install covers it). Don't run the migration wizard
   until the call — that's the show.

Cleanup: **SE: Teardown old demo repos** (dry-run by default) archives or
deletes repos with the `blacksmith-demo` topic older than N days.

## The "before" state (what the customer sees)

All workflows on `ubuntu-latest`, zero caching. Measured on GitHub-hosted
runners at `ANVIL_WORKLOAD: 200`:

| Job | Duration | Deliberate inefficiency |
| --- | --- | --- |
| Frontend (Node 20/22/24) | ~4–6 min × 3 legs | matrix fan-out, npm ci every run, CPU-bound test sweep |
| Backend unit tests | ~4 min | module download every run, CPU-bound sweep |
| Backend build | ~1 min | no Go build cache, 140-file generated package |
| Backend integration | ~1.5 min | Postgres service, row-at-a-time retention sweep |
| Dashboard E2E | ~2.5 min | Playwright browsers re-downloaded every run, 3 browsers |
| Docker backend/frontend | ~1.5–2.5 min each | single-stage builds, full base images, no layer cache |

CI wall-clock ≈ 6–7 min; ~20–25 billable job-minutes per push.

**Timing knob:** `ANVIL_WORKLOAD` (env at the top of the template's
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
2. **Onboarding (5 min).** Install the Blacksmith GitHub App on the
   repo, run the **migration wizard** live, merge its `runs-on` PR.
   "One line per job." *(Fallback: draft PR "Migrate CI to Blacksmith
   runners".)*
3. **The speedup (5 min).** Push a trivial commit (or re-run a
   workflow). Compare wall-clock side by side, then open the Blacksmith
   dashboard: per-job timings against the seeded GitHub-hosted baseline,
   cost view.
4. **Codesmith rightsizing (3 min).** The migration PR deliberately
   oversizes the frontend matrix (8 vCPU for lint/unit) and undersizes
   E2E (2 vCPU). Apply Codesmith's recommendation live.
5. **Adoption ladder (10 min).** One feature per persona (below), live
   with Codesmith where possible, otherwise its draft PR. Each PR is one
   feature with one measurable dashboard delta.
6. **Wrap (2 min).** Dashboard cumulative view: minutes saved, cost
   delta at their scale.

## Feature PR toolkit

Every demo repo gets these draft PRs, in this order (PR #1–#7), created
from the patches in `features/`. Preferred path: generate live with the
wizard/Codesmith; the PR is the zero-dead-air fallback.

| PR | Patch | What it shows | Notes |
| --- | --- | --- | --- |
| #1 | `01-migrate-runners` | runs-on swap, 2x CPU at half price | Wizard generates this live — prefer that |
| #2 | `02-dependency-caching` | npm/Go/Playwright caches via standard actions | Blacksmith transparently accelerates `actions/cache` & `setup-*` — no forked actions |
| #3 | `03-docker-caching` | `useblacksmith/build-push-action` layer cache | 2nd run is the payoff — build twice on the call |
| #4 | `04-sticky-disks` | hot-mounted cache disks, ~3s regardless of size | Overlaps dependency-caching — pick ONE per demo |
| #5 | `05-git-checkout-caching` | `useblacksmith/checkout` drop-in | Repo has 22MB fixtures + chunky history on purpose |
| #6 | `06-static-ip` | CI → IP-allowlisted prod replica | Skips cleanly until `REPLICA_DATABASE_URL` secret set |
| #7 | `07-bazel-caching` | zero-config Bazel remote cache | Bazel workflow is paths-gated; trigger via workflow_dispatch |

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
  git-checkout-caching → bazel-caching (dispatch the Bazel workflow).
- **Security-sensitive fintech:** migrate → static-ip (have a tiny
  hosted Postgres allowlisting only the org's static IP; set
  `REPLICA_DATABASE_URL` in the demo repo) → dependency-caching. Mention
  SSH-into-runner debugging.
- **Quick 15-min call:** migrate only, dashboard before/after, cost
  view. Lower `ANVIL_WORKLOAD` to ~80 first so runs fit.

## Extras in the toolkit

- Live runner SSH debugging, ARM (`-arm` labels) and macOS runners, test
  analytics, automatic free tier upgrades — all demoable from the
  dashboard without repo changes.
- Bazel edge case: `bazel.yml` only triggers on `bazel/**` changes or
  manual dispatch, so it stays out of the Actions tab unless you want it.
- If a customer browses code: `bazel/` is a small "export integrity
  tooling" subproject; `data/` is "sample telemetry exports"; the slow
  KDF tests are "SEC-114 compliance". It all reads in-character.

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
