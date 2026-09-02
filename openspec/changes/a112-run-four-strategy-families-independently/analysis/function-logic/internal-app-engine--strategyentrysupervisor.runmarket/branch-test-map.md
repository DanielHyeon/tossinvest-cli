# Branch Test Map: `StrategyEntrySupervisor.runMarket`

- Source SHA-256: `4e457c677157b2f8c73f813f8250575657b6beedddc1ad467db209a35579986d`; AST branch locations are authoritative.
- Revision: base — 이 change 는 이 함수를 편집하지 않는다. RED 칸이 모두 `no (base)`
  인 이유가 그것이다. 이 번들은 태스크 5.3.2 가 인용할 분기를 열거하기 위해 만들었다.

커버리지는 주장이 아니라 **측정**이다. "Test" 칸은 시험마다 따로
`go test ./internal/app/engine/ -run '^<Test>$' -coverprofile=…` 를 돌려 그
프로파일이 해당 줄을 포함하는 블록을 `count>0` 으로 보고한 것만 적었고,
"**없음**" 은 패키지 전체를 한 번에 돌린 프로파일에서도 `count=0` 인 것이다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | select at 770:2 — 배리어 전에 취소되면 사이클 0 회 | 배리어 경합 시험(`TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`) | no (base) | yes |
| B2 | for at 775:2 — 시장 하나를 도는 **단일** 소비자 루프 | 같은 시험 | no (base) | yes |
| B3 | select at 776:3 — 취소 vs 큐 도착 | `TestShutdownAndTriggerShareBarrierAndDrainBothQueues` | no (base) | yes |
| B4 | if at 785:4 — 권한 만료가 평가 전에 잠근다 | `TestExpiredAuthorityLatchesBeforeEvaluation` | no (base) | yes (block 785-786) |
| B5 | if at 787:5 — 만료 잠금 자체가 실패한다 | **없음** | no (base) | **no — block 787-790 count=0** |
| B6 | if at 791:5 — 만료 뒤 재시작 대기가 실패한다 | `TestExpiredAuthorityLatchesBeforeEvaluation` | no (base) | 부분 — 취소 갈래만 yes, `:795`–`:796` 은 count=0 |
| B7 | if at 792:6 — 그 실패가 ctx 취소 때문이다 | `TestExpiredAuthorityLatchesBeforeEvaluation` | no (base) | yes (block 792-794 count=1) |
| B8 | if at 800:4 — 꺼졌거나 잠긴 worker 가 사이클을 건너뛴다 | **없음** | no (base) | **no — block 800-801 count=0** |
| B9 | if at 804:4 — 마감 시한을 넘긴 사이클을 버려진 것으로 표시 | `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction` 외 3 | no (base) | yes |
| B10 | if at 807:4 — 취소된 사이클이 루프를 끝낸다 | `TestShutdownAndTriggerShareBarrierAndDrainBothQueues` 외 2 | no (base) | yes |
| B11 | if at 810:4 — 성공한 사이클이 다음 투입을 기다린다 | `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive` 외 2 | no (base) | yes |
| B12 | if at 813:4 — 권한 갱신 전용 worker 의 오류는 잠그지 않는다 | **없음** | no (base) | **no — block 813-814 count=0** |
| B13 | if at 816:4 — 중앙 무결성 오류가 모든 신규 진입을 멈춘다 | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety` | no (base) | yes |
| B14 | if at 821:4 — 보통 오류의 잠금 자체가 실패한다 | **없음** | no (base) | **no — block 821-824 count=0** |
| B15 | if at 825:4 — 잠근 뒤 재시작 대기가 실패한다 | `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority` 외 3 | no (base) | 부분 — 취소 갈래만 yes, `:829`–`:830` 은 count=0 |
| B16 | if at 826:5 — 그 실패가 ctx 취소 때문이다 | `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction` 외 3 | no (base) | yes (block 826-828 count=1) |

## 측정으로 확인한 빈칸

패키지 전체 스위트에서도 `count=0` 인 블록은 여섯이다: `787-790`, `795-796`,
`800-801`, `813-814`, `821-824`, `829-830`. 공통점이 있다 — **잠금·재시작 자체가
실패하는 경로와, 사이클을 아예 돌리지 않는 경로**다. 태스크 5.7 이 가져갈 자리이고,
이 번들은 그 사실을 적을 뿐 메우지 않는다.
