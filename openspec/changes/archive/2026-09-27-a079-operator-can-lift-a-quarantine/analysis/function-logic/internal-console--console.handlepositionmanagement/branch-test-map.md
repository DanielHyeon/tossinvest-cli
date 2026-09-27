# Branch Test Map: `Console.handlePositionManagement`

번호는 2026-09-27 HEAD f1e82e79 에서 재생성한 AST 기준이다(refresh — 아래 표 뒤 주석). a063이 추가한
분기는 B12·B13·B14(격리 capability 탐지, 조회 오류, 목록 적재)와 B20(행별 활성 격리)다.
나머지는 위치만 밀렸고 조건과 동작이 그대로다. B10·B11·B16·B17·B18 은 a111(`882a0b49`)이 넣은 분기다.

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
| B10 | 청산 원장을 읽을 수 있을 때만 저장된 청산 줄을 적재한다 (a111 `882a0b49`) | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` (a111 BTM B10) | a111 | pass |
| B11 | 읽은 청산 줄을 position id 로 색인한다 (a111) | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` (a111 BTM B11) | a111 | pass |
| B12 | 격리 capability가 있을 때만 격리를 읽는다 | `TestAnUnwiredCommanderShowsNoReleaseAction`, `TestTheConsoleOffersReleaseOnlyForAQuarantinedRow` | yes | pass |
| B13 | 격리 조회 실패는 화면을 막지 않고 이유를 표시한다 | `TestAFailedQuarantineReadDoesNotBlankTheWholeScreen` | yes | pass |
| B14 | 읽은 격리를 position id로 적재한다 | `TestTheConsoleOffersReleaseOnlyForAQuarantinedRow` | yes | pass |
| B15 | 각 포지션 행을 조립한다 | 같음 | yes | pass |
| B16 | 청산 원장을 못 읽으면 불명 사유를 `journal_unavailable` 로 둔다 (a111) | `TestA111PositionManagementNeverResurrectsAStoppedMarkerAfterClockRollback` (a111 BTM B16) | a111 | pass |
| B17 | 저장된 청산이 있고 MANAGED·청산 적격인 행만 신선도 판정한 줄을 받는다 (a111) | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` (a111 BTM B17) | a111 | pass |
| B18 | 정본 snapshot 이 있으면 값과 출처를 싣는다 (a111) | `TestA111PositionManagementSamplesResponseTimeAfterMarkerRead` (a111 BTM B18) | a111 | pass |
| B19 | 행 단위 대사 차단을 표시한다 | 기존 | no | pass |
| B20 | 활성 격리가 있는 행만 badge와 해제 action을 받는다 | `TestTheConsoleOffersReleaseOnlyForAQuarantinedRow` | yes | pass |
| B21 | MANAGED 행에 정책 action을 붙인다 | 기존 | no | pass |
| B22 | MANAGED가 아니면 정책 action을 붙이지 않는다 | 기존 | no | pass |
| B23 | 등록된 공통 정책마다 override action을 만든다 | 기존 | no | pass |
| B24 | 외부 편입만 「자동관리 해제」를 받는다 | 기존 | no | pass |
| B25 | RELEASED 외부 편입만 「재편입」을 받는다 | 기존 | no | pass |

RED observed = a079 구현 전 코드에서 실패함을 2026-08-04에 확인. B20(옛 B15)은 세대 비교
방식을 잘못 잡았을 때도 RED가 되며, 실제로 구현 중 `State.AdoptionGeneration`으로
비교하던 초안이 재편입된 포지션의 badge를 숨기는 것을 이 경로에서 잡아냈다.

Refresh(2026-09-27): 옛 번호는 448dfeb1 소스의 AST(B1..B20)였다. 옛/새 분기를 (kind, 소스 줄)로 difflib 정렬한
결과 B1..B9 → 같은 번호, B10..B13 → B12..B15, B14..B20 → B19..B25, 새 분기 B10·B11·B16·B17·B18 은 a111 이
넣었다. 새 분기 행의 시험은 a111 의 Branch Test Map(같은 소스 sha 287825be 의 AST 로 번호를 매긴 표)이
이름으로 인용한 기존 시험이다 — 새 시험 저술 0.
