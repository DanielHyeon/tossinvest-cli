## 1. Contract and pre-edit logic maps

- [ ] 1.1 Capture the implementation base, refresh CodeGraph evidence, and map the current `ProfileProtection=UNWIRED` assembly, attestation loader, protection controller, official gateway and engine interlock call paths without changing runtime state
- [ ] 1.2 Create pre-edit Function Logic Maps and Branch Test Maps for `protection.AssessReadiness`, `Controller.Register`, `Controller.Replace`, `Controller.Reconcile`, `Controller.Recover`, `execgw.Gateway.checkProtection`, engine assembly, and every other existing function that implementation will change
- [ ] 1.3 Freeze a branch-to-scenario matrix for KR/US independent readiness, signed scope validation, expiry/build drift, restart recovery, duplicate events, replace coverage, oversell prevention and OFF-state safety-loop continuity
- [x] 1.4 Record the new isolated-core Function Logic and Branch Test Map in `analysis/function-logic/isolated-core.md`; no existing runtime function is edited in this wave

## 2. RED attestation and protection lifecycle tests

- [x] 2.1 Add RED strict-attestation tests for pinned trust root, key ID, signature algorithm allowlist, revocation, bounded rotation overlap, monotonic serial anti-rollback, maximum lifetime, trusted-time unavailability/rollback, schema version, account/profile, market, order type, session, quantity range, trigger source, replace semantics, tool/build/evidence digests, issued/expiry times, unknown fields and file ownership/permission failures
- [x] 2.2 Add RED market-isolation tests proving KR-only evidence cannot authorize US, US-only evidence remains valid when KR expires, readiness alone never enables a lane/autostart/automation/LIVE approval, and missing evidence keeps both markets `UNWIRED`
- [x] 2.3 Add RED capability tests requiring exact broker client-key echo, lookup fields, identity uniqueness scope, pending/terminal and cancel-result query semantics, and idempotency/dedup behavior; prove any absent capability yields exactly `UNWIRED` plus a typed refusal
- [ ] 2.4 Add RED controller tests for submit/cancel unknown outcomes, register-response crash, duplicate fill, stable generation/revision operation identity, exact broker-ID recovery, orphan discovery, atomic/continuous replacement refusal, non-retreat trigger and sell-claim oversubscription; prove unattested idempotency never permits resubmission and unowned orphans are never guessed/canceled
- [ ] 2.5 Add RED engine/Gateway tests proving attestation drift at dispatch blocks only exposure raising, existing broker protection and reduce-only exit continue, and reconciliation/recovery never infer identity from symbol/time
- [x] 2.6 Add RED transport guards for the isolated readiness core proving it has no live-host transport dependency and cannot change a toggle, lane, activation or approval

## 3. Protection readiness implementation

- [x] 3.1 Implement the strict signed protection-attestation schema and canonical verifier with pinned trust roots, key/algorithm/revocation/rotation policy, durable monotonic serial and trusted-time floor, maximum lifetime, typed per-market refusal, current build/evidence binding and fail-closed file validation
- [ ] 3.2 Replace the production-global readiness constant at the decision boundary with immutable KR/US readiness snapshots derived from exact attestation scope plus actual supervisor wiring, while preserving the shipped `UNWIRED` default
- [ ] 3.3 Implement exact broker identity/query/dedup capability parsing and durable generation/revision operation identity plus desired/observed/unknown broker state needed for registration, safer replacement, cancellation, restart recovery and reconciliation
- [ ] 3.4 Enforce attested replace semantics, continuous coverage, broker-reserved/local sell-claim bounds and entry-latch closure without weakening current ACTIVE protection
- [ ] 3.5 Wire the official protection gateway and supervisor into production engine assembly without exposing a second mutation path or changing lane, autostart, automation gate or LIVE approval settings
- [ ] 3.6 Implement submit/cancel unknown and orphan reconciliation so resubmission occurs only under attested broker idempotency, otherwise remains no-resubmit/entry-latched until exact reconciliation or human resolution

## 4. Isolated integration and failure recovery

- [ ] 4.1 Build an isolated KR/US official-broker integration fixture that exercises signed readiness, partial fill, registration, quantity convergence, safer replacement, cancellation and exact broker-state reconciliation with live hostname mutation structurally blocked
- [ ] 4.2 Prove process loss after broker acceptance but before local commit recovers the same protection once when exact query/dedup is attested, and otherwise enters no-resubmit reconciliation without duplicating a conditional order while new exposure stays blocked
- [ ] 4.3 Prove KR attestation/recovery failure does not lower valid US readiness and vice versa, while protection, exit, reconciliation and fill handling continue in both markets
- [ ] 4.4 Run restart/idempotence/crash-point tests and the official-only/WTS-isolation matrix; confirm no fixture flips a toggle, creates approval or sends a live order

## 5. VERIFY and review gates

- [ ] 5.1 Refresh post-edit AST, Function Logic Maps and Branch Test Maps for every changed existing function and pass the repository analysis checker
- [ ] 5.2 Run targeted protection/attestation/execgw/engine tests, race tests for affected packages, journal crash/restart suites, full tests and vet, and strict OpenSpec/PM validation
- [ ] 5.3 Run `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=a071-wire-kr-us-protection-readiness`, then complete adversarial independent review before marking the high-risk change complete
- [ ] 5.4 Verify the built default remains lane/autostart/automation/LIVE OFF or unapproved, missing attestation remains `UNWIRED`, and protection/exit/reconciliation/fill paths remain available without any live broker mutation
- [x] 5.5 Run isolated-core unit, race, vet, fuzz, coverage, static dependency and strict OpenSpec validation; preserve the production integration gates above as pending
