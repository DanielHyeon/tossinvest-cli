# Function Logic Map: `Gateway.checkEntry`

- Source: `internal/execgw/gateway.go` (850-860, 편집 뒤)
- Qualified: `Gateway.checkEntry`
- AST evidence: `ast.json`(편집 뒤, `source_sha256` b4b020607839345a…) · `ast-base.json`(편집 전, base `a000cf29`, `source_sha256` 96264a796aa63f4a…, 편집 전 분기수 1) — 현재 branches 2 · returns 3
- Risk scan: `risk-pattern-report.md`

**역할.** 노출을 늘리는 변이의 진입 관문 질의.

**a137 편집.** 전략 계획(`plan.strategy != nil`)일 때만 거절에 일일 손실 표지를 닮(B2 추가). 레거시 경로는 바이트 단위로 같음.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| plan | raisesExposure · strategy | submit | 매도는 B1 에서 nil(관문 없음) |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if g.entry == nil \|\| !plan.raisesExposure {` (if 851) | — | 분기 본문 참조 | `go test ./internal/execgw -run Gateway` |
| B2 | `if plan.strategy != nil {` (if 855) | — | 분기 본문 참조 | `TestA137AFenceCausedByTheDailyLossEscalationIsAnAccountBlock`(submit :588 · :665 경유) · `TestA137OtherFencesBehaveAsBefore` |

Exact AST return positions: 852 · 857 · 859

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `g.entry.CheckEntryFor` | 관문 판정 | — | AST calls |
| `g.entry.markStrategyRefusal` | 표지(a137) | — | AST calls |

## State mutations and fallbacks

- 없음.

## Safety conclusion

- Safe edit boundary: 전략 계획의 거절 표지.
- High-risk impact: yes — 매도(raisesExposure 거짓)는 B1 에서 그대로 통과(§0.3).
