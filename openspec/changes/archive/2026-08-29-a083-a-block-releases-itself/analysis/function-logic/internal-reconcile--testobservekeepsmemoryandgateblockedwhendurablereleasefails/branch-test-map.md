# Branch Test Map: `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (426) `range` — for _, tc := range []struct | `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails` 자신 | n/a | yes |
| B2 | (438) `if` — if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest | `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails` 자신 | n/a | yes |
| B3 | (446) `if` — if err := tracker.Restore(ctx); err != nil | `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails` 자신 | n/a | yes |
| B4 | (457) `if` — if err == nil | `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails` 자신 | n/a | yes |
| B5 | (460) `if` — if len(out.Cleared) != 0 || !out.Blocked | `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails` 자신 | n/a | yes |
| B6 | (463) `if` — if blocks := tracker.Blocks(); len(blocks) != 1 || blocks[0].Cause != journal.… | `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails` 자신 | n/a | yes |
| B7 | (466) `if` — if gate.CheckEntryFor("us", "AAPL") == nil | `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails` 자신 | n/a | yes |
| B8 | (469) `if` — if len(store.requests) != 1 || store.requests[0].ExpectCause != journal.Reconc… | `TestObserveKeepsMemoryAndGateBlockedWhenDurableReleaseFails` 자신 | n/a | yes |

이 함수는 테스트 자체이므로 검증 주체와 대상이 같다. a083이 바꾼 것은 credit·diff의
as-of와 관측 사이의 시계 진행뿐이고 단언은 그대로다.
