# Function Logic Map (편집 전): `bindProductionRiskInputs`

- Source: `internal/riskbucket/production_snapshot_authority.go`
- Source SHA-256: `c5e0c64b2de6524804c4b0fbab5fb8596191379d764c68c319905ee92aa87ed2`
- Signature: `bindProductionRiskInputs(params=3, results=4)`
- Source range: `276:1`–`342:2`
- AST evidence: `ast.json` — 편집 **전**(a112 5.2.2.2 리뷰 수리 로트, J4 = (A) 판정 2026-10-01).

## Inputs and invariants

- 이 로트의 편집은 **오류 신원 배관**뿐이다 — 갈래 판정(어느 입력이 거절되는가)은 바꾸지 않는다(Manager 조건 ①).

## Branches and early returns

- Exact AST return nodes: `282:3, 286:3, 290:3, 294:3, 303:3, 307:3, 314:3, 332:3, 338:4, 341:2`.

| Branch | AST kind | Source location | Condition (source line) | Edit |
|---|---|---|---|---|
| B1 | if | 278:2 | `if !result.ValidProposal() \\|\\| result.Code != strategyflow.RefusalNone \\|\\| result.Quantity == 0 \\|\\| !lineage.Comp` | 판정 갈래 — 편집 안 함 |
| B2 | if | 285:2 | `if horizon != HorizonShort && horizon != HorizonMedium {` | 판정 갈래 — 편집 안 함 |
| B3 | if | 289:2 | `if !ok {` | 판정 갈래 — 편집 안 함 |
| B4 | if | 293:2 | `if !ok {` | **편집 대상** — 서명 정책에 그 종목의 섹터 매핑이 없음(`symbol sector mapping unavailable`) — 범위 국소 사유(그 종목만 정책 밖)인데 결함과 같은 평문 오류 |
| B5 | if | 300:2 | `if !entryOK \\|\\| !entryMajorOK \\|\\| entry.Currency() != body.QuoteCurrency \\|\\| entry.UnitVersion() != "minor-v1" \\|\\|` | 판정 갈래 — 편집 안 함 |
| B6 | if | 306:2 | `if err != nil {` | 판정 갈래 — 편집 안 함 |
| B7 | if | 313:2 | `if observed.After(config.ObservedAt) \\|\\| fresh.Before(config.ObservedAt) {` | 판정 갈래 — 편집 안 함 |
| B8 | if | 331:2 | `if _, _, _, _, _, _, err := validateReservePolicy(reserve); err != nil {` | 판정 갈래 — 편집 안 함 |
| B9 | range | 336:2 | `for _, dimension := range requiredDimensions {` | 판정 갈래 — 편집 안 함 |
| B10 | if | 337:3 | `if _, err := parseMinor(limits[dimension], reserve.MaxDecimalBits); err != nil {` | 판정 갈래 — 편집 안 함 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `result.ValidProposal` | 278:6 |
| `lineage.Valid` | 278:121 |
| `terms.Valid` | 278:141 |
| `Market` | 279:45 |
| `terms.AccountRef` | 279:88 |
| `terms.Market` | 280:3 |
| `terms.Symbol` | 280:39 |
| `terms.Quantity` | 280:75 |
| `terms.LineageIdentity` | 281:3 |
| `strings.ToUpper` | 281:92 |
| `strings.TrimSpace` | 281:108 |
| `errors.New` | 282:53 |
| `Horizon` | 284:13 |
| `errors.New` | 286:53 |
| `exactProductionRiskStrategy` | 288:18 |
| `errors.New` | 290:53 |
| `exactProductionRiskSymbol` | 292:16 |
| `errors.New` | 294:53 |
| `terms.Entry` | 296:11 |
| `canonicalProductionRiskTime` | 297:28 |
| `entry.AsOf` | 297:56 |
| `entry.MajorDecimal` | 298:30 |
| `UTC` | 299:16 |
| `time.Unix` | 299:16 |
| `entry.Currency` | 300:34 |
| `entry.UnitVersion` | 300:76 |
| `entry.MinorScale` | 301:3 |
| `productionRiskMinorScale` | 301:25 |
| `entry.Source` | 301:68 |
| `entry.Version` | 301:92 |
| `entry.Digest` | 301:117 |
| `entryObserved.After` | 302:3 |
| `entryFresh.Before` | 302:45 |
| `errors.New` | 303:53 |
| `input.FX.EvidenceAt` | 305:13 |
| `canonicalProductionRiskTime` | 309:23 |
| `canonicalProductionRiskTime` | 310:20 |
| `latestProductionRiskTime` | 311:14 |
| `fx.ObservedAt` | 311:70 |
| `earliestProductionRiskTime` | 312:11 |
| `fx.FreshUntil` | 312:63 |
| `observed.After` | 313:5 |
| `fresh.Before` | 313:42 |
| `errors.New` | 314:53 |
| `productionRiskDigest` | 316:19 |
| `(unnamed)` | 316:40 |
| `strings.Join` | 316:47 |
| `terms.Identity` | 316:110 |
| `input.FX.Digest` | 316:128 |
| `config.ObservedAt.Format` | 317:3 |
| `strings.TrimPrefix` | 318:42 |
| `entry.Version` | 324:128 |
| `productionRiskDigest` | 325:14 |
| `(unnamed)` | 325:35 |
| `strings.Join` | 325:42 |
| `entry.Source` | 325:64 |
| `entry.Version` | 325:80 |
| `entry.Digest` | 325:97 |
| `entry.PriceMinor` | 325:113 |
| `terms.Identity` | 325:133 |
| `fx.RateQuoteToBase` | 327:35 |
| `fx.Haircut` | 327:66 |
| `fx.Source` | 327:107 |
| `fx.Version` | 327:129 |
| `fx.Digest` | 328:12 |
| `fx.ObservedAt` | 328:67 |
| `fx.FreshUntil` | 328:96 |
| `validateReservePolicy` | 331:30 |
| `parseMinor` | 337:16 |
| `fmt.Errorf` | 338:54 |

## State mutations and fallbacks

- 원장은 읽기 전용(`mode=ro`, `query_only`). 쓰기 없음.

## Safety conclusion

- High-risk(위험 권한). 편집 전 모든 실패는 `ErrProductionRiskSnapshotUnavailable` 하나로 접힌다 — 엔진이 범위 국소 거절과 원장 결함을 구별할 수 없는 근본 원인.
