# Function Logic Map: `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder`

- Source: `internal/reconcile/restore_test.go` (lines 476–531)
- AST evidence: `ast.json` (`source_sha256: 2fd89803606be1f4664a60d079030df381cc426a784a0322c030ec81c71afdc4`)
- Risk scan: `risk-pattern-report.md`
- 위험 등급: **Normal** — 테스트 코드. 주문·손절·원장 판정 경로가 아니다.

## What it does

부분 persist 실패 후 남은 credit이 재시도에서 해제된다. a083의 D4 회계(answered-but-not-committed 보존)를 고정한다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| 테스트 fixture | 해당 테스트가 구성한 값 | 테스트 본문 | 단언 실패로 드러난다 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return/error | Required test |
|---|---|---|---|---|
| B1 | (480) `range` — for _, symbol := range []string{"AAPL", "MSFT"} | 테스트 단언 | — | 아래 Branch Test Map |
| B2 | (481) `if` — if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest | 테스트 단언 | — | 아래 Branch Test Map |
| B3 | (490) `if` — if err := tracker.Restore(ctx); err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B4 | (499) `if` — if err == nil | 테스트 단언 | — | 아래 Branch Test Map |
| B5 | (502) `if` — if len(out.Cleared) != 1 || out.Cleared[0].Symbol != "AAPL" | 테스트 단언 | — | 아래 Branch Test Map |
| B6 | (505) `if` — if blocks := tracker.Blocks(); len(blocks) != 1 || blocks[0].Symbol != "MSFT" | 테스트 단언 | — | 아래 Branch Test Map |
| B7 | (509) `if` — if err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B8 | (512) `if` — if len(active) != 1 || active[0].Symbol != "MSFT" | 테스트 단언 | — | 아래 Branch Test Map |
| B9 | (515) `if` — if gate.CheckEntryFor("us", "AAPL") != nil || gate.CheckEntryFor("us", "MSFT")… | 테스트 단언 | — | 아래 Branch Test Map |
| B10 | (522) `if` — if err != nil | 테스트 단언 | — | 아래 Branch Test Map |
| B11 | (525) `if` — if len(out.Cleared) != 1 || out.Cleared[0].Symbol != "MSFT" || out.Blocked | 테스트 단언 | — | 아래 Branch Test Map |
| B12 | (528) `if` — if gate.CheckEntryFor("us", "MSFT") != nil | 테스트 단언 | — | 아래 Branch Test Map |

## Calls and live bindings

| Callee | Why called | Error/timeout/retry contract | Evidence |
|---|---|---|---|
| `{'kind': 'call', 'at': {'line': 477, 'column': 9}, 'text': 'context.Background'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 478, 'column': 9}, 'text': 'clock.NewFake'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 479, 'column': 7}, 'text': 'openJournal'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 481, 'column': 19}, 'text': 'j.EnterReconcile'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 485, 'column': 4}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 488, 'column': 10}, 'text': 'noStaleGate'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 489, 'column': 13}, 'text': 'trackerOn'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 490, 'column': 12}, 'text': 'tracker.Restore'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 491, 'column': 3}, 'text': 't.Fatalf'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 495, 'column': 28}, 'text': 'asOfAt'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 495, 'column': 2}, 'text': 'tracker.AdjustmentApplied'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |
| `{'kind': 'call', 'at': {'line': 496, 'column': 2}, 'text': 'clk.Advance'}` | 본문 참조 | 호출부 계약 유지 | AST `calls` |

## State mutations and fallbacks

- 테스트 로컬 상태만. 프로덕션 상태를 변경하지 않는다.

## Safety conclusion

- Safe edit boundary: a083이 바꾼 것은 credit·diff에 비교 as-of를 붙이고 관측 사이에 시계를 진행시킨 것뿐이다. 어떤 단언도 완화하지 않았다.
- High-risk impact: no — 테스트 함수다. 다만 이 테스트들이 검증하는 대상은 High-risk 경로이므로, 단언을 약화하지 않았다는 것이 이 map의 요점이다.
