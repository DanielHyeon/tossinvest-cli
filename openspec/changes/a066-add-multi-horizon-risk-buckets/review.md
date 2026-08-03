# Review — a066-add-multi-horizon-risk-buckets

- Date: 2026-08-03
- Stage: proposal freeze; production implementation not started
- Voices: Manager scope/safety review, independent adversarial risk review, final semantic re-review

## Findings and disposition

- **Accepted:** bucket dimensions are horizon/market/strategy/sector/symbol and canonical quantity fields
  are `q_candidate` and `q_final`.
- **Accepted:** HELD is a monetary reservation at worst executable price plus worst fees and fresh official
  FX haircut/ceil. `q_final` is fixed before final RiskIntent and GuardianDecision are issued atomically.
- **Accepted:** owner release requires the prior generation's mutations, protection, sell claims,
  reduce-only attempts and unresolved fills to be authoritatively clean.
- **Accepted after round 2:** each deduplicated fill transfers proportional HELD and records
  `filled=max(transferred conservative amount, actual monetary exposure)` in every applicable bucket.
  Overage or unknown actual price/fee/FX preserves the fill/Position and latches only new exposure as
  `RISK_OVERAGE` or `UNKNOWN_ACTUAL_RISK`.

## Verification

- Strict OpenSpec validation: PASS.
- Final independent semantic re-review: PASS, no open blocker.
- Property/crash tasks cover partial fills, replacements, predecessor late fills, retries and atomic replay.

## Verdict

Proposal freeze approved for RED implementation. Missing official FX evidence yields zero exposure-raising
quantity and never delays fill, reconciliation, protection or reduce-only exit.
