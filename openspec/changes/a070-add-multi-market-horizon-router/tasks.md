## 1. Logic Mapping and RED Contracts

- [x] 1.1 No existing function was edited; record the new-package Function Logic and Branch Test Map in `analysis/function-logic/new-package.md`
- [x] 1.2 Add RED owner tests for exact `(account,market,symbol,position_generation)` key, horizon exclusion, all-active-horizon lookup, short-vs-weekly race, multiple-owner corruption, generation rollover and stale snapshot
- [x] 1.3 Add concurrent KR/US RED lifecycle tests proving either market OFF/closed/stale/CAS-conflicted/failed does not gate the peer and no combined activation authority exists
- [x] 1.4 Add RED shared physical endpoint/reset-generation tests proving market/horizon are anti-replay subscopes, shared exhaustion defers all subscopes, concurrent last-slot acquire cannot multiply quota, replay is rejected and safety reserve survives
- [x] 1.5 Add RED per-market record tests for revision/lock/activation CAS, independent rollback, transaction crash/restart and peer revision preservation
- [x] 1.6 Add RED legacy migration tests: disabled to both OFF, verified KR/US only to that market with peer OFF, corrupt/combined/unverified to both OFF, idempotent crash retry and no synthesized activation
- [x] 1.7 Add RED dormant tests for current scheduler DISABLED/OFF, no market selected, runtime UNOBSERVED and zero routing/order/toggle mutation

## 2. Router and Scheduler Implementation

- [x] 2.1 Implement pure routing over exact ownership key with all-horizon owner lookup, owner preservation and typed ambiguity/corruption refusal
- [x] 2.2 Implement independent KR/US durable desired records, monotonic revisions, market locks, activation bindings and CAS/rollback
- [x] 2.3 Implement fail-closed idempotent legacy migration without broadening market authority
- [x] 2.4 Extend capability anti-replay preimage with market/horizon while retaining one shared physical endpoint/reset-generation quota/commitment/absolute-cap authority
- [x] 2.5 Bind short/weekly cadence to scoped capabilities without creating per-scope capacity or touching safety reservations

## 3. Integration and Safety

- [ ] 3.1 Integrate a067–a069 eligible lanes and a066 owner snapshots so each generation reaches at most one owner across all horizons
- [x] 3.2 Add independent KR/US, short/weekly, legacy migration, owner race, CAS crash and shared quota exhaustion matrices
- [x] 3.3 Prove router/scheduler cannot create campaign, owner, broker, journal, activation or toggle mutations and OFF yields zero entry routing
- [ ] 3.4 Confirm exit/fill/reconciliation/protection/emergency loops continue under shared quota exhaustion, migration refusal or one-market failure

## 4. VERIFY

- [x] 4.1 Run targeted router/owner/scheduler/migration/budget tests with race detection and cross-horizon plus shared-quota property tests
- [ ] 4.2 Run broader candidate/strategy/campaign/risk/exit/runtime regressions including concurrent CAS, crash/replay and partial-market failure
- [ ] 4.3 Run `make sdd-sync`, `make sdd-check` and `make gate CHANGE=a070-add-multi-market-horizon-router`, recording no market selection, lane enablement or live order
