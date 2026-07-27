# Function Logic Map: `TestACancelRetriesWhileTheBrokerSaysAlreadyProcessing`

- Source: `internal/verifylive/us_market_test.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `apply-us-measurement-fixes`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 입력 | 테스트가 구성한 하네스 상태 | 테스트 코드 | 실패는 t.Error/t.Fatal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if _, err := h.run(Options{HoldingSymbol: "005930"}); err != nil {` (internal/verifylive/us_market_test.go:240, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B2 | `if anyFailureMentioning(h.entries(), "already-processing") {` (internal/verifylive/us_market_test.go:246, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B3 | `for _, a := range Outstanding(h.entries()) {` (internal/verifylive/us_market_test.go:249, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B4 | `if !a.Deliberate {` (internal/verifylive/us_market_test.go:252, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B5 | `if !anyObservation(h.entries(), "order.cancel.retries") {` (internal/verifylive/us_market_test.go:257, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 하네스·fake broker 헬퍼 | 시나리오 구동 | 테스트 전용 | AST callees |

## State mutations and fallbacks

- 테스트/하네스 함수다 — httptest·fake broker만 사용하고 실계좌 side effect가 없다. 이 change에서의 변경은 시장 매개변수와 US 보유를 재현하기 위한 것이다.

## Safety conclusion

- Safe edit boundary: 거부 경로의 주장(‘mutating broker call 0건’)이 약해지지 않을 것.
- High-risk impact: no — 테스트 코드이며 프로덕션 경로가 아니다.
