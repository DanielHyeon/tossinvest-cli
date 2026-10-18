# a157 proposal-freeze adversarial Eng review (Opus, separate context)

Target: `openspec/changes/a157-first-leg-binding-before-sizing` (proposal · design · tasks · spec delta · AST bundles).
Code read at HEAD `dc6625687` (no source change since base `7b1748481` — `git diff --stat 7b1748481 HEAD -- internal/` is empty, and the AST bundle `source_sha256` values match the current files).
Real repo untouched. All measurements ran in a scratch copy:
`/tmp/claude-1000/-mnt-D-Axipient-workspace-TossOS/1711ae2d-4cc3-48b0-97b1-c0fa1bb7c4da/scratchpad/review-a157/copy-3975299`
(copied go.mod, go.sum, internal/, cmd/, tools/, fixtures/, schemas/; no `.git`; `/tmp/.git` confirmed absent).

## Verdict: **APPROVE-WITH-FIXES**

No P0. The core safety claim holds. The **set of accepted first-leg requests is unchanged**, and no branch of D2 produces more issuances or orders.
Three P1s must be fixed before freeze:
- a spec SHALL is false (measured);
- design option D5-1 as written can panic;
- the test and mutation plan cannot kill a wrong partial reorder. The existing suite pins nothing about the order (measured).

---

## Measurement: A1 applied in the scratch copy

I implemented A1 exactly as design D1 describes it:
- add the nil guard;
- move B2–B8 ahead of `PrecheckQFinalEntry` with the inputs switched to `request.Entry.Admission.{TransactionID,Owner}` and `request.Entry.QCandidate`;
- leave B9 after sizing.

Diff: `first_leg_a1.go.txt` in the review dir. Results:

| Run | Result |
|---|---|
| `go vet ./internal/execgw/` | clean |
| `go test -count=1 ./internal/execgw/` (untagged) | ok |
| `go test -count=1 -tags tossos_testseams ./internal/execgw/ ./internal/app/engine/` (clean rerun, file sha `2e41c970…` asserted before and after) | **ok / ok** |
| `go test -count=1 -trimpath ./internal/app/engine/` (untagged) | only `TestA111ObserverUsesClockLeaseHelpersForTheUseLease` and `TestA111FallbackSequenceRecoveryIsLazyAndPriceEvidenceUsesTheGateDuration` fail. Both are `-trimpath` artefacts: they open `exitloop.go` through a module-path name. They are unrelated to A1. |

**Consequence:** no existing test pins the old order. M1 ("move the block back") can only be killed by the new rows in 0.4/1.1. See P1-3.

Probe test (`zz_a157_probe_testseam_test.go.txt` in the review dir) ran against HEAD and against A1 using the existing `qFinalCampaignFirstLegIssuanceFixture`:

| Row | HEAD code/field | A1 code/field |
|---|---|---|
| control (US, valid) | accept | accept |
| **single** `Entry.QCandidate = 0` | `ZERO_QUANTITY / q_candidate` | `RISK_CALCULATION_INVALID / first_leg_lineage` |
| **single** `Entry.Symbol = "aapl"` (lineage `AAPL`) | `UNKNOWN_SYMBOL / symbol` | `RISK_CALCULATION_INVALID / first_leg_lineage` |
| single `Entry.Market = " US"` | `first_leg_lineage` | `first_leg_lineage` |
| single cap 0 (KR weekly lane, `RiskBudget = 1 KRW`) | `EXISTING_GUARDIAN_CAP` | `EXISTING_GUARDIAN_CAP` |
| cap 0 + `Weekly = nil` | `EXISTING_GUARDIAN_CAP` | `weekly_reservation` |
| cap 0 + TransactionID `"tx\|bad"` | `EXISTING_GUARDIAN_CAP` | `journal.ErrInvalidRequest` |
| cap 0 + `ProspectiveGeneration = "pg"` | `EXISTING_GUARDIAN_CAP` | `OWNER_CONFLICT / prospective_generation` |
| cap 0 + `ExpectedPositionGeneration = 3` | `EXISTING_GUARDIAN_CAP` | `OWNER_CONFLICT / position_generation` |
| cap 0 + `AttemptID = " a"` | `EXISTING_GUARDIAN_CAP` | `first_leg_strategy` |
| nil Guardian | `RISK_CALCULATION_INVALID / guardian` | same |

---

## P0

None. Evidence is in "Verified correct" at the end.

## P1

### P1-1. "A request with only one defect is reported the same" is false. Measured.
- **Where the claim is made:** spec ADDED scenario 「결함이 하나뿐인 요청은 그대로다」 (spec.md:9-11) and design D3 bullet 1. Both say a request with one binding defect **or** one entry refusal keeps its code · field · message · cause.
- **Why it fails:** one defect can trip a moved check *and* a `PrecheckQFinalEntry` check, because they read the same request fields. Two cases measured (probe table above):
  - `Entry.QCandidate = 0` → HEAD `ZERO_QUANTITY/q_candidate` (riskguardian_qfinal.go:106), A1 `first_leg_lineage` (riskguardian_first_leg.go:71).
  - non-canonical `Entry.Symbol` → HEAD `UNKNOWN_SYMBOL/symbol` (qfinal.go:103), A1 `first_leg_lineage`.
- **Other overlaps by inspection (same mechanism):**
  - `lineage.AccountRef ≠ g.accountRef` with `owner.Key.AccountID = lineage.AccountRef` → HEAD `OWNER_CONFLICT/owner` (qfinal.go:120), A1 `first_leg_lineage`.
  - A mismatched or unparsable Entry/Stop/Target price → HEAD `EXISTING_GUARDIAN_CAP` (qfinal.go:147/151), A1 `first_leg_lineage`. When the mismatch drives the cap to 0, this is the intended flip.
- **Impact:** none on orders or engine classification. All of these were and stay `AUTHORITY_MISMATCH` in `admit`, and `validateStrategyFirstLegAuthority` (strategy_first_leg_admission.go:222-227) already rejects them in production. The problem is that the SHALL as written cannot pass. Also, D2's claim that it "enumerates everything" is only true at the `admit`-classification level, not at the code/field level.
- **Fix:**
  - Define "defect" as "failing check". Reword the scenario to: "a request that passes all seven binding checks keeps the entry refusal's code · field · message · cause; a request that passes `PrecheckQFinalEntry` keeps the binding refusal's".
  - Add a D2 row (f), "one input trips both judgements", listing the four overlaps above with the measured codes.
  - Add the two measured rows to 0.4 so the new reporting is pinned rather than unspecified.

### P1-2. D5 option 1 as described can panic: typed-nil interface.
- **The claim:** design D5-1 says "`EntryGate: c.Entry` — nil 이면 지금과 같고".
- **Why it is false:**
  - `c.Entry` is `*execgw.EntryGate` (engine.go:181). Assigning a nil pointer to `StrategyEntrySupervisorOptions.EntryGate StrategyEntryBlocker` (strategy_entry_supervisor.go:126) gives a **non-nil interface**.
  - `blockEntryOnCentralIntegrity` checks `s.entry == nil` (:1046). That check is then false, and it calls `s.entry.Block(...)`.
  - `(*EntryGate).Block` does `g.mu.Lock()` with no nil-receiver guard (execgw/retry.go:543-545), so it is a nil dereference.
- **Today:** the assembly supervisor is never `Run`, so this is latent.
- **Risk:** the first change that does `Run` it would turn "process-wide fail-closed" into a goroutine panic. If unrecovered, the engine stops, and with protection UNWIRED that means no stop-loss (safety invariant 4).
- **Fix:** if option 1 is chosen, write `if c.Entry != nil { opts.EntryGate = c.Entry }`. Add a test with `c.Entry == nil` asserting the supervisor's `entry` is the nil interface, and one with a real gate asserting the Block path latches entries instead of returning `ErrStrategyCentralIntegrity`. Correct the D5 text.

### P1-3. The planned tests and mutants cannot catch a partial or duplicated reorder.
- **What I measured:** the whole existing suite passes with A1 (above). Only the new rows can kill M1.
- **Gap 1: partial move.** M1 rows are "cap 0 + B2 · B5 · B8" only. A wrong implementation that moves B2/B5/B8 but leaves **B3, B4, B6 or B7** after `PrecheckQFinalEntry` passes every planned row. The ADDED SHALL names all seven checks.
  - All of them are constructible at execgw level. The probe shows B2, B3, B6, B7 and B8 each flipping with `RiskBudget = 1 KRW` on the KR weekly fixture.
  - B4 is reachable by leaving `Result.ExecutionTerms` zero-valued: `MajorDecimal` fails on an empty `priceMinor` (strategyflow/types.go:192-195), and B2/B3 still pass.
  - **Fix:** add a "cap 0 + Bn" row for each of the 7 checks. For B5, add one row per changed input source (q_candidate disjunct and owner LaneID/CampaignID disjunct).
- **Gap 2: duplicated checks.** No structural pin covers spec 「같은 판정을 사이징 앞과 뒤 두 자리에 두어서는 안 된다 (SHALL NOT)」. A mutant that keeps B2–B8 after the call *and* adds them before it is behaviourally equivalent and survives every row.
  - **Fix:** add an AST test on `PrecheckQFinalCampaignFirstLeg`. It must assert:
    - exactly one `g.PrecheckQFinalEntry` call;
    - every binding-refusal `return` precedes it;
    - each field literal (`prospective_generation`, `first_leg_price_unit`, `first_leg_lineage`, `position_generation`, `first_leg_strategy`, `weekly_reservation`) and the single `journal.QFinalPolicyVersion` call occur exactly once in the function.
  - Kill it with a "duplicate block" mutant (M14).
- **Engine-level note:** only B2 and B8 can flip through `admit`. `validateStrategyFirstLegAuthority` blocks B3–B7 first, and B5's `AccountRef` disjunct is caught before the cap by qfinal.go:120. Engine rows (a146 fixture with weekly injection / bad TransactionID) should therefore be B2 and B8 only. Say this in D3 so nobody tries to build unreachable engine rows.

## P2

### P2-1. The "source sha unchanged" check for `IssuePrecheckedQFinalCampaignFirstLeg` is impossible as specified.
- **Where:** D3 bullet 3 and task 1.3 require the sha of `PrecheckQFinalEntry` · `IssueQFinalEntry` · `IssuePrecheckedQFinalCampaignFirstLeg` to stay unchanged, "AST 번들 대조".
- **Why it cannot pass:** the bundle's `source_sha256` is the **file** sha.
  - Measured: bundle `a31ccdfd…` == `sha256sum internal/execgw/riskguardian_first_leg.go`.
  - Measured: bundle `6fcace26…` == `riskguardian_qfinal.go`.
  - `IssuePrecheckedQFinalCampaignFirstLeg` lives in the edited file (riskguardian_first_leg.go:117), so its file sha must change.
- **Fix:** hash the gofmt-normalised `FuncDecl` source span, for example with a small `go/ast` extractor or by adding a body sha to `tools/logic-map` output. Leave the qfinal.go pair as a file-sha check.

### P2-2. The A2 scenario "same dedup rule" is false for the assembly supervisor's lifetime, and the planned "Run" test cannot reach its target.
- **Dedup scope:** dedup state lives on each supervisor's `strategyMarketRuntime` (`dailyLossNoticeDay` in a137_daily_loss_notice.go; `entrySkipNoticeDay` / `entrySkipNoticed` in a146_entry_skip.go).
- **Lifetime:** a new assembly supervisor is built on **every refresh wave** (`refreshPairedStrategyEntryProductionAssembly` → `collectStrategyRefreshWave` → `NewPairedStrategyEntryProductionAssembly`, strategy_entry_supervisor.go:611-631).
- **Consequence:** if one were ever `Run`, "once per account day (and symbol)" would reset every wave. If it ran alongside the production supervisor, the same process would emit two lines per key. Both break the main SHALL 「한 프로세스 안에서 … 한 번」.
- **Today:** `Run` is never called, so A2 is harmless. No double notice exists. nil `c.Log` → the factories return a nil func, which is safe.
- **Reachability of the D4 test:** the assembly workers' `Cycle` is `c.productionStrategyCycle` → `runProductionStrategyMarketCycle` → a full authority refresh plus dispatch. Driving a skip or unmeasurable record into *that* supervisor from a seam test is impractical. M12 will end up killed by whatever the test really asserts.
- **Fix:**
  - Reword the scenario to "the assembly supervisor is constructed with the same notice functions (projection-only; a change that Runs it must lift dedup state to `Context` and wire EntryGate)".
  - Test: build the assembly with a minimal `Context{Log: …}` and assert both notice fields are non-nil. Then call the package-internal recorders (`recordGuardianCapZeroSkip` and the a137 recorder) on that supervisor and assert the `obs` events reach `c.Log`. Use the production supervisor as the control.

### P2-3. Spec contradictions created by the delta.
- **Inside MODIFIED:** the requirement keeps 「같은 주기의 범위 거절 보고, 멈추는 결함, 일일 손실 결과의 처리는 이 요구 이전과 같아야 한다 (SHALL)」. A157 changes those very outcomes for co-occurring requests (D2 b/c/d).
  - Fix: append the qualifier 「(결속 결함이 함께 있는 요청은 「전략 첫 다리 결속 판정은 Guardian 사이징보다 먼저 한다」가 정한다)」.
- **Against the a137 requirement:** 「일일 손실 한도의 입력은 원장에서 측정된다」 (spec.md:558ff) says 「측정이 한도 도달을 판정하면 생산 진입 경로는 운영 모드를 ENTRY_BLOCKED … 로 강화 · 영속해야 한다 (SHALL)」, and its scenario 「자본 0 인 계좌 → 강화」 has no binding-defect exception. The new scenario says 「강화가 일어나지 않으며」.
  - Fix: either MODIFY the a137 requirement with one clause ("a request refused by a first-leg binding check never reaches the chain; escalation happens on the next request that does") or state that cross-reference explicitly in the ADDED requirement.

### P2-4. D2-c: I agree with **accept**, but the trade-off shown to the human is incomplete.
- **Verified:** design claims 2 and 3 are true. The cycle stops (deliverEachStrategyHandoff :91-95). The manual path escalates on its own chain refusal (riskguardian.go:452 → `escalateFor` :663).
- **Missing argument for accept:** today, binding defects of the B3–B7 class are already rejected by `validateStrategyFirstLegAuthority` at admit:100, *before* `MeasureDailyLoss`/precheck (:104-105). So "a binding defect hides daily-loss escalation" is existing behaviour, and a157 only extends it to B2/B8.
- **Missing cost:**
  - In today's production shape (refreshOnly worker, deterministic scope order), a **persistent** B2/B8 defect on the first scope means this market's strategy path never escalates that day. There is no ENTRY_BLOCKED mode row and no mode announcement (`g.announcer`). The operator only sees the swallowed-fault counter.
  - New exposure is still zero: the defective scope is refused and later scopes are not evaluated.
  - The (d) branch also loses the `engine.daily_loss_unmeasurable` warning for that cycle.
- **Fix:** add both points to D2-c so the human decides with them in view.

## P3

1. **Wrong line.** The assembly supervisor's `NewStrategyEntrySupervisor` call is at **:394**, not :392. AST `engine--context.newpairedstrategyentryproductionassembly` calls list `(394, NewStrategyEntrySupervisor)`; :392 is inside the `range` worker loop. Fix all three places: proposal Why, design F8, and D4.
2. **Overstated clock claim.** Proposal "What Does NOT Change" says 「clock 은 B9 앞의 issuedAt 에서만 읽음」. B1 (`PrecheckQFinalEntry`) reads `g.clk.Now()` at 118, 127, 132 and 165 (AST calls). The argument still holds because B2–B8 read no clock. Restate it as "B2–B8 read no clock and have no side effects".
3. **Hand-read claims cited as AST.** F3 and F4 cite AST "assignments", but the bundles carry only `kind`/line, not the assigned target. The field-level statements ("only `QCandidate`/`EvaluatedAt`/`AccountID`/`Policy`/`QExistingGuardian` are written"; "only :66 sets `QCandidate`") are hand reads. I checked them against the source and they are **true**. Label them as source reads (qfinal.go:113-175, admission.go:66-124), or extend logic-map to emit assignment targets.
4. **SHALL NOT too broad.** ADDED 「같은 판정을 사이징 앞과 뒤 두 자리에 두어서는 안 된다」 literally collides with:
   - the new nil guard, which duplicates `PrecheckQFinalEntry` B1;
   - the issue-time `QFinalPolicyVersion` recomputation (riskguardian_first_leg.go:135).
   Scope it to "the seven binding judgements, within the precheck".
5. **Unclear scope of "모든 생산 조립".** MODIFIED 「감독자를 만드는 모든 생산 조립(권한 갱신 감독자와 조립 감독자)」 is ambiguous. The package-level value-only `NewPairedStrategyEntryProductionAssembly(snapshot)` (strategy_entry_supervisor.go:648) and `NewDormantStrategyEntrySupervisor` also build supervisors without notices. They are dormant (no Cycle), and cmd/tossctl/engine_strategy_entry_dormant_test.go:33 does `Run` one. Name the two `(*Context)` constructors explicitly.
6. **M13 is not a mutant.** "Compile failure expected" proves nothing about the tests; record it as a build note, not as CAUGHT.

## Verified correct (coverage for the Manager)

- **F1:** branch ids and lines (B1 51 … B9 105) and the clock at 95 match the AST.
- **F2:** B2–B8 call only pure helpers.
  - `QFinalPolicyVersion` (journal/risk_bucket_issuance.go:48-57) is total over strings.
  - `guardianWeeklyFirstLegBinding` (:216-235) is nil-safe and returns a clone.
  - `MajorDecimal` and `strings.*` are pure.
- **F3 / F4:** true. `OwnerKey`, `Owner` and `TransactionID` are value types (riskbucket/owner.go, journal/risk_bucket.go:64-72). The clone at :113 copies them. `CalculateAdmission` writes `QCandidate` only at :66.
- **`policyVersion` / `accountRef`:** immutable after construction (riskguardian.go:308 is the only assignment), so reading them before the call is equivalent.
- **Canonicalisation (question 1):** B5 compared the *raw* `request.Entry.Market/Symbol` before and still does. `PrecheckQFinalEntry` canonicalises only its own copy (qfinal.go:174 assigns its by-value parameter), so no ACCEPT changes. Non-canonical inputs that were accepted before (`"kr"`, through EqualFold plus canonicalisation) are still accepted. Others change only *which* refusal is reported (P1-1).
- **No new panics:** `PrecheckQFinalEntry` now sees a subset of its former inputs. B2–B8's new inputs are the Entry-derived strings and ints only, and none can panic there; B6's `uint64(...)+1` is guarded. The nil guard covers `g.policyVersion` and `g.accountRef`.
- **Accepted set unchanged:** shown by the argument above and by the whole existing execgw + engine suite (incl. `tossos_testseams`) passing with A1. That suite includes tests asserting `QCandidate`/`QFinal`/receipt values on the success path.
- **D2 at the admit level:** complete. No branch becomes less conservative in orders. X is always `RISK_CALCULATION_INVALID` / `OWNER_CONFLICT` / `ErrInvalidRequest`, with causes that cannot match `DailyLossUnmeasurable`, `DailyLossLimitReached`, the cap-0 sentinel or `BucketCapExhausted`. So X is always `AUTHORITY_MISMATCH` and stops the cycle.
- **F6:** callers verified by grep — `PrecheckQFinalEntry` ← first leg :50 and `IssueQFinalEntry` :323; `PrecheckQFinalCampaignFirstLeg` ← `admit` :105.
- **F7 / F8:** branch ids verified. The other F8 lines are correct: 433-434, :541, :418-421.
- **Spec MODIFIED:** a faithful copy of main spec.md:719-757. The only edits are: inserted a157 sentence, 「그 뒤의」 → 「q_final 위에서 도는」, 「정책 버전 · 주간 예약 결속 ·」 removed, the trailing notice SHALL, and two new scenarios.
