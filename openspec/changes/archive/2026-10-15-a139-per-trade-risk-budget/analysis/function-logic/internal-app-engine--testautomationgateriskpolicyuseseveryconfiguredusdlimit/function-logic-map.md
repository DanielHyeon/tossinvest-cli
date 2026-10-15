# Function Logic Map: `TestAutomationGateRiskPolicyUsesEveryConfiguredUSDLimit`

- Source: `internal/app/engine/guardian_wiring_internal_test.go` (15-48)
- Qualified: `TestAutomationGateRiskPolicyUsesEveryConfiguredUSDLimit`
- AST evidence: `ast.json`(현재) — 현재 AST branches 3 · returns 0
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 18e770a3ee196c4c…) — 편집 전 갈래 수 3
- Risk scan: `risk-pattern-report.md`

**역할.** 시험 — USD 게이트의 정책 유도 전 칸.

**a139 편집.** 기대 예산 50.75 → 16.91(floor(50.75 / 3, 0.01)).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if !reflect.DeepEqual(got, want) {` (if 37) | — | 분기 본문 참조 | `go test ./internal/app/engine -run RiskPolicy` |
| B2 | `if err != nil {` (if 42) | — | 분기 본문 참조 | `go test ./internal/app/engine -run RiskPolicy` |
| B3 | `if want := gateLimits(gate); !reflect.DeepEqual(limits, want) {` (if 45) | — | 분기 본문 참조 | `go test ./internal/app/engine -run RiskPolicy` |

Exact AST return positions: (없음)

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 6종 — 대표: `execgw.ExposureLimitsFor`, `gateLimits`, `reflect.DeepEqual`, `risk.DefaultPolicy`, `riskPolicyFromAutomationGate`, `t.Fatalf`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a139 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a139 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
