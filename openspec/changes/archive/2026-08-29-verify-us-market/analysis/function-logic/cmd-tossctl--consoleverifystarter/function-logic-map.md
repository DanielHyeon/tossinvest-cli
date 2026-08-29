# Function Logic Map: `consoleVerifyStarter`

- Source: `cmd/tossctl/console.go` (revision: current)
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
| B1 | `if err != nil {` (cmd/tossctl/console.go:390, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B2 | `if err != nil {` (cmd/tossctl/console.go:394, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B3 | `if err != nil {` (cmd/tossctl/console.go:398, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B4 | `if err != nil {` (cmd/tossctl/console.go:403, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B5 | `if err != nil {` (cmd/tossctl/console.go:425, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |
| B6 | `if runErr != nil && (errors.Is(runErr, context.Canceled) \|\| errors.Is(runErr, context.DeadlineExceeded)) {` (cmd/tossctl/console.go:434, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | internal/verifylive/us_market_test.go, internal/console/us_market_test.go 및 기존 패키지 테스트 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| AST callees 참조 | 콘솔 배선이 두 시장의 기록을 넘기고, run마다 시장에 맞는 기록·심볼·보유를 해석한다. | 호출자 계약을 따른다 | ast.json |

## State mutations and fallbacks

- 콘솔 배선이 두 시장의 기록을 넘기고, run마다 시장에 맞는 기록·심볼·보유를 해석한다.
- 주문 전송·취소·상한·계획 인가·승인 레일은 이 함수 밖이며 무변경이다.

## Safety conclusion

- Safe edit boundary: 승인 채널(클릭)과 단계별 확인 부재는 무변경. 시장별 기록 경로가 어긋나면 증거가 섞인다.
- High-risk impact: no — 배선·판독·렌더링이며 주문 자체를 만들지 않는다.
