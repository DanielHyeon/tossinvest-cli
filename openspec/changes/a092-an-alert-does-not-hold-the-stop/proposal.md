# a092 · 알림이 손절을 붙잡지 않는다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a092`
- **Spec**: `exit-policy` (MODIFIED 1) · `engine-safety` (MODIFIED 1)
- **위험 등급**: **High-risk** (§0.3·§0.5 — 손절 경로의 동기 체류와 알림 배선)

> **작성 순서**: 이 문서의 분기 주장은 전부 `analysis/function-logic/`의 AST 산출물에서
> 나왔다. 산출물이 문서보다 **먼저** 만들어졌다 (`.claude/CLAUDE.md`「단계 건너뛰기 금지」).
> 17판이 더한 함수 9개(`Journal.ClaimAlertForDelivery`·`claimOwed`·`Journal.PendingAlerts`·
> `Journal.MarkAlertDelivered`·`Journal.MarkAlertAttemptFailed`·`Journal.UndeliveredCount`·
> `EntryGate.Block`·`NewRuntime`·`Notifier.Acknowledge`)의 `ast.json`도 이 문서보다 먼저
> 만들어졌다 — D0.3의 뮤텍스 논증이 그 열거 위에 서 있기 때문이다.
>
> ---
>
> ## 17판은 재설계다 — 전송을 exit 관측 루프 밖으로 옮긴다
>
> **16판까지의 a092는 전송을 루프 안에 두고 예산으로 묶으려 했다.** 16라운드가 그 전제를
> 깼다(A-1): 세 필드를 채우는 것은 *"명시된 안전 목표를 달성하지 않는다"*. 예산은 전송에만
> 기한을 씌우고 원장 쓰기·로그·뮤텍스 대기에는 못 씌우므로 체류 상한이 아니고, 예산 안에
> 들어가려면 시도를 1회로 줄여야 하는데 그 1회가 a096의 3회 계약을 거짓으로 만든다(A-3).
>
> **답은 이 change의 두 spec 델타가 세 번 적어 두었다** — *"전송이 루프 밖으로
> 나가야 한다"*(engine-safety 16판 :28·:30·:34, exit-policy 16판 :28). 열여섯 판이
> 그것을 **범위 밖**에 두었고, 그래서 열여섯 판이 도달하지 못했다.
>
> **그리고 그 기계는 이미 있다.** `Notifier.Flush`(`internal/obs/notifier.go:427`)가
> 자기 주석에 *"It is what a supervising loop calls periodically"*라고 적혀 있고,
> 프로덕션 호출자가 **0**이다(테스트 3곳뿐). 재설계는 만드는 것이 아니라 **배선하는 것**이다.
>
> **그 부재는 오늘 살아 있는 결함이다.** `Gateway.parkAlert`(`internal/execgw/replay.go:534`)가
> outbox에 넣는 `order.unresolved_in_doubt`는 기록되고 **전송되지 않는다.** 도달 경로는
> `internal/reconcile/recovery.go:318` → `ReplayInDoubt` → `parkAlert`이고, 프로덕션 주석
> `replay.go:107`의 *"the notifier's Flush picks the row up and delivers it"*는 **거짓이다.**
> a092가 이 주석을 참으로 만든다. 유도와 대가는 `design.md` D0.
>
> ---
>
> **17판이다.** 열여섯 판이 열여섯 라운드에서 거부됐다. 판별 이력은 `review.md`가 정본이고
> 여기에는 **형태**만 적는다 — 그 형태가 다음 판이 무엇을 다르게 하는지를 정하기 때문이다.
>
> **9~11판의 형태는 하나였다: 검사 도구가 리뷰의 대상이 됐다.** 7판이 값의 사본을 막으려고
> `check_values.py`를 만든 뒤, 7·8·9·10·11라운드의 차단이 거의 전부 그 도구와 그 도구를
> 검사하는 도구 안에서 나왔다. 열한 판 동안 **프로덕션 Go는 0줄**이고, 이 change가 고치려는
> 손절 경로의 34초 체류는 HEAD에 그대로 있다.
>
> **12판이 도구를 강등하자 열한 라운드가 못 낸 것이 나왔다.** 12라운드(codex, 교차 모델)가
> **P 9건**을 냈다 — 그 전 열한 라운드의 P는 **0건**이었다. 가설은 확인됐다.
>
> **그러나 12판의 기제는 반증됐다.** 12판이 도구 대신 넣은 `rg` 검사는 *정답의 부재*를
> 물었고, 그래서 **틀린 값을 평범한 ASCII로 적으면 통과했다.** 게이트에 배선되지도
> 않았다. 13판은 그 검사를 **철회한다**(tasks 4.42) — 문서 값 일치는 기계로 강제하지
> 않고, 강제는 컴파일러·고정 테스트·사람 리뷰 셋이 한다.
>
> **13판의 실질은 P1이다: 엔진 루프의 전송 시도를 3회에서 1회로 줄인다.** 12판까지의
> 설계는 어떤 지연 구간에서 **손절을 더 늦췄다**. 근거와 표는 `tasks.md` §4.38~4.45.
>
> | 판 | 거부된 형태 |
> |---|---|
> | 1 | 범위가 배달 루프까지 가서 landed 테스트 16건을 재작성시켰다 → 미배정 후속 분리 |
> | 2 | 열거와 측정 다섯 군데가 틀렸다 |
> | 3 | 창을 커밋 날짜로 잡았고, 예산을 주기와 **같게** 잡아 자기 테스트가 자기 spec을 반증했다 |
> | 4 | `n.mu` 배수를 안 셌고, reserve 프로브가 journal 경합을 뺐다 |
> | 5 | 훑기 표를 23행으로 만들었으나 **한 행이 자기 산출물을 잘못 보고**했다 |
> | 6 | **정정이 리뷰어가 가리킨 그 문서에만 착지하고 같은 값이 사는 다른 문서에는 안 갔다** |
> | 7 | **6판을 고치려고 만든 방어 두 개를 돌려 보지 않았다** — 그리고 여섯 판이 못 본 기제 결함이 하나 있었다: 컴파일 단언이 `reserve == 0`을 허용한다 |
> | 8 | **고친 방어의 *새로 들어간 부분*을 자체 시험이 한 건도 안 덮었다** — 그리고 로그 열거 정정이 요구 문단에만 착지하고 같은 파일 Scenario에는 안 갔다 |
>
> **6라운드의 진단이 작업 단위를 바꾼다.** 여섯 판의 정정 단위는 "값"이 아니라
> "리뷰어가 준 file:line"이었고, 그래서 매번 아직 지목되지 않은 문서에서 같은 오류가
> 살아남았다. 6라운드 차단 3건 중 **2건이 이 문서 하나**에서 나왔다 —
> proposal이 design·analysis의 수치를 **복제**하고 있었기 때문이다.
>
> 그래서 7판의 첫 산출물은 문서 편집이 아니라 `tools/check_values.py`이고,
> **이 문서는 수치의 정본이 아니게 된다.** 유도는 design D2, 측정은
> `analysis/delivery-latency.md`가 정본이며 여기 남는 등식은 그 스크립트가
> **산술과 입력 상수를 기계로 검사하는 형태**만이다.
>
> **7라운드의 진단은 그 산출물 자신을 향했다.** 차단 5건 중 4건이 7판이 새로 만든
> 두 방어 안에 있었다 — 스크립트는 자기가 잡는다고 적은 값을 못 잡았고, 프로브는
> 컴파일되지 않았다. **방어를 만들었으나 돌려 보지 않은 것**이다. 그래서 8판은
> 검사를 고치는 것과 **그 검사를 우회해 보는 것**을 같은 task로 묶는다
> (`tools/check_values_selftest.py`). 다섯 번째 차단은 종류가 달랐다 —
> 컴파일 단언이 `reserve == 0`을 허용했고 그것이 여섯 판을 살아남은 유일한
> **기제** 결함이다.
>
> **`check_analysis`는 17판에서 PASS다** (2026-08-09 측정 —
> `evidence complete or diff-proven exempt`). 산출물은 **36개**이고, 17판이 더한
> 9개와 앞서 남아 있던 4개, 합쳐서 13개의 산문을 task 3.5가 채웠다.
> **AST가 먼저**라는 조건은 그 전에 이미 지켜졌다 — 이 문서의 분기 주장은
> 산문이 아니라 `ast.json` 위에 서 있었고, 3.5는 그 열거를 사람이 읽는 형태로
> 옮긴 것이다. **둘은 다른 조건이고 둘 다 충족됐다.**
>
> 3.5가 **새 항목 셋**을 만들었다: R17-12(`EntryGate.Block` 멱등성 미관측) ·
> R17-13(이미 배달된 행이 주기를 끊는다) · `ReasonAlertUndelivered` 래치의
> **프로덕션 해제 경로 부재**. 마지막 것은 a092가 만든 결함이 아니라 발견한
> 결함이고, 새 운영 표면은 a092의 범위 밖이다 (review 17.9).
>
> ---
>
> ## 18판 — 17라운드가 재설계 자체에서 P를 냈고, 셋을 사용자가 결정했다
>
> 17라운드는 codex 두 보이스 병렬로 돌았고 **둘 다 BLOCK**(A: P7·T0, B: P7·T5,
> 고유 P 11). 16라운드까지의 P가 *"목표를 달성하지 않는다"*였다면 17라운드의 P는
> **"재설계 자체가 틀렸다"**이다. 사용자 결정 3건으로 18판이 답한다.
>
> | 결정 | 답 | 어디에 |
> |---|---|---|
> | 세고-푸는 구간의 배제 (A-P1) | **잠금을 남기되 원격 전송 위에서는 안 잡는다** | design D0.3a · engine-safety 델타 |
> | 배달 루프가 죽으면 (A-P5) | **엔진을 내리지 않는다.** 게이트 래치 + 모드 승격 | design D0.9 · engine-safety 델타 |
> | a092의 범위 (A-P7) | **끝까지 구현한다.** non-goal 넷 철회 | design D0.10 · 아래 목록의 취소선 |
>
> 저자가 결정 없이 고친 것: **A-P4**(두 델타가 일반 등급의 내구 기록에서 모순 —
> engine-safety를 정본으로 갈랐다) · **A-P3**(≈16.5s가 상한이 아니다 — 큐 대기를
> 넣어 **≈44.5s**로 다시 쓰고 굶주림 방지 규칙을 SHALL로 넣었다) ·
> **B-P5·P6·P7**(task 3.5가 만든 거짓 커버리지 주장 셋 — 원본 대조로 확인하고 고쳤다).

## Why

exit 관측 루프는 원격 알림의 왕복을 **동기로, 최대 34초** 기다린다. 그리고 `Run`은
사이클이 끝난 **뒤에** 주기를 잔다.

```go
// exitloop.go:353-363 — AST calls 순서가 :358 → :359
o.reportCycle(o.ObserveOnce(ctx))                       // :358  여기 34초가 들어갈 수 있다
if err := o.clk.Sleep(ctx, o.Interval()); err != nil {  // :359  그리고 5초를 더 잔다
```

**체류는 주기를 대체하지 않고 늘린다.** — 근거: `internal-app-engine--exitobserver.run`
AST (calls 5, `:358`→`:359`).

| 사이클 | 동기 체류 | 실제 관측 간격 | 의도(5s) 대비 |
|---|---|---|---|
| 알림 없음 | ~0 | 5s | 1× |
| critical 알림 1건 | **34s** | 39s | **7.8×** |
| critical 알림 N건 | **34N s** | 34N+5 s | — |

### 사이클은 알림을 여러 번 올릴 수 있다

`ObserveOnce` B5 `:453`의 `range states`가 포지션을 순차 순회하고, `judge`가 알림
헬퍼들을 부른다. 그중 **`alertProposalRefused`(`exitloop.go:1548-1565`)는 억제가 없다** —
그 함수의 `ast.json`이 **`"branches": null` · `"returns": null`**를 내놓는다. 조건문이 0개이므로
조기 반환도 중복 억제도 없다. 호출자는 `submit` 안의 `:1264`·`:1309` 두 곳이고,
포지션마다·레벨마다 올라간다. 프로덕션 로그가 그것을 보인다
(005930, engine.log line 6896·6899·6910·6912).

**"래치가 없다"는 부재에 관한 주장이고, 손으로 읽어서는 증명되지 않는다.**
AST 열거만이 그것을 말한다. — 근거:
`analysis/function-logic/internal-app-engine--exitobserver.alertproposalrefused/`.

⇒ **한 사이클의 최악 체류는 예산 × 그 사이클이 올린 알림 수**이고, 그 수는 보유 포지션
수에 비례한다.

**16판까지 여기에는 "a092는 알림 하나의 예산을 정하고 사이클 총합은 정하지 않는다 —
총합을 유계로 만들려면 전송이 사이클 밖으로 나가야 하고 그것이 미배정 후속이다"라고
적혀 있었다.** 17판은 그 후속을 이 change 안으로 들인다. 사이클 총합이 **유계가 되는
것은 아니다** — 알림 수에 비례하는 항은 남는다. 바뀌는 것은 그 항이 `원격 왕복 × 시도 +
시도 간 대기`에서 **로컬 SQLite 쓰기 하나**가 되는 것이다. 위 표의 34s·34N s는
`로그 한 줄 + outbox 트랜잭션` × N이 된다.

### 두절 사이클은 오늘 한 번 알린다 (2판 정정)

2판은 두절 사이클이 `Notify`에 두 번 도달해 68초라고 썼다. **오늘은 한 번이다.**

`TransitionOperatingMode`의 AST가 B14 `switch@409` → **B15 `case direction == 0`
→ `return :415`**를 열거한다. `AnnounceOperatingMode`는 `:479`이고 `tx.Commit()`은
`:468`이다 — **반환은 announce보다 64줄 앞, 커밋보다 53줄 앞이다.**
`ModeTriggerExitObservationOutage`의 목표는 `ENTRY_BLOCKED`(`:537-549`)이고,
계정은 **2026-07-31T09:55:49부터 ENTRY_BLOCKED이며 완화된 적이 없다**
(live journal `operating_modes` 1행, read-only 확인).

로그가 같은 말을 한다: `AnnounceOperatingMode`가 쓰는 모양의 줄(`from_state` 필드를 가진
`engine.operating_mode`)은 **전체 로그에 line 372 하나뿐**이다. — 근거:
`analysis/delivery-latency.md` §1.3, `…--journal.transitionoperatingmode` FLM.

⇒ **68초는 NORMAL 계정의 상한이고 오늘의 헤드라인은 34초다.**
그리고 두절 알림은 `o.outageRaised`로 래치되므로(`checkOutage` B3 `:775-778`)
**두절 에피소드당 1회**지 사이클마다가 아니다.

### 34초의 유도 — 세 필드가 조립부에서 비어 있다

```text
critical 최악 = Attempts × (publish 1회 상한) + (Attempts-1) × RetryDelay
              = 3        × 10s                + 2            × 2s        = 34s
정상 최악     =            publish 1회 상한                               = 10s
```

| 항 | 실효값 | 기본값을 고르는 분기 | 조립부 |
|---|---|---|---|
| `Attempts` | 3 | `deliver` B1 `:245` | `exitwiring.go:71-81` `newNotifier`가 **안 채운다** (AST branches 0, calls 0) |
| `RetryDelay` | 2s | `wait` B1 `:292` | 같은 자리에서 **안 채운다** |
| publish 1회 | 10s | `Ntfy.Publish` B3 `:96`, B7 `:122` | `resolveNotificationPublisher`의 `&obs.Ntfy{...}` 리터럴이 **`Timeout`을 안 채운다** |

상위 ctx에 deadline은 없다 — `cmd/tossctl/engine.go:274`·`runtime.go:272`는 취소 전용,
프로덕션 clock은 실시계.

### 같은 저장소가 이미 이 판단을 한 번 내렸다

`Runtime.alert`(`runtime.go:444-456`)는 `context.WithTimeout(..., alertDeliveryBound)`로
감싼다. 주석은 **"finite so a dead transport cannot hold the shutdown open"**이다.
**종료 경로에 씌운 기한을 관측 경로에는 씌우지 않았다** — `ExitObserver.alert`의
AST defers는 0이고 `ReconcileDriver.alert`도 0이다.

부수적으로: `alertDeliveryBound = 30s`인데 최악은 34초다. **30 < 34, 주석이 틀렸다.**
그리고 같은 감독자의 **두 번째** 동기 전송 대기(`runtime.go:415`
`EscalateOperatingMode`)에는 기한이 아예 없다 — `Runtime.escalate` AST가 `r.alert`(`:396`)와
`EscalateOperatingMode`(`:415`)를 둘 다 열거하고 앞만 `alertCtx`를 받는다.
**주석 문구의 정정은 a092가 한다**(tasks 8.5) — 12판까지 이 자리는 미배정 후속으로 미뤘고
13라운드 P7이 그것을 기각했다. `"three bounded publish attempts"`를 거짓으로 만드는 것이
a092의 편집이므로, 자기가 만든 거짓 진술을 다음 change에 넘기는 것은 침묵한 생략이다.
**주석 한 줄이고 `Runtime.alert`의 AST는 바이트 단위로 그대로다.**
`runtime.go:415`의 무기한 대기는 여전히 미배정 후속이다 — 그것은 a092가 만든 것이 아니다.

### 실측이 말하는 것과 말하지 않는 것

전문은 `analysis/delivery-latency.md`.

- **창은 배선 커밋이 아니라 프로세스 재시작이다.** 3판은 커밋 날짜(2026-08-04)로 창을
  잡았는데, 08-05 00:11의 로그가 `no notification publisher is configured`
  (`notifier.go:253`, `Publisher == nil`일 때만 닿는다)를 쓴다 — **돌던 프로세스**에
  publisher가 없었다(코드가 없었는지 설정이 없었는지는 로그가 가르지 못한다). 창은 **engine.log line 6866**(2026-08-05T00:36:18Z 재시작) 이후다
- 측정 가능한 표본 **6건**, 전부 exit 관측 루프 생산이고 **전부 창 안이다**
- **냉연결 표본은 1건뿐이고 그것이 최댓값이다.** 나머지 5건은 온연결이고 한 자릿수 배
  빠르다. 여섯 표본의 값은 `analysis/delivery-latency.md`가 정본이다
- **그런데 냉이 이 워크로드의 정상 상태다** — 창 안에서 publish 유발 줄 **39개 중 18개(46%)**가
  90초(`http.Transport.IdleConnTimeout`) 넘게 떨어져 있다. 그 18개 중 측정 가능한 것이 1개다
- **창 안에서 실패한 publish는 한 건도 없다.** 34초는 유도된 상한이지 관측된 값이 아니다

> **3판의 "47개 중 24개(51%)"는 틀렸다.** 창이 24.5시간 일러 publish가 일어날 수 없던
> 줄을 포함했고 종류 열거도 불완전했다. **5판의 확정치**: `.Notify(`에 닿는 종류는
> **16종**이고, 그중 flatten 4종은 프로덕션에서 `Notifier`가 nil이라 로그 전용이므로
> (`cmd/tossctl/flatten.go:247`) 실제 publish 유발은 **12종**이다.
> **방향은 살아남는다: 46% ≈ 절반.**

## What Changes

### 17판 — 전송을 exit 관측 루프 밖으로 옮긴다

**델타가 커진다. 그것을 숨기지 않는다.** 16판까지 a092는 조립부 세 필드를 채우는
change였고 *"`obs` 패키지의 함수는 한 줄도 바꾸지 않는다"*가 그 selling point였다.
17판은 `internal/obs`의 배달 경로를 바꾼다. High-risk 패키지의 제어 흐름 변경이고,
따라서 Function Logic Map과 Branch Test Map이 면제 없이 필수다.

| 파일 | 무엇이 바뀌나 | 제어 흐름 |
|---|---|---|
| `internal/obs/notifier.go` `notifyCritical` | claim까지 하고 반환한다. `deliver` 호출이 사라진다 | **바뀐다** |
| `internal/obs/notifier.go` `claimAndDeliver` | claim만 남는다. **배달 뮤텍스는 계속 잡되 그 아래는 로컬 원장 작업뿐이다** (D0.3a — 18판이 D0.3의 결론을 철회했다) | **바뀐다** |
| `internal/obs/notifier.go` `Flush` | 배치 상한 · 행당 1시도 · 시도 소진 시 래치와 승격 · nil publisher를 실패 시도로 기록 | **바뀐다** |
| `internal/obs/notifier.go` `publishBestEffort` | 비차단 이관. 버퍼가 차면 버리고 기록한다 (D0.8) | **바뀐다** |
| `internal/obs/notifier.go` `deliver` · `wait` | 인라인 재시도가 배달 루프로 간다 | **바뀐다** |
| `internal/app/engine/notifications.go` | 상수 집합이 D0.6으로 교체된다 | 변화 없음 |
| 엔진 조립부 | 알림 배달 `SupervisedLoop`을 등록한다 | 추가 |
| `internal/execgw/replay.go:107` | 주석이 **참이 된다**. 편집이 아니라 검증 대상이다 | 변화 없음 |

**`obs`의 공개 계약은 유지된다.** `Notify`는 여전히 durable 기록 실패에서만 오류를
반환하고(`notifier.go:118-123`), 전달 실패의 결과 셋(outbox 행 보존 · 신규 진입 차단 ·
운영 모드 강화)도 그대로다. 바뀌는 것은 그 결과가 **어느 goroutine에서 언제** 일어나느냐다.

**바뀌지 않는 것을 먼저 적는다.** 등급 표(`event.go:279-299`), outbox 스키마,
`ClaimAlertForDelivery`의 재무장 판정(`claimOwed`), CAS 술어(`WHERE state = PENDING`),
CLI 시험 발송 10초(D6), 그리고 두 spec 델타가 *"이 요구가 만드는 것은 예산이지 보장이
아니다"*라고 적은 부분.

---

### 16판까지의 What Changes (기록 — 채택 근거로 인용하지 않는다)

> 아래는 전송이 루프 **안에** 있다는 전제 위의 설계다. D0.7이 무효 범위를 적는다.

### 전송 예산을 조립하는 자리에 적는다

`Attempts` · `RetryDelay` · `Ntfy.Timeout` 세 필드를 엔진 조립부에서 **채운다.**

```text
alertPublishAttempts = 1                                              ← 13판이 바꾼 것
alertTransportBudget = alertPublishAttempts × alertPublishTimeout
                     + (alertPublishAttempts − 1) × alertRetryDelay  ← 계수가 0이다
alertOverheadReserve = alertBudget − alertTransportBudget             ← outbox·게이트·승격 몫
정상 등급             = 1 × alertPublishTimeout
```

> **시도가 하나인 이유는 예산이 아니라 안전이다.** 12라운드가 12판의 3회 설계에서
> **손절이 더 늦어지는 구간**을 찾았다 — publish 지연이 `alertPublishTimeout`보다 크고
> `alertTransportBudget`보다 작으면, 오늘은 1회 성공인데 12판 설계는 세 번 타임아웃하고
> 실패한다. 시도가 하나면 체류는 `min(실제 지연, alertPublishTimeout)`이고 **어떤
> 지연에서도 오늘 이하다.** 표는 `tasks.md` §4.38.
>
> **이 문서는 그 상수들의 값을 적지 않는다.** 값은 **이름으로만** 참조한다 — 유도는
> **design D2**, 강제는 **`notifications.go`의 컴파일 타임 단언**.
> 12판은 이 규칙을 `rg` 검사로 강제하려 했고 **12라운드가 그 검사를 깼다**
> (틀린 값을 적으면 통과한다). 13판은 검사를 철회하고 규칙만 남긴다 — 강제는
> **컴파일러·고정 테스트·사람 리뷰**가 한다(tasks 4.42).
>
> 6판이 여기 적었던 `2 × 0.2s`는 **기각된 후보 #2의 재시도 지연이었고 산술도 안
> 닫혔다.** 두 frozen 문서가 손절 경로 프로덕션 상수에 서로 다른 값을 말한 것이고,
> 그것이 6라운드 차단 B1이다. 7판은 그 재발을 **검사로** 막으려 했고 다섯 라운드가
> 그 검사의 우회를 찾았다. **이름은 우회할 수 없다** — 사본이 없으면 갈라질 것이 없다.

**예산은 주기와 같아서는 안 된다.** 같은 `Notify` 호출 안에 전송 말고도 **다섯 가지**가
더 있고 그것이 0이 아니기 때문이다: `ClaimAlertForDelivery` 1회(outbox 기록),
`MarkAlertAttemptFailed` 시도 수만큼(시도별 실패 기록), `Gate.Block`(게이트 래치),
**소진 시 `n.escalate`의 `EscalateOperatingMode` 승격 트랜잭션**
(`notifier.go:222` 호출 → `:312` 원장 호출),
그리고 **그 호출이 쓰는 구조화 로그 줄**(`notifier.go:148` `logEvent`의 Warn ·
`:399` `deliver` 소진 Error는 항상, `:323` `escalate`의 승격 Warn은 승격이 실제로
일어났을 때). 3판은 transport를 `alertBudget`과 **같게** 잡아
주기를 넘었고(design D2의 회차 표), 4판이 transport를 내리고 남는 몫에 **이름을 준다**
(`alertOverheadReserve`).

> **19판이 이 문단의 좌표 다섯과 이름 하나를 다시 고정했다 (HEAD `285c7619` 기준).**
> `notifier.go:187`·`:131`·`:279`·`:228`은 a097이 `notifyCritical`을 다시 쓰기 전의
> 줄번호였다. 그리고 **이름 하나는 좌표가 아니라 주장이 낡은 것이다**: a097 이후
> `notifyCritical`은 `EnqueueAlert`를 부르지 않는다. outbox 기록은
> `claimAndDeliver`(`:194`)가 `ClaimAlertForDelivery`(`:244`)로 한다.
> `EnqueueAlert`는 `internal/execgw/replay.go:551` 한 곳만 남은 얇은 위임 wrapper다.
> 숫자만 갈아 끼웠으면 **틀린 함수를 가리키는 맞는 줄번호**가 됐을 것이다.
> 좌표 옆에 *무엇을 하는 줄인지*를 함께 적는 것은 그래서다 — 다음 표류를 다시 찾기 위해서.

**그중 셋만 journal 커넥션 하나**(`SetMaxOpenConns(1)`)**에 줄을 선다.**
`Gate.Block`은 아니다 — `internal/execgw/retry.go:498-505`가 `g.mu` 뮤텍스와 맵 쓰기뿐이고
원장에 닿지 않는다. 6판은 이것을 "그 다섯이 전부 줄을 선다"고 적었고 **같은 거짓이
design D3의 프로덕션 Go 주석 초안에도 들어가 있었다**(6라운드 H2).

**측정은 이 문서가 갖지 않는다.** 채택 상수의 체류·초과분·마진은
`analysis/delivery-latency.md` §7.2가 정본이고, 그 값이 회차마다 흩어진다는 것과
유도 규칙이 무엇을 입력으로 받는지는 §7.2.1과 design D2가 쓴다.
6판이 여기 옮겨 적은 "10회 · `GOMAXPROCS=2` · 초과분 28.9ms"는 **채택하지 않은 후보 #2 회차의 조건**이었고 `28.9`는 어떤 측정 표에도 없었다(6라운드 차단 B2). <!-- rejected-value -->

**세 필드가 전부 동작을 바꾼다.** 12판까지는 `Attempts`가 `obs.DefaultCriticalAttempts`와
같은 3이라 **no-op**이었다. 13판은 그것을 **1로 내린다** — 이것이 이 change에서 가장
큰 동작 변경이고, 손절 체류를 어떤 지연 구간에서도 오늘 이하로 만드는 유일한 항이다.

**`alertRetryDelay`는 실행되지 않는데도 채운다.** 회차가 하나이므로 `wait`에 닿지
않는다. 채우는 이유는 누가 `alertPublishAttempts`를 올렸을 때 빈 자리가
`DefaultRetryDelay`(2s)로 돌아가고 그 값이 예산 밖이기 때문이다. 주석에 그렇게 적는다.

**`obs` 패키지의 함수는 한 줄도 바꾸지 않는다.** 세 필드는 이미 존재하고 이미 문서화된
손잡이다(`notifier.go:92-96`, `ntfy.go:72-73`).

### 편집하는 함수는 둘이고 둘 다 제어 흐름이 안 바뀐다

| 함수 | AST 현재 | 편집 후 기대 |
|---|---|---|
| `newNotifier` (`exitwiring.go:71-81`) | **branches 0**, returns 1, calls 0 | branches 0, returns 1, calls 0 |
| `resolveNotificationPublisher` (`notifications.go`) | branches 5, returns 4 | branches 5, returns 4 |

## Impact

- **Specs**: `exit-policy` (MODIFIED 1) · `engine-safety` (MODIFIED 1)
- **Code**:
  - `internal/app/engine/notifications.go` — 예산 상수 **5개**(`alertFlushInterval`,
    `alertPublishAttempts`, `alertPublishTimeout`, `alertFlushBatch`,
    `alertLoopShare`) + 유도 주석 + 컴파일 타임 단언 **6줄**.
    정본은 **design D0.6의 `// notifications.go — 17판` 블록**(`design.md:378-402`)이고
    단언과 그 실측은 **`tasks.md` §8.2**다.
    그리고 `resolveNotificationPublisher`의 `&obs.Ntfy{...}` 리터럴에 `Timeout`.
    **`import "time"` 추가 필요** (현재 import는 `os`·`strings`·`config`·`obs`뿐)

    > **⚠⚠ 19판 4차까지 이 줄은 「상수 6개(`alertBudget`, `alertRetryDelay`,
    > `alertTransportBudget`, `alertOverheadReserve` …)」였다** — 19라운드 A-P7 = B-P3.
    > **17판이 그 넷을 지웠는데 Impact은 16판 델타를 계속 약속하고 있었다.**
    > Impact은 이 change가 **프로덕션에 무엇을 넣는지**를 말하는 자리이므로,
    > 여기가 틀리면 폭발 반경·되돌리기 단위·리뷰 대상이 전부 틀린 것을 가리킨다.
    >
    > **단언 여섯 줄의 성질도 여기서 과장하지 않는다.** 그 여섯은
    > **0·음수·단위 누락·단위 오타**를 잡고 **채택값의 근처는 안 지킨다** —
    > `X - k` 꼴은 `X = k`를 허용한다(`[0]struct{}`가 합법). 못 잡는 구성
    > **여덟 개를 §8.2가 실측으로 열거**하며, 그중 `alertPublishAttempts = 1`은
    > `engine-safety`의 SHALL NOT을 위반하면서 **BUILD OK**다.
  - `internal/app/engine/exitwiring.go` — `newNotifier`에 채우는 필드.
    **`Attempts`·`RetryDelay` 둘이 아니다** — `RetryDelay`가 읽는 `alertRetryDelay`는
    17판이 지웠고 `Notifier.wait`는 claim 경로에서 도달 불가능해진다(D0.6).
    **무엇을 채우는지는 §8.7의 형태가 정해지면 확정된다** — 19라운드 A-P2가
    §8.7과 D0.2의 모순을 열어 두었고, 그 결정 전에는 이 줄을 못 채운다.
    **빈칸에 이름을 붙여 둔다 — 침묵한 생략이 아니다.**
- **Tests (새로)** — 신설 파일 2개, 기존 `_test.go` 편집 없음:
  - `internal/app/engine/a092_alert_budget_test.go` (`package engine` — `newNotifier`와
    상수가 전부 unexported다) — 조립부 값 3건, 상수 값 고정, **알림 하나의 실시계 체류**
  - `cmd/tossctl/a092_testsend_source_test.go` — CLI 시험 발송이 범위 밖임을
    `go/parser` 소스 스캔으로 고정
- **Tests (기존)**: **재작성 0건으로 예상한다.** 조립부에 도달하는 테스트 중 어느 것도
  `Notifier.wait`에 닿지 않는다: 주입되는 유일한 publisher `countingPublisher`의 `fail`
  필드가 **저장소 어디에서도 설정되지 않고**, 나머지는 전부 `Publisher == nil`이라
  `deliver`가 첫 시도에서 `break`한다(`notifier.go:252-255`).
  **task 7.3이 이 예상을 실행으로 검증하고, 도달 테스트 수는 7.1.1의 프로브가 센다** —
  세지 않은 수를 여기 적지 않는다. 7.1의 `go test ./...`는 `ok`/`FAIL`만 내므로
  도달 수를 만들지 못한다(6라운드 H-3)
- **Schema**: **없음**
- **Config**: **없음** — 예산은 코드 상수이므로 runtime toggle이 아니고 audit 대상이 아니다
- **§0.3**: 손절을 **덜** 지연시킨다 — **알림 한 건에 대해, 어떤 publish 지연에서도
  오늘 이하다.** 12판까지는 이 문장에 구간 조건이 붙어야 했다(12라운드 P1).
  시도를 1회로 줄여 그 조건을 없앴다. 판정·수량·가격은 무변경
- **§0.3의 두 번째 경계 — 에피소드 누적으로는 오늘 이하가 아니다.** 13판 초안은
  위 문장을 "**어떤** publish 지연에서도"라고만 적었고, **그것은 실측으로 거짓이다.**
  관측 5회를 순차로 돌려 재면(design D2, 14판 재측정):

  | 서버 지연 | 오늘 | 13판(기각) | **14판 = `alertPublishTimeout` × 1회** |
  |---|---|---|---|
  | 500ms | 0.51s · DELIVERED | 0.51s · DELIVERED | **0.51s · DELIVERED** |
  | **1.795s** (실측 평균, 냉연결 20건) | 1.81s · DELIVERED | **6.55s · PENDING · 래치** | **1.81s · DELIVERED** |
  | 2.125s (첫 **8**건의 최댓값 — **표본이 작아 폐기된 값**) | 2.14s · DELIVERED | **6.55s · PENDING · 래치** | **2.14s · DELIVERED** |
  | **2.721s** (20건의 실측 최댓값) | 2.73s · DELIVERED | 6.55s · PENDING · 래치 | **2.73s · DELIVERED** |
  | 4.0s | 4.02s · DELIVERED | 6.56s · PENDING · 래치 | **17.57s · PENDING · 래치** |

  오늘은 발송이 **성공**하므로 `RemindAfter`(1시간)가 이후 관측을 억제해 체류가
  0이 된다. 발송이 실패하면 행이 PENDING으로 남고 `claimOwed`가 PENDING을 창 없이
  곧바로 owed로 돌려주므로(`outbox.go:277-279`) **매 관측이 다시 낸다.**
  **이것은 산문이 아니라 잰 값이다.**

  **14판이 13판과 갈리는 곳이 이 표다.** 13판의 1회 상한은 실측 전송 지연의 평균
  **아래**여서 정상 ntfy가 매 관측 실패했다. 14판의 상한은 실측 최댓값 **위**라서
  실측 구간 전체에서 **오늘과 같은 수**를 낸다. 누적 역행은 없어지지 않고
  13판 상한 위에서 시작하던 것이 **`alertPublishTimeout` 위로 밀려난다** — 4.0초 행이
  그 구간이고, 그 안에서는
  14판이 13판보다 **더 비싸다**(17.57s vs 6.56s). 그 교환을 택하는 근거는
  **실측된 지연이 어느 구간에 있느냐**다. 진입 차단은 보수 방향이고(안전 불변식 6)
  청산에는 영향이 없지만, **오늘 없던 차단이 새로 생기는 것**이므로 여기 적는다.
  RED는 6.12(R6)가 진다

  > **⚠ 13판이 여기 적었던 "그 승격은 재시작에도 남는다"는 거짓이다** — 13라운드
  > 보이스 B가 기각했고 재현했다. `ENTRY_BLOCKED` **기록**은 원장에 남지만
  > 새 프로세스의 게이트에 그것을 투영하는 프로덕션 경로가 없다
  > (`RestoreOperatingModeProjection` 호출자 0). 거짓 차단은 **재시작으로 풀린다** —
  > 영구가 아니라 프로세스 수명만큼이다.
- **§0.3의 경계 — 이 change가 유계로 만들지 못하는 것**: 위 문장은 **transport**에
  대한 것이다. 같은 호출 안의 `ClaimAlertForDelivery`·시도별 journal 쓰기·운영 모드
  승격·동기 로그에는 기한이 없고 journal은 커넥션 하나다. `alertOverheadReserve`는
  **관측된 최악의 배수로 잡은 예산이지 상한이 아니다**(12라운드 P2). 호출 전체에
  기한을 씌우는 길은 D1 안 B가 기각했다(`BeginTx` 전에 만료되면 원장 트랜잭션이
  시작조차 안 된다) → **미배정 후속**
- **§0.4**: 브로커 요청 무변경 · **§0.9**: 임계·가격·수량 무변경

### 새로 생기는 결합 하나

`alertBudget = DefaultExitObservationInterval`을 컴파일 타임 단언이 강제하므로,
누가 `DefaultExitObservationInterval`을 **`alertTransportBudget` 이하로** 내리면
**빌드가 깨진다.** 그것이 이 단언의 목적이지만 **대가이기도 하다** — 관측 주기를
바꾸려는 사람은 알림 상수 세 개를 다시 유도해야 한다. 실패 메시지가
`alertOverheadReserve`를 이름으로 말하므로 어디를 봐야 하는지는 드러난다.

> **경계는 "transport보다 작게"가 아니라 "transport 이하로는 안 된다"이다**(7라운드 M-3).
> 주기를 `alertTransportBudget`과 **같게** 두면 reserve가 **정확히 0**이 되고, 그것은
> 두 delta가 `SHALL NOT`으로 금지한 상태다. 7판의 단언
> (`var _ [alertOverheadReserve]struct{}`)은 그 0을 **통과시켰다** — `[0]struct{}`가
> 합법이기 때문이다. 8판의 단언은 `- 1`이 붙어 0에서 깨진다.
> **실측 표는 design D3이 갖는다** — 경계값과 컴파일러 메시지를 이 문서가 복제하면
> 그것이 6라운드 B1의 형태다.

또한 실효 주기는 `o.Interval()`(`exitloop.go:325-331`)이고 예산은 **기본값**에 묶인다.
프로덕션은 `opts.Interval`을 설정하지 않으므로 오늘은 같지만, 누가 설정하면 둘이 갈린다.
**a092는 이 drift를 막지 않는다** — 막으려면 상수가 아니라 실행 시점 검사가 필요하고
그것은 이 change의 범위 밖이다.

## Non-goals

- ~~**사이클 총 체류의 상한** → 미배정 후속~~ **18판: 부분적으로 a092 안으로.**
  유계가 되지는 않지만 비례 계수가 원격 왕복에서 로컬 원장 작업으로 바뀐다 (D0.4)
- ~~**`n.mu` 교차 루프 경합** → 미배정 후속~~ **18판: a092 안으로.**
  경합 구조가 아니라 **잠금이 덮는 구간**이 바뀐다 — 원격 전송 위에서 잡지 않는다 (D0.3a)
- ~~**배달 루프와 outbox backlog 재시도** → 미배정 후속~~ ~~**18판: a092 안으로.**~~
  **19판: `a098-nobody-sends-what-the-outbox-keeps`로 분리.**
  `Notifier.Flush`의 프로덕션 호출자가 0이고 PENDING 행이 한 번도 재시도되지 않은 것은
  미배정 후속이 아니라 **오늘 살아 있는 결함이다**(D0.1). 18판은 그것을 a092 안으로
  끌어왔고, 19판은 **그것만 따로 낸다** — 사용자 결정 3 = 안 1
- ~~**`obs` 패키지 함수의 편집** — 세 필드는 이미 있다~~ **18판: a092의 델타다.**
  `claimAndDeliver`·`Flush`·`publishBestEffort`가 바뀐다 (D0.10)

> **18판이 위 네 줄을 뒤집었다 (사용자 결정 2026-08-10, 결정 3 = 안 1).**
> 17판은 *"재설계는 만드는 것이 아니라 배선하는 것"*(위 :30)이라 쓰면서 동시에
> 그 배선이 필요로 하는 것들을 non-goal에 남겨 두었다. 17라운드 A-P7 = B-P1·P2가
> 그 모순을 잡았다 — **두 문장을 다 따르면 재설계가 구현되지 않는다.**
> 취소선은 지우지 않고 남긴다: 어느 판이 무엇을 범위 밖이라고 했는지가 증거다.
- **`ExitObserver.alert`·`ReconcileDriver.alert`에 ctx 기한 씌우기** — 호출자 수준 기한은
  `BeginTx` 전에 만료돼 원장 트랜잭션이 **시작조차 안 된다**(design D1 안 B)
- **CLI 시험 발송**(`cmd/tossctl/notificationsettings.go:151`) — 엔진 루프가 아니고,
  4판의 engine-safety SHALL은 범위를 엔진 루프로 한정한다. 결정과 대가는 design D6
- **`runtime.go:415`의 무기한 대기** → 미배정 후속 (`alertDeliveryBound` **주석 정정은
  13판에서 a092로 옮겼다** — tasks 8.5, 13라운드 P7)
- **`RestoreOperatingModeProjection`·`SetModeProjector` 배선** → **미배정 후속.**
  a092가 발견했고 a092가 만들지 않았다 — `EscalateOperatingMode`가 쓰는 `ENTRY_BLOCKED`는
  **산 프로세스에서도 재시작 후에도** 아무것도 막지 않는다(design D5의 표). a092는 이
  사실에 **의존하지 않도록** 문장을 걷어내는 것까지만 한다. 배선을 여기 얹으면
  되돌리기 단위가 둘이 된다

  > **19판이 이 줄을 강화했다 — 재측정 결과 공백이 더 넓다 (HEAD `285c7619`).**
  > 여기 있던 문장은 *"재시작 후 아무것도 막지 않는다"*였고, 그것은 공백을
  > **재시작 창으로 한정한다.** 측정값은 그렇지 않다.
  >
  > | 잰 것 | 값 |
  > |---|---|
  > | `SetModeProjector` 프로덕션 호출자 | **0** |
  > | 같은 함수 테스트 호출자 | 18 (파일 5개) |
  > | `ReasonOperatingModeBlocked` 래치를 세우는 유일한 자리 | `modegate.go:50` — `ProjectOperatingMode` 안 |
  >
  > `TransitionOperatingMode`는 커밋 **직후** projector를 부른다(`operating_mode.go:475-476`).
  > bind된 projector가 없으면 그 호출이 아무 일도 안 하므로, `ENTRY_BLOCKED` 행은
  > 재시작을 기다릴 것도 없이 **처음부터** 진입 게이트에 닿지 않는다.
  > 이것은 a092가 새로 만든 결함이 아니라 a092가 **의존하지 않기로 한 사실**이고,
  > 그 사실이 A-P4의 답을 바꾼다 (아래 design D5·tasks §8.7).
- **경로별 시도 횟수** → 범위 밖. 조립부는 하나이고(`gateway.go:280`) 그 산물을
  호출자 6곳이 공유하므로(tasks 7.5), 경로마다 다른 예산을 주려면 `obs.Notifier`에
  **필드**를 더하고 호출자 6곳이 저마다 그 값을 고르게 해야 한다. 대가는 design D5가
  경로별로 적는다

  > **19판이 이 줄의 근거를 바꿨다 (18라운드 B-P12).**
  > 여기 있던 이유는 *"a092가 `obs`를 한 줄도 안 건드리기로 한 것과 정면으로 충돌한다"*
  > 였고, 그것은 같은 문서 위 :435-436이 **이미 뒤집은 문장**이다 — 18판은
  > `claimAndDeliver`·`Flush`·`publishBestEffort`를 a092의 델타로 끌어왔다.
  > 뒤집힌 전제를 근거로 남겨 두면 범위 판단이 **없는 규칙**에 기대게 된다.
  > 진짜 이유는 그대로 남는다: 배달 루프 재설계는 **한 예산을 언제 쓰는가**를 바꾸고,
  > 경로별 시도 횟수는 **예산이 몇 개인가**를 바꾼다. 다른 결정이고 되돌리기 단위도 다르다.
- **`o.Interval()` drift 방어** → 범위 밖 (위 Impact에 기록)
- **`EventExitProposalCapped` 승격** → **어느 change도 소유하지 않는다.**

  > **⚠ 12·13판은 이것을 "→ a091"로 적었고 거짓이다** (14라운드 보이스 B R14-P7).
  > a091은 이 승격을 **명시적으로 기각한다** — design.md:12의 안 A가
  > *"부분 캡까지 critical이 된다"*를 이유로 버려졌고, proposal.md:108이 같은 것을
  > non-goal로 적는다. a091이 하는 것은 `EventExitProposalCapped`를 **부분 캡
  > 전용으로 좁히는 것**이지 승격이 아니다. **위임 대상을 확인하지 않고 적은
  > 포인터였다.** 아래 §미배정 후속에 옮긴다.
- **관측 누락 계측** → a090 · **outbox 재발 장부** → a089 · **보호 청산의 가격** → a087

### 형제 change와의 충돌 2건 — a092가 남의 문서를 고치지 않으므로 여기 적는다

15라운드 보이스 B N6이 찾았다. **둘 다 a092가 landed되는 순간 참이 아니게 된다.**
a092는 두 문서를 편집하지 않는다 — 고칠 이유가 그 change에 있고, 여기서 고치면
되돌리기 단위가 셋이 된다. **대신 무엇이 거짓이 되는지를 적는다.**

| 형제 | 그 문서가 적은 것 | a092 뒤에 |
|---|---|---|
| **a096** | `proposal.md:157` — *"`deliver`의 **3회·2초**와 소진 뒤 `ModeEntryBlocked`는 a074 계약"* | **3회·2초가 1회·(미실행)이 된다.** `ModeEntryBlocked` 부분은 그대로 참이다. a096의 리마인더가 "최초 전달과 완전히 같은 코드를 걷는다"는 문장도 참인 채로 남지만, **그 코드의 예산이 달라진다** — a096의 리마인더는 이제 3.5초 안에 못 나가면 실패다 |
| **a089** | `proposal.md:91` — `PendingAlerts`가 행을 안 돌려줘 *"`Flush`가 영영 재시도하지 않는다"* | **고쳐도 재시도는 안 일어난다.** `Notifier.Flush`(`notifier.go:427`)는 **프로덕션 호출자가 0**이고 테스트 3곳뿐이다. a089가 고치는 것은 `Flush`에게 **줄 것이 생기는 것**이지 `Flush`가 **불리는 것**이 아니다 |

**a089 쪽이 더 무겁다.** a092의 미배정 1번(「루프 밖 유계 재전송기」)이 없으면
a089의 수리는 **아무도 안 부르는 함수의 입력을 고치는 것**으로 끝난다. 그리고
1번을 `Flush` 배선으로 갚으면 **a092가 없애려는 버그가 되살아난다** —
`Flush`는 `n.mu`를 쥔 채(`:433-434`) `PendingAlerts(ctx, 0)`을 부르고
`outbox.go:392-404`가 `limit > 0`일 때만 `LIMIT`을 붙이므로 **0은 무제한**이다.
PENDING 9행이면 9 × 3.5s = **31.5초**다.

**그러므로 a089와 a092는 같은 미배정 항 하나를 함께 기다린다.** 어느 쪽도 그것을
자기 안에서 갚지 못하고, 그 사실이 두 change 어디에도 안 적혀 있었다.

## 미배정 후속 — a092가 남기는 의무 (소유자 없음)

**이 목록에는 change ID가 없다.** 12·13판은 이것들을 `a093`으로 위임했는데
**`a093`은 `openspec/changes/`에도 `archive/`에도 없다** — 만들어진 적이 없다
(14라운드 보이스 A P3 · 보이스 B R14-P3, 두 목소리가 독립으로 찾았다).
없는 change를 소유자로 적으면 그 의무는 위임된 것이 아니라 **사라진 것**이다.

**17판이 이 목록의 절반을 a092 안으로 들인다.** 16판까지 1·3·4·8이 미배정이었고, 그
넷이 전부 "전송이 루프 안에 있다"에서 나온 의무였다. 전송이 나가면 넷이 같이 닫힌다.

| # | 의무 | 17판 | 근거 |
|---|---|---|---|
| 1 | **루프 밖 유계 재전송기** | **a092 안으로** | D0.2·D0.6. 유계·감독·사이클당 1시도가 spec SHALL이 된다 |
| 2 | **`RestoreOperatingModeProjection`·`SetModeProjector` 배선** | 미배정 유지 | `ENTRY_BLOCKED`가 재시작 후 아무것도 막지 않는다. 전송 위치와 무관하다 |
| 3 | **`n.mu` 교차 루프 경합 해소** | **a092 안으로** | **18판 정정** — D0.3a. spec SHALL NOT은 *"기록 경로가 배달 잠금을 잡지 않는다"*가 아니라 **"배달 경로가 원격 전송 위에서 잠금을 잡지 않는다"**이다. 잠금은 남고 그 잠금이 덮는 구간이 로컬 원장 작업으로 줄어든다 (17라운드 A-P1) |
| 4 | **사이클 총 체류의 상한** | **부분적으로 a092 안으로** | 유계가 되지는 않는다. 비례 계수가 원격 왕복에서 로컬 쓰기로 바뀐다 (D0.4) |
| 5 | **`runtime.go:415`의 무기한 대기** | 미배정 유지 | 주석 정정만 a092(tasks 8.5) |
| 6 | **`Notify` 호출 전체에 기한 씌우기** | 미배정 유지 → **불필요해진다** | D1 안 B가 기각한 길이고, 17판에서는 씌울 대상이 로컬 쓰기뿐이다 |
| 7 | **`EventExitProposalCapped` 승격** | 미배정 유지 | a091이 명시적으로 기각. 다만 17판은 이 이벤트가 **루프를 붙잡던 것**을 고친다 (D0.8) |
| 8 | **`Notifier.Acknowledge` 배선** | **a092 안으로** | 배달 루프가 게이트를 걸면 그것을 푸는 경로가 같은 change에 있어야 한다. 오늘 호출자 0이고, 없으면 **재시작만이 해제 수단이다** |

> **⚠ 1번은 "`Flush`를 부른다"가 아니다 — 그렇게 하면 같은 버그를 다시 만든다.**
> 15라운드 보이스 B N3이 이것을 짚었고 17판이 소스와 AST로 재확인했다.
> `Notifier.Flush`(`notifier.go:427`)는 `n.mu`를 **잡은 채**(`:434-435`)
> `PendingAlerts(ctx, 0)`을 부르는데, `outbox.go:392-404`가 `if limit > 0`일 때만
> `LIMIT`을 붙이므로 **0은 무제한이다.** 그 전량을 잠금 안에서 동기로 하나씩 publish한다.
> PENDING이 9행이면 9 × 3.5s = **31.5초** 동안 `claimAndDeliver`가 막힌다 —
> **a092가 없애려는 34초가 a092가 지목한 해결책 안에 그대로 있다.**
> `internal-obs--notifier.flush/ast.json`이 같은 것을 열거한다(분기 6 · 반환 4 · 호출 9 ·
> `n.mu.Lock` + `defer n.mu.Unlock` · 행마다 `Publisher.Publish`).
>
> 그래서 17판이 배선하는 것은 **네 가지를 갖춘 배달 루프**다: 사이클당 행 수 상한
> (`alertFlushBatch`), 기록 경로가 배달 잠금을 잡지 않는 잠금 규율(D0.3), 사이클당
> 1행 1시도라는 재시도 규율(D0.6), 그리고 해제 경로(8번). 그 넷을 spec이 SHALL로 적는다.

**a095의 의존이 다시 참이 된다** (14라운드 보이스 B R14-P2). `a095 design.md:203`은
*"outbox와 재시도가 그 반복을 대신 책임진다"*를 억제 유지의 근거로 쓴다. 16판까지 a092는
재시도를 1회로 줄이고 outbox 행을 다시 보내는 것을 갖지 않아 **그 문장을 거짓으로 만들었다.**
17판에서는 시도가 3회로 돌아가고 배달 루프가 outbox 행을 다시 보내므로 **a095가 고칠 것이
없다.** a096의 3회·2초 계약(`a096/proposal.md:154`)과 *"pending 알림은 계속 재시도되어야
한다"*(`a096 engine-safety/spec.md:19`)도 같은 이유로 다시 참이다.

## 미해결

- **`alertPublishTimeout`은 측정이 아니라 판단이다.** 14판의 3.5초는 실제 전송기
  실측 **20건**(`https://ntfy.sh/` 읽기 전용 GET, 전부 냉연결)의 최댓값 위에 있다.
  비율과 값은 design D2와 `delivery-latency.md` §9.1이 갖는다.
  20건은 분포가 아니고, **그것은 topic POST가 아니라
  homepage GET이다** — 경계짓는 것은 네트워크 경로이지 publish 핸들러가 아니다.
  진짜 POST를 재려면 사용자 topic으로 **실제 알림이 발송된다**(휴대폰에 뜬다).
  그것은 사람의 승인 없이 하지 않는다. 재측정 절차는 tasks 9.8.

  > **13판의 1.3초는 이 측정이 반증했다** (14라운드). 1.3s는 실측 **평균 아래**여서
  > 정상 ntfy를 매 관측 실패로 기록했을 것이다 — design D2의 측정 표가 그것이다.
  > 그 구성은 "느린 전송기를 실패로 부른다"가 아니라 **정상 전송기를 실패로 부른다**였다.
- **`alertOverheadReserve`의 근거는 로컬 프로브다.** 운영 디스크의 fsync가 더 느리면
  초과분이 커진다. 유도 규칙은 **전 회차 관측 최댓값의 2배**이고 그 판정값은 design D2와
  `delivery-latency.md` §7.2.1이 갖는다. **어느 한 회차의 마진을 여기 옮겨 적지 않는다** —
  6판이 4판의 절대 마진을 그대로 들고 있던 것이 6라운드 H1이고, 그 규칙은 5판에서
  이미 폐기됐다. 남는 사실은 규칙이 아니라 **측정의 지역성**이다: 로컬 SQLite·로컬 루프백에서
  잰 값이고, 운영 기계에서 다시 재는 것이 tasks 7.2의 산출이다
- **a092는 미전달을 늘린다. 13판은 그것을 더 늘린다.** (`alertPublishTimeout`,
  오늘의 publish 상한] 구간에서 성공했을 전달이 실패로 바뀌고, **재시도가 없으므로
  일시적 흔들림 한 번이 곧바로 미전달이다.** 서버가 받았는데 응답만 늦은 경우는
  **중복 발송 + 거짓 게이트 래치**가 된다(design D4).
  **미배정 후속 전까지 그 행을 *루프 밖에서* 재시도하는 것은 없다.**

  **재시도를 대신하는 것은 다음 관측인데, 그것이 없는 경로가 여섯이다.**

  > **⚠ 12·13판은 여기에 "그것이 없는 경로가 하나 있다"고 적었고 거짓이다** —
  > 14라운드 보이스 B R14-P1이 기각했고 소스에서 확인했다. **틀린 것은 나다.**
  > `runtime.go:453` 하나만 셌는데, exit 루프의 critical 알림 **넷이 전부 래치를
  > 먼저 세우고 알린다** — `checkOutage`(`exitloop.go:777`) ·
  > `alertRefused`(`:1520`) · `announceQuarantine`(`exit_quarantine_announce.go:71`) ·
  > `noteDelay`(`:1579`). 래치가 서면 다음 관측은 `alert`에 닿기 전에 early return하므로
  > **PENDING 행이 다시 owed여도 아무도 다시 부르지 않는다.** 다섯째는 운영 모드
  > 전이다 — `direction == 0`이 `operating_mode.go:415`에서 early return하고
  > 그것은 투영(`:475`)보다 **앞이다.**

  PENDING 행이 창 없이 다시 owed인 것(`outbox.go:277-279`)이 재발송으로 이어지는 것은
  **래치 없는 두 경로뿐이다** — exit 판정(`exitwiring.go:103,141`)과 대사
  (`reconcileloop.go:556`). 나머지 여섯에서 **1회 실패는 곧 미전달이다.**
  그중 하나가 `exit.judgement_refused` — **손절도 평가되지 않는다**는 알림이다.
  경로별 표는 design D5, 기계적 전수는 tasks 7.5.

  **`ENTRY_BLOCKED`가 그 자리를 메우지도 않는다**: 기록은 원장에 남지만 새 프로세스의
  게이트는 그것을 읽지 않는다(`RestoreOperatingModeProjection` 프로덕션 호출자 0).
  **기록은 남고 강제는 사라진다.**

  이것이 14판이 고른 교환이다: **전달 신뢰도를 미배정 후속(§1번)까지 낮춘 채로 두고,
  그 대신 손절 체류를 관측당 조건 없이 오늘 이하로 만든다.** 진입 차단은 보수 방향이고
  청산은 무관하므로 안전 불변식이 이 방향을 가리킨다. **"조건 없이"가 걸리는 단위는
  관측이며 에피소드 누적이 아니다** — 누적은 §0.3 항목의 두 번째 경계가 적는다.
  **그리고 이 교환이 성립하려면 3.5초가 실측 위여야 한다** — 13판의 1.3초에서는
  성립하지 않았다
