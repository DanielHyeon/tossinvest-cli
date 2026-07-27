# Function Logic Map: `runState.snapshot`

- Source: `internal/console/run.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `verify-us-market`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이 함수의 입력 | 시그니처가 정의한 범위 | 호출자 | 범위 밖 값은 정규화되거나 거부된다 |
| run/화면의 시장 | KR 또는 US(zero value = KR) | verifylive.NormalizeMarket | 미지정은 KR로 해석 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if r.partial != "" {` (internal/console/run.go:160, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B2 | `if r.err != nil {` (internal/console/run.go:163, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | run이 시작 시점에 시장을 고정하고 그것으로 StartVerify를 부른다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- run이 시작 시점에 시장을 고정하고 그것으로 StartVerify를 부른다.
- 주문 전송·취소·상한·계획 인가·승인 레일은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: 실행 중 시장이 바뀔 수 있으면 승인된 계획과 기록되는 증거가 서로 다른 시장을 말하게 된다.
- High-risk impact: no — 배선·판독·렌더링이며 주문 자체를 만들지 않는다.
