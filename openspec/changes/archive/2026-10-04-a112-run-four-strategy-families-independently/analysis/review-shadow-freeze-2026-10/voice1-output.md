# SHADOW freeze voice1 — SubagentHandback 원문 (agent a8ef03efb4591dbc6, 저장 2026-10-04)

Voice 1 (MUST NOT / exposure) — a112 7.3.1 SHADOW freeze review

Safety: I read nothing under `~/.codex`. I did not edit the repository. All experiments ran in a scratchpad copy of c7219640 (`/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/4b178b0a-b918-4419-8142-f266a03d8e82/scratchpad/v1copy`): two probe tests and one sketch file, nothing else. The brief's sha256 matches (96baede4…d762).

**Verdict: REJECT — revise and re-freeze.** There is no P0: built as designed, nothing produces an order, raises exposure or turns activation ON. But P1-1 shows the §0 premise ("pure evaluation over proposals that already exist, no new I/O") is false in the code. Fixing it moves the edit into the order-path function `collectMarket`/`coordinateMarketProposals`, which changes the risk profile the freeze approved. P1-2 to P1-5 are pins that cannot support the MUST NOT claims as written.

## P1

**P1-1. The SHADOW input set is empty or tautological, so §0 is wrong about the code.**
- Evidence: `evaluate`'s inputs are `strategyLaneInputs(c.AccountRef, fresh.proposals.forMarket(market))` (strategy_entry_supervisor.go:543-546).
- `authority.entries` is set only by `arbitration.entries()`, which holds only the winners the coordinator selected (strategy_proposal_authority.go:435-457). Every closure path returns entries that are empty.
- So the proposal of a lane that is OFF in a declared market is gated as DORMANT before arbitration and never reaches the lane runtime.
- Repro (copy, `zz_v1_shadow_input_probe_test.go`, tag `tossos_testseams`, using the existing `collectUnderLoad` fixture):
  - declared, only continuation ON → `gated=[DORMANT] laneInputs=[000660/CONTINUATION 005930/CONTINUATION]`. The REVERSAL proposal is absent, so the shadowed REVERSAL lane can only ever report NO_INPUT.
  - undeclared (shadow-only deployment) → `laneInputs=[000660/CONTINUATION 005930/REVERSAL]`. These are exactly the legacy winners that are being dispatched anyway; the losing continuation proposal on 005930 is absent.
- Result: as designed, the SHALL "pure evaluation and counterfactual projection" is met vacuously — the passes-on-an-empty-sample pattern. A meaningful SHADOW needs the pre-gate batch (`batch.LanesFor`) carried out of `collectMarket`. That is a High-risk edit next to `entries`/`dispatchHandoffs`.
- Proposal:
  - Rewrite §0.
  - Carry pre-gate inputs in a type that `dispatchHandoffs` and `coordinateMarketProposals` cannot take.
  - Add a structural pin that `dispatchHandoffs` reads only `entries`.
  - Add `collectMarket` and `coordinateMarketProposals` to the FLM list with mutation.

**P1-2. The §3 census does not block the ON path that is actually open.**
- `FamilyActivation` is protected only by unexported fields of `strategyrouter`. The shadow loader and `FamilyShadow` are planned for the same package and are told to share the activation loader's functions.
- No test today pins where `FamilyActivation` can be minted. The comment at production_family_activation.go:331-333 is a claim, not a pin.
- Repro (copy): I added `FamilyShadow` to `strategyrouter` with `ShadowFor(...) bool` and `Promotion() FamilyActivation`, which mints ON.
  - The design's census ("FamilyShadow methods returning DesiredState = 0") passes (methods=2, DesiredState-returning=0).
  - `worker.Effective(shadow.Promotion())` = ON, and `Run(..., Input{})` returns REFUSED rather than DORMANT, so the activation gate was passed.
  - Both the existing untagged and tagged strategyrouter and strategyworker suites stay green with this sketch present (`rtk proxy go test`: ok/ok).
- Proposal, either:
  - (a) Put `FamilyShadow` and its loader in a separate package that cannot mint `FamilyActivation`, sharing only byte-level helpers that return no authority type; or
  - (b) Add a type-identity census over `strategyrouter` production files: functions whose results or out-params carry `FamilyActivation` = exactly {LoadProductionFamilyActivation, the tagged seam}, and composite literals or conversions of `FamilyActivation` only there. Add a source-digest freeze, since `mint_census_test.go` itself retracts completeness (generics, `any`, twin-struct conversion).
- Both variants also need: `FamilyShadow` has no field or method carrying `FamilyActivation`.

**P1-3. "ShadowCycle has no Envelope" is not a capability boundary, and the zero-count behaviour pin is wrong for the target deployment.**
- The engine already holds the `Input` (Scope, SnapshotDigest, Proposal), and `strategycoordinator.Envelope` is a public literal it builds itself (strategy_market_coordinator.go:79-81). One line turns WOULD_EMIT into a Submit.
- §3 ③ asserts "Submit 0 · dispatchHandoffs 0 · broker 0". In the shadow-only deployment that F1 is designed for (activation undeclared), the gate is not installed (strategy_family_activation.go:132-133), so the legacy path Submits and dispatches normally. The test can only pass if the activation is declared, which leaves the real configuration unmeasured.
- Proposal: replace it with a differential test. With and without the shadow pin, in both undeclared and declared markets, the Submit, handoff, broker, lease, latch and recovery traces must be identical.
- Add a reader allowlist: shadow values and fields may be read only by `projection()`. Today `runtime.observed` is read only by `observations()` and `projection()` (grep), but nothing pins that.
- `FamilyShadow` must not be a parameter of `familyGateFor`, `admit`, `coordinateMarketProposals` or `dispatch*`. `TestTheRollbackPathOnlyReads` already covers `admit`; keep `lane.Shadow` off that allowlist.

**P1-4. A fault in SHADOW can latch a lane durably or halt the market cycle (a ledger write, and it survives restart).**
- Inside the lane step, a panic or deadline goes through `settle → failLocked`, which latches the lane (bounded.go:229-244). `persistMarketLatches` then writes `strategy_lane_latches` (strategy_lane_runtime.go:252). The latch comes back after restart (`TestALatchedLaneComesBackLatchedAfterTheProcessRestarts`).
- Outside the step, a panic is re-raised at the join (strategy_lane_runtime.go:244-247). Any `evaluate` error skips that cycle's dispatch (strategy_entry_supervisor.go:543-548) and latches the market (`latchMarket`, :1093-1124).
- If the shadow load sits in `collectMarket`, its `recover` turns a panic into INTERNAL_FAILURE and closes the legacy path (strategy_proposal_authority.go:270-276).
- So a read-only diagnostic gains entry-closing power, and the §3 claim "lease/latch/recovery rows 0" is measured only on the happy path. The premise of `TestTheProductionStepNeverLatchesSoTheLedgerStaysEmpty` (DORMANT, no error) also needs re-measuring once Shadow runs in the step.
- Proposal: read the manifest once per wave, outside the lanes, and pass it as a value (the activation precedent). Shadow computation must never return an error and must recover its own panic into a shadow outcome rather than a lane failure. Add fault-injection pins: shadow panic → 0 lane latches, 0 ledger rows, legacy dispatch unchanged, market not latched.

**P1-5. Sharing with the activation loader makes the activation gate an unplanned edit target.**
- Manager ② says to share the binding comparison, lifetime and revocation code. Today those live inside `validateProductionFamilyActivation` and `LoadProductionFamilyActivation`, return activation sentinels and take the activation body type (production_family_activation.go:429-624). Sharing them means refactoring the code that decides exposure.
- The §7 FLM list (`collectMarket`, `evaluate`, projection, around `familyGateFor`) does not name these functions, and §7.4 does not plan re-running mutation on them.
- Shared helpers that return `ErrProductionFamilyActivation*` would give shadow errors activation identities; `familyGateFor` branches on `errors.Is(err, ErrProductionFamilyActivationUndeclared)`.
- Proposal:
  - Add the three activation functions to Pre-Edit FLM/BTM.
  - Re-run the activation loader's mutation ledger after the extraction.
  - Pin `errors.Is(shadowErr, <any activation sentinel>) == false` and the reverse.
  - Add a cross-feed test: shadow bytes rejected by the activation loader and activation bytes rejected by the shadow loader.

**P1-6. Two separate judgements of "OFF lane".**
- §3 says SHADOW applies only to OFF lanes ("ON wins"); §9 adds the projection rule runtime=SHADOW ⇒ desired=OFF ∧ effective=OFF.
- desired ON / effective OFF is a legitimate activation state (production_family_activation.go:609-612). Canonical-bytes equality does not prove the generator wrote the file — json.Marshal of a hand-built body passes (:523-525). So the state is reachable even though decision 63 says only the generator makes manifests.
- If `worker.Shadow` reuses `Run`'s `Effective != ON` test, those lanes become SHADOW with desired=ON. The widened `validateLane` then rejects the whole snapshot on every read path (fail-closed, but the projection disappears entirely).
- Proposal: one exported predicate (desired OFF ∧ effective OFF) used by both `worker.Shadow` and `validateLane`. Pin it in both places, and mutate each side.

## P2

- **Restart pin.** The R2 pin as designed reads the projection before any wave runs after restart, so it holds structurally (the observed map is empty). It should run a wave after restart for three cases (no pin, pin but invalid file, file but no pin) and assert UNOBSERVED with wave > 0.
- **Spec scenario narrowed.** The amended :91-92 ("restart without the shadow manifest pin") no longer covers "pin present, file unusable". The requirement text (:89) still covers it, so add it to the scenario or the pin.
- **Latched lanes.** `worker.Shadow` sits on `FamilyWorker` and never sees the latch. A LATCHED lane can therefore report WOULD_EMIT. Run the shadow through the lane with latch first, or add a LATCHED outcome.
- **shadowOutcome with no pin.** With pin 0 the additive field must be null or absent, or "behaviour change 0" is false on the API. Pin shadowOutcome≠nil ⇔ runtime=SHADOW in `validateLane`, and forbid SHADOW when trigger is nil or the lane is unobserved.
- **Vocabulary census.** It counts only typed const declarations in router and projection. A `RuntimeState("SHADOW")` conversion, or a const placed in strategyworker, slips past. `validateLane` is the real gate, so say so.
- **ShadowCycle field census.** Use an exact field list. A "carries no Envelope" check is defeated by an `any` or func-typed field, as `mint_census_test.go` itself records.
- **Coordinates and wording.** The brief cites design :289/:291; at c7219640 these are :291/:293 (the amendment added two lines). The R2 test message still says "signed shadow manifest … golden amendment"; it will be renamed per §5 and §8 ④.

## Ungraded (out of scope per the brief)

spec.md:38 "SHADOW counterfactual 외 dispatch handoff는 0건" reads as if a SHADOW counterfactual were a kind of dispatch handoff. That conflicts with :89 MUST NOT and with Manager ③ (lane-only). Reconcile the wording; it must not become a justification for routing counterfactuals through `dispatchHandoffs`.

## Attacks tried that held

- **strategyworker minting ON.** Impossible: `FamilyActivation`'s fields are unexported, so the worker package can only pass the zero value (DORMANT). The sketch had to live in strategyrouter.
- **Observation memory leaking into the next wave's activation decision.** The gate re-reads activation through the loader every wave, and `runtime.observed` has no reader outside `projection()`/`observations()`. This holds but is unpinned (see P1-3).
- **First projection after restart, before any wave.** The store starts as `DormantSnapshot`, the new runtime has no observations, and `productionWorker` fixes runtime at UNOBSERVED. Nothing persists: `strategyprojection.Store` is memory only.
- **Recovery generation replaced by a shadow generation.** The argument expression is frozen by `TestTheRecoveryGenerationComesFromTheVerifiedActivationAndNothingElse` (a112_lane_latch_durability_test.go:306-360).
- **Shadow bytes accepted as an activation manifest.** Rejected by `DisallowUnknownFields`, canonical equality and the schema/domain check (production_family_activation.go:515, :523-525, :545-546), provided the shadow body is a distinct struct. Pin both directions (P1-5).
- **8.6 / A100.** Building the capability is not deploying it. 8.6 stays BLOCKED, and with production shadow pin 0 nothing deploys. The exception is the code that runs regardless of the pin (P1-4, P2 shadowOutcome).

## Repo unchanged

Start and end, identical:
```
 M openspec/changes/a112-run-four-strategy-families-independently/tasks.md
?? .reticle-setup-crash.log
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/review-shadow-freeze-2026-10/
?? openspec/changes/a112-run-four-strategy-families-independently/analysis/shadow-2026-10/
?? w4.log
HEAD c7219640d08408a650941c62cf0df6efff77c036
```
