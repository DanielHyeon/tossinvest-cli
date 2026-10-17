# Function Logic Map (편집 전): `collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `dc5fbecc120d415e2c7607f6892741395d555b137e4fe6ad25fbf04c60011fa7`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- Source range: `293:1`–`449:2`
- AST evidence: `ast.json` — 편집 **전**(a112 8.5 응답 로트(HEAD 178cc196)).

## Inputs and invariants

- 편집 계획: 응답 로트 ①(F1, Manager 판정 Y): 관문 판정 계산을 옛 자리(coordinateMarketProposals 바로 앞)에 되살린다 — 조기 계산(B1 뒤)은 그대로 두고 조정 앞 닫힘 여섯의 진단 carry 전용으로; 분기 추가 0(대입 한 문장).

## Branches and early returns

- Exact AST return nodes: `311:3`, `318:3`, `327:3`, `330:3`, `335:3`, `347:4`, `359:3`, `368:3`, `398:3`, `406:3`, `415:3`, `424:3`, `433:3`, `440:3`, `444:2`.

| Branch | AST kind | Source location | Condition |
|---|---|---|---|
| B1 | if | 315:2 | `if !routes.snapshot.Ready // len(routes.entries) == 0 // !schedule.snapshot.Ready // schedule.restore.Activation == nil {` |
| B2 | if | 326:2 | `if !fx.snapshot.Ready // !fx.read.valid {` |
| B3 | if | 329:2 | `if loader.getenv == nil // loader.load == nil // loader.configDir == "." // loader.evidencePath == "." // loader.journalPath == "." // loader.accountRef == "" {` |
| B4 | if | 334:2 | `if err != nil // base64.StdEncoding.EncodeToString(key) != encoded // len(key) != ed25519.PublicKeySize {` |
| B5 | if | 338:2 | `if market == StrategyMarketUS {` |
| B6 | range | 344:2 | `for _, entry := range routes.entries {` |
| B7 | if | 346:3 | `if symbol == "" // bySymbol[symbol].approved.Valid() {` |
| B8 | if | 358:2 | `if err != nil // batch.ManifestDigest() != digest {` |
| B9 | if | 363:2 | `if absence, lost := batch.Fault(); lost {` |
| B10 | if | 393:2 | `if arbitration.erasedScopes != 0 {` |
| B11 | if | 400:2 | `if arbitration.collision {` |
| B12 | if | 409:2 | `if outcome.Overflow {` |
| B13 | if | 417:2 | `if outcome.Refusal != strategyarbiter.RefusalNone {` |
| B14 | if | 427:2 | `if !resolved {` |
| B15 | if | 435:2 | `if len(entries) == 0 {` |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 312:90 |
| `len` | 315:31 |
| `fail` | 318:10 |
| `loader.familyGateFor` | 325:9 |
| `fail` | 327:10 |
| `fail` | 330:10 |
| `strings.TrimSpace` | 332:13 |
| `loader.getenv` | 332:31 |
| `DecodeString` | 333:14 |
| `base64.StdEncoding.Strict` | 333:14 |
| `base64.StdEncoding.EncodeToString` | 334:19 |
| `len` | 334:72 |
| `fail` | 335:10 |
| `strings.TrimSpace` | 341:12 |
| `loader.getenv` | 341:30 |
| `make` | 342:13 |
| `len` | 342:58 |
| `make` | 343:14 |
| `len` | 343:59 |
| `entry.approved.Symbol` | 345:13 |
| `bySymbol.approved.Valid` | 346:22 |
| `fail` | 347:11 |
| `append` | 350:13 |
| `entry.route.Request` | 350:97 |
| `loader.load` | 352:16 |
| `strategyrouter.Market` | 353:42 |
| `strings.TrimSpace` | 353:111 |
| `loader.getenv` | 353:129 |
| `ed25519.PublicKey` | 354:15 |
| `strings.TrimSpace` | 357:23 |
| `loader.getenv` | 357:41 |
| `batch.ManifestDigest` | 358:19 |
| `fail` | 359:10 |
| `batch.Fault` | 363:22 |
| `fail` | 364:13 |
| `absence.String` | 366:37 |
| `coordinateMarketProposals` | 382:26 |
| `len` | 383:30 |
| `distinctGatedOutcomes` | 383:54 |
| `fail` | 394:13 |
| `fail` | 401:13 |
| `fail` | 410:13 |
| `fail` | 418:13 |
| `string` | 420:40 |
| `arbitration.entries` | 426:23 |
| `fail` | 428:13 |
| `len` | 435:5 |
| `fail` | 436:13 |
| `len` | 446:17 |
| `len` | 446:53 |
| `strategyProposalSetDigest` | 447:23 |

## Safety conclusion

- 판정 = 편집 전(f473d815)과 같은 함수 · 같은 순간 — 적재 중 취소 · 철회가 다시 판정에 보인다(허용 방향 경합 닫힘). 비용은 조정 도달 주기의 소형 읽기 1회(오늘 핀 0).
