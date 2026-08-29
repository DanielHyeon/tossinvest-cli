# Branch Test Map: `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (480) `range` — for _, symbol := range []string{"AAPL", "MSFT"} | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B2 | (481) `if` — if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B3 | (490) `if` — if err := tracker.Restore(ctx); err != nil | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B4 | (499) `if` — if err == nil | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B5 | (502) `if` — if len(out.Cleared) != 1 || out.Cleared[0].Symbol != "AAPL" | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B6 | (505) `if` — if blocks := tracker.Blocks(); len(blocks) != 1 || blocks[0].Symbol != "MSFT" | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B7 | (509) `if` — if err != nil | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B8 | (512) `if` — if len(active) != 1 || active[0].Symbol != "MSFT" | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B9 | (515) `if` — if gate.CheckEntryFor("us", "AAPL") != nil || gate.CheckEntryFor("us", "MSFT")… | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B10 | (522) `if` — if err != nil | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B11 | (525) `if` — if len(out.Cleared) != 1 || out.Cleared[0].Symbol != "MSFT" || out.Blocked | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |
| B12 | (528) `if` — if gate.CheckEntryFor("us", "MSFT") != nil | `TestObservePublishesOnlyDurablePartialReleasesAndRetriesTheRemainder` 자신 | n/a | yes |

이 함수는 테스트 자체이므로 검증 주체와 대상이 같다. a083이 바꾼 것은 credit·diff의
as-of와 관측 사이의 시계 진행뿐이고 단언은 그대로다.
