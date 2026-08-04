# Branch Test Map: `runConsole`

편집은 분기를 추가하지 않았다. 구조체 리터럴에 필드 한 줄을 더한 것이고, 필드 대입은
조건이 아니다. a075 당시의 36개 분기는 같은 조건·같은 순서로 남아 있다. 통합된 현재
함수의 B37–B41은 이후 a072가 추가한 strategy runtime projection endpoint 탐색이며
a075의 알림 seam과 독립적이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `ctx == nil` → 기본 context | 기존 console 테스트 | no | pass |
| B2 | 프로필 해석 실패 | 기존 console 테스트 | no | pass |
| B3 | 옵션 검증 실패 ① | 기존 console 테스트 | no | pass |
| B4 | 옵션 검증 실패 ② | 기존 console 테스트 | no | pass |
| B5 | 옵션 검증 실패 ③ | 기존 console 테스트 | no | pass |
| B6 | 옵션 검증 실패 ④ | 기존 console 테스트 | no | pass |
| B7 | 옵션 검증 실패 ⑤ | 기존 console 테스트 | no | pass |
| B8 | remote 옵션 해석 실패 | 기존 remote 테스트 | no | pass |
| B9 | 원장 경로 있음 | 기존 console 테스트 | no | pass |
| B10 | 원장 열기 실패 ① | 기존 console 테스트 | no | pass |
| B11 | 원장 열림 ① | 기존 console 테스트 | no | pass |
| B12 | 원장 열기 실패 ② | 기존 console 테스트 | no | pass |
| B13 | 원장 열림 ② | 기존 console 테스트 | no | pass |
| B14 | 엔진 journal 디렉터리 있음 | 기존 a059 테스트 | no | pass |
| B15 | 엔진 journal 디렉터리 없음 | 기존 a059 테스트 | no | pass |
| B16 | 컨테이너 → self-update 미배선 | `TestContainerBuildsDoNotStageALocalUpdate` 계열 | no | pass |
| B17 | self 경로 실패 | 기존 update 테스트 | no | pass |
| B18 | self 경로 실패(같은 절) | 기존 update 테스트 | no | pass |
| B19 | self 경로 확보 | 기존 update 테스트 | no | pass |
| B20 | candidate 확인 실패 | 기존 update 테스트 | no | pass |
| B21 | candidate 확인 성공 | 기존 update 테스트 | no | pass |
| B22 | updater 생성 실패 | 기존 update 테스트 | no | pass |
| B23 | updater 생성 성공 | 기존 update 테스트 | no | pass |
| B24 | updater 있음 → 업데이트 seam 주입 | 기존 update 테스트 | no | pass |
| B25 | 릴리스 검사 실패 | 기존 update 테스트 | no | pass |
| B26 | 릴리스 검사 성공 | 기존 update 테스트 | no | pass |
| B27 | 엔진 디렉터리 있음 → 마커 경로 | 기존 a059 테스트 | no | pass |
| B28 | 마커 해석 실패 | 기존 a059 테스트 | no | pass |
| B29 | 자동 시작 seam 있음 | 기존 autostart 테스트 | no | pass |
| B30 | 기동 결과 문구 있음 | 기존 autostart 테스트 | no | pass |
| B31 | 엔진 디렉터리 있음 → 제어 평면 탐색 | 기존 a059 테스트 | no | pass |
| B32 | descriptor 있음 → dial | 기존 제어 평면 테스트 | no | pass |
| B33 | stat 오류가 NotExist가 아님 | 기존 제어 평면 테스트 | no | pass |
| B34 | dial 실패 | 기존 제어 평면 테스트 | no | pass |
| B35 | dial 성공 → commander 주입 | 기존 제어 평면 테스트 | no | pass |
| B36 | stat 오류(같은 절) | 기존 제어 평면 테스트 | no | pass |
| B37 | strategy projection descriptor 있음 → dial | strategy runtime integration 테스트 | no | pass |
| B38 | strategy descriptor stat 오류 분기 | strategy runtime integration 테스트 | no | pass |
| B39 | strategy projection dial 실패 → dormant 유지 | strategy runtime integration 테스트 | no | pass |
| B40 | strategy projection dial 성공 → reader 주입 | strategy runtime integration 테스트 | no | pass |
| B41 | strategy descriptor stat 오류가 NotExist가 아님 → 사유 기록 | strategy runtime integration 테스트 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 주입 | 프로필이 해석되면 알림 seam이 주입된다 | `TestEnablingCreatesAChannelAndWritesTheThreeKeys` | no | pass |
| typed nil | 프로필이 없으면 인터페이스에 typed nil이 들어가지 않는다 | `TestTheNotificationSeamIsAbsentRatherThanTypedNil` | no | pass |
| 능력 | 주입된 필드가 콘솔 능력 목록에 열거된다 | `TestEveryCapabilityTheConsoleReceivesIsEnumeratedAndDeclaresNothingButReads` | **yes** | pass |
| 기동 | 콘솔이 여전히 뜬다 | 기존 console 기동 테스트 | no | pass |

기존 36개 분기에 RED가 없는 것은 편집이 그것들을 건드리지 않았기 때문이며, 그
사실 자체가 이 표의 주장이다. 새 동작의 RED는 seam 자신의 파일(`notificationsettings.go`)과
화면(`settings_notifications.go`)의 Branch Test Map이 아니라 — 그 둘은 **새 파일**이라
기존 함수 편집이 아니다 — `a075_notification_seam_test.go`와
`a075_notification_card_test.go`가 직접 관측한다.
