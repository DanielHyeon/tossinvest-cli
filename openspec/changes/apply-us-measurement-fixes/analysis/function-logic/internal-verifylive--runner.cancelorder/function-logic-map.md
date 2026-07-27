# Function Logic Map: `Runner.cancelOrder`

- Source: `internal/verifylive/mutate.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `apply-us-measurement-fixes`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 이 함수의 입력 | 시그니처가 정의한 범위 | 호출자 | 범위 밖 값은 정규화되거나 거부된다 |
| run/화면의 시장 | KR 또는 US(zero value = KR) | verifylive.NormalizeMarket | 미지정은 KR로 해석 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := r.gate(sr, request{` (internal/verifylive/mutate.go:297, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B2 | `for attempts = 1; ; attempts++ {` (internal/verifylive/mutate.go:309, for) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B3 | `if err == nil {` (internal/verifylive/mutate.go:313, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B4 | `if !transient \|\| attempts > CancelRetryAttempts {` (internal/verifylive/mutate.go:317, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B5 | `if sleepErr := r.sleep(ctx, wait); sleepErr != nil {` (internal/verifylive/mutate.go:322, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B6 | `if attempts > 1 {` (internal/verifylive/mutate.go:327, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B7 | `if err != nil {` (internal/verifylive/mutate.go:333, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B8 | `if id := strings.TrimSpace(res.CurrentOrderID); id != "" && id != orderID {` (internal/verifylive/mutate.go:338, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 정정 요청이 시장의 브로커 계약을 따른다 — KR은 수량 필수, US는 수량 전달 불가. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 정정 요청이 시장의 브로커 계약을 따른다 — KR은 수량 필수, US는 수량 전달 불가.
- 주문 전송·취소·상한·계획 인가·승인 레일은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: US에 수량을 실으면 400 us-modify-quantity-not-supported로 거절된다(측정 실패). 반대로 KR에서 수량을 빼면 정정 자체가 거절된다. 승인 목록의 문구도 실제 요청과 일치해야 한다.
- High-risk impact: yes — 실계좌 주문 요청의 내용·대상을 결정한다.
