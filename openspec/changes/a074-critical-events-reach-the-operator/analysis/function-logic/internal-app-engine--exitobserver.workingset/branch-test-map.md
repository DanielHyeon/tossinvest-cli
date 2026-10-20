# Branch Test Map: `ExitObserver.workingSet`

**2026-10-20 refresh.** 이 표의 번호는 HEAD 3415cf89 의 `ast.json`(분기 31) 기준이다. a074 편집
당시(base 448dfeb1) 분기는 20개였고, base 뒤 이웃 change 가 분기 11개를 더했다(B7~B13 · B25 · B26 ·
B30 · B31). 옛 번호 → 새 번호 대응과 그 근거는 `function-logic-map.md` 「Refresh (2026-10-20)」 절에 있다.
이웃 분기 행은 **a074 의 관측이 아니다** — 그 분기를 만든 change 의 현존 시험을 적었고, 그 시험이 그
갈래에 진입하는지는 여기서 재지 않았다. RED 는 `n/a` 다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Positions 실패 → 사이클 실패 | 기존 exitloop 테스트 | no | pass |
| B2 | exit state 읽기 실패 → 사이클 실패 | 기존 exitloop 테스트 | no | pass |
| B3 | state 결과 인덱싱 | 기존 exitloop 테스트 | no | pass |
| B4 | 포지션 순회 | 기존 exitloop 테스트 | no | pass |
| B5 | CLOSED/0수량 skip | 기존 exitloop 테스트 | no | pass |
| B6 | 미관리 보유 알림 후 skip | 기존 `unmanaged` 테스트 | no | pass |
| B7 | (이웃 a150) 편입 포지션만 승인 보호를 대조한다 | `TestA150ExitSeedsAdmittedPolicyAfterConfigurationChange` | n/a | pass |
| B8 | (이웃 a150) 승인 보호와 exit 가 어긋나면 latch 경로 | `TestA150ForeignGenerationDoesNotSeedButExistingExitSurvives` | n/a | pass |
| B9 | (이웃 a150) 승인 증거 읽기 실패 | `TestA150ExistingExitRemainsManagedWhenAdmissionEvidenceUnreadable` | n/a | pass |
| B10 | (이웃 a150) 캠페인이 있을 때만 latch | `TestA150RestartAfterLatchSeedsAdmittedExit` | n/a | pass |
| B11 | (이웃 a150) latch 실패 → 사이클 실패, 기존 exit 는 관리 유지 | `TestA150ExistingExitRemainsManagedWhenLatchFails` | n/a | pass |
| B12 | (이웃 a150) 첫 실패만 `cycle.Err`에 담긴다 | `TestA150ExistingExitRemainsManagedWhenLatchFails` | n/a | pass |
| B13 | (이웃 a150) state·정책 둘 다 없으면 세고 경고 | `TestA150UnrecoverablePositionIsCountedAndWarned` | n/a | pass |
| B14 | 열린 state 없음 → 개설 | 기존 exitloop 테스트 | no | pass |
| B15 | 개설 실패 → skip | 기존 exitloop 테스트 | no | pass |
| B16 | 첫 실패만 `cycle.Err`에 담긴다 | 기존 exitloop 테스트 | no | pass |
| B17 | 완결된 state → skip | 기존 exitloop 테스트 | no | pass |
| B18 | corrupt snapshot 격리 생성 → 생성 이벤트 발행 (latch로 1회) | `TestACorruptSnapshotQuarantineIsAnnouncedWhenItIsCreated` | **yes** (M3) | pass |
| B19 | 격리 write 실패 → skip | 기존 exitloop 테스트 | no | pass |
| B20 | 첫 실패만 `cycle.Err`에 담긴다 | 기존 exitloop 테스트 | no | pass |
| B21 | 활성 격리 read 실패 → skip | 기존 exitloop 테스트 | no | pass |
| B22 | 이미 활성인 격리 → 판정 거부, 생성 이벤트 **없음** | `TestAnAlreadyActiveQuarantineIsNotAnnouncedAgain` | no | pass |
| B23 | 첫 실패만 `cycle.Err`에 담긴다 | 기존 exitloop 테스트 | no | pass |
| B24 | 활성 격리 행이 `refused`에 들어간다 (a084 뒤로는 재판정이 필요 없는 격리만) | `TestAnAlreadyActiveQuarantineIsNotAnnouncedAgain` | no | pass |
| B25 | (이웃 a084) 재판정이 필요한 활성 격리의 else 사슬 | `TestASupersededQuarantineIsReJudgedAndReleased` | n/a | pass |
| B26 | (이웃 a084) 선택기가 바뀐 격리 → 한 번 재판정 | `TestASupersededQuarantineIsReJudgedAndReleased` | n/a | pass |
| B27 | identity 해석 실패 격리 생성 → 생성 이벤트 발행 | `TestALegacyIdentityQuarantineIsAnnouncedWhenItIsCreated` | no | pass |
| B28 | 격리 write 실패 → skip | 기존 exitloop 테스트 | no | pass |
| B29 | 첫 실패만 `cycle.Err`에 담긴다 | 기존 exitloop 테스트 | no | pass |
| B30 | (이웃 a150) 같은 세대의 승인 정책이 있을 때만 floor 비교 | `TestA150MismatchedStopCannotWeakenAdmittedProtection` | n/a | pass |
| B31 | (이웃 a150) 승인 손절가가 baseline 보다 높으면 floor 로 올린다 | `TestA150MismatchedStopCannotWeakenAdmittedProtection` | n/a | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 순서 | 격리 행이 여전히 반환 집합의 마지막에 온다 | `TestCorruptPositionAlertCannotDelayAnotherEmergencyExit` (기존) | no | pass |
| 무개입 | 알림 추가가 반환 집합의 내용을 바꾸지 않는다 | `TestAnnouncingAQuarantineDoesNotChangeTheWorkingSet` | no | pass |
| latch | 판정 거부가 이미 latch된 포지션도 격리 이벤트를 받는다 | `TestANewQuarantineVersionIsAnnouncedAgain` | **yes** (M4) | pass |

M3·M4의 RED는 실제 관측이다 (2026-08-04). latch 제거는 3사이클에서 3회 발행으로,
latch 키의 version 제거는 재격리 미발행으로 각각 실패했다.
