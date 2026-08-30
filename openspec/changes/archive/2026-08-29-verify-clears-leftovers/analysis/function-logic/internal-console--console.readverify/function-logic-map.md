# Function Logic Map: `Console.readVerify`

- Source: `internal/console/data.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `verify-clears-leftovers`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이 함수의 입력 | 시그니처가 정의한 범위 | 호출자 | 범위 밖 값은 정규화되거나 거부된다 |
| 증거 기록의 Outstanding | 이 도구가 만들고 취소되지 않은 객체 | capability-verify*.jsonl | 기록이 없으면 대상도 없다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if strings.TrimSpace(v.Record) == "" {` (internal/console/data.go:298, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B2 | `if err != nil {` (internal/console/data.go:302, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B3 | `for _, s := range progress.Steps {` (internal/console/data.go:317, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |
| B4 | `if s.Verdict.Terminal() {` (internal/console/data.go:318, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 |  | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 승인 게이트·계획 인가(Plan.Authorises)·노출 상한·1주 규칙은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: 
- High-risk impact: no — 배선·판독·렌더링이며 요청 자체를 만들지 않는다.
