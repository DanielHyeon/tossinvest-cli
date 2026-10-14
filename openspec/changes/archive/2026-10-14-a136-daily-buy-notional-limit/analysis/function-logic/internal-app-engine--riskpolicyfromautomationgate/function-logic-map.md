# Function Logic Map: `riskPolicyFromAutomationGate`

- Source: `internal/app/engine/guardian_wiring.go` (34-67, 편집 뒤)
- Qualified: `riskPolicyFromAutomationGate`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` cfc8eee17fbb7442…, 편집 전 분기 수 0) — 현재 branches 0 · returns 2
- Risk scan: `risk-pattern-report.md`

**역할.** 감사된 게이트 → Guardian 정책 단일 유도.

**a136 편집.** `policy.MaxDailyBuyNotional = {number(gate.EffectiveMaxDailyBuyNotional()), currency}` 한 줄. 해석(기본값)은 config 한 곳.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `gate` | 임의(검증은 Validate) | config | Validate 가 거부 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| (없음) | 직선 함수 — AST branches 0 | — | — | `TestA136ProductionPolicyResolvesTheDailyBuyCeiling` · `TestAutomationGateRiskPolicyUsesEveryConfiguredUSDLimit` |

Exact AST return positions: 39 · 66

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `risk.DefaultPolicy` · `EffectiveMaxDailyBuyNotional` | 기본 + 해석 | — | AST |

## State mutations and fallbacks

- 지역 정책.

## Safety conclusion

- Safe edit boundary: 대입 한 줄.
- High-risk impact: yes(Guardian 정책)
