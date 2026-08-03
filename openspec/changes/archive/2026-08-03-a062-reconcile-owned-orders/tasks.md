## 1. Contract and Evidence

- [x] 1.1 Reserve `STORY-TOS-a062`, create the OpenSpec proposal/design/spec deltas, capture the base commit, and pass strict validation
- [x] 1.2 Record live read-only reproduction evidence and the external-order ownership root cause without secrets or account identifiers
- [x] 1.3 Create Function Logic Maps and Branch Test Maps for every edited existing function before implementation

## 2. Owned Order Tracking

- [x] 2.1 RED: prove a non-terminal external snapshot without a confirmed attempt or lineage is excluded from `TrackedFillOrders`
- [x] 2.2 GREEN: filter tracked snapshots by positive local ownership evidence while preserving confirmed unseen orders and replacements
- [x] 2.3 Verify an external order disappearing across cycles cannot create `MissingOrders` or increment permanent-promotion failures
- [x] 2.4 Bind owned snapshots to account/market/trading-day/symbol/side, migrate additively to schema v16, and fail closed on identifier collisions or later-day reuse
- [x] 2.5 Add canonically scoped lineage evidence/resolution and prove cross-account/day/market and migration-boundary reuse cannot alter local open-order identity
- [x] 2.6 Scope startup terminal reservation recovery to the exact decision-bound owner and preserve all holds on cross-scope or ambiguous identifiers
- [x] 2.7 Wire reservation recovery before engine decisions and nonce pruning, and retain consumed-nonce evidence for every held reservation

## 3. Authoritative Reconcile Blocking

- [x] 3.1 RED: prove every active journal RECONCILE cause blocks covered automatic adoption and appears in the runtime projection
- [x] 3.2 GREEN: project all active account journal states into the tracker/runtime without letting the quantity comparer clear other-producer causes
- [x] 3.3 RED→GREEN: preserve tracker/gate blocks on durable release failure and return from `RunOnce` before adoption or price reads
- [x] 3.4 Validate include-only adoption requires a price reader and fails closed instead of panicking

## 4. Audited Recovery

- [x] 4.1 RED: cover confirmation, operator/note, engine-lock, stable-snapshot, blocking-diff refusal, and successful durable release
- [x] 4.2 GREEN: add local-only `tossctl engine reconcile-resolve` using official reads and the existing operator release contract
- [x] 4.3 Verify the command exposes no broker mutation path, HTTP route, console form, or operating-toggle change

## 5. Verification and Delivery

- [x] 5.1 Run focused tests, race tests, full tests, vet, strict OpenSpec, PM, logic-map, SDD and `make gate` checks
- [x] 5.2 Obtain independent security, test, and maintainability review; resolve all P0/P1 findings
- [x] 5.3 Merge to local main, push remote main, build and deploy containers without invoking a live order command
- [x] 5.4 Stop the engine, run the verified operator release once, restart, and confirm KR/US holdings become managed with persisted exit lines
- [x] 5.5 Confirm archive readiness: PM/memory evidence is prepared, SDD indexes are current, and the deployed services are healthy
