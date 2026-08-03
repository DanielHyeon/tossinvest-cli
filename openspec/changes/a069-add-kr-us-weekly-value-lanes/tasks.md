## 1. Logic Mapping and RED Contracts

- [ ] 1.1 Before editing existing registry/evaluator/calendar/reservation functions, record Function Logic Maps and Branch Test Maps for filing revisions, market week, reservation, leg count/allocation, RR, invalidation and OFF
- [ ] 1.2 Add concurrent peer RED schemas for OpenDART KR and EDGAR US filing/revision/as-of/observed/ingested/freshness, currency/unit, diluted shares/dilution, model/config/fair-value fixed-point preimage; neither source waits for the other
- [ ] 1.3 Add RED future-revision, correction-chain, missing unit/currency/shares, overflow and deterministic model replay tests
- [ ] 1.4 Add RED official-calendar/IANA tests for canonical `(campaign_id, market, stable_market_week_identity)`, KR/US holiday and US DST, calendar correction generation A→B identity stability/no second slot, concurrent/replay atomic uniqueness, positive partial-fill consumption, zero-fill cancel/expiry release, same-idempotency retry, crash/restart and no server-local fallback
- [ ] 1.5 Add RED seven-leg/allocation/risk property tests for positive-fill ordinal count, immutable Q/ceilings, q_leg planned/a066 caps, exact account-valuation-minor `max(transferred reserve, ceil(qty*stop-distance + entry fees + estimated exit fees/levies))`, a066-frozen US FX quote/direction/haircut/ceil, filled floor plus held/proposed conservative reservations, checked overflow/refusal, actual overage or unknown-risk latch, partial fill/duplicate/cancel and no upward reallocation
- [ ] 1.6 Add RED exact RR tests for `reward=max(target-entry,0)*qty-costs`, `risk=max(entry-stop,0)*qty+costs`, `floor(reward*1e6/risk)`, risk<=0/overflow refusal, inclusive minimum threshold, same a066 FX snapshot, `target=min(staged,fair_value)`, structural-stop cap and typed invalidation/common-exit authority separation
- [ ] 1.7 Add RED same-release and market/source isolation tests proving one unavailable source/OFF market does not gate its peer

## 2. Concurrent KR and US Implementation

- [ ] 2.1 Implement strict OpenDART filing/model schema and pure `kr_weekly_disclosure_value_v1` evaluator
- [ ] 2.2 Concurrently with 2.1, implement peer EDGAR schema and `us_weekly_disclosure_value_v1`; this task MUST NOT depend on KR credentials, stabilization or activation
- [ ] 2.3 Implement official-calendar stable market-week identity and canonical three-part atomic reservation key, keeping calendar generation as evidence and preserving uniqueness across correction/replay
- [ ] 2.4 Implement immutable seven-leg plan/count, a066 cap, the frozen account-valuation-minor filled/held/proposed risk formula and same-snapshot US FX conversion, actual-fill overage latch, per-leg revalidation and no upward reallocation
- [ ] 2.5 Implement capped target, the exact checked RR/ppm formula with inclusive threshold and same-snapshot FX, stop non-retreat and typed invalidation/refusal only
- [ ] 2.6 Register both lane versions in one registry unit with independent market scope and default OFF

## 3. Integration and Safety

- [ ] 3.1 Integrate a064 disclosure/calendar evidence and a065/a066 campaign/risk contracts with filing, revision, week reservation, risk-budget and planned-leg lineage
- [ ] 3.2 Add concurrency/crash, correction, holiday/DST, partial-fill, replay, source-isolation and common-exit integration fixtures
- [ ] 3.3 Prove lanes do not call source APIs directly and have no broker, journal/owner writer, exit authority or operating-toggle dependency

## 4. VERIFY

- [ ] 4.1 Run targeted KR/US weekly, calendar/reservation, model/RR and allocation property/fuzz tests with race detection
- [ ] 4.2 Run broader evidence/campaign/risk/strategy/exit/scheduler regressions and confirm missing KR credentials do not gate US or create live mutation
- [ ] 4.3 Run `make sdd-sync`, `make sdd-check` and `make gate CHANGE=a069-add-kr-us-weekly-value-lanes`, recording both lanes remain OFF
