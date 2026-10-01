# Function Logic Map: `newProductionRouteFixture`

- Source: `internal/strategyrouter/production_test.go` (`212`–`251`)
- Qualified: `newProductionRouteFixture`
- AST evidence: `ast.json` (`source_sha256` 21d69cb2a4f5da5a…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 7 · return 1 · 호출 19

**역할.** route 생산 픽스처. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:217` `if err != nil {` |
| B2 | range | `:220` `for _, statement := range []string{` |
| B3 | if | `:225` `if _, err := db.Exec(statement); err != nil {` |
| B4 | if | `:229` `if err := db.Close(); err != nil {` |
| B5 | if | `:232` `if err := os.Chmod(journal, 0o600); err != nil {` |
| B6 | if | `:236` `if err != nil {` |
| B7 | range | `:241` `for _, market := range []Market{MarketKR, MarketUS} {` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: user_version 과 config 주입 값을 시험 상수 1 로(행 · 단언 불변).
- **High-risk impact**: no — 시험. 시험 전용 — 생산 동작 무관.
