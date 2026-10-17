# Function Logic Map: `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops`

- Source: `internal/app/engine/a112_owner_scope_trading_test.go` (`574`–`622`)
- Qualified: `TestAScopeLatchIsRefusedAloneButALatchReadFaultStops`
- AST evidence: `ast.json` (`source_sha256` 9f190f0d32244592…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 10 · return 0 · 호출 30

**역할.** a112 scope latch 시험. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:577` `if err != nil {` |
| B2 | if | `:580` `if _, err := stub.Exec(`INSERT INTO risk_bucket_scope_latches(account_ref,market,symbol,prospective_generat…` |
| B3 | if | `:586` `if got := strings.Join(latched.placedSymbols(), ","); got != "000660" {` |
| B4 | if | `:589` `if refusal := a112ScopeRefusalOf(err); refusal == nil \|\| refusal.scope.Symbol != "005930" \|\| !errors.Is…` |
| B5 | if | `:595` `if err != nil {` |
| B6 | range | `:600` `for _, statement := range []string{`DROP TABLE risk_bucket_scope_latches`,` |
| B7 | if | `:604` `if _, err := stub.Exec(statement); err != nil {` |
| B8 | range | `:611` `for _, scope := range unreadable.risk.kr.scopes {` |
| B9 | if | `:614` `if readiness["005930"] \|\| !readiness["000660"] {` |
| B10 | if | `:618` `if placed := unreadable.placedSymbols(); len(placed) != 0 \|\| err == nil \|\| a112ScopeRefusalOf(err) != n…` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: latch 행 · 뷰를 stub 에 심으므로 useRiskStub 으로 stub 을 명시 선택(판정 · 단언 불변).
- **High-risk impact**: no — 시험. 시험 전용 — 생산 동작 무관.
