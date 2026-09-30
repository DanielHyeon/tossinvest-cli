# Function Logic Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `a356e5ead7d719e7b791423645a86b2a2f8b2eed26066127eafb1928ec411288`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- Source range: `295:1`–`441:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 제안 집합 digest 의 식은 `strategyProposalSetDigest` 하나다 — 조립과 dispatch 대조가 같은 함수를 부른다.

## Branches and early returns

- Exact AST return nodes: `310:3, 315:3, 318:3, 321:3, 326:3, 338:4, 350:3, 359:3, 390:3, 398:3, 407:3, 416:3, 425:3, 432:3, 436:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 314:2 | 편집 전과 같은 분기(좌표만) |
| B2 | if | 317:2 | 편집 전과 같은 분기(좌표만) |
| B3 | if | 320:2 | 편집 전과 같은 분기(좌표만) |
| B4 | if | 325:2 | 편집 전과 같은 분기(좌표만) |
| B5 | if | 329:2 | 편집 전과 같은 분기(좌표만) |
| B6 | range | 335:2 | 편집 전과 같은 분기(좌표만) |
| B7 | if | 337:3 | 편집 전과 같은 분기(좌표만) |
| B8 | if | 349:2 | 편집 전과 같은 분기(좌표만) |
| B9 | if | 354:2 | 편집 전과 같은 분기(좌표만) |
| B10 | if | 385:2 | 편집 전과 같은 분기(좌표만) |
| B11 | if | 392:2 | 편집 전과 같은 분기(좌표만) |
| B12 | if | 401:2 | 편집 전과 같은 분기(좌표만) |
| B13 | if | 409:2 | 편집 전과 같은 분기(좌표만) |
| B14 | if | 419:2 | 편집 전과 같은 분기(좌표만) |
| B15 | if | 427:2 | 편집 전과 같은 분기(좌표만) |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 311:90 |
| `len` | 314:31 |
| `fail` | 315:10 |
| `fail` | 318:10 |
| `fail` | 321:10 |
| `strings.TrimSpace` | 323:13 |
| `loader.getenv` | 323:31 |
| `DecodeString` | 324:14 |
| `base64.StdEncoding.Strict` | 324:14 |
| `base64.StdEncoding.EncodeToString` | 325:19 |
| `len` | 325:72 |
| `fail` | 326:10 |
| `strings.TrimSpace` | 332:12 |
| `loader.getenv` | 332:30 |
| `make` | 333:13 |
| `len` | 333:58 |
| `make` | 334:14 |
| `len` | 334:59 |
| `entry.approved.Symbol` | 336:13 |
| `bySymbol.approved.Valid` | 337:22 |
| `fail` | 338:11 |
| `append` | 341:13 |
| `entry.route.Request` | 341:97 |
| `loader.load` | 343:16 |
| `strategyrouter.Market` | 344:42 |
| `strings.TrimSpace` | 344:111 |
| `loader.getenv` | 344:129 |
| `ed25519.PublicKey` | 345:15 |
| `strings.TrimSpace` | 348:23 |
| `loader.getenv` | 348:41 |
| `batch.ManifestDigest` | 349:19 |
| `fail` | 350:10 |
| `batch.Fault` | 354:22 |
| `fail` | 355:13 |
| `absence.String` | 357:37 |
| `loader.familyGateFor` | 373:9 |
| `coordinateMarketProposals` | 374:26 |
| `len` | 375:30 |
| `distinctGatedOutcomes` | 375:54 |
| `fail` | 386:13 |
| `fail` | 393:13 |
| `fail` | 402:13 |
| `fail` | 410:13 |
| `string` | 412:40 |
| `arbitration.entries` | 418:23 |
| `fail` | 420:13 |
| `len` | 427:5 |
| `fail` | 428:13 |
| `len` | 438:17 |
| `len` | 438:53 |
| `strategyProposalSetDigest` | 439:23 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- 동작 불변 리팩터(같은 식) — `TestTheProposalSetDigestMatchesWhatTheAssemblyRecords` 가 실제 중재 경로로 잰다; 변이 X18(조립 쪽만 바뀜) CAUGHT.
