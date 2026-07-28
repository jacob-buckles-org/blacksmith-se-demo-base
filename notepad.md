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

## Still open

Is there a good way to track the evolution of the improvements as we
onboard onto the platform? Maybe Codesmith can tell me? Not a demo-repo
question anymore, but worth asking Blacksmith/Codesmith independently —
real prospects will ask this.

Could Codesmith guide me through onboarding new features?

