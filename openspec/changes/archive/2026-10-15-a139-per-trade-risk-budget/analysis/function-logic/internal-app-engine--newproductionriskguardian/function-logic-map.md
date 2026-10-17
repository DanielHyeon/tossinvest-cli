# Function Logic Map: `newProductionRiskGuardian`

- Source: `internal/app/engine/guardian_wiring.go` (80-116)
- Qualified: `newProductionRiskGuardian`
- AST evidence: `ast.json`(현재) — 현재 AST branches 4 · returns 5
- 편집 전 기준: `ast-base.json`(base, `source_sha256` c21b52d0ebb66586…) — 편집 전 갈래 수 3
- Risk scan: `risk-pattern-report.md`

**역할.** 생산 Guardian 생성 — 사전 검사 · 정책 유도 · 검증 · 생성.

**a139 편집.** Validate 앞에서 `risk.PerTradeRiskBudget` 유도 실패를 이름으로 밝혀 ErrLimitsRequired — 새 분기 하나(freeze 리뷰 P3).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if v := gate.EffectiveMaxOpenPositions(); !wholeOpenPositions(v) {` (if 88) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'ProductionPolicy|A138|A139'` · `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` · `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` |
| B2 | `if audit := perTradeRiskBudgetAudit(gate); audit.value == "" {` (if 93) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'ProductionPolicy|A138|A139'` · `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` · `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` |
| B3 | `if err := policy.Validate(); err != nil {` (if 97) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'ProductionPolicy|A138|A139'` · `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` · `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` |
| B4 | `if err != nil {` (if 112) | — | 분기 본문 참조 | `go test ./internal/app/engine -run 'ProductionPolicy|A138|A139'` · `TestA139AnUnderivableBudgetRefusesTheProductionGuardianByName` · `TestA139ProductionPolicyDerivesAThirdOfTheDailyLoss` |

Exact AST return positions: 89 · 94 · 98 · 113 · 115

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 9종 — 대표: `costs.DefaultModel`, `execgw.NewManualBuyBoard`, `execgw.NewRiskGuardian`, `fmt.Errorf`, `gate.EffectiveMaxOpenPositions`, `perTradeRiskBudgetAudit`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a139 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a139 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
