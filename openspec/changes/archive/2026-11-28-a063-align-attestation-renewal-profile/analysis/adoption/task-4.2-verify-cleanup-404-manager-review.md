# a063 task 4.2 conditional-cleanup 404 — Manager review

**Decision: accepted as a blocked diagnostic; no operational retry is allowed.**
Task 4.2 remains open and this record is not an attestation, a broker
reconciliation, a service activation, a final-gate result, or archive approval.

## Reviewed evidence

- The diagnostic is SHA-256
  `cc7ffa73cf112958c36591a5392879c47968b50393814764d80d5fb8a4dd0b8a`.
- The independent adversarial review is CLEAR at SHA-256
  `8002a771513909eb7506612335fe31c28226a9332066c2589346fa96fd81be68`.
- The subsequent gstack review is CLEAR at SHA-256
  `08b2ab41e66c0bc7783bd9b38d57e7f0109bf3412ba8c5e021f85f1035fa4e22`.
- A separate Terra verification reran `go test ./internal/verifylive`,
  `go test ./internal/official ./internal/trading`, strict a063 validation, and
  `git diff --check`; all passed.

The source confirms the fail-closed behavior: a conditional cleanup appends a
`Cancelled` artifact event only after a nil official DELETE result. A 404 leaves
the record outstanding. `Outstanding` and `PendingCleanup` therefore continue
to plan that artifact, and the current code has no append-only reconciliation
event backed by a fresh official read.

## Evidence boundary

The user observed no current active orders after the approved cleanup request
received `conditional-order-not-found`. This is a valuable operator observation
that exposes a mismatch with the local record, but it does not identify the
conditional object's terminal state, object type, identity match, freshness, or
causal relationship to the 404. It cannot make the 404 a successful cancel or
clear the local record.

No further `verify run`, `--resume`, `--redo`, `--record`, manual cancel, engine
start, or systemd action is approved from this evidence. A focused follow-on
OpenSpec change must first specify an account-scoped, fresh official read and a
distinct idempotent `reconciled-absent` record event. That future event must
retain the failed cleanup call and must not count as cancellation success or
engine-attestation coverage.

## Scope decision

This repair is outside a063: its approved requirements preserve the existing
qualification and interlock semantics and explicitly exclude order-side-effect
work. a063 stays blocked on the resulting capability/interlock evidence. The
follow-on change must be completed and accepted before a063's approved
operations can be reconsidered.
