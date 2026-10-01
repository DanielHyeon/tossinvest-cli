# Function Logic Map: `readProductionRiskUsage`

- Source: `internal/riskbucket/production_snapshot_authority.go` (`479`–`513`)
- Qualified: `readProductionRiskUsage`
- AST evidence: `ast.json` (`source_sha256` 3aa9b66c00cdcded…) — **편집 전**(base `de3b4f65` 의 바이트, 커버리지 `analysis/impl/coverage-pre-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 3 · return 3 · 호출 7

**역할.** 계좌 bucket 의 예약 행을 읽는 원장 사용량의 유일한 SQL(a126 의 떠남 사실 포함). a127: 질의 문자열을 패키지 상수로 옮겨 적재기가 같은 상수로 prepare 하게 함(D7 — 둘째 철자 금지).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `db` | UsageQueryer(`*sql.DB` · `*sql.Tx`) | 호출자 | B1 질의 오류 되던짐 |
| 예약 · snapshot · policy · 영수증 · owners · scope_latches · final_decisions | 현재 원장 표 | 원장 | B3 스캔 오류 |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 커버리지로 만들었다. 「창의 return」은 위치, 「진입 실측」은 그 줄로 시작하는 커버리지 블록.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:498` `if err != nil {` | :499 | 아니오 |
| B2 | for | `:503` `for rows.Next() {` | — | 예 |
| B3 | if | `:505` `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,` | :508, :512 | — |

## Calls and live bindings

`db.QueryContext`(SQL 한 개) · `rows.Scan` · `rows.Err`. 브로커 없음.

## State mutations and fallbacks

없음 — 읽기 전용.

## Safety conclusion

- **Safe edit boundary**: 편집 전 — a127 은 SQL 문자열을 상수로 옮기는 것만(내용 · 분기 불변).
- **High-risk impact**: yes — 사이징 사용량 합의 입력.
