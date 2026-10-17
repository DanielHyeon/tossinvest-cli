# Function Logic Map: `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter`

- Source: `internal/strategyworker/policy_receipt_test.go` (362-422, 편집 뒤)
- Qualified: `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` a768c56c186c7d77…) · `ast-base.json`(편집 전, base `63fd0c92`, `source_sha256` fa9fe7b92430e295…, 편집 전 분기수 8) — 현재 branches 8 · returns 3
- Risk scan: `risk-pattern-report.md`

**역할.** 엔진의 시장별 런타임 상태 칸 목록을 고정해, 임계값 1(엔진은 실패를 세지 않음)의 근거를 지키는 구조 시험.

**a137 편집.** a137 재리뷰 P1-N1 의 새 칸 셋(dailyLossUnmeasurable · firstDailyLossCause · dailyLossNoticeDay)을 기대 목록에 더하고 근거 주석(쓰는 곳 하나 · 읽는 곳 Snapshot · latchMarket 은 안 봄). 분기 불변.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 엔진 파일 AST | strategyMarketRuntime 칸 | 시험 | 목록 불일치 → FAIL |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, file := range engineFiles(t) {` (range 364) | — | 분기 본문 참조 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B2 | `if !ok \|\| spec.Name.Name != "strategyMarketRuntime" {` (if 367) | — | 분기 본문 참조 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B3 | `if !ok {` (if 371) | — | 분기 본문 참조 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B4 | `for _, field := range structure.Fields.List {` (range 375) | — | 분기 본문 참조 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B5 | `for _, name := range field.Names {` (range 376) | — | 분기 본문 참조 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B6 | `if fields == nil {` (if 383) | — | 분기 본문 참조 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B7 | `if strings.Join(fields, ",") != strings.Join(want, ",") {` (if 412) | — | 분기 본문 참조 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B8 | `if threshold := ProductionRuntimePolicy().FailureThreshold(); threshold != 1 {` (if 418) | — | 분기 본문 참조 | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |

Exact AST return positions: 368 · 372 · 380

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `engineFiles` | 엔진 소스 AST | — | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 기대 목록 · 주석.
- High-risk impact: no — 시험.
