# Function Logic Map: `runConsole`

- Source: `cmd/tossctl/console.go`
- AST evidence: `ast.json` (구현 후 추출)
- Risk scan: `risk-pattern-report.md`

기존 함수 — Options 리터럴에 Settings seam 1행 주입만 추가. 다른 배선 무변경.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| (함수 서술 참조 — 이 change의 delta·design D1~D6이 계약) | — | 현재 HEAD + 위 테스트 | 테스트 실패로 관측 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | if | 없음 | — | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green |
| B2 | if | 없음 | — | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green |
| B3 | if | 없음 | — | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green |
| B4 | if | 없음 | — | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green |
| B5 | if | 없음 | — | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green |
| B6 | if | 없음 | — | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green |
| B7 | else | 없음 | — | internal/console settings_test.go 전체(주입된 seam 동작) + cmd/tossctl 기존 콘솔 테스트 green |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| ast.json calls 참조 | 목적 서술 참조 | 기존 계약 무변경 | ast.json + HEAD |

## State mutations and fallbacks

- 없음(주입 구성만)

## Safety conclusion

- Safe edit boundary: Settings 필드 1행 — 다른 seam 무접촉
- High-risk impact: no (콘솔·config·배선 — 주문·원장 무접촉)
