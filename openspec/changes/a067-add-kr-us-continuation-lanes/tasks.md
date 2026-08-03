## 1. Logic Mapping and RED Contracts

- [ ] 1.1 Before editing any existing registry/evaluator/dispatch-adjacent function, record its Function Logic Map and Branch Test Map for OFF, market/schema mismatch, allocation, invalidation and stop branches
- [ ] 1.2 Add concurrent peer RED schema tests for KR flow and US participation fields, units, timestamp ordering/freshness, threshold/config digest, overflow and exact ppm boundary arithmetic; neither market test waits for the other
- [ ] 1.3 Add RED allocation/risk property tests for immutable budget/Q, weights/14, floor+final remainder, q_leg planned/a066 caps, exact account-valuation-minor `max(transferred reserve, ceil(qty*stop-distance + entry fees + estimated exit fees/levies))`, a066-frozen US FX quote/direction/haircut/ceil, filled floor plus held/proposed conservative reservations, checked overflow/refusal, actual overage or unknown-risk latch, partial fill/duplicate/cancel and prohibited upward reallocation
- [ ] 1.4 Add RED authority tests proving lanes emit typed invalidation/refusal but never an exit decision/order, while the common exit engine independently proceeds and invalidation suppresses every add
- [ ] 1.5 Add RED same-release/market-isolation tests proving both lane versions exist and either market OFF/closed/failure does not gate the peer

## 2. Concurrent KR and US Implementation

- [ ] 2.1 Implement the strict versioned KR flow schema and pure `kr_short_flow_continuation_v1` evaluator with checked integer arithmetic
- [ ] 2.2 Concurrently with 2.1, implement the peer US participation schema and `us_short_participation_continuation_v1`; this task MUST NOT depend on KR stabilization or activation
- [ ] 2.3 Implement immutable 8:4:2 ceilings plus the frozen account-valuation-minor filled/held/proposed monetary-risk formula, same-snapshot US FX conversion and actual-fill overage/unknown-risk latch without blocking fill/common exit
- [ ] 2.4 Apply fresh continuation confirmation, a066 cap and stop non-retreat, returning only entry/add decision or typed invalidation/refusal
- [ ] 2.5 Register both lane ID/versions in one registry unit with independent market scope and default desired/effective OFF

## 3. Integration and Safety

- [ ] 3.1 Integrate a064 evidence and a065/a066 campaign/risk contracts while preserving schema/config, risk-budget, planned-leg and market lineage
- [ ] 3.2 Add replay, cross-market contamination, stale evidence, partial-fill and invalidation/common-exit integration fixtures
- [ ] 3.3 Prove dependency closure contains no broker, journal/owner writer, exit authority or operating-toggle writer and dormant registration changes no LIVE/activation state

## 4. VERIFY

- [ ] 4.1 Run targeted KR/US continuation, registry, property/fuzz and common-exit tests with race detection in the same verification run
- [ ] 4.2 Run broader evidence/campaign/risk/strategy/exit/scheduler regressions and confirm zero live hostname mutation, toggle or approval changes
- [ ] 4.3 Run `make sdd-sync`, `make sdd-check` and `make gate CHANGE=a067-add-kr-us-continuation-lanes`, recording both lanes still OFF
