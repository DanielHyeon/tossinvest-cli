# Function Logic Map: `TestTrackerReleaseFailureStopsBeforePriceAndAdoption`

- Source: `internal/app/engine/reconcileloop_test.go` (lines 404–440)
- AST evidence: `ast.json` (`source_sha256: 2456d7022ac6ce4a6de30e65d7c57d0b10b7a67ded5c6c404c38faa40e550106`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 코드. 주문·손절·원장 판정 경로가 아니다.

## What it does

tracker 실패가 시세 조회와 편입 앞에서 사이클을 멈춘다. a083에서 credit에 as-of를 붙여, 거부하는 것이 stamp가 아니라 durable 해제 실패임을 분명히 했다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 해당 테스트가 구성한 값 | 테스트 본문 | 단언 실패로 드러난다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (403) `if` — if _, _, err := h.journal.EnterReconcile(context.Background(), journal.EnterRe… | 테스트 단언 | — | 아래 Branch Test Map |
| B2 | (409) `if` — if err := h.tracker.Restore(context.Background()); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B3 | (419) `if` — if cycle.Err == nil | 테스트 단언 | — | 아래 Branch Test Map |
| B4 | (422) `if` — if cycle.Blocked != 1 || cycle.Released != 0 | 테스트 단언 | — | 아래 Branch Test Map |
| B5 | (425) `if` — if cycle.Adopted != 0 || cycle.Unmanaged != 0 || h.prices.calls != 0 | 테스트 단언 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `newDriverHarness` | (405) h := newDriverHarness(t, nil) | 호출부 계약 유지 | AST `calls` |
| `h.holds` | (406) h.holds("005930", "10", "55000", 70000) | 호출부 계약 유지 | AST `calls` |
| `h.journal.EnterReconcile` | (414) if _, _, err := h.journal.EnterReconcile(context.Background(), journal.EnterReconcileRequest{ | 호출부 계약 유지 | AST `calls` |
| `context.Background` | (414) if _, _, err := h.journal.EnterReconcile(context.Background(), journal.EnterReconcileRequest{ | 호출부 계약 유지 | AST `calls` |
| `t.Fatalf` | (418) t.Fatalf("EnterReconcile: %v", err) | 호출부 계약 유지 | AST `calls` |
| `h.tracker.Restore` | (420) if err := h.tracker.Restore(context.Background()); err != nil { | 호출부 계약 유지 | AST `calls` |
| `h.tracker.AdjustmentApplied` | (427) h.tracker.AdjustmentApplied(h.clk.Now().UTC().Format(time.RFC3339), "000660") | 호출부 계약 유지 | AST `calls` |
| `Format` | (427) h.tracker.AdjustmentApplied(h.clk.Now().UTC().Format(time.RFC3339), "000660") | 호출부 계약 유지 | AST `calls` |
| `UTC` | (427) h.tracker.AdjustmentApplied(h.clk.Now().UTC().Format(time.RFC3339), "000660") | 호출부 계약 유지 | AST `calls` |
| `h.clk.Now` | (427) h.tracker.AdjustmentApplied(h.clk.Now().UTC().Format(time.RFC3339), "000660") | 호출부 계약 유지 | AST `calls` |
| `h.cycle` | (429) cycle := h.cycle() | 호출부 계약 유지 | AST `calls` |
| `t.Fatal` | (431) t.Fatal("cycle must fail when the durable release is refused") | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 상태만. 프로덕션 상태를 변경하지 않는다.

## Safety conclusion

- Safe edit boundary: a083이 바꾼 것은 credit·diff에 비교 as-of를 붙이고 관측 사이에 시계를 진행시킨 것뿐이다. 어떤 단언도 완화하지 않았다.
- High-risk impact: no — 테스트 함수다. 다만 이 테스트들이 검증하는 대상은 High-risk 경로이므로, 단언을 약화하지 않았다는 것이 이 map의 요점이다.
