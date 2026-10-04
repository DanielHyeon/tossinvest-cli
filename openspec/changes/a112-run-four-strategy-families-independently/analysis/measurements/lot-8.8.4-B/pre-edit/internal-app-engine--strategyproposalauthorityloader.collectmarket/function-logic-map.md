# Function Logic Map (편집 전): `collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `60e0ef9270e3102bb69ce930cdffb41a5c4653dca69efdc94d32374823161bae`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- Source range: `293:1`–`439:2`
- AST evidence: `ast.json` — 편집 **전**(HEAD f473d815).

## Inputs and invariants

- 편집 계획: a112 8.8.4 로트 B 항목 2(Q-B2): gate 계산(familyGateFor)만 RouteNotReady 가드 직후로 옮긴다 — 실패 kind 판정 자리 · 우선순위 · FamilyGateClosed 판정 자리 불변, 옮긴 값은 닫힘 갈래가 싣는 활성화로만 쓰인다. RouteNotReady 는 zero-carry 유지.

## Branches and early returns

- Exact AST return nodes: `308:3`, `313:3`, `316:3`, `319:3`, `324:3`, `336:4`, `348:3`, `357:3`, `388:3`, `396:3`, `405:3`, `414:3`, `423:3`, `430:3`, `434:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 312:2 | `if !routes.snapshot.Ready // len(routes.entries) == 0 // !schedule.snapshot.Ready // schedule.restore.Activation == nil {` |
| B2 | if | 315:2 | `if !fx.snapshot.Ready // !fx.read.valid {` |
| B3 | if | 318:2 | `if loader.getenv == nil // loader.load == nil // loader.configDir == "." // loader.evidencePath == "." // loader.journalPath == "." // loader.accountRef == "" {` |
| B4 | if | 323:2 | `if err != nil // base64.StdEncoding.EncodeToString(key) != encoded // len(key) != ed25519.PublicKeySize {` |
| B5 | if | 327:2 | `if market == StrategyMarketUS {` |
| B6 | range | 333:2 | `for _, entry := range routes.entries {` |
| B7 | if | 335:3 | `if symbol == "" // bySymbol[symbol].approved.Valid() {` |
| B8 | if | 347:2 | `if err != nil // batch.ManifestDigest() != digest {` |
| B9 | if | 352:2 | `if absence, lost := batch.Fault(); lost {` |
| B10 | if | 383:2 | `if arbitration.erasedScopes != 0 {` |
| B11 | if | 390:2 | `if arbitration.collision {` |
| B12 | if | 399:2 | `if outcome.Overflow {` |
| B13 | if | 407:2 | `if outcome.Refusal != strategyarbiter.RefusalNone {` |
| B14 | if | 417:2 | `if !resolved {` |
| B15 | if | 425:2 | `if len(entries) == 0 {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 309:90 |
| `len` | 312:31 |
| `fail` | 313:10 |
| `fail` | 316:10 |
| `fail` | 319:10 |
| `strings.TrimSpace` | 321:13 |
| `loader.getenv` | 321:31 |
| `DecodeString` | 322:14 |
| `base64.StdEncoding.Strict` | 322:14 |
| `base64.StdEncoding.EncodeToString` | 323:19 |
| `len` | 323:72 |
| `fail` | 324:10 |
| `strings.TrimSpace` | 330:12 |
| `loader.getenv` | 330:30 |
| `make` | 331:13 |
| `len` | 331:58 |
| `make` | 332:14 |
| `len` | 332:59 |
| `entry.approved.Symbol` | 334:13 |
| `bySymbol.approved.Valid` | 335:22 |
| `fail` | 336:11 |
| `append` | 339:13 |
| `entry.route.Request` | 339:97 |
| `loader.load` | 341:16 |
| `strategyrouter.Market` | 342:42 |
| `strings.TrimSpace` | 342:111 |
| `loader.getenv` | 342:129 |
| `ed25519.PublicKey` | 343:15 |
| `strings.TrimSpace` | 346:23 |
| `loader.getenv` | 346:41 |
| `batch.ManifestDigest` | 347:19 |
| `fail` | 348:10 |
| `batch.Fault` | 352:22 |
| `fail` | 353:13 |
| `absence.String` | 355:37 |
| `loader.familyGateFor` | 371:9 |
| `coordinateMarketProposals` | 372:26 |
| `len` | 373:30 |
| `distinctGatedOutcomes` | 373:54 |
| `fail` | 384:13 |
| `fail` | 391:13 |
| `fail` | 400:13 |
| `fail` | 408:13 |
| `string` | 410:40 |
| `arbitration.entries` | 416:23 |
| `fail` | 418:13 |
| `len` | 425:5 |
| `fail` | 426:13 |
| `len` | 436:17 |
| `len` | 436:53 |
| `strategyProposalSetDigest` | 437:23 |

## Safety conclusion

- familyGateFor 는 읽기 전용(파일 읽기 · 검증) — 원장 · 브로커 · 토글 쓰기 0. 조기 실행으로 FX · 권한 실패 주기에도 그 읽기가 돈다(시장당 파도당 소형 파일 1회).
