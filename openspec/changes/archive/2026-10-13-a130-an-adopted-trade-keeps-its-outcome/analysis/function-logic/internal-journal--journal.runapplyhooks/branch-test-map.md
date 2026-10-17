# Branch Test Map: `Journal.runApplyHooks`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 분기 몸체를 실행한 시험(`analysis/verify/btm-coverage.md` 측정, 시험 하나씩 커버리지).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:276` `if hooks.Project == nil && hooks.Campaign == nil && hooks.Exit == nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B2 | `:283` `if hooks.Project != nil {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40 | no | yes |
| B3 | `:284` `if err := hooks.Project(ctx, handle, fill); err != nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B4 | `:288` `if hooks.Campaign != nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B5 | `:289` `if err := hooks.Campaign(ctx, handle, fill); err != nil {` | 패키지 전체 시험이 실행(후보 패턴 밖 — 개별 이름 미측정) | no | yes |
| B6 | `:293` `if err := j.applyRiskBucketOwnerBindingInTx(ctx, tx, fill); err != nil {` | not-applicable: 오류 반환 — SQLite 읽기 · 쓰기 · 커밋 실패(또는 열기 실패)에서만 열림; 건강한 원장에서 도달 불가이고 결함 주입 시험 없음. a130 무편집 갈래(편집은 이 함수의 다른 줄) | no | n/a |
| B7 | `:297` `if hooks.Exit != nil {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AFoldInBeforeOrAtTheAdoptionIsNotALot` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` 외 40 | no | yes |
| B8 | `:298` `if err := hooks.Exit(ctx, handle, fill); err != nil {` | `TestA130ARolledBackCloseReportsNothing` | no | yes |
