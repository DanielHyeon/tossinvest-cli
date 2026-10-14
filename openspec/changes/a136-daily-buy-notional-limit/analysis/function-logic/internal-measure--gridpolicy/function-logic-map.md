# Function Logic Map: `gridPolicy`

- Source: `internal/measure/counterfactual.go` (353-373, 편집 뒤)
- Qualified: `gridPolicy`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` ac05b5382bc4c368…, 편집 전 분기 수 2) — 현재 branches 2 · returns 2
- Risk scan: `risk-pattern-report.md`

**역할.** 측정 도구의 넉넉한 격자 정책(기하만 거부하게).

**a136 편집.** `policy.MaxDailyBuyNotional = money(headroom, currency)` — 새 필수 칸이 US 격자를 통화 불일치로 깨지 않게(freeze 리뷰 P2-9).

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `market` | KR · US | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if market == costs.MarketUS {` (if 358) | — | 분기 본문 참조 | `go test ./internal/measure` |
| B2 | `if err := policy.Validate(); err != nil {` (if 369) | — | 분기 본문 참조 | `go test ./internal/measure` |

Exact AST return positions: 370 · 372

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `risk.DefaultPolicy` · `policy.Validate` | 기본 + 검증 | 오류 | AST |

## State mutations and fallbacks

- 지역 정책.

## Safety conclusion

- Safe edit boundary: 대입 한 줄.
- High-risk impact: no(측정 도구)
