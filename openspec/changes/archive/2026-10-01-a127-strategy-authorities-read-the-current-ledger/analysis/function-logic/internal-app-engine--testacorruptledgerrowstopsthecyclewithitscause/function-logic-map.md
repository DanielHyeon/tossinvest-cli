# Function Logic Map: `TestACorruptLedgerRowStopsTheCycleWithItsCause`

- Source: `internal/app/engine/a112_owner_scope_trading_test.go` (`541`–`571`)
- Qualified: `TestACorruptLedgerRowStopsTheCycleWithItsCause`
- AST evidence: `ast.json` (`source_sha256` 9f190f0d32244592…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 7 · return 0 · 호출 18

**역할.** a112 손상 원장 행 결함 시험. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:544` `if err != nil {` |
| B2 | if | `:547` `if _, err := stub.Exec(`INSERT INTO risk_bucket_reservations(reservation_id,account_ref,bucket_dimension,bu…` |
| B3 | range | `:555` `for _, scope := range fixture.risk.kr.scopes {` |
| B4 | if | `:558` `if scopes["005930"] \|\| !scopes["000660"] {` |
| B5 | if | `:562` `if placed := fixture.placedSymbols(); len(placed) != 0 {` |
| B6 | if | `:565` `if err == nil \|\| a112ScopeRefusalOf(err) != nil {` |
| B7 | if | `:568` `if !errors.Is(err, riskbucket.ErrProductionRiskSnapshotUnavailable) \|\| errors.Is(err, riskbucket.ErrProdu…` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: 손상 행을 실제 원장의 트리거가 막으므로 useRiskStub 으로 stub 을 명시 선택(판정 · 단언 불변).
- **High-risk impact**: no — 시험. 시험 전용 — 생산 동작 무관.
