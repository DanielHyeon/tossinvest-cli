## 1. Contract and pre-edit logic maps

- [ ] 1.1 Capture the implementation base, refresh CodeGraph evidence, and map the existing strategy-runtime console handler, shared descriptor, HTTP read model/SSE, runtime-only Unix transport, lane-performance projector and Compose health/deploy paths
- [ ] 1.2 Create pre-edit Function Logic Maps and Branch Test Maps for `handleStrategyRuntime`, runtime-reading validation/projection, HTTP runtime mapping/router handlers, performance attribution and every other existing function that implementation will change
- [x] 1.3 Freeze a console/API golden field matrix and branch scenarios for exact `WIRED`/`UNWIRED` readiness plus typed refusal, KR/US partial availability, dormant OFF truth, deterministic campaign lineage, partial-fill/staged-close conservation, activation drift, schema compatibility and bounded partial rollback

## 2. RED console and API contract tests

- [x] 2.1 Add RED shared-projection tests for per-market lane desired/effective, evidence digest/freshness, campaign/leg, horizon risk, scheduler/calendar, activation, exact `WIRED`/`UNWIRED` ProtectionReady plus typed refusal, reconciliation and observed-at
- [x] 2.2 Add RED console tests proving KR and US render independently, one unavailable market remains typed unknown without zero/default/cross-market fallback, dormant blockers remain visible, and no new order/gate/LIVE/activation/protection mutation control exists
- [x] 2.3 Add RED private API/OpenAPI/SSE tests proving schema parity with console, partial-market envelopes, stable unknown semantics, body/method guards and absence of new mutation routes
- [x] 2.4 Add RED runtime-only Unix transport tests proving the sidecar can read the shared projection across Compose namespaces but cannot obtain preview/apply, activation, order or protection mutation authority
- [ ] 2.5 Add RED lane-performance tests requiring market, lane/version, campaign/leg and full decision-to-close identifiers, rejecting symbol/time or same-ticker cross-market attribution, deduplicating fill deltas/corrections, and preserving `link_missing`/`not_measured`
- [ ] 2.6 Add RED quantity and accounting conservation tests for partial entry fills, staged closes, residual open quantity, authoritative cost-basis allocation, entry/exit fees, taxes, FX source/rate/as-of, rounding and gross-to-net PnL; prove missing fee/FX is never coerced to zero

## 3. Read-only operational projection implementation

- [x] 3.1 Implement one server-owned market-keyed runtime projection and descriptor registry with typed per-market unknown/refusal values and no operating-setting writer or broker mutator dependency
- [x] 3.2 Extend the authenticated runtime-only Unix endpoint and reader for the projection with strict schema/unknown-field validation, bounded reads and independent market error envelopes
- [x] 3.3 Adapt the console strategy-runtime view to the shared projection, preserving authenticated GET-only behavior, OFF/unobserved blocker honesty, responsive rendering and zero free-input/mutation surface
- [x] 3.4 Adapt private REST/SSE/OpenAPI models to the same projection and golden field registry without duplicating defaults, labels, refusal mapping or effective-state calculations
- [ ] 3.5 Extend performance attribution and queries with market/lane/version/campaign/leg composite lineage, fill/close-leg deltas, authoritative cost policy/version and quantity/fee/FX/PnL conservation while retaining the isolated derived store, bounded pruning and read-only authority

## 4. Integration and deployment guards

- [x] 4.1 Add an integrated engine-runtime/Unix/console/API fixture proving simultaneous KR/US snapshots, partial-market failure isolation, reconnect/full-snapshot convergence and exact console/API parity
- [ ] 4.2 Add Compose preimage tests requiring exact current/target image digests, rendered Compose/config/activation/protection digests, environment key set, volume/mount identity, schema versions/compatibility ranges and baseline health; mutable tags or incomplete preimages must fail before replacement
- [x] 4.3 Add dormant health checks for console/API schema, authenticated Unix projection connectivity and KR/US OFF/not-configured truth without starting entry runtime or contacting a broker mutation endpoint
- [ ] 4.4 Add deployment regression tests proving one-service-at-a-time replacement honors frozen order and ≤5 minute per-service timeout, cannot change autostart/automation/lane/LIVE/protection state, and cannot create an order/audit mutation
- [ ] 4.5 Add partial-failure rollback tests proving only the replaced subset rolls back in reverse order to exact preimage digests, untouched services/config/volumes/journal/protection stay unchanged, and incompatible rollback keeps the new service with entry OFF and safety continuity

## 5. VERIFY and release gates

- [ ] 5.1 Refresh post-edit AST, Function Logic Maps and Branch Test Maps for every changed existing function and pass the repository analysis checker
- [ ] 5.2 Run targeted console/httpapi/runtime-transport/performance tests, affected-package race tests, OpenAPI and responsive/static route guards, full tests and vet, and strict OpenSpec/PM validation
- [ ] 5.3 Run `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=a073-operate-multi-market-strategy-lanes`, then complete independent implementation review before release
- [ ] 5.4 Verify `docker compose config`, exact image digests, schema read/write compatibility and the frozen OFF-state/config/volume preimage before replacing any running service

## 6. Dormant deploy and post-deploy checks

- [ ] 6.1 Record the complete immutable deployment preimage with exact service image digests, rendered Compose/config/activation/lane/autostart/automation/LIVE/protection digests, environment keys, volumes/mounts, schema compatibility and baseline health using read-only commands; abort if any field or OFF/unapproved baseline is not provable
- [ ] 6.2 Replace Compose services one at a time in frozen order within the ≤5 minute per-service bound, preserving persistent volumes and without starting entry runtime, changing a toggle/approval or issuing a live order
- [ ] 6.3 Run post-deploy console/API/Unix health and KR/US dormant projection checks, compare pre/post state digests, and prove no order, protection mutation, activation or operating-setting audit event was created
- [ ] 6.4 If health or state preservation fails, stop further replacement and roll back only the replaced subset in reverse order to exact preimage digests; leave config/approval/journal/volumes and broker-resident protection untouched, and if schema compatibility forbids rollback keep the new service with entry OFF/safety continuity and typed recovery
