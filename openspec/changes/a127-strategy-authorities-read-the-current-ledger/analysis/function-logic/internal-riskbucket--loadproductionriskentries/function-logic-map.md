# Function Logic Map: `loadProductionRiskEntries`

- Source: `internal/riskbucket/production_snapshot_authority.go` (`361`–`448`)
- Qualified: `loadProductionRiskEntries`
- AST evidence: `ast.json` (`source_sha256` 3aa9b66c00cdcded…) — **편집 전**(base `de3b4f65` 의 바이트, 커버리지 `analysis/impl/coverage-pre-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 14 · return 14 · 호출 47

**역할.** 원장을 읽기 전용으로 열어 버전 확인 · 범위 scope latch · 다섯 dimension 사용량을 읽고 snapshot 항목을 만듦. **편집 전 결함**: B5 `:384` 가 `PRAGMA user_version` 을 동결 리터럴 `productionRiskJournalSchema = 27` 과 비교 — 실제 원장(v35)은 항상 거절. 판독은 각각 autocommit(`db`), 사용량 질의는 scope latch 가 0 일 때만 실행(조건부).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 원장 파일 | 소유자 · 0600 · 정규 파일(심링크 아님) | `validateProductionRiskJournalFile` | B2 거절 |
| `PRAGMA user_version` | 편집 전: == 27(결함) · 편집 뒤: == 주입 값 | 원장 | B5 거절(편집 전 진입 0 — 픽스처가 27) |
| scope latch 행 | 0 | `risk_bucket_scope_latches` | B6 오류 · B7 범위 국소 거절(ScopeRefused) |
| 권한 창 | manifest · 가격 · FX 관측 ≤ AsOf < fresh | 정책 · 입력 | B8 거절 |
| 다섯 dimension 사용량 | 판독 가능 · latch 없음 | `ReadJournalBucketUsage` | B10 · B11 거절 |

## Branches and early returns

> 분기 표는 `analysis/harness/branch_table.py` 가 ast · 소스 · 커버리지로 만들었다. 「창의 return」은 위치, 「진입 실측」은 그 줄로 시작하는 커버리지 블록.

| Branch | 종류 | 조건 (원문) | 창의 return | 진입 실측 |
|---|---|---|---|---|
| B1 | if | `:363` `if !ok {` | :364 | 아니오 |
| B2 | if | `:366` `if err := validateProductionRiskJournalFile(config.JournalPath, owner); err != nil {` | :367 | 예 |
| B3 | if | `:375` `if err != nil {` | :376 | 아니오 |
| B4 | if | `:380` `if err := db.PingContext(ctx); err != nil {` | :381 | 아니오 |
| B5 | if | `:384` `if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil \|\| version != productionRiskJournalSchema {` | :385 | 아니오 |
| B6 | if | `:389` `if err := db.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND market=? AND symbol=?`,` | :391 | — |
| B7 | if | `:393` `if scopeLatches != 0 {` | :394 | 예 |
| B8 | if | `:403` `if authorityObserved.After(scope.AsOf) \|\| authorityFresh.Before(scope.AsOf) {` | :404 | 아니오 |
| B9 | range | `:407` `for _, dimension := range requiredDimensions {` | — | 예 |
| B10 | if | `:409` `if err != nil {` | :410 | 아니오 |
| B11 | if | `:413` `if usage.Latched {` | :414 | 예 |
| B12 | if | `:423` `if err != nil {` | :424 | 아니오 |
| B13 | if | `:433` `if err != nil {` | :434 | 아니오 |
| B14 | if | `:442` `if err != nil {` | :443, :447 | 아니오 |

## Calls and live bindings

`productionRiskOwnerUID` · `validateProductionRiskJournalFile` · `sql.Open`(mode=ro · query_only · busy_timeout) · `db.SetMaxOpenConns(1)` · `db.PingContext` · `db.QueryRowContext`(PRAGMA user_version · scope latch count) · `ReadJournalBucketUsage`(dimension 마다, `db`) · provenance 생성자 · `newRiskSnapshotAuthorityMaterialEntry`. 원장 쓰기 · 브로커 없음.

## State mutations and fallbacks

없음 — 읽기 전용 연결. 실패는 오류 반환(호출자가 Unavailable 로 감쌈).

## Safety conclusion

- **Safe edit boundary**: 편집 전 — a127 은 B5 의 비교를 주입 값 정확 일치 + 방향 문구로, 판독 전부를 읽기 전용 tx 하나로, 버전 확인 직후 · latch 판독 전에 SQL 상수 둘을 prepare 하도록 바꿀 예정(D1 · D7). 판독 SQL · 사용량 판정 불변.
- **High-risk impact**: yes — 사이징 사용량 판독.
