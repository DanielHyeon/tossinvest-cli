# Branch Test Map: `recordGateSettings`

편집 후 분기는 3개에서 6개가 되었다. 새 분기 셋(B2·B3·B4)은 전부 알림 detail 문자열의
조립이며, 기존 순회(B5)와 실패 전파(B6)는 그대로 새 항목에도 적용된다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 거부된 adoption 블록의 사유가 detail에 | 기존 interlock 테스트 | no | pass |
| B2 | 거부된 알림 블록의 사유가 detail에 | `TestARefusedNotificationBlockIsAudited` | no | pass |
| B3 | 거부되지 않은 경우 | `TestNotificationSettingsAreAudited` | no | pass |
| B4 | 공개 ntfy 서비스를 향하면 경고 detail | `TestNotificationSettingsAreAudited` | no | pass |
| B5 | 모든 항목이 선언 순서대로 기록된다 | `TestExistingAuditEntriesKeepTheirOrder` | no | pass |
| B6 | audit write 실패 → 기동 거부 | 기존 interlock 테스트 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| §0.8 | topic 값도 token 값도 audit에 나타나지 않는다 | `TestTheAuditTrailCarriesNoNotificationSecret` | **yes** (M6) | pass |
| 순서 | 기존 11항목의 순서와 이름이 그대로다 | `TestExistingAuditEntriesKeepTheirOrder` | no | pass |
| 항목 | 알림 4항목이 기록된다 | `TestNotificationSettingsAreAudited` | no | pass |

**M6의 RED는 실제 관측이다** (2026-08-04): `topic_configured` 항목을 topic 값 자체로
바꾸자 `TestTheAuditTrailCarriesNoNotificationSecret`이 그 줄을 그대로 인용하며
실패했다.
