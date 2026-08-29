# Branch Test Map: `Console.openOrdersPanelFrom`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 일반 주문 읽기가 실패하면 그 건수를 합계에 넣지 않는다 | `TestAPlainReadThatFailedIsNeverAddedIntoTheTotalEither`, `TestAWholeSeamFailureIsUnmeasuredAndNotAnEmptyAccount` | yes — I-7 변이표: 두 그룹이 잘림 플래그를 공유하게 되돌리면 `TestTheLiveCountIsANumberEvenWhenTheClosedPageWasTruncated`가 문다 | yes |
| B2 | 조건주문 읽기가 실패하면 그 건수를 합계에 넣지 않는다 | `TestAConditionalReadThatFailedIsNeverAddedIntoTheTotal`, `TestTheOpenOrderCountIsUnmeasuredAndNotZero` | yes | yes |
| B3 | (무분기 본문) 오래된 판독이 측정된 숫자로 렌더되지 않는다 — 시각·나이·TTL이 값과 같은 호흡에 붙는다 | `TestTheOverviewNeverRendersAStaleOrdersReadingAsAMeasuredNumber` | yes — I-7 변이표: 개요에서 시각·나이·TTL 표시를 제거하면 이 테스트가 문다 | yes |
| B4 | (무분기 본문) seam이 배선되면 실제 값이 뜬다 | `TestTheOverviewOpenOrdersPanelHasARealValueOnceTheSeamIsWired` | — | yes |
