# Function Logic Map: `bindProductionRiskInputs`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `9a74db4abd523da823e02e716258d428d3f385647c3d18121e744f9fac45119e`
- Signature: `bindProductionRiskInputs(params=3, results=4)`
- Source range: `321:1`–`388:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 범위 국소 신원은 B4 하나 — 나머지 실패는 결함으로 남긴다(넓히지 않음).

## Branches and early returns

- Exact AST return nodes: `327:3, 331:3, 335:3, 340:3, 349:3, 353:3, 360:3, 378:3, 384:4, 387:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 323:2 | 봉인된 결과 불일치(무결성 — 결함) |
| B2 | if | 330:2 | 지원하지 않는 horizon |
| B3 | if | 334:2 | 전략 위험 매핑 없음(결함 — 범위 국소로 넓히지 않음) |
| B4 | if | 338:2 | 서명 정책에 그 종목 섹터 매핑 없음 → **`ErrProductionRiskScopeRefused`**(범위 국소) |
| B5 | if | 346:2 | 최악 체결가 권한 무효 |
| B6 | if | 352:2 | FX 결속 실패 |
| B7 | if | 359:2 | 정책 창 불일치 |
| B8 | if | 377:2 | 예약 정책 무효 |
| B9 | range | 382:2 | 차원 순회 |
| B10 | if | 383:3 | 한도 해석 실패 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `result.ValidProposal` | 323:6 |
| `lineage.Valid` | 323:121 |
| `terms.Valid` | 323:141 |
| `Market` | 324:45 |
| `terms.AccountRef` | 324:88 |
| `terms.Market` | 325:3 |
| `terms.Symbol` | 325:39 |
| `terms.Quantity` | 325:75 |
| `terms.LineageIdentity` | 326:3 |
| `strings.ToUpper` | 326:92 |
| `strings.TrimSpace` | 326:108 |
| `errors.New` | 327:53 |
| `Horizon` | 329:13 |
| `errors.New` | 331:53 |
| `exactProductionRiskStrategy` | 333:18 |
| `errors.New` | 335:53 |
| `exactProductionRiskSymbol` | 337:16 |
| `fmt.Errorf` | 340:53 |
| `terms.Entry` | 342:11 |
| `canonicalProductionRiskTime` | 343:28 |
| `entry.AsOf` | 343:56 |
| `entry.MajorDecimal` | 344:30 |
| `UTC` | 345:16 |
| `time.Unix` | 345:16 |
| `entry.Currency` | 346:34 |
| `entry.UnitVersion` | 346:76 |
| `entry.MinorScale` | 347:3 |
| `productionRiskMinorScale` | 347:25 |
| `entry.Source` | 347:68 |
| `entry.Version` | 347:92 |
| `entry.Digest` | 347:117 |
| `entryObserved.After` | 348:3 |
| `entryFresh.Before` | 348:45 |
| `errors.New` | 349:53 |
| `input.FX.EvidenceAt` | 351:13 |
| `canonicalProductionRiskTime` | 355:23 |
| `canonicalProductionRiskTime` | 356:20 |
| `latestProductionRiskTime` | 357:14 |
| `fx.ObservedAt` | 357:70 |
| `earliestProductionRiskTime` | 358:11 |
| `fx.FreshUntil` | 358:63 |
| `observed.After` | 359:5 |
| `fresh.Before` | 359:42 |
| `errors.New` | 360:53 |
| `productionRiskDigest` | 362:19 |
| `(unnamed)` | 362:40 |
| `strings.Join` | 362:47 |
| `terms.Identity` | 362:110 |
| `input.FX.Digest` | 362:128 |
| `config.ObservedAt.Format` | 363:3 |
| `strings.TrimPrefix` | 364:42 |
| `entry.Version` | 370:128 |
| `productionRiskDigest` | 371:14 |
| `(unnamed)` | 371:35 |
| `strings.Join` | 371:42 |
| `entry.Source` | 371:64 |
| `entry.Version` | 371:80 |
| `entry.Digest` | 371:97 |
| `entry.PriceMinor` | 371:113 |
| `terms.Identity` | 371:133 |
| `fx.RateQuoteToBase` | 373:35 |
| `fx.Haircut` | 373:66 |
| `fx.Source` | 373:107 |
| `fx.Version` | 373:129 |
| `fx.Digest` | 374:12 |
| `fx.ObservedAt` | 374:67 |
| `fx.FreshUntil` | 374:96 |
| `validateReservePolicy` | 377:30 |
| `parseMinor` | 383:16 |
| `fmt.Errorf` | 384:54 |

## State mutations and fallbacks

- 순수 함수 — 상태 없음.

## Safety conclusion

- High-risk. 판정 불변(같은 입력이 같은 자리에서 거절).

a112 6.1: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)

a112 게이트 준비(2026-10-04): a127 82080177(전략 권한 적재기가 현재 원장을 읽음)이 같은 파일을 편집해 이 번들이 낡았다 — 분기 · return · 호출 구조 동일
