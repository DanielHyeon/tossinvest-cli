# Function Logic Map: `PricingBasis.DescribeKO`

- Source: `internal/verifylive/plan.go` (revision: current)
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
| B1 | `switch b {` (internal/verifylive/plan.go:164, switch) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B2 | `case PriceFarBuy:` (internal/verifylive/plan.go:165, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B3 | `case PriceFarSell:` (internal/verifylive/plan.go:169, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B4 | `case PriceFarStop:` (internal/verifylive/plan.go:173, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B5 | `case PriceOneTickFurther:` (internal/verifylive/plan.go:176, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B6 | `case PriceIdenticalBody:` (internal/verifylive/plan.go:178, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B7 | `default:` (internal/verifylive/plan.go:181, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 계획 줄의 가격 설명이 시장을 말한다 — 밴드가 없는 시장에서 clamp를 주장하지 않는다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 계획 줄의 가격 설명이 시장을 말한다 — 밴드가 없는 시장에서 clamp를 주장하지 않는다.
- 주문 전송·취소·상한·계획 인가·승인 레일은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: KR 영문 문구는 바이트 단위로 보존되어야 한다(plan digest의 입력). 움직이면 재개된 run의 승인 증거가 깨진다.
- High-risk impact: no — 배선·판독·렌더링이며 주문 자체를 만들지 않는다.
