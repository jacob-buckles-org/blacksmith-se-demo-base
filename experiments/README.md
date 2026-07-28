# experiments/

Throwaway instruments for personal learning. **Nothing here is part of the
demo** — files in this folder are not in `app/` or `workflows/`, so they never
ship into `se-demo-app` / `se-demo-app-unmigrated`, and `Validate` ignores them.

## cache-benchmark.yml

Confirms whether Blacksmith's colocated cache restores faster than GitHub
Actions' cache, holding payload size constant.

1. Copy `cache-benchmark.yml` into an app-installed repo's
   `.github/workflows/` (`se-demo-app` works). It has to live in an
   app-installed repo because one of its two jobs runs on a Blacksmith runner.
2. Dispatch it once (Actions → Cache benchmark → Run) — both jobs miss and
   populate their cache (~300 MB by default; tune with the `payload_mb` input).
3. Dispatch it again — both jobs hit. Compare **restore wall time** in the two
   jobs' logs: the `ubuntu-latest` job used GitHub's cache, the
   `blacksmith-*` job used Blacksmith's colocated cache.

Delete it from the demo repo when you're done — it's not meant to live there
permanently.
