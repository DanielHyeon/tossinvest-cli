# Review — a061-show-history-instrument-names

Status: approved — final security, test, and maintainability reviews report no A061 findings

## Pre-edit scope

- Read-only history projection and template only.
- New dependency is batch stock metadata with no mutation method.
- No journal schema, frozen performance value, input, script, route, toggle, or order change.

## Function Logic Map

`Console.history` and every modified existing command/concurrency function are
bound to regenerated AST, branch, and risk evidence under
`analysis/function-logic/`. `check_analysis.py` passes. Branchless functions keep
the gate-required synthetic `B1` happy-path row.

## Independent findings and resolution

- The first red-team pass found stale public-client/reset maps after the adapter
  was simplified. All affected design, Function Logic Map, Branch Test Map, AST,
  and risk artifacts were regenerated against the final single account-resolved
  client design.
- The first security/red-team pass found a multi-owner advisory-marker race,
  verification-start starvation, and stale abort targets. All three supervised
  entrypoints now order execution flock → profile intent marker → kernel
  rate-budget lease. Abort is refused while another verifier is active and reloads
  the evidence record after exclusive admission. Behavior and race tests cover
  cancellation, marker ownership, broker avoidance, and target addition/removal.
- Final red-team review found that the older advisory marker helper swallowed a
  publication error. A061 admission now publishes its marker through a strict
  path and refuses before lease/broker construction when the marker is unwritable;
  a read-only stale-marker regression test covers the failure.
- Partial official responses now cache accepted names only and retry only omitted
  keys after backoff. Dedicated tests prove the accepted cache survives.
- Credential rotation of the pre-existing shared console client was reported by
  the first broad review. Comparison with the persisted base confirms the
  `SaveOpenAPI` behavior and long-lived broker cache are unchanged by A061; the
  final diff-scoped security review classifies it as pre-existing/out of scope.
- Final security review: **NO FINDINGS** for the A061 diff.
- Final test-evidence review: **NO FINDINGS** after direct list/empty-abort
  broker-avoidance and late-addition/late-settlement assertions were added.
- Final maintainability/red-team review: **NO FINDINGS** after strict intent
  publication and regenerated evidence were verified.

## Verification evidence

- Focused regular and race suites for `cmd/tossctl`, `internal/console`, and
  `internal/ratebudget`: pass.
- `go vet ./...`: pass.
- Strict A061 OpenSpec and Function Logic Map checker: pass.
- Full repository tests: pass.
- Strict validation of all 58 OpenSpec items: pass.
- `make sdd-sync` and `make sdd-check`: pass (hard-evidence index matches this worktree).
- Full `make gate CHANGE=a061-show-history-instrument-names`: recorded immediately below when executed.
