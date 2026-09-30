# Function Logic Map: `createStrategyRiskLoaderJournal`

- Source: `internal/app/engine/strategy_risk_authority_test.go` (`201`–`224`)
- Qualified: `createStrategyRiskLoaderJournal`
- AST evidence: `ast.json` (`source_sha256` afb21e1c30354837…) — **편집 뒤**(시험 전용 fixture)
- Risk scan: `risk-pattern-report.md`
- AST branches 4 · return 0 · 호출 8

**역할.** v27식 축소 원장을 만드는 시험 fixture(engine 의 전략 위험 loader 시험용). 비례 원칙(WORKFLOW 「비례 원칙」): 시험 전용 — 무거운 규율 대상 아님, 게이트가 수정 함수로 세므로 번들을 둔다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `t` · 경로 · 원장 | 시험 임시 원장 | 시험 | `t.Fatal` |

## Branches and early returns

> 표는 `analysis/harness/branch_table.py` 가 만들었다(시험 코드라 커버리지 블록 없음 — `—`).

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:204` `if err != nil {` | — | — |
| B2 | range | `:207` `for _, statement := range []string{`PRAGMA user_version=27`,` | — | — |
| B3 | if | `:216` `if _, err := db.Exec(statement); err != nil {` | — | — |
| B4 | if | `:221` `if err := os.Chmod(path, 0o600); err != nil {` | — | — |

## Calls and live bindings

원장(시험 임시 DB) 문장 실행과 `t.Fatal`. 생산 코드 · 브로커 호출 없음.

## State mutations and fallbacks

시험 임시 원장만.

## Safety conclusion

- **Safe edit boundary**: `CREATE TABLE` 문만 바뀌었다 — reservations 에 owner 키 열 4, scope_latches 에 prospective_generation, 빈 표 셋(영수증 · owners · final_decisions). 사용량 reader 의 새 조인이 성립하게 하는 것뿐이며 어떤 INSERT · 단언 · 기대값도 바뀌지 않았다(Manager 조건 영수증 — review 1.0.4).
- **High-risk impact**: no — 시험 fixture.
