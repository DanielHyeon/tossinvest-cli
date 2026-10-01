# Function Logic Map: `loadProductionRiskEntries`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `3aa9b66c00cdcdedf09e0bea1b0eeeaf56d46d0ba40149f28edf70e73d7e26b4`
- Signature: `loadProductionRiskEntries(params=6, results=2)`
- Source range: `361:1`–`448:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 판정 무변 — 편집 전 B6 이 거절하던 입력은 B6 또는 B7 이 같은 결론(거절)으로 거절한다.

## Branches and early returns

- Exact AST return nodes: `364:3, 367:3, 376:3, 381:3, 385:3, 391:3, 394:3, 404:3, 410:4, 414:4, 424:4, 434:4, 443:4, 447:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 363:2 | 소유자 UID 없음 |
| B2 | if | 366:2 | 원장 파일 검증 실패 |
| B3 | if | 375:2 | 열기 실패 |
| B4 | if | 380:2 | ping 실패 |
| B5 | if | 384:2 | 스키마 핀 불일치(R1 — 별도 change) |
| B6 | if | 389:2 | **(분리)** latch 조회 결함 → `scope latch unreadable: %w`(결함) |
| B7 | if | 393:2 | **(분리)** 그 범위에 scope latch 있음 → `ErrProductionRiskScopeRefused`(범위 국소) |
| B8 | if | 403:2 | 권한 창 불일치 |
| B9 | range | 407:2 | 차원 순회 |
| B10 | if | 409:3 | 사용량 읽기 실패(원장 결함 — 손상 행 포함) |
| B11 | if | 413:3 | latch 된 사용량 |
| B12 | if | 423:3 | 정책 출처 구성 실패 |
| B13 | if | 433:3 | 스냅숏 출처 구성 실패 |
| B14 | if | 442:3 | 권한 항목 구성 실패 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `productionRiskOwnerUID` | 362:15 |
| `validateProductionRiskJournalFile` | 366:12 |
| `query.Set` | 370:2 |
| `query.Add` | 371:2 |
| `query.Add` | 372:2 |
| `query.Encode` | 373:69 |
| `sql.Open` | 374:13 |
| `dsn.String` | 374:32 |
| `db.Close` | 378:8 |
| `db.SetMaxOpenConns` | 379:2 |
| `db.PingContext` | 380:12 |
| `Scan` | 384:12 |
| `db.QueryRowContext` | 384:12 |
| `errors.New` | 385:15 |
| `Scan` | 389:12 |
| `db.QueryRowContext` | 389:12 |
| `string` | 390:20 |
| `fmt.Errorf` | 391:15 |
| `fmt.Errorf` | 394:15 |
| `string` | 396:51 |
| `string` | 396:91 |
| `canonicalProductionRiskTime` | 398:25 |
| `canonicalProductionRiskTime` | 399:22 |
| `latestProductionRiskTime` | 400:23 |
| `earliestProductionRiskTime` | 401:20 |
| `scope.AsOf.Add` | 402:3 |
| `authorityObserved.After` | 403:5 |
| `authorityFresh.Before` | 403:44 |
| `errors.New` | 404:15 |
| `make` | 406:13 |
| `len` | 406:59 |
| `ReadJournalBucketUsage` | 408:17 |
| `errors.New` | 414:16 |
| `productionRiskDigest` | 418:19 |
| `(unnamed)` | 418:40 |
| `strings.Join` | 418:47 |
| `string` | 418:112 |
| `NewPolicyProvenance` | 422:28 |
| `productionRiskDigest` | 426:21 |
| `(unnamed)` | 426:42 |
| `strings.Join` | 426:49 |
| `string` | 426:94 |
| `scope.AsOf.Format` | 427:15 |
| `strings.TrimPrefix` | 428:44 |
| `NewSnapshotProvenance` | 432:30 |
| `newRiskSnapshotAuthorityMaterialEntry` | 441:17 |
| `append` | 445:13 |

## State mutations and fallbacks

- 원장 읽기 전용(`mode=ro` · `query_only`), 쓰기 없음.

## Safety conclusion

- High-risk. 원장 읽기 전용. 새로 통과 · 새로 거절 0.

a112 6.1: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
