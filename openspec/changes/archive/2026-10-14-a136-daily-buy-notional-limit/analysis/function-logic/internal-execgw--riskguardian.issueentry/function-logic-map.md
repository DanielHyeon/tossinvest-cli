# Function Logic Map: `IssueEntry`

- Source: `internal/execgw/riskguardian.go` (418-575, 편집 뒤)
- Qualified: `RiskGuardian.IssueEntry`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 88cf7c183018aae8…, 편집 전 분기 수 13) — 현재 branches 14 · returns 13
- Risk scan: `risk-pattern-report.md`

**역할.** 일반 진입 발급 문(체인 1회 → 원자 발급).

**a136 편집.** collectIssue 클로저의 ReserveRequest 에 `DailyBuyLimit: g.dailyBuyLimit()` 한 칸. (재추출 2026-10-14: a137 a1cfe0c2 가 같은 함수 앞부분에 일일 손실 측정(`g.MeasureDailyLoss` · `g.dailyLossFloor` · `evaluateChainWithDailyLoss`)을 넣어 호출 목록이 바뀜 — a136 편집 지점(collectIssue 의 `dailyBuyLimit`)은 무변경.)

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `req` | BUY | 호출자 | 오류 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if req.Collect == nil {` (if 419) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B2 | `if err != nil {` (if 423) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B3 | `if intent.Side != risk.SideBuy {` (if 426) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B4 | `if verdict := evaluateChainWithDailyLoss(in, measureErr); !verdict.Allowed {` (if 444) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B5 | `if !verdict.Allowed {` (if 453) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B6 | `if err != nil {` (if 488) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B7 | `if err != nil {` (if 492) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B8 | `if err != nil {` (if 497) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B9 | `if req.strategyPlan == nil {` (if 529) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B10 | `} else {` (else 531) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B11 | `if collectErr != nil {` (if 535) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B12 | `if strategyErr == nil {` (if 540) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B13 | `if err != nil {` (if 546) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |
| B14 | `if refused, ok := refusal.(*IssueRefusal); ok && refused.Stage == StageIssuance {` (if 553) | — | 분기 본문 참조 | `TestA136IssueEntryRefusesTheBuyThatReachesTheDailyCeiling` · `TestA136AManualBuyOutsideTheEngineCountsTowardTheDay` · `TestA136IssueEntryRefusesWhenForeignManualBuysCannotBeConverted` · 기존 발급 시험군 |

Exact AST return positions: 420 · 424 · 427 · 450 · 455 · 489 · 493 · 498 · 500 · 536 · 538 · 557 · 568

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.journal.RecordDecisionAndReserveWithRecollection` | 원자 발급 | 한도 거부 즉시 | AST |

## State mutations and fallbacks

- journal tx.

## Safety conclusion

- Safe edit boundary: 요청 칸 하나.
- High-risk impact: yes
