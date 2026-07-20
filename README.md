# Blacksmith SE Demo Base

Internal demo factory for Blacksmith SE demos. This repo stamps out
fresh, one-per-demo customer repos that start in a deliberately slow
"before" state, ready for a live Blacksmith migration on a call.

**Start here:** [`DEMO_RUNBOOK.md`](DEMO_RUNBOOK.md) — talk track,
persona presets, timings, and per-feature demo notes.

## How it works

```
this repo                                per-demo repo
─────────                                ─────────────
template/            ── provision.yml ─▶ se-demo-<customer>
  (demo app + its                          main: template/ contents,
   slow workflows)                               fresh history
features/*.patch     ───────────────────▶ feature branches + draft PRs
```

- **`template/`** is the entire demo repo: a realistic demo-customer app
  (React/TS dashboard + Go API + Postgres, plus chunky data fixtures)
  with deliberately unoptimized GitHub Actions
  workflows — everything on `ubuntu-latest`, zero caching, matrix
  fan-out, per-run browser downloads. Only this directory reaches
  customers; keep everything inside it in-character.
- **`features/NN-<name>.patch`** files each hold one commit adding one
  Blacksmith feature to the template (`git format-patch` output, commit
  message included). Provisioning applies each patch in the demo repo as
  branch `feature/<name>` and opens a draft PR (subject → title, body →
  body; the `NN` prefix controls PR order). You pick features at demo
  time by which PRs you open (or generate them live with the migration
  wizard/Codesmith — the PRs are the fallback). This base repo itself
  has a single branch: `main`.
- **Root** (this level) is SE-facing only and never reaches customers:
  this README, the runbook, `CLAUDE.md`, and the workflows below.

## Workflows in this repo

| Workflow | What it does |
| --- | --- |
| `SE: Provision demo repo` | Creates `se-demo-<customer>` from `template/`, pushes feature branches, opens draft PRs, seeds baseline CI history |
| `SE: Teardown old demo repos` | Archives/deletes repos tagged `blacksmith-se-demo-generated` older than N days (dry-run by default) |
| `Validate template` | Base-repo CI: fast correctness checks on `template/` changes |

## Setup (one-time)

Create the org secret `DEMO_PROVISION_TOKEN`: a PAT with `repo`,
`workflow`, and `delete_repo` scopes on `jacob-buckles-org`.

## Provisioning a demo

Actions → **SE: Provision demo repo** → enter `customer_name`.
**Provision the day before the call** — baseline history needs
wall-clock time on GitHub's queue. Details, knobs, and the demo arc are
in the runbook.

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
