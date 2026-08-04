# Branch Test Map: `consoleNotificationSeam`

분기 하나. 프로필이 해석되면 seam을, 아니면 nil을 준다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 프로필이 해석되면 구현을, 아니면 nil 인터페이스를 반환한다 | `TestTheNotificationSeamIsAbsentRatherThanTypedNil` | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 배선 | 주입된 seam이 실제로 파일을 쓴다 | `TestEnablingCreatesAChannelAndWritesTheThreeKeys` | no | pass |
| 미배선 | seam이 nil이면 카드가 사유를 렌더하고 버튼을 렌더하지 않는다 | `TestAnUnwiredAlertSeamSaysWhyInsteadOfOfferingAButton` | no | pass |
| 능력 | 주입되는 인터페이스가 능력 목록에 열거된다 | `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads` | **yes** | pass |
