# Review — a065-add-position-campaign-leg-core

- Date: 2026-08-03
- Stage: proposal freeze; production implementation not started
- Voices: Manager scope/safety review, independent adversarial ledger review, final semantic re-review

## Findings and disposition

- **Accepted:** campaign creation before first fill requires a prospective position-generation CAS and a
  set-once first-fill binding; manual/external position drift enters reconciliation rather than renumbering.
- **Accepted:** Campaign and Leg transition tables include rejection, cancel pending/cancelled, expiry,
  submit ambiguity, replacement, late fill and recovery terminality.
- **Accepted:** cumulative fill watermarks are broker-order scoped and replacement lineage aggregates
  predecessor/child deltas without assuming the new order continues the predecessor counter.
- **Accepted after round 2:** a late positive fill on a replaced/cancelled predecessor advances the immutable
  order watermark and Position exactly once in the same transaction. Cap or lineage ambiguity preserves
  the fill, recalculates remaining quantity and latches campaign `RECONCILE`/new-entry block.

## Verification

- Strict OpenSpec validation: PASS.
- Final independent semantic re-review: PASS, no open blocker.
- Exit-first and non-retreating stop authority remain outside the campaign policy and are not weakened.

## Verdict

Proposal freeze approved for RED implementation. The core is strategy neutral and grants no broker or
activation capability.
