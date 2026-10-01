# Function Logic Map: `a112MirrorLedgerIntoRiskStub`

- Source: `internal/app/engine/a112_owner_scope_trading_test.go` (`172`–`208`) — base 판본
- Qualified: `a112MirrorLedgerIntoRiskStub`
- AST evidence: `ast.json` (`source_sha256` 31c63a8926764220…, revision base) — **편집 전(base — 이 함수는 a127 에서 지워짐)**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 8 · return 1 · 호출 24

**역할.** 원장 → stub 복사 다리(a112 임시). 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:175` `if err != nil {` |
| B2 | if | `:180` `if err != nil {` |
| B3 | if | `:185` `if len(tables) == 0 {` |
| B4 | range | `:189` `for table, columns := range tables {` |
| B5 | if | `:191` `if _, err := stub.Exec("DELETE FROM " + table); err != nil {` |
| B6 | range | `:196` `for _, row := range source {` |
| B7 | if | `:197` `if _, err := stub.Exec(fmt.Sprintf("INSERT INTO %s(%s) VALUES(%s)", table, columns, placeholders), row...);…` |
| B8 | if | `:202` `if copied := a112ReadRows(t, stub, table, columns); !reflect.DeepEqual(copied, source) {` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: a127 이 스키마 핀을 고쳐 다리가 필요 없어져 지움(트립와이어가 예고한 제거 조건).
- **High-risk impact**: no — 시험. 시험 전용 — 생산 동작 무관.
