# Branch Test Map: `invokeBoundedStrategyCycle`

- Source SHA-256: `627c647d087032586c4b63ca315a30fd9fad6b51af329fa4e8bf4fecd7104e08`; AST branch locations are authoritative.
- Revision: base — 이 change 는 이 함수를 편집하지 않는다. RED 칸이 모두 `no (base)`
  인 이유가 그것이다. 이 번들은 태스크 5.3.2 가 인용할 분기를 열거하기 위해 만들었다.

커버리지는 주장이 아니라 **측정**이다. "Test" 칸은 시험마다 따로
`go test ./internal/app/engine/ -run '^<Test>$' -coverprofile=…` 를 돌려 그
프로파일이 해당 줄을 포함하는 블록을 `count>0` 으로 보고한 것만 적었고,
"**없음**" 은 패키지 전체를 한 번에 돌린 프로파일에서도 `count=0` 인 것이다.

| Branch | Scenario anchor | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | select at 924:2 — 상위 취소 / 사이클 완료 / 마감 시한 셋 중 먼저 오는 것 | `TestShutdownAndTriggerShareBarrierAndDrainBothQueues`(889-890), `TestMarketFailureEmitsExactIrreversibleFaultAndKeepsPeerSafetyAlive`(891-892), `TestContextIgnoringCycleWatchdogLatchesOnceAndLateResultHasNoAction`(897) | no (base) | yes — 세 갈래 모두 count>0 |
| B2 | if at 930:3 — 마감 시한이 울렸는데 상위가 **이미** 취소되어 있다 | `TestTheWatchdogRechecksCancellationInsteadOfTrustingItsOwnTimer` | no (base) | yes (block 930-932 count=1 — **5.6 이 처음 실행**) |

## 측정으로 확인한 빈칸

**닫혔다 (2026-09-03, 태스크 5.6).** 원래는 "B2 는 어느 시험도 돌리지 않는다.
태스크 5.7 이 가져갈 자리"라고 적혀 있었다. 5.7 은 가져가지 않았고 5.6 이 가져갔다.

왜 아무도 밟지 못했는지가 이제 이름을 갖는다: **밟지 않은 것이 아니라 관측할 수
없었다.** 실제 취소에서는 `ctx.Done()` 갈래와 마감 시한 갈래가 둘 다 준비되고,
Go 는 준비된 갈래 중 하나를 무작위로 고르며, 두 갈래의 반환값이 **같아서**
결과로도 구별되지 않는다.

`Done()` 이 절대 준비되지 않으면서 `Err()` 는 취소를 보고하는 context 를 넣으면
마감 시한 갈래만 열려 B2 가 결정적으로 돈다. 그때 이 갈래가 하는 일이 드러난다:
자기 타이머를 믿지 않고 취소를 다시 확인해 **취소를 `abnormal` 로 승격시키지
않는다.** 생산 임계값이 1 이므로, 그 승격은 곧 "정상 종료가 시장을 잠근다"이다.
`internal/app/engine/a112_watchdog_cancellation_test.go` 가 그 두 칸을 대조군과
함께 재고, 재확인을 지우는 변이(M6)가 잡힌다.
