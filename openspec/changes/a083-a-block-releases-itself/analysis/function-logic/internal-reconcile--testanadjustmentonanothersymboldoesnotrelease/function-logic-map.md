# Function Logic Map: `TestAnAdjustmentOnAnotherSymbolDoesNotRelease`

- Source: `internal/reconcile/mismatch_test.go` (lines 222–237)
- AST evidence: `ast.json` (`source_sha256: 797c553eb4d9cc2c161b1a7e6bb6cc78f6f77d4a3a38e9c9b768061466bf61a1`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 코드. 주문·손절·원장 판정 경로가 아니다.

## What it does

credit은 심볼별이다. a083에서 as-of를 붙였고, 단언은 바뀌지 않았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 해당 테스트가 구성한 값 | 테스트 본문 | 단언 실패로 드러난다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (231) `if` — if len(out.Cleared) != 0 | 테스트 단언 | — | 아래 Branch Test Map |
| B2 | (234) `if` — if tracker.EntryAllowed("us", "AAPL") == nil | 테스트 단언 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 223, 'column': 9}, 'text': 'clock.NewFake'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 224, 'column': 13}, 'text': 'newTracker'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 226, 'column': 22}, 'text': 'mismatchDiffAt'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 226, 'column': 2}, 'text': 'observe'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 227, 'column': 28}, 'text': 'asOfAt'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 227, 'column': 2}, 'text': 'tracker.AdjustmentApplied'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 228, 'column': 2}, 'text': 'clk.Advance'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 230, 'column': 29}, 'text': 'cleanDiffAt'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 230, 'column': 9}, 'text': 'observe'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 231, 'column': 5}, 'text': 'len'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 232, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 234, 'column': 5}, 'text': 'tracker.EntryAllowed'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 상태만. 프로덕션 상태를 변경하지 않는다.

## Safety conclusion

- Safe edit boundary: a083이 바꾼 것은 credit·diff에 비교 as-of를 붙이고 관측 사이에 시계를 진행시킨 것뿐이다. 어떤 단언도 완화하지 않았다.
- High-risk impact: no — 테스트 함수다. 다만 이 테스트들이 검증하는 대상은 High-risk 경로이므로, 단언을 약화하지 않았다는 것이 이 map의 요점이다.
