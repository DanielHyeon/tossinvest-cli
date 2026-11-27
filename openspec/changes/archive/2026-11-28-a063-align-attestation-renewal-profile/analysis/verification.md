# Implementation verification

All feature tests use temporary directories, local fake HTTP servers, or a
loopback console. They do not call a live broker.

| Command | Result | Evidence |
|---|---|---|
| `go test ./internal/soak ./cmd/tossctl ./internal/console` | pass, exit 0 | Latest run: soak 0.128s, CLI 49.330s, console 45.331s (before the final portable-reader test skip adjustment; a rerun is pending below). |
| `make test` | pass, exit 0 | Completed before FIFO/platform follow-up edits; must be rerun against final source. |
| `make vet` | pass, exit 0 | 2026-09-05. |
| `make validate` | blocked, exit 2 | a063 validated. Unrelated `a119-codex-session-handoff-and-gbrain-startup` failed the all-change validation. |
| `make sdd-sync` | pass, exit 0 | CodeGraph synced 13 changed files; GBrain advisory was busy, previous freshness retained. |
| Windows `GOOS=windows GOARCH=amd64 go test -c ./internal/soak` | pass, exit 0 | Secure reader intentionally compiles as unsupported/unknown diagnostics. |
| Darwin release command build | blocked | Existing `internal/strategyaccount/production_owner_unix.go:17` references undefined `productionFileUID`; not introduced by this change. |

Observed REDs during implementation were fixture/compile mistakes: missing path
argument for the new writer, then an empty endpoint fixture. They were corrected
and do not establish pre-change behavioral failure. The feature regression tests
now cover issued/refused status, override rejection, writer failure retaining a
prior status, symlink/mode/FIFO rejection, 72-hour warning, stale/mismatched
unknown state, and status non-effect on attestation usability.

Pending at this snapshot: complete `make sdd-check`, rerun the focused matrix and
`make test` against final source, and run `make test-seams` plus `git diff --check`.

## Current-source rerun

`make test` completed with exit 0 after the path-snapshot fix and hostile-input
tests (2026-09-05). This is the current full-suite result; other final checks
are intentionally recorded separately when they are rerun on this same source.

`make test-seams && make test-race` completed with exit 0 on 2026-09-05. The
tagged full suite passed, followed by the configured race package sets and the
engine race selection.

## Administrative rerun

| Command | Exit | Current result |
|---|---:|---|
| `make vet` | 0 | Passed. |
| `make sdd-sync` | 0 | CodeGraph synchronized; GBrain freshness remained advisory/busy. |
| `make sdd-check` | 0 | Passed; advisory GBrain source/freshness warnings only. |
| `make validate` | 2 | a063 passed; unrelated `a119-codex-session-handoff-and-gbrain-startup` failed strict all-change validation. |
| `make gate CHANGE=a063-align-attestation-renewal-profile` | 2 | Expected: tasks 4.1–4.5 remain unchecked, including human-approved install and operational proof. |

## Independent continuation verification (2026-09-06)

| Command | Exit | Current result |
|---|---:|---|
| `python3 tools/logic-map/check_analysis.py --change a063-align-attestation-renewal-profile` | 1 | 327 missing-evidence rows outside the a063 scoped bundles; direct validation of the eleven scoped bundles passed. |
| `make sdd-sync` | unknown | The CodeGraph phase reported `Already up to date`; the detached CodeGraphContext process ended, but this interface did not return its final exit status. It is not recorded as a pass. |
| `make sdd-check` | 0 | Passed. CodeGraph freshness matched; GBrain freshness/busy messages remained advisory. |
| `git diff --check` | 0 | Passed. |

The deployment proposal is reviewed but not executed. Same-profile operational
evidence and explicit human approval remain pending; archive and final
acceptance stay blocked.
