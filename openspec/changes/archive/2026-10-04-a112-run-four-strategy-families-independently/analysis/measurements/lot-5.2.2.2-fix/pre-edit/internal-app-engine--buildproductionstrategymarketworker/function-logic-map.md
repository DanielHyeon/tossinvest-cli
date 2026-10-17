# Function Logic Map: `buildProductionStrategyMarketWorker`

- Source: `internal/app/engine/strategy_entry_supervisor.go`
- Source SHA-256: `9e24e93028b2728071d71d1d6ccea2c2a83fe768f6efe2dc09a57906c435a373`
- Signature: `buildProductionStrategyMarketWorker(params=13, results=1)`
- Source range: `432:1`–`504:2`
- AST evidence: `ast.json` — **편집 뒤**(a112 5.2.2.2).
- Risk scan: `risk-pattern-report.md`.

## Inputs and invariants

- 활성화 없는 시장(오늘 생산 전부)은 handoff 가 하나라 편집 전과 같은 판정이다(토글 OFF = upstream).
- 관측(보호 · 진입 관문)은 준비 확인일 뿐이고 주문마다 제출 경로가 범위 단위로 다시 검사한다(`withStrategyEntryGateAuthority`).

## Branches and early returns

- Exact AST return nodes: `439:3, 453:3, 478:3, 499:3, 501:2`.

| Branch | AST kind | Source location | Meaning |
|---|---|---|---|
| B1 | if | 438:2 | 배선 미완/nil → dormant(편집 전 B1 불변) |
| B2 | if | 451:2 | 시장 권한(일정 · 후보 · 경로 · 환율 · 위험 · 계좌) 준비 미완 → dormant. **편집: handoff 조건을 뺐다**(B3~B4 로) |
| B3 | range | 460:2 | **(새)** 주문 경로와 같은 handoff 목록(`dispatchHandoffs`) 순회 — 활성화 없는 시장은 시장 단위 하나(오늘), 서명 활성화 시장은 범위마다 하나 |
| B4 | if | 462:3 | **(새, 편집 전 B2 의 handoff 절반 + 편집 전 B3)** 그 handoff 가 거절했거나 봉인 깨진 제안 → 그 범위는 승격 근거가 못 됨(continue) |
| B5 | if | 468:3 | 보호 관측 실패 → **그 범위만** 건너뜀(편집 전 B4 는 시장 dormant — 범위 하나면 B7 로 같은 결과) |
| B6 | if | 471:3 | 진입 관문 관측 실패 → **그 범위만** 건너뜀(J3 — 한 범위의 거절이 다른 범위의 승격을 굶기지 않음) |
| B7 | if | 477:2 | **(새)** 승격 근거가 된 범위 없음 → dormant |
| B8 | if | 498:2 | digest / revision / 만료 → dormant(편집 전 B6 불변) |

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
| `gateway.ObserveStrategyProtection` | 468:16 |
| `strings.ToLower` | 468:55 |
| `string` | 468:71 |
| `gateway.ObserveStrategyEntryGate` | 471:16 |
| `strings.ToLower` | 471:54 |
| `string` | 471:70 |
| `strategyWorkerEvidenceDigest` | 495:12 |
| `validStrategyDigest` | 498:6 |
| `IsZero` | 498:64 |
| `a.authority.FreshUntil` | 498:64 |
| `a.authority.FreshUntil` | 502:23 |

## State mutations and fallbacks

- 이 함수 자신은 원장을 쓰지 않는다. 봉투 · 시장 값으로의 폴백 없음(범위 권한이 없으면 거절 또는 준비 안 됨).

## Safety conclusion

- High-risk 인접(승격은 화면 · 주기 가동만 움직이고 주문은 dispatch 가 낸다). 편집은 활성화 시장에서만 승격을 **넓힌다** — 한 범위라도 모든 관측을 통과해야 한다. 새로 승격되는 입력: 서명 활성화된 두 범위 시장(편집 전 OverCapacity 로 dormant). 활성화 없는 시장의 새 통과 입력 0.
