# Function Logic Map: `loadProductionRiskEntries`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `38de0b7d846b0a1af1b01bc24eb94adc673ca953a3f125f3130adbc3bb4f58c2`
- Signature: `loadProductionRiskEntries(params=6, results=2)`
- Source range: `352:1`–`439:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 판정 무변 — 편집 전 B6 이 거절하던 입력은 B6 또는 B7 이 같은 결론(거절)으로 거절한다.

## Branches and early returns

- Exact AST return nodes: `355:3, 358:3, 367:3, 372:3, 376:3, 382:3, 385:3, 395:3, 401:4, 405:4, 415:4, 425:4, 434:4, 438:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 354:2 | 소유자 UID 없음 |
| B2 | if | 357:2 | 원장 파일 검증 실패 |
| B3 | if | 366:2 | 열기 실패 |
| B4 | if | 371:2 | ping 실패 |
| B5 | if | 375:2 | 스키마 핀 불일치(R1 — 별도 change) |
| B6 | if | 380:2 | **(분리)** latch 조회 결함 → `scope latch unreadable: %w`(결함) |
| B7 | if | 384:2 | **(분리)** 그 범위에 scope latch 있음 → `ErrProductionRiskScopeRefused`(범위 국소) |
| B8 | if | 394:2 | 권한 창 불일치 |
| B9 | range | 398:2 | 차원 순회 |
| B10 | if | 400:3 | 사용량 읽기 실패(원장 결함 — 손상 행 포함) |
| B11 | if | 404:3 | latch 된 사용량 |
| B12 | if | 414:3 | 정책 출처 구성 실패 |
| B13 | if | 424:3 | 스냅숏 출처 구성 실패 |
| B14 | if | 433:3 | 권한 항목 구성 실패 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRiskOwnerUID` | 353:15 |
| `validateProductionRiskJournalFile` | 357:12 |
| `query.Set` | 361:2 |
| `query.Add` | 362:2 |
| `query.Add` | 363:2 |
| `query.Encode` | 364:69 |
| `sql.Open` | 365:13 |
| `dsn.String` | 365:32 |
| `db.Close` | 369:8 |
| `db.SetMaxOpenConns` | 370:2 |
| `db.PingContext` | 371:12 |
| `Scan` | 375:12 |
| `db.QueryRowContext` | 375:12 |
| `errors.New` | 376:15 |
| `Scan` | 380:12 |
| `db.QueryRowContext` | 380:12 |
| `string` | 381:20 |
| `fmt.Errorf` | 382:15 |
| `fmt.Errorf` | 385:15 |
| `string` | 387:51 |
| `string` | 387:91 |
| `canonicalProductionRiskTime` | 389:25 |
| `canonicalProductionRiskTime` | 390:22 |
| `latestProductionRiskTime` | 391:23 |
| `earliestProductionRiskTime` | 392:20 |
| `scope.AsOf.Add` | 393:3 |
| `authorityObserved.After` | 394:5 |
| `authorityFresh.Before` | 394:44 |
| `errors.New` | 395:15 |
| `make` | 397:13 |
| `len` | 397:59 |
| `ReadJournalBucketUsage` | 399:17 |
| `errors.New` | 405:16 |
| `productionRiskDigest` | 409:19 |
| `(unnamed)` | 409:40 |
| `strings.Join` | 409:47 |
| `string` | 409:112 |
| `NewPolicyProvenance` | 413:28 |
| `productionRiskDigest` | 417:21 |
| `(unnamed)` | 417:42 |
| `strings.Join` | 417:49 |
| `string` | 417:94 |
| `scope.AsOf.Format` | 418:15 |
| `strings.TrimPrefix` | 419:44 |
| `NewSnapshotProvenance` | 423:30 |
| `newRiskSnapshotAuthorityMaterialEntry` | 432:17 |
| `append` | 436:13 |

## State mutations and fallbacks

- 원장 읽기 전용(`mode=ro` · `query_only`), 쓰기 없음.

## Safety conclusion

- High-risk. 원장 읽기 전용. 새로 통과 · 새로 거절 0.
