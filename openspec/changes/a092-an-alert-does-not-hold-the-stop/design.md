# a092 설계 — 예산은 조립하는 자리에 적는다 (11판)

> **이 문서가 상수 유도의 정본이다.** proposal은 7판부터 수치를 복제하지 않고 여기를
> 가리킨다 — 6라운드 차단 2건이 그 복제에서 나왔다. 값이 문서마다 같은지는
> `tools/check_values.py`가 검사한다.

## D1 — 코드가 아니라 조립을 바꾼다

34초는 `obs`의 결함이 아니다. `obs`는 세 개의 손잡이를 **이미** 갖고 있고 문서화까지 했다.

| 손잡이 | 선언 | 기본값을 고르는 분기 |
|---|---|---|
| `Notifier.Attempts` | `notifier.go:92-94` | `deliver` B1 `:245` |
| `Notifier.RetryDelay` | `notifier.go:95-96` | `wait` B1 `:292` |
| `Ntfy.Timeout` | `ntfy.go:72-73` | `Publish` B3 `:96` |

34초는 **조립부가 셋 다 비워 둔 결과**다. 고칠 자리는 조립부다.

| 안 | 방법 | 문제 |
|---|---|---|
| A | `deliver`에 총 예산 기한을 넣는다 | High-risk 함수 본문 편집. `MarkAlertAttemptFailed`가 같은 ctx를 쓰므로(`:264`) 기한이 지나면 **원장 기록까지 실패한다** |
| B | `ExitObserver.alert`/`checkOutage`의 ctx에 기한을 씌운다 | 아래 |
| C | **조립부에서 세 필드를 채운다** | 함수 본문 0줄. `Notify`에 도달하는 7경로 전부에 걸린다 |

**C를 택한다.**

### 안 B의 기각 이유 — 2판이 부정확하게 썼다

2판은 "거기 기한을 씌우면 journal 트랜잭션이 잘린다"고 썼다. **틀렸다.**
`TransitionOperatingMode` AST가 `tx.Commit()` **B25 `:468`** → `AnnounceOperatingMode`
**B27/B28 `:478-479`** 순서를 열거한다 — announce는 트랜잭션 **밖**이다
(`internal-journal--journal.transitionoperatingmode` FLM).

정확한 이유는 둘이다:

1. **호출자 수준 기한은 그 아래 이탈들에 누적된다.** `checkOutage`가 `:780`의 알림에
   기한을 다 쓰면 `:796`의 `EscalateOperatingMode`는 **이미 만료된 ctx로 `BeginTx`
   (`operating_mode.go:391`)를 부른다** — 트랜잭션이 잘리는 게 아니라 **시작되지 않는다.**
   운영 모드 승격 자체가 사라진다
2. **경로를 다 덮지 못한다.** `Notify`에 도달하는 자리는 7곳이고
   (`analysis/notify-reach.md`) 그중 넷(P1·P3·P5·P6)은 exit 루프의 함수를 지나지 않는다

**C의 부작용은 균일성이다** — 예산이 대사 루프·Retrier·Guardian·감독자에게도 걸린다.
D5에서 각각의 대가를 계산한다.

## D2 — 숫자를 고른다. 그리고 무엇이 측정이고 무엇이 판단인지 갈라 쓴다

### 예산이 걸리는 것은 전송이 아니라 **호출**이다 — 3판이 여기서 틀렸다

`ExitObserver.Run` AST가 `:358 ObserveOnce` → `:359 Sleep(Interval)` 순서를 보인다.
체류는 주기에 **더해진다.** 그러므로 유계로 만들 것은 **한 알림이 사이클을 붙잡는
전체 시간**이고, 거기에는 전송 말고도 같은 `Notify` 호출 안의 작업이 들어간다:
`EnqueueAlert` 1회(outbox 기록) + `MarkAlertAttemptFailed` 시도 수만큼(시도별 실패 기록) + `Gate.Block`(게이트 래치) + **`n.escalate`의 `EscalateOperatingMode` 승격 트랜잭션**(`notifier.go:187` — 예산이 소진된 그 경로에서만 돈다) + **그 호출이 쓰는 구조화 로그 줄**(`notifier.go:131`·`:279`는 항상, `:228`은 승격이 실제로 일어났을 때만).

> **이 열거가 다섯 항인 것은 두 delta의 SHALL과 같아야 한다.** 9판까지 이 줄은
> 네 항이었다 — 로그 쓰기를 D3 주석과 두 delta에만 더하고 **유도 정본인 여기**와
> 제안서와 측정 정본에는 안 더했다(9라운드 차단 2). 유도 정본이 spec보다 적게
> 세면 그 차이만큼 예산의 근거가 비어 있다. `tools/check_values.py`의
> `check_dwell_enumeration`이 이제 이 다섯 문서를 전부 본다.

3판은 이 항을 빠뜨리고 등식을 이렇게 썼다.

```text
(3판, 거짓) Attempts × Timeout + (Attempts-1) × RetryDelay  =  Interval (5s)
```

**등호가 문제다.** 비전송 작업이 0이 아니므로 실제 체류는 **반드시** 주기를 넘는다.
3라운드 리뷰가 그 상태를 실측했다 — 네트워크만으로 **5.003초**, 원장·게이트까지 **5.023초**.
그리고 tasks가 쓰려던 테스트는 `≤ 예산 + slack`을 단언하고 있었다.
**a092가 자기 spec을 반증하는 테스트를 실으려던 것이다.**

4판의 등식은 부등식이고, 남는 몫에 **이름을 준다.**

```text
alertTransportBudget = Attempts × Timeout + (Attempts-1) × RetryDelay
alertOverheadReserve = alertBudget − alertTransportBudget        ≥ 0
alertBudget          = DefaultExitObservationInterval (5s)
```

### 기준 루프

**기준 루프가 exit 관측인 것은 주기가 가장 짧아서가 아니다.** 감독자는 1초 주기다
(`DefaultHealthInterval`, `runtime.go:84`). exit 관측인 이유는 **손절 판정이 거기 살고
§0.3이 거기 걸리기 때문**이다.

감독자가 면제되는 근거는 **하나뿐이다**: `CheckHealth` B5(`runtime.go:383`)의
`takeLatch`가 루프 이름별로 래치하고, 해제는 복구 시(B4 `:375`)뿐이므로 매 초 전송하지
않는다(`…--runtime.takelatch`·`…--runtime.checkhealth` FLM).

**3판이 두 번째 근거로 쓴 "감독자는 이미 기한을 갖고 있다"는 거짓이다.**
`Runtime.escalate`는 `Notify`에 두 번 도달하는데, `r.alert`(`:396`)만
`alertDeliveryBound` 30초를 갖고 `EscalateOperatingMode`(`:415`)는 **평범한 감독자 ctx를
그대로 넘긴다**(`…--runtime.escalate` FLM). 근거는 둘이 아니라 하나다.

### 후보

**측정 기준은 냉연결이다** — 창(engine.log line 6866 이후) 안에서 publish 유발 줄
**39개 중 18개(46%)** 가 `IdleConnTimeout` 90초를 넘겨 떨어져 있다.
**그리고 냉 측정은 1건, 0.754초다.**

| # | Attempts | Timeout | RetryDelay | transport | **reserve** | 냉 여유 |
|---|---|---|---|---|---|---|
| 1 (3판이 택함) | 3 | 1.5s | 250ms | 5.00s | **0ms — 채택 불가** | 1.99× |
| 2 (4판 초안) | 3 | 1.4s | 200ms | 4.60s | 400ms | 1.86× |
| **3 (채택)** | **3** | **1.3s** | **150ms** | **4.20s** | **800ms** | **1.72×** |
| 4 | 2 | 2.0s | 250ms | 4.25s | 750ms | 2.65× |
| 5 | 3 | 1.0s | 500ms | 4.00s | 1000ms | 1.33× |
| 6 | 3 | 0.7s | 200ms | 2.50s | 2500ms | **0.93× — 냉 측정보다 작다** |
| 7 (7라운드 M-4) | 3 | 1.3s | **50ms** | 4.00s | **1000ms** | 1.72× |

**#3을 택한다.**

- **#1은 reserve가 0이라 채택 불가다.** 실측 5.023초로 주기를 넘는다. 3판의 오류다
- **#2는 reserve가 모자란다.** 4판 초안이 이것을 택했는데, 그때의 최악 초과분 71.9ms는
  **유휴 journal**에서 잰 값이었다. 4라운드가 다른 루프의 원장 쓰기를 넣고 재니
  초과분이 **168.3ms**로 커졌다(아래). 400ms는 그 위에 2.4배밖에 안 남는다
- **#6은 채택 불가다.** 유일한 냉 측정이 1회 상한을 넘는다 — 건강한 transport에서도
  매 시도 실패한다
- **#5(1.33×)는 여유가 너무 얇다.** 표본 1건에서 33% 여유를 주장할 수 없다
- **#4는 여유가 가장 크지만 시도를 2회로 줄인다.** `DefaultCriticalAttempts`의 주석
  (`notifier.go:41-44`)은 3의 이유를 "transient (a DNS blip, a restarting ntfy
  container)"라고 적었고 a092에 그 판단을 뒤집을 근거가 없다
- **#3은 냉 여유를 1.86×에서 1.72×로 팔아 reserve를 400ms에서 800ms로 산다.**
  그 거래를 하는 이유는 두 여유가 **지키는 것이 다르기** 때문이다: reserve는
  **spec이 약속한 상한**을 지키고, 냉 여유는 **성공했을지 모르는 전달 하나**를 지킨다.
  전자가 깨지면 계약이 거짓이 되고, 후자가 깨지면 outbox 행이 PENDING으로 남아
  a093이 재시도한다. **하드한 쪽에 여유를 준다.**
- **#7은 그 논리를 한 칸 더 밀면 나오는 값이고, 재지 않고는 답할 수 없었다.**
  아래에서 따로 쓴다.

#### #7을 택하지 않는 이유 — 7라운드 M-4

M-4의 지적은 정당하다. 이 문서는 아래 유도 4번에서 "150ms든 250ms든 재시작 중인
컨테이너는 아직 안 떠 있다"고 **재시도 대기의 무의미함을 스스로 적어 놓고**, 그 항에
300ms를 남겼다. 그렇다면 D를 50ms로 내려 reserve를 1000ms로 만드는 것이  <!-- not-a-measurement: M-4 대안(D=50ms)의 유도값 — 채택하지 않은 후보의 산술이다 -->
**같은 문서의 우선순위("하드한 쪽에 여유를 준다")를 더 잘 따르는 것 아닌가**.

**그래서 쟀다** — `delivery-latency.md` §7.5 셀 E. 결과는 **초과분이 D에 의존하지
않는다**는 것이다: 같은 조건에서 D=50ms 셀이 123.6~218.9 ms, D=150ms 셀(셀 D)이
112.0~209.6 ms로 **같은 띠**다. 당연하다 — D는 transport 항이고 초과분은
비transport 항이다.

**그러므로 #7이 사는 200ms는 막을 대상이 관측되지 않은 200ms다.** 판단은 이렇다.

1. 유도 규칙은 `reserve ≥ 2 × (관측 전 회차 최악)`이고, 전 회차 최악 356.1 ms에서
   **800ms는 이미 2.25배**로 규칙을 만족한다. 8판 재측정의 최악 319.3 ms도 그 아래다.
2. **규칙이 만족되는데 상수를 더 옮기지 않는다.** 이것은 이 change의 **선언된 판정
   규칙**이고(8라운드 M-11이 "어디에도 원칙으로 선언돼 있지 않다"고 짚어 여기 적는다),
   "하드한 쪽에 여유를 준다"보다 **우선한다**. 두 규칙이 충돌할 때 —
   #7이 정확히 그 경우다 — 순서는 이렇다:

   > **(i) 유도 규칙이 만족되는가**를 먼저 본다. 만족되면 상수는 그 자리에 둔다.
   > **(ii) 만족되지 않을 때만** "하드한 쪽에 여유를 준다"가 어느 항을 늘릴지 정한다.

   (ii)가 (i)보다 앞서면 규칙이 값을 고정하지 못한다 — 언제든 "더 하드한 쪽"을
   찾을 수 있기 때문이다. 이 change가 일곱 판 동안 기각된 형태가 정확히 그것,
   **근거 없이 옮긴 값이 문서 여섯 곳의 유도를 무효로 만드는 것**이었다.
3. D를 50ms로 만들면 세 시도가 **100ms 안에** 다 나간다. transient를 사는 것은  <!-- not-a-measurement: M-4 대안(D=50ms)에서 3×50ms로 유도 — 측정이 아니다 -->
   시도 횟수라는 판단은 유지되지만, 대기를 사실상 없애는 쪽은 **보수적 방향이 아니라
   중립**이고, reserve 쪽의 이득은 위 1에 의해 **불필요**하다.

**그래서 #3을 유지하고, #7은 기각이 아니라 대기로 둔다.** 재검토 조건은 명시적이다 —
**초과분이 400 ms를 넘는 것이 한 번이라도 관측되면**(그러면 800ms가 2배 규칙을
깬다) #7이 첫 후보다. 그 관측을 가능하게 하는 계측은 #3의 냉 상한과 같이 **a090**이다.

### `alertRetryDelay = 150ms`의 유도 — 3라운드 M1

3판은 세 상수 중 이것 하나에만 이유를 안 적었다. 유도는 이렇다.

1. 시도 수는 3으로 고정한다(위 #4 논의).
2. reserve는 **관측된 모든 회차의 최악 초과분의 2배 이상**이어야 한다. 전 회차 최악은
   **356.1ms**(5라운드 리뷰어, 아래) → reserve ≥ 712ms.  <!-- not-a-measurement: reserve의 유도 하한 — 356.1ms 실측의 2배로 계산한 값이다 -->
3. 1회 상한은 냉 측정(0.754s) 대비 1.7배 이상을 남긴다 → Timeout ≥ 1.28s.  <!-- not-a-measurement: Timeout의 유도 하한 — 냉 측정 0.754s의 1.7배로 계산한 값이다 -->
4. 둘을 만족하는 **가장 읽기 쉬운 반올림 값**이 Timeout 1.3s · reserve 800ms이고,
   그때 `RetryDelay = (5000 − 3×1300 − 800) / 2 = 150ms`다.

즉 **재시도 대기는 자유 변수가 아니라 나머지 항**이다. 그것이 이 값의 이유다.
그리고 이 크기에서는 재시도 대기가 transient를 실질적으로 돕지 않는다 —
150ms든 250ms든 재시작 중인 컨테이너는 아직 안 떠 있다. transient를 실제로 사는 것은
**시도 횟수**이고, 그보다 긴 회복을 사는 것은 **outbox 행의 생존**(a093)이다.

> **규칙이 두 번 바뀌었고, 두 번째가 5라운드 M1이다.**
>
> - 3·4판은 **비율**("최악의 5배")로 썼다. 분모(실측 최악)가 측정을 세게 할수록
>   커져서(71.9ms → 142.2ms → 168.3ms) 규칙이 값을 고정하지 못했다.
> - 4판 확정은 **절대 마진 400ms**로 바꿨다. 그런데 5라운드가 같은 조건에서
>   **356.1ms**를 얻었고, 재측정은 같은 기계에서 **31.9~77.8ms**를 얻었다.
>   **관측 전체에서 11.2배가 흩어진다**(356.1 ÷ 31.9 — `delivery-latency.md` §7.2.1이
>   이름 붙인 "산포 배수") — 지배 요인은 프로브 조건이 아니라 측정 순간의 주변 부하다.
>   절대 마진도 "어느 회차의 최악이냐"에 종속이었다.
> - 5판은 **전 회차 최악의 2배**로 쓴다. 입력이 한 회차가 아니라 **관측 전체**이므로
>   새 회차가 더 큰 값을 내도 규칙이 무너지는 대신 **재유도를 요구한다**.
>   현재: 전 회차 최악 356.1ms, reserve 800ms = **2.25배**.
>
> 남는 질문은 그대로다 — "이 프로브가 재현하지 못하는 운영 조건(느린 fsync·WAL
> checkpoint)에 얼마를 남기는가". 답을 한 회차의 마진으로 쓰지 않을 뿐이다.
> **배포 기계에서 다시 재는 것이 tasks 7.2의 산출이다.**

### 이것은 판단이지 측정이 아니다 — 그리고 재시도가 그것을 보강해 주지 않는다

여유 1.72배의 근거는 **표본 1건**이다. 그리고 냉 핸드셰이크가 1.3초를 넘어 타임아웃 나면
그 연결은 풀에 남지 않으므로 **다음 시도도 식은 채 시작한다.** 세 번 시도해도 세 번 다
실패한다. 여유는 여유만큼만 믿는다.

#### 로그에는 1.3초를 넘는 간격이 있다 — 그것을 배제한 근거는 측정이 아니다 (5라운드 H3)

`engine.log`의 연속 `exit.position_unmanaged` 줄 간격에는 **1.811초·1.836초**가 있고,
두 건이 연달아 나간 사이클 하나는 **2.499초**다. 전부 1.3초보다 크다.

**이 값들은 왕복이 아니다.** `delivery-latency.md` §5.3대로 그 타입은 조립 자리가 4곳이고
셋이 대사 루프 goroutine이라(`runtime.go:277-283`) 줄 간격이 어느 루프의 체류도 재지
못한다. 1.836초를 만드는 뒤쪽 줄(7814→7815)은 `adoption.go:455` 생산이고 그다음이
`reconcile.clean`이다 — 대사 goroutine이다.

**하지만 배제의 근거가 "재 봤더니 왕복이 아니었다"가 아니라 "생산자를 추적해 보니
다른 루프였다"라는 것을 분명히 적는다.** 두 산출물(`publishBestEffort`·`ntfy.Publish`)이
5판까지 이 숫자를 **"왕복 실측"이라고 부르고 있었고** 어느 문서도 그것을 쓰지도
반박하지도 않았다. 그것이 5라운드 H3다.

그래서 유도가 서 있는 자리는 이렇다:

| | 값 | 성격 |
|---|---|---|
| 유효 왕복 표본 (§3) | 0.1983 ~ **0.7540 s**, n=6 (냉 1건) | 짝지은 줄이 모두 인접함을 확인 |
| 무효 간격 표본 | 0.202 ~ **1.836 s**, n=9 (+2.499 s) | **생산자가 섞여 있다 — 왕복이 아니다** |
| 채택 1회 상한 | 1.3 s | 유효 최댓값의 1.72배 |

**만약 무효 표본이 왕복이었다면 1.3초는 작다.** 이 change는 그 가능성을 측정으로
닫지 않았다. 닫는 것은 두 가지다 — design D2의 재검토 조건(냉 publish가 1.3초를 넘는
것이 **한 번이라도** 관측되면 재유도)과 그 관측을 가능하게 하는 **a090의 계측**.
a090 이전까지 이 상한은 **판단이며, 그 판단이 틀렸을 때 무엇이 그것을 알려주는지가
정해져 있다**는 것이 이 절이 주장하는 전부다.

**그래서 이 숫자에는 재검토 조건을 붙인다**: 냉연결 publish가 1.3초를 넘는 것이 한 번이라도
관측되면 이 값을 다시 유도한다. 그 관측을 가능하게 하는 계측은 a090이다.

### reserve 800ms는 측정에서 나왔다

`analysis/delivery-latency.md` §7의 국소 프로브(저장소 무편집, `go test -overlay`).

**초과분 측정이 두 번 커졌다는 것을 먼저 쓴다.**

| 측정 회차 | 조건 | 비전송 초과분 최악 |
|---|---|---|
| 4판 초안 (10회) | `-race` · `GOMAXPROCS=2` · CPU 3배 초과구독, **유휴 journal** | 71.9 ms |
| 4라운드 리뷰 | + 다른 루프의 연속 원장 쓰기 | 142.2 ms |
| 4판 확정 | 위 전부 + `EnqueueAlert` 연속 경합 | 168.3 ms |
| **5라운드 리뷰** | 같은 조건 | **356.1 ms** ← **전 회차 최악** |
| 5판 재측정 A | `-race` 없음 · writer 0/1/2/4 | 36.9 ~ 54.9 ms |
| 5판 재측정 B | `-race` · `GOMAXPROCS=60`(코어 20) · writer 0/1/2/4 | 31.9 ~ 77.8 ms |
| 8판 재측정 — **로거 nil** | `-race` · `GOMAXPROCS=60` · writer 0/2 · 셀당 5회 | 79.9 ~ **319.3 ms** |
| 8판 재측정 — **로거 실제** | 같은 조건 | 75.7 ~ 209.6 ms |

**조건을 세게 하면 커진다는 것이 4판의 진단이었는데, 5판 재측정 B는 같은 조건에서
77.8ms를 얻었다.** 조건이 아니라 **주변 부하**가 지배한다 — `delivery-latency.md`
§7.2.1. 그래서 규칙의 입력은 한 회차가 아니라 **관측 전체의 최댓값**이다.

**커진 이유는 journal이 `SetMaxOpenConns(1)`이기 때문이다**
(`internal/journal/journal.go:151`, `synchronous=FULL` · `_txlock=immediate`).
`Notify` 한 번이 그 **한 개 커넥션**에서 트랜잭션 여러 개를 돌리므로 다른 루프의 쓰기가
전부 줄을 선다. **4판 초안의 프로브는 아무도 안 쓰는 새 임시 DB를 써서 이 항을 통째로 뺐다.**

확정 상수(#3)의 실측:

| | 값 |
|---|---|
| transport 예산 | 4.200 s |
| 실측 체류 범위 (이 회차) | 4.2903 ~ **4.3683 s** |
| 비전송 초과분 (이 회차) | 90.3 ~ 168.3 ms |
| **비전송 초과분 (전 회차 관측 범위)** | **31.9 ~ 356.1 ms** |
| **reserve ÷ 전 회차 최악** | **800 / 356.1 = 2.25배** ← 유도 규칙의 판정값 |
| 주기(5s)까지 남은 여유 (이 회차) | 631.7 ~ 709.7 ms |

**"이 회차"와 "전 회차"를 구별해서 읽는다.** 유도가 서는 것은 아래 줄이고,
위 줄은 한 번의 관측이다 — 4판이 631.7ms를 설계의 성질처럼 쓴 것이 5라운드 M1이다.

`-race` 자체는 체류를 의미 있게 늘리지 않는다. 늘리는 것은 **원장 경합**이다.

**한계 — 프로브가 재현하지 못하는 것을 먼저 적는다.**

1. **로컬 SSD의 fsync다.** 운영 디스크가 느리면 `synchronous=FULL` 커밋이 길어진다.
2. **WAL checkpoint를 강제하지 않았다.** 운영 WAL은 3.9MB이고, `Notify`가 도는
   트랜잭션 중 하나 안에서 checkpoint가 일어날 수 있다.
3. **`n.mu` 경합은 이 표에 없다.** 그것은 초과분이 아니라 **배수**이고 아래에서 따로 쓴다.

800ms 몫은 1·2를 위한 것이지 남는 시간이 아니다.

**4번째 한계는 8판이 지웠다 — 7라운드 H2.** 7판까지의 프로브는 `newNotifier`의
4번째 인자(`log *obs.Logger`, `exitwiring.go:71-72`)에 `nil`을 넘겼고, `Notifier`의
로그 쓰기가 전부 `n.Log != nil` 뒤에 있으므로 **소진 경로의 구조화 로그 세 줄을
구조적으로 제거한 상태**에서 쟀다. 위 D3 주석의 열거에도 그 항이 없었다.

로거를 유일한 변수로 두고 다시 쟀다(`delivery-latency.md` §7.5).
**로거 nil 셀의 최악 초과분 319.3 ms**가 **로거 실제 셀의 최악 초과분 218.9 ms**보다
크게 나왔다 — 부호가 반대다.

**그 역전을 "로그 쓰기가 측정 한계 아래"의 증거로는 쓰지 않는다**(8라운드 H4).
셀당 n=5의 max 비교이고, §7.2.1이 이름 붙인 11.2배 산포가 지배하는 분포에서
그것이 말하는 것은 **"이 설계로는 안 잡힌다"**이지 "효과가 없다"가 아니다.
세 줄의 크기는 **미측정으로 남는다.**

**그래도 800ms는 안 바뀐다** — 하중을 지는 것은 하나다: 이번 회차 최악 319.3 ms가
전 회차 최악 **356.1 ms**보다 작으므로 **유도 규칙의 입력이 그대로**다.

**바뀐 것은 열거다.** 크기를 몰라도 열거는 완전해야 한다 — 열거에서 빠진 항은
**다음 편집이 예산 밖으로 읽는 항**이기 때문에 D3 주석과 두 delta의 SHALL 열거에
로그 쓰기를 더했다. 이 change가 반복해서 배운 것이 "빠진 것은 없다고 쓰지 말고
세어서 쓰라"였다.

### `n.mu` 경합은 이 상한 밖이고, 그것을 spec이 말하게 한다

`deliver`는 `n.mu`를 **재시도 전체(대기 포함)** 보유하고(`notifier.go:241-242`)
`*obs.Notifier`는 다섯 루프가 공유하는 **하나**다(`gateway.go:280`).
두 루프가 동시에 critical을 올리면 뒤쪽의 `Notify` 체류에 앞쪽 체류가 **통째로** 더해진다.

**채택 상수(T=1.3초·D=150ms, transport 4.200초)에서 실측 — 앞쪽 4.234초, 뒤쪽 8.458초.**
관측 주기의 **1.69배**, 자기 차례의 **2.00배**다.

> **4판이 여기 적은 9.231초는 후보 #2(T=1.4초·D=200ms, transport 4.600초)에서 잰 값이었다.** <!-- rejected-value -->
> 5라운드 H1이 그것을 짚었고, 대조군을 돌려 확정했다 — 후보 #2로 재현하면
> 앞쪽 4.6531~4.6631초·뒤쪽 9.1539~9.1637초가 나온다. 4판 §7.4가 실은
> `loopA dwell = 4.652 s`가 그 앞쪽 값이다. **채택하지 않은 상수에서 잰 수를
> spec의 SHALL 근거로 실었던 것**이고, 같은 문서 §7.2의 4.2903~4.3683초와
> 양립할 수 없다는 것이 문서 안에서 이미 보였어야 했다.

a092는 이것을 **고치지 않는다**(전송이 루프 밖으로 나가야 하고 그것은 a093이다).
고치지 않는 대신 **spec이 그것을 말하게 한다** — engine-safety 델타는 상한을
"자기 차례"로 한정하고, 직렬화가 그 상한의 배수를 만든다는 것을 실측값과 함께 적는다.

> **4판 초안은 exit-policy ¶20에서 이 사실을 인정해 놓고 같은 요구의 ¶16에서
> 반대로 SHALL NOT을 걸었다.** engine-safety 델타에는 `n.mu`가 한 글자도 없었다.
> 2라운드 C3가 이름 붙인 것이 네 판째 같은 자리에서 재발한 것이고, 그것이 4라운드 C1이다.

### 정상 등급

`publishBestEffort`는 publish 1회다(AST branches 2, 루프 없음). 10s → **1.3s**.
실패가 로그 한 줄이므로 축소의 대가가 더 작다.

### 사이클 총합은 이 예산이 정하지 않는다

`alertProposalRefused`의 AST가 **branches 0 · returns 0**이다
(`…--exitobserver.alertproposalrefused` FLM) — 억제가 없다는 것이 열거로 확정됐다.
그러므로 한 사이클이 알림을 여러 번 올린다. 사이클 최악 = **알림당 상한 × 그 사이클의
알림 수**이고, **알림당 상한은 transport가 아니라 `alertBudget`(5.0초)이다** —
비전송 작업이 0이 아니기 때문이고, 그것이 이 절 첫머리에서 3판을 기각한 이유 그대로다.

> **8판은 여기서 `4.2s × N`이라고 썼다**(8라운드 H2). 같은 절이 `D2` 첫머리에서
> "예산이 걸리는 것은 전송이 아니라 **호출**이다 — 3판이 여기서 틀렸다"고 적어 놓고,
> 절 끝에서 그 치환을 저질렀다. 실측 체류는 언제나 `4.2s`보다 크므로
> `4.2s × N`은 **알림당 초과분만큼 낙관적**이고, `exit-policy` 델타의 시나리오는
> 옳게 "각 알림은 **관측 주기** 안에 … 그 3배까지"라고 쓴다. **유도 정본이 spec보다
> 낙관적이면 유도 정본이 틀린 것이다.**
>
> **9판은 여기에 "알림당 최소 90 ms"라고 적었고 그것은 이 change 자신의 코퍼스가
> 반증한다**(9라운드 H-4). 90ms는 **한 회차**(`-race`·원장 경합)의 초과분 하한이고,
> 이 문서는 75줄 뒤에서 그 오류에 이미 이름을 붙여 놓았다 —
> *"4판이 631.7ms를 설계의 성질처럼 쓴 것이 5라운드 M1이다."* 채택 상수에서 관측된
> 초과분의 하한은 **31.9 ms**이고(§7.2.1 재측정 B, `check_values.py`의 명명값
> "초과분 관측 범위의 하한"), §7.4 loopA의 최솟값은 `4.234387s − 4.200s = 34.4ms`다. 그러므로 **수량을 쓰지
> 않는다** — 결론(`4.2s × N`은 하한이고 알림당 상한은 `alertBudget`이다)은 초과분이
> 양수라는 것만으로 성립하고, 그 이상은 회차에 종속이다.

그리고 `deliver`가 `n.mu`를 쥐므로(`notifier.go:241-242`) 다른 루프의 전송을 기다린
시간이 더 붙는다. **a092는 뮤텍스를 쥐고 있는 시간을 34초에서 4.2초로 줄일 뿐이다.**

## D3 — 상수는 한 자리에 모으고 유도를 함께 적는다

세 값이 두 파일에 흩어지면 합을 읽으려면 두 파일을 읽어야 한다. 그것이 34초를 만든
조건이다. 셋을 `internal/app/engine/notifications.go`에 모으고 `exitwiring.go`는
같은 패키지이므로 그대로 참조한다.

```go
// notifications.go — import "time" 을 추가해야 한다 (현재 os·strings·config·obs)

// alertBudget is the wall clock one alert may hold the exit observation cycle.
//
// It is DefaultExitObservationInterval and not a number of its own: the exit
// observation loop sleeps its interval *after* the cycle (exitloop.go:358-359),
// so a cycle's alert time is added to the observation gap rather than absorbed
// by it.
//
// It is not the shortest loop period in the engine — the health supervisor runs
// at DefaultHealthInterval, one second. It is the loop where the stop lives.
const alertBudget = DefaultExitObservationInterval

// alertPublishAttempts is obs.DefaultCriticalAttempts restated, not changed. It
// is written here so all three terms of the budget are read in one place.
const alertPublishAttempts = obs.DefaultCriticalAttempts // 3

// alertPublishTimeout bounds one publish. 1.3s is a judgement on one cold-pool
// measurement of 0.754s, not a distribution: see the change's delivery-latency
// analysis. Revisit if a cold publish is ever observed above it.
const alertPublishTimeout = 1300 * time.Millisecond

// alertRetryDelay is what is left once the timeout and the reserve are chosen,
// not a free variable: (alertBudget - 3*1300ms - 800ms) / 2.
const alertRetryDelay = 150 * time.Millisecond

// alertTransportBudget is what the three fields above cost on the network.
const alertTransportBudget = alertPublishAttempts*alertPublishTimeout +
	(alertPublishAttempts-1)*alertRetryDelay

// alertOverheadReserve is what alertBudget leaves for the work Notify does
// around the sending. The list is meant to be exhaustive for the exhaustion
// path, because a term left out of it is a term the next edit reads as free:
// the outbox insert (notifier.go:177), one MarkAlertAttemptFailed per attempt,
// the entry-gate latch (notifier.go:284), the EscalateOperatingMode transaction
// that exhaustion reaches through escalate (called at notifier.go:187, the
// journal call itself at notifier.go:217), and three structured log writes --
// logEvent's Warn (notifier.go:131), deliver's Error on exhaustion
// (notifier.go:279), and escalate's Warn when the mode actually changed
// (notifier.go:228).
//
// Everything but the latch and the log queues on one journal connection;
// EntryGate.Block takes a mutex and writes a map (execgw/retry.go:498).
//
// Measured worst across every probe session is 356ms; the same nominal
// conditions have also produced 32ms, so this number tracks ambient load more
// than it tracks the code. 800ms is therefore not "worst plus a margin" but
// twice the largest value ever observed. Re-measure before trusting it on a
// machine other than the one in the change's delivery-latency analysis.
const alertOverheadReserve = alertBudget - alertTransportBudget

// The four arrays are the assertion. Each subtracts one so that the *smallest
// legal value* is the one that still compiles, and each names its own constant
// in the failure.
//
// Zero is not "unset" at the callee, it is a different, larger number: attempts
// <= 0 becomes DefaultCriticalAttempts (notifier.go:245), delay <= 0 becomes
// DefaultRetryDelay, 2s (notifier.go:292), and timeout <= 0 becomes 10s
// (ntfy.go:96). A zero written here would compile into a green build whose
// reserve is positive and whose real transport is 7.9s. The first three arrays  <!-- not-a-measurement: alertRetryDelay=0 반증의 유도값 — 3×1.3s+2×2s의 결과다 -->
// are what makes that not compile.
//
// The fourth is the budget itself, and it is strict: both spec deltas say the
// transport budget SHALL NOT equal the observation interval, because the work
// enumerated above is not zero.
var _ [alertPublishAttempts - 1]struct{}
var _ [alertPublishTimeout - 1]struct{}
var _ [alertRetryDelay - 1]struct{}
var _ [alertOverheadReserve - 1]struct{}

// The four above count nanoseconds, so they cannot tell 1300ms from 1300ns. A
// dropped unit -- alertPublishTimeout = 1300 -- passes all four and produces a
// green build whose per-publish timeout is 1300ns, which times out before any
// dial completes: every alert exhausts, every alert latches the entry gate and
// escalates the operating mode. That is worse than the zero these four catch,
// and until now nothing caught it. The last two assert the order of magnitude,
// which is the part a unit carries.
var _ [alertPublishTimeout/time.Millisecond - 100]struct{}
var _ [alertRetryDelay/time.Millisecond - 10]struct{}
```

**단언이 왜 넷이고 왜 전부 `- 1`인가 — 7라운드 B1·H1.**

7판까지의 단언은 `var _ [alertOverheadReserve]struct{}` 하나였다. **`[0]struct{}`는
합법적인 Go 배열이므로 그것이 강제한 것은 `transport ≤ budget`뿐이다.** 두 spec delta는
`SHALL NOT`으로 **`<`** 를 요구한다(engine-safety:26 · exit-policy:16). 3판이 기각당한
후보 #1(T=1.5s·D=250ms, transport = 주기)을 그 단언 아래 넣으면 **BUILD OK**가 난다 —
**기각의 이유였던 바로 그 구성이 컴파일을 통과한다.** `- 1`이 그 한 칸을 닫는다.

그리고 단언은 **합**만 보고 **항의 부호**를 안 봤다. 세 필드 전부 피호출자에 0 폴백이
있으므로(위 주석) `alertRetryDelay = 0`은 컴파일 reserve **+1.1초**와 실제 transport  <!-- not-a-measurement: 같은 반증의 유도값 — 기본 대체값 2s가 만드는 차이다 -->
**7.9초**를 동시에 만든다. 그것은 engine-safety:24가 `SHALL NOT`으로 금지한  <!-- not-a-measurement: 같은 반증의 유도값 — 3×1.3s+2×2s의 결과다 -->
"그 자리에서 비워 두어 피호출자의 기본값이 쓰이는 구성"이다. **앞의 세 배열이 그 세
항을 각각 본다.**

**실측 — 아래 표의 모든 구성을 `go build -overlay`로 이 패키지에 넣어 돌렸다**
(저장소 무편집, go1.26.5). 메시지는 잘라내지 않았다.

> **구성 수를 산문에 적지 않는다.** 9판까지 여섯 곳이 그 수를 손으로 박았는데
> 같은 절의 표는 아홉 행이었다(9라운드 H-2) — 8라운드 H5가 한 행을 타입별 둘로 가르면서
> 손으로 박은 카운트가 낡았다. *"기대값을 박은 검사는 검사가 아니라 또 하나의 주장이다"*
> 가 이 파일의 원칙이고, 그 원칙은 검사 코드만이 아니라 산문에도 적용된다.
> 이제 `tools/check_values.py`의 `check_derived_counts`가 이 표의 행 수를 세어
> 산문의 "N 구성" 주장과 대조한다 — 그래서 이 문장은 수를 말하지 않는다.

| 구성 | 결과 | 컴파일러 메시지 |
|---|---|---|
| 채택 #3 (T=1.3s·D=150ms) | **BUILD OK** | — |
| reserve = 1ns (허용되는 최솟값) | **BUILD OK** | — |
| M-4 대안 (D=50ms, reserve 1000ms) | **BUILD OK** | — |
| **후보 #1 (T=1.5s·D=250ms, reserve 0)** | **FAIL** | `invalid array length alertOverheadReserve - 1 (constant -1 of int64 type "time".Duration)` |
| T=1.6s·D=200ms (reserve −200ms) | FAIL | `invalid array length alertOverheadReserve - 1 (constant -200000001 of int64 type "time".Duration)` |
| `alertRetryDelay = 0 * time.Millisecond` | **FAIL** | `invalid array length alertRetryDelay - 1 (constant -1 of int64 type "time".Duration)` · `invalid array length alertRetryDelay / time.Millisecond - 10 (constant -10 of int64 type "time".Duration)` |
| `alertRetryDelay = 0` (맨 정수) | **FAIL** | `invalid array length alertRetryDelay - 1 (untyped int constant -1)` · `invalid array length alertRetryDelay / time.Millisecond - 10 (constant -10 of int64 type "time".Duration)` |
| `alertPublishTimeout = 0 * time.Millisecond` | **FAIL** | `invalid array length alertPublishTimeout - 1 (constant -1 of int64 type "time".Duration)` · `invalid array length alertPublishTimeout / time.Millisecond - 100 (constant -100 of int64 type "time".Duration)` |
| `alertPublishAttempts = 0` | **FAIL** | `invalid array length alertPublishAttempts - 1 (untyped int constant -1)` |
| **`alertPublishTimeout = 1300` (단위 누락)** | **FAIL** | `invalid array length alertPublishTimeout / time.Millisecond - 100 (constant -100 of int64 type "time".Duration)` |
| **`alertRetryDelay = 150` (단위 누락)** | **FAIL** | `invalid array length alertRetryDelay / time.Millisecond - 10 (constant -10 of int64 type "time".Duration)` |
| **`alertPublishTimeout = 1300 * time.Microsecond` (단위 오타)** | **FAIL** | `invalid array length alertPublishTimeout / time.Millisecond - 100 (constant -99 of int64 type "time".Duration)` |
| **`alertRetryDelay = 150 * time.Microsecond` (단위 오타)** | **FAIL** | `invalid array length alertRetryDelay / time.Millisecond - 10 (constant -10 of int64 type "time".Duration)` |

> **굵은 넷이 10판이 더한 것이다** — 9라운드 H-1. 9판의 단언 넷은 **나노초만 세므로
> 1300ms와 1300ns를 구별하지 못했고**, 단위를 빠뜨린 구성이 초록 빌드로 통과했다.
> 그 빌드에서는 모든 publish가 다이얼 전에 기한을 넘겨 **매 알림이 소진되고 매 알림이
> 진입 게이트를 래치한다** — 9판의 논증이 막겠다던 `0`보다 나쁘다. 마지막 두 배열이
> 단위가 지고 있던 부분, 즉 **자릿수**를 본다.

**메시지가 상수 이름을 말한다** — 그것이 이 change가 테스트 대신 배열을 쓰는 이유다
(아래). **어느 행도 잘라내지 않았다.**

**타입 부분은 상수를 *어떻게 쓰느냐*에 달렸고, 그것이 표에 두 행으로 있다**
(8라운드 H5). `0 * time.Millisecond`로 쓰면 `time.Duration`이라
따옴표 붙은 `"time".Duration`이 나오고, **맨 `0`으로 쓰면 무타입 정수**라
`untyped int constant -1`이 나온다. `alertPublishAttempts`는
`obs.DefaultCriticalAttempts`(무타입 상수)라 늘 후자다. 8판은 행 라벨을
`alertRetryDelay = 0`으로 적고 메시지는 typed 형태를 실어 **둘이 어긋났다.**
**어느 쪽으로 쓰든 FAIL이므로 방어는 그대로 성립한다** — 어긋난 것은 문자열뿐이고,
5라운드 H1이 문자열에도 측정 문맥을 요구한 것이 이 자리다.

**배열 길이는 어느 구성에서도 int 범위 안이다.** 가장 큰 것이
`[alertPublishTimeout - 1]` = 1,299,999,999이고, reserve > 0을 만족하는 어떤 구성에서도
1.3e9을 못 넘는다(transport < budget = 5e9ns이고 T는 그 1/3 아래다). **32비트 타깃에서도
int32 범위 안**이므로 GOARCH 의존이 없다 — 원소가 `struct{}`라 배열 자체의 크기는 0이다.

**등호를 강제하지는 않는다.** 필요한 것은 `transport < budget`이지
`transport == budget`이 아니다. 3판의 양방향 배열은 등호를 강제했고, 등호가 바로 이
change를 거짓으로 만든 조건이었다(D2). 7판은 그것을 고치려다 **반대쪽으로 한 칸
넘어가** 등호를 허용했다. `- 1` 하나가 두 오류 사이의 정확한 자리다.

**그리고 이 등식을 테스트로 지킬 수는 없다.** 단언과 테스트가 같은 패키지에 있으므로
등식이 깨지면 **테스트 바이너리가 만들어지지 않는다** — 테스트는 영영 실행되지 않는다.
3판은 "컴파일 실패는 읽기 어렵고 테스트는 이유를 말한다"고 썼는데, 실행되지 않는 테스트는
아무 이유도 말하지 않는다. 위 표의 메시지가 그 자리를 대신한다.

> **어디서 쟀는지를 같이 적는다** — 5라운드 H1이 준 규칙이 상수뿐 아니라 문자열에도
> 걸린다. 위 표의 메시지는 **이 저장소 안 `internal/app/engine`에 `go build -overlay`로
> 상수 파일을 넣어** 얻은 것이고 `go build`·`go vet`·`go test`가 모두 같은 문자열을 낸다.
> **따옴표는 여기서 나온다**: 패키지가 `time`을 import한 문맥이라 Go가 타입을
> `"time".Duration`으로 적는다. 같은 코드를 `time`을 import하는 **독립 모듈**에서 돌리면
> 따옴표 없이 `time.Duration`이 나온다. 6라운드 적대적 보이스가 그 독립 모듈에서 재고
> "재현되지 않는다"고 보고했으나, **실제 시나리오에서는 위가 맞다** — 그 지적은 오탐으로
> 기각했고, 남는 교훈은 **문자열도 측정 문맥을 밝혀야 한다**는 것이다.
>
> **7판은 여기에 `-500000000`을 실었다**(7라운드 M-2). 그 값을 만드는 구성은
> transport 5.5초(예: T=1.7s·D=200ms)인데 **이 문서의 어느 시나리오도 그것이 아니다.**  <!-- not-a-measurement: 반증 예시(T=1.7s·D=200ms)의 유도값 — 채택 구성이 아니다 -->
> 형태만 맞고 값은 아무 데서도 안 나온 수였다 — 그래서 지웠고, 그 자리에 **실제로 돌린
> 구성들의 표**를 놓았다. 6라운드 M-2가 "10배·5배"에서 잡은 것과 같은 결함이
> 문자열에서 되풀이된 것이다.

테스트가 할 수 있는 일은 **값을 고정하는 것**이다(tasks 6.5). 단언이 잡는 것과 못 잡는 것이
갈린다 — 누가 1.3s를 3s로 바꾸면 transport가 `3 × 3s + 2 × 150ms = 9.3s`가 되어 예산을
넘으므로 **reserve가 음수가 되고 컴파일이 깨진다**. 그러나 1.3s를 1.0s로 바꾸면
transport가 `3 × 1.0s + 2 × 150ms = 3.3s`라 reserve는 양수이고 **컴파일을 통과한다** —
합법적이지만 의도하지 않은 재배분이다. 값 고정 테스트가 잡는 것은 후자뿐이다.
등식을 증명하는 것이 아니라 **선택을 고정**한다.

## D4 — 잃는 것을 계산한다 (완전성을 주장하지 않는다)

예산 축소가 전달에 미치는 영향. **아래가 전부라고 주장하지 않는다** — 2판은 "하나뿐이다"라고
썼고 그것이 틀렸다.

| publish의 실제 거동 | 오늘(10s/시도) | a092(1.3s/시도) | 차이 |
|---|---|---|---|
| 빠른 실패 (DNS NXDOMAIN·connection refused·non-2xx) | 3회 실패 | 3회 실패 | 없음 |
| 타임아웃 (블랙홀) | 30s 쓰고 실패 | 4.2s 쓰고 실패 | 없음 (시간만 줄어듦) |
| (1.3s, 10s] 구간에서 **성공** | 성공 | **실패** | **손실 ①** |
| 서버는 **접수**했는데 응답이 1.3s 안에 안 옴 | 대개 성공 | **실패로 기록 + 재시도** | **손실 ②** |

### 손실 ② — 2판이 빠뜨린 모드

`client.Do`(`ntfy.go:125`)가 기한 오류를 돌려주면 `deliver`는 `MarkAlertAttemptFailed`를
쓰고 재시도한다. 메시지는 이미 나갔으므로:

- 운영자는 **같은 알림을 2~3건 중복 수신**한다
- outbox 행은 PENDING으로 남는다
- `Gate.Block(ReasonAlertUndelivered)`가 래치된다 — **실제로는 배달된 알림에 대해**
- `Flush` 프로덕션 호출자가 없으므로 **아무도 정정하지 않는다**

1회 상한을 10초에서 1.3초로 줄이면 이 모드는 **더 있음직해진다.**

`ntfy.go:99`의 `context.WithTimeout`은 **지역 ctx를 만든다** — 그래서 짧아진 상한이
`MarkAlertAttemptFailed`·`MarkAlertDelivered`(호출자 ctx)에는 걸리지 않는다.
원장 기록은 상한과 무관하게 완료된다.

**그래도 §0.3·§0.6 위반은 아니다**: 결과는 진입 차단이고 청산은 건드리지 않는다.
보수 방향이다. 다만 **대가는 인정하고 spec에 적는다**(engine-safety 델타의
"받았는데 실패로 기록된다" 시나리오).

### 손실 ①이 실현될 확률

관측 6건 중 (1.3s, 10s] 구간에 든 것은 **0건**, 최댓값 0.754초. 다만 **표본이 6건이고
냉은 1건**이다.

**그래도 진행하는 이유**: 손실은 관측되지 않은 구간의 확률이고, 34초의 비용은
**transport가 죽는 즉시** 매 사이클 결정적으로 발생하며 §0.3이 직접 걸린다.

## D5 — 균일성의 대가를 경로별로

| 경로 | 오늘 | a092 후 | 평가 |
|---|---|---|---|
| P4 exit 관측 | 34s | 4.2s | **목적** |
| P2·P5·P6 대사 루프 (60s 주기) | 34s | 4.2s | 이득. 주기 대비 여유가 크므로 손실 ①의 노출도 낮다 |
| P3 감독자 (`Runtime.alert`) | min(30s, 34s) = 30s | **4.2s** | **대가.** 엔진이 왜 죽었는지 설명하는 알림이고 프로세스는 종료 중이며 `Flush` 호출자가 없다. 느리지만 살아 있는 transport라면 6초에 배달됐을 것이 이제 안 된다 |
| P1b 감독자 승격 (`runtime.go:415`) | 34s, **기한 없음** | 4.2s, 기한 없음 | 기한이 없다는 사실은 a092가 고치지 않는다. 예산이 줄어 노출이 줄 뿐이다 |
| P1 Announcer | 34s | 4.2s | 오늘 발동하지 않음(`analysis/notify-reach.md`) |
| P7 flatten | — | — | 프로덕션 `Notifier` nil |

P3의 대가를 받아들이는 근거: 종료 알림은 **durable하게 기록되고** 종료 자체를 막지
않는다. 그리고 `alertDeliveryBound`의 30초가 이제 4.2초보다 커서 **주석이 처음으로 참이 된다** —
오늘은 34초 상한이 30초 기한보다 커서 그 주석이 거짓이다. **주석 자체의 수정은 a093이
한다**(a092는 `runtime.go`를 편집하지 않는다).

## D6 — CLI 시험 발송은 10초로 남긴다. 그 대가도 적는다

`cmd/tossctl/notificationsettings.go:151`은 별도 `&obs.Ntfy{}`를 만들고 엔진 루프가 아니다.
1.3초를 씌우면 운영자가 손으로 한 번 보내는 시험이 차가운 연결에서 헛되이 실패한다
(재시도가 없다 — `Publish` 1회).

**3판은 여기서 자기 SHALL을 위반했다.** engine-safety 델타의 조립 SHALL이 범위를
한정하지 않아 이 자리가 위반이 됐고, tasks 6.9는 그 위반을 소스 스캔 테스트로
**고정**하려 했다. 4판은 SHALL의 범위를 **엔진 루프가 동기로 기다리는 경로**로 한정한다 —
사람이 명령을 입력해 기다리는 대화형 시험 발송은 루프를 붙잡지 않는다.

**대가**: 운영자의 유일한 채널 점검이 **프로덕션 예산을 시험하지 않게 된다.**
10초 발송이 "채널 정상"이라고 보고하는데 프로덕션은 1.3초에 세 번 실패하고 게이트를 걸 수 있다.

**그래도 10초를 유지한다** — 시험의 목적은 "토픽·토큰·URL이 맞는가"이고 예산 검증이 아니다.
예산 검증을 채널 점검에 얹으면 운영자가 설정 오류와 지연을 구분할 수 없다.
**이 결정을 소스 스캔 테스트로 고정한다**(tasks 6.9) — 이제 그것은 위반의 고정이 아니라
**범위 밖이라는 결정의 고정**이다.

## D7 — Pre-Edit 선언 (High-risk)

| 항목 | 내용 |
|---|---|
| 편집 대상 | `newNotifier`(`exitwiring.go:71-81`), `resolveNotificationPublisher`(`notifications.go`의 `&obs.Ntfy{...}` 리터럴 한 줄) |
| 편집 성격 | **구조체 리터럴에 필드 추가 + 상수·import 선언.** 조건문·이탈·호출을 더하지 않는다 |
| 무변화 증명 | 편집 후 AST가 `newNotifier` branches 0/returns 1/calls 0, `resolveNotificationPublisher` branches 5/returns 4 유지 |
| 되돌리기 | 세 필드를 지우면 정확히 오늘로 돌아온다. 스키마·설정·원장 변화 없음 |
| 안 건드리는 것 | `obs` 전부, `exitloop.go` 전부, `runtime.go` 전부, `journal` 전부. **`cmd/tossctl`은 프로덕션 코드를 안 건드린다 — 다만 6.9가 테스트 파일 하나를 신설한다**(4라운드 N5: 3판 문구가 그것과 모순됐다) |
| 토글 | **없다.** 없애려는 것이 upstream 동작(34초)이므로 OFF 상태가 곧 결함이 된다. §0.6("명확한 근거가 있는 보수 방향")으로 정당화하고 근거는 `analysis/delivery-latency.md` |

> 줄 번호를 상수·import 삽입 뒤의 값으로 쓰지 않는다. `notifications.go`에 `import "time"`과
> 상수 블록이 들어가면 `&obs.Ntfy{...}` 리터럴의 줄 번호가 내려간다. tasks는 줄 번호가
> 아니라 **식별자**로 자리를 지시한다.

## 검증

- `newNotifier`가 `Attempts`·`RetryDelay`를 상수 값으로 채운다
- `resolveNotificationPublisher`가 돌려준 `obs.Publisher`를 `*obs.Ntfy`로 단언했을 때
  `Timeout`이 상수 값이다
- **예산 부등식은 컴파일 타임이 지킨다.** 테스트는 등식을 증명하지 않고 **값을 고정**한다 —
  같은 패키지의 컴파일 단언이 깨지면 테스트는 실행되지 못한다
- `resolveNotificationPublisher`의 세 nil 이탈(**B2 `:77`·B3 `:83`·B5 `:94`**) **무변화** — B1 `:69`는 `getenv == nil`이고 반환하지 않는다(4라운드 M1)
- CLI 시험 발송 `Ntfy`는 `Timeout` **무설정** 유지 (소스 스캔, `go/parser`)
- **알림 하나**의 실시계 체류가 `alertTransportBudget` 이상 `alertBudget` 미만 —
  실제 `*obs.Ntfy`(`BaseURL`·`Topic` 채움) + 응답하지 않는 리스너, 실시계.
  **가짜 시계로는 측정할 수 없다**(`ntfy.go:99`의 `context.WithTimeout`은 실시계를 쓰고
  `obs.Ntfy`에 `Clock` 필드가 없다 — 3라운드에서 12초 워치독을 넘겨 막히는 것을 확인했다)
- `obs` 패키지 테스트 회귀 0 — 이 change는 그 패키지를 편집하지 않는다
- **회귀와 경합은 두 갈래로 본다** — `make test`(= `go test -timeout 30m ./...`)로 나무 전체
  회귀 0, `-race`는 **폭발 반경 5개 패키지에서만**. 나무 전체 `-race`는 이 기계에서
  **완주하지 못한다**(`internal/journal`이 30분·60분 기한을 둘 다 넘긴다).
  명령과 근거는 tasks 9.4가 갖는다. **완주하지 못하는 명령을 검증 항목으로 적지 않는다** —
  5라운드 B1의 정정이 tasks에만 착지하고 이 목록에 안 온 것이 6라운드 H-1이다
