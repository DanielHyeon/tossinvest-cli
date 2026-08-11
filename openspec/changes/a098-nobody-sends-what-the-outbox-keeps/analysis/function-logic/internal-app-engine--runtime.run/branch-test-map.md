# Branch Test Map: `Runtime.Run`

**GREEN 칸은 호출 지점을 실측해서 채운다** — 덮이지 않은 것을 덮였다고 적지 않는다.
`ast.json`의 열거가 정본이다: 분기 4 · 이탈 3.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:262` `Recover`가 배선돼 있다 | 기존 회복 테스트(`cmd/tossctl` 조립 경로, `//go:build tossos_testseams`) | no | **yes (기존)** |
| B2 | `:267` 회복이 실패하면 루프를 하나도 안 띄운다 | 기존 | no | **yes (기존)** |
| B3 | `:277` **배달 실행자는 `Loops`에 안 들어간다** — 감독 루프 집합이 안 는다 | **a098 R12** — `LoopNames()`가 **오늘과 같은 넷**(`reconcile`·`exit`·`filldetect`·`strategy-entry`)을 낸다 | **no — 오늘도 넷이다**. 이것은 **회귀 핀**이다 | **yes (오늘)** |
| **신설 B5** | B3 뒤 · `range r.opts.Auxiliary` — 배달 실행자가 여기서 뜬다 | **a098 R1** — 기록된 critical 알림이 유한 시간에 `DELIVERED`가 된다 | **yes — 오늘 배달 주체가 없으므로 행이 영원히 PENDING** | no |
| **신설 B5** (같은 자리, 다른 성질) | 그 goroutine이 **`wg`에 들어간다** — `stops`에는 안 보낸다 | **a098 R14** — 원장 쓰기가 끝난 뒤에 `Runtime.Run`이 반환한다 | **yes — `wg.Add(1)`을 뺀 채로 먼저 관측한다** | no |
| B4 | `:304` **취소 경로로 나간다 — 그리고 상한 안에 나간다** | **a098 R7** | **yes** (아래 ⚠) | no |
| 이탈 `:331` | **배달 실행자가 반환해도 여기로 오지 않는다** | **a098 R3** | **아래 ⛔ — 오늘 관측할 방법이 없다** | no |

> **⛔ `:331` 줄의 RED는 오늘 관측할 수 없다 — 그것을 「덮였다」로 적지 않는다.**
> 결정 9-2 아래에서 배달 실행자는 `stops`에 **아무것도 안 보낸다.** 그래서
> *"배달 실행자가 반환해도 `:331`로 안 간다"*는 **구조적으로 참**이고,
> 그것을 반증하는 오늘의 상태가 없다 — 오늘은 배달 실행자 자체가 없다.
>
> **RED를 만들려면 잘못된 구현을 먼저 짜야 한다**: 배달 실행자를 `Loops`에 넣어
> `TestALoopReturningOnItsOwnStopsEverythingAndIsCritical`(`runtime_test.go:171`)이
> 그것을 잡는 것을 보고, 그다음 `Auxiliary`로 옮겨 R3가 초록이 되는 것을 본다.
> **그 두 관측을 안 하면 R3는 born-GREEN이다** — 아무것도 안 한 코드도 통과한다.

> **⚠ R7의 RED를 어떻게 관측하는가** (19라운드 B-P6이 이 칸을 고치게 했다).
> *"유한 시간에 반환"*으로 쓰면 **나쁜 구현도 통과한다** — 유한한 `time.Sleep(d)`은
> 끝나므로 결국 반환한다. RED가 성립하려면 **상한을 걸어야 한다**:
> 가짜 시계로 `alertFlushInterval` 안에 `Runtime.Run`이 반환하는 것을 본다.
>
> 오늘 배달 루프가 없으므로 이 테스트는 **루프를 만든 직후에 의미가 있다.**
> born-GREEN을 피하려면 **취소를 안 보는 루프를 먼저 짜서 `wg.Wait()`(`:300`)가
> 안 돌아오는 것을 보고**, 그다음 취소 가능한 대기로 고쳐 상한 안에 들어오는 것을 본다.

## 이 함수에 **테스트를 새로 쓰지 않는 분기**

**B1·B2** — `Recover` 경로. 결정 5-1·9-2 어느 쪽도 그 조건을 안 바꾸고,
`TestRecoveryRunsBeforeAnyLoopStarts`(`runtime_test.go:488`)와
`TestAnIncompleteRecoveryStartsNothing`(`runtime_test.go:529`)가 이미 덮는다.
a098 task 4.6이 **그 콜백 안에** 게이트 복원을 넣지만 **이 함수의 분기는 안 바꾼다.**

## 덮이지 않은 것을 이름으로 적는다

- **`supervisorWG.Wait()`(`:301`)의 상한** — `superviseHealth`가 ctx 취소에
  반응하지 않으면 `wg.Wait()`와 같은 형태로 종료가 멈춘다. **a098은 그 함수를
  안 건드리므로 이 change의 RED 대상이 아니다.** 다만 R7이 재는 상한은
  `wg.Wait()`와 `supervisorWG.Wait()` **둘의 합**이므로, R7이 실패하면
  원인이 배달 루프가 아닐 수 있다 — 그때 이 줄을 본다.
- **`drain(stops)`(`:302`)의 내용** — a098은 배달 루프의 정지 사유가 이 목록에
  어떻게 나타나는지 단언하지 않는다. `not-applicable`: 이 change는 그 문자열을
  근거로 쓰지 않는다.
