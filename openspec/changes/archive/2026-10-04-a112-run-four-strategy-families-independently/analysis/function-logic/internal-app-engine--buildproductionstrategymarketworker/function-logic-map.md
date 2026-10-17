# Function Logic Map: `buildProductionStrategyMarketWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `9cb510c1c9c7f44ac109de5fd8ddac8e6c559adb4d654b1acd65dce19cc9e577`
- Signature: `buildProductionStrategyMarketWorker(params=13, results=1)`
- Source range: `428:1`–`514:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 활성화 없는 시장은 범위 하나 — 최소 만료 = 그 범위의 FreshUntil = 편집 전 값, forScope 는 그 범위 그대로(새 거절 0).

## Branches and early returns

- Exact AST return nodes: `435:3, 449:3, 486:3, 509:3, 511:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 434:2 | 배선 미완/nil → dormant |
| B2 | if | 447:2 | 시장 권한 준비 미완 → dormant |
| B3 | range | 456:2 | 주문 경로와 같은 handoff 목록 순회 |
| B4 | if | 458:3 | handoff 거절 · 봉인 깨짐 → 그 범위 건너뜀 |
| B5 | if | 464:3 | **(새)** 범위 키 정규화 실패 → 건너뜀 |
| B6 | if | 467:3 | **(새)** 그 범위의 위험 권한 준비 안 됨 → 건너뜀(승격 근거 아님) |
| B7 | if | 470:3 | **(새)** 그 범위의 계좌 권한 준비 안 됨 → 건너뜀 |
| B8 | if | 476:3 | 보호 관측 실패 → 그 범위 건너뜀 |
| B9 | if | 479:3 | 진입 관문 관측 실패 → 그 범위 건너뜀(J3) |
| B10 | if | 485:2 | 승격 근거 범위 없음 → dormant |
| B11 | if | 508:2 | digest / revision / **만료(준비된 계좌 범위의 최소 FreshUntil)** → dormant |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `schedule.forMarket` | 443:27 |
| `candidate.forMarket` | 443:55 |
| `route.forMarket` | 443:84 |
| `fx.forMarket` | 444:3 |
| `proposal.forMarket` | 444:25 |
| `riskAuthority.forMarket` | 444:53 |
| `account.forMarket` | 444:86 |
| `p.dispatchHandoffs` | 456:26 |
| `handoff.Single` | 457:24 |
| `result.ValidProposal` | 458:21 |
| `strategyOwnerKeyOf` | 463:17 |
| `r.forScope` | 467:18 |
| `a.forScope` | 470:18 |
| `gateway.ObserveStrategyProtection` | 476:16 |
| `strings.ToLower` | 476:55 |
| `string` | 476:71 |
| `gateway.ObserveStrategyEntryGate` | 479:16 |
| `strings.ToLower` | 479:54 |
| `string` | 479:70 |
| `strategyWorkerEvidenceDigest` | 503:12 |
| `a.earliestFreshUntil` | 507:15 |
| `validStrategyDigest` | 508:6 |
| `expiresAt.IsZero` | 508:64 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk 인접(승격). 편집은 승격을 **좁히기만** 한다(권한 없는 범위로의 승격 · 늦은 만료 제거). 주문은 dispatch 가 범위마다 다시 검사.

a112 6.3 (c) — 같은 파일 재검증 판정 이동(6 줄 감소)으로 줄만 밀림

a112 7.3.1 SHADOW 로트 — 같은 파일의 다른 함수 편집으로 줄만 밀림(본문 불변)
