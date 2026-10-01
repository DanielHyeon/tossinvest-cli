# Function Logic Map: `bindProductionRiskInputs`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `38de0b7d846b0a1af1b01bc24eb94adc673ca953a3f125f3130adbc3bb4f58c2`
- Signature: `bindProductionRiskInputs(params=3, results=4)`
- Source range: `283:1`–`350:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 범위 국소 신원은 B4 하나 — 나머지 실패는 결함으로 남긴다(넓히지 않음).

## Branches and early returns

- Exact AST return nodes: `289:3, 293:3, 297:3, 302:3, 311:3, 315:3, 322:3, 340:3, 346:4, 349:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 285:2 | 봉인된 결과 불일치(무결성 — 결함) |
| B2 | if | 292:2 | 지원하지 않는 horizon |
| B3 | if | 296:2 | 전략 위험 매핑 없음(결함 — 범위 국소로 넓히지 않음) |
| B4 | if | 300:2 | 서명 정책에 그 종목 섹터 매핑 없음 → **`ErrProductionRiskScopeRefused`**(범위 국소) |
| B5 | if | 308:2 | 최악 체결가 권한 무효 |
| B6 | if | 314:2 | FX 결속 실패 |
| B7 | if | 321:2 | 정책 창 불일치 |
| B8 | if | 339:2 | 예약 정책 무효 |
| B9 | range | 344:2 | 차원 순회 |
| B10 | if | 345:3 | 한도 해석 실패 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `result.ValidProposal` | 285:6 |
| `lineage.Valid` | 285:121 |
| `terms.Valid` | 285:141 |
| `Market` | 286:45 |
| `terms.AccountRef` | 286:88 |
| `terms.Market` | 287:3 |
| `terms.Symbol` | 287:39 |
| `terms.Quantity` | 287:75 |
| `terms.LineageIdentity` | 288:3 |
| `strings.ToUpper` | 288:92 |
| `strings.TrimSpace` | 288:108 |
| `errors.New` | 289:53 |
| `Horizon` | 291:13 |
| `errors.New` | 293:53 |
| `exactProductionRiskStrategy` | 295:18 |
| `errors.New` | 297:53 |
| `exactProductionRiskSymbol` | 299:16 |
| `fmt.Errorf` | 302:53 |
| `terms.Entry` | 304:11 |
| `canonicalProductionRiskTime` | 305:28 |
| `entry.AsOf` | 305:56 |
| `entry.MajorDecimal` | 306:30 |
| `UTC` | 307:16 |
| `time.Unix` | 307:16 |
| `entry.Currency` | 308:34 |
| `entry.UnitVersion` | 308:76 |
| `entry.MinorScale` | 309:3 |
| `productionRiskMinorScale` | 309:25 |
| `entry.Source` | 309:68 |
| `entry.Version` | 309:92 |
| `entry.Digest` | 309:117 |
| `entryObserved.After` | 310:3 |
| `entryFresh.Before` | 310:45 |
| `errors.New` | 311:53 |
| `input.FX.EvidenceAt` | 313:13 |
| `canonicalProductionRiskTime` | 317:23 |
| `canonicalProductionRiskTime` | 318:20 |
| `latestProductionRiskTime` | 319:14 |
| `fx.ObservedAt` | 319:70 |
| `earliestProductionRiskTime` | 320:11 |
| `fx.FreshUntil` | 320:63 |
| `observed.After` | 321:5 |
| `fresh.Before` | 321:42 |
| `errors.New` | 322:53 |
| `productionRiskDigest` | 324:19 |
| `(unnamed)` | 324:40 |
| `strings.Join` | 324:47 |
| `terms.Identity` | 324:110 |
| `input.FX.Digest` | 324:128 |
| `config.ObservedAt.Format` | 325:3 |
| `strings.TrimPrefix` | 326:42 |
| `entry.Version` | 332:128 |
| `productionRiskDigest` | 333:14 |
| `(unnamed)` | 333:35 |
| `strings.Join` | 333:42 |
| `entry.Source` | 333:64 |
| `entry.Version` | 333:80 |
| `entry.Digest` | 333:97 |
| `entry.PriceMinor` | 333:113 |
| `terms.Identity` | 333:133 |
| `fx.RateQuoteToBase` | 335:35 |
| `fx.Haircut` | 335:66 |
| `fx.Source` | 335:107 |
| `fx.Version` | 335:129 |
| `fx.Digest` | 336:12 |
| `fx.ObservedAt` | 336:67 |
| `fx.FreshUntil` | 336:96 |
| `validateReservePolicy` | 339:30 |
| `parseMinor` | 345:16 |
| `fmt.Errorf` | 346:54 |

## State mutations and fallbacks

- 순수 함수 — 상태 없음.

## Safety conclusion

- High-risk. 판정 불변(같은 입력이 같은 자리에서 거절).
