# Function Logic Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `60e0ef9270e3102bb69ce930cdffb41a5c4653dca69efdc94d32374823161bae`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- Source range: `293:1`–`439:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 제안 집합 digest 의 식은 `strategyProposalSetDigest` 하나다 — 조립과 dispatch 대조가 같은 함수를 부른다.

## Branches and early returns

- Exact AST return nodes: `308:3, 313:3, 316:3, 319:3, 324:3, 336:4, 348:3, 357:3, 388:3, 396:3, 405:3, 414:3, 423:3, 430:3, 434:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 312:2 | 편집 전과 같은 분기(좌표만) |
| B2 | if | 315:2 | 편집 전과 같은 분기(좌표만) |
| B3 | if | 318:2 | 편집 전과 같은 분기(좌표만) |
| B4 | if | 323:2 | 편집 전과 같은 분기(좌표만) |
| B5 | if | 327:2 | 편집 전과 같은 분기(좌표만) |
| B6 | range | 333:2 | 편집 전과 같은 분기(좌표만) |
| B7 | if | 335:3 | 편집 전과 같은 분기(좌표만) |
| B8 | if | 347:2 | 편집 전과 같은 분기(좌표만) |
| B9 | if | 352:2 | 편집 전과 같은 분기(좌표만) |
| B10 | if | 383:2 | 편집 전과 같은 분기(좌표만) |
| B11 | if | 390:2 | 편집 전과 같은 분기(좌표만) |
| B12 | if | 399:2 | 편집 전과 같은 분기(좌표만) |
| B13 | if | 407:2 | 편집 전과 같은 분기(좌표만) |
| B14 | if | 417:2 | 편집 전과 같은 분기(좌표만) |
| B15 | if | 425:2 | 편집 전과 같은 분기(좌표만) |

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

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- 동작 불변 리팩터(같은 식) — `TestTheProposalSetDigestMatchesWhatTheAssemblyRecords` 가 실제 중재 경로로 잰다; 변이 X18(조립 쪽만 바뀜) CAUGHT.

a112 7.3 — 같은 파일 편집(관측 필드 · record 인자 / QueueDropCount 주석)으로 줄만 밀림
