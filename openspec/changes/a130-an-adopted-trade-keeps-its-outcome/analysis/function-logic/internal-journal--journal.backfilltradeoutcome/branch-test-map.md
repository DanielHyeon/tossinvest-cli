# Branch Test Map: `Journal.BackfillTradeOutcome`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 함수의 a130 시험 · 기존 시험 묶음(무편집 갈래는 기존 시험).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:892` `if _, err := j.TradeOutcomeOf(ctx, id); err == nil {` | 기존 `TestABackfillRecoversTheGapAndRefusesToRewriteIt` · `TestAPartlyEngineSoldAdoptedPositionFreezesNothing` | no | yes |
| B2 | `:894` `} else if !errors.Is(err, ErrTradeOutcomeNotFound) {` | 기존 `TestABackfillRecoversTheGapAndRefusesToRewriteIt` · `TestAPartlyEngineSoldAdoptedPositionFreezesNothing` | no | yes |
| B3 | `:894` `} else if !errors.Is(err, ErrTradeOutcomeNotFound) {` | 기존 `TestABackfillRecoversTheGapAndRefusesToRewriteIt` · `TestAPartlyEngineSoldAdoptedPositionFreezesNothing` | no | yes |
| B4 | `:899` `if reason != "" {` | 기존 `TestABackfillRecoversTheGapAndRefusesToRewriteIt` · `TestAPartlyEngineSoldAdoptedPositionFreezesNothing` | no | yes |
| B5 | `:903` `if _, err := j.db.ExecContext(ctx, `` | 기존 `TestABackfillRecoversTheGapAndRefusesToRewriteIt` · `TestAPartlyEngineSoldAdoptedPositionFreezesNothing` | no | yes |
| B6 | `:913` `if isUniqueViolation(err) {` | 기존 `TestABackfillRecoversTheGapAndRefusesToRewriteIt` · `TestAPartlyEngineSoldAdoptedPositionFreezesNothing` | no | yes |
