# Function Logic Map: `Runner.preflightStatic`

- Source: `internal/verifylive/runner.go` (revision: current)
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
| B1 | `if step.Deferred != "" {` (internal/verifylive/runner.go:512, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B2 | `if step.OptIn != "" && !r.optedIn(step) {` (internal/verifylive/runner.go:518, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B3 | `for _, dep := range step.DependsOn {` (internal/verifylive/runner.go:521, range) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B4 | `if !passed(dep) {` (internal/verifylive/runner.go:522, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B5 | `if step.NeedsHolding && r.holdingSymbol == "" {` (internal/verifylive/runner.go:526, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B6 | `if symbol := r.mutationSymbol(step); step.Mutates && !SameMarket(MarketOf(symbol), r.market) {` (internal/verifylive/runner.go:531, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | run이 시장을 갖고, preflight가 하드코딩된 KR 대신 run의 시장과 심볼의 시장을 비교한다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- run이 시장을 갖고, preflight가 하드코딩된 KR 대신 run의 시장과 심볼의 시장을 비교한다.
- 주문 전송·취소·상한·계획 인가·승인 레일은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: 비교가 느슨해지면(예: 시장 불일치를 통과) 다른 시장의 규칙으로 실주문을 내게 된다. 승인·계획 인가·상한 레일은 이 함수 밖이며 무변경.
- High-risk impact: yes — 실계좌 주문 요청의 내용·대상을 결정한다.
