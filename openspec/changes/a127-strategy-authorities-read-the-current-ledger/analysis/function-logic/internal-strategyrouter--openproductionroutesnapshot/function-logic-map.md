# Function Logic Map: `openProductionRouteSnapshot`

- Source: `internal/strategyrouter/production.go` (`600`–`622`)
- Qualified: `openProductionRouteSnapshot`
- AST evidence: `ast.json` (`source_sha256` 7d60a867a87576ca…) — **편집 전**(base `de3b4f65` 의 바이트, 커버리지 `analysis/impl/coverage-pre-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 4 · return 5 · 호출 10

**역할.** 원장 파일을 검증하고 읽기 전용 tx 를 연 뒤 `PRAGMA user_version` 을 확인. **편집 전 결함**: B4 `:616` 이 동결 리터럴 `productionRouteJournalV = 27` 과 비교 — 실제 원장 항상 거절, 원인은 맨 sentinel. a127: 주입 값 정확 일치 + 방향 문구(`%w`), 버전 확인 직후 owners · campaign SQL 상수 prepare(D1 · D7).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 원장 파일 | 소유자 · 0600 · 정규 파일 | `validateProductionRouteJournalFile` | B1 거절 |
| `PRAGMA user_version` | 편집 전: == 27(결함) · 편집 뒤: == 주입 값 | 원장 | B4 거절(편집 전 진입 0) |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 커버리지로 만들었다. 「창의 return」은 위치, 「진입 실측」은 그 줄로 시작하는 커버리지 블록.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:601` `if err := validateProductionRouteJournalFile(journalPath, ownerUID); err != nil {` | :602 | 아니오 |
| B2 | if | `:606` `if err != nil {` | :607 | 아니오 |
| B3 | if | `:611` `if err != nil {` | :613 | 아니오 |
| B4 | if | `:616` `if err := tx.QueryRowContext(ctx, `PRAGMA user_version`).Scan(&version); err != nil \|\| version != productionRouteJournalV {` | :619, :621 | 아니오 |

## Calls and live bindings

`validateProductionRouteJournalFile` · `sql.Open`(mode=ro · query_only · busy_timeout) · `db.SetMaxOpenConns(1)` · `db.BeginTx`(ReadOnly) · `tx.QueryRowContext`(PRAGMA user_version) · `tx.Rollback` · `db.Close`.

## State mutations and fallbacks

실패 시 tx rollback · db close. 성공 시 db · tx 를 호출자에게 넘김.

## Safety conclusion

- **Safe edit boundary**: 편집 전 — a127 은 B4 비교 · 문구와 prepare 추가만 바꿀 예정.
- **High-risk impact**: yes — route 권한의 원장 입구.
