# a063 task 4.2 conditional-cleanup 404 adversarial review

**Verdict: CLEAR — the diagnostic is fail-closed and does not turn the DELETE 404 or user observation into an unsupported cancellation claim.**

Re-reviewed diagnosis SHA-256 `cc7ffa73cf112958c36591a5392879c47968b50393814764d80d5fb8a4dd0b8a` and current static source/tests. The prior reviewed body is unchanged; the only normalization is removal of trailing blank EOF lines. No live `tossctl`, order, engine, systemd, survey, or configuration command was executed.

## Evidence and assessment

- The report correctly treats the two supplied facts separately: the official DELETE returned a non-success response for the named conditional identifier; the later human observation reported no active orders. Neither establishes the broker-side cause, a child order, a position change, or a terminal state for the local artifact.
- In `internal/verifylive/mutate.go`, `cancelConditional` logs the DELETE attempt and returns immediately on a non-nil error. It calls `sr.cancelled` only after a nil error. `internal/official/conditional_writes.go` propagates a non-2xx DELETE as an error. Therefore a 404 does not produce a local `Cancelled=true` event.
- `Outstanding` in `internal/verifylive/record.go` is append-only-record derived and only removes an artifact when its newest terminal event is `Cancelled` or `Filled`. The stale conditional thus remains visible after the failed cleanup, matching the report.
- Alternate terminal paths exist but are materially different from a DELETE 404: a successful conditional modification records the old identifier as cancelled and a successor as created; the trigger observation records the conditional and child as filled only after observing the triggering/child outcome. These paths reinforce the diagnostic’s decision not to label a 404 cancelled, modified, triggered, or filled.
- `cleanup.go` and its tests support the described retry hazard: a failed cleanup remains planned/outstanding rather than silently disappearing. A resumed verification would therefore be a new approved live cleanup attempt, not reconciliation.
- The report correctly limits the current human “no active orders” observation. Its proposed future reconciliation contract requires a fresh, account-scoped authoritative absence read with explicit object-type and identity semantics. This avoids treating a generic order observation as proof about a conditional resource or any successor.
- It correctly prevents the failed conditional DELETE from becoming supervised endpoint evidence: current soak borrowing is limited to the required ordinary order place/cancel evidence, not conditional endpoints.

The proposed next step is appropriately non-operational: an OpenSpec-scoped, append-only reconciliation design with fresh authoritative-read, identity/freshness, idempotency, and non-promotion tests. It expressly forbids retry/resume/new record/manual cancel until then, leaves tasks 4.2–4.5 open, and does not claim an archive or completion.

No dangerous false remedy, 404-as-success conversion, or confusion between “no active orders” and a proven conditional lifecycle was found.
