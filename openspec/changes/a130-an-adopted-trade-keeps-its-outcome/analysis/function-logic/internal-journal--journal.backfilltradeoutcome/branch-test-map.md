# Branch Test Map: `Journal.BackfillTradeOutcome`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 분기 몸체를 실행한 시험(`analysis/verify/btm-coverage.md` 측정, 시험 하나씩 커버리지).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:892` `if _, err := j.TradeOutcomeOf(ctx, id); err == nil {` | `TestABackfillRecoversTheGapAndRefusesToRewriteIt` | no | yes |
| B2 | `:894` `} else if !errors.Is(err, ErrTradeOutcomeNotFound) {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B3 | `:894` `} else if !errors.Is(err, ErrTradeOutcomeNotFound) {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B4 | `:899` `if reason != "" {` | `TestAPartlyEngineSoldAdoptedPositionFreezesNothing` · `TestAnAdoptedPositionWithNoEngineSellFreezesNothing` | no | yes |
| B5 | `:903` `if _, err := j.db.ExecContext(ctx, `` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B6 | `:913` `if isUniqueViolation(err) {` | not-applicable: 앞의 TradeOutcomeOf 확인과 INSERT 사이에 다른 쓰기가 끼어야 열림(경합 전용); a130 무편집 갈래 | no | n/a |
