# Function Logic Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `dc5fbecc120d415e2c7607f6892741395d555b137e4fe6ad25fbf04c60011fa7`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- Source range: `293:1`–`449:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.8.4-B).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 13 닫힘의 kind 순서와 관문 계산 자리 하나 · fail 클로저 하나 · 성공 반환 하나가 gate.activation 을 싣는다 — census 가 못 박음(새 갈래는 표 편집 강제).
- familyGateFor 는 읽기 전용 — env 읽기, `LoadProductionFamilyActivation`(매니페스트 파일 읽기 · 검증), 레인 목록 조회. 원장 · 브로커 · 토글 · 게이트웨이 쓰기 0.

## Branches and early returns

- Exact AST return nodes: `311:3, 318:3, 327:3, 330:3, 335:3, 347:4, 359:3, 368:3, 398:3, 406:3, 415:3, 424:3, 433:3, 440:3, 444:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 315:2 | 경로 · 스케줄 미준비 → ROUTE_NOT_READY(**영값** — 관문 계산 전; 결속 값이 없거나 믿을 수 없음) |
| B2 | if | 326:2 | FX 미준비 → FX_NOT_READY(관문 활성화) |
| B3 | if | 329:2 | 적재기 설정 결손 → INTERNAL_FAILURE(관문 활성화) |
| B4 | if | 334:2 | 제안 공개 열쇠 무효 → AUTHORITY_INVALID(관문 활성화) |
| B5 | if | 338:2 | US 시장 env 이름 |
| B6 | range | 344:2 | 경로 항목 순회 |
| B7 | if | 346:3 | 종목 빈값 · 중복 → INTERNAL_FAILURE(관문 활성화) |
| B8 | if | 358:2 | 제안 적재 실패 · digest 불일치 → AUTHORITY_INVALID(관문 활성화) |
| B9 | if | 363:2 | 범위가 제안을 잃음 → PROPOSAL_PRODUCTION_FAULT(관문 활성화) |
| B10 | if | 393:2 | 관문이 범위를 통째로 지움 → FAMILY_GATE_CLOSED(판정 자리 불변) |
| B11 | if | 400:2 | 계보 신원 충돌 → INTERNAL_FAILURE |
| B12 | if | 409:2 | 조정자 넘침 → QUEUE_OVERFLOW |
| B13 | if | 417:2 | 중재 거절 → ARBITRATION_REFUSED |
| B14 | if | 427:2 | 선택을 되돌리지 못함 → INTERNAL_FAILURE |
| B15 | if | 435:2 | 받아들인 범위 0 → NO_ACCEPTED_SCOPE(관문 뒤 대조군) |

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

## State mutations and fallbacks

- 상태 변경 없음 — 권위 값을 조립해 돌려준다.

## Safety conclusion

- 관문 계산이 앞당겨져 FX · 설정 · 열쇠 · 중복 · 적재 · 고장으로 닫히는 주기에도 활성화 매니페스트 읽기가 돈다(시장당 파도당 소형 파일 1회, 읽기 전용). 진입 경로라 손절 즉시성과 무관하다(Manager 판정 Q-B2 수용).
- 닫힌 시장은 항목이 0 이라 실은 활성화로 주문이 나가지 않는다 — 실은 활성화는 그 주기의 레인 관측(승격)에만 쓰인다(TestAClosedMarketStillCarriesTheGatesActivation).
