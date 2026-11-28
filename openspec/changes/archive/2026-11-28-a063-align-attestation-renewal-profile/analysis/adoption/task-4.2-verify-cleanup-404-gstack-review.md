# a063 task 4.2 — conditional-cleanup 404 independent gstack review

**Verdict: CLEAR — diagnostic only.** The record correctly preserves the
unresolved conditional artifact and does not authorize a retry, reconciliation,
attestation, engine action, task completion, gate, or archive.

## Reviewed inputs

- diagnosis SHA-256:
  `cc7ffa73cf112958c36591a5392879c47968b50393814764d80d5fb8a4dd0b8a`;
- adversarial review SHA-256:
  `8002a771513909eb7506612335fe31c28226a9332066c2589346fa96fd81be68`;
- a063 delta requirements and `tasks.md`, where 4.2–4.5 remain unchecked;
- current cleanup, record, mutation, official-client, soak, and named unit-test
  source.

Mechanical re-review confirms the diagnosis body is unchanged; the sole change is removal of trailing blank EOF line(s). The refreshed adversarial re-review remains CLEAR.

Read-only checks performed: `git diff --check` returned `0`; the relevant
`internal/verifylive`, `internal/official`, and `cmd/tossctl` paths have no
uncommitted source diff. The two diagnostic/review records are untracked
metadata, and unrelated a119 work was not inspected as a change to this
finding. No test or live command was run during this review.

## Technical assessment

The diagnosis separates two supplied runtime facts: a human-approved
conditional DELETE returned official HTTP 404 (`conditional-order-not-found`),
and a later human observation found no active orders. The first establishes a
non-success response for the named identifier at request handling; the second
is time-bounded observation evidence. Neither identifies a broker lifecycle,
a successor, child order, fill, cancellation, or position effect. The report
therefore makes no broker-state claim beyond those inputs, which is the correct
scope.

Current behavior corroborates the fail-closed reading. `cancelConditional` in
`internal/verifylive/mutate.go` records the endpoint call then returns a
non-nil error without calling `sr.cancelled`; the official client propagates a
non-2xx DELETE as an error. `Outstanding` in `record.go` remains append-only
and considers only an artifact whose latest event is `Cancelled` or `Filled`
terminal. `PendingCleanup` intentionally releases a conditional after the
later failed `conditional-cancel` verdict, so a future resume would plan a new
live cleanup attempt rather than silently reconcile it.

The static tests directly reflect those claims:
`TestAFailedCleanupIsRecordedAndDoesNotStopTheRun` asserts a failed cleanup is
not a pass and remains outstanding, while
`TestAFailedCancelAfterTheHoldStillReleases` asserts that the failed conditional
is a cleanup target. The diagnosis also retains its reported test commands as
PASS, but no command-output capture was supplied here; this review relies on
the current static test contracts and does not independently certify a new test
execution.

`soak.LiveOnlyEndpoints` permits only ordinary order placement and cancellation
for supervised borrowing. Conditional endpoints are excluded, so the 404 cannot
be promoted into a successful endpoint measurement or fresh attestation proof.

## Boundary and next work

The proposed repair is correctly outside a063 operational completion: it needs
a separately specified reconciliation contract with a fresh, account-scoped,
object-type-specific authoritative absence read, verifiable identity and
freshness, append-only/idempotent event semantics, and tests that reject stale,
wrong-account, ambiguous, or unverifiable observations. The 404 alone must
remain nonterminal, original failure evidence must remain preserved, and a
reconciled-absent event must not count as cancellation, fill, or soak proof.

Until such a change is designed, implemented, independently reviewed, and
accepted, retry/resume/new-record/manual cleanup remains out of scope and
requires separate human authorization. No finding blocks keeping this
fail-closed diagnostic record.
