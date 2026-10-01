# Function Logic Map: `buildProductionStrategyMarketWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `6f1f6804cfd437116c16a48d1526327c28360433f09e0ba3567eb3537442ee5b`
- Signature: `buildProductionStrategyMarketWorker(params=13, results=1)`
- Source range: `426:1`–`512:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2 리뷰 수리).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 활성화 없는 시장은 범위 하나 — 최소 만료 = 그 범위의 FreshUntil = 편집 전 값, forScope 는 그 범위 그대로(새 거절 0).

## Branches and early returns

- Exact AST return nodes: `433:3, 447:3, 484:3, 507:3, 509:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 432:2 | 배선 미완/nil → dormant |
| B2 | if | 445:2 | 시장 권한 준비 미완 → dormant |
| B3 | range | 454:2 | 주문 경로와 같은 handoff 목록 순회 |
| B4 | if | 456:3 | handoff 거절 · 봉인 깨짐 → 그 범위 건너뜀 |
| B5 | if | 462:3 | **(새)** 범위 키 정규화 실패 → 건너뜀 |
| B6 | if | 465:3 | **(새)** 그 범위의 위험 권한 준비 안 됨 → 건너뜀(승격 근거 아님) |
| B7 | if | 468:3 | **(새)** 그 범위의 계좌 권한 준비 안 됨 → 건너뜀 |
| B8 | if | 474:3 | 보호 관측 실패 → 그 범위 건너뜀 |
| B9 | if | 477:3 | 진입 관문 관측 실패 → 그 범위 건너뜀(J3) |
| B10 | if | 483:2 | 승격 근거 범위 없음 → dormant |
| B11 | if | 506:2 | digest / revision / **만료(준비된 계좌 범위의 최소 FreshUntil)** → dormant |

## Calls and live bindings

| Callee expression | Position |
|---|---|
| `schedule.forMarket` | 441:27 |
| `candidate.forMarket` | 441:55 |
| `route.forMarket` | 441:84 |
| `fx.forMarket` | 442:3 |
| `proposal.forMarket` | 442:25 |
| `riskAuthority.forMarket` | 442:53 |
| `account.forMarket` | 442:86 |
| `p.dispatchHandoffs` | 454:26 |
| `handoff.Single` | 455:24 |
| `result.ValidProposal` | 456:21 |
| `strategyOwnerKeyOf` | 461:17 |
| `r.forScope` | 465:18 |
| `a.forScope` | 468:18 |
| `gateway.ObserveStrategyProtection` | 474:16 |
| `strings.ToLower` | 474:55 |
| `string` | 474:71 |
| `gateway.ObserveStrategyEntryGate` | 477:16 |
| `strings.ToLower` | 477:54 |
| `string` | 477:70 |
| `strategyWorkerEvidenceDigest` | 501:12 |
| `a.earliestFreshUntil` | 505:15 |
| `validStrategyDigest` | 506:6 |
| `expiresAt.IsZero` | 506:64 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk 인접(승격). 편집은 승격을 **좁히기만** 한다(권한 없는 범위로의 승격 · 늦은 만료 제거). 주문은 dispatch 가 범위마다 다시 검사.

a112 6.3 (c) — 같은 파일 재검증 판정 이동(6 줄 감소)으로 줄만 밀림
