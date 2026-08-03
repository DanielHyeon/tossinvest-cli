# Review — a069-add-kr-us-weekly-value-lanes

- Date: 2026-08-03
- Stage: proposal freeze; implementation not started
- Voices: Manager strategy/safety review, independent architecture/test review, round-2 adversarial review

## Findings and disposition

- OpenDART KR and EDGAR US inputs preserve filing/revision/as-of/availability, units/currency,
  diluted shares/dilution facts, model/config digest and exact fair-value preimage.
- The target is `min(staged_target, fair_value)`. Reward/risk and cost-aware RR have checked fixed-point
  formulas, explicit fee/levy/FX placement, non-positive-risk refusal and inclusive threshold semantics.
- The weekly unique key is `(campaign_id, market, stable_market_week_identity)`; calendar generation is
  evidence only, so a calendar A→B correction cannot create a second allowance.
- Positive fill consumes the slot; authoritative zero-fill cancel/expiry releases it; retries and crashes
  cannot create another slot. Seven-leg count is based on distinct positive-fill ordinals.
- Actual filled monetary risk never falls below the transferred conservative reservation and cannot be
  reallocated upward after a partial fill.

## Verification

- Strict OpenSpec validation: PASS.
- KR/US disclosure outages are isolated and neither market waits for the other's stability.
- Missing disclosure, FX or model evidence means typed refusal and zero new exposure, not inferred value.

## Verdict

Proposal freeze approved for paired KR/US RED implementation.
