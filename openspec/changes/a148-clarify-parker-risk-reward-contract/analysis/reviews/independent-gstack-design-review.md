# a148–a156 independent engineering design review

Review date: 2026-10-18. Repository: `/mnt/D/Axipient/workspace/TossOS`.
Scope: read-only review of nine proposed changes, current tasks, PM stories, WORKFLOW/SDD rules and selected existing module/test surfaces. No code, proposal, task, PM or operational changes. No tests or aggregate gates executed. Applied plan-eng-review architecture, code-quality, testability and performance checks; host setup, telemetry, interactive approval workflow and additional outside voice omitted by assignment. This report is independent review output, not a claim that a separate cross-model pass ran.

## Verdict

**DESIGN FREEZE PASS — current a148–a156 design set.** Re-read current a150/a151/a153/a155 design and normative spec deltas after Manager revisions. All substantive findings from this review are resolved at design level. No outstanding design decision blocks freezing this proposal batch.

This is not implementation READY, passing implementation tests, profitable-strategy validation, operational approval or archive approval. Each change retains its predecessor, fresh evidence, implementation, independent review and gate requirements. a156 remains incomplete until separately approved actual observation/live evidence and reconciliation tasks are satisfied.

Final resolution evidence:
- a150 distinguishes pre-submit rejection from post-fill recovery, retains admitted protection and recorded fills, blocks additional exposure and reports unresolved protection. Scenarios now sit under the correct requirement and explicitly separate timing.
- a151 fixes next eligible 1-minute open entry, adverse slippage/tick rounding, gap-through-stop execution, common absolute 2.1% target, no trailing/partial exits in the first experiment, and paired censored treatment without invented fills.
- a153 fixes market-separated equal-session net-R aggregation, 95% paired percentile CI, seed 153/2000 resamples, common mark-to-market MDD and exactly one primary candidate chosen before holdout access. Best-holdout selection cannot promote.
- a155 fixes 256-record/64KiB bounds, nonblocking enqueue, drop-newest counters, no disk/retry on trading or exit threads, fsync-before-checkpoint and stable-ID recovery without evidence truncation.
- Current lifecycle tasks distinguish preparation from actual gate/archive execution; a156 tasks 4.1/4.2 remain blocking.

Dependencies remain acyclic; frozen v1/r1/a128/a147 preservation and separate operational authority remain intact.

## Findings and disposition

All four original findings below are **RESOLVED** in the final design/spec reread. Their quoted earlier text and recommendations are retained as review history and test guidance, not outstanding defects. The implementation tests described below remain to be written/run by the implementation owner.

### 1. [P1] a150 needs separate pre-dispatch rejection and post-fill protection failure contracts (confidence 8/10)

Evidence: `openspec/changes/a150-bind-strategy-stop-sizing-and-exit/design.md:18`: “최초 stop을 exit state가 전달받는 경계의 누락·불일치는 신규 진입을 fail-closed한다. 이미 보유한 포지션의 기존 보호 exit는 유지한다.” Spec `specs/risk-management/spec.md:7–8` likewise says an exit-initialization mismatch refuses new entry.

That is sufficient before dispatch, but does not define what happens when a new position already partially filled before exit initialization/recovery discovers missing or inconsistent protection. “Keep existing exit” does not establish a valid exit for that newly filled position. This is a specification ambiguity, not a demonstrated current-code defect.

Recommendation: bind and validate immutable approved stop/terms before order submission; distinguish (a) pre-dispatch mismatch: no order/reservation mutation, and (b) post-dispatch/fill/restart mismatch: deny further exposure, preserve/recover the last durable authorized protection through existing safety/reconciliation paths, retain explicit unresolved state and alert. Never erase the fill, fabricate terminal order state, silently initialize a wider stop or wait for a research/observation path. Name the existing recovery owner before editing.

Tests: injected mismatch before dispatch → broker count zero; partial fill then journal/exit initialization failure → no additional entry and existing safety path remains available; restart with incomplete snapshot → no synthetic valid snapshot; duplicate fill → no duplicate protection/entry. Use integration fixtures across terms/reservation/journal/exit, not only pure sizing tests.

### 2. [P1] a153 promotion criterion lacks an exact statistical decision contract (confidence 8/10)

Evidence: `openspec/changes/a153-evaluate-cost-aware-strategy-variants/design.md:17`: “세션 단위 paired aggregation 및 session bootstrap CI(고정 seed, 2000회)” and line 18: “paired net expectancy 개선 CI lower>0”.

Seed, confidence level, interval construction, session statistic/weighting and multiple-variant holdout selection are not fixed. Two implementations can obey the text and produce different promotion eligibility from identical data. Twenty sessions is correctly described as a floor rather than sufficient evidence; it does not remove this ambiguity.

Recommendation: preregister one primary metric and one candidate selected solely on train/validation; freeze a literal seed, confidence level (e.g. two-sided 95%), percentile interval rule/quantile interpolation, common paired session resampling indices, session weighting and missing/no-trade-session treatment. Keep additional holdout contrasts exploratory, or explicitly fix simultaneous/multiplicity handling. Define equity series and comparison basis for MDD nonworsening. Missing required definitions must refuse a confirmatory run, rather than get defaults after outcomes.

Tests: independent golden small session table with expected aggregation; exact replay of interval endpoints; candidate selection/parameter changes after holdout → consumed/exploratory; 19 versus 20 sessions; zero/no-trade and missing-cost sessions; differing trade counts cannot silently change the primary denominator. This requires a short manifest contract, not a statistics framework.

### 3. [P2] a151 fixes stop arms but leaves the executable fill/exit model to interpretation (confidence 8/10)

Evidence: `openspec/changes/a151-compare-stop-policies-offline/design.md:15` fixes “동일 후보·진입 시점·비용 모델·청산 정책”; line 17 says “다음 시점 executable price 모델 없이 신호봉 종가에 즉시 체결했다고 가정하지 않는다.”

The prohibition is good but not a selected model. Exact entry timing, gap-through-stop execution, tick rounding and whether “same exit” fixes absolute targets or an R-multiple are required for reproducible stop-only comparisons. With different stops an R-multiple can change target prices; that may be intended, but must be disclosed as the estimand rather than silently chosen by implementer.

Recommendation: freeze these fields before results. Prefer one simple documented OHLC replay model with conservative gap fills and explicit unsupported cases; refuse manifests with missing execution semantics. Specify absolute-target-fixed versus same-R-rule explicitly. Add a mini numeric example where all three arms share entry/cost assumptions and show expected stops, quantities and exits.

Tests: next-bar open gap beyond stop; signal before bar close; target and stop same bar; invalid structural stop above entry; tick rounding reaches entry; risk cap rounds quantity to zero. Existing tasks list gaps and ambiguous bars but do not supply expected execution outcomes yet.

### 4. [P2] a155 resource isolation remains qualitative (confidence 8/10)

Evidence: `openspec/changes/a155-observe-versioned-strategy-decisions/design.md:16`: “append-only, bounded queue/drop counters, duplicate ID 처리와 재시작 checkpoint를 명세한다. 기록 실패/부하는 기존 protective exit를 지연시키지 않는다.”

No actual queue bound, maximum record size, overflow policy, shutdown/drain policy or append/checkpoint crash-order rule is specified. Therefore disk/backpressure tests can pass under incompatible implementations. Strictly “no delay” cannot be proved using a wall-clock comparison alone.

Recommendation: use one existing bounded channel/consumer if available; fix queue capacity and byte bounds in versioned configuration, reject blocking writes on any trading/protection path, choose a documented drop policy and count gaps. Specify append-before-checkpoint plus deterministic duplicate identity and crash recovery behavior, or simpler replay-only semantics if no checkpoint is needed yet. No new scheduler or service framework.

Tests: stalled writer and full queue → producer returns without waiting; oversized record refused; crash after append before checkpoint → replay is deterministic/deduplicated; failed append never advances committed checkpoint; stop/exit test proceeds despite permanently blocked observation sink. Capacity and test timeout values must be selected before this integration is READY.

## Architecture and dependencies

```text
a148 -> a149 -> a150 -> a151 ----+-> a153 -> a154 -> a155 -> a156
           \-> a152 ------------+          ^        ^
                    a150 ------------------+--------+
                    a149 ---------------------------+
```

Graph is acyclic. Explicit design dependencies: a149←148; a150←148,149; a151←150; a152←149; a153←151,152; a154←150,153; a155←149,150,153,154; a156←155. a152 could technically be independent read-only work once a149 exists, but the user's numbered sequence can remain sequential. No speculative parallel implementation needed.

a150 spans many existing high-risk modules. Keep one end-to-end contract trace first and classify each strategy as supported/unsupported; do not introduce a universal stop-policy service or claim all strategy paths connected. Scope smell is manageable by explicit capability table and one writer, not automatically by creating more changes. a155 correctly refuses the order-capable coordinator rather than borrowing a paper/shadow switch. a154 correctly reuses settings lifecycle and separates initial stops from common exits.

## What already exists / reuse

- `internal/strategyflow/types.go:230`: ExecutionTerms structure; a150 explicitly reuses it.
- `internal/risk/account_base.go:244`: AccountBaseStrategyEntryQuantity; `internal/riskbucket/admission.go:65`: CalculateAdmission; reuse caps/risk admission boundaries rather than parallel calculations.
- `internal/optimization/store.go:523,672`: existing Preview/Apply entry points. Existing test names include `TestApplyIsVersionBoundOneShotCASAndAppendOnly`, `TestEvidenceBackedApplyRevalidatesCurrentEvidenceBeforeMutation`, and actor/tamper/replay cases. a154 should extend these contracts and revalidate current evidence/approval server-side, not rely on disabled UI alone. Tests were located, not executed or proved sufficient.
- Existing engine exit tests cover isolation/concurrency surfaces; map new a150 behavior onto them after fresh FLM/BTM. This review makes no unsupported claim about internal branches.
- a147 already distinguishes offline implementation from separately authorized collection. Its immutable evidence boundary is consistent with the new series; do not reinterpret its past observations as independent confirmatory data.

## Test-plan map and failure handling

Tests below are planned requirements, not passing evidence.

```text
Immutable source inputs
  -> a148 RR diagnostics        denominator invalid -> typed invalid/unknown [planned]
  -> a149 session calculator    gap/late/future/warm-up -> not_ready/reject [planned]
  -> a150 terms + sizing        pre-dispatch mismatch -> deny [planned]
       -> fill/journal/exit     post-fill mismatch -> recovery contract [contract resolved; integration test required]
  -> a151 paired replay         same-bar ambiguity -> stop-first + bounds [planned]
                                gap execution price -> exact model [contract resolved; test required]
  -> a152 VP feasibility        no true trades/bins -> unsupported [planned]
  -> a153 holdout report        small/reused sample -> inconclusive [planned]
                                CI/primary comparison -> deterministic gate [contract resolved; test required]
  -> a154 UI preview/CAS        stale/unknown/unapproved -> refused/read-only [planned, integration]
  -> a155 observation consumer  disk/backpressure -> visible drops [planned]
                                crash checkpoint ordering -> fixed semantics [contract resolved; test required]
  -> a156 live gate             no approval -> inactive/open change [planned]
                                uncertain cancel -> no new exposure [planned, integration]
```

Add byte/hash equality checks over frozen v1 fixtures and a128/r1/a147 input paths at each relevant change, plus negative broker/reservation/operating-setting side-effect assertions. Protection timing needs call-path isolation tests, not profit metrics. UI requires keyboard/disabled/unknown states and server validation; synthetic fixtures are sufficient for UI development but not live approval.

## Performance

No new throughput architecture warranted. a149 can process a bounded regular-session prefix with checked arithmetic; document maximal input length and avoid unbounded rational growth/repeated full-prefix work in later streaming use. a153's 2000 bootstrap repetitions are bounded but input/session/variant counts should be manifest-bounded; no runtime path should wait for research output. a155's explicit bounds are now specified; implementation must verify them under a stalled sink. No benchmark was run; no operational latency guarantee is asserted.

## Completion and archive semantics (current files rechecked)

The initial lifecycle ordering concern has been corrected during review. Current a156 tasks 3.2/3.4 are gate/lifecycle preparation, not false execution receipts; its Post-gate lifecycle is after the operational task section. Tasks 4.1/4.2 explicitly block live completion/archive until separate approval, real evidence and reconciliation. Do not report the earlier task ordering as an outstanding defect.

- a152 may validly complete with unsupported data: feasibility is its deliverable; spec explicitly permits this. That does not authorize a VP arm.
- a153 may deliver a reproducible inconclusive report without producing an approved policy; make this acceptance meaning explicit when implementation is scoped. Inconclusive is neither a successful strategy nor fabricated evidence.
- a154 may implement read-only/disabled choices with synthetic fixtures; no qualifying candidate means no selectable live option.
- a155 scope is offline replay/adapter verification; archive would not assert actual observation happened. a156 still requires actual observation evidence separately.
- a156 remains incomplete while live tasks are unchecked; preparation, dry runs and code gates cannot close it.
- No current new change is asserted complete. All future completion remains subject to fresh base/context, scoped tests, independent review, aggregate gate receipts and formal spec/PM/archive lifecycle.

## NOT in scope

- New data collection, API/model runs, subscriptions or schedules: separate authorization required.
- Settings changes, restarts, orders, new live legs or widening protection: not authorized by these designs.
- Production VP engine: a152 only qualifies data; separate approved implementation required.
- New strategy/configuration framework: existing contracts and finite descriptors suffice.
- r1 prompt/selection changes or retroactive evidence repair: explicitly prohibited.
- Extra TODO writes: no repository edits authorized for this review. Existing TODOS' live sell/cancel validation limitations should be considered before a156; no reason to perform them now.

## GSTACK REVIEW REPORT

| Runs | Status | Findings |
|---|---|---|
| Independent substantive plan review + final reread | PASS for design freeze | All four original findings resolved at contract level |
| Architecture/dependencies | Reviewed | No cycle; pre-submit/post-fill boundary explicit |
| Code quality/KISS | Reviewed | Existing terms/risk/settings reused; no framework justified |
| Testability | Reviewed | Exact replay, statistical and failure contracts available; implementation tests not run |
| Performance | Reviewed | Observation queue/record/overflow/durability bounds now explicit |
| Separate outside voice / cross-model | Not run | Assignment prohibits additional agents; no absent reviewer claimed |
| Runtime validation / live evidence | Not run | Outside approved review scope |

VERDICT: DESIGN FREEZE PASS. Implementation and empirical validation remain outstanding. a156 remains NOT READY for live completion/archive without separately approved operational evidence.

NO UNRESOLVED DECISIONS
