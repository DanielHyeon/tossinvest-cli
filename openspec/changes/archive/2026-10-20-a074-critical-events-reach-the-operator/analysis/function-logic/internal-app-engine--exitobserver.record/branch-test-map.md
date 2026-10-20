# Branch Test Map: `ExitObserver.record`

a074 편집 당시(base 448dfeb1 기준) 분기는 11개에서 12개가 되었다. a074 의 새 분기는
`ErrExitSnapshotQuarantined` sentinel 이며 `ErrProposalPending` 뒤에 온다 — 현재 번호로 B17·B16.

**2026-10-20 refresh.** 이 표의 번호는 HEAD 3415cf89 의 `ast.json`(분기 18) 기준이다. base 뒤 이웃
change 가 이 함수에 분기 6개를 더했다(B1·B2·B5·B6·B8·B9). 옛 번호 → 새 번호 대응과 그 근거는
`function-logic-map.md` 「Refresh (2026-10-20)」 절에 있다. 이웃 분기 행은 **a074 의 관측이 아니다** —
그 분기를 지나는 현존 시험을 그 분기를 만든 change 의 번들에서 옮겨 적었고 RED 는 `n/a` 다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | (이웃 a111) 시세가 쓸 수 없으면 판정 보류 | `TestA094TheParkCheckChangesNoOutcome` (a091 번들 인용) | n/a | pass |
| B2 | (이웃) 관측 출처가 비었을 때만 출처를 정한다 | `TestABaselineBreachProposesTheWholePosition` (a091 번들 인용) | n/a | pass |
| B3 | `quote.FetchedAt`가 zero → 관측 출처 `cycle` | 기존 exitloop 테스트 | no | pass |
| B4 | 그 외 → 관측 출처 `quote_fetched_at` | 기존 exitloop 테스트 | no | pass |
| B5 | (이웃 a150) 승인 floor 를 먼저 커밋하고 이번 주기 발의 억제 | `TestA150MismatchedStopCannotWeakenAdmittedProtection` (a150, `ArmSuppressedFloorReconciliation` 단언) | n/a | pass |
| B6 | (이웃 a150) floor 갱신이 없을 때의 else-if 사슬 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` (a091 번들 인용) | n/a | pass |
| B7 | 전량청산/취소 필요 시 심볼 정리 | 기존 exitloop 테스트 | no | pass |
| B8 | (이웃 a084) 재판정 패스의 비보호 발의는 정리하지 않고 억제 | `TestCrossingTheFirstTakeProfitKeepsThePositionUnderJudgement` (a091 번들 인용) | n/a | pass |
| B9 | (이웃 a084) 재판정이 아니면 정리 시도 | `TestAWorkingEntryIsCancelledBeforeTheLiquidation` (a091 번들 인용) | n/a | pass |
| B10 | 정리 실패 → 반환 | 기존 exitloop 테스트 | no | pass |
| B11 | 정리 미완료 → 지연 기록, 발의 억제 | 기존 exitloop 테스트 | no | pass |
| B12 | 정리 완료 → 지연 시계 해제 | 기존 exitloop 테스트 | no | pass |
| B13 | orderable → intent 생성과 발의 부착 | 기존 exitloop 테스트 | no | pass |
| B14 | decision id에서 intent id를 못 만들면 새로 발급 | 기존 exitloop 테스트 | no | pass |
| B15 | 판정 기록 실패 | `TestAnAmbiguousRecoveryQuarantineIsAnnouncedInTheSameCycle` | no | pass |
| B16 | `ErrProposalPending`은 여전히 조용한 보류 | `internal/journal` 발의 수명주기 테스트 | no | pass |
| B17 | 격리 sentinel → 같은 사이클에 생성 이벤트, error는 그대로 반환 | `TestAnAmbiguousRecoveryQuarantineIsAnnouncedInTheSameCycle` | no | pass |
| B18 | 무장되지 않으면 제출 없음 | 기존 exitloop 테스트 | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 무개입 | 격리 알림이 보호 상태도 주문 경로도 건드리지 않는다 | `TestAQuarantineAnnouncementWritesNothing` | no | pass |
| 사이클 | 격리 생성 사이클이 실패로 보고된다 | `TestRunReportsAFailedCycle` | **yes** (M1) | pass |
| latch | `record`와 `workingSet`이 같은 격리를 두 번 알리지 않는다 | `TestAQuarantineIsAnnouncedOnceAcrossBothPaths` | **yes** (M3) | pass |

`ErrExitSnapshotQuarantined` 되읽기가 실패해도 원래 error가 그대로 반환된다는 것은
`announceQuarantineFromLedger`의 구조가 보장한다 — 그 함수는 값을 반환하지 않고
호출자는 반환문을 조건에 걸지 않는다.
