# Function Logic Map: `transientCancelRefusal`

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
| B1 | `if !errors.As(err, &apiErr) \|\| apiErr.Code != http.StatusConflict {` (internal/verifylive/mutate.go:370, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B2 | `if jsonErr := json.Unmarshal([]byte(apiErr.Body), &body); jsonErr != nil {` (internal/verifylive/mutate.go:381, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B3 | `if body.Error.Code != "already-processing" {` (internal/verifylive/mutate.go:384, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B4 | `if wait <= 0 {` (internal/verifylive/mutate.go:388, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B5 | `if wait > CancelRetryMaxWait {` (internal/verifylive/mutate.go:391, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

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
