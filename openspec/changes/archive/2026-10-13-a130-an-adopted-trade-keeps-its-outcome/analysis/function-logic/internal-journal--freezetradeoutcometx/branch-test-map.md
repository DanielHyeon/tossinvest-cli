# Branch Test Map: `freezeTradeOutcomeTx`

편집 뒤 좌표(4판 GREEN). 분기별 RED 는 따로 재지 않았음(no) — 4판 시험 전체가 base 에서 컴파일 RED(`analysis/red/red-4.txt`), 조작 행 모양은 행동 RED(red-2 · red-4). GREEN = 패키지 시험 통과(`analysis/verify/receipts.md`). Test 칸은 그 분기 몸체를 실행한 시험(`analysis/verify/btm-coverage.md` 측정, 시험 하나씩 커버리지).

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:131` `if reason != "" {` | `TestA130AFailedAmendDoesNotDateALocalSell` · `TestA130AGrownPositionIsRefusedInsideTheFillTransaction` · `TestA130AGrownPositionPricesEachLotAtItsMinuteBar` · `TestA130ALaterInstanceFlattenIsNotThisTripsSell` 외 21 | no | yes |
| B2 | `:144` `if err != nil {` | not-applicable: INSERT 실패 — 행 존재는 앞에서 확인했고 같은 트랜잭션이 쓰기 잠금을 쥠; 사유 코드만 a130 | no | n/a |
