# Branch Test Map: `TestWriteThroughRecordsTheStateWithItsEvidence`

AST의 모든 분기를 1행씩 덮는다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (721) `if` — if _, err := tracker.Observe(ctx, mismatchDiff("AAPL", "10", "4")); err != nil | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B2 | (725) `if` — if err != nil | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B3 | (728) `if` — if len(active) != 1 | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B4 | (732) `if` — if state.Symbol != "AAPL" || state.Cause != journal.ReconcileCauseQuantityMism… | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B5 | (735) `if` — if state.Evidence == "" | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B6 | (743) `if` — if _, err := tracker.Observe(ctx, reconcile.Diff{AccountRef: "acct-7", Matched… | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B7 | (746) `if` — if active, err = j.ActiveReconcileStates(ctx); err != nil | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B8 | (749) `if` — if len(active) != 1 | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B9 | (758) `if` — if _, err := tracker.Observe(ctx, cleanDiffAt(clk)); err != nil | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B10 | (761) `if` — if active, err = j.ActiveReconcileStates(ctx); err != nil | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B11 | (764) `if` — if len(active) != 0 | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B12 | (768) `if` — if err != nil | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |
| B13 | (771) `if` — if len(history) != 1 || history[0].ReleaseCause != journal.ReconcileReleaseAdj… | `TestWriteThroughRecordsTheStateWithItsEvidence` 자신 | n/a | yes |

이 함수는 테스트 자체이므로 검증 주체와 대상이 같다. a083이 바꾼 것은 credit·diff의
as-of와 관측 사이의 시계 진행뿐이고 단언은 그대로다.
