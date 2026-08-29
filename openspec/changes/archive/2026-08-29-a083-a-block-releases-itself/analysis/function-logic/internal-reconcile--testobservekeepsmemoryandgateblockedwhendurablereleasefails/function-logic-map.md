# Function Logic Map: `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails`

- Source: `internal/reconcile/restore_test.go` (lines 425–474)
- AST evidence: `ast.json` (`source_sha256: 2fd89803606be1f4664a60d079030df381cc426a784a0322c030ec81c71afdc4`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 코드. 주문·손절·원장 판정 경로가 아니다.

## What it does

durable 해제가 거부되면 메모리와 gate가 차단을 유지한다. a083에서 as-of를 붙였고, 단언은 바뀌지 않았다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 해당 테스트가 구성한 값 | 테스트 본문 | 단언 실패로 드러난다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (426) `range` — for _, tc := range []struct | 테스트 단언 | — | 아래 Branch Test Map |
| B2 | (438) `if` — if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest | 테스트 단언 | — | 아래 Branch Test Map |
| B3 | (446) `if` — if err := tracker.Restore(ctx); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B4 | (457) `if` — if err == nil | 테스트 단언 | — | 아래 Branch Test Map |
| B5 | (460) `if` — if len(out.Cleared) != 0 || !out.Blocked | 테스트 단언 | — | 아래 Branch Test Map |
| B6 | (463) `if` — if blocks := tracker.Blocks(); len(blocks) != 1 || blocks[0].Cause != journal.… | 테스트 단언 | — | 아래 Branch Test Map |
| B7 | (466) `if` — if gate.CheckEntryFor("us", "AAPL") == nil | 테스트 단언 | — | 아래 Branch Test Map |
| B8 | (469) `if` — if len(store.requests) != 1 || store.requests[0].ExpectCause != journal.Reconc… | 테스트 단언 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 431, 'column': 37}, 'text': 'errors.New'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 434, 'column': 3}, 'text': 't.Run'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 435, 'column': 11}, 'text': 'context.Background'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 436, 'column': 11}, 'text': 'clock.NewFake'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 437, 'column': 9}, 'text': 'openJournal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 438, 'column': 20}, 'text': 'j.EnterReconcile'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 442, 'column': 5}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 444, 'column': 12}, 'text': 'noStaleGate'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 445, 'column': 15}, 'text': 'trackerOn'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 446, 'column': 14}, 'text': 'tracker.Restore'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 447, 'column': 5}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 453, 'column': 30}, 'text': 'asOfAt'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 상태만. 프로덕션 상태를 변경하지 않는다.

## Safety conclusion

- Safe edit boundary: a083이 바꾼 것은 credit·diff에 비교 as-of를 붙이고 관측 사이에 시계를 진행시킨 것뿐이다. 어떤 단언도 완화하지 않았다.
- High-risk impact: no — 테스트 함수다. 다만 이 테스트들이 검증하는 대상은 High-risk 경로이므로, 단언을 약화하지 않았다는 것이 이 map의 요점이다.
