# Branch Test Map: `TestRestoreProjectsStatesThisTrackerDidNotEnter`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (321) `if` — if _, _, err := j.EnterReconcile(ctx, journal.EnterReconcileRequest | `TestRestoreProjectsStatesThisTrackerDidNotEnter` 자신 | n/a | yes |
| B2 | (331) `if` — if err := tracker.Restore(ctx); err != nil | `TestRestoreProjectsStatesThisTrackerDidNotEnter` 자신 | n/a | yes |
| B3 | (336) `if` — if rejected == nil | `TestRestoreProjectsStatesThisTrackerDidNotEnter` 자신 | n/a | yes |
| B4 | (339) `if` — if rejected.Reason != execgw.ReasonReconcilePermanent | `TestRestoreProjectsStatesThisTrackerDidNotEnter` 자신 | n/a | yes |
| B5 | (346) `if` — if len(blocks) != 1 || blocks[0].Cause != journal.ReconcileCauseIdentifierConf… | `TestRestoreProjectsStatesThisTrackerDidNotEnter` 자신 | n/a | yes |
| B6 | (351) `if` — if _, err := tracker.Observe(ctx, cleanDiffAt(clk)); err != nil | `TestRestoreProjectsStatesThisTrackerDidNotEnter` 자신 | n/a | yes |
| B7 | (354) `if` — if blocks = tracker.Blocks(); len(blocks) != 1 || blocks[0].Cause != journal.R… | `TestRestoreProjectsStatesThisTrackerDidNotEnter` 자신 | n/a | yes |
| B8 | (357) `if` — if gate.CheckEntryFor("us", "AAPL") == nil | `TestRestoreProjectsStatesThisTrackerDidNotEnter` 자신 | n/a | yes |

이 함수는 테스트 자체이므로 검증 주체와 대상이 같다. a083이 바꾼 것은 credit·diff의
as-of와 관측 사이의 시계 진행뿐이고 단언은 그대로다.
