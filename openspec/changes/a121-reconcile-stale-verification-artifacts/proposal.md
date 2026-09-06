# a121 — Reconcile stale verification artifacts

## Why

An approved live verification cleanup received official
`conditional-order-not-found`, while the local append-only verification record
kept the conditional artifact outstanding. A subsequent operator observation
found no active orders, but it did not establish the artifact's identity,
terminal lifecycle, or causal relationship to the failed DELETE. The current
fail-closed behavior correctly avoids calling the 404 a successful cancel, but
it has no read-backed path to reconcile a stale local artifact.

This blocks a063 operational acceptance. It must be repaired without retrying
the live cleanup, creating a new verification record, weakening the engine
interlock, or treating a 404 as success.

## What changes

- Add a read-only, account-scoped reconciliation path for an artifact already
  owned by a verification record.
- Append a distinct, idempotent `reconciled-absent` local event only after a
  fresh, complete official conditional-order read proves the same record-owned
  artifact absent under explicit identity, account, object-type, and pagination
  rules.
- Preserve the failed cleanup call and its failure. Reconciliation is neither a
  cancel, fill, endpoint success, nor capability-attestation evidence.
- Keep absence ambiguous or unproven when the official read is stale, partial,
  wrong-account, wrong-type, identity-mismatched, or otherwise unverifiable.

## Non-goals

- Retry, cancel, amend, create, or otherwise mutate any live order.
- Infer a broker terminal lifecycle from DELETE 404 or a generic operator view.
- Reclassify failed verification steps or promote reconciliation into soak or
  engine-interlock capability coverage.
- Complete, archive, or relax a063. a063 remains blocked until its own
  operational evidence is independently satisfied.

## Impact

- `internal/verifylive` append-only record, outstanding-artifact projection,
  and verification CLI surface.
- Official conditional-order read adapter and focused contract tests.
- `order-execution` specification only; runtime trading, risk limits, and
  engine startup semantics remain unchanged.
