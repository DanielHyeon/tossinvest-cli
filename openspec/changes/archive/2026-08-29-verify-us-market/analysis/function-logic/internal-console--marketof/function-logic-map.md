# Function Logic Map: `marketOf`

- Source: `internal/console/pages.go` (revision: current)
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
| B1 | `if v := strings.TrimSpace(r.URL.Query().Get("market")); v != "" {` (internal/console/pages.go:127, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 화면과 폼이 시장을 실어 나른다. 승인 형식(클릭 1회)은 무변경이다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 화면과 폼이 시장을 실어 나른다. 승인 형식(클릭 1회)은 무변경이다.
- 주문 전송·취소·상한·계획 인가·승인 레일은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: 폼이 시장을 잃으면 US 화면의 버튼이 KR run을 시작한다 — 사용자가 보지 않은 시장의 계획을 승인하게 된다. 재측정 대상은 계속 기록에서만 계산한다.
- High-risk impact: no — 배선·판독·렌더링이며 주문 자체를 만들지 않는다.
