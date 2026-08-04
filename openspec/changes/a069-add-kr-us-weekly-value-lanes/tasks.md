## 1. Logic Mapping and RED Contracts

- [x] 1.1 Record Function Logic Maps and Branch Test Maps for evaluator, stop, calendar, reservation CAS, evidence validation and actual-risk transitions before the adversarial hardening edits
- [x] 1.2 Add concurrent peer RED schemas for OpenDART KR and EDGAR US filing/revision/as-of/observed/ingested/freshness, currency/unit, diluted shares/dilution, model/config/fair-value fixed-point preimage; neither source waits for the other
- [x] 1.3 Add RED future-revision, correction-chain, missing unit/currency/shares, overflow and deterministic model replay tests
- [x] 1.4 Add RED official-calendar/IANA tests for canonical `(campaign_id, market, stable_market_week_identity)`, KR/US holiday and US DST, calendar correction generation A→B identity stability/no second slot, concurrent/replay atomic uniqueness, positive partial-fill consumption, zero-fill cancel/expiry release, same-idempotency retry, crash/restart and no server-local fallback
- [x] 1.5 Add RED seven-leg/allocation/risk property tests for positive-fill ordinal count, immutable Q/ceilings, q_leg planned/a066 caps, exact account-valuation-minor `max(transferred reserve, ceil(qty*stop-distance + entry fees + estimated exit fees/levies))`, a066-frozen US FX quote/direction/haircut/ceil, filled floor plus held/proposed conservative reservations, checked overflow/refusal, actual overage or unknown-risk latch, partial fill/duplicate/cancel and no upward reallocation
- [x] 1.6 Add RED exact RR tests for `reward=max(target-entry,0)*qty-costs`, `risk=max(entry-stop,0)*qty+costs`, `floor(reward*1e6/risk)`, risk<=0/overflow refusal, inclusive minimum threshold, same a066 FX snapshot, `target=min(staged,fair_value)`, structural-stop cap and typed invalidation/common-exit authority separation
- [x] 1.7 Add RED same-release and market/source isolation tests proving one unavailable source/OFF market does not gate its peer

## 2. Concurrent KR and US Implementation

- [x] 2.1 Implement strict OpenDART filing/model schema and pure `kr_weekly_disclosure_value_v1` evaluator
- [x] 2.2 Concurrently with 2.1, implement peer EDGAR schema and `us_weekly_disclosure_value_v1`; this task MUST NOT depend on KR credentials, stabilization or activation
- [x] 2.3 Implement official-calendar stable market-week identity and canonical three-part atomic reservation key, keeping calendar generation as evidence and preserving uniqueness across correction/replay
- [x] 2.4 Implement immutable seven-leg plan/count, a066 cap, the frozen account-valuation-minor filled/held/proposed risk formula and same-snapshot US FX conversion, actual-fill overage latch, per-leg revalidation and no upward reallocation
- [x] 2.5 Implement capped target, the exact checked RR/ppm formula with inclusive threshold and same-snapshot FX, stop non-retreat and typed invalidation/refusal only
- [x] 2.6 Register both lane versions in one registry unit with independent market scope and default OFF

## 3. Integration and Safety

- [x] 3.1 Expose only sealed pure ports for a064 disclosure/calendar evidence and a065/a066 campaign/risk contracts; runtime, scheduler and journal wiring remains outside a069 scope
- [x] 3.2 Add concurrency/crash, correction, holiday/DST, partial-fill, replay, source-isolation and common-exit integration fixtures
- [x] 3.3 Prove lanes do not call source APIs directly and have no broker, journal/owner writer, exit authority or operating-toggle dependency

## 4. VERIFY

- [x] 4.1 Run targeted KR/US weekly, calendar/reservation, model/RR and allocation property/fuzz tests with race detection
- [x] 4.2 Run broader evidence/campaign/risk/strategy/exit/scheduler regressions and confirm missing KR credentials do not gate US or create live mutation
- [x] 4.3 Run `make sdd-sync`, `make sdd-check` and `make gate CHANGE=a069-add-kr-us-weekly-value-lanes`, recording both lanes remain OFF; shared gate result is pending concurrent journal logic-map/test repair recorded in status

## 5. Independent-review hardening

- [x] 5.1 Reject stale a066 cap/frozen FX at disclosure `evaluated_at` and require exact reserved decision quantity
- [x] 5.2 Derive stable market-week identity exactly from market, IANA-zone Monday session and ISO week; seal trusted reservation evaluation time
- [x] 5.3 Scope CAS version, positive-leg count and consumed ordinals by `(campaign_id, market)` and require sequential distinct ordinals 1..7
- [x] 5.4 Seal decoded evidence and cover the complete immutable revision/PIT/dilution/financial/model preimage in the decision digest
- [x] 5.5 Replace caller `Enabled` with an unexported dormant evaluation authorization; external callers cannot mint activation before a072
- [x] 5.6 Make positive-fill reservation consumption available only through one reservation+risk aggregate transition
- [x] 5.7 Seal stop candidate version/digest/freshness and add complete calendar/revision/position/leg/risk/cap result lineage
- [x] 5.8 Run package, race, fuzz, vet and selected evidence/campaign/risk/strategy/exit/scheduler regressions
- [x] 5.9 Seal private RiskState balances/latches/fills, expose read-only accessors, and reject scalar or copied-map tampering in admission and fill application
