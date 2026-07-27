# Function Logic Map: `SameMarket`

- Source: `internal/verifylive/pricing.go` (revision: current)
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
| B1 | 분기 없음 — 단일 경로 (internal/verifylive/pricing.go) | 아래 State mutations 참조 | 정상 반환 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 시장 상수와 시장 비교를 도입했다. 가격 산식은 손대지 않았다 — US 밴드 부재는 이미 `lowerLimit > 0` 가드로 처리되고 있었고 TickSize도 US 그리드를 이미 구현한다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 시장 상수와 시장 비교를 도입했다. 가격 산식은 손대지 않았다 — US 밴드 부재는 이미 `lowerLimit > 0` 가드로 처리되고 있었고 TickSize도 US 그리드를 이미 구현한다.
- 주문 전송·취소·상한·계획 인가·승인 레일은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: zero value가 KR이어야 한다. 여기서 기본값이 US로 바뀌면 시장을 말하지 않은 모든 호출자가 다른 시장에 주문을 내게 된다.
- High-risk impact: no — 배선·판독·렌더링이며 주문 자체를 만들지 않는다.
