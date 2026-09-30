# Branch Test Map: `StrategyEntrySupervisor.runMarket`

- Source SHA-256: `9e24e93028b2728071d71d1d6ccea2c2a83fe768f6efe2dc09a57906c435a373`; AST branch locations are authoritative.
- Revision: **modified (태스크 5.6.2.1, 2026-09-30)** — 그 앞은 8.8.4(2026-09-05, B12 에 세기 호출). 이번 편집은 **B12 본문 안에
  새 분기 하나(B13)**: 중앙 무결성 오류를 삼키지 않고 신규 진입을 닫는다(사람 결정 (6) — fail-closed 의 수단은 EntryGate).
  판정 순서(refreshOnly 가 중앙 판정보다 앞)는 그대로다. 편집 전 번들은 `analysis/measurements/lot-5.6.2-5.2.2/pre-edit/` 에 보존.
- **재번호**: 편집 전 B1~B12 불변, 새 B13, 편집 전 B13~B16 → B14~B17(위치 정렬 — 편집 지점 뒤 네 분기가 하나씩 밀림).
- 측정: `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json` — 격리 사본(HEAD `3260f4eb` archive + 이 로트 파일), `./internal/app/engine` 시험 109개를 하나씩, 실패 0.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | select at 881:2 — 배리어 전에 취소되면 사이클 0 회 | 배리어 경합 시험(`TestStrategyEntrySupervisorStartsKRAndUSCyclesConcurrently`) | no (base) | yes (하네스가 select 블록 좌표를 못 잡음 — 이 행은 앞 측정 유지) |
| B2 | for at 886:2 — 시장 하나를 도는 **단일** 소비자 루프 | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue`, `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine` 외 24 | no (base) | yes (block 884.6-885.10, 시험 26개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B3 | select at 887:3 — 취소 vs 큐 도착 | `TestShutdownAndTriggerShareBarrierAndDrainBothQueues` | no (base) | yes (하네스가 select 블록 좌표를 못 잡음 — 이 행은 앞 측정 유지) |
| B4 | if at 896:4 — 권한 만료가 평가 전에 잠근다 | `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping` | no (base) | yes (block 894.15-896.19, 시험 2개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B5 | if at 898:5 — 만료 잠금 자체가 실패한다 | `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping` | no (base) | yes (block 896.19-899.6, 시험 1개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B6 | if at 902:5 — 만료 뒤 재시작 대기가 실패한다 | `TestExpiredAuthorityLatchesBeforeEvaluation`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping` | no (base) | yes (block 900.70-901.26, 시험 2개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B7 | if at 903:6 — 그 실패가 ctx 취소 때문이다 | `TestExpiredAuthorityLatchesBeforeEvaluation` | no (base) | yes (block 901.26-903.7, 시험 1개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B8 | if at 911:4 — 꺼졌거나 잠긴 worker 가 사이클을 건너뛴다 | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue` | no (base) | yes (block 909.16-910.13, 시험 1개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B9 | if at 915:4 — 마감 시한을 넘긴 사이클을 버려진 것으로 표시 | `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`, `TestMarketQueueSaturationDoesNotConsumePeerQueue` 외 4 | no (base) | yes (block 913.17-915.5, 시험 6개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B10 | if at 918:4 — 취소된 사이클이 루프를 끝낸다 | `TestMarketQueueSaturationDoesNotConsumePeerQueue`, `TestShutdownAndTriggerShareBarrierAndDrainBothQueues` 외 3 | no (base) | yes (block 916.17-918.5, 시험 5개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B11 | if at 921:4 — 성공한 사이클이 다음 투입을 기다린다 | `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine`, `TestARefreshOnlyWorkerCentralIntegrityErrorLeavesTheEngineRunning` 외 11 | no (base) | yes (block 919.18-920.13, 시험 13개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B12 | if at 924:4 — 권한 갱신 전용 worker 의 오류는 잠그지 않되 **세어진다** — 5.6.2.1 부터 그 안에서 중앙 무결성 오류는 신규 진입을 닫는다(B13) | `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine`, `TestARefreshOnlyWorkerCentralIntegrityErrorLeavesTheEngineRunning` 외 4 | yes (5.6.2.1 편집 — 본문에 B13 추가) · `red-5.6.2.1.log` | yes (block 922.19-940.83, 시험 6개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B13 | if at 942:5 — (새 분기, 5.6.2.1) 권한 갱신 전용 worker 의 **중앙 무결성** 오류 — 게이트가 있으면 `blockEntryOnCentralIntegrity` 가 신규 진입을 닫고(`ReasonStrategyCentralIntegrity`) `continue`, 게이트가 없을 때만 이 본문(프로세스 전체 fail-closed 로 중계) | `TestWithoutAnEntryGateACentralFaultIsNotSwallowed` | yes — `red-5.6.2.1.log`(`TestWithoutAnEntryGateACentralFaultIsNotSwallowed` · `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine` FAIL); 변이 E02 · E03 · E04 · E05 CAUGHT | yes (block 940.83-943.6, 시험 1개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`); 게이트 갈래(본문 밖)는 `TestARefreshOnlyCentralIntegrityFaultBlocksNewEntryNotTheEngine` 가 게이트 사유 · CheckEntry · Run 불반환 · 두 시장 계속으로 잼 |
| B14 | if at 948:4 — (effective worker) 중앙 무결성 오류가 프로세스 전체 fail-closed 로 올라간다 — 생산 0(effective worker 없음); 활성화 로트 전 처분은 이월 표(5.6.2.1) | `TestCentralIntegrityFailureEscapesOuterLoopAndDrainsSafety` | no (base) | yes (block 946.39-949.5, 시험 1개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B15 | if at 953:4 — 보통 오류의 잠금 자체가 실패한다 | `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt`, `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping` | no (base) | yes (block 951.18-954.5, 시험 2개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B16 | if at 957:4 — 잠근 뒤 재시작 대기가 실패한다 | `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction` 외 6 | no (base) | yes (block 955.69-956.25, 시험 8개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |
| B17 | if at 958:5 — 그 실패가 ctx 취소 때문이다 | `TestAnEffectiveMarketFaultLeavesItsPeerAndTheSupervisorAlone`, `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction` 외 5 | no (base) | yes (block 956.25-958.6, 시험 7개 — `analysis/measurements/lot-5.6.2-5.2.2/coverage-post-5.6.2.1-engine.json`) |

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

> **커버리지 블록 번호는 옮겨 적지 않고 프로파일로 다시 잰다.** 5.6.1 이 적어 둔
> 번호는 5.1.2.1(+16)·5.2.1(+3) 의 삽입 뒤 19줄 밀려 있었고 아무도 옮기지 않았다.
> 5.2.1 이 다시 재어 맞췄고(29개 중 28개가 정확히 +19, `count` 도 전부 일치),
> 남은 하나는 5.6.1 이 적을 때 실제 블록보다 한 줄 짧게 적혀 있어 잰 값으로
> 바꿨다 — 산술로 옮겼다면 그 오류를 그대로 옮겨 적었을 것이다. 5.3.3 이 다시
> 옮기면서(+17) 모든 번호를 프로파일과 대조했다. 프로파일은
> `go test -count=1 -tags tossos_testseams -coverprofile ./internal/app/engine/`
> (2026-09-03, 77.9% of statements).


