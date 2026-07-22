# Blacksmith SE Demo Base

Internal demo factory for Blacksmith SE demos. This repo maintains a
small fixed pool of long-lived demo repos and **resets** one to a
deliberately slow "before" state before each call, ready for a live
Blacksmith migration.

**Start here:** [`DEMO_RUNBOOK.md`](DEMO_RUNBOOK.md) — talk track,
persona presets, timings, and per-feature demo notes.

## How it works

```
this repo                                pool repos (reused, reset per demo)
─────────                                ───────────────────────────────────
template/            ── reset.yml ──────▶ se-demo-app          (primary)
  (demo app + its                         se-demo-app-backup-2 (standby)
   slow workflows)                        se-demo-app-backup-3 (standby)
features/*.patch     ───────────────────▶   main: template/ contents (force-pushed)
                                            feature branches + draft PRs
```

- **A fixed pool of 3 repos**, tagged with the topic
  `blacksmith-se-demo-pool`: `se-demo-app` (the one shown in most demos)
  plus two standbys. Reusing a bounded set keeps the Blacksmith dashboard
  from filling with dead entries (Blacksmith has no way to remove a repo's
  runs, even after the GitHub repo is deleted). `reset.yml` creates a pool
  repo the first time it targets one, and resets it every time after.
- **`template/`** is the entire demo repo: a realistic demo-customer app
  (React/TS dashboard + Go API + Postgres, plus chunky data fixtures)
  with deliberately unoptimized GitHub Actions workflows — everything on
  `ubuntu-latest`, zero caching, matrix fan-out, per-run browser
  downloads. Only this directory reaches customers; keep everything
  inside it in-character.
- **`features/NN-<name>.patch`** files each hold one commit adding one
  Blacksmith feature to the template (`git format-patch` output, commit
  message included). Reset applies each patch in the demo repo as branch
  `feature/<name>` and can open a draft PR (subject → title, body → body;
  the `NN` prefix controls PR order). You pick features at demo time by
  which PRs you open (or generate them live with the migration
  wizard/Codesmith — the PRs are the fallback). This base repo itself has
  a single branch: `main`.
- **Reset re-derives from the base repo's current `template/` + `features/`**,
  so any demo-flow change you make here propagates to a pool repo the next
  time you reset it — no separate sync step.
- **Root** (this level) is SE-facing only and never reaches customers:
  this README, the runbook, `CLAUDE.md`, and the workflows below.

## Workflows in this repo

| Workflow | What it does |
| --- | --- |
| `SE: Reset demo repo` | Resets a pool repo (`primary` / `backup-2` / `backup-3` / `all`) to the current template: force-pushes `main`, recreates feature branches (+ optional draft PRs), seeds baseline CI history. Creates the repo on first use. |
| `SE: Generate CI activity` | Pushes N trivial commits to a pool repo's `main` (`primary` / `backup-2` / `backup-3` / `all`) to trigger its workflows on demand — quick CI activity / warmup runs without hand-committing. Does not reset. |
| `SE: Decommission demo pool` | Deletes every repo tagged `blacksmith-se-demo-pool` (dry-run by default) — the rare full teardown |
| `Validate template` | Base-repo CI: fast correctness checks on `template/` changes |

## Setup (one-time)

1. Create the org secret `DEMO_PROVISION_TOKEN`: a PAT with `repo`,
   `workflow`, and `delete_repo` scopes on `jacob-buckles-org`.
2. Run **SE: Reset demo repo** once per pool repo (`primary`, then
   `backup-2`, `backup-3`) to create them.
3. Install the Blacksmith GitHub App on the 3 pool repos (or rely on an
   org-wide install). It stays installed across resets.

## Resetting for a demo

Actions → **SE: Reset demo repo** → pick a `target` (usually `primary`).
**Reset the day before the call** — baseline history needs wall-clock
time on GitHub's queue. Because the app is already installed, demos start
at the migration wizard rather than app-install. Details, knobs, and the
demo arc are in the runbook.

## Conventions

- Feature patches touch only `template/` paths and contain exactly one
  commit. Their commit messages are customer-visible (subject → PR
  title, body → PR body) — write them in-character.
- `Validate template` CI applies every patch to current `main`, so a
  template change that breaks a patch fails loudly instead of at
  provision time.
- Editing a feature patch:

  ```sh
  git checkout -b tmp main
  git am features/01-migrate-runners.patch   # patch -> commit
  # ...edit files, then: git commit -a --amend
  git format-patch -1 --stdout > features/01-migrate-runners.patch
  git checkout main && git branch -D tmp
  # commit the regenerated patch on main
  ```

- The "before" CI must be slow *authentically* — the timing knob is
  `METRICS_WORKLOAD` in the template workflows (CPU-bound test sweeps, no
  sleeps), so faster runners genuinely speed it up.
