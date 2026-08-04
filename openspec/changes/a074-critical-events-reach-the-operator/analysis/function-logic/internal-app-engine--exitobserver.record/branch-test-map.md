# Branch Test Map: `ExitObserver.record`

편집 후 분기는 11개에서 12개가 되었다. 새 분기는 B12 — `ErrExitSnapshotQuarantined`
sentinel — 이며 `ErrProposalPending`(B11) 뒤에 온다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `quote.FetchedAt`가 zero → 관측 출처 `cycle` | 기존 exitloop 테스트 | no | pass |
| B2 | 그 외 → 관측 출처 `quote_fetched_at` | 기존 exitloop 테스트 | no | pass |
| B3 | 전량청산/취소 필요 시 심볼 정리 | 기존 exitloop 테스트 | no | pass |
| B4 | 정리 실패 → 반환 | 기존 exitloop 테스트 | no | pass |
| B5 | 정리 미완료 → 지연 기록, 발의 억제 | 기존 exitloop 테스트 | no | pass |
| B6 | 정리 완료 → 지연 시계 해제 | 기존 exitloop 테스트 | no | pass |
| B7 | orderable → intent 생성과 발의 부착 | 기존 exitloop 테스트 | no | pass |
| B8 | decision id에서 intent id를 못 만들면 새로 발급 | 기존 exitloop 테스트 | no | pass |
| B9 | 판정 기록 실패 | `TestAnAmbiguousRecoveryQuarantineIsAnnouncedInTheSameCycle` | no | pass |
| B10 | `ErrProposalPending`은 여전히 조용한 보류 | `internal/journal` 발의 수명주기 테스트 | no | pass |
| B11 | 무장되지 않으면 제출 없음 | 기존 exitloop 테스트 | no | pass |
| B12 | 격리 sentinel → 같은 사이클에 생성 이벤트, error는 그대로 반환 | `TestAnAmbiguousRecoveryQuarantineIsAnnouncedInTheSameCycle` | no | pass |

## 추가 시나리오 (분기 밖의 계약)

| 계약 | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| 무개입 | 격리 알림이 보호 상태도 주문 경로도 건드리지 않는다 | `TestAQuarantineAnnouncementWritesNothing` | no | pass |
| 사이클 | 격리 생성 사이클이 실패로 보고된다 | `TestRunReportsAFailedCycle` | **yes** (M1) | pass |
| latch | `record`와 `workingSet`이 같은 격리를 두 번 알리지 않는다 | `TestAQuarantineIsAnnouncedOnceAcrossBothPaths` | **yes** (M3) | pass |

`ErrExitSnapshotQuarantined` 되읽기가 실패해도 원래 error가 그대로 반환된다는 것은
`announceQuarantineFromLedger`의 구조가 보장한다 — 그 함수는 값을 반환하지 않고
호출자는 반환문을 조건에 걸지 않는다.
