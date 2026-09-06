# a063 task 4.2 — conditional-cleanup 404 diagnosis

**Verdict: local verification state is stale with respect to the supplied
account-state observation.  It is not safe to infer the broker-side terminal
transition from the cleanup response alone.**

This is a diagnostic record, not a repair, attestation, or task completion.

## Scope and safety boundary

The manager supplied two authoritative, redacted runtime facts for this
diagnosis:

1. A human-approved resumed verification sent exactly one cleanup request for
   the prior conditional artifact.  The official endpoint returned HTTP 404
   with the broker code `conditional-order-not-found`.
2. A subsequent human account observation found no active orders.  No
   identifiers are retained here.

The first fact proves an accepted request reached the official API and that the
API did not find the requested conditional identifier at its handling point.
The second is broker-state evidence that the local record's remaining
outstanding artifact is stale/inconsistent with the observed account state.

Neither fact, alone or together, identifies *why* the conditional is absent.
This record therefore does not label it cancelled, expired, modified,
triggered, or filled.  It also does not infer a child order or position change.

The diagnosis did not run `verify run`, `--resume`, `abort`, any order or
engine command, a systemd action, a survey action, or a configuration mutation.
It did not read the raw JSONL record.  Read-only `verify status` and
`verify report` were rendered through redaction only; they still derived one
outstanding artifact from the local record.

## Code evidence

### Cleanup lifecycle

- `Runner.Run` approves the planned batch and invokes the cleanup prologue
  before the catalogue ([runner.go](../../../../../internal/verifylive/runner.go)
  lines 492-515).  A resumed run therefore may repeat a cleanup target even
  when normal measurement steps are already terminal.
- `PendingCleanup` selects a held conditional once its
  `conditional-cancel` gate has a later terminal verdict.  A failed
  conditional-cancel is terminal for resumption purposes, so it releases the
  held artifact into the cleanup plan ([cleanup.go](../../../../../internal/verifylive/cleanup.go)
  lines 119-190; [record.go](../../../../../internal/verifylive/record.go) lines
  90-118).
- The cleanup request uses `cancelConditional`.  That method appends the
  endpoint call, but appends the artifact's `Cancelled=true` record event
  **only when** `CancelConditionalOrder` returns nil
  ([mutate.go](../../../../../internal/verifylive/mutate.go) lines 617-644).
  The official client returns a non-nil API error for a non-2xx DELETE response
  ([conditional_writes.go](../../../../../internal/official/conditional_writes.go)
  lines 10-36).  No existing classification maps this 404 to a successful
  cancellation.
- `Outstanding` is record-derived: it keeps an artifact outstanding until its
  newest event is `Cancelled` or `Filled`
  ([record.go](../../../../../internal/verifylive/record.go) lines 500-575).
  Consequently, the observed 404 leaves the old artifact outstanding in the
  local record even if the broker no longer has an active object.
- The intended fail-closed behavior is covered by
  `TestAFailedCleanupIsRecordedAndDoesNotStopTheRun` and
  `TestAFailedCancelAfterTheHoldStillReleases`: a failed cleanup remains
  visible and is eligible for another planned cleanup; it is never silently
  removed ([cleanup_test.go](../../../../../internal/verifylive/cleanup_test.go)
  lines 179-218, [hold_test.go](../../../../../internal/verifylive/hold_test.go)
  lines 106-126).

### Resume and attestation effects

- A normal `verify run` refuses to overwrite an existing record; `--resume`
  reuses it so completed measurements do not place a second live order
  ([verify.go](../../../../../cmd/tossctl/verify.go) lines 305-328; [runner.go](../../../../../internal/verifylive/runner.go)
  lines 701-715).  Repeating it now would plan another live cleanup attempt
  against an already observed-absent broker object, so it is not the safe
  reconciliation mechanism.
- A failed endpoint call does not become successful supervised evidence.  The
  only order mutations that the read-only soak may borrow for the engine gate
  are order placement and order cancellation; conditional-order endpoints are
  deliberately excluded ([soak.go](../../../../../internal/soak/soak.go) lines
  88-101; [soak.go](../../../../../cmd/tossctl/soak.go) lines 427-473).
  Thus this conditional 404 neither proves a cleanup nor repairs the missing
  engine attestation.
- A future reconciliation must preserve the original failed call and its
  digest.  It must not retroactively convert the HTTP 404 itself into a
  successful endpoint measurement or use it as capability-attestation evidence.

## Unknowns that remain deliberately unresolved

- The official 404 response does not establish whether the resource was
  cancelled, expired, replaced, or triggered.
- The human account observation establishes no active orders at its observation
  time, but this diagnostic contains no retained identifier, terminal status,
  or causality evidence that would explain the absent conditional.
- Current source and tests contain no reconciliation path that consumes an
  authoritative broker read plus the recorded artifact and appends a distinct,
  auditable terminal reconciliation event.  In particular, there is no tested
  rule that treats a DELETE 404 alone as closure.

## Safe next action

Do **not** retry, cancel, resume, abort, or start the engine from this state.

The next SDD task is an OpenSpec-scoped reconciliation design and implementation
before any operational retry.  Its contract must require a read-only,
account-scoped authoritative absence observation with explicit freshness and
object-type semantics.  Only that observation—not the DELETE 404—may support a
new append-only `reconciled-absent` event.  The event must retain the failed
cleanup evidence, state its basis, be idempotent, and cause `Outstanding` to
stop presenting this stale artifact without claiming a cancellation, fill, or
attestation success.

Required test cases for that task include:

- DELETE 404 with no verified authoritative absence remains outstanding.
- DELETE 404 plus a matching, fresh authoritative absence reconciles exactly
  one matching artifact and preserves the original failure.
- Wrong account/profile, stale read, wrong object type, ambiguous list result,
  or unverifiable identity fails closed and leaves it outstanding.
- Re-running reconciliation is idempotent; resume then contains no cleanup
  mutation for a reconciled artifact.
- The reconciliation event cannot be counted by `SucceededEndpoints` or
  promoted into `soak attest` supervised proof.

After that change is independently reviewed, gstack-reviewed, and accepted,
the normal a063 sequence remains: establish valid same-profile evidence,
collect the required future renewal window, then perform the read-only final
attestation checks.  Tasks 4.2–4.5 remain open.

## Static verification

- `go test ./internal/verifylive` — PASS
- `go test ./internal/official ./internal/trading` — PASS
- CodeGraph callers/callees were checked for `cleanup`, `Outstanding`, and
  `CancelConditionalOrder`; direct source above was used for function-internal
  branch and side-effect evidence.
- `git diff --check` — PASS before writing this diagnostic; the existing
  unrelated a119 untracked paths were left untouched.
