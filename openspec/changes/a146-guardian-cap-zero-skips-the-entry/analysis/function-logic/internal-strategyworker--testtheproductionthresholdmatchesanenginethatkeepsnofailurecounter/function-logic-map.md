# Function Logic Map: `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter`

- Source: `internal/strategyworker/policy_receipt_test.go` (362-422)
- Qualified: `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter`
- AST evidence: `ast.json`(현재) — 현재 AST branches 8 · returns 3
- 편집 전 기준: `ast-base.json`(base, `source_sha256` a768c56c186c7d77…) — 편집 전 갈래 수 8
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 — 엔진 시장 런타임 칸 목록이 「실패를 세지 않는다」는 임계값 1 의 영수증.

**a146 편집.** 목록에 a146 칸 넷(`guardianCapZeroSkips` · `firstGuardianCapZeroSkip` · `entrySkipNoticeDay` · `entrySkipNoticed`)과 근거 주석 — 쓰는 곳은 `recordGuardianCapZeroSkip` 하나, 읽는 곳은 `Snapshot` 하나, `latchMarket` 은 넷 중 어느 것도 보지 않음(세는 것은 실패가 아니라 정책 결과).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, file := range engineFiles(t) {` (range 364) | — | 편집 전과 같음(소스 줄 참조) | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` · `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B2 | `if !ok \|\| spec.Name.Name != "strategyMarketRuntime" {` (if 367) | — | 편집 전과 같음(소스 줄 참조) | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` · `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B3 | `if !ok {` (if 371) | — | 편집 전과 같음(소스 줄 참조) | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` · `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B4 | `for _, field := range structure.Fields.List {` (range 375) | — | 편집 전과 같음(소스 줄 참조) | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` · `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B5 | `for _, name := range field.Names {` (range 376) | — | 편집 전과 같음(소스 줄 참조) | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` · `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B6 | `if fields == nil {` (if 383) | — | 편집 전과 같음(소스 줄 참조) | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` · `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B7 | `if strings.Join(fields, ",") != strings.Join(want, ",") {` (if 412) | — | 편집 전과 같음(소스 줄 참조) | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` · `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |
| B8 | `if threshold := ProductionRuntimePolicy().FailureThreshold(); threshold != 1 {` (if 418) | — | 편집 전과 같음(소스 줄 참조) | `go test ./internal/strategyworker -run TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` · `TestTheProductionThresholdMatchesAnEngineThatKeepsNoFailureCounter` |

Exact AST return positions: 368 · 372 · 380

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 9종 — 대표: `FailureThreshold`, `ProductionRuntimePolicy`, `append`, `ast.Inspect`, `engineFiles`, `make`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a146 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a146 편집」에 적은 한 자리.
- High-risk impact: yes(전략 첫 다리 진입 분류 — 주문을 늘리지 않음: 건너뛴 진입은 브로커 0, 매도 · 손절 · 청산 경로 무관)
