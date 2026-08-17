# Branch Test Map: `reportStrategyProjectionDegraded`

편집 **전** 상태다(a109 base `016da624`). 이 함수의 분기는 하나이고, 그 분기가 아니라
**부작용의 성질**(등급·비차단·원장 미접촉)이 이 함수의 계약이다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | Notifier 가 없으면 stderr 한 줄만 남기고 돌아온다 | 직접 핀 없음 — 조립 실패 배포의 최소 표면. non-nil 경로는 `TestAFailedStrategyProjectionDoesNotStopTheEngine` 이 지난다 | no | no |

**분기 밖 계약 핀**(이 함수가 실제로 지켜야 하는 것):

| 계약 | Test | RED observed | GREEN observed |
|---|---|---|---|
| 강등이 사유를 stderr 로 말한다 | `TestAFailedStrategyProjectionDoesNotStopTheEngine` | yes (a108) | yes |
| 보고가 미전달 outbox 행을 남기지 않는다 | `TestTheDegradedBootWritesNoUndeliveredOutboxRow` | yes (a108) | yes |
| 반복 강등이 다음 부팅의 진입 게이트를 잠그지 않는다 | `TestASecondDegradedBootLeavesTheNextBootsEntryGateUnlatched` | yes (a108) | yes |
| 느린 publisher 가 부팅을 붙잡지 않고, 보고는 닿고, 종료가 그것을 끊지 않는다 | `TestTheDegradedBootDoesNotWaitForTheNotifier` | yes (a108 gstack Fix) | yes |

**a109 §2.2 가 더하는 시나리오**:

| 시나리오 | Test | RED observed | GREEN observed |
|---|---|---|---|
| 세 형제 endpoint 각각의 강등이 어느 표면인지 말한다(문구·scope 구분) | a109 §2.2 (`cmd/tossctl/a109_*_test.go`) | pending | pending |
| 세 형제 강등도 outbox 행 0 · 진입 게이트 미잠금 | a109 §2.2 | pending | pending |
