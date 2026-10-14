# Function Logic Map: `engineRuntime`

- Source: `cmd/tossctl/engine.go` (655-750, 편집 뒤)
- Qualified: `engineRuntime`
- AST evidence: `ast.json`(편집 뒤) · `ast-base.json`(편집 전, base f504cc52, `source_sha256` 458cde7b2cbdef09…, 편집 전 분기 수 6) — 현재 branches 6 · returns 7
- Risk scan: `risk-pattern-report.md`

**역할.** 엔진 런타임 조립(감독 루프 · 보조 실행자).

**a136 편집.** 보조 실행자 목록 끝에 `ectx.ManualBuyRefresher(clk)` 추가.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 인자 | — | 호출자 | — |

## Branches and early returns

| Branch | Condition (소스 줄 인용) | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | `if err != nil {` (if 658) | — | 분기 본문 참조 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| B2 | `if err != nil {` (if 667) | — | 분기 본문 참조 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| B3 | `if err != nil {` (if 682) | — | 분기 본문 참조 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| B4 | `if err != nil {` (if 687) | — | 분기 본문 참조 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| B5 | `if err != nil {` (if 691) | — | 분기 본문 참조 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |
| B6 | `if err != nil {` (if 703) | — | 분기 본문 참조 | `TestProductionRuntimeStartsExactlyOneAlertDeliverer` |

Exact AST return positions: 659 · 668 · 683 · 688 · 692 · 704 · 707

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `ectx.ManualBuyRefresher` | 수동 매수 읽기 | 보조 실행자(정지가 다른 루프를 내리지 않음) | AST |

## State mutations and fallbacks

- 없음(구성만).

## Safety conclusion

- Safe edit boundary: 칸 · 등록 하나 추가 — 기존 분기 무변경.
- High-risk impact: yes(런타임 조립 — 감독 루프 무변경)
