# Function Logic Map: `fakeBroker.PlaceOrder`

- Source: `internal/verifylive/fake_broker_test.go` (revision: current)
- AST evidence: `ast.json`
- Risk scan: `risk-pattern-report.md`
- Change: `apply-us-measurement-fixes`

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 입력 | 테스트가 구성한 하네스 상태 | 테스트 코드 | 실패는 t.Error/t.Fatal |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if f.placeAlreadyProcessing > 0 {` (internal/verifylive/fake_broker_test.go:368, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B2 | `if err := f.throttled("place"); err != nil {` (internal/verifylive/fake_broker_test.go:376, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B3 | `if f.placeErr != nil {` (internal/verifylive/fake_broker_test.go:379, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B4 | `if strings.EqualFold(intent.Side, "sell") && f.rejectOversell && intent.Quantity > f.sellable[intent.Symbol] {` (internal/verifylive/fake_broker_test.go:382, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B5 | `if intent.ClientOrderID != "" {` (internal/verifylive/fake_broker_test.go:388, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B6 | `if seen {` (internal/verifylive/fake_broker_test.go:392, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B7 | `switch {` (internal/verifylive/fake_broker_test.go:393, switch) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B8 | `case prior.body == body && f.honourIdempotency:` (internal/verifylive/fake_broker_test.go:394, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B9 | `case prior.body != body && f.conflictOnDifferentBody:` (internal/verifylive/fake_broker_test.go:396, case) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |
| B10 | `if intent.ClientOrderID != "" {` (internal/verifylive/fake_broker_test.go:408, if) | 아래 State mutations 참조 | 조건 불충족 시 조기 반환 또는 분기 | 이 함수 자체가 테스트다 |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| 하네스·fake broker 헬퍼 | 시나리오 구동 | 테스트 전용 | AST callees |

## State mutations and fallbacks

- 테스트/하네스 함수다 — httptest·fake broker만 사용하고 실계좌 side effect가 없다. 이 change에서의 변경은 시장 매개변수와 US 보유를 재현하기 위한 것이다.

## Safety conclusion

- Safe edit boundary: 거부 경로의 주장(‘mutating broker call 0건’)이 약해지지 않을 것.
- High-risk impact: no — 테스트 코드이며 프로덕션 경로가 아니다.
