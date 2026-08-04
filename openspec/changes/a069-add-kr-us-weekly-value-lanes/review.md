# Review — a069-add-kr-us-weekly-value-lanes

- Date: 2026-08-04
- Stage: adversarial hardening complete; repository gate pending
- Voices: Manager strategy/safety review, independent architecture/test review, round-2 adversarial review

## Findings and disposition

### Independent adversarial review remediation

- **Resolved:** stale cap/FX self-time validation now uses disclosure `evaluated_at`; cap reservation quantity is exact.
- **Resolved:** stable week is exactly derived from market/exchange + IANA Monday ISO week, and reserve uses a sealed trusted evaluation time.
- **Resolved:** reservation CAS version, leg count and consumed ordinals are campaign+market scoped, so KR and US never share counters.
- **Resolved:** only sequential distinct ordinals 1..7 consume; the public reservation-only transition rejects positive fill.
- **Resolved:** decoded evidence, dormant evaluation authority, reservation state and stop candidates are private-sealed; post-decode/literal mutation fails closed.
- **Resolved:** decision digest covers the complete immutable filing/revision/PIT/dilution/financial/model snapshot.
- **Resolved:** positive fill returns one reservation+risk aggregate, retaining consumed fill plus unknown/overage latch atomically.
- **Resolved:** mutable exported RiskState fields/maps were made private and canonical-sealed; scalar mutation and copied-map latch clearing now fail closed in both admission and fill application.
- **Resolved:** outcome lineage now carries superseded revision/sequence, calendar generation/session, position generation, risk budget, leg ceiling and cap/FX identifiers.

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
- Package tests, race detection and vet: PASS after both adversarial hardening rounds.
- Allocation and missing-fill fuzz tests (2 seconds each): PASS.
- Evidence/campaign/risk/strategy/exit/scheduler regression selection: PASS.
- Pure-package dependency inspection (no source API, broker, journal, runtime, toggle or exit imports): PASS.
- Full-worktree regression completed with a069/strategyrouter green; only concurrent journal changes failed `internal/position` eligibility-spelling enforcement outside a069 ownership.
- `make sdd-sync` and one `make sdd-check` passed. A final check after documentation updates reported the shared fingerprint stale while hard `codegraph status .` remained up to date at HEAD `23794f8626a20691431d5452b76e800255b0ee74`. The change gate passed task/review checks and stopped at Function Logic Map step only for concurrent `internal/journal/**` functions owned by another workstream; no a069 map finding remained.

## Verdict

Paired KR/US implementation satisfies the frozen contracts and passed its independent hardening review.
Final shared integration remains conditional only on the root-owned journal repair, full regression rerun
and change-gate rerun after concurrent changes settle.
