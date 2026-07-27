# Function Logic Map: `newHarness`

- Source: `internal/console/console_test.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `verify-clears-leftovers`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 입력 | 테스트가 구성한 하네스 상태 | 테스트 코드 | 실패는 t.Error/t.Fatal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `for _, f := range tweak {` (internal/console/console_test.go:111, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트 하네스다 |
| B2 | `if err != nil {` (internal/console/console_test.go:116, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트 하네스다 |
| B3 | `if err != nil {` (internal/console/console_test.go:121, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트 하네스다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 하네스·fake broker 헬퍼 | 시나리오 구동 | 테스트 전용 | AST callees |

## State mutations and fallbacks

- 테스트/하네스 함수다 — httptest·fake broker만 사용하고 실계좌 side effect가 없다.
- 이 change에서는 gofmt 정렬만 바뀌었다 — 테스트 하네스의 동작 무변경.

## Safety conclusion

- Safe edit boundary: 테스트 하네스이며 실계좌 side effect가 없다.
- High-risk impact: no — 테스트 코드이며 프로덕션 경로가 아니다.
