# a108 뮤테이션 원장 — T2 (tasks 3.4)

대상: `cmd/tossctl/engine.go` (겹2) · `cmd/tossctl/httpapi.go` (겹3).
baseline: GREEN 커밋 `1d9451e0`(겹2) + `cb2df8b3`(겹3). 뮤테이션은 **커밋된 baseline 위에서**
적용했고, 원복은 `git checkout --` 뒤 **심볼 개수 대조**로 확인했다 — 커밋 전 작업본에
`git checkout` 을 쓰면 GREEN 까지 지워지고 `git diff --quiet` 가 그것을 초록으로 통과시킨다.

실행: `python3 scratchpad/mutate.py` (드라이버는 임시 파일, 원장이 산출물이다).
각 뮤테이션마다 `go test ./cmd/tossctl/ -run '<a108 11건>' -count=1 -json` 을 돌려
**어느 테스트가 죽는지**를 이름으로 기록했다. 컴파일 실패는 뮤테이션이 아니므로
전부 컴파일되는 형태로만 만들었다(9/9 컴파일·실행됨).

## baseline (11건 전부 GREEN)

```text
TestAFailedStrategyProjectionDoesNotStopTheEngine
TestTheDegradedBootLeavesADurableCriticalAlert
TestEachDegradedRunEarnsItsOwnAlertAndOnlyOne
TestASucceedingProjectionIsStillServedAndClosed
TestTheDegradedBootStillHoldsTheJournalFlock
TestTheDegradedBootPublishesReadyOnlyAfterRecovery
TestTheGateIsRefusedBeforeTheProjectionEndpointIsTouched
TestADeadDescriptorDoesNotStopTheDaemon
TestAnAbsentDescriptorAndADeadOneBootTheSame
TestAnUninspectableDescriptorIsStillFatal
TestTheDaemonReadsTheDescriptorBesideItsOwnJournal
```

## 원장

| # | 대상 | 뮤테이션 | 죽은 테스트 | 생존 |
|---|---|---|---|---|
| M1 | engine.go | 강등을 `return projErr` 로 되돌린다 — **사고 당시 코드 그 자체** | `TestAFailedStrategyProjectionDoesNotStopTheEngine`, `TestTheDegradedBootLeavesADurableCriticalAlert`, `TestEachDegradedRunEarnsItsOwnAlertAndOnlyOne`, `TestTheDegradedBootStillHoldsTheJournalFlock`, `TestTheDegradedBootPublishesReadyOnlyAfterRecovery` | 6 |
| M2 | engine.go | 강등은 하되 durable 알림을 지운다 (stderr 한 줄만 = 「기동 1회 유실형」) | `TestTheDegradedBootLeavesADurableCriticalAlert`, `TestEachDegradedRunEarnsItsOwnAlertAndOnlyOne` | 9 |
| M3 | engine.go | 알림 event key 에서 실행 토큰을 빼 디렉터리로 고정 | `TestEachDegradedRunEarnsItsOwnAlertAndOnlyOne` | 10 |
| M4 | engine.go | 성공 경로의 `defer strategyRuntime.Close()` 제거 | `TestASucceedingProjectionIsStillServedAndClosed` | 10 |
| M5 | engine.go | projection 기동을 automation gate 평가보다 **앞으로** 이동 | `TestTheGateIsRefusedBeforeTheProjectionEndpointIsTouched`, `TestASucceedingProjectionIsStillServedAndClosed` | 9 |
| M6 | engine.go | 강등 시점에 a102 ready 신호를 발행 | `TestTheDegradedBootPublishesReadyOnlyAfterRecovery` | 10 |
| M7 | engine.go | 강등 직전에 journal flock 을 놓는다 | `TestTheDegradedBootStillHoldsTheJournalFlock` | 10 |
| M8 | httpapi.go | dial 실패 강등을 `return fmt.Errorf(...)` 로 되돌린다 — **crash loop 당시 코드** | `TestADeadDescriptorDoesNotStopTheDaemon`, `TestAnAbsentDescriptorAndADeadOneBootTheSame` | 9 |
| M9 | httpapi.go | 비-NotExist `os.Stat` 오류까지 강등 (강등의 상한 제거) | `TestAnUninspectableDescriptorIsStillFatal` | 10 |

**생존한 뮤테이션 0건.** 9건 전부 최소 하나의 테스트를 죽였다.

## 이 원장이 실제로 증명하는 것

- **M1·M8 이 task 3.4 가 요구한 항목이다.** 강등을 fatal 로 되돌리면 3.1 과 3.3 이 죽는다.
  M1 은 다섯 개를 죽이는데, 그중 셋(`...StillHoldsTheJournalFlock`,
  `...PublishesReadyOnlyAfterRecovery`, `EachDegradedRun...`)은 **강등된 부팅이 존재하지
  않으면 잴 것이 없기 때문**에 같이 죽는다 — 안전 핀은 강등 위에 서 있다.
- **M2 는 「알림이 있다」가 진짜 측정임을 보인다.** stderr 한 줄만 남기는 구현은
  `TestAFailedStrategyProjectionDoesNotStopTheEngine` 을 그대로 통과한다(그 테스트는
  stderr 문자열도 보므로). 원장 행을 보는 테스트가 없으면 「조용한 강등 = 은폐」가
  통과했을 것이다.
- **M3 은 event key 설계가 우연이 아님을 보인다.** key 를 디렉터리로 고정해도 다른 열 개
  테스트는 전부 통과한다. 두 번째 강등 기동이 조용해지는 것을 잡는 것은
  `TestEachDegradedRunEarnsItsOwnAlertAndOnlyOne` 하나뿐이다.
- **M5 는 평가 순서가 측정되고 있음을 보인다.** 게이트가 꺼진 계좌에서 socket 이 생기는
  것을 잡는 테스트가 실재한다.
- **M9 는 강등의 상한이 측정되고 있음을 보인다.** 「전부 강등」은 나머지 열 개를 통과한다.

## 원복 검증 (심볼 대조)

매 뮤테이션 뒤 `git checkout -- cmd/tossctl/engine.go cmd/tossctl/httpapi.go` 를 하고
아래 심볼 개수가 baseline 과 **정확히 일치**하는지 확인했다. 9회 전부 일치.

```text
engine.go   engineStrategyProjectionStart              3
            reportStrategyProjectionDegraded           3
            engineStrategyProjectionAlertType          4
            EnqueueAlert                               2
            "if strategyRuntime != nil {"              1
            "run = at.UTC().Format(time.RFC3339Nano)"  1
httpapi.go  "inspect strategy runtime projection"      1
            "strategy runtime projection에 연결하지 못했다"  1
```

`git status --porcelain` 도 매회 비어 있음을 확인했다(드라이버가 아니면 중단).

## 이 원장이 덮지 못한 것 (숨기지 않고 적는다)

- **겹1(`internal/strategyprojectionrpc`)은 T1 소유라 뮤테이션하지 않았다.** T2 테스트는
  실패를 seam 으로 주입하므로, T1 이 회수 규칙을 어떻게 바꾸든 이 원장의 결론은 바뀌지
  않는다 — 반대로 말하면 **회수의 전체성 자체는 이 원장이 증명하지 않는다**(tasks 2.4 몫).
- **강등 상태의 httpapi 가 실제로 `strategyRuntime = nil` 인지는 간접 측정이다.**
  `Dial` 실패 시 `client` 는 nil 이고 대입되지 않으므로 코드에 제3의 경로가 없다는 것이
  근거이며, 라우터 내부를 들여다보는 테스트는 만들지 않았다(경계 인증 fixture 가
  이 change 의 측정 대상보다 커진다). M8 이 「fatal 로 되돌리면 죽는다」를 잡는다.
- **비-unix 빌드는 측정하지 않았다.** 두 테스트 파일 모두 `//go:build unix` 다.
  참고로 그 빌드에서는 `Start` 가 **항상** 실패하므로, 이 change 이전에는
  `tossctl engine run` 이 애초에 뜰 수 없었다(`transport_other.go`).
