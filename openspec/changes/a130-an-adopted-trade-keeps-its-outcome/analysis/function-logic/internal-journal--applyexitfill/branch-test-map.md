# Branch Test Map: `ApplyExitFill`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 함수의 a130 시험 · 기존 시험 묶음(무편집 갈래는 기존 시험).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:973` `if err != nil {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B2 | `:976` `if !found {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B3 | `:983` `if errors.Is(err, ErrExitStateNotFound) {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B4 | `:988` `if err != nil {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B5 | `:992` `if strings.EqualFold(strings.TrimSpace(where.Side), "SELL") {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B6 | `:995` `if err != nil {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B7 | `:999` `if cmp, err := riskcalc.CompareDecimal(next, state.TakenRatioTotal); err != nil ` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B8 | `:1002` `} else if cmp > 0 {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B9 | `:1002` `} else if cmp > 0 {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B10 | `:1003` `if err := tx.MoveTakenRatioTotal(ctx, where.PositionID, next); err != nil {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B11 | `:1009` `if fill.Terminal && state.Pending() && state.PendingIntentID != "" &&` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B12 | `:1011` `if err := tx.ResolvePending(ctx, where.PositionID); err != nil {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B13 | `:1014` `if err := appendExitEventFromHook(ctx, tx, exitEventRow{` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B14 | `:1023` `if where.PositionState == PositionClosed && !state.Completed {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B15 | `:1024` `if err := markExitCompletedTx(ctx, tx, where.PositionID); err != nil {` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
| B16 | `:1036` `if reason := freezeTradeOutcomeTx(ctx, tx, where.PositionID, tx.costs); reason !` | `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130ARolledBackCloseReportsNothing` · 기존 exit apply 시험 | no | yes |
