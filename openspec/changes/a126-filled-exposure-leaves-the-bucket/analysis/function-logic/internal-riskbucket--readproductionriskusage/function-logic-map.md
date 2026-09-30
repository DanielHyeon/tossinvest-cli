# Function Logic Map: `readProductionRiskUsage`

- Source: `internal/riskbucket/production_snapshot_authority.go` (`458`–`492`)
- Qualified: `readProductionRiskUsage`
- AST evidence: `ast.json` (`source_sha256` c5e0c64b2de65248…) — **편집 뒤**(구현 로트, 격리 워크트리). 편집 전 판은 `467322df`
- Risk scan: `risk-pattern-report.md`
- AST branches 3 · return 3 · 호출 7

**역할.** 계좌 bucket 의 모든 예약 행을 읽는다 — 원장 사용량의 유일한 SQL. a126 뒤에는 행마다 **떠남 판정의 원장 사실**을 함께 싣는다(D1): 예약 행 owner 키(r.*)의
영수증 존재 · 영수증 released_at · owner released_at · 그 owner 키의 scope latch 존재 · 예약 행과 결정 사본(d.*)의 owner 키 일치 · 결정 사본 쪽 영수증 존재.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `db` | UsageQueryer | 호출자 | 질의 오류 B1 되던짐 |
| 예약 · snapshot · policy · 영수증 · owners · scope_latches · final_decisions | v24+ 원장 표(축소 픽스처는 빈 표) | 원장 | 스캔 오류 B3 |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 편집 뒤 커버리지(`analysis/impl/coverage-post-edit.out`, `-coverpkg=./internal/riskbucket,./internal/journal`)로 만들었다. 「창의 return」은 위치다.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:477` `if err != nil {` | :478 | 아니오 |
| B2 | for | `:482` `for rows.Next() {` | — | 예 |
| B3 | if | `:484` `if err := rows.Scan(&row.ReservationID, &row.PolicyVersion, &row.HeldMinor, &row.FilledMinor, &row.State,` | :487, :491 | — |

## Calls and live bindings

`db.QueryContext`(SQL 한 개 — LEFT JOIN 영수증 · owners · 결정, EXISTS scope latch · 결정 쪽 영수증) · `rows.Scan` · `rows.Err`. 브로커 호출 없음. 오류는 되던짐.

## State mutations and fallbacks

없다 — 읽기 전용. 규칙은 싣지 않는다: 떠남 · 거절 판정은 `aggregateProductionRiskUsage` 한 곳.

## Safety conclusion

- **Safe edit boundary**: SQL 의 열 · 조인만 늘었다. 분기 B1~B3 는 편집 전과 같은 구조(질의 오류 · 순회 · 스캔 오류). 스키마 · 저장값 무변.
- **High-risk impact**: yes — 사이징의 사용량 합.
