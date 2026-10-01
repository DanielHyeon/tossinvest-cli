# Function Logic Map: `loadProductionRouteOwnersFrom`

- Source: `internal/strategyrouter/production.go` (`624`–`685`)
- Qualified: `loadProductionRouteOwnersFrom`
- AST evidence: `ast.json` (`source_sha256` 7d60a867a87576ca…) — **편집 전**(base `de3b4f65` 의 바이트, 커버리지 `analysis/impl/coverage-pre-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 14 · return 12 · 호출 23

**역할.** owner 이력을 읽어 active owner 를 재구성하고, active owner 가 있을 때만 position_campaigns 를 읽어 대조. **조건부 질의**: B10 `:659` 에서 active owner 가 없으면 campaign 질의 전에 성공 반환(codex 1R P1). a127: 두 SQL 을 패키지 상수로 옮겨 opener 가 같은 상수로 prepare(D7).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `queryer` | nil 아님 | 호출자(tx) | B1 거절 |
| owner 이력 행 | ≤ 상한 · 정규 값 | `risk_bucket_owners` | B2 · B4 · B5 거절 |
| active owner | ≤ 1 · actual 있음 · latch 없음 | 원장 | B9 · B11 · B12 거절 |
| campaign | 키 일치 · ACTIVE · 진입 막힘 아님 | `position_campaigns` | B13 거절 |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 커버리지로 만들었다. 「창의 return」은 위치, 「진입 실측」은 그 줄로 시작하는 커버리지 블록.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:625` `if queryer == nil {` | :626 | 아니오 |
| B2 | if | `:629` `if err != nil {` | :630 | 아니오 |
| B3 | for | `:634` `for rows.Next() {` | — | 예 |
| B4 | if | `:635` `if len(history) >= productionRouteMaxOwners {` | :636 | 아니오 |
| B5 | if | `:639` `if err := rows.Scan(&value.prospective, &value.laneID, &value.campaignID, &value.actual, &value.acquired, &value.released, &value.overage…` | :641 | — |
| B6 | if | `:645` `if err := rows.Err(); err != nil {` | :646 | 예 |
| B7 | range | `:651` `for _, value := range history {` | — | 예 |
| B8 | if | `:652` `if value.released == "" {` | — | 예 |
| B9 | if | `:656` `if len(active) > 1 {` | :657 | 예 |
| B10 | if | `:659` `if len(active) == 0 {` | :660 | 예 |
| B11 | if | `:663` `if value.actual == "" \|\| value.overage != 0 \|\| value.unknown != 0 {` | :664 | 예 |
| B12 | if | `:667` `if err != nil \|\| actual != key.PositionGeneration \|\| strconv.FormatUint(actual, 10) != value.actual {` | :668 | 아니오 |
| B13 | if | `:672` `if err := queryer.QueryRowContext(ctx, `SELECT account_ref,market,symbol,lane_version,prospective_token,coalesce(actual_position_generati…` | :676 | — |
| B14 | if | `:679` `if !ok \|\| descriptor.LaneVersion != laneVersion {` | :680, :684 | 아니오 |

## Calls and live bindings

`queryer.QueryContext`(owners SQL) · `rows.Scan` · `validProductionOwnerRow` · `rows.Err` · `productionOwnerHistoryDigest` · `strconv.ParseUint` · `queryer.QueryRowContext`(campaign SQL) · `productionRouteDescriptors`.

## State mutations and fallbacks

없음 — 읽기 전용.

## Safety conclusion

- **Safe edit boundary**: 편집 전 — a127 은 두 SQL 문자열을 상수로 옮기는 것만(내용 · 분기 불변).
- **High-risk impact**: yes — owner 재구성(route 권한의 기반).
