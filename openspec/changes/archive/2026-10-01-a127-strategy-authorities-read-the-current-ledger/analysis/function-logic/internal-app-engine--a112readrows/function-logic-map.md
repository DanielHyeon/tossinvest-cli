# Function Logic Map: `a112ReadRows`

- Source: `internal/app/engine/a112_owner_scope_trading_test.go` (`246`–`270`) — base 판본
- Qualified: `a112ReadRows`
- AST evidence: `ast.json` (`source_sha256` 31c63a8926764220…, revision base) — **편집 전(base — 이 함수는 a127 에서 지워짐)**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 5 · return 1 · 호출 15

**역할.** 다리의 행 읽기 헬퍼. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | if | `:249` `if err != nil {` |
| B2 | for | `:255` `for rows.Next() {` |
| B3 | range | `:258` `for i := range values {` |
| B4 | if | `:261` `if err := rows.Scan(pointers...); err != nil {` |
| B5 | if | `:266` `if err := rows.Err(); err != nil {` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: 다리와 함께 지움.
- **High-risk impact**: no — 시험. 시험 전용 — 생산 동작 무관.
