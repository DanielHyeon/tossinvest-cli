# Function Logic Map: `readProductionRiskUsage`

- Source: `internal/riskbucket/production_snapshot_authority.go` (`452`–`473`)
- Qualified: `readProductionRiskUsage`
- AST evidence: `ast.json` (`source_sha256` 5f26bf28bbf8207f…) — **편집 전**(base `a189e74f`, freeze census `989ab031` 과 sha · 분기 일치)
- Risk scan: `risk-pattern-report.md`
- AST branches 3 · return 3

**역할.** 계좌 bucket(dimension, value)의 모든 예약 행을 읽는다 — 원장 사용량의 **유일한** SQL(a066 5.6.1 F1). 오늘은 owner 해제 여부를
보지 않는다(design 「현재 동작」).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `db` | `*sql.DB` 또는 `*sql.Tx`(UsageQueryer) | 호출자 | 질의 오류는 B1 로 되던짐 |
| `account, dimension, value` | 계좌 · dimension enum · bucket 값 | 호출자 | WHERE 조건 — 행 0 이면 빈 목록 |
| `risk_bucket_reservations` · `risk_bucket_snapshots` · `risk_bucket_policies` | v22+ 스키마 | 원장 | 스캔 오류 B3 되던짐 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `:459` `if err != nil {` — 질의 오류 | — | `:460` `nil, err` | 미진입(pre-edit 커버리지) — 기존, a126 무변 |
| B2 | `:464` `for rows.Next() {` | `result` 에 행 추가 | — | 진입 — a126 1.1 전부 |
| B3 | `:466` 스캔 오류 | — | `:468` `nil, err` | 미진입 — 기존, a126 무변 |

## Calls and live bindings

`db.QueryContext`(SQL 한 개) · `rows.Scan` · `rows.Err`. 브로커 호출 없음. 오류는 호출자(`ReadJournalBucketUsage`)로 되던진다.

## State mutations and fallbacks

없다 — 읽기 전용.

## Safety conclusion

- **Safe edit boundary (a126 D1)**: SQL 에 영수증 · owner released_at · scope latch · decision 사본의 조인을 더해 행마다 떠남 판정 입력을 읽는다.
  분기 구조(B1~B3)는 그대로다. 떠남 **규칙**은 `aggregateProductionRiskUsage` 한 곳이 적용한다 — SQL 은 사실만 싣는다.
- **High-risk impact**: yes — 사이징(진입 cap)의 사용량 합을 정한다.
