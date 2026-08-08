# Proposal-freeze review

- Date: 2026-07-30
- Risk: Small — Python SDD advisory tooling, tests, and workflow documentation only
- Voices: Manager self-review with failure-mode and developer-experience perspectives
- Product runtime impact: none; no Go product code, account state, order path, Guardian, journal, or installed binary change

## Evidence reviewed

- Two live `gbrain serve` processes had the same
  `/mnt/D/project/axipient/TossOS/.sdd/gbrain-home`.
- PID 479088 owned the refreshed PGLite lock and had the database files open.
- PID 2920715 was a same-home `serve` child without the PGLite lock.
- `python3 tools/sdd/gbrain_project.py sources list` reproduced
  `gbrain sources: connect timed out` after about 10 seconds.
- Repository CodeGraph identified `gbrain_project.py`, `sdd_sync.sync`, and their
  tests as the complete project-local assembly surface.
- GBrain source confirms PGLite is single-connection and its lock already has
  stale-owner recovery; deleting `.gbrain-lock` in TossOS would duplicate and
  weaken that ownership contract.

## Findings and decisions

1. **Blind stale-lock deletion** — rejected. A currently refreshed live owner was observed.
2. **Automatically kill the lock owner to make CLI sync work** — rejected. It would disrupt
   the active agent that legitimately owns the advisory engine.
3. **Move to HTTP/PostgreSQL in this hot path** — deferred. It enables simultaneous GBrain
   clients but adds credentials, service lifecycle, and data migration beyond this incident.
4. **Kernel singleton in the wrapper** — accepted. It prevents a second process before
   PGLite, releases on crash, and needs no PID-file cleanup.
5. **Busy as nonblocking advisory in sdd-sync** — accepted only for exact exit 75/marker
   contention. Other GBrain failures remain visible and incomplete.
6. **Legacy owner bridge** — accepted. A live PGLite lock created before deployment must be
   recognized so a new wrapper does not acquire its own lock and then wait inside GBrain.

## Requirement re-review — stale heartbeat

Implementation review found that PID liveness alone would be stricter than GBrain upstream:
a dead holder's PID can be reused by an unrelated process while its heartbeat remains stale.
The legacy bridge requirement was amended to require both live PID and heartbeat within the
same upstream steal grace (default 600 seconds). Stale/malformed locks are not deleted by the
wrapper; they are delegated to GBrain's token-checked recovery. A subprocess regression with a
live recycled PID and 700-second-old heartbeat is GREEN. Decision: amendment approved.

## Gate classification

- Proposal/design/spec are consistent with GBrain's advisory rank.
- The change does not weaken CodeGraph hard-evidence freshness.
- Regression tests must execute the wrapper as a subprocess, not only mock `flock`.
- Recovery must target only the verified same-home non-owner GBrain child.
- Function Logic Map: not-applicable — only Python SDD tooling is modified; no existing Go
  function or high-risk runtime function is touched.

Decision: approved for RED/GREEN implementation.

## Independent implementation review

- Reviewer: `hotfix_maintainability_review`
- Date: 2026-07-30
- Initial findings:
  1. A general nonzero source probe could continue and incorrectly record GBrain freshness.
  2. Source registration failure could still reach sync and incorrectly record GBrain freshness.
  3. A legacy-owner diagnostic could expose arbitrary command arguments.
- Resolution:
  1. Nonzero probes now record hard-evidence freshness only and return immediately.
  2. Registration failures and exceptions now return before sync and never record GBrain
     freshness.
  3. Legacy diagnostics now emit only an allowlisted constant subcommand; a subprocess
     regression verifies that option values are not exposed.
- Re-review evidence: 15 focused tests, Python compilation, and `git diff --check` passed.
- Final verdict: **NO FINDINGS**.
