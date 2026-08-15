# Branch Test Map: `strategyRuntimeReaderFor`

편집 **전** 상태다(a109 base `016da624`). 네 분기의 **판정**은 a108 이 이미 핀했다 —
a109 가 여는 것은 「그 판정 뒤에 다시 시도하는 경로가 없다」는 결함이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 엔진 디렉터리를 못 정하면 경고 + dormant | 미고정 (경로 해석 실패는 환경 이상) | no | no |
| B2 | descriptor 부재는 조용한 dormant | `TestADialFailureRendersUnavailableRatherThanNotConfigured` · `TestAnAbsentDescriptorAndADeadOneBootTheSame` | no | yes |
| B3 | 조사 불가능(ENOTDIR)은 경고 + dormant, **fatal 아님** | `TestAnUninspectableDescriptorDegradesLikeTheConsole` | yes (a108 D4-2 뒤집기) | yes |
| B4 | dial 실패는 sentinel → unavailable | `TestADeadDescriptorDoesNotStopTheDaemon` · `TestASocketFileWithNoOwnerDegradesTheDaemon` · `TestADialFailureRendersUnavailableRatherThanNotConfigured` | yes (a108) | yes |

**a109 §2.3 이 더하는 시나리오**(이 표의 분기에는 없다 — wrapper 의 성질이다):

| 시나리오 | Test | RED observed | GREEN observed |
|---|---|---|---|
| 냉부팅 순서: httpapi 가 먼저 뜨고 엔진이 나중에 endpoint 를 발행 → 재시작 없이 회복 | a109 §2.3 (`cmd/tossctl/a109_*_test.go`) | pending | pending |
| 가동 중 엔진 재시작(새 socket·새 토큰) → live 부착이 실패하기 시작 → 회복 | a109 §2.3 | pending | pending |
| rate limit 안에서만 시도한다(창 안 재시도 없음) | a109 §2.3 | pending | pending |
| 요청 goroutine 이 dial·probe 를 부르지 않는다 | a109 §2.3 | pending | pending |
| 재부착 전 화면 값은 dormant/unavailable 구분을 유지한다 | a109 §2.3 | pending | pending |
