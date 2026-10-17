# Function Logic Map: `collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `a349829128400c6ae6845aa49d271b318d582b0151532d2fcbd62515debfffbf`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=6, results=1)`
- Source range: `301:1`–`468:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 7.3.1 SHADOW).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- shadow 대입은 둘뿐(첫 문장 부재 값 · 조정 뒤) — AST 핀.

## Branches and early returns

- Exact AST return nodes: `321:3, 328:3, 339:3, 342:3, 347:3, 359:4, 371:3, 380:3, 417:3, 425:3, 434:3, 443:3, 452:3, 459:3, 463:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 325:2 | 경로 · 스케줄 미준비 → ROUTE_NOT_READY |
| B2 | if | 338:2 | FX 미준비 |
| B3 | if | 341:2 | 적재기 설정 결함 |
| B4 | if | 346:2 | 제안 공개 열쇠 무효 |
| B5 | if | 350:2 | US env 이름 |
| B6 | range | 356:2 | 경로 항목 순회 |
| B7 | if | 358:3 | 종목 중복 |
| B8 | if | 370:2 | 제안 적재 실패 · digest 불일치 |
| B9 | if | 375:2 | 받아들인 범위가 제안을 잃음 |
| B10 | if | 412:2 | 관문이 범위를 지움 → FAMILY_GATE_CLOSED(묶음 실음) |
| B11 | if | 419:2 | 계보 충돌(조정자의 부재 값 그대로) |
| B12 | if | 428:2 | 큐 넘침 |
| B13 | if | 436:2 | 중재 거절 |
| B14 | if | 446:2 | 미해결 선택 |
| B15 | if | 454:2 | 받아들인 범위 0 |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `len` | 322:90 |
| `len` | 325:31 |
| `fail` | 328:10 |
| `loader.familyGateFor` | 337:9 |
| `fail` | 339:10 |
| `fail` | 342:10 |
| `strings.TrimSpace` | 344:13 |
| `loader.getenv` | 344:31 |
| `DecodeString` | 345:14 |
| `base64.StdEncoding.Strict` | 345:14 |
| `base64.StdEncoding.EncodeToString` | 346:19 |
| `len` | 346:72 |
| `fail` | 347:10 |
| `strings.TrimSpace` | 353:12 |
| `loader.getenv` | 353:30 |
| `make` | 354:13 |
| `len` | 354:58 |
| `make` | 355:14 |
| `len` | 355:59 |
| `entry.approved.Symbol` | 357:13 |
| `bySymbol.approved.Valid` | 358:22 |
| `fail` | 359:11 |
| `append` | 362:13 |
| `entry.route.Request` | 362:97 |
| `loader.load` | 364:16 |
| `strategyrouter.Market` | 365:42 |
| `strings.TrimSpace` | 365:111 |
| `loader.getenv` | 365:129 |
| `ed25519.PublicKey` | 366:15 |
| `strings.TrimSpace` | 369:23 |
| `loader.getenv` | 369:41 |
| `batch.ManifestDigest` | 370:19 |
| `fail` | 371:10 |
| `batch.Fault` | 375:22 |
| `fail` | 376:13 |
| `absence.String` | 378:37 |
| `loader.familyGateFor` | 399:9 |
| `coordinateMarketProposals` | 400:37 |
| `collected.boundTo` | 401:12 |
| `loader.shadowConfig` | 401:30 |
| `len` | 402:30 |
| `distinctGatedOutcomes` | 402:54 |
| `fail` | 413:13 |
| `fail` | 420:13 |
| `fail` | 429:13 |
| `fail` | 437:13 |
| `string` | 439:40 |
| `arbitration.entries` | 445:23 |
| `fail` | 447:13 |
| `len` | 454:5 |
| `fail` | 455:13 |
| `len` | 465:17 |
| `len` | 465:53 |
| `strategyProposalSetDigest` | 466:23 |

## State mutations and fallbacks

- 호출자의 shadow 칸 두 번 대입(그 밖 쓰기 0).

## Safety conclusion

- High-risk(조정 · 제안 권한 · 조립) 경로의 편집은 운반뿐이다 — 조정 · admit · Submit · Arbitrate · dispatch 의 입력 · 순서 · 반환은 편집 전과 같고(차등 dispatch 시험 · 변이 S01~S08), shadow 값은 authority 구조체에 들어가지 않는다(census ② `TestOnlyTheAllowedFunctionsEverTouchAShadowType`).

0.5 응답 로트: 같은 파일의 주석 · 한 글자 편집(본문 구조 불변)으로 파일 SHA 만 바뀜
