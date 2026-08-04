# Branch Test Map: `Console.settingsView`

편집 후 분기는 9개에서 11개가 되었다. 새 분기는 B9·B10 — 알림 seam의 배선 여부와
읽기 실패 — 이며 `EngineBoot`(B7·B8) 뒤, `SystemUpdater`(B11) 앞에 온다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 편입 seam 배선됨 | 기존 settings 테스트 | no | pass |
| B2 | 편입 읽기 실패 | 기존 settings 테스트 | no | pass |
| B3 | 한도 seam 배선됨 | 기존 settings 테스트 | no | pass |
| B4 | 한도 읽기 실패 | 기존 settings 테스트 | no | pass |
| B5 | 거래 정책 seam 배선됨 | 기존 settings 테스트 | no | pass |
| B6 | 거래 정책 읽기 실패 | 기존 settings 테스트 | no | pass |
| B7 | 자동 시작 seam 배선됨 | 기존 settings 테스트 | no | pass |
| B8 | 자동 시작 읽기 실패 | 기존 settings 테스트 | no | pass |
| B9 | 알림 seam 배선됨 → 카드가 버튼을 렌더 | `TestTheAlertCardOffersOneButtonWhenAlertsAreOff` | no | pass |
| B10 | 알림 읽기 실패 → 사유만 렌더, 다른 카드는 산다 | `TestAnUnreadableAlertConfigDoesNotTakeTheTabDown` | no | pass |
| B11 | 업데이트 seam 배선됨 | 기존 settings 테스트 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 미배선 | seam 없는 빌드는 사유를 렌더하고 버튼을 렌더하지 않는다 | `TestAnUnwiredAlertSeamSaysWhyInsteadOfOfferingAButton` | no | pass |
| 카드 표준 | 모든 카드가 적용 후 preview를 갖는다 | `TestEveryCardShowsWhatChangesAndWhen` | **yes** | pass |
| 카드 표준 | 모든 카드가 버튼이거나 이름 붙은 사유다 | `TestEveryCardEitherSavesOrSaysWhyNot` | no | pass |
| §0.8 | 채널은 본문에만 나오고 리다이렉트 URL에는 없다 | `TestTheNoticeNeverCarriesTheChannel` | **yes** (M4) | pass |
| 구독 | 카드가 구독 주소를 표시한다 | `TestTheCardShowsTheSubscribeAddress` | no | pass |
| §0.7 | 토큰 입력란이 없다 | `TestThereIsNoTokenInputOnTheAlertCard` | no | pass |

`TestEveryCardShowsWhatChangesAndWhen`의 RED는 실제 관측이다 (2026-08-04): 카드를 먼저
붙이고 `fullSettingsHarness`에 seam을 주입하지 않은 상태에서
`/settings/tools/notifications renders no 적용 후 preview`로 실패했다 — 미배선 카드는
`cardblocked`를 렌더하므로 preview가 없다.
