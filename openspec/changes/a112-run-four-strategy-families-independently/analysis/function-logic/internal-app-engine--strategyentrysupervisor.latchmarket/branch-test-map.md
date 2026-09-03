# Branch Test Map: `StrategyEntrySupervisor.latchMarket`

- Source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`; AST branch locations are authoritative.
- Revision: base — 이 change 는 이 함수를 편집하지 않는다. RED 칸이 모두
  `no (base)` 인 이유가 그것이다. 태스크 5.6 이 인용할 분기를 열거하려고 만들었다.

커버리지는 주장이 아니라 **측정**이다. "Test" 칸은 시험마다 따로
`go test ./internal/app/engine/ -run '^<Test>$' -coverprofile=…` 를 돌려 그
프로파일이 해당 줄을 포함하는 블록을 `count>0` 으로 보고한 것만 적었고,
"**없음**" 은 패키지 전체를 한 번에 돌린 프로파일에서도 `count=0` 인 것이다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | if at 982:2 — 실패 문구가 비어 있다 | **없음** | no (base) | **no — block 953-955 count=0** |
| B2 | if at 986:2 — 비정상 사이클의 refusal | `TestPairedMarketAbnormalReturnSchedulesOnlyLocalBoundedRestartAndKeepsEverySafetyLoopAlive` 외 1 | no (base) | yes (block 957-959) |
| B3 | if at 989:2 — 권한 만료의 refusal | `TestExpiredAuthorityLatchesBeforeEvaluation` 외 1 | no (base) | yes (block 960-962) |
| B4 | if at 993:2 — 관측 시각이 없어 잠금을 거부한다 | `TestTheFourEscalationsThatStopTheEngineAreExactlyTheSupervisorsOwnBrokenBookkeeping`/"관측 시각이 없으면…" | no (base) | yes (block 964-966 count=1, **5.6 이 처음 실행**) |
| B5 | if at 997:2 — latch revision 소진 | 같은 시험의 "latch revision 이 소진되면…"·"권한 만료의 잠금도…" 및 `TestBrokenSupervisorBookkeepingTakesTheSafetyLoopsDownWithIt` | no (base) | yes (block 968-971 count=1, **5.6 이 처음 실행**) |
| B6 | if at 1003:2 — 첫 refusal 만 기록한다 | `TestMarketRestartAttemptAndDeadlineSaturateWithoutOverwritingFirstTypedRefusal` 외 4 | no (base) | yes (block 974-976) |
| B7 | if at 1006:2 — 첫 실패 문구·latchID·revision++ | 같은 시험 외 4 | no (base) | yes (block 977-982) |
| B8 | if at 1012:2 — 재시작 시도 수 증가(포화 전) | 같은 시험 외 4 | no (base) | yes (block 983-985) |
| B9 | select at 1026:2 — fault 를 스트림에 건넸다 (`case` 팔) | `TestEveryWorkerCanHandOffItsFaultWithoutAnybodyDraining` 외 5 | no (base) | yes (block 998-999) |
| B9-default | 같은 select at 997:2 의 `default` — 스트림 포화 | **없음** | no (base) | **no — block 1000-1001 count=0, 구조적으로 도달 불가(아래)** |

## 측정으로 확인한 빈칸

둘이다. 둘 다 "시험을 안 썼다"가 아니라 **"오늘 코드에서 열리지 않는다"** 이고,
그 이유가 서로 다르다.

- `917-919` (B1 true) — 잠금 이유가 빈 문자열이 되려면 평가가
  `errors.New("")` 같은 것을 돌려줘야 한다. 오늘 그 값을 만드는 생산 경로가 없다.
  자리표시자는 방어이지 관측된 동작이 아니다.
- `964-965` (B9 의 `default`) — fault 스트림이 포화해야 열린다. 용량은 2 이고(`:584`),
  잠긴 시장은 `evaluationState`(`893:2`)가 거부하므로 시장당 잠금은 한 번이며,
  시장은 정확히 둘이다. **2 = 2 라서 넘칠 수 없다.** 이 균형은 어디에도 적혀
  있지 않았고, 5.1.2 가 시장 둘을 lane 여덟으로 바꾸면 깨진다. 태스크 5.6 은
  그래서 그 등식을 `TestTheFaultStreamHoldsOneSlotForEveryWorkerThatCanLatch`
  로 못 박았다(용량을 1 로 만드는 변이 M7 이 잡힌다).
