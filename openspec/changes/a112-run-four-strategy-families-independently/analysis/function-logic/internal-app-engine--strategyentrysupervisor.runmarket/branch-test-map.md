# Branch Test Map: `StrategyEntrySupervisor.runMarket`

- Source SHA-256: `66150078e25dfad6d1fec322b955e5f23e3aad77f0525321867a500e0960f58f`; AST branch locations are authoritative.
- Revision: base — 이 change 는 이 함수를 편집하지 않는다. RED 칸이 모두 `no (base)`
  인 이유가 그것이다. 이 번들은 태스크 5.3.2 가 인용할 분기를 열거하기 위해 만들었다.

커버리지는 주장이 아니라 **측정**이다. "Test" 칸은 시험마다 따로
`go test ./internal/app/engine/ -run '^<Test>$' -coverprofile=…` 를 돌려 그
프로파일이 해당 줄을 포함하는 블록을 `count>0` 으로 보고한 것만 적었고,
"**없음**" 은 패키지 전체를 한 번에 돌린 프로파일에서도 `count=0` 인 것이다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | select at 786:2 — 배리어 전에 취소되면 사이클 0 회 | 배리어 경합 시험(`TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`) | no (base) | yes |
| B2 | for at 791:2 — 시장 하나를 도는 **단일** 소비자 루프 | 같은 시험 | no (base) | yes |
| B3 | select at 792:3 — 취소 vs 큐 도착 | `TestShutdownAndTriggerShareBarrierAndDrainBothQueues` | no (base) | yes |
| B4 | if at 801:4 — 권한 만료가 평가 전에 잠근다 | `TestExpiredAuthorityLatchesBeforeEvaluation` | no (base) | yes (block 785-786) |
| B5 | if at 803:5 — 만료 잠금 자체가 실패한다 | `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`/"권한 만료의 잠금도…" | no (base) | yes (block 787-790 count=1 — **5.6 이 처음 실행**) |
| B6 | if at 807:5 — 만료 뒤 재시작 대기가 실패한다 | 취소 갈래 `TestExpiredAuthorityLatchesBeforeEvaluation`; 비취소 갈래 `TestTheFourEscalations…`/"만료 뒤의 재시작 대기도…" | no (base) | yes (block 792-794 count=1, block 795-796 count=1 — **후자를 5.6 이 처음 실행**) |
| B7 | if at 808:6 — 그 실패가 ctx 취소 때문이다 | `TestExpiredAuthorityLatchesBeforeEvaluation` | no (base) | yes (block 792-794 count=1) |
| B8 | if at 816:4 — 꺼졌거나 잠긴 worker 가 사이클을 건너뛴다 | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue` | no (base) | yes (block 800-801 count=1 — **5.6 이 처음 실행**) |
| B9 | if at 820:4 — 마감 시한을 넘긴 사이클을 버려진 것으로 표시 | `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction` 외 3 | no (base) | yes |
| B10 | if at 823:4 — 취소된 사이클이 루프를 끝낸다 | `TestShutdownAndTriggerShareBarrierAndDrainBothQueues` 외 2 | no (base) | yes |
| B11 | if at 826:4 — 성공한 사이클이 다음 투입을 기다린다 | `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive` 외 2 | no (base) | yes |
| B12 | if at 829:4 — 권한 갱신 전용 worker 의 오류는 잠그지 않는다 | `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo` | no (base) | yes (block 813-814 count=1 — **5.6 이 처음 실행. 이것이 오늘 생산이 실제로 도는 구성이다**) |
| B13 | if at 832:4 — 중앙 무결성 오류가 모든 신규 진입을 멈춘다 | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety` | no (base) | yes |
| B14 | if at 837:4 — 보통 오류의 잠금 자체가 실패한다 | `TestTheFourEscalations…`/"관측 시각이 없으면…"·"latch revision 이 소진되면…", `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt` | no (base) | yes (block 821-824 count=1 — **5.6 이 처음 실행**) |
| B15 | if at 841:4 — 잠근 뒤 재시작 대기가 실패한다 | 취소 갈래 `TestMarketPanicIsContainedAndCannotRecoverMemoryAuthority` 외 3; 비취소 갈래 `TestTheFourEscalations…`/"재시작 기한이 계약 밖이면…" | no (base) | yes (block 826-828 count=1, block 829-830 count=1 — **후자를 5.6 이 처음 실행**) |
| B16 | if at 842:5 — 그 실패가 ctx 취소 때문이다 | `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction` 외 3 | no (base) | yes (block 826-828 count=1) |

## 측정으로 확인한 빈칸 — **닫혔다 (2026-09-03, 태스크 5.6)**

원래 이 절은 이렇게 적혀 있었다: 패키지 전체 스위트에서도 `count=0` 인 블록이
여섯이고(`787-790`, `795-796`, `800-801`, `813-814`, `821-824`, `829-830`),
공통점은 **잠금·재시작 자체가 실패하는 경로와 사이클을 아예 돌리지 않는 경로**
라는 것. 태스크 5.7 이 가져갈 자리로 적었지만, 5.7 의 리허설은 새 타입을 재고
엔진은 재지 않았다. 실제로 가져간 것은 **5.6** 이다.

여섯 모두 이제 `count=1` 이고, 그 여섯이 하나의 문장을 이룬다: **전략 고장이
엔진을 세우는 경로는 넷뿐이고 넷 다 감독자 자신의 장부가 깨진 경우다.**
평가 실패(보통 오류·panic·마감 시한)는 그 넷에 없다. 나머지 둘은 사이클을
돌리지 않는 두 갈래이고, 그중 `813-814` 는 **오늘 생산이 실제로 도는 구성**이다.

시험은 `internal/app/engine/a112_fault_scope_test.go` 에 있고, 반증 10/10 이
잡혔다(상세는 review.md 의 5.6 절).
