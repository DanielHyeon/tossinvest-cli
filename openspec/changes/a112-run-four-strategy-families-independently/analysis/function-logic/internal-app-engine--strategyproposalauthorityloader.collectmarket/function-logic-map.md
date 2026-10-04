# Function Logic Map: `strategyProposalAuthorityLoader.collectMarket`

- Source: `internal/app/engine/strategy_proposal_authority.go`
- Source SHA-256: `2c546898bc4178d0ee44238d51fb8e05a4c908cee721041ee24713c4c98ac385`
- Signature: `strategyProposalAuthorityLoader.collectMarket(params=5, results=1)`
- Source range: `293:1`–`458:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 8.5-R).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 관문 대입은 최상위 문장 둘: 조기(B1 가드 바로 다음) · 판정(`coordinateMarketProposals` 바로 앞, 그 호출의 관문 인자가 `gate`) — census 가 이웃 문장까지 못 박음.
- 13 닫힘의 kind 순서 · fail 클로저 하나 · 성공 반환 하나가 gate.activation 을 싣는다 — census(새 갈래는 표 편집 강제).
- 조정 앞 닫힘(B2 · B3 · B4 · B7 · B8 · B9)은 조기 값, 조정 뒤 닫힘(B10~B14)과 성공 · B15 는 판정 값을 싣는다. **잔여 비대칭(관측 전용, Manager 판정 수용)**: 적재 중 철회 경합에서 조정 앞 닫힘은 철회 전 조기 값을 싣는다 — 그 갈래는 항목 0 이라 조정 · handoff · 주문에 닿지 않고 그 주기의 레인 관측에만 쓰인다.
- **census 의 한계(8.5 보이스 3 P2-1)**: B11 · B14 갈래 안에서 활성화를 다른 값으로 바꾸는 선택자-대입 변이(E5 · E6)는 census 도 행동도 못 본다. 두 갈래는 입력으로 도달 불가한 공간이다(계보 신원이 LaneID · LaneVersion 을 해시 — strategyflow/types.go; 중복 종목은 B7 이 먼저 닫음) — census-only 를 유지(보이스 3 지지).
- familyGateFor 는 읽기 전용 — env 읽기, `LoadProductionFamilyActivation`(매니페스트 파일 읽기 · 검증), 레인 목록 조회. 원장 · 브로커 · 토글 · 게이트웨이 쓰기 0.

## Branches and early returns

- Exact AST return nodes: `312:3, 319:3, 330:3, 333:3, 338:3, 350:4, 362:3, 371:3, 407:3, 415:3, 424:3, 433:3, 442:3, 449:3, 453:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 316:2 | 경로 · 스케줄 미준비 → ROUTE_NOT_READY(**영값** — 관문 계산 전; 결속 값이 없거나 믿을 수 없음) |
| B2 | if | 329:2 | FX 미준비 → FX_NOT_READY(조기 · 진단 활성화) |
| B3 | if | 332:2 | 적재기 설정 결손 → INTERNAL_FAILURE(조기 · 진단 활성화) |
| B4 | if | 337:2 | 제안 공개 열쇠 무효 → AUTHORITY_INVALID(조기 · 진단 활성화) |
| B5 | if | 341:2 | US 시장 env 이름 |
| B6 | range | 347:2 | 경로 항목 순회 |
| B7 | if | 349:3 | 종목 빈값 · 중복 → INTERNAL_FAILURE(조기 · 진단 활성화) |
| B8 | if | 361:2 | 제안 적재 실패 · digest 불일치 → AUTHORITY_INVALID(조기 · 진단 활성화) |
| B9 | if | 366:2 | 범위가 제안을 잃음 → PROPOSAL_PRODUCTION_FAULT(조기 · 진단 활성화) |
| B10 | if | 402:2 | 관문이 범위를 통째로 지움 → FAMILY_GATE_CLOSED(**판정** 관문 — 조정 바로 앞에서 다시 계산; 판정 활성화를 실음) |
| B11 | if | 409:2 | 계보 신원 충돌 → INTERNAL_FAILURE(판정 활성화) |
| B12 | if | 418:2 | 조정자 넘침 → QUEUE_OVERFLOW(판정 활성화) |
| B13 | if | 426:2 | 중재 거절 → ARBITRATION_REFUSED |
| B14 | if | 436:2 | 선택을 되돌리지 못함 → INTERNAL_FAILURE(판정 활성화) |
| B15 | if | 444:2 | 받아들인 범위 0 → NO_ACCEPTED_SCOPE(관문 뒤 대조군 — 판정 활성화, 적재 2 회) |

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

## State mutations and fallbacks

- 상태 변경 없음 — 권위 값을 조립해 돌려준다.

## Safety conclusion

- 판정 = 편집 전과 같은 함수 · 같은 순간 — 적재 중 ctx 취소 · 매니페스트 철회가 다시 그 주기 판정에 보인다(허용 방향 경합 닫힘; 철회는 하류에 다시 막는 자리가 없었다).
- 미선언(토글 OFF) 시장은 두 계산 모두 미선언 선반환이라 기존 경로 그대로 — 취소 주기에도 스냅숏 동일(`a112_gate_decision_recompute_test.go` `TestAnUndeclaredMarketCancelledDuringTheLoadKeepsTheLegacyPath`).
- 대가: 조정에 닿는 주기에 활성화 읽기 1 회 추가(64 KiB 이하 로컬 0400 파일, 핀 선언 시장만 — 오늘 생산 핀 0). 진입 경로라 손절 즉시성과 무관.
- 닫힌 시장은 항목이 0 이라 실은 활성화로 주문이 나가지 않는다 — 실은 활성화는 그 주기의 레인 관측(승격)에만 쓰인다.
