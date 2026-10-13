# Branch Test Map: `Journal.RecordFill`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 함수의 a130 시험 · 기존 시험 묶음(무편집 갈래는 기존 시험).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:314` `if strings.TrimSpace(obs.OrderID) == "" {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B2 | `:318` `if !scope.complete() && !scope.legacyUnscoped() {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B3 | `:324` `if err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B4 | `:328` `if math.IsNaN(filled) \|\| math.IsInf(filled, 0) {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B5 | `:340` `if err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B6 | `:346` `switch {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B7 | `:347` `case err == nil:` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B8 | `:348` `case errors.Is(err, ErrFillNotFound):` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B9 | `:350` `default:` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B10 | `:356` `if ownerErr != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B11 | `:364` `if hadPrev && scope.complete() {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B12 | `:368` `if owners == 1 && prev.CommittedAt <= ownershipAt {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B13 | `:374` `if err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B14 | `:380` `if hadPrev {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B15 | `:385` `if refusal := classifyFillRefusal(obs, prev, hadPrev, filled, prevFilled); refus` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B16 | `:392` `if err := markFillRefused(ctx, tx, scope, obs, refusal, now, hadPrev); err != ni` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B17 | `:401` `if err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B18 | `:405` `if err := tx.Commit(); err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B19 | `:413` `if nearlyZero(delta, filled) {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B20 | `:420` `if hadPrev && delta == 0 && sameSnapshot(obs, prev) {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B21 | `:421` `if err := tx.Commit(); err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B22 | `:437` `if err := upsertFillSnapshot(ctx, tx, scope, obs, now); err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B23 | `:441` `if correction {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B24 | `:442` `if err := recordExecutionCorrection(ctx, tx, ExecutionCorrection{` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B25 | `:464` `if delta > 0 {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B26 | `:465` `if _, err := tx.ExecContext(ctx,` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B27 | `:504` `if err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B28 | `:512` `if obs.Terminal && locallyOwned {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B29 | `:516` `if err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B30 | `:526` `if locallyOwned {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B31 | `:536` `} else if err := latchRiskBucketFillFailureForScope(ctx, tx, applied,` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B32 | `:527` `if err := j.applyRiskBucketFillInTx(ctx, tx, applied); err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B33 | `:530` `if err := j.releaseTerminalRiskBucketOrderInTx(ctx, tx, applied); err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B34 | `:533` `if err := applyWeeklyReservationLifecycleInTx(ctx, tx, applied); err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B35 | `:536` `} else if err := latchRiskBucketFillFailureForScope(ctx, tx, applied,` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B36 | `:550` `if locallyOwned {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B37 | `:551` `if err := j.runApplyHooks(withOutcomeRefusals(ctx, &refusals), tx, applied); err` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
| B38 | `:556` `if err := tx.Commit(); err != nil {` | 기존 체결 시험 · `TestA130RecordFillReportsRefusalsOnlyAfterItsCommit` · `TestA130ARolledBackCloseReportsNothing` | no | yes |
