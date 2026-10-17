# Function Logic Map (편집 전): `loadProductionRiskEntries`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `c5e0c64b2de6524804c4b0fbab5fb8596191379d764c68c319905ee92aa87ed2`
- Signature: `loadProductionRiskEntries(params=6, results=2)`
- Source range: `344:1`–`427:2`
- AST evidence: `ast.json` — 편집 **전**(a112 5.2.2.2 리뷰 수리 로트, J4 = (A) 판정 2026-10-01).

## Inputs and invariants

- 이 로트의 편집은 **오류 신원 배관**뿐이다 — 갈래 판정(어느 입력이 거절되는가)은 바꾸지 않는다(Manager 조건 ①).

## Branches and early returns

- Exact AST return nodes: `347:3, 350:3, 359:3, 364:3, 368:3, 373:3, 383:3, 389:4, 393:4, 403:4, 413:4, 422:4, 426:2`.

| Branch | AST kind | Source location | Condition (source line) | Edit |
|---|---|---|---|---|
| B1 | if | 346:2 | `if !ok {` | 판정 갈래 — 편집 안 함 |
| B2 | if | 349:2 | `if err := validateProductionRiskJournalFile(config.JournalPath, owner); err != nil {` | 판정 갈래 — 편집 안 함 |
| B3 | if | 358:2 | `if err != nil {` | 판정 갈래 — 편집 안 함 |
| B4 | if | 363:2 | `if err := db.PingContext(ctx); err != nil {` | 판정 갈래 — 편집 안 함 |
| B5 | if | 367:2 | `if err := db.QueryRowContext(ctx, "PRAGMA user_version").Scan(&version); err != nil \\|\\| version != productionRi` | 판정 갈래 — 편집 안 함 |
| B6 | if | 371:2 | `if err := db.QueryRowContext(ctx, `SELECT count(*) FROM risk_bucket_scope_latches WHERE account_ref=? AND mark` | **편집 대상** — scope latch 조회: `err != nil || scopeLatches != 0` 이 원장 조회 결함과 범위 국소 latch 를 한 오류(`scope latch present`)로 합침 |
| B7 | if | 382:2 | `if authorityObserved.After(scope.AsOf) \\|\\| authorityFresh.Before(scope.AsOf) {` | 판정 갈래 — 편집 안 함 |
| B8 | range | 386:2 | `for _, dimension := range requiredDimensions {` | 판정 갈래 — 편집 안 함 |
| B9 | if | 388:3 | `if err != nil {` | 판정 갈래 — 편집 안 함 |
| B10 | if | 392:3 | `if usage.Latched {` | 판정 갈래 — 편집 안 함 |
| B11 | if | 402:3 | `if err != nil {` | 판정 갈래 — 편집 안 함 |
| B12 | if | 412:3 | `if err != nil {` | 판정 갈래 — 편집 안 함 |
| B13 | if | 421:3 | `if err != nil {` | 판정 갈래 — 편집 안 함 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRiskOwnerUID` | 345:15 |
| `validateProductionRiskJournalFile` | 349:12 |
| `query.Set` | 353:2 |
| `query.Add` | 354:2 |
| `query.Add` | 355:2 |
| `query.Encode` | 356:69 |
| `sql.Open` | 357:13 |
| `dsn.String` | 357:32 |
| `db.Close` | 361:8 |
| `db.SetMaxOpenConns` | 362:2 |
| `db.PingContext` | 363:12 |
| `Scan` | 367:12 |
| `db.QueryRowContext` | 367:12 |
| `errors.New` | 368:15 |
| `Scan` | 371:12 |
| `db.QueryRowContext` | 371:12 |
| `string` | 372:20 |
| `errors.New` | 373:15 |
| `string` | 375:51 |
| `string` | 375:91 |
| `canonicalProductionRiskTime` | 377:25 |
| `canonicalProductionRiskTime` | 378:22 |
| `latestProductionRiskTime` | 379:23 |
| `earliestProductionRiskTime` | 380:20 |
| `scope.AsOf.Add` | 381:3 |
| `authorityObserved.After` | 382:5 |
| `authorityFresh.Before` | 382:44 |
| `errors.New` | 383:15 |
| `make` | 385:13 |
| `len` | 385:59 |
| `ReadJournalBucketUsage` | 387:17 |
| `errors.New` | 393:16 |
| `productionRiskDigest` | 397:19 |
| `(unnamed)` | 397:40 |
| `strings.Join` | 397:47 |
| `string` | 397:112 |
| `NewPolicyProvenance` | 401:28 |
| `productionRiskDigest` | 405:21 |
| `(unnamed)` | 405:42 |
| `strings.Join` | 405:49 |
| `string` | 405:94 |
| `scope.AsOf.Format` | 406:15 |
| `strings.TrimPrefix` | 407:44 |
| `NewSnapshotProvenance` | 411:30 |
| `newRiskSnapshotAuthorityMaterialEntry` | 420:17 |
| `append` | 424:13 |

## State mutations and fallbacks

- 원장은 읽기 전용(`mode=ro`, `query_only`). 쓰기 없음.

## Safety conclusion

- High-risk(위험 권한). 편집 전 모든 실패는 `ErrProductionRiskSnapshotUnavailable` 하나로 접힌다 — 엔진이 범위 국소 거절과 원장 결함을 구별할 수 없는 근본 원인.
