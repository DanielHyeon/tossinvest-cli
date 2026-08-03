## 1. Logic Mapping and RED Contracts

- [ ] 1.1 Before editing existing registry/evaluator/dispatch-adjacent functions, record Function Logic Maps and Branch Test Maps for OFF, schema arithmetic, structural timing, allocation, invalidation and stop branches
- [ ] 1.2 Add concurrent peer RED schema tests for KR absorption and US dislocation fields, units, exact `effective_at<=observed_at<=ingested_at<=evaluated_at<=fresh_until`, inclusive equality and one-tick stale/reversed boundaries, threshold/config digest, overflow and exact ppm arithmetic; neither market test waits for the other
- [ ] 1.3 Add RED structural tests for exact `sweep_at <= break_at <= reclaim_at <= evaluated_at`, bounded window, stale/missing event, cross-market/generation scope and price-decline-only refusal
- [ ] 1.4 Add RED allocation/risk property tests for immutable budget/Q, 2:4:8 weights/14, floor+remainder, q_leg planned/a066 caps, exact account-valuation-minor `max(transferred reserve, ceil(qty*stop-distance + entry fees + estimated exit fees/levies))`, a066-frozen US FX quote/direction/haircut/ceil, filled floor plus held/proposed conservative reservations, checked overflow/refusal, actual overage or unknown-risk latch, partial fill/duplicate/cancel and no upward reallocation
- [ ] 1.5 Add RED authority tests proving typed invalidation/refusal suppresses add but lane exit decisions/orders remain zero while common exit engine proceeds independently
- [ ] 1.6 Add RED same-release and failure-isolation tests for both lane versions, OFF defaults and independent KR/US progression

## 2. Concurrent KR and US Implementation

- [ ] 2.1 Implement strict KR absorption schema and pure `kr_short_absorption_reversal_v1` evaluator with checked integer arithmetic
- [ ] 2.2 Concurrently with 2.1, implement peer US dislocation schema and `us_short_dislocation_reversal_v1`; this task MUST NOT depend on KR stabilization or activation
- [ ] 2.3 Implement causal/bounded sweep-break-reclaim confirmation with stable event lineage and fail-closed scope validation
- [ ] 2.4 Implement immutable 2:4:8 ceilings, a066 cap, the frozen account-valuation-minor filled/held/proposed risk formula and same-snapshot US FX conversion, actual-fill overage latch, stop non-retreat and no reallocation
- [ ] 2.5 Register both lane ID/versions together as default OFF and return only entry/add decision or typed invalidation/refusal

## 3. Integration and Safety

- [ ] 3.1 Integrate a064 structural evidence and a065/a066 campaign/risk contracts with schema/config, event, risk-budget and leg lineage
- [ ] 3.2 Add crash/replay, partial-fill, duplicate-event, cross-market contamination and invalidation/common-exit fixtures
- [ ] 3.3 Prove dependency closure contains no broker, journal/owner writer, exit authority or operating-toggle writer and dormant registration creates no activation/LIVE change

## 4. VERIFY

- [ ] 4.1 Run targeted KR/US reversal, structural-order/window, allocation property/fuzz and common-exit tests with race detection
- [ ] 4.2 Run broader evidence/campaign/risk/strategy/exit/scheduler regressions including adversarial price-decline-only and partial-fill cases
- [ ] 4.3 Run `make sdd-sync`, `make sdd-check` and `make gate CHANGE=a068-add-kr-us-reversal-lanes`, recording both lanes remain OFF with zero live mutation
