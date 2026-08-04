# Branch Test Map: `fullSettingsHarness`

분기 없는 함수다. 계약은 "모든 seam이 배선된 콘솔"을 만든다는 것 하나이고,
a065는 그 목록에 하나를 더했다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 모든 seam이 배선된 콘솔을 만든다 (행복 경로) | `TestEveryCardShowsWhatChangesAndWhen` | **yes** | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 덮어쓰기 | tweak이 개별 seam을 nil로 되돌린다 | `TestAnUnwiredAlertSeamSaysWhyInsteadOfOfferingAButton` | no | pass |
