# Review — a073-operate-multi-market-strategy-lanes

- Date: 2026-08-03
- Stage: proposal freeze; implementation/deployment not started
- Voices: Manager scope review, independent operations/security/UI review, final semantic re-review

## Findings and disposition

- Console, JSON and SSE consume one server-owned market projection. KR and US have separate status/error
  envelopes, exact `WIRED|UNWIRED` readiness and honest unavailable/not-configured states.
- The existing authenticated GET-only strategy-runtime pattern is reused; no order, gate, activation,
  autostart, protection-weakening route or free input is added.
- Performance attribution requires the complete persisted market-to-close identity chain and conserves
  partial/staged-close quantity, basis, fees, taxes, FX and PnL. Missing facts are `link_missing` or
  `not_measured`, never zero/current-FX inference.
- Compose replacement pins image/schema/config/activation/protection/volume preimages, replaces one service
  at a time and reverse-rolls only the replaced subset. Incompatible rollback keeps entry OFF and forbids
  destructive downgrade.

## Design and DX disposition

The plan extends the existing `/strategy-runtime` information architecture rather than introducing a new
interaction model. Primary scan order is market identity → desired/effective/refusal → evidence/scheduler/
protection/reconciliation → campaign/risk → provenance/performance. Loading/unavailable, dormant, partial,
current and lineage-missing states are specified; mobile/accessibility and console/API parity reuse existing
golden-contract patterns. Operator time-to-diagnosis improves because every blocked market exposes its first
typed refusal without requiring journal joins.

## Verification

- Strict OpenSpec validation: PASS.
- Final independent operations/UI re-review: PASS, no open blocker.
- Deployment tests are read-only/dormant and prohibit live orders or operating-toggle changes.

## Verdict

Proposal freeze approved. Build/deploy authorization applies only to an OFF/unapproved dormant release;
market activation remains an explicit later human decision.
