## 1. Contract and pre-edit logic maps

- [ ] 1.1 Capture the implementation base, refresh CodeGraph evidence, and map the dormant descriptor, scheduler, horizon router, strategy dispatch, Guardian/Gateway and engine supervisor call paths without starting the engine or changing activation
- [ ] 1.2 Create pre-edit Function Logic Maps and Branch Test Maps for `runEngineRun`, `engineRuntime`, engine runtime supervision, scheduler decision/restore, `strategydispatch.Dispatch`, `strategyengine.DormantRuntimeDescriptor`, and every other existing function that implementation will change
- [ ] 1.3 Freeze a branch-to-scenario matrix covering KR/US concurrent progress, independent calendar/activation/budget/refusal, complete lineage, irreversible lease states and A→B→A drift, owner epoch/fencing, attested broker idempotency, restart ambiguity, worker failure isolation, central integrity fallback and safety-loop continuity

## 2. RED concurrent runtime and lease tests

- [ ] 2.1 Add RED coordinator tests proving KR and US workers start together, KR wait/close/refusal cannot pause eligible US work, US evidence/budget failure cannot pause KR, and no combined market authority is accepted
- [ ] 2.2 Add RED strategy tests for approved-candidate to router/lane/campaign/leg lineage, typed router/lane refusal, unsupported bindings and static absence of broker/journal/operating writers from pure router and lane dependencies
- [ ] 2.3 Add RED durable dispatch-lease state-machine tests for `ISSUED→CLAIMED→SUBMITTING→SUBMITTED|AMBIGUOUS|REFUSED`, every claim/validation terminal consumption, authority A→B→A non-revival, exact missing/changed/expired/stale/cross-market `REFUSED + RELEASED`, and consumed-lease replay refusal that releases only any retry-attempt reservation while preserving the original terminal disposition; all pre-transport cases have zero broker requests and terminal states have zero outgoing transitions
- [ ] 2.4 Add RED durable owner-epoch/fencing tests proving stale or concurrent owners cannot claim, transition or call Gateway, restart increments the epoch, and every journal/Gateway mutation compares the current fencing token
- [ ] 2.5 Add RED restart/outcome-classification tests proving CLAIMED pre-transport crash is `REFUSED + RELEASED`, exact acceptance is `SUBMITTED + TRANSFERRED`, definitive rejection or authoritative no-accept/no-fill is `REFUSED + RELEASED`, only durable transport uncertainty is `AMBIGUOUS + HELD`, no path submits twice, and broker resubmit requires a071-attested exact identity/query/dedup/idempotency
- [ ] 2.6 Add RED reservation-disposition crash tests proving outcome code/operation identity/query digest/time and lease/disposition commit atomically; pre-transport refusal and definitive post-transport no-accept release exact reservation, submitted transfers it, ambiguous freezes it, and later reconciliation never revives a terminal lease
- [ ] 2.7 Add RED failure-isolation tests proving an abnormal KR/US worker return latches only that market OFF with bounded restart while peer evaluation and all safety loops continue with reserved API budget
- [ ] 2.8 Add RED central-integrity fault tests proving journal/Gateway/owner/fence corruption closes all entry, the external supervisor fences the old owner and starts entry-incapable safety fallback within the manifest RTO of at most 60 seconds, and fallback failure preserves broker protection plus critical alerting
- [ ] 2.9 Add RED production-assembly guards proving central dispatch or its mutually fenced safety-only fallback is the only mutation owner, official Gateway is the only order path, WTS/paper/shadow/canary mutation count is zero, and no test reaches a live hostname or changes a toggle/approval

## 3. Concurrent runtime implementation

- [ ] 3.1 Implement a supervised coordinator with independent KR/US worker state, context, calendar/activation binding, evidence cursor, endpoint budget key, bounded cycle retry and typed first refusal
- [ ] 3.2 Connect approved candidates through the market/horizon router and pure lane evaluation, preserving market, candidate/evidence, router/lane/version and campaign/leg lineage before any Guardian or broker capability is available
- [ ] 3.3 Implement the irreversible durable dispatch lease state machine and outcome/disposition transaction that binds every safety generation, owner epoch/fencing token, risk reservation and attempt lineage; every pre-transport validation failure is exact `REFUSED + RELEASED`, and A→B→A cannot revive a lease
- [ ] 3.4 Implement one fenced central dispatch owner that reloads current durable authority immediately before the official Gateway call and rejects missing, changed, expired, replayed, stale-owner or cross-market leases before broker transport
- [ ] 3.5 Implement `SUBMITTING` recovery with exact broker identity/query: definitive rejection/no-accept/no-fill atomically becomes `REFUSED + RELEASED`, acceptance becomes `SUBMITTED + TRANSFERRED`, only durable uncertainty becomes `AMBIGUOUS + HELD`, and same-operation-key retry requires current attested idempotency/dedup capability
- [ ] 3.6 Implement the separate durable reservation disposition and atomic release/transfer/hold rules; exact reconciliation may change disposition but must never revive or rewrite a terminal lease
- [ ] 3.7 Extend engine supervision so cycle and abnormal worker failures latch/restart only that market while peer market and safety-class loops continue; reserve process fail-closed for central integrity faults
- [ ] 3.8 Implement the external heartbeat supervisor contract and fenced safety-only fallback with versioned RTO ≤60 seconds, no entry-lease capability, and explicit fallback-unavailable alert state
- [ ] 3.9 Publish immutable market-keyed runtime state for later read-only operation without adding activation, order, protection or operating-setting mutation capability

## 4. Production assembly integration

- [ ] 4.1 Wire evidence, scheduler, router, lane, campaign/risk, Guardian, dispatch lease and official Gateway into the production engine profile after existing recovery/interlock prerequisites
- [ ] 4.2 Integrate KR and US workers with fake clocks/calendars and bounded queues so one market's slow API or unavailable evidence cannot starve the other market or any safety-class loop
- [ ] 4.3 Prove end-to-end isolated KR and US decisions progress concurrently through the fenced central dispatch owner with deterministic lineage, non-revivable leases and no duplicate submission across restart/unknown outcome
- [ ] 4.4 Prove lane, scheduler, autostart or automation OFF yields zero new entries/legs/scale-ins while existing protection, exit, reconciliation, fill detection and emergency reduction continue
- [ ] 4.5 Preserve current production baseline defaults: runtime may be deployed but lane/autostart/entry remain OFF, activation remains absent, ProtectionReady remains evidence-gated, and no LIVE approval is synthesized
- [ ] 4.6 Fault-inject abnormal market workers and central integrity loss to prove the former isolates one market and the latter reaches fenced safety-only fallback within the frozen RTO while protection/exit/reconciliation remain continuous

## 5. VERIFY and review gates

- [ ] 5.1 Refresh post-edit AST, Function Logic Maps and Branch Test Maps for every changed existing function and pass the repository analysis checker
- [ ] 5.2 Run targeted strategy/runtime/scheduler/risk/execgw/engine tests, affected-package race tests, lease crash/restart suites, official-only mutation guards, full tests and vet, and strict OpenSpec/PM validation
- [ ] 5.3 Run `make sdd-sync`, `make sdd-check`, and `make gate CHANGE=a072-wire-multi-market-strategy-runtime`, then complete adversarial independent review before marking the high-risk change complete
- [ ] 5.4 Verify with read-only assertions that KR/US activation and LIVE approval did not change, no live broker mutation occurred, and OFF still preserves protection, exit, reconciliation and fill continuity
