# Function Logic Map: `loadProductionRiskEntries`

- Source: `internal/riskbucket/production_snapshot_authority.go` (`390`–`500`)
- Qualified: `loadProductionRiskEntries`
- AST evidence: `ast.json` (`source_sha256` 9a74db4abd523da8…) — **편집 뒤**(구현 로트, 커버리지 `analysis/impl/coverage-post-edit.out`)
- Risk scan: `risk-pattern-report.md`
- AST branches 19 · return 18 · 호출 54

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
| B1 | if | `:392` `if !ok {` | :393 | 아니오 |
| B2 | if | `:395` `if err := validateProductionRiskJournalFile(config.JournalPath, owner); err != nil {` | :396 | 예 |
| B3 | if | `:404` `if err != nil {` | :405 | 아니오 |
| B4 | if | `:409` `if err := db.PingContext(ctx); err != nil {` | :410 | 아니오 |
| B5 | if | `:415` `if err != nil {` | :416 | 아니오 |
| B6 | if | `:422` `if err := tx.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil {` | :423 | 아니오 |
| B7 | if | `:425` `if version > config.JournalSchemaVersion {` | :426 | 예 |
| B8 | if | `:428` `if version < config.JournalSchemaVersion {` | :429 | 예 |
| B9 | range | `:433` `for _, statement := range []string{productionRiskScopeLatchSQL, productionRiskUsageSQL} {` | — | 예 |
| B10 | if | `:435` `if err != nil {` | :436 | 예 |
| B11 | if | `:442` `if err := tx.QueryRowContext(ctx, productionRiskScopeLatchSQL, scope.AccountID, string(scope.Market), scope.Symbol).Scan(&scopeLatches); …` | :443 | 아니오 |
| B12 | if | `:445` `if scopeLatches != 0 {` | :446 | 예 |
| B13 | if | `:455` `if authorityObserved.After(scope.AsOf) \|\| authorityFresh.Before(scope.AsOf) {` | :456 | 아니오 |
| B14 | range | `:459` `for _, dimension := range requiredDimensions {` | — | 예 |
| B15 | if | `:461` `if err != nil {` | :462 | 아니오 |
| B16 | if | `:465` `if usage.Latched {` | :466 | 예 |
| B17 | if | `:475` `if err != nil {` | :476 | 아니오 |
| B18 | if | `:485` `if err != nil {` | :486 | 아니오 |
| B19 | if | `:494` `if err != nil {` | :495, :499 | 아니오 |

## Calls and live bindings

`productionRiskOwnerUID` · `validateProductionRiskJournalFile` · `sql.Open`(mode=ro · query_only · busy_timeout) · `db.SetMaxOpenConns(1)` · `db.PingContext` · `db.BeginTx`(ReadOnly) · `tx.Rollback`(defer) · `tx.QueryRowContext`(PRAGMA user_version · `productionRiskScopeLatchSQL`) · `tx.PrepareContext`(`productionRiskScopeLatchSQL` · `productionRiskUsageSQL`) · `ReadJournalBucketUsage`(dimension 마다, `tx`) · provenance 생성자 · `newRiskSnapshotAuthorityMaterialEntry`. 원장 쓰기 · 브로커 없음(1.6 리뷰 P3-5 로 편집 뒤 호출 갱신).

## State mutations and fallbacks

없음 — 읽기 전용 연결. 실패는 오류 반환(호출자가 Unavailable 로 감쌈).

## Safety conclusion

- **Safe edit boundary**: B5 비교 · 문구, tx 묶음, prepare 선행만 바뀜. scope latch · 창 · 사용량 판정과 오류 신원 불변.
- **High-risk impact**: yes — 사이징 사용량 판독.
