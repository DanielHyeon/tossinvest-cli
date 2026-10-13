# Branch Test Map: `bindApplyHooks`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 분기 몸체를 실행한 시험(`analysis/verify/btm-coverage.md` 측정, 시험 하나씩 커버리지).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:193` `if len(refused) > 0 {` | `TestA130TheLoggedBindingCarriesTheSink` | no | yes |
| B2 | `:196` `if err := j.SetApplyHooks(journal.ApplyHooks{` | `TestA130TheLoggedBindingCarriesTheSink` · `TestBindingWiresBothApplyHooks` · `TestTheBoundExitApplierIsTheJournalsOwn` · `TestTheProjectionGuardAsksTheBindingAndNotTheData` | no | yes |
