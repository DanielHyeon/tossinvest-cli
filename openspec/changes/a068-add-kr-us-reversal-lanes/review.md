# Review — a068-add-kr-us-reversal-lanes

- Date: 2026-08-03
- Stage: paired KR/US focused implementation complete; repository integration gates pending
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
- RED-to-GREEN captured strict/duplicate JSON, exact timestamp boundaries, causal structure,
  immutable allocation, a066 cap/FX provenance, stop non-retreat, replay and market-isolation branches.
- Missing fill identity uses a campaign/leg/order/quantity/price/fee/time/source/FX preimage digest:
  identical raw retries are idempotent and distinct unidentified fills remain distinct evidence.
- A066 risk-cap and frozen-FX sealers plus authority identifiers are package-private. Risk caps are
  sealed to the exact immutable plan digest, reject zero quantity/risk bases, and bind the proposed
  reservation quantity exactly to the evaluated leg quantity.
- Plan-state mismatch, corrupt held/release accounting and 256-bit filled-risk overflow preserve the
  event and prior accounting, latch unknown risk, and block all subsequent exposure raising.
- Missing-ID and zero-quantity fills preserve existing held/filled values; their unknown-risk latch is
  the conservative control that prevents any new admission until authoritative reconciliation.
- Missing cancel identity uses the campaign/leg/order/release/time/source preimage digest. Exact raw
  retries are idempotent, distinct unidentified observations remain separate, and neither releases
  held risk without authoritative identity.
- Zero-quantity fill evidence remains non-applied even when its FillID conflicts with an earlier
  positive fill; the conflict latches unknown risk without synthesizing positive-fill accounting.
- `go test -count=1 -race ./internal/reversallane`: PASS.
- `go vet ./internal/reversallane`: PASS.
- Strict OpenSpec validation for both a067 continuation and a068 reversal: PASS.
- Allocation and fill-retry fuzz targets (2 seconds each): PASS.
- Dependency/source authority tests and standard-library-only dependency closure: PASS.

## Verdict

Focused paired implementation is ready for broader repository regression and the root-owned
`make sdd-sync`, `make sdd-check`, and `make gate CHANGE=a068-add-kr-us-reversal-lanes` gates.
