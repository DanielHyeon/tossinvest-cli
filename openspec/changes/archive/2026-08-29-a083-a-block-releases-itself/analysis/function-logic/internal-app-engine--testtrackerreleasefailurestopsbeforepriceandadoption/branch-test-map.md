# Branch Test Map: `TestTrackerReleaseFailureStopsBeforePriceAndAdoption`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (403) `if` — if _, _, err := h.journal.EnterReconcile(context.Background(), journal.EnterRe… | `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` 자신 | n/a | yes |
| B2 | (409) `if` — if err := h.tracker.Restore(context.Background()); err != nil | `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` 자신 | n/a | yes |
| B3 | (419) `if` — if cycle.Err == nil | `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` 자신 | n/a | yes |
| B4 | (422) `if` — if cycle.Blocked != 1 || cycle.Released != 0 | `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` 자신 | n/a | yes |
| B5 | (425) `if` — if cycle.Adopted != 0 || cycle.Unmanaged != 0 || h.prices.calls != 0 | `TestTrackerReleaseFailureStopsBeforePriceAndAdoption` 자신 | n/a | yes |

이 함수는 테스트 자체이므로 검증 주체와 대상이 같다. a083이 바꾼 것은 credit·diff의
as-of와 관측 사이의 시계 진행뿐이고 단언은 그대로다.
