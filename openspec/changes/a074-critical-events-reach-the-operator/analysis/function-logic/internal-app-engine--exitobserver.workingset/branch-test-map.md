# Branch Test Map: `ExitObserver.workingSet`

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Positions 실패 → 사이클 실패 | 기존 exitloop 테스트 | no | pass |
| B2 | exit state 읽기 실패 → 사이클 실패 | 기존 exitloop 테스트 | no | pass |
| B3 | state 결과 인덱싱 | 기존 exitloop 테스트 | no | pass |
| B4 | 포지션 순회 | 기존 exitloop 테스트 | no | pass |
| B5 | CLOSED/0수량 skip | 기존 exitloop 테스트 | no | pass |
| B6 | 미관리 보유 알림 후 skip | 기존 `unmanaged` 테스트 | no | pass |
| B7 | 열린 state 없음 → 개설 | 기존 exitloop 테스트 | no | pass |
| B8 | 개설 실패 → skip | 기존 exitloop 테스트 | no | pass |
| B9 | 첫 실패만 `cycle.Err`에 담긴다 | 기존 exitloop 테스트 | no | pass |
| B10 | 완결된 state → skip | 기존 exitloop 테스트 | no | pass |
| B11 | corrupt snapshot 격리 생성 → 생성 이벤트 발행 (latch로 1회) | `TestACorruptSnapshotQuarantineIsAnnouncedWhenItIsCreated` | **yes** (M3) | pass |
| B12 | 격리 write 실패 → skip | 기존 exitloop 테스트 | no | pass |
| B13 | 첫 실패만 `cycle.Err`에 담긴다 | 기존 exitloop 테스트 | no | pass |
| B14 | 활성 격리 read 실패 → skip | 기존 exitloop 테스트 | no | pass |
| B15 | 이미 활성인 격리 → 판정 거부, 생성 이벤트 **없음** | `TestAnAlreadyActiveQuarantineIsNotAnnouncedAgain` | no | pass |
| B16 | 첫 실패만 `cycle.Err`에 담긴다 | 기존 exitloop 테스트 | no | pass |
| B17 | 활성 격리 행이 `refused`에 들어간다 | `TestAnAlreadyActiveQuarantineIsNotAnnouncedAgain` | no | pass |
| B18 | identity 해석 실패 격리 생성 → 생성 이벤트 발행 | `TestALegacyIdentityQuarantineIsAnnouncedWhenItIsCreated` | no | pass |
| B19 | 격리 write 실패 → skip | 기존 exitloop 테스트 | no | pass |
| B20 | 첫 실패만 `cycle.Err`에 담긴다 | 기존 exitloop 테스트 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 순서 | 격리 행이 여전히 반환 집합의 마지막에 온다 | `TestCorruptPositionAlertCannotDelayAnotherEmergencyExit` (기존) | no | pass |
| 무개입 | 알림 추가가 반환 집합의 내용을 바꾸지 않는다 | `TestAnnouncingAQuarantineDoesNotChangeTheWorkingSet` | no | pass |
| latch | 판정 거부가 이미 latch된 포지션도 격리 이벤트를 받는다 | `TestANewQuarantineVersionIsAnnouncedAgain` | **yes** (M4) | pass |

M3·M4의 RED는 실제 관측이다 (2026-08-04). latch 제거는 3사이클에서 3회 발행으로,
latch 키의 version 제거는 재격리 미발행으로 각각 실패했다.
