# Function Logic Map: `buildProductionStrategyMarketWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `64f1cc0b85ecf5693dc5df0622b0f19f665697ea1b6f7616eb6755598f546e97`
- Signature: `buildProductionStrategyMarketWorker(params=13, results=1)`
- Source range: `432:1`–`518:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 활성화 없는 시장은 범위 하나 — 최소 만료 = 그 범위의 FreshUntil = 편집 전 값, forScope 는 그 범위 그대로(새 거절 0).

## Branches and early returns

- Exact AST return nodes: `439:3, 453:3, 490:3, 513:3, 515:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 438:2 | 배선 미완/nil → dormant |
| B2 | if | 451:2 | 시장 권한 준비 미완 → dormant |
| B3 | range | 460:2 | 주문 경로와 같은 handoff 목록 순회 |
| B4 | if | 462:3 | handoff 거절 · 봉인 깨짐 → 그 범위 건너뜀 |
| B5 | if | 468:3 | **(새)** 범위 키 정규화 실패 → 건너뜀 |
| B6 | if | 471:3 | **(새)** 그 범위의 위험 권한 준비 안 됨 → 건너뜀(승격 근거 아님) |
| B7 | if | 474:3 | **(새)** 그 범위의 계좌 권한 준비 안 됨 → 건너뜀 |
| B8 | if | 480:3 | 보호 관측 실패 → 그 범위 건너뜀 |
| B9 | if | 483:3 | 진입 관문 관측 실패 → 그 범위 건너뜀(J3) |
| B10 | if | 489:2 | 승격 근거 범위 없음 → dormant |
| B11 | if | 512:2 | digest / revision / **만료(준비된 계좌 범위의 최소 FreshUntil)** → dormant |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `schedule.forMarket` | 447:27 |
| `candidate.forMarket` | 447:55 |
| `route.forMarket` | 447:84 |
| `fx.forMarket` | 448:3 |
| `proposal.forMarket` | 448:25 |
| `riskAuthority.forMarket` | 448:53 |
| `account.forMarket` | 448:86 |
| `p.dispatchHandoffs` | 460:26 |
| `handoff.Single` | 461:24 |
| `result.ValidProposal` | 462:21 |
| `strategyOwnerKeyOf` | 467:17 |
| `r.forScope` | 471:18 |
| `a.forScope` | 474:18 |
| `gateway.ObserveStrategyProtection` | 480:16 |
| `strings.ToLower` | 480:55 |
| `string` | 480:71 |
| `gateway.ObserveStrategyEntryGate` | 483:16 |
| `strings.ToLower` | 483:54 |
| `string` | 483:70 |
| `strategyWorkerEvidenceDigest` | 507:12 |
| `a.earliestFreshUntil` | 511:15 |
| `validStrategyDigest` | 512:6 |
| `expiresAt.IsZero` | 512:64 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk 인접(승격). 편집은 승격을 **좁히기만** 한다(권한 없는 범위로의 승격 · 늦은 만료 제거). 주문은 dispatch 가 범위마다 다시 검사.
