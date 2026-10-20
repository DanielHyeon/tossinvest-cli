# Branch Test Map: `mergeEngine`

편집은 분기를 추가하지 않았다 — `mergeNotifications` 호출은 문이지 조건이 아니고,
`mergeAdoption`이 같은 자리에 이미 그렇게 있다. 분기는 5개 그대로다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | engine 블록 없음 → 기본값(전부 off) | `TestAConfigWithNoNotificationsBlockIsUnchanged` | no | pass |
| B2 | autostart 병합 | 기존 config 테스트 | no | pass |
| B3 | exit_policy 병합·거부 | 기존 config 테스트 | no | pass |
| B4 | automation_gate 없음 → 조기 반환 | `TestNotificationsMergeWithoutAnAutomationGate` | no | pass |
| B5 | gate.enabled 병합 | 기존 config 테스트 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| §0.2 | 알림 블록 없음 → zero value, 오늘과 동일 | `TestAConfigWithNoNotificationsBlockIsUnchanged` | no | pass |
| 병합 | 알림 블록이 병합된다 | `TestNotificationsAreMerged` | no | pass |
| 전부-또는-전무 | 거부된 블록은 통째로 0이 된다 | `TestNotificationsRefuseANonHTTPBaseURL` | **yes** (M5) | pass |
| 순서 | B4의 조기 반환 앞에서 병합된다 | `TestNotificationsMergeWithoutAnAutomationGate` | no | pass |
| §0.8 | 토큰을 담을 필드가 없다 | `TestNotificationsHaveNoFieldForASecret` | no | pass |
| 순수성 | 이 패키지는 환경변수를 읽지 않는다 | `TestConfigReadsNoEnvironment` | no | pass |

**M5의 RED는 실제 관측이다** (2026-08-04): 거부된 블록을 통째로 0으로 만드는 대신
`Rejected`만 채워 넣자 `TestNotificationsRefuseANonHTTPBaseURL`이
`{Enabled:true BaseURL:ftp://elsewhere …}`를 인용하며 실패했다.
