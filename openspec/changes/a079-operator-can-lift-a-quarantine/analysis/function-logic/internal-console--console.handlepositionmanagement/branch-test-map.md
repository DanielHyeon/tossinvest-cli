# Branch Test Map: `Console.handlePositionManagement`

번호는 구현 **후** 재생성한 AST 기준이다. a063이 추가한 분기는 B10·B11·B12(격리
capability 탐지, 조회 오류, 목록 적재)와 B15(행별 활성 격리)다. 나머지는 위치만
밀렸고 조건과 동작이 그대로다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | 설정 read seam이 없으면 문구를 남기고 계속 그린다 | 기존 `position_policy_test.go` | no | pass |
| B2 | 설정 seam이 있으면 실제로 읽는다 | 같음 | no | pass |
| B3 | 설정 읽기 실패는 화면을 막지 않는다 | 같음 | no | pass |
| B4 | 설정 읽기 성공은 desired를 채운다 | 같음 | no | pass |
| B5 | commander가 없으면 조회 전용으로 뜨고 해제 action이 없다 | `TestAnUnwiredCommanderShowsNoReleaseAction` | yes | pass |
| B6 | runtime 조회 실패는 문구만 남긴다 | 기존 | no | pass |
| B7 | runtime 조회 성공은 실효 설정을 채운다 | 기존 | no | pass |
| B8 | 대사 차단 목록을 투영한다 | 기존 | no | pass |
| B9 | 정책 목록 실패는 조기 return이다 | 기존 | no | pass |
| B10 | 격리 capability가 있을 때만 격리를 읽는다 | `TestAnUnwiredCommanderShowsNoReleaseAction`, `TestTheConsoleOffersReleaseOnlyForAQuarantinedRow` | yes | pass |
| B11 | 격리 조회 실패는 화면을 막지 않고 이유를 표시한다 | `TestAFailedQuarantineReadDoesNotBlankTheWholeScreen` | yes | pass |
| B12 | 읽은 격리를 position id로 적재한다 | `TestTheConsoleOffersReleaseOnlyForAQuarantinedRow` | yes | pass |
| B13 | 각 포지션 행을 조립한다 | 같음 | yes | pass |
| B14 | 행 단위 대사 차단을 표시한다 | 기존 | no | pass |
| B15 | 활성 격리가 있는 행만 badge와 해제 action을 받는다 | `TestTheConsoleOffersReleaseOnlyForAQuarantinedRow` | yes | pass |
| B16 | MANAGED 행에 정책 action을 붙인다 | 기존 | no | pass |
| B17 | MANAGED가 아니면 정책 action을 붙이지 않는다 | 기존 | no | pass |
| B18 | 등록된 공통 정책마다 override action을 만든다 | 기존 | no | pass |
| B19 | 외부 편입만 「자동관리 해제」를 받는다 | 기존 | no | pass |
| B20 | RELEASED 외부 편입만 「재편입」을 받는다 | 기존 | no | pass |

RED observed = a079 구현 전 코드에서 실패함을 2026-08-04에 확인. B15는 세대 비교
방식을 잘못 잡았을 때도 RED가 되며, 실제로 구현 중 `State.AdoptionGeneration`으로
비교하던 초안이 재편입된 포지션의 badge를 숨기는 것을 이 경로에서 잡아냈다.
