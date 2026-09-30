# Function Logic Map: `createProductionRiskDB`

- Source: `internal/riskbucket/production_snapshot_authority_test.go` (`215`–`240`)
- Qualified: `createProductionRiskDB`
- AST evidence: `ast.json` (`source_sha256` 308092e288586bd0…) — **편집 뒤**(시험 전용 fixture)
- Risk scan: `risk-pattern-report.md`
- AST branches 4 · return 0 · 호출 8

**역할.** v27식 축소 원장을 만드는 시험 fixture(생산 snapshot reader 시험용). 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님, 게이트가 수정 함수로 세므로 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` · 경로 · 원장 | 시험 임시 원장 | 시험 | `t.Fatal` |

## Branches and early returns

> 표는 `analysis/harness/branch_table.py` 가 만들었다(시험 코드라 커버리지 블록 없음 — `—`).

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | range | `:229` `for _, statement := range statements {` | — | — |
| B2 | if | `:230` `if _, err := db.Exec(statement); err != nil {` | — | — |
| B3 | if | `:234` `if err := db.Close(); err != nil {` | — | — |
| B4 | if | `:237` `if err := os.Chmod(path, 0o600); err != nil {` | — | — |

## Calls and live bindings

원장(시험 임시 DB) 문장 실행과 `t.Fatal`. 생산 코드 · 브로커 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: 위와 같은 `CREATE TABLE` 편집 — 단언 무변.
- **High-risk impact**: no — 시험 fixture.
