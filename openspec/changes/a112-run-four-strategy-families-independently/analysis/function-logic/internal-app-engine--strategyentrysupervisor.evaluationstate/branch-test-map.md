# Branch Test Map: `StrategyEntrySupervisor.evaluationState`

- Source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`; AST branch locations are authoritative.
- Revision: base — 편집하지 않는다. 태스크 5.6 이 인용한다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 893:2 — 종료 중·nil·**잠김**·사이클 없음·권한 없는 dormant | `TestALatchedMarketSkipsTheTriggersAlreadySittingInItsQueue` | no (base) | yes (block 894-896 count=1, **5.6 이 처음 잠긴 갈래로 실행**) |
| B2 | if at 897:2 — 권한 갱신 worker 는 만료를 보지 않는다 | `TestTheOnlyWorkerProductionActuallyRunsSwallowsEveryCycleError`, `TestARefreshOnlyWorkerSwallowsACentralIntegrityErrorToo` | no (base) | yes (block 897-899) |
| 본문 | 900:2 — effective worker 의 신선도 판정 | `TestExpiredAuthorityLatchesBeforeEvaluation` 외 다수 | no (base) | yes |

## 측정으로 확인한 빈칸

없다. 세 블록 모두 `count>0` 이다.

다만 B1 은 **다섯 조건의 논리합**이고 커버리지는 그중 어느 조건이 참이었는지
말해 주지 않는다. 5.6 이 새로 채운 것은 그중 `worker.latched` 하나이며,
그것이 `runMarket` 의 `800-801`(잠긴 시장이 큐에 남은 요청을 버린다)을 여는
조건이다. 나머지 넷(종료 중·nil·사이클 없음·권한 없는 dormant)은 기존 스위트가
같은 블록을 통해 밟는다.
