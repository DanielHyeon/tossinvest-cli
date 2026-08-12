# Branch Test Map: `Runtime.Run`

**GREEN 칸은 호출 지점을 실측해서 채운다** — 덮이지 않은 것을 덮였다고 적지 않는다.
`ast.json`의 열거가 정본이다: **편집 후** 분기 5 · 이탈 3.

> **⛔ 이 표의 예측이 신설 갈래를 `B5`라고 적고 있었다. 실측은 `B4`다.**
> 갈래를 가운데 넣으면 **뒤의 id 가 전부 밀린다** — 옛 `B4`(`gracefulStop`)가 `B5`가 됐다.
> 아래는 실측 id 로 다시 적은 것이다. 조건은 하나도 안 바뀌었다.

| Branch | Scenario | Test | RED observed | GREEN observed |
|---|---|---|---|---|
| B1 | `:289` `Recover`가 배선돼 있다 | 기존 회복 테스트(`cmd/tossctl` 조립 경로, `//go:build tossos_testseams`) | no | **yes (기존)** |
| B2 | `:294` 회복이 실패하면 루프를 하나도 안 띄운다 | 기존 | no | **yes (기존)** |
| B3 | `:304` **배달 실행자는 `Loops`에 안 들어간다** — 감독 루프 집합이 안 는다 | **a098 R12** — `LoopNames()`가 **오늘과 같은 넷**(`reconcile`·`exit`·`filldetect`·`strategy-entry`)을 낸다 | **no — 오늘도 넷이다**. 이것은 **회귀 핀**이다 | **yes (오늘)** · 보조는 여기 안 나타난다: `TestAnAuxiliaryExecutorIsNotInTheSupervisedLoopSet` |
| B4 | **신설** `:322` — `range r.opts.Auxiliary`. 세 성질을 한 자리가 진다 — **① 실행자가 실제로 뜬다** · **② `stops`에 안 보낸다** · **③ `wg`에 들어간다** | ① `TestEveryAuxiliaryExecutorIsStarted` · **`TestTheEngineStopsPromptlyWithTheDelivererAttached`**(런타임이 진짜 실행자를 띄운다) — ② **a098 R2** `TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine` + **대조군** `TestTheDelivererAsASupervisedLoopStopsTheEngine` — ③ **a098 R14** `TestTheRuntimeDrainsAnAuxiliaryExecutorBeforeItReturns` | **yes — 성질마다 다른 뮤테이션이 잡았다.** ① 뮤테이션 B(launcher 삭제) → 네 테스트가 기동 대기 타임아웃 · ② 뮤테이션 C(`stops <- stop{…}` 삽입) → `ErrLoopFailed` · ③ 뮤테이션 A(`wg.Add(1)` 삭제) → `panic: sync: negative WaitGroup counter`. **④ 4.2 가 본문을 채웠다** — `r.runAuxiliary(**loopCtx**, aux)`가 패닉을 잡고 `loopCtx`로 판정하고 죽음만 `OnStop`에 넘긴다 (**R3·R13·R15·R18**: `TestAStoppedDelivererLocksTheEntryGate` · `TestCancellingTheRuntimeDoesNotLockTheEntryGate` · `TestASupervisedLoopFailingDoesNotLockTheEntryGate` · `TestAPanickingDelivererDoesNotKillTheEngine`). 뮤테이션 셋이 각자 다르게 잡았다 — G(`recover` 삭제) → **테스트 바이너리가 죽는다** · H'(판정만 부모 ctx) → **R13② 하나만** FAIL · I(판정 생략) → R13 ①②가 FAIL | **yes (2026-08-12)** |
| B5 | `:351` **취소 경로로 나간다 — 그리고 상한 안에 나간다** | **a098 R7** — `TestTheEngineStopsPromptlyWithTheDelivererAttached` (가짜 시계를 **한 번도 안 전진시킨다**) | **yes** (아래 ⚠) | **yes** |
| 이탈 `:379` | **배달 실행자가 반환해도 여기로 오지 않는다** | **a098 R2** | **yes — 뮤테이션 C 가 정확히 이 이탈로 나갔다** | **yes** |

> **✅ 뮤테이션 넷이 이 표의 GREEN 칸을 만들었다 (2026-08-12).**
> 넷 다 **다른 자리에서** 잡혔다 — A 는 `wg.Done()`의 패닉, B 는 네 테스트의 기동 대기
> 타임아웃, C 는 `ErrLoopFailed` 단언, F(조립 검증 삭제)는 네 하위 케이스 전부.
> **한 뮤테이션이 여러 테스트를 죽이는 것과, 여러 뮤테이션이 각자 다른 테스트를 죽이는
> 것은 다른 증거다.** 뒤엣것이라야 각 테스트가 **자기 성질**을 진다고 말할 수 있다.

> **⛔ `:331` 줄의 RED는 오늘 관측할 수 없다 — 그것을 「덮였다」로 적지 않는다.**
> 결정 9-2 아래에서 배달 실행자는 `stops`에 **아무것도 안 보낸다.** 그래서
> *"배달 실행자가 반환해도 `:331`로 안 간다"*는 **구조적으로 참**이고,
> 그것을 반증하는 오늘의 상태가 없다 — 오늘은 배달 실행자 자체가 없다.
>
> **RED를 만들려면 잘못된 구현을 먼저 짜야 한다**: 배달 실행자를 `Loops`에 넣어
> `TestALoopReturningOnItsOwnStopsEverythingAndIsCritical`(`runtime_test.go:171`)이
> 그것을 잡는 것을 보고, 그다음 `Auxiliary`로 옮겨 R3가 초록이 되는 것을 본다.
> **그 두 관측을 안 하면 R3는 born-GREEN이다** — 아무것도 안 한 코드도 통과한다.
>
> ## ✅ 그 두 관측을 실제로 했다 (2026-08-12, task 4.1)
>
> **위 문단이 요구한 것은 R2이고 R3가 아니다** — 이 칸이 두 요구를 한 이름으로 적고
> 있었다. `:379`로 안 가는 것은 **R2**(다른 루프가 산다)이고, **R3**는 그 죽음이
> **게이트를 잠근다**이며 `OnStop`이 없는 지금은 아직 아무도 안 진다(4.2·4.3).
>
> | 관측 | 어떻게 | 결과 |
> |---|---|---|
> | 잘못된 구현이 실제로 엔진을 내린다 | `TestTheDelivererAsASupervisedLoopStopsTheEngine` — **같은 함수**를 `Loops`에 등록 | `ErrLoopFailed` · exit 루프 정지 · critical 1건 |
> | 옳은 구현은 안 내린다 | `TestAnAuxiliaryExecutorThatReturnsDoesNotStopTheEngine` — 등록 자리만 바꾼다 | `Run` 이 안 깨어남 · exit 루프 생존 · critical 0건 |
> | 그 차이가 **`stops` 한 줄** 때문이다 | **뮤테이션 C** — 보조 launcher 에 `stops <- stop{…}` 삽입 | 뒤엣것이 **FAIL**, `:379` 이탈로 나갔다 |
>
> **셋째 줄이 없으면 앞의 둘은 「두 테스트가 서로 다른 결과를 낸다」까지만 말한다.**
> 그 차이의 **원인**이 등록 자리라는 것은 원인을 되돌려 봐야 나온다.

> **⚠ R7의 RED를 어떻게 관측하는가** (19라운드 B-P6이 이 칸을 고치게 했다).
> *"유한 시간에 반환"*으로 쓰면 **나쁜 구현도 통과한다** — 유한한 `time.Sleep(d)`은
> 끝나므로 결국 반환한다. RED가 성립하려면 **상한을 걸어야 한다**:
> 가짜 시계로 `alertFlushInterval` 안에 `Runtime.Run`이 반환하는 것을 본다.
>
> 오늘 배달 루프가 없으므로 이 테스트는 **루프를 만든 직후에 의미가 있다.**
> born-GREEN을 피하려면 **취소를 안 보는 루프를 먼저 짜서 `wg.Wait()`(`:300`)가
> 안 돌아오는 것을 보고**, 그다음 취소 가능한 대기로 고쳐 상한 안에 들어오는 것을 본다.
>
> **✅ 그 형태로 관측했다 (2026-08-12).** 취소를 안 보는 실행자는 지어낸 것이 아니라
> `TestTheRuntimeDrainsAnAuxiliaryExecutorBeforeItReturns`의 **실행자 자체**다 —
> 그것이 풀리기 전에는 `Run`이 반환하지 **않는 것**을 200ms 창으로 보고, 풀린 뒤
> 2초 상한 안에 반환하는 것을 본다. 상한 쪽은
> `TestTheEngineStopsPromptlyWithTheDelivererAttached`가 **진짜 실행자**로 지고,
> 거기서 가짜 시계를 **한 번도 안 전진시킨다** — 취소 불가능한 `Sleep(2s)` 구현은
> 그 테스트를 통과할 수 없다.
>
> ⚠ 4.0의 `alertDeliveryInterval`(2 s)이 옛 이름 `alertFlushInterval`을 대신한다.

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
