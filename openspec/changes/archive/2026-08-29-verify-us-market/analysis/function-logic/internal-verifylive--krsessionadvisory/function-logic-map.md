# Function Logic Map: `KRSessionAdvisory`

- Source: `internal/verifylive/hours.go` (revision: base)
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
| B1 | `switch {` (internal/verifylive/hours.go:74, switch) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B2 | `case weekend:` (internal/verifylive/hours.go:75, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B3 | `case hhmm >= krRegularOpen && hhmm < krRegularClose:` (internal/verifylive/hours.go:80, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B4 | `default:` (internal/verifylive/hours.go:86, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 시장별 세션 판정을 추가했다. US는 internal/clock(IANA·DST)을 쓰고, 문구는 휴장 응답이 미측정임을 명시한다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 시장별 세션 판정을 추가했다. US는 internal/clock(IANA·DST)을 쓰고, 문구는 휴장 응답이 미측정임을 명시한다.
- 주문 전송·취소·상한·계획 인가·승인 레일은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: advisory는 차단하지 않는다. US 문구가 KR 실측 코드를 자기 근거로 제시하면 관측하지 않은 사실을 단언하는 것이 된다.
- High-risk impact: no — 배선·판독·렌더링이며 주문 자체를 만들지 않는다.
