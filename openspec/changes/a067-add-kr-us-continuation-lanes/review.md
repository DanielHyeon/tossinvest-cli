# Review — a067-add-kr-us-continuation-lanes

- Date: 2026-08-03
- Stage: proposal freeze; implementation not started
- Voices: Manager strategy/safety review, independent architecture/test review, round-2 adversarial review

## Findings and disposition

- KR flow and US participation remain different versioned schemas with exact integer ppm arithmetic,
  freshness, units, threshold/config digests and independent lineage.
- The lane emits only entry/add decisions or typed invalidation/refusal; the common exit engine remains
  the sole exit-decision authority and evaluates independently.
- The immutable 8:4:2 plan uses `floor(Q*8/14)`, `floor(Q*4/14)` and final remainder. Partial fill,
  retry, cancel or unused quantity never reallocates upward.
- Round 2 froze campaign loss-to-stop monetary usage in account valuation minor units, conservative
  reservation floors, official FX identity/direction/as-of/haircut and checked overflow refusal.

## Verification

- Strict OpenSpec validation: PASS.
- KR and US registry conformance requires both lanes in the same release and rejects one-sided builds.
- No broker, journal writer, operating toggle or LIVE activation capability is injected into either lane.

## Verdict

Proposal freeze approved subject to the common a064–a066 contracts entering first; KR operational
stability is not a prerequisite for US implementation.
