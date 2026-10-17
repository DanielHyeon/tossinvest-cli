# Function Logic Map: `newProductionRiskGuardian`

- Source: `internal/app/engine/guardian_wiring.go` (69-96, 편집 뒤)
- Qualified: `newProductionRiskGuardian`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` cfc8eee17fbb7442…, 편집 전 분기 수 2) — 현재 branches 2 · returns 3
- Risk scan: `risk-pattern-report.md`

**역할.** 생산 Guardian 구성.

**a136 편집.** `ManualBuys: execgw.NewManualBuyBoard()` 한 칸(a136 D9=b) — 보조 실행자가 게시할 보드.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err := policy.Validate(); err != nil {` (if 77) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` · `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver` |
| B2 | `if err != nil {` (if 92) | — | 분기 본문 참조 | `TestA136TheRefresherPublishesTodaysManualBuys` · `TestProductionGuardianUsesConfiguredUSDLimitsAndReachesExitObserver` |

Exact AST return positions: 78 · 93 · 95

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `execgw.NewRiskGuardian` | 생성 | 오류 반환 | AST |

## State mutations and fallbacks

- 없음(구성만).

## Safety conclusion

- Safe edit boundary: 칸 · 등록 하나 추가 — 기존 분기 무변경.
- High-risk impact: yes(Guardian 구성)
