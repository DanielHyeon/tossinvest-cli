# a092 · 알림이 손절을 붙잡지 않는다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a092`
- **Spec**: `exit-policy` (MODIFIED 1) · `engine-safety` (MODIFIED 1)
- **위험 등급**: **High-risk** (§0.3·§0.5 — 손절 경로의 동기 체류와 알림 배선)

> **작성 순서**: 이 문서의 분기 주장은 전부 `analysis/function-logic/`의 AST 산출물에서
> 나왔다. 산출물이 문서보다 **먼저** 만들어졌다 (`.claude/CLAUDE.md`「단계 건너뛰기 금지」).
> `check_analysis` PASS — 함수 **23개**.
>
> **10판이다.** 아홉 판이 아홉 라운드에서 거부됐다. 판별 이력은 `review.md`가 정본이고
> 여기에는 **형태**만 적는다 — 그 형태가 9판이 무엇을 다르게 하는지를 정하기 때문이다.
>
> | 판 | 거부된 형태 |
> |---|---|
> | 1 | 범위가 배달 루프까지 가서 landed 테스트 16건을 재작성시켰다 → a093 분리 |
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
> `check_analysis` PASS — 함수 **23개**.

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

**a092는 알림 하나의 예산을 정한다. 사이클 총합은 정하지 않는다** — 총합을 유계로
만들려면 전송이 사이클 밖으로 나가야 하고 그것이 a093이다.

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
**주석 문구의 정정은 a093이 한다**(a092는 `runtime.go`를 편집하지 않는다).

### 실측이 말하는 것과 말하지 않는 것

전문은 `analysis/delivery-latency.md`.

- **창은 배선 커밋이 아니라 프로세스 재시작이다.** 3판은 커밋 날짜(2026-08-04)로 창을
  잡았는데, 08-05 00:11의 로그가 `no notification publisher is configured`
  (`notifier.go:253`, `Publisher == nil`일 때만 닿는다)를 쓴다 — **돌던 프로세스**에
  publisher가 없었다(코드가 없었는지 설정이 없었는지는 로그가 가르지 못한다). 창은 **engine.log line 6866**(2026-08-05T00:36:18Z 재시작) 이후다
- 측정 가능한 표본 **6건**, 전부 exit 관측 루프 생산이고 **전부 창 안이다**
- **냉연결 표본은 1건뿐이고 그것이 최댓값(0.754 s)이다.** 나머지 5건은 온연결(0.198~0.282 s)
- **그런데 냉이 이 워크로드의 정상 상태다** — 창 안에서 publish 유발 줄 **39개 중 18개(46%)**가
  90초(`http.Transport.IdleConnTimeout`) 넘게 떨어져 있다. 그 18개 중 측정 가능한 것이 1개다
- **창 안에서 실패한 publish는 한 건도 없다.** 34초는 유도된 상한이지 관측된 값이 아니다

> **3판의 "47개 중 24개(51%)"는 틀렸다.** 창이 24.5시간 일러 publish가 일어날 수 없던
> 줄을 포함했고 종류 열거도 불완전했다. **5판의 확정치**: `.Notify(`에 닿는 종류는
> **16종**이고, 그중 flatten 4종은 프로덕션에서 `Notifier`가 nil이라 로그 전용이므로
> (`cmd/tossctl/flatten.go:247`) 실제 publish 유발은 **12종**이다.
> **방향은 살아남는다: 46% ≈ 절반.**

## What Changes

### 전송 예산을 조립하는 자리에 적는다

`Attempts` · `RetryDelay` · `Ntfy.Timeout` 세 필드를 엔진 조립부에서 **채운다.**

```text
transport = 3 × 1.3s + 2 × 150ms = 4.200s   ← critical, 오늘 34s
reserve   = 5.000s − 4.200s      = 800ms    ← outbox·게이트·승격 몫
정상      = 1 × 1.3s             = 1.3s     ← 오늘 10s
```

> **이 세 줄은 `tools/check_values.py`가 검사한다.** 산술을 평가하고, 입력이 채택
> 상수인지 본다. 6판이 여기 적었던 `2 × 0.2s = 4.2s`는 **기각된 후보 #2의 재시도
> 지연이었고 산술도 안 닫혔다**(3×1.3+2×0.2 = 4.3s). 두 frozen 문서가 손절 경로
> 프로덕션 상수에 서로 다른 값을 말한 것이고, 그것이 6라운드 차단 B1이다.
> **값을 유도하는 자리는 design D2이고 이 등식은 그 결과를 기계 검사 아래 다시 적는다.**

**예산은 주기와 같아서는 안 된다.** 같은 `Notify` 호출 안에 전송 말고도 **다섯 가지**가
더 있고 그것이 0이 아니기 때문이다: `EnqueueAlert` 1회(outbox 기록),
`MarkAlertAttemptFailed` 시도 수만큼(시도별 실패 기록), `Gate.Block`(게이트 래치),
**소진 시 `n.escalate`의 `EscalateOperatingMode` 승격 트랜잭션**(`notifier.go:187`),
그리고 **그 호출이 쓰는 구조화 로그 줄**(`notifier.go:131`·`:279` 항상,
`:228`은 승격이 실제로 일어났을 때). 3판은 5.0s를 그대로 써서 주기를 넘었고(design D2의 회차 표),
4판이 transport를 내리고 남는 몫에 **이름을 준다**(`alertOverheadReserve`).

**그중 셋만 journal 커넥션 하나**(`SetMaxOpenConns(1)`)**에 줄을 선다.**
`Gate.Block`은 아니다 — `internal/execgw/retry.go:498-505`가 `g.mu` 뮤텍스와 맵 쓰기뿐이고
원장에 닿지 않는다. 6판은 이것을 "그 다섯이 전부 줄을 선다"고 적었고 **같은 거짓이
design D3의 프로덕션 Go 주석 초안에도 들어가 있었다**(6라운드 H2).

**측정은 이 문서가 갖지 않는다.** 채택 상수의 체류·초과분·마진은
`analysis/delivery-latency.md` §7.2가 정본이고, 그 값이 회차마다 흩어진다는 것과
유도 규칙이 무엇을 입력으로 받는지는 §7.2.1과 design D2가 쓴다.
6판이 여기 옮겨 적은 "10회 · `GOMAXPROCS=2` · 초과분 28.9ms"는 **채택하지 않은 후보 #2 회차의 조건**이었고 `28.9`는 어떤 측정 표에도 없었다(6라운드 차단 B2). <!-- rejected-value -->

**세 필드 중 둘만 동작을 바꾼다.** `Attempts`에 넣는 값은 `obs.DefaultCriticalAttempts`와
같은 3이므로 **no-op이다** — 채우는 이유는 예산 세 항이 조립부 한 자리에서 읽히게
하기 위한 것이지 동작을 바꾸기 위한 것이 아니다. 그렇게 적는다.

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
  - `internal/app/engine/notifications.go` — 예산 상수 **6개**(`alertBudget`,
    `alertPublishAttempts`, `alertPublishTimeout`, `alertRetryDelay`,
    `alertTransportBudget`, `alertOverheadReserve`) + 유도 주석 + 컴파일 타임 단언 **6줄**
    (`- 1` 넷은 0과 음수를, `/time.Millisecond` 둘은 **단위 누락**을 잡는다),
    그리고 `resolveNotificationPublisher`의 `&obs.Ntfy{...}` 리터럴에 `Timeout`.
    **`import "time"` 추가 필요** (현재 import는 `os`·`strings`·`config`·`obs`뿐)
  - `internal/app/engine/exitwiring.go` — `newNotifier`에 `Attempts`·`RetryDelay`
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
- **§0.3**: 손절을 **덜** 지연시킨다. 판정·수량·가격은 무변경
- **§0.4**: 브로커 요청 무변경 · **§0.9**: 임계·가격·수량 무변경

### 새로 생기는 결합 하나

`alertBudget = DefaultExitObservationInterval`을 컴파일 타임 단언이 강제하므로,
누가 `DefaultExitObservationInterval`을 **4.2초 이하로** 내리면 **빌드가 깨진다.**
그것이 이 단언의 목적이지만 **대가이기도 하다** — 관측 주기를 바꾸려는 사람은 알림
상수 세 개를 다시 유도해야 한다. 실패 메시지가 `alertOverheadReserve`를 이름으로
말하므로 어디를 봐야 하는지는 드러난다.

> **경계는 "4초 아래"가 아니라 "4.2초 이하"다**(7라운드 M-3). transport가 4.200초이므로
> 주기를 4.2초로 두면 reserve가 **정확히 0**이 되고, 그것은 두 delta가 `SHALL NOT`으로
> 금지한 상태다. 7판의 단언(`var _ [alertOverheadReserve]struct{}`)은 그 0을
> **통과시켰다** — `[0]struct{}`가 합법이기 때문이다. 8판의 단언은 `- 1`이 붙어
> 0에서 깨진다. 실측 두 줄(design D3에 표 전부):
>
> ```text
> 주기 4.2초 → invalid array length alertOverheadReserve - 1 (constant -1 …)
> 주기 4.0초 → invalid array length alertOverheadReserve - 1 (constant -200000001 …)
> ```

또한 실효 주기는 `o.Interval()`(`exitloop.go:325-331`)이고 예산은 **기본값**에 묶인다.
프로덕션은 `opts.Interval`을 설정하지 않으므로 오늘은 같지만, 누가 설정하면 둘이 갈린다.
**a092는 이 drift를 막지 않는다** — 막으려면 상수가 아니라 실행 시점 검사가 필요하고
그것은 이 change의 범위 밖이다.

## Non-goals

- **사이클 총 체류의 상한** — 알림당 예산으로는 도달할 수 없다. 전송이 사이클 밖으로
  나가야 하고 그것이 **a093**이다
- **`n.mu` 교차 루프 경합** — `deliver`가 재시도 전체에 걸쳐 뮤텍스를 쥐고
  `*obs.Notifier`는 다섯 루프가 공유한다(`analysis/notify-reach.md`).
  a092가 줄이는 것은 **쥐고 있는 시간**(34s→4.2s)이고 경합 구조는 그대로다 → a093
- **배달 루프와 outbox backlog 재시도** → a093. `Notifier.Flush`는 프로덕션 호출자가
  0곳이고 PENDING 9행은 한 번도 재시도되지 않았다
- **`obs` 패키지 함수의 편집** — 세 필드는 이미 있다
- **`ExitObserver.alert`·`ReconcileDriver.alert`에 ctx 기한 씌우기** — 호출자 수준 기한은
  `BeginTx` 전에 만료돼 원장 트랜잭션이 **시작조차 안 된다**(design D1 안 B)
- **CLI 시험 발송**(`cmd/tossctl/notificationsettings.go:151`) — 엔진 루프가 아니고,
  4판의 engine-safety SHALL은 범위를 엔진 루프로 한정한다. 결정과 대가는 design D6
- **`alertDeliveryBound` 주석의 문구 정정, `runtime.go:415`의 무기한 대기** → a093
- **`o.Interval()` drift 방어** → 범위 밖 (위 Impact에 기록)
- **`EventExitProposalCapped` 승격** → a091 (a092 뒤에 발효)
- **관측 누락 계측** → a090 · **outbox 재발 장부** → a089 · **보호 청산의 가격** → a087

## 미해결

- **1.3초는 측정이 아니라 판단이다.** 냉연결 표본이 **1건**(0.754 s)이고 1.3초는 그것의
  **1.72배**다. 냉이 정상 상태인데(39건 중 18건) 측정은 하나다. design D2가 대안 배분을
  표로 놓고 이 판단을 정면으로 쓴다
- **reserve 800ms의 근거는 로컬 프로브다.** 운영 디스크의 fsync가 더 느리면 초과분이
  커진다. 유도 규칙은 **전 회차 관측 최댓값의 2배**이고 그 판정값은 design D2와
  `delivery-latency.md` §7.2.1이 갖는다. **어느 한 회차의 마진을 여기 옮겨 적지 않는다** —
  6판이 4판의 절대 마진을 그대로 들고 있던 것이 6라운드 H1이고, 그 규칙은 5판에서
  이미 폐기됐다. 남는 사실은 규칙이 아니라 **측정의 지역성**이다: 로컬 SQLite·로컬 루프백에서
  잰 값이고, 운영 기계에서 다시 재는 것이 tasks 7.2의 산출이다
- **a092는 미전달을 늘린다.** (1.3s, 10s] 구간에서 성공했을 전달이 실패로 바뀌고,
  **서버가 받았는데 응답이 늦은 경우는 중복 발송 + 거짓 게이트 래치가 된다**(design D4).
  **a093 전까지 그 행을 재시도하는 것은 없다**
