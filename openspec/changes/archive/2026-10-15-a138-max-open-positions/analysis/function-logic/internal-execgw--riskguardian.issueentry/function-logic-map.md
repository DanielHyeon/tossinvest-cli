# Function Logic Map: `IssueEntry`

- Source: `internal/execgw/riskguardian.go` (420-578)
- Qualified: `RiskGuardian.IssueEntry`
- AST evidence: `ast.json`(현재) — 현재 AST branches 14 · returns 13
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 7bf1bbe7623c179e…) — 편집 전 갈래 수 14
- Risk scan: `risk-pattern-report.md`

**역할.** 진입 발급 문 1 — 체인 · 재수집 · 원자 발급.

**a138 편집.** collectIssue 클로저의 ReserveRequest 에 `OpenPositionLimit: g.openPositionLimit()` 한 칸(분기 무변경).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if req.Collect == nil {` (if 421) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B2 | `if err != nil {` (if 425) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B3 | `if intent.Side != risk.SideBuy {` (if 428) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B4 | `if verdict, unmeasured := evaluateChainWithDailyLoss(in, measureErr); !verdict.Allowed {` (if 446) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B5 | `if !verdict.Allowed {` (if 455) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B6 | `if err != nil {` (if 490) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B7 | `if err != nil {` (if 494) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B8 | `if err != nil {` (if 499) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B9 | `if req.strategyPlan == nil {` (if 532) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B10 | `} else {` (else 534) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B11 | `if collectErr != nil {` (if 538) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B12 | `if strategyErr == nil {` (if 543) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B13 | `if err != nil {` (if 549) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |
| B14 | `if refused, ok := refusal.(*IssueRefusal); ok && refused.Stage == StageIssuance {` (if 556) | — | 분기 본문 참조 | `go test ./internal/execgw -run 'IssueEntry|A138'` · `TestA138IssueEntryRefusesTheEntryThatExceedsTheMaximum` |

Exact AST return positions: 422 · 426 · 429 · 452 · 457 · 491 · 495 · 500 · 502 · 539 · 541 · 560 · 571

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 26종 — 대표: `UTC`, `chainRefusal`, `collectIssue`, `dailyLossChainRefusal`, `errors.Join`, `errors.New`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
