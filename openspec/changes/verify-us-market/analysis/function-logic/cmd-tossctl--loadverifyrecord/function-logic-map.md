# Function Logic Map: `loadVerifyRecord`

- Source: `cmd/tossctl/verify.go` (revision: current)
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
| B1 | `if err != nil {` (cmd/tossctl/verify.go:492, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B2 | `if err != nil {` (cmd/tossctl/verify.go:496, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | CLI가 --market을 받고, 기록 경로·보유 종목·프로브 심볼이 그 시장을 따른다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- CLI가 --market을 받고, 기록 경로·보유 종목·프로브 심볼이 그 시장을 따른다.
- 주문 전송·취소·상한·계획 인가·승인 레일은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: 기본값은 KR이어야 하며, US run이 KR 기본 심볼을 그대로 쓰면 계획이 전부 배제된다.
- High-risk impact: no — 배선·판독·렌더링이며 주문 자체를 만들지 않는다.
