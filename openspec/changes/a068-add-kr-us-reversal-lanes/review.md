# Review — a068-add-kr-us-reversal-lanes

- Date: 2026-08-03
- Stage: proposal freeze; implementation not started
- Voices: Manager strategy/safety review, independent architecture/test review, round-2 adversarial review

## Findings and disposition

- KR absorption and US dislocation use separate strict schemas and exact ppm arithmetic.
- Final-leg structure is the same scope and bounded causal order:
  `sweep_at <= break_at <= reclaim_at <= evaluated_at`.
- Evidence time is exact: `effective_at <= observed_at <= ingested_at <= evaluated_at <= fresh_until`;
  equality, one-tick stale and reversed-order branches are RED fixtures.
- The immutable 2:4:8 plan uses first-two floor plus final remainder, never reallocates unused quantity,
  preserves the conservative monetary-risk floor and leaves exit decisions to the common engine.

## Verification

- Strict OpenSpec validation: PASS.
- KR and US same-release/same-wave conformance is explicit.
- Price decline alone cannot authorize the final leg; LIVE/toggle/broker authority remains absent.

## Verdict

Proposal freeze approved for paired KR/US RED implementation.
