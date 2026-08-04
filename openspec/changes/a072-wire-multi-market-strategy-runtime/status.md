# Status — a072-wire-multi-market-strategy-runtime

- Overall: VERIFYING the paired KR/US release candidate
- Isolated core: GREEN (`internal/strategyruntime`)
- Release/default: paired KR/US, both entry `OFF` / runtime `UNOBSERVED`
- Coordinator: independent market calendar/activation/evidence/budget/refusal/latch and bounded restart state
- Lineage: complete immutable candidate→evidence→router→lane/version→campaign/leg→risk/reservation→Guardian→attempt chain
- Lease: sealed, revisioned, irreversible `ISSUED→CLAIMED→SUBMITTING→terminal`
- Validation: exact authority generations/digests, protection serial, account/market/symbol, owner epoch/fence and expiry
- Outcomes: atomic `REFUSED+RELEASED`, `SUBMITTED+TRANSFERRED`, or `AMBIGUOUS+HELD`
- Recovery: CLAIMED crash is pre-transport refusal; SUBMITTING requires non-contradictory authoritative lookup/response proof observed no earlier than lease issue
- Resubmit: `AMBIGUOUS+HELD` only, same operation ID, exact current authority/owner and complete attested broker capability with bounded attempts
- Ordering: nonterminal out-of-order submit/classification consumes the lease as `REFUSED+RELEASED`, zero broker requests
- Worker hardening: immutable first typed refusal, exact latch/OFF invariant and saturating restart attempt/deadline
- Identity: common canonical UTF-8 boundary, maximum 256 bytes, no whitespace/control; post-mint reseal cannot bypass it
- Fallback: newer fenced safety-only owner, frozen RTO ≤60 seconds, no entry/lease issuance
- Authority: package-private constructors; no broker/live transport, journal/gateway writer, toggle or approval

## Final paired production integration

- KR and US are implemented in the same release wave. The production engine assembles both markets through
  scheduler, candidate, router, all six concrete lane evaluators, official FX, one account-base Guardian,
  five-bucket risk authority, atomic first-leg admission, fenced dispatch ownership and the official Gateway.
- Final market-local schedule and transport authority are revalidated inside the entry-gate critical section
  immediately before send. Owner/revision/lease fencing prevents expired, replaced or ABA-replayed authority
  from crossing the transport boundary for either market.
- Missing activation/evidence remains a typed, market-local OFF refusal. The release creates no activation,
  operating-setting or LIVE approval and does not weaken protection, exit, reconciliation or fill loops.
- The final independent red-team review is CLEAN with no CRITICAL/P1/high finding. Its last regression proves
  that a gate wait cannot use an expired/replaced KR or US dispatch owner.
- Pure strategy composition: GREEN (`internal/strategyflow`), paired KR/US continuation/reversal/weekly
  bindings, all `OFF/UNOBSERVED`, sealed candidate→router→lane→campaign/leg/risk lineage
- Durable journal checkpoint: additive v25 preserves v24 and records the central owner fence,
  per-market KR/US authority shape, q_final-bound lease shape and cold-restart discovery. Every
  authority-bearing mutation is deliberately `ErrStrategyDispatchDormant`; no production mint exists.
- Additive v26 first-leg checkpoint: one journal-owned transaction now atomically records paired KR/US
  q_final, aggregate plus five monetary holds, immutable strategy lineage, campaign/claim, prospective owner
  and leg 1. Its immutable companion includes the fixed router ID/release and future leases must repeat the
  exact client operation identity. It still mints no dispatch lease or broker capability.
- Dispatch-owner checkpoint: transport-free task 3.4a is GREEN. Current owner, authority revision/digest,
  first-leg binding, attempt/client/manifest/router, campaign/claim/leg/owner and exact final reservation
  disposition are revalidated in one CAS transaction. The strategy-only Gateway now atomically commits the
  core attempt and lease into `DISPATCH_STARTED + SUBMITTING` immediately before its callback. After restart,
  a newer current owner can close only a durable `CLAIMED` row with no transport-start marker; optional
  prepared core, lease, aggregate and five buckets become `NOT_DISPATCHED + REFUSED + RELEASED` in one
  rollback-safe transaction. Recovery that could retry or resubmit remains deliberately closed.
- Engine supervisor checkpoint: paired KR/US evaluation children start behind one barrier with independent
  bounded queues. Market faults irreversibly disable only that in-memory child and emit immutable fault
  evidence. Production now assembles the single outer `strategy-entry` loop from an explicit read-only
  automation/protection snapshot. The command boundary contains no activation/threshold/risk/Gateway/journal/
  broker capability, so both markets remain OFF with nil cycles even when ordinary automation facts are true.
- Paired production schedule checkpoint: `engine run` now delegates one context-owned same-wave load that
  freezes one observation instant, reads exact KR/US desired files and official calendars concurrently, and
  verifies separate digest-pinned Ed25519 activation manifests. A bad or absent market artifact refuses only
  that market; the command receives scalar schedule observations and no opaque activation authority. This
  checkpoint deliberately does not promote a worker because risk/lane/first-leg/Gateway cycle completeness is
  still missing.
- Paired production candidate checkpoint: scheduler-ready KR and US now consume exact owner-only threshold,
  opaque evidence and human approval files in independent goroutines at the scheduler's one frozen instant.
  Each activation record is externally SHA-256 pinned and binds the canonical set, evidence, market/session,
  version, actor and approval time. The shared discovery database is opened through separate SQLite `mode=ro`
  and `query_only` handles, and the audited `strategycandidate` sanitizer lets only measured-and-clear values
  cross as immutable `strategy.ApprovedSnapshot`. A broken KR artifact leaves US ready and the mirror case is
  tested; schedule OFF performs zero authority/store reads. Public output is scalar counts/digests only, and
  both workers remain dormant because risk/lane/first-leg/Gateway cycle authority is incomplete.
- Paired production FX checkpoint: candidate-ready KR and US now enter the existing
  `officialfx.ProductionAuthorityService` concurrently at the same frozen observation. KR independently
  re-verifies the official account and mints same-currency identity; US independently verifies the pinned,
  owner-only signed monotonic haircut policy and reads the official quote-to-account-base rate. Either
  refusal leaves its peer result intact, candidate OFF performs zero market reads, and opaque evidence stays
  private while only pair/digest readiness is projected. Workers remain dormant until five-bucket state,
  lane inputs, first-leg admission and Gateway cycle authority are complete.
- Paired lane-ordering decision: production lane evaluation is now specified as a two-stage contract.
  All six KR/US bindings first produce a sealed cap-free `q_candidate` proposal from exact router,
  evidence, plan, stop and execution-term authority; the account-base Guardian and five-bucket
  transaction then derive `q_final`. A fabricated unlimited cap or feeding `q_final` backward into the
  proposal is forbidden. RED/GREEN and production wiring remain pending and must land for KR and US in
  the same wave.
- Paired q_candidate implementation checkpoint: continuation 8:4:2, reversal 2:4:8 and weekly-value
  now expose KR/US cap-free proposal evaluators over their shared fail-closed validation paths. Proposal
  quantity is exact immutable leg remaining; weekly proposals still require the durable market-week
  reservation. `strategyflow.Propose` seals quantity, complete lineage, execution terms and zero-authority
  counters. Mutation breaks the seal, legacy cap-admitted evaluation is not a proposal, and the production
  five-bucket loader now accepts only `ValidProposal`. `FinalizeProposalQuantity` can only reduce a valid
  proposal to q_final and preserves exact lineage plus all non-quantity terms. Paired lane, strategyflow,
  riskbucket and engine tagged tests plus normal tests/vet are GREEN.
- Paired production proposal checkpoint: separate owner-only, digest-pinned, Ed25519-signed KR/US
  proposal manifests are consumed at one frozen observation. The loader replays exact immutable
  `evidence.db` snapshots through `mode=ro + query_only`, rechecks the separately sealed router and
  official FX authority, and dispatches only the exact six-market matrix (continuation, reversal and
  weekly-value for both KR and US). Weekly proposals additionally require the schema-v27 durable
  market/week reservation and bind that exact reservation into the later atomic first-leg request.
  A corrupt or missing market artifact refuses only that market; only sealed cap-free `q_candidate`
  results remain inside the engine package. Multiple symbol proposals are never collapsed into an
  arbitrary risk input. Paired and all-six matrix tests are GREEN; workers remain OFF until the
  authority-complete admission/lease/Gateway cycle lands.
- Market recovery checkpoint: cycle error and panic latch only the failed market, publish an absolute bounded
  restart deadline and preserve peer plus five safety loops. The latch never auto-restores entry authority;
  central integrity still exits through process fail-closed. Independent review is CLEAN.
- Sealed bridge audit: exact entry/stop/target contracts, cap-free q_candidate proposal, one account-base
  Guardian, paired production five-bucket snapshot loader, KR identity evidence and US official
  quote-to-base FX mint now exist in the same wave. The production engine still lacks the signed lane-input/
  router-owner adapter, prospective owner/exposure collector and authority-complete
  proposal→q_final→first-leg assembler, so US and KR production entry remain closed.
- Read-only snapshot authority: one market-scoped contract now seals exactly five ordered bucket/policy
  provenances and their matching immutable journal references for both KR and US. It rejects stale,
  missing, duplicate, tampered, currency-mismatched and cross-market input. Its source is package-private

## Paired production five-bucket authority checkpoint

- KR and US now consume separate owner-only `0400`, externally digest-pinned and Ed25519-signed
  `risk-bucket-policy-{KR,US}.json` files in one component wave. The policy owns the exact account and
  currency scope, horizon/market limits, lane-to-server-risk mapping, symbol-to-sector limits and
  versioned worst-case fee model; TossOS ships no writer or signer.
- `riskbucket` revalidates the sealed strategy result and official FX at the engine's frozen instant,
  treats the exact sealed BUY limit as the official mutation-contract worst price, and opens the existing
  schema-v26 journal with `mode=ro + query_only` to sum bounded exact prior `filled + HELD` usage.
- A malformed, stale, latched, cross-market, wrong-owner/mode/signature/digest policy or journal makes
  only that market unavailable. Paired normal/tagged tests prove the exact valid peer remains ready.
- The command boundary receives only market/symbol/sector/count/digest observations. Opaque bundles stay
  inside the engine package. Because no production lane-input result exists yet, the normal context path
  reports `LANE_NOT_READY`, performs no risk-policy read and leaves both workers dormant.
  and its production constructor remains unavailable, so this adds no runtime authority or mutation.
- Six-lane execution authority: KR and US continuation, reversal and weekly-value results now carry opaque
  exact entry/stop/target provenance, currency/scale/unit and policy lineage. Weekly RR and reversal/saved-stop
  authority are sealed; a public scalar cannot retreat an existing stop. Final adversarial re-review is CLEAN.
- FX/q_final authority: official evidence remains opaque through Guardian issuance and is revalidated at the
  Guardian clock. Client configuration is immutable after construction and origin validation plus the FX GET
  are atomic. The paired production authority service re-verifies KR account identity and accepts US haircut
  policy only from a pinned, signed, monotonic manifest; its engine config/trust-pin loader remains absent, so
  neither market gains entry.
- Currency decision: KR and US will share one account-base-currency Guardian with request-scoped frozen
  official FX; per-market quote Guardians are forbidden because they split the account-wide cap.
- Paired currency delivery: KR identity FX and US official quote-to-base FX are one implementation wave;
  neither market's operational stability may gate the peer's design, RED tests, adapter or Gateway work.
- Composite settlement checkpoint: paired KR/US strategy dispatch derives its terminal result only from the
  actual dispatch callback and durable attempt state. One cancellation-detached transaction now owns core
  outcome, terminal lease/disposition, five risk-order mappings, campaign watermark/leg/campaign state,
  strategy execution lineage and ACK-window missed-fill repair. Public caller finalization was removed;
  ordinary `DispatchVerified` is unchanged and no retry/resubmit authority was added.
- Independent-review P1 remediation: strategy dispatch now requires a non-nil official existence verifier
  before any durable transition or transport. Terminal ACK-window zero/partial fills release aggregate and
  five-bucket remainders in the same confirmed transaction. A prepared pre-transport refusal closes core,
  lease, aggregate and five buckets in one rollback-safe transaction. Production fill wiring now includes
  Campaign with Project/Exit/Costs while both market workers remain dormant. Independent re-review is CLEAN
  with P0/P1/P2 = 0/0/0.
- Pre-transport cleanup checkpoint: every exact CLAIMED refusal, including early protection, policy, FX,
  reservation and final-fence failures, normalizes the aggregate plus five buckets to RELEASED and terminally
  consumes the lease. Safe prior partial release is preserved; missing, filled, partial-held, mapped or
  substituted authority rolls back unchanged.
- Remaining: package-owned production account/exposure loader,
  authority-complete q_candidate→q_final→atomic-first-leg production adapter and lease mint, full paired a071 protection assembly,
  SUBMITTING exact-outcome recovery, full safety-loop fault injection, repository gates and final review

## Paired production router implementation checkpoint

- The route authority contract is frozen before code: separate pinned/signed KR and US manifests each
  contain a bounded canonical set of symbol scopes, every scope contains exactly the market's
  continuation, reversal and weekly descriptors, and the manifest repeats the private scheduler
  activation/calendar binding.
- Owner state is reconstructed only from schema-v26 journal history. Empty history permits first-leg
  candidate routing; an existing owner requires an exact active campaign, actual numeric generation,
  unblocked entry and no risk latch. Prospective tokens are never converted into position generations.
- Package-private owner/market seals remain private, one-market failure preserves its peer, and the
  loader owns no writer, signer, activation, Guardian, Gateway or broker capability. Every market batch
  reads its manifest once and reconstructs all approved symbols in one read-only SQLite transaction;
  KR and US batches start concurrently and a missing signed symbol scope is counted as a local refusal.
- Paired RED/GREEN, market-failure isolation, multi-symbol partial refusal, tagged integration, vet,
  Function Logic analysis, strict OpenSpec validation and whitespace checks pass in the same checkpoint.

No market was activated, no lease reached a broker, and no live order or operating setting changed.
