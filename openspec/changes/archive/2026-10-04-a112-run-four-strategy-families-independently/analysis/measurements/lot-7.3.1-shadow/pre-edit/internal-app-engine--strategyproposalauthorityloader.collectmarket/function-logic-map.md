# Function Logic Map (편집 전): `collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `2c546898bc4178d0ee44238d51fb8e05a4c908cee721041ee24713c4c98ac385`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- Source range: `293:1`–`458:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD 9e5f3ccf).

## Inputs and invariants

- 편집 계획: 브리프 §4: coordinateMarketProposals 의 shadow 묶음을 authority 밖 별도 값으로 받아 (strategyProposalMarketAuthority, strategyShadowBatch) 로 반환; 조정 앞 닫힘 일곱은 부재 값, 조정 뒤 닫힘 다섯 · 성공은 수집 묶음.

## Branches and early returns

- Exact AST return nodes: `312:3`, `319:3`, `330:3`, `333:3`, `338:3`, `350:4`, `362:3`, `371:3`, `407:3`, `415:3`, `424:3`, `433:3`, `442:3`, `449:3`, `453:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 316:2 | `if !routes.snapshot.Ready // len(routes.entries) == 0 // !schedule.snapshot.Ready // schedule.restore.Activation == nil {` |
| B2 | if | 329:2 | `if !fx.snapshot.Ready // !fx.read.valid {` |
| B3 | if | 332:2 | `if loader.getenv == nil // loader.load == nil // loader.configDir == "." // loader.evidencePath == "." // loader.journalPath == "." // loader.accountRef == "" {` |
| B4 | if | 337:2 | `if err != nil // base64.StdEncoding.EncodeToString(key) != encoded // len(key) != ed25519.PublicKeySize {` |
| B5 | if | 341:2 | `if market == StrategyMarketUS {` |
| B6 | range | 347:2 | `for _, entry := range routes.entries {` |
| B7 | if | 349:3 | `if symbol == "" // bySymbol[symbol].approved.Valid() {` |
| B8 | if | 361:2 | `if err != nil // batch.ManifestDigest() != digest {` |
| B9 | if | 366:2 | `if absence, lost := batch.Fault(); lost {` |
| B10 | if | 402:2 | `if arbitration.erasedScopes != 0 {` |
| B11 | if | 409:2 | `if arbitration.collision {` |
| B12 | if | 418:2 | `if outcome.Overflow {` |
| B13 | if | 426:2 | `if outcome.Refusal != strategyarbiter.RefusalNone {` |
| B14 | if | 436:2 | `if !resolved {` |
| B15 | if | 444:2 | `if len(entries) == 0 {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 313:90 |
| `len` | 316:31 |
| `fail` | 319:10 |
| `loader.familyGateFor` | 328:9 |
| `fail` | 330:10 |
| `fail` | 333:10 |
| `strings.TrimSpace` | 335:13 |
| `loader.getenv` | 335:31 |
| `DecodeString` | 336:14 |
| `base64.StdEncoding.Strict` | 336:14 |
| `base64.StdEncoding.EncodeToString` | 337:19 |
| `len` | 337:72 |
| `fail` | 338:10 |
| `strings.TrimSpace` | 344:12 |
| `loader.getenv` | 344:30 |
| `make` | 345:13 |
| `len` | 345:58 |
| `make` | 346:14 |
| `len` | 346:59 |
| `entry.approved.Symbol` | 348:13 |
| `bySymbol.approved.Valid` | 349:22 |
| `fail` | 350:11 |
| `append` | 353:13 |
| `entry.route.Request` | 353:97 |
| `loader.load` | 355:16 |
| `strategyrouter.Market` | 356:42 |
| `strings.TrimSpace` | 356:111 |
| `loader.getenv` | 356:129 |
| `ed25519.PublicKey` | 357:15 |
| `strings.TrimSpace` | 360:23 |
| `loader.getenv` | 360:41 |
| `batch.ManifestDigest` | 361:19 |
| `fail` | 362:10 |
| `batch.Fault` | 366:22 |
| `fail` | 367:13 |
| `absence.String` | 369:37 |
| `loader.familyGateFor` | 390:9 |
| `coordinateMarketProposals` | 391:26 |
| `len` | 392:30 |
| `distinctGatedOutcomes` | 392:54 |
| `fail` | 403:13 |
| `fail` | 410:13 |
| `fail` | 419:13 |
| `fail` | 427:13 |
| `string` | 429:40 |
| `arbitration.entries` | 435:23 |
| `fail` | 437:13 |
| `len` | 444:5 |
| `fail` | 445:13 |
| `len` | 455:17 |
| `len` | 455:53 |
| `strategyProposalSetDigest` | 456:23 |

## Safety conclusion

- High-risk(제안 권한) — authority 값 · 닫힘 사유 · 반환 authority 무변경, shadow 값은 authority 구조체에 들어가지 않음(census ②).
