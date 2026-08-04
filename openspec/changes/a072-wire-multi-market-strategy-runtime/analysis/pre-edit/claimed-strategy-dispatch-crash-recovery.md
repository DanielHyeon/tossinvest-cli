# Claimed strategy-dispatch crash recovery

## Scope

This slice closes only a durable `CLAIMED + RESERVED` lease whose
`transport_started_at` is still `NULL`. It applies to KR and US in the same
implementation and test wave. It creates no lease, activation, approval,
Gateway call, broker request, retry, resend, or `SUBMITTING` outcome authority.

## Frozen contract

`RecoverClaimedStrategyDispatchLease` receives the newly acquired current
central-owner CAS and the exact old lease revision. In one transaction it must:

1. prove the caller is the current owner epoch/fence;
2. load the exact lease and require `CLAIMED + RESERVED`, the expected revision,
   and no durable transport-start marker;
3. prove the v26 first-leg binding and the exact aggregate plus five monetary
   reservations remain normalizable and unmapped;
4. terminally write `REFUSED + RELEASED` with
   `RECOVERY_CLAIMED_NO_TRANSPORT`; and
5. prove the aggregate and all five dimensions are released before commit.

The lease retains its original owner epoch/fence as immutable issue provenance.
The current recovery owner is proved by the current-owner row in the same
transaction and cannot use this API on `ISSUED`, `SUBMITTING`, or terminal rows.

## Rollback and replay

- Any proof, update, transition-log, or commit failure preserves the full
  `CLAIMED + RESERVED` preimage and all six holds.
- A repeated call returns `ErrStrategyDispatchLeaseConsumed` and does not alter
  the terminal outcome.
- A stale recovery owner returns `ErrStrategyDispatchFenced` before mutation.
- KR and US cases must pass the same table-driven RED/GREEN suite; neither
  market is reported complete independently.

