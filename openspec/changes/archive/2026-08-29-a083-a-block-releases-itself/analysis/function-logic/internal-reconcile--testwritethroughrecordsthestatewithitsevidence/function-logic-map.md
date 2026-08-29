# Function Logic Map: `TestWriteThroughRecordsTheStateWithItsEvidence`

- Source: `internal/reconcile/restore_test.go` (lines 715–774)
- AST evidence: `ast.json` (`source_sha256: 2fd89803606be1f4664a60d079030df381cc426a784a0322c030ec81c71afdc4`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 코드. 주문·손절·원장 판정 경로가 아니다.

## What it does

해제가 ADJUSTMENT_APPLIED 원인과 함께 원장에 기록된다. a083에서 credit에 as-of를 붙였다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 해당 테스트가 구성한 값 | 테스트 본문 | 단언 실패로 드러난다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (721) `if` — if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B2 | (725) `if` — if err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B3 | (728) `if` — if len(active) != 1 | 테스트 단언 | — | 아래 Branch Test Map |
| B4 | (732) `if` — if state.Symbol != "AAPL" || state.Cause != journal.ReconcileCauseQuantityMism… | 테스트 단언 | — | 아래 Branch Test Map |
| B5 | (735) `if` — if state.Evidence == "" | 테스트 단언 | — | 아래 Branch Test Map |
| B6 | (743) `if` — if _, err := tracker.Observe(ctx, reconcile.Diff{AccountRef: "acct-7", Matched… | 테스트 단언 | — | 아래 Branch Test Map |
| B7 | (746) `if` — if active, err = j.ActiveReconcileStates(ctx); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B8 | (749) `if` — if len(active) != 1 | 테스트 단언 | — | 아래 Branch Test Map |
| B9 | (758) `if` — if _, err := tracker.Observe(ctx, cleanDiffAt(clk)); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B10 | (761) `if` — if active, err = j.ActiveReconcileStates(ctx); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B11 | (764) `if` — if len(active) != 0 | 테스트 단언 | — | 아래 Branch Test Map |
| B12 | (768) `if` — if err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B13 | (771) `if` — if len(history) != 1 || history[0].ReleaseCause != journal.ReconcileReleaseAdj… | 테스트 단언 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 716, 'column': 9}, 'text': 'context.Background'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 717, 'column': 9}, 'text': 'clock.NewFake'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 718, 'column': 7}, 'text': 'openJournal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 719, 'column': 13}, 'text': 'trackerOn'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 719, 'column': 28}, 'text': 'noStaleGate'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 721, 'column': 15}, 'text': 'tracker.Observe'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 721, 'column': 36}, 'text': 'mismatchDiff'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 722, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 724, 'column': 17}, 'text': 'j.ActiveReconcileStates'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 726, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 728, 'column': 5}, 'text': 'len'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 729, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 상태만. 프로덕션 상태를 변경하지 않는다.

## Safety conclusion

- Safe edit boundary: a083이 바꾼 것은 credit·diff에 비교 as-of를 붙이고 관측 사이에 시계를 진행시킨 것뿐이다. 어떤 단언도 완화하지 않았다.
- High-risk impact: no — 테스트 함수다. 다만 이 테스트들이 검증하는 대상은 High-risk 경로이므로, 단언을 약화하지 않았다는 것이 이 map의 요점이다.
