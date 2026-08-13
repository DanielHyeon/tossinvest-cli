# a108 뮤테이션 원장 — T2 (tasks 3.4 · Fix 라운드 6.9)

대상: `cmd/tossctl/engine.go`(겹2) · `cmd/tossctl/httpapi.go`(겹3) ·
`cmd/tossctl/httpapi_reader.go`(겹3 Fix).

이 문서는 **두 라운드**를 담는다.

- **§원 라운드** — GREEN 커밋 `1d9451e0`(겹2)·`cb2df8b3`(겹3) 위에서 잰 9건.
  A2 적대 리뷰가 그중 **M2·M3·M9 가 지키던 계약 자체를 뒤집었다.** 뒤집힌 계약을
  지키던 뮤테이션은 「살아남았다/죽었다」와 무관하게 **폐기**다 — 아래에 사유와 함께
  남긴다(지우지 않는다: 무엇이 왜 틀렸는지가 이 문서의 값이다).
- **§Fix 라운드** — GREEN 커밋 `aecc03e0` 위에서 다시 잰 12건. 이것이 **현재 유효한
  원장**이고, branch-test-map 이 인용하는 번호도 이쪽이다.

## §Fix 라운드 (현재 유효)

baseline: `aecc03e0`. 뮤테이션은 **커밋된 baseline 위에서** 적용했다.
실행: `python3 <scratchpad>/mutate.py` + `mutate_m5.py`(드라이버는 임시 파일, 원장이
산출물이다). 각 뮤테이션마다 `go test ./cmd/tossctl/ -run '<a108 14건>' -count=1 -json`
을 돌려 **어느 테스트가 죽는지**를 이름으로 기록했다. 컴파일 실패는 뮤테이션이
아니므로 전부 컴파일·실행되는 형태로만 만들었다(12/12 실행됨).

### baseline (14건 중 13 GREEN + 1 의도된 RED)

```text
TestAFailedStrategyProjectionDoesNotStopTheEngine                PASS
TestTheDegradedBootWritesNoUndeliveredOutboxRow                  PASS
TestASecondDegradedBootLeavesTheNextBootsEntryGateUnlatched      PASS
TestASucceedingProjectionIsStillServedAndClosed                  PASS
TestTheDegradedBootStillHoldsTheJournalFlock                     PASS
TestTheDegradedBootLeavesReadyToTheRuntimeSeam                   PASS
TestTheGateIsRefusedBeforeTheProjectionEndpointIsTouched         PASS (+2 subtests)
TestADeadDescriptorDoesNotStopTheDaemon                          PASS
TestAnAbsentDescriptorAndADeadOneBootTheSame                     PASS
TestAnUninspectableDescriptorDegradesLikeTheConsole              PASS
TestTheDaemonReadsTheDescriptorBesideItsOwnJournal               PASS
TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding    PASS
TestAnAbsentStrategyReaderStaysDormantRatherThanUnavailable      PASS
TestASocketFileWithNoOwnerDegradesTheDaemon                      FAIL ← 의도된 RED
```

**의도된 RED 하나를 숨기지 않는다.** `TestASocketFileWithNoOwnerDegradesTheDaemon`
(S3 — descriptor·socket 파일 잔존 + 주인 사망)은 T1-fix 의 `Dial` connect probe(6.4)를
기다린다. 이 라운드 시점의 `Dial` 은 socket 의 Lstat·모드·perm 만 보므로 S3 를
**성공**으로 읽는다. 이 뮤테이션 라운드에서 그 테스트는 baseline 이 이미 FAIL 이므로
어떤 뮤테이션의 「죽인 테스트」로도 세지 않았다.

**정산(gstack Fix 라운드, 2026-08-14).** T1-fix 가 병합되면서 그 RED 는 **테스트를
손대지 않고** GREEN 이 됐다 — 교차 RED 가 병합의 검증으로 작동한 기록이다. 지금
14건은 전부 GREEN 이고, 그 테스트의 실패는 「아직 안 만들어진 것」이 아니라
**probe 가 사라졌다**는 회귀 신호다. 테스트의 주석과 실패 문구도 그렇게 고쳤다
(gstack A5①). 아래 §gstack 라운드가 그 baseline 위에서 다시 잰 원장이다.

### 원장

| # | 대상 | 뮤테이션 | 죽은 테스트 | 생존 |
|---|---|---|---|---|
| M1 | engine.go | 강등을 `return projErr` 로 되돌린다 — **사고 당시 코드 그 자체** | `TestAFailedStrategyProjectionDoesNotStopTheEngine`, `TestTheDegradedBootWritesNoUndeliveredOutboxRow`, `TestASecondDegradedBootLeavesTheNextBootsEntryGateUnlatched`, `TestTheDegradedBootStillHoldsTheJournalFlock`, `TestTheDegradedBootLeavesReadyToTheRuntimeSeam` | 10 |
| **M2b** | engine.go | **D3-2 를 되돌린다** — 강등이 durable critical outbox 행을 다시 쓴다(원 D3 의 코드) | `TestTheDegradedBootWritesNoUndeliveredOutboxRow`, `TestASecondDegradedBootLeavesTheNextBootsEntryGateUnlatched` | 13 |
| **M3b** | engine.go | 이벤트 타입을 obs 등급표에 **등재된** critical 이름(`obs.EventOperatingMode`)으로 바꾼다 — outbox rail 재진입 | `TestTheDegradedBootWritesNoUndeliveredOutboxRow`, `TestASecondDegradedBootLeavesTheNextBootsEntryGateUnlatched` | 13 |
| M4 | engine.go | 성공 경로의 `defer strategyRuntime.Close()` 제거 | `TestASucceedingProjectionIsStillServedAndClosed` | 14 |
| M5 | engine.go | projection 기동을 automation gate 평가보다 **앞으로** 옮긴다 | `TestTheGateIsRefusedBeforeTheProjectionEndpointIsTouched`(+ `게이트 OFF` 서브테스트) | 13 |
| M6 | engine.go | 강등 시점에 a102 ready 신호를 발행 | `TestTheDegradedBootLeavesReadyToTheRuntimeSeam` | 14 |
| M7 | engine.go | 강등 직전에 journal flock 을 놓는다 | `TestTheDegradedBootStillHoldsTheJournalFlock` | 14 |
| M8 | httpapi.go | dial 실패 강등을 `return fmt.Errorf(...)` 로 되돌린다 — **crash loop 당시 코드** | `TestADeadDescriptorDoesNotStopTheDaemon`, `TestAnAbsentDescriptorAndADeadOneBootTheSame` | 13 |
| **M9b** | httpapi.go | **D4-2 를 되돌린다** — 비-NotExist `os.Stat` 오류를 다시 fatal 로(원 D4 의 코드) | `TestAnUninspectableDescriptorDegradesLikeTheConsole` | 14 |
| **M10** | httpapi_reader.go | 집계의 전략 흡수를 제거 — 읽기 실패가 다시 스냅샷 전체를 죽인다 | `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` | 14 |
| **M11** | httpapi_reader.go | 읽기 실패를 reader 부재와 **같은 값**으로 접는다(`Unavailable`→`Dormant`) | `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` | 14 |
| **M12** | httpapi_reader.go | 성공한 읽기까지 `Unavailable` 로 덮는다 — 「항상 강등」 | `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` | 14 |

**생존한 뮤테이션 0건.** 12건 전부 최소 하나의 테스트를 죽였다.

### 이 원장이 실제로 증명하는 것

- **M2b·M3b 가 D3-2 의 핵심이다.** 둘은 서로 다른 길로 같은 사고에 도달한다 —
  하나는 `EnqueueAlert` 를 직접 부르고, 하나는 이벤트 이름을 등급표에 올려
  `Notify` 가 알아서 outbox 로 흘리게 한다. **두 길 다 같은 두 테스트가 잡는다.**
  이것이 중요한 이유: 계약을 「EnqueueAlert 를 부르지 마라」가 아니라 「미전달 행이
  생기지 마라」로 잡았기 때문에, 아직 존재하지 않는 세 번째 길도 잡힌다.
- **M11 은 「접지 마라」가 진짜 측정임을 보인다.** 두 값을 하나로 접어도 집계는 살고
  화면은 뜬다 — 열세 개 테스트가 통과한다. 그것을 잡는 것은 refusal code 를 시장별로
  보는 한 테스트뿐이고, 그 구분이 없으면 운영자는 「전략 기능을 안 켰다」와
  「엔진이 죽었다」를 화면에서 구별할 수 없다.
- **M12 는 대조군이 실재함을 보인다.** 「항상 강등」은 M10·M11 이 겨누는 성질을
  전부 통과한다. 살아 있는 projection 이 `EVIDENCE_STALE` 로 그대로 실리는지 보는
  절반이 없으면 그 구현이 살아남는다.
- **M1 이 다섯을 죽이는 이유는 여전히 같다.** 강등된 부팅이 존재하지 않으면
  flock·ready·outbox 를 잴 대상 자체가 없다 — 안전 핀은 강등 위에 서 있다.
- **M9b 는 D4-2 의 반전이 측정되고 있음을 보인다.** 「조사 불가는 fatal」로 되돌리면
  ENOTDIR 한 번이 다시 `Restarting (1)` 을 만들고, 그것을 잡는 테스트가 실재한다.

### 원복 검증 (내용 동일성 + 심볼 대조)

매 뮤테이션 뒤 드라이버가 **원문 문자열 전체와의 동일성**을 먼저 확인하고
(`path.read_text() != original` 이면 즉시 중단), 그다음 아래 심볼 개수가 baseline 과
정확히 일치하는지 확인했다. 12회 전부 일치.

```text
engine.go        engineStrategyProjectionStart           3
                 reportStrategyProjectionDegraded        3
                 engineStrategyProjectionDegradedEvent   3
                 EnqueueAlert                            0   ← D3-2 이후 0 이 계약이다
                 "if strategyRuntime != nil {"           1
                 ectx.Notifier.Notify                    1
                 "marker.Ready("                         1
                 "lock.Release()"                        1
httpapi.go       "inspect strategy runtime projection"   0   ← 6.8① 이후 0
                 "…연결하지 못했다"                        1
                 "…확인할 수 없다"                         1
                 "return fmt.Errorf"                     0   ← 전략 경로에 fatal 없음
httpapi_reader.go strategyprojection.UnavailableSnapshot 1
                 strategyprojection.DormantSnapshot      1
                 r.strategyRuntime.Read                  1
                 "return nil, err"                       7   ← B1~B7 의 fail-closed
```

`git diff --exit-code -- cmd/tossctl/engine.go cmd/tossctl/httpapi.go
cmd/tossctl/httpapi_reader.go` 로도 라운드 종료 시점의 무잔재를 확인했다.

**원복 방식에 대한 주의(전 라운드에서 배운 것):** 커밋 전 작업본에 `git checkout` 을
쓰면 GREEN 까지 지워지고 `git diff --quiet` 가 그것을 초록으로 통과시킨다. 그래서
이번 드라이버는 파일 원문을 **메모리에 들고** 되돌린다.

## §원 라운드 (폐기·정정 기록)

baseline: `1d9451e0` + `cb2df8b3`. 9건 전부 최소 하나를 죽였고 생존은 0 이었다.
그중 셋은 **뒤집힌 계약을 지키던 것**이라 폐기한다.

| # | 원 뮤테이션 | 판정 |
|---|---|---|
| M2 | 강등은 하되 durable 알림을 지운다 (stderr 한 줄만 = 「기동 1회 유실형」) | **폐기.** 그 durable 알림이 A2 가 찾은 결함 그 자체다. 오늘의 계약은 정반대이며 `M2b` 가 반대 방향을 잰다 |
| M3 | 알림 event key 에서 실행 토큰을 빼 디렉터리로 고정 | **폐기.** dedup 토큰 기계가 outbox 와 함께 사라졌다. 잴 대상이 없다 |
| M9 | 비-NotExist `os.Stat` 오류까지 강등 (강등의 상한 제거) | **폐기.** 그 「상한」이 원 D4 이고 D4-2 가 그것을 철회했다. `M9b` 가 반대 방향을 잰다 |

### 원 라운드가 **틀리게** 주장했던 것 (A2 정정)

- 「`M2` 는 **원장 행을 보는 테스트가 없으면 조용한 강등 = 은폐가 통과했을 것**임을
  보인다」 — 전제가 틀렸다. 은폐를 막는 표면은 원장 outbox 말고도 있었다(콘솔·httpapi
  의 dormant 전략 화면). 그리고 원장 행이 만드는 **비용**(다음 부팅의 진입 차단)을
  세지 않았다.
- 「그 행은 전달될 때까지 PENDING 으로 남는다 = **미해소 유지**」 — 미구현이었다.
  transport 가 살아 있으면 ~2초 뒤 DELIVERED 로 사라진다. 그것을 「확인」한 테스트는
  전달 루프가 없는 harness 에서 PENDING 을 본 것이고, 그것은 「유지」의 측정이 아니다
  (A2 F2). Fix 라운드는 같은 harness 특성을 **반대로** 쓴다: 전달 루프가 없으니
  어떤 행이든 남는다 → 「0」은 「안 썼다」의 증거다.

### `internal/execgw` 선례 오독 정정 (A2 F1)

원 라운드의 코드 주석은 이 이벤트를 등급표에 올리지 않는 것을 **execgw 의 park 알림
선례**로 정당화했다. 그것은 선례가 아니다:

- execgw 의 park 이벤트는 obs 등급표에 **등재돼 있다**(criticalEvents 멤버).
- 그것이 entry gate 를 막는 것은 **의도된 판정**이다 — 주문 실행 경로가 막힌 상태에서
  새 진입을 여는 것이 위험하기 때문이다.

a108 의 강등은 정반대다. 막힌 것은 **화면**이고 실행 경로는 멀쩡하다. 그러므로 같은
rail 을 태우는 것은 선례 따르기가 아니라 **분류 착오**다. 이 정정은 코드 주석
(`engineStrategyProjectionDegradedEvent`)에도 반영했다.

### RED 재구성 사실 기록 (A2 F5)

원 라운드의 RED 는 git 히스토리에 없었다 — 테스트와 구현이 한 커밋(`1d9451e0`)에
들어갔다. 원장과 branch-test-map 이 적은 RED 서술은 **작업 중 관측의 사후 재구성**이며,
그 진위를 독립적으로 받친 것은 A2 의 M1 재현이다(같은 뮤테이션·같은 죽은 테스트).

Fix 라운드는 그것을 고쳤다: 테스트만 담은 커밋 `d8b27021` 이 GREEN 커밋 `aecc03e0`
보다 **먼저** 있고, 그 커밋 하나를 체크아웃하면 5건이 실패한다.

## 이 원장이 덮지 못한 것 (숨기지 않고 적는다)

- **겹1(`internal/strategyprojectionrpc`)은 T1 소유라 뮤테이션하지 않았다.** T2 테스트는
  실패를 seam 으로 주입하므로 T1 이 회수 규칙을 어떻게 바꾸든 이 원장의 결론은 바뀌지
  않는다 — 반대로 말하면 **회수의 전체성 자체는 이 원장이 증명하지 않는다**(tasks 2.4·6.5).
- **S3 의 강등은 아직 뮤테이션으로 잴 수 없다.** 그 경로의 판정은 T1-fix 의 `Dial`
  connect probe 가 진다. 지금 이 원장이 가진 것은 「그 probe 가 없으면 강등이 발동하지
  않는다」는 RED 하나다.
- **강등 상태의 httpapi 가 실제로 `strategyRuntime = nil` 인지는 여전히 간접 측정이다.**
  다만 Fix 라운드가 그 nil 의 **결과**는 직접 재게 됐다 — `httpAPIReader.Snapshot` 이
  nil 을 dormant 로, 죽은 reader 를 unavailable 로 그리는 것을 M11 이 지킨다.
- **비-unix 빌드의 위험은 존재하지 않는다** (A2 F9 — 원 라운드의 위험 항목 **삭제**).
  두 테스트 파일이 `//go:build unix` 인 것은 맞지만, 그 빌드에서 문제가 되는 지점은
  7단계 projection 이 아니다: `internal/enginelock/flock_other.go:16` 의
  `ErrLockUnsupported` 가 **부팅 1단계**(`enginelock.Acquire`, engine.go:196)에서 기동을
  거절하므로 7단계에 도달할 수 없다. 원 라운드는 「비-unix 는 측정하지 않았다」를
  위험으로 적었는데, 그것은 도달 불가능한 코드에 대한 걱정이었다.

---

## §gstack 라운드 (2026-08-14, Fix-First A1~A6) — 7건 적용, 7건 사망

baseline: 커밋 `fbb34fa1`(+ 뮤테이션 하나를 닫은 `e525d1d2`). 드라이버는 파일 원문을
메모리에 들고 되돌리고, 매 회 원문 동일성과 아래 심볼 개수를 확인했다.

```text
reportStrategyProjectionDegraded 3 · engineStrategyProjectionDegradedEvent 4 ·
engineStrategyProjectionStart 3 · EnqueueAlert 1(테스트 대조군 전용) ·
ectx.Notifier.Notify 1 · context.WithoutCancel 1 · unavailableStrategyRuntime 6 ·
strategyRuntimeReaderFor 4 · UnavailableSnapshot 1 · DormantSnapshot 1 ·
r.strategyRuntime.Read 1 · "return fmt.Errorf" 1(sentinel 의 Read) · MUTANT 0
```

### RED 관측 (GREEN 보다 먼저)

```text
느린 알림 publisher 하나가 부팅을 3s 넘게 붙잡았다 — 손절 루프는 rt.Run 안에서
  시작하고, 이 보고는 그 앞에 있다
KR/US 판정 = "NOT_CONFIGURED", want "RUNTIME_UNAVAILABLE"   (dial 실패·S3 두 행)
  같은 표의 「descriptor 가 없다」 대조군은 그때도 통과했다
```

### 원장

| # | 뒤집은 판정 | 뮤테이션 | 죽은 테스트 |
|---|---|---|---|
| M13 | A1 보고의 실행 자리 | `Notify` 를 다시 **동기**로 | `TestTheDegradedBootDoesNotWaitForTheNotifier` |
| M14 | A1 보고의 존재 | `Notify` 호출을 통째로 삭제 | `TestTheDegradedBootDoesNotWaitForTheNotifier` |
| M15 | A1 보고의 수명 | `WithoutCancel` 제거 — 부모 ctx 상속 | `TestTheDegradedBootDoesNotWaitForTheNotifier` (아래 주) |
| M16 | A2 강등의 신호값 | dial 실패를 다시 `nil` 로 | `TestADialFailureRendersUnavailableRatherThanNotConfigured/{descriptor_는_있는데_못_붙는다,주인_없는_socket_파일이_남았다}` |
| M17 | A3 경고의 존재 | 조사 불가의 `Fprintf` 를 무력화 | `TestAnUninspectableDescriptorDegradesLikeTheConsole` |
| M18 | A2 두 값의 구분 | **부재까지** sentinel 로 접는다 | `TestADialFailureRendersUnavailableRatherThanNotConfigured/descriptor_가_없다` |
| M19 | 겹3 흡수의 실재 | 전략 읽기 자체를 건너뛴다 | `TestADialFailureRenders…/{2행}`, `TestALiveStrategyProjectionThatDiesLeavesTheAggregateStanding` |

**M13·M14 가 A1 의 짝이다.** 하나는 「기다린다」를, 하나는 「조용해진다」를 만든다 —
비동기화의 흔한 실패 모드가 둘 다 후자로 미끄러지는 것이므로, 같은 테스트가 양쪽을
잡는 것이 계약이다(핀의 ①·②).

**M16·M18 이 A2 의 짝이다.** 하나는 장애를 배포 선택으로 접고, 하나는 배포 선택을
장애로 접는다. 어느 쪽이든 운영자는 화면에서 둘을 구별할 수 없다.

### 살아남았다가 닫은 것 — M15

첫 실행에서 **M15 가 살아남았다.** 핀이 재던 것은 ①늦지 않는다 ②닿는다 둘뿐이었고,
`WithoutCancel` 이 지키는 성질(③ 종료가 보고를 끊지 않는다)은 부팅 ctx 가 테스트에서
한 번도 취소되지 않아 관측 대상이 아니었다. 가짜 publisher 에 `ctx.Done()` 기록을
더하고, 명령이 돌아온 뒤 부팅 ctx 를 취소한 다음 발행이 여전히 살아 있는지 보는
절을 추가해서 닫았다(커밋 `e525d1d2`). 그 뒤 M15 는 죽는다.

이것이 [[passing-test-is-not-evidence]] 의 전형이다 — 세 성질을 담은 구현에 두 성질만
재는 핀을 붙이면, 나머지 하나는 **코드에만 있고 측정에는 없다.**
