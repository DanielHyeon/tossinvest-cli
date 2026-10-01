# Function Logic Map: `createProductionRiskDB`

- Source: `internal/riskbucket/production_snapshot_authority_test.go` (`220`–`245`)
- Qualified: `createProductionRiskDB`
- AST evidence: `ast.json` (`source_sha256` c578d9fd751b2112…, revision current) — **편집 뒤**(시험 전용)
- Risk scan: `risk-pattern-report.md`
- AST branches 4 · return 0 · 호출 9

**역할.** riskbucket 축소 원장 픽스처. 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님. 게이트가 수정 함수로 세므로 경량 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 시험 임시 원장 · fixture | 시험 | 생산 작성자 경로 + 손상만 SQL | `t.Fatal`/`t.Errorf` |

## Branches and early returns

> AST 분기 열거(시험 코드 — 커버리지 블록 없음).

| Branch | 종류 | 조건 (원문) |
|---|---|---|
| B1 | range | `:234` `for _, statement := range statements {` |
| B2 | if | `:235` `if _, err := db.Exec(statement); err != nil {` |
| B3 | if | `:239` `if err := db.Close(); err != nil {` |
| B4 | if | `:242` `if err := os.Chmod(path, 0o600); err != nil {` |

## Calls and live bindings

시험 fixture · 생산 journal/riskbucket API 호출. 브로커 · 실계좌 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: user_version 을 시험 상수 1 로(행 · 단언 불변).
- **High-risk impact**: no — 시험. 시험 전용 — 생산 동작 무관.
