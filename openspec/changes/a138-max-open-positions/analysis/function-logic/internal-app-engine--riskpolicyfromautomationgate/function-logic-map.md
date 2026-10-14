# Function Logic Map: `riskPolicyFromAutomationGate`

- Source: `internal/app/engine/guardian_wiring.go` (34-69)
- Qualified: `riskPolicyFromAutomationGate`
- AST evidence: `ast.json`(현재) — 현재 AST branches 0 · returns 2
- 편집 전 기준: `ast-base.json`(base, `source_sha256` 6acfab5929684cce…) — 편집 전 갈래 수 0
- Risk scan: `risk-pattern-report.md`

**역할.** 감사된 게이트 한도 → Guardian 정책 유도(직선).

**a138 편집.** `policy.MaxOpenPositions = openPositionsFromGate(gate)` 한 줄(해석 · 정수 판정은 새 leaf 함수).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 함수 인자 · 수신자 상태 | 소스 참조 | 현재 HEAD | 분기 표 참조 |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | 직선 함수 — 분기 없음 | — | 정상 경로 | `go test ./internal/app/engine -run 'RiskPolicy|A136|A138'` · `TestA138ProductionPolicyResolvesTheOpenPositionMaximum` |

Exact AST return positions: 39 · 68

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| (호출 8종 — 대표: `config.WholeOpenPositions`, `gate.EffectiveMaxDailyBuyNotional`, `gate.EffectiveMaxOpenPositions`, `number`, `risk.DefaultPolicy`, `strconv.FormatFloat`) | 본문 참조 | 소스 참조 | AST |

## State mutations and fallbacks

- 편집은 위 「a138 편집」 한 곳. 그 밖의 상태 변화 · 대체 경로는 편집 전과 같음.

## Safety conclusion

- Safe edit boundary: 「a138 편집」에 적은 한 자리.
- High-risk impact: yes(Guardian 진입 한도 — 거부 방향만 추가, 매도 경로 무관)
