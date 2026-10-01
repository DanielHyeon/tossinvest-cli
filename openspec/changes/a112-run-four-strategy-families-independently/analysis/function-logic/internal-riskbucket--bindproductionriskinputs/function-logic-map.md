# Function Logic Map: `bindProductionRiskInputs`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `3aa9b66c00cdcdedf09e0bea1b0eeeaf56d46d0ba40149f28edf70e73d7e26b4`
- Signature: `bindProductionRiskInputs(params=3, results=4)`
- Source range: `292:1`–`359:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 범위 국소 신원은 B4 하나 — 나머지 실패는 결함으로 남긴다(넓히지 않음).

## Branches and early returns

- Exact AST return nodes: `298:3, 302:3, 306:3, 311:3, 320:3, 324:3, 331:3, 349:3, 355:4, 358:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 294:2 | 봉인된 결과 불일치(무결성 — 결함) |
| B2 | if | 301:2 | 지원하지 않는 horizon |
| B3 | if | 305:2 | 전략 위험 매핑 없음(결함 — 범위 국소로 넓히지 않음) |
| B4 | if | 309:2 | 서명 정책에 그 종목 섹터 매핑 없음 → **`ErrProductionRiskScopeRefused`**(범위 국소) |
| B5 | if | 317:2 | 최악 체결가 권한 무효 |
| B6 | if | 323:2 | FX 결속 실패 |
| B7 | if | 330:2 | 정책 창 불일치 |
| B8 | if | 348:2 | 예약 정책 무효 |
| B9 | range | 353:2 | 차원 순회 |
| B10 | if | 354:3 | 한도 해석 실패 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `result.ValidProposal` | 294:6 |
| `lineage.Valid` | 294:121 |
| `terms.Valid` | 294:141 |
| `Market` | 295:45 |
| `terms.AccountRef` | 295:88 |
| `terms.Market` | 296:3 |
| `terms.Symbol` | 296:39 |
| `terms.Quantity` | 296:75 |
| `terms.LineageIdentity` | 297:3 |
| `strings.ToUpper` | 297:92 |
| `strings.TrimSpace` | 297:108 |
| `errors.New` | 298:53 |
| `Horizon` | 300:13 |
| `errors.New` | 302:53 |
| `exactProductionRiskStrategy` | 304:18 |
| `errors.New` | 306:53 |
| `exactProductionRiskSymbol` | 308:16 |
| `fmt.Errorf` | 311:53 |
| `terms.Entry` | 313:11 |
| `canonicalProductionRiskTime` | 314:28 |
| `entry.AsOf` | 314:56 |
| `entry.MajorDecimal` | 315:30 |
| `UTC` | 316:16 |
| `time.Unix` | 316:16 |
| `entry.Currency` | 317:34 |
| `entry.UnitVersion` | 317:76 |
| `entry.MinorScale` | 318:3 |
| `productionRiskMinorScale` | 318:25 |
| `entry.Source` | 318:68 |
| `entry.Version` | 318:92 |
| `entry.Digest` | 318:117 |
| `entryObserved.After` | 319:3 |
| `entryFresh.Before` | 319:45 |
| `errors.New` | 320:53 |
| `input.FX.EvidenceAt` | 322:13 |
| `canonicalProductionRiskTime` | 326:23 |
| `canonicalProductionRiskTime` | 327:20 |
| `latestProductionRiskTime` | 328:14 |
| `fx.ObservedAt` | 328:70 |
| `earliestProductionRiskTime` | 329:11 |
| `fx.FreshUntil` | 329:63 |
| `observed.After` | 330:5 |
| `fresh.Before` | 330:42 |
| `errors.New` | 331:53 |
| `productionRiskDigest` | 333:19 |
| `(unnamed)` | 333:40 |
| `strings.Join` | 333:47 |
| `terms.Identity` | 333:110 |
| `input.FX.Digest` | 333:128 |
| `config.ObservedAt.Format` | 334:3 |
| `strings.TrimPrefix` | 335:42 |
| `entry.Version` | 341:128 |
| `productionRiskDigest` | 342:14 |
| `(unnamed)` | 342:35 |
| `strings.Join` | 342:42 |
| `entry.Source` | 342:64 |
| `entry.Version` | 342:80 |
| `entry.Digest` | 342:97 |
| `entry.PriceMinor` | 342:113 |
| `terms.Identity` | 342:133 |
| `fx.RateQuoteToBase` | 344:35 |
| `fx.Haircut` | 344:66 |
| `fx.Source` | 344:107 |
| `fx.Version` | 344:129 |
| `fx.Digest` | 345:12 |
| `fx.ObservedAt` | 345:67 |
| `fx.FreshUntil` | 345:96 |
| `validateReservePolicy` | 348:30 |
| `parseMinor` | 354:16 |
| `fmt.Errorf` | 355:54 |

## State mutations and fallbacks

- 순수 함수 — 상태 없음.

## Safety conclusion

- High-risk. 판정 불변(같은 입력이 같은 자리에서 거절).

a112 6.1: 같은 파일의 다른 함수 편집으로 줄만 이동(본문 불변)
