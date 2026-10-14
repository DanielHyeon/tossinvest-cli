# Function Logic Map: `TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker`

- Source: `internal/execgw/riskguardian_qfinal_test.go` (215-258)
- Qualified: `TestGatewayRefusesQFinalMarkedDecisionWithoutExactAdmissionBeforeBroker`
- AST evidence: `ast.json`(현재) — 현재 AST branches 6 · returns 0
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 2f2a1e6036fcf9c6…) — 편집 전 갈래 수 6
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 — q_final 표시 결정의 Gateway 거부.

**a138 편집.** 요청에 `OpenPositionLimit` 넉넉한 칸 하나.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 219) | — | 분기 본문 참조 | `go test ./internal/execgw -run QFinal` |
| B2 | `if err != nil {` (if 223) | — | 분기 본문 참조 | `go test ./internal/execgw -run QFinal` |
| B3 | `if err != nil {` (if 227) | — | 분기 본문 참조 | `go test ./internal/execgw -run QFinal` |
| B4 | `if err != nil {` (if 245) | — | 분기 본문 참조 | `go test ./internal/execgw -run QFinal` |
| B5 | `if err != nil {` (if 249) | — | 분기 본문 참조 | `go test ./internal/execgw -run QFinal` |
| B6 | `if !errors.As(err, &rejected) \|\| rejected.Reason != execgw.ReasonGuardianRiskBucketMismatch \|\| places != 0 {` (if 255) | — | 분기 본문 참조 | `go test ./internal/execgw -run QFinal` |

Exact AST return positions: (없음)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 16종 — 대표: `Add`, `a136ZeroManualBuy`, `broker.totals`, `clk.Now`, `context.Background`, `errors.As`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
