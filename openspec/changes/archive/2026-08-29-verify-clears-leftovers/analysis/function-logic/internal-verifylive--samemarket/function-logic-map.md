# Function Logic Map: `SameMarket`

- Source: `internal/verifylive/pricing.go` (revision: current)
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
| B1 | 분기 없음 — 단일 경로 (internal/verifylive/pricing.go) | 아래 State mutations 참조 | 정상 반환 | internal/verifylive/cleanup_test.go, internal/console/retry_after_run_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 이 change에서는 gofmt 정렬만 바뀌었다 — 동작·시그니처·분기 무변경. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 이 change에서는 gofmt 정렬만 바뀌었다 — 동작·시그니처·분기 무변경.
- 승인 게이트·계획 인가(Plan.Authorises)·노출 상한·1주 규칙은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: 시장 비교의 의미가 바뀌면 다른 시장 규칙으로 주문이 나간다. 이 change는 그것을 건드리지 않는다.
- High-risk impact: no — 배선·판독·렌더링이며 요청 자체를 만들지 않는다.
