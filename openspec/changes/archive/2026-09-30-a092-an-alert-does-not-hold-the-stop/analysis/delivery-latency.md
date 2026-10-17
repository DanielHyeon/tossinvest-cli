# 알림 동기 체류 실측 (11판 본문 + 14판 §9)

> ## ⚠ §7의 "확정 상수"는 **폐기된 구성**을 잰 것이다
>
> **§7.2 이하의 모든 측정은 시도 3회 · 1회 상한 1.3s · 대기 150ms 구성에서 잰 것이다.**
> 그 구성은 **13판에서, 그리고 다시 14판에서 기각됐다.** 현재 채택값은
> **시도 1회 · 1회 상한 3.5s · reserve 1500ms**이며, 그 근거 측정은 **§9**에 있다.
>
> 14라운드에서 두 목소리가 독립으로 이것을 찾았다(보이스 A P5 · 보이스 B R14-P6):
> *"측정 정본이라 이름 붙은 문서가 채택되지 않은 구성을 재고 있다."* 맞다.
>
> **§7의 숫자를 고쳐 적지 않는다.** 그 값들은 그 구성에서 **실제로 잰 것**이고,
> 새 구성의 수로 갈아 끼우면 그것은 재측정이 아니라 **날조**다.
> §7이 여전히 정본인 것은 **비전송 초과분**(outbox·게이트·승격·로그)뿐이다 —
> 그 항은 시도 횟수와 1회 상한에 의존하지 않는다(§7.5 셀 E가 그것을 보인다).
> **transport에 딸린 수**(4.200s · 800ms · 8.458s · 9.231s)는 전부 폐기 구성의 것이다.  <!-- rejected-value -->

- 원본: `~/.config/tossctl/engine.log` (읽기 전용), live journal (`mode=ro&immutable=1`)
- **네 판 연속 열거나 측정에서 틀렸다.** 1판은 생산자를, 2판은 **연결 상태**를,
  3판은 **창**을 틀렸고, 4판 초안은 **종류 열거**(flatten)와 **경합 없는 프로브**를
  틀렸다. 5판이 고치는 것은 넷째다.
- **이 문서가 측정의 정본이다.** 7판부터 proposal·design·tasks·spec이 인용하는 지연
  계열 수치는 전부 여기(또는 `function-logic/`)에 실려 있어야 하고,
  `tools/check_values.py`가 그것을 검사한다. 6판이 proposal에 적었던 `28.9 ms`는 **어떤 측정 표에도 없는 수**였다(6라운드 차단 B2). <!-- rejected-value -->
- **기각된 수를 이 문서가 인용할 때는 그 줄에 `<!-- rejected-value -->`를 단다.**
  7라운드 B2: 표지가 없으면 `check_values.py`의 측정 코퍼스가 이 문장에서 그 수를
  긁어 가 **기각된 값을 영구히 면역시킨다.** 기각을 적는 문장이 기각을 무효로 만드는
  구조였다.

## 0. 창 — 커밋이 아니라 **프로세스 재시작**이다

3판은 창을 "publisher 배선 커밋 `e540668f`(2026-08-04) 이후"로 썼다. **틀렸다.**
커밋이 존재하는 것과 **돌던 프로세스가 그것을 갖고 있는 것**은 다르다.

증거는 로그 자신에 있다.

```text
line 6837  2026-08-05T00:11:44.202Z  ERROR engine.alert_undelivered
           error = "no notification publisher is configured"
```

그 문자열은 `internal/obs/notifier.go:253` 한 곳에만 있고, 그 줄은
**`n.Publisher == nil`일 때만** 닿는다. 즉 커밋 다음 날 00:11에도 **돌던 프로세스**에는
publisher가 없었다.

> **"바이너리"라고 쓰지 않는다.** `resolveNotificationPublisher`는 **설정** 때문에도
> nil을 돌려준다(`:81`·`:87`·`:97`). 로그가 증명하는 것은 *그 프로세스에 publisher가
> 없었다*는 것이고, 코드가 없었는지 설정이 없었는지는 가르지 않는다.
> **창의 결론은 어느 쪽이든 같다** — 그 구간에 publish는 일어날 수 없었다.

publisher를 가진 프로세스는 **line 6862-6871의 재시작**에서 시작한다
(`journal lock` 배너 = 새 프로세스). 첫 JSON 줄이 line 6866
(`2026-08-05T00:36:18.597Z`)이다.

> **창 = engine.log line 6866 이후.** 3판의 창은 여기보다 **약 24.5시간 이르고**,
> 그 구간에는 publish가 한 건도 일어날 수 없었다.
> (3판 자신도 §5에서 "배선 이전 … 2026-07-31 ~ 08-05"라고 써서 §4와 모순됐다.)

## 1. "publish를 유발하는 줄"의 정의

전송 지연을 논하려면 **직전에 어떤 발송이 있었는지**를 알아야 하고, 그러려면 로그의
어떤 줄이 `Publisher.Publish`를 부르는지 확정해야 한다.

**`severity` 필드는 판별자가 아니다.** `Logger.emit`(`log.go:178-188`)이 **모든** 줄에
`FieldSeverity`를 붙인다. `severity: critical`인 줄이 발송을 뜻하지 않는다.

**등급도 판별자가 아니다.** `Notify`(`notifier.go:107-116`)는 critical이면
`notifyCritical`, 정상이면 `publishBestEffort`(`:138-150`)로 가고 **둘 다 `Publish`를
부른다.** 정상 등급도 같은 전역 연결 풀을 데운다 — 2판 C1이 여기서 틀렸다.

### 1.1 `.Notify(`에 닿는 이벤트 **16종** — 그중 프로덕션에서 실제로 publish하는 것은 12종

| 조립 자리 | 이벤트 종류 | 등급 | 프로덕션 publish |
|---|---|---|---|
| `exitloop.go:780` | `exit.observation_outage` | critical | ✅ |
| `exitloop.go:1430` | `exit.proposal_capped` | normal | ✅ |
| `exitloop.go:1500` | `exit.position_unmanaged` | normal | ✅ |
| `exitloop.go:1526` | `exit.judgement_refused` | critical | ✅ |
| `exitloop.go:1550` | `exit.proposal_refused` | critical | ✅ |
| `exitloop.go:1580` | `exit.liquidation_delayed` | critical | ✅ |
| `exit_quarantine_announce.go:71` | `exit.snapshot_quarantined` | critical | ✅ |
| `adoption.go:353` | `exit.position_adopted` | normal | ✅ |
| `adoption.go:417`·`:455` | `exit.position_unmanaged` | normal | ✅ |
| `exitwiring.go:103` | `exit.position_unmanaged` | normal | ✅ |
| `exitwiring.go:141` | `exit.position_closed_externally` | normal | ✅ |
| `runtime.go:316` | `engine.loop_failed` | critical | ✅ |
| `runtime.go:396` | `engine.loop_degraded` | critical | ✅ |
| `obs/mode.go:57` | `engine.operating_mode` | critical | ✅ |
| `flatten/flatten.go:186`·`:278`·`:403`, `liquidate.go:238`·`:245` | `flatten.started` · `flatten.stalled` · **`flatten.complete`** · `order.in_doubt` | started·stalled·in_doubt는 critical, **complete는 normal** | ❌ **로그 전용** |

**flatten 4종은 프로덕션에서 `Notify`에 닿지 않는다.** `Saga.event`(`flatten.go:689-700`)는
`s.Notifier == nil`이면 `s.logf` 가지를 타는데, 프로덕션 `flatten.Saga` 리터럴
(`cmd/tossctl/flatten.go:247-263`)은 **`Notifier`를 채우지 않는다.**
그러므로 실제 publish 유발 종류는 **12종**이다.

> **3판·4판 초안이 두 번 틀렸다.**
> 3판은 `exit.snapshot_quarantined`를 빠뜨렸고, 4판 초안은 flatten을 **3종**으로 세면서
> (`flatten.complete` 누락) 동시에 그것들을 **publish 유발로 분류**했다.
> **두 오류 다 자매 산출물이 이미 답을 갖고 있었다** — `analysis/notify-reach.md`가
> `exit_quarantine_announce.go:71`을 P4 호출자 표에, P7을 "프로덕션에서 `Notifier`가
> nil"로 이미 적어 두었다. **산출물을 만들고도 안 쓰면 산출물이 없는 것과 결과가 같다.**

### 1.2 로그 전용 생산자를 가진 4종 — 판별자가 필요하다

engine 12종 중 나머지 8종은 `Notify` 말고는 그 종류를 쓰는 자리가 없어
**종류만으로 정확하다.**

| 이벤트 종류 | 로그 전용 생산자 | 판별자 |
|---|---|---|
| `engine.operating_mode` | **7곳** — `notifier.go:221`·`:228`, `runtime.go:417`·`:423`, `interlock.go:467`·`:470`, `exitloop.go:799` | **`from_state` 필드 보유** (`obs/mode.go:68`에서만 쓴다) |
| `engine.loop_failed` | `runtime.go:330` (`r.log(..., true, ...)` → **WARN, 레벨이 겹친다**) | **`stopped_with` 필드 보유** (`:316`의 Fields에만 있다) |
| `exit.proposal_capped` | `exitloop.go:1412` (`o.logErr` → ERROR) | 레벨 **INFO** (정상 등급이라 `Log.Event`) |
| `exit.snapshot_quarantined` | `exitloop.go:576` (`o.log(..., false, ...)` → INFO) | 레벨 **WARN** (critical이라 `Log.Warn`) |

> **3판은 `engine.operating_mode`의 생산자를 "6곳"이라고 썼고 표에는 5행만 있었다.
> 실제는 8곳(`Notify` 1 + 로그 전용 7)이다.**

### 1.3 그래서 `AnnounceOperatingMode`는 몇 번 발동했는가

`from_state`를 가진 줄은 **전체 로그에서 line 372 하나다** —
`2026-07-31T09:55:49`, `NORMAL → ENTRY_BLOCKED`, `AUTO`, `BROKER_AUTH_REJECTED`.
live journal의 `operating_modes` 테이블도 정확히 그 **1행**뿐이고 이후 완화된 적이 없다.

> **⇒ `AnnounceOperatingMode`는 전 기간 1회 발동했고, 창 안에서는 0회다.**
> `notify-reach.md`의 P1a(`checkOutage:796`)는 **한 번도 발동한 적이 없다.**
> `TransitionOperatingMode` B15(`operating_mode.go:410` → `return :415`)가
> announce(`:479`) 전에 반환하기 때문이다. 이것이 H1이다.

## 2. 무엇을 재는가

`deliver:259`의 `alert_undelivered` 줄은 **`Publisher.Publish`가 nil을 돌려준 뒤에만**
쓰인다(`notifier.go:257-260`). 그러므로

```text
[logEvent :109 의 이벤트 줄]  →  [바로 다음 줄이 deliver:259 의 alert_undelivered]
```

간격 = `EnqueueAlert` + **성공한 publish 1회** + 실패한 `MarkAlertDelivered`.

**엄격한 인접만 쓴다.** 창을 `n+2`까지 넓히면 한 `alert_undelivered` 줄이 두 이벤트 줄에
겹쳐 짝지어져 **없는 표본이 생긴다**(3판 작업 중 실제로 발생시켰다가 폐기).

그 줄이 왜 나오는가: `EnqueueAlert`는 같은 `event_key`를 두 번째로 넣으면 **기존 행 id를
그대로 돌려준다**(`outbox.go:128-131`, AST B4/B5). 그 행이 이미 DELIVERED면
`MarkAlertDelivered`의 `WHERE ... AND state = PENDING`이 0행을 갱신해 오류가 된다
(`outbox.go:154-159`). 로그 문구가 정확히 그것이다 —
`journal: no such alert: 12 (or it is no longer pending)`.

## 3. 표본 6건 — 창 안에서 재현된다

직전 publish는 **§1.1의 프로덕션 publish 12종 전부**를 대상으로 찾았다(정상 등급 포함). 임계는
`http.Transport`의 `IdleConnTimeout` **90초**이고, `Ntfy.Publish` B7 `:122`가 만드는
`&http.Client{Timeout: timeout}`은 `Transport`가 nil이라 **프로세스 전역
`http.DefaultTransport` 풀**을 쓴다.

| # | 줄 | 이벤트 | 조립 자리 | 직전 publish 간격 | 연결 | 체류 |
|---|---|---|---|---|---|---|
| 1 | 6896→6897 | `exit.proposal_refused` | `exitloop.go:1550` | 11.0 s | 온 | **0.1983 s** |
| 2 | 6899→6900 | `exit.proposal_refused` | `exitloop.go:1550` | 51.6 s | 온 | **0.2819 s** |
| 3 | 6910→6911 | `exit.proposal_refused` | `exitloop.go:1550` | **260.8 s** | **냉** | **0.7540 s** |
| 4 | 6912→6913 | `exit.proposal_refused` | `exitloop.go:1550` | 6.0 s | 온 | **0.2358 s** |
| 5 | 6980→6981 | `exit.judgement_refused` | `exitloop.go:1526` | 0.7 s | 온 | **0.2027 s** |
| 6 | 7181→7182 | `exit.judgement_refused` | `exitloop.go:1526` | 0.4 s | 온 | **0.2061 s** |

**표본 6건은 전부 창(line 6866 이후) 안에 있다.** C1이 무너뜨린 것은 §4의 모집단
통계이지 이 표가 아니다 — 4판에서 값이 하나도 바뀌지 않았다.

두 이벤트 종류 다 프로덕션 조립 자리가 **1곳뿐**이고 둘 다 exit 관측 루프다.

## 4. 냉연결 비율 — 39건 중 18건(46%)

창 안에서 §1의 규칙(프로덕션 publish 12종 + 판별자 4개)으로 센 결과:

| | 값 |
|---|---|
| 종류만으로 센 후보 줄 | 44 |
| **판별자를 통과한 publish 유발 줄** | **39** |
| **그중 냉(직전 publish로부터 90초 초과, 또는 재시작 직후)** | **18 (46%)** |
| 창 안의 프로세스 재시작 | 5회 (line 6971·7172·7745·7803·7806) |

종류 내역: `exit.position_unmanaged` 16 · `exit.position_adopted` 12 ·
`exit.proposal_refused` 5 · `exit.judgement_refused` 4 ·
`exit.position_closed_externally` 2. **`engine.operating_mode`는 0건이다** — §1.3과 일치한다.

> **3판은 "47건 중 24건(51%)"이라고 썼다.** 그 47은 창이 24.5시간 일러서
> publish가 일어날 수 없던 줄을 포함했고, 종류 열거와 판별자도 불완전했다.
> **방향은 살아남는다(46% ≈ 절반) — 냉은 이 워크로드의 예외가 아니라 정상 상태다.**

그리고 그 18건 중 **측정 가능한 것은 1건뿐**이다 — 나머지는 `MarkAlertDelivered`가
성공해 흔적을 남기지 않았다.

| | 표본 수 | 범위 |
|---|---|---|
| **냉** | **1** | **0.7540 s** |
| 온 | 5 | 0.1983 ~ 0.2819 s |

**⇒ 설계는 냉을 기준으로 잡아야 하고, 냉에 대한 측정은 하나다.**

> **2라운드가 요구한 둘째 냉 표본 0.62 s는 여기 없다. 왜 없는지를 적는다**
> (6라운드 M4 — 폐기를 침묵으로 처리한 것이 지적이었다).
>
> 그 값은 2라운드 리뷰가 "재기동 직후라 냉"이라고 본 **로그 두 줄의 간격**
> (`6978→6979`)이고, 뒤쪽 줄은 `exit.position_unmanaged`다. **그 타입의 줄 간격은
> §5.3이 이미 무효로 판정한 방법이다** — 조립 자리가 4곳이고 셋이 대사 루프
> goroutine이라(`runtime.go:277-283`) 간격이 어느 루프의 체류도 재지 못한다.
> 같은 표본군이 `publishBestEffort` 산출물의 무효 9건에 `0.623 s`로 들어 있다.
>
> **즉 폐기의 이유는 "빼고 싶어서"가 아니라 "같은 판정이 그 값에도 걸리기 때문"이다.**
> 그 판정을 0.62 s에만 적용하지 않는 것이 오히려 일관성을 깨뜨린다.
> 냉 표본이 **n=1**이라는 것은 그래서 완화되지 않는다 — design D2가 그것을 정면으로 쓴다.

## 5. 이 표본의 한계 — 먼저 쓴다

1. **성공만 잡힌다.** 실패한 publish는 이 창을 만들지 않는다. 창 안에서 transport 오류로
   인한 `alert_undelivered`는 **0건**이다. **오늘의 34초는 유도된 상한이고 관측된 적이 없다.**
2. **재발송에 치우쳐 있다.** 최초 발송은 `MarkAlertDelivered`가 성공해 줄을 안 남긴다.
   publish 자체는 같은 호출이다.
3. **정상 등급은 표본이 없다.** `exit.position_unmanaged`는 조립 자리가 **4곳**이고
   그중 셋이 대사 루프 goroutine이다(`runtime.go:277-283`). 줄 간격은 어느 루프의
   체류도 재지 못한다.
4. **인접 짝짓기는 goroutine 간섭에 취약하다.** `logEvent`는 `notifier.go:109`에서
   **`n.mu`를 잡기 전에** 돌고, 하나의 `*obs.Notifier`를 다섯 루프가 공유한다
   (`gateway.go:280`). 다른 goroutine의 이벤트 줄이 사이에 끼면 잘못 짝지어질 수 있다.
   **오차 방향은 보수적이다** — 끼어든 줄만큼 측정값이 실제 publish보다 **길게** 나온다.
   §3의 6건에서 그런 끼어듦은 없었다(짝지은 줄이 모두 인접).
5. **§3의 분해에는 `n.mu` 획득이 빠져 있다.** 그것도 측정 창 안에 있고, 역시
   측정값을 길게 만드는 방향이다.
6. **n=1에서 1회 상한을 고르는 것은 측정이 아니라 판단이다.** design D2가 그렇게 쓴다.

## 6. 창 이전 (참고)

창 이전 `alert_undelivered` 줄의 오류는 전부
`no notification publisher is configured`(`notifier.go:253`)다 — **발송이 시도된 적
없는 구간**이다. §3의 6건은 전부 창 안이다.

## 7. 후보 상수의 국소 실측 — 이것은 live 데이터가 아니다

§3~§5는 운영 로그다. 이 절은 **저장소를 건드리지 않고**(`go test -overlay`) 돌린
국소 프로브이며, live 계좌와 무관하다.

프로브 모양: `net.Listen`만 하고 **accept 하지 않는다**(커널이 backlog에서 핸드셰이크를
끝내므로 클라이언트는 응답 읽기에서 기한까지 막힌다) → `openTestJournal` +
`execgw.NewEntryGate` + `newNotifier` + critical `Notify` 1회.

### 7.1 초과분 측정이 두 번 커졌다 — 프로브가 뺀 조건이 있었다

| 회차 | 후보 상수 | 조건 | 최악 초과분 |
|---|---|---|---|
| 3판 (3라운드 리뷰) | T=1.5s · D=250ms (transport = 주기) | 네트워크만 / 원장·게이트까지 | **체류 5.003 s / 5.023 s** — 초과분이 아니라 **주기 초과** |
| 4판 초안 (10회) | T=1.4s · D=200ms | `-race` · `GOMAXPROCS=2` · CPU 3배 초과구독, **유휴 journal** | 71.9 ms |
| 4라운드 리뷰 | 같음 | + 다른 루프의 연속 원장 쓰기 | 142.2 ms |
| **4판 확정 (6회)** | **T=1.3s · D=150ms** | 위 전부 + `EnqueueAlert` 연속 경합 | **168.3 ms** |

**같은 3라운드가 잰 것이 하나 더 있다 — 잘못 조립한 프로브의 체류 521 ms.**
`Ntfy`에 `Topic`을 안 채우면 `Publish`가 **다이얼 전에** `ErrNtfyNotConfigured`로
반환하므로(`ntfy.go:86-89`) 세 시도가 전부 즉시 실패한다. 3판의 테스트 리터럴이
그 상태였고 체류가 **521 ms**로 나왔다 — 기한이 아니라 조립 실패를 잰 값이다.
tasks 6.6이 이 수를 함정 경고의 근거로 인용하므로 여기 싣는다. **9판 전까지 이 값은
tasks에만 있고 어떤 측정 표에도 없었다** — 8판의 값 단위 검사가 정수 측정치를
못 보던 동안(8라운드 M-7) 여덟 판을 살아남은 고아다.

**첫 행이 이 표에 있는 이유**: 3판은 transport를 주기와 **같게** 잡았고(후보 #1,
reserve 0), 그래서 비전송 작업만큼 반드시 초과했다. 5.003 s는 네트워크만, 5.023 s는
원장·게이트까지 포함한 체류다. **이 두 값은 초과분이 아니라 "등호가 왜 틀렸는지"의
증거**이고, design D2가 그것을 근거로 부등식으로 갈아탔다. 6판까지 이 수치는
design과 proposal에만 있고 측정 표에는 없었다 — 7판이 여기로 옮긴다.

**커진 이유**: journal은 `SetMaxOpenConns(1)`이고(`internal/journal/journal.go:151`,
`synchronous=FULL` · `_txlock=immediate`) `Notify` 한 번이 그 **한 개 커넥션**에서
트랜잭션 여러 개를 돌린다. 다른 루프의 쓰기가 있으면 전부 줄을 선다.
**4판 초안의 프로브는 아무도 안 쓰는 새 임시 DB를 써서 이 항을 통째로 뺐다.**

### 7.2 12판까지의 후보 #3 실측 — **폐기된 구성이다** <!-- rejected-value -->

> **⚠ 이 절의 제목은 12판까지 "확정 상수의 실측"이었다.** 그 구성(시도 3회)은
> 13판이 기각했고, 그 뒤 13판의 구성(1회·1.3s)도 14판이 기각했다.
> **현재 채택 구성의 측정은 §9다.** 아래 값은 폐기 구성의 것이며, 살아 있는 부분은
> **비전송 초과분**(90.3~168.3 ms)뿐이다 — 그 항은 시도 횟수에 의존하지 않는다.

transport 예산 = 3 × 1300 ms + 2 × 150 ms = **4.200 s**, reserve = **800 ms**. <!-- rejected-value -->

**이 두 수는 더 이상 채택값이 아니다.** 현재는 transport 3.500 s · reserve 1500 ms.

| 조건 | 실행 | 체류 | 초과분 | 주기(5s)까지 여유 |
|---|---|---|---|---|
| `-race`, 유휴 journal | 3 | 4.2903 / 4.2930 / — s | 90.3 ~ 93.0 ms | 707.0 ~ 709.7 ms |
| `-race`, **원장 경합** | 3 | 4.3608 / **4.3683 s** | 160.8 ~ **168.3 ms** | 631.7 ~ 639.2 ms |

- **최악 체류 4.3683 s < 5.000 s.** 이 회차의 절대 마진 **631.7 ms**.
- 초과분은 `EnqueueAlert` 1회(outbox 기록) + `MarkAlertAttemptFailed` 3회(시도별 실패 기록)
  + `Gate.Block`(게이트 래치) + `n.escalate`의 `EscalateOperatingMode` 승격 트랜잭션
  + 그 호출이 쓰는 구조화 로그 줄 + 기한 오버슈트의 합이다.
  **이 회차는 로거가 nil이라 마지막 항이 0이었다** — §7.5가 로거를 실제로 달고 다시 잰다.
  9판까지 이 줄은 로그 항 없이 넷만 셌다(9라운드 차단 2). 측정 정본이 구성을 적게
  세면 그 항이 예산 밖으로 읽힌다.
- `-race` 자체는 체류를 의미 있게 늘리지 않는다. 늘리는 것은 **원장 경합**이다.

#### 7.2.1 이 숫자는 기계·주변 부하에 종속이며 재현되지 않는다 (5라운드 M1)

**같은 기계·같은 채택 상수에서 초과분이 11.2배 흩어진다.**
(아래에서 "산포 배수"라 이름 붙이는 값이고, **이 문서군 전체에서 한 표기로 쓴다** —
8라운드 M-5: 8판까지 `11배`와 `11.2배`가 같은 양을 두 표기로 말했고,
그 이름의 검사가 `11` 쪽만 읽었다.)
5라운드 리뷰어가 다른 값을 얻어 이것을 드러냈고, 재측정으로 확인했다.

| 회차 | 조건 | 최악 초과분 |
|---|---|---|
| 4판 확정 (위 표) | `-race` · CPU 3배 · `EnqueueAlert` 경합 | 168.3 ms |
| 5라운드 리뷰어 | `-race` · CPU 3배 · `EnqueueAlert` 경합 | **356.1 ms** |
| 5판 재측정 A | `-race` 없음 · writer 0/1/2/4 | 36.9 ~ 54.9 ms |
| 5판 재측정 B | `-race` · `GOMAXPROCS=60`(코어 20의 3배) · writer 0/1/2/4 | 31.9 ~ **77.8 ms** |
| 4라운드 N1 — **다른 기계** | 같은 채택 상수 | 60.7 ~ 82.9 ms |

**두 개의 비를 구별해서 읽는다 — 하나만 "산포 배수"라고 부른다.**

- **산포 배수 = 11.2배** — 관측 전체의 범위 356.1 ÷ 31.9. 유도 규칙이 상대하는 값이다.
- 조건을 문자 그대로 맞춘 두 회차 사이만 봐도 356.1 ÷ 77.8 = 4.6배다. 이것은
  **범위가 아니라 재현 실패의 폭**이고, 위 값과 다른 양이다.

6판은 이 둘을 같은 문서 안에서 "10배"와 "5배"로 적어 어느 쌍의 비인지 말하지 않았다
(6라운드 M-2). **두 값 다 참이지만 이름이 없으면 둘 중 하나는 오독을 만든다.**

어느 쪽으로 읽어도 결론은 같다 — 이 값을 지배하는 것은 프로브의 조건이 아니라
**측정 순간의 주변 부하**이고, 다른 기계에서 잰 4라운드 N1의 60.7~82.9 ms가
그것을 한 번 더 보인다. 그래서:

1. **어떤 한 회차의 마진도 설계의 성질이 아니다.** 4판이 "절대 마진 631.7 ms"라고
   쓴 것은 그 회차의 값이지 a092가 보장하는 값이 아니다.
2. **유도 규칙은 관측 최댓값을 입력으로 받아야 한다** — 회차 최댓값이 아니라.
   전 회차 최악은 **356.1 ms**이고 reserve 800 ms는 그것의 **2.25배**다.
   design D2의 규칙은 그렇게 다시 쓴다.
3. **배포 기계에서 다시 재는 것이 tasks 7.2의 산출이다.** 여기 실린 값은 이 기계의
   값이고, 다른 기계에서 규칙을 만족하는지는 그 기계에서 확인해야 한다.

reserve 800 ms는 **관측된 어떤 초과분보다도 2배 이상 크다** — 이 결론은 네 회차 전부에서
성립하고, 그것이 규칙이 값의 산포를 견디는 방식이다.

### 7.3 한계 — 프로브가 재현하지 못하는 것

1. **로컬 SSD의 fsync다.** 운영 디스크가 느리면 `synchronous=FULL` 커밋이 길어진다.
   800 ms 몫은 그 여지를 위한 것이지 여유분이 아니다.
2. **WAL checkpoint를 강제하지 않았다.** 운영 WAL은 3.9MB이고, `Notify`가 도는
   트랜잭션 중 하나 안에서 checkpoint가 일어날 수 있다.
3. **`checkOutage`의 두 번째 동기 대기는 이 프로브 밖이다.** `EscalateOperatingMode`
   (`exitloop.go:796`)는 `Notify` 호출이 아니라 그 다음 호출이다. a092의 계약은
   **`Notify` 하나**를 유계로 만든다.
4. **`n.mu` 경합은 이 표에 없다** — 초과분이 아니라 **배수**다. §7.4.

### 7.4 `n.mu` 경합 — 상한의 배수, 채택 상수에서 실측 8.458초

`deliver`는 `n.mu`를 재시도 전체(대기 포함) 보유하고(`notifier.go:241-242`)
`*obs.Notifier`는 다섯 루프가 공유하는 **하나**다(`gateway.go:280`).
두 goroutine이 같은 전송기로 critical을 올렸을 때, **채택 상수
T=1.3초·D=150ms·시도 3회(transport 4.200초)** 에서 3회 실측:

```text
run0  loopA = 4.240848032s   loopB = 8.454018981s   ratio(주기 5s) = 1.691
run1  loopA = 4.234955218s   loopB = 8.457964606s   ratio(주기 5s) = 1.692
run2  loopA = 4.234387020s   loopB = 8.453599198s   ratio(주기 5s) = 1.691
```

최악 **loopB = 8.458초** — 관측 주기의 **1.69배**, 자기 차례의 **2.00배**.
뒤에 잠금을 얻는 쪽의 `Notify` 체류에 앞쪽 체류가 **통째로** 더해진다.
그 대기는 `Notify` 호출 **안**이므로 호출자의 동기 체류다.

#### 4판이 여기 실었던 9.231초는 채택하지 않은 상수의 값이다 <!-- rejected-value -->

5라운드 H1이 짚었고 **대조군으로 확정했다.** 같은 프로브를 후보 #2
(T=1.4초·D=200ms, transport 4.600초)로 돌리면:

```text
run0  loopA = 4.655796603s   loopB = 9.156676151s
run1  loopA = 4.653055753s   loopB = 9.153851745s
run2  loopA = 4.663135757s   loopB = 9.163681864s
```

4판이 적은 `loopA dwell = 4.652 s`가 **이 값이다** — 소수 셋째 자리까지 일치한다.
그리고 같은 문서 §7.2는 채택 상수의 loopA를 4.2903~4.3683초로 싣고 있었다.
**한 문서가 같은 프로브 모양의 loopA에 두 개의 양립 불가능한 값을 싣고 어느 쪽도
어느 상수에서 나왔는지 말하지 않았다.** 그 상태로 engine-safety 델타의 SHALL이
9.231초를 근거로 인용했다 — **a092가 만들지 않는 수**를. <!-- rejected-value -->

그래서 5판부터 배수를 적을 때는 **어느 상수에서 쟀는지를 같은 줄에 적는다**.
그 규칙은 spec 델타에도 SHALL로 들어갔다.

#### a092는 이 배수를 고치지 않는다

전송이 루프 밖으로 나가야 하고 그것이 미배정 후속이다(proposal §미배정 후속 4번).
고치지 않는 대신 **spec이 그것을 말한다**: engine-safety 델타는 상한을 "자기 차례"로
한정하고 직렬화가 배수를 만든다는 것을 이 값과 함께 적는다.
**4판 초안은 이 사실을 exit-policy ¶20에서 인정해 놓고 ¶16에서 반대로 SHALL NOT을
걸었고, engine-safety에는 아예 안 썼다.** 그것이 4라운드 C1이다.

> **프로브가 `Notify`의 반환을 버리는 이유**: `err`는 nil이다 — `Notify`는 발송
> 실패를 호출자 오류로 돌려주지 않는다(`notifier.go:109-113` 주석). 게이트 래치와
> outbox 행이 실패의 기록이다. 그래서 체류는 재는 값이고 오류는 재는 값이 아니다.

### 7.5 8판 재측정 — 로거를 달고 다시 쟀다 (7라운드 H2)

**7라운드가 짚은 것**: 위 표를 만든 프로브는 `newNotifier`의 **4번째 인자**
(`log *obs.Logger`, `exitwiring.go:71-72`)에 `nil`을 넘겼다. `Notifier`의 로그 쓰기는
전부 `n.Log != nil` 뒤에 있으므로, 그 프로브는 소진 경로의 구조화 로그 쓰기를
**구조적으로 제거한 상태**에서 초과분을 쟀다. 그리고 D3의 reserve 구성 열거에도
그 항이 없었다. **측정과 열거가 같은 곳에서 같은 항을 빠뜨렸다.**

**소진 경로가 쓰는 로그 줄 — 소스 열거**:

| # | 자리 | 등급 | 조건 |
|---|---|---|---|
| 1 | `Notify` `:109` → `logEvent` `:131` `Log.Warn(e.Type, …)` | Warn | critical이면 **항상** |
| 2 | `deliver` `:279` `Log.Error(EventAlertUndelivered, lastErr, …)` | Error | 시도 소진 시 **항상** |
| 3 | `escalate` `:228` `Log.Warn(EventOperatingMode, …)` | Warn | 승격이 **실제로 바뀌었을 때** |

3은 조건부다 — 이미 ENTRY_BLOCKED면 `changed`가 거짓이라 줄이 안 나온다. 그러므로
**3줄은 이 경로의 최댓값**이고 반복 알림은 2줄이다. 그 밖에 `:221`·`:259`·`:265`가
있으나 셋 다 원장 쓰기 자체가 실패했을 때만 돈다(정상 소진 경로 아님).

**프로브**: 로거를 **유일한 변수**로 두고 나머지를 고정한다. 로거는 프로덕션과 같은
모양 — `obs.NewLogger(obs.LogOptions{Writer: <파일>, JSON: true, Clock: clock.System()})`
(`cmd/tossctl/engine.go:201`이 `errOut`에 JSON으로 쓰고 배포에서 그 fd가 파일이다).
조건: go1.26.5 · 코어 20 · `GOMAXPROCS=60` · `-race` · 셀마다 5회 · `go test -overlay`
(저장소 무편집).

| 셀 | 로거 | 원장 경합 | transport | 초과분 best ~ worst | 로그 줄 |
|---|---|---|---|---|---|
| A | **nil** | 없음 | 4.200 s | 79.9 ~ 88.3 ms | 0 |
| B | **실제** | 없음 | 4.200 s | 75.7 ~ 94.9 ms | **3** |
| C | **nil** | writer 2 | 4.200 s | 127.7 ~ **319.3 ms** | 0 |
| D | **실제** | writer 2 | 4.200 s | 112.0 ~ 209.6 ms | **3** |
| E | 실제 | writer 2 | 4.000 s | 123.6 ~ 218.9 ms | **3** |

**이 회차의 두 최악 — 아래 결론이 쓰는 값은 이 둘뿐이다.**

- **로거 nil 셀의 최악 초과분 319.3 ms** (셀 C)
- **로거 실제 셀의 최악 초과분 218.9 ms** (셀 E — 채택 상수만 보면 셀 D의 209.6 ms)

**결론 — 이 설계로는 로그 세 줄이 안 잡힌다.**

1. **A↔B(로거만 다름)**: 76~95 ms 대 80~88 ms로 **겹친다.** B의 best가 A의 best보다
   **낮다.**
2. **C↔D(로거만 다름)**: 로거 nil 셀의 최악 초과분 319.3 ms가 로거 실제 셀의
   최악보다 **크다.** 부호가 반대다.
3. **그러나 2를 "로그 쓰기가 측정 한계 아래"의 증거로 쓰지 않는다**(8라운드 H4).
   셀당 n=5이고 보고한 통계는 min/max뿐이다. 11.2배 산포가 지배하는 분포에서
   max-of-5의 순서 역전은 **"효과가 없다"가 아니라 "이 설계로는 안 잡힌다"**이다.
   세 줄의 실제 크기는 **미측정**이고, 그것을 재려면 짝지은 반복과 분포 통계가
   필요하며 이 change는 그것을 하지 않는다.
4. **하중을 지는 것은 이것 하나다**: 전 회차 최악은 여전히 **356.1 ms**(5라운드)이고
   이번 회차 최악 319.3 ms는 그보다 작다. **로거를 달아도 유도 규칙의 입력이 안 바뀐다** —
   판정값 `800 / 356.1 = 2.25배`가 그대로 선다. reserve 800 ms는 유지된다.
5. **틀린 것은 값이 아니라 열거였다.** D3 주석과 두 delta의 SHALL 열거에 로그 쓰기를
   **더한다** — 크기 때문이 아니라, 열거가 완전하지 않으면 다음 편집이 그 항을
   예산 밖으로 읽기 때문이다. **크기를 몰라도 열거는 완전해야 한다.**

**셀 E는 M-4에 답한다** — `alertRetryDelay`를 150 ms에서 50 ms로 내려 transport를
4.000 s로 만든 대안이다. 초과분은 123.6~218.9 ms로 **D와 같은 띠**다. 즉 **초과분은
D에 의존하지 않는다** — reserve를 200 ms 더 사도 그 200 ms가 막을 새 항이 관측되지
않는다. 판단은 design D2에 적는다.

#### 7.5.1 재현 — 프로브 **골자** (9판이 실었고, 실행되지 않았다)

> **⚠ 아래는 실행 가능한 파일이 아니다.** 9라운드 H-3이 확인했다 —
> `package`·`import`·`func` 선언이 없고 `withLog`·`timeout`·`attempts`·`delay`·
> `writers`·`stop`·`key`·`logBytes`·`t` 아홉 개가 미정의이며 overlay 인자가
> 자리표시자다. 그래서 §7.5의 다섯 셀은 **9판에서도 독립 재검증이 불가능**했다.
> 실행 가능한 전문과 그것으로 다시 잰 결과는 **§7.6**에 있다.
> 이 절은 9판이 실은 것을 이력으로 남긴 것이고, 근거로는 §7.6을 쓴다.

**8판은 이 측정으로 두 delta의 SHALL 열거를 바꿔 놓고 프로브를 안 실었다**
(8라운드 H3). 7.1.1은 명령을 통째로 실었는데 여기는 조건 서술만 있었다.
아래를 `internal/app/engine`에 `-overlay`로 넣으면 위 표가 재현된다 — **저장소는
편집하지 않는다.**

```go
// a092reserve_probe_test.go — package engine
//
// 로거를 유일한 변수로 두고 소진 경로 한 번의 체류를 잰다.
// 응답하지 않는 전송: listen만 하고 accept 하지 않는다 — 커널이 backlog에서
// 핸드셰이크를 끝내므로 클라이언트는 응답 읽기에서 기한까지 막힌다.

ln, _ := net.Listen("tcp", "127.0.0.1:0")      // accept 하지 않는다
j := openTestJournal(t)                        // interlock_internal_test.go:75
gate := execgw.NewEntryGate(clock.System(), nil)

var lg *obs.Logger
if withLog {                                   // ← 이 셀의 유일한 변수
    f, _ := os.Create(filepath.Join(t.TempDir(), "engine.log"))
    // 프로덕션과 같은 모양: cmd/tossctl/engine.go:201 이
    // obs.NewLogger(obs.LogOptions{Writer: errOut, JSON: true, Clock: clk})
    lg = obs.NewLogger(obs.LogOptions{Writer: f, JSON: true, Clock: clock.System()})
}

pub := &obs.Ntfy{BaseURL: "http://" + ln.Addr().String(),
    Topic: "a092-probe", Timeout: timeout}
n := newNotifier(j, gate, "a092-probe-account", lg, pub, clock.System())
n.Attempts, n.RetryDelay = attempts, delay

// 원장 경합: 다른 루프가 같은 한 개 커넥션에 쓰는 상태(journal.go:151)
for w := 0; w < writers; w++ {
    go func(w int) {
        for k := 0; ; k++ {
            select { case <-stop: return; default: }
            j.EnqueueAlert(context.Background(), journal.Alert{
                EventKey: fmt.Sprintf("a092-noise-%d-%d", w, k),
                Type: "probe.noise", Severity: "normal", Title: "noise"})
        }
    }(w)
}

start := time.Now()
err := n.Notify(context.Background(), obs.Event{
    Type: obs.EventExitProposalRefused, Key: key, Title: "a092 reserve probe",
    Body: "the probe holds the cycle for one alert",
    Fields: map[string]any{obs.FieldSymbol: "PROBE", obs.FieldReason: "a092-probe"}})
elapsed := time.Since(start)                   // 초과분 = elapsed - transport

// 로그 줄 수와 **종류**를 함께 센다 — 세 줄이 실제로 나갔는지가 이 표의 주장이다.
for _, want := range []string{
    string(obs.EventExitProposalRefused),      // notifier.go:131 logEvent WARN
    string(obs.EventAlertUndelivered),         // notifier.go:279 deliver ERROR
    string(obs.EventOperatingMode),            // notifier.go:228 escalate WARN
} {
    if !bytes.Contains(logBytes, []byte(want)) {
        t.Errorf("이 셀은 자기가 주장하는 경로를 안 지났다: %q 줄이 없다", want)
    }
}
```

```bash
GOMAXPROCS=60 go test -overlay <overlay.json> -race -count=1 -timeout 20m -v \
  -run TestA092OverheadReserve ./internal/app/engine/
```

**"로그 줄 3"이 왜 매 회차 3인지**: `openTestJournal(t)`가 **회차마다 새 DB**를 열므로
`escalate`의 `changed`가 항상 참이고 `notifier.go:228`이 매번 나간다. 프로덕션에서
이미 ENTRY_BLOCKED인 계정에 두 번째 알림이 오면 그 줄은 **안 나가고 2줄**이다 —
그러므로 **3은 이 경로의 최댓값**이다. 위 세 종류 확인이 없으면 이 표의 "3"은
숫자일 뿐이므로 프로브가 `t.Errorf`로 단언한다.

**§7.3의 한계 4건은 이 재측정에도 그대로 걸린다.** 로컬 SSD의 fsync, WAL checkpoint,
`checkOutage`의 두 번째 대기, `n.mu` 배수 — 로거를 달았다고 이 넷이 재현되지는 않는다.

### 7.6 10판 재측정 — 프로브를 싣고, 실어 놓은 그것으로 쟀다 (9라운드 H-3)

9라운드 H-3의 지적: §7.5.1이 "프로브 전문"이라고 적었으나 package·import·func
선언이 없고 변수 아홉 개가 미정의여서 **실행할 수 없었다.** 그래서 §7.5의 다섯 셀은
9판에서도 독립 재검증이 불가능한 채로 남았다.

10판은 프로브를 **컴파일되는 파일 하나로** 만들고, 그 파일로 다시 쟀다.
§7.6.1의 전문은 실행한 것 그대로이며 아무것도 잘라내지 않았다.

**결과** — 셀당 5회, best ~ worst. go1.26.5, `--- PASS: TestA092OverheadCells (118.18s)`:

| 셀 | 로거 | 원장 경합 | transport | 초과분 best ~ worst | 로그 줄 |
|---|---|---|---|---|---|
| A′ | nil | 없음 | 4.200 s | 53.0 ~ 75.9 ms | 0 |
| B′ | 실제 | 없음 | 4.200 s | 47.6 ~ 59.2 ms | 3 |
| C′ | nil | writer 2 | 4.200 s | 103.5 ~ 145.1 ms | 0 |
| D′ | 실제 | writer 2 | 4.200 s | 62.6 ~ **189.3 ms** | 3 |
| E′ | 실제 | writer 2 | 4.000 s | 92.2 ~ 113.8 ms | 3 |

**§7.5(9판 회차)와 절대값이 다르다. 그것이 이 문서의 예고다** — §7.2.1이
*"이 숫자는 기계·주변 부하에 종속이며 재현되지 않는다"*라고 적었고 이번 회차가
그것을 다시 보인다. 재현되는 것은 **절차와 구조**이고 그것이 프로브를 싣는 이유다.

**재현된 것** — 두 회차가 같게 말하는 것:

1. **로그 줄이 정확히 3이다** — 매 회차, 두 회차 모두. `logEvent`의 Warn
   (`notifier.go:131`), `deliver`의 소진 Error(`notifier.go:279`),
   `escalate`의 Warn(`notifier.go:228`). 셋째 줄이 매번 나오는 이유는
   `openTestJournal`이 회차마다 새 DB를 열어 `changed`가 언제나 참이기 때문이다.
   **프로덕션에서는 계정이 이미 ENTRY_BLOCKED이므로 2줄이다**(§1.3) —
   이 프로브는 그 점에서 프로덕션보다 한 줄 비싸다.
2. **원장 경합이 초과분을 키운다** — A′→C′가 약 2배, B′→D′가 약 3배. 두 회차 동일.
3. **초과분이 D에 의존하지 않는다**(7라운드 M-4의 질문). D=50ms 셀 E′와
   D=150ms 셀 D′가 이번에도 같은 띠다. D는 transport 항이고 초과분은 비transport 항이다.
4. **모든 셀이 reserve 800ms 아래**다. 이번 회차 최악은 셀 D′의 189.3 ms다.

**재현되지 않은 것** — 정직하게 적는다:

- **로거의 부호가 회차마다 뒤집힌다.** 9판 회차는 셀 B′의 최댓값이 셀 A′보다 컸고,
  이번 회차는 B가 A보다 **작다**(59.2 < 75.9). n=5 두 회차로는 세 줄의 쓰기 비용이
  잡히지 않는다 — **이 설계로는 안 잡힌다**는 뜻이지 비용이 0이라는 뜻이 아니다.
  **세 줄의 크기는 미측정으로 남긴다.**
- **절대값 전부.** 이번 회차 최악은 9판 회차의 최악보다 작고, 전 회차 최악
  356.1 ms(5라운드)보다 한참 작다. **reserve 유도가 쓰는 값은 계속 전 회차 최악
  356.1 ms이고**(800 / 356.1 = 2.25배) 이번 회차가 그것을 바꾸지 않는다.
  낮게 나온 회차로 상수를 옮기는 것은 이 change가 일곱 판 동안 기각된 형태다.

#### 7.6.1 프로브 전문과 명령

이 파일은 `internal/app/engine` 패키지 안에 들어간다 — `openTestJournal`과
`newNotifier`가 그 패키지의 비공개 이름이기 때문이다. **저장소에는 쓰지 않는다.**
`-overlay`로만 넣고, `go vet` 통과를 확인한 뒤 `go test`로 돌렸다.

```bash
mkdir -p /tmp/a092
# 아래 Go 전문을 /tmp/a092/a092_overhead_probe_test.go 로 저장한다.
ROOT=$(git rev-parse --show-toplevel)
echo "{\"Replace\":{\"$ROOT/internal/app/engine/a092_overhead_probe_test.go\":\"/tmp/a092/a092_overhead_probe_test.go\"}}" \
  > /tmp/a092/overlay.json
go vet  -overlay /tmp/a092/overlay.json ./internal/app/engine/
go test -overlay /tmp/a092/overlay.json -run TestA092OverheadCells \
        -count=1 -timeout 900s -v ./internal/app/engine/
```

```go
package engine

import (
	"context"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/JungHoonGhae/tossinvest-cli/internal/clock"
	"github.com/JungHoonGhae/tossinvest-cli/internal/execgw"
	"github.com/JungHoonGhae/tossinvest-cli/internal/journal"
	"github.com/JungHoonGhae/tossinvest-cli/internal/obs"
)

const (
	p75Attempts = 3
	p75Delay    = 150 * time.Millisecond
)

// cell 하나를 한 번 돌리고 (초과분, 로그 줄 수)를 돌려준다.
func p75Once(t *testing.T, withLog bool, writers int, timeout, delay time.Duration) (time.Duration, int) {
	t.Helper()
	transport := p75Attempts*timeout + (p75Attempts-1)*delay

	ln, err := net.Listen("tcp", "127.0.0.1:0") // listen만, accept 하지 않는다
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()

	j := openTestJournal(t)
	clk := clock.System()
	gate := execgw.NewEntryGate(clk, nil)
	pub := &obs.Ntfy{BaseURL: "http://" + ln.Addr().String(), Topic: "a092-probe", Timeout: timeout}

	var lg *obs.Logger
	logPath := filepath.Join(t.TempDir(), "engine.log")
	if withLog {
		lf, ferr := os.Create(logPath)
		if ferr != nil {
			t.Fatalf("log file: %v", ferr)
		}
		defer func() { _ = lf.Close() }()
		lg = obs.NewLogger(obs.LogOptions{Writer: lf, JSON: true, Clock: clk})
	}

	n := newNotifier(j, gate, "acct-probe", lg, pub, clk)
	n.Attempts = p75Attempts
	n.RetryDelay = delay

	// 원장 경합 — 같은 커넥션(SetMaxOpenConns(1))에 줄을 세운다.
	stop := make(chan struct{})
	var wg sync.WaitGroup
	for w := 0; w < writers; w++ {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			for i := 0; ; i++ {
				select {
				case <-stop:
					return
				default:
				}
				_, _ = j.EnqueueAlert(context.Background(), journal.Alert{
					EventKey: fmt.Sprintf("probe-writer-%d-%d", id, i),
					Type:     string(obs.EventExitProposalRefused),
					Severity: string(obs.SeverityCritical),
					Title:    "contention",
				})
			}
		}(w)
	}

	start := time.Now()
	nerr := n.Notify(context.Background(), obs.Event{Type: obs.EventExitJudgementRefused})
	elapsed := time.Since(start)
	close(stop)
	wg.Wait()
	if nerr != nil {
		t.Fatalf("outbox 쓰기가 실패하면 이 실행은 아무것도 안 잰다: %v", nerr)
	}

	lines := 0
	if withLog {
		b, rerr := os.ReadFile(logPath)
		if rerr == nil {
			for _, c := range b {
				if c == '\n' {
					lines++
				}
			}
		}
	}
	return elapsed - transport, lines
}

func TestA092OverheadCells(t *testing.T) {
	const runs = 5
	cells := []struct {
		name    string
		withLog bool
		writers int
		timeout time.Duration
		delay   time.Duration
	}{
		{"A  로거 nil   · 경합 없음 · transport 4.200s", false, 0, 1300 * time.Millisecond, p75Delay},
		{"B  로거 실제  · 경합 없음 · transport 4.200s", true, 0, 1300 * time.Millisecond, p75Delay},
		{"C  로거 nil   · writer 2 · transport 4.200s", false, 2, 1300 * time.Millisecond, p75Delay},
		{"D  로거 실제  · writer 2 · transport 4.200s", true, 2, 1300 * time.Millisecond, p75Delay},
		{"E  로거 실제  · writer 2 · transport 4.000s", true, 2, 1300 * time.Millisecond, 50 * time.Millisecond},
	}
	for _, c := range cells {
		best, worst, lines := time.Hour, time.Duration(0), -1
		for i := 0; i < runs; i++ {
			over, ln := p75Once(t, c.withLog, c.writers, c.timeout, c.delay)
			if over < best {
				best = over
			}
			if over > worst {
				worst = over
			}
			lines = ln
		}
		fmt.Printf("CELL %s | %.1f ~ %.1f ms | 로그 줄 %d\n",
			c.name, float64(best.Microseconds())/1000, float64(worst.Microseconds())/1000, lines)
	}
}
```


## 8. 재현

```python
# 창 시작 : engine.log line 6866 (publisher를 가진 프로세스의 첫 JSON 줄)
# 창 끝   : line 8353 = 2026-08-06T00:52:55Z   ← §4를 계산한 시점의 마지막 줄
#           **끝을 고정하지 않으면 §4의 숫자가 재현되지 않는다.** 로그는 계속 자라고,
#           같은 규칙을 나중에 돌리면 다른 값이 나온다(4라운드 M5가 42/20을 얻었다).
# 유발 줄 : §1.1의 **프로덕션 publish 12종**. flatten 4종은 제외한다 —
#           cmd/tossctl/flatten.go:247 이 Notifier를 안 채워 Saga.event가 로그 가지를 탄다.
#           단 아래 4종은 판별자를 통과해야 한다
#           engine.operating_mode      -> 'from_state' 필드 보유
#           engine.loop_failed         -> 'stopped_with' 필드 보유
#           exit.proposal_capped       -> level == INFO
#           exit.snapshot_quarantined  -> level == WARN
# 짝짓기  : 이벤트 줄 **바로 다음** 줄이 alert_id 없는 engine.alert_undelivered일 때만
# 냉 판정 : 직전 유발 줄과 90초 초과(http.Transport.IdleConnTimeout),
#           또는 사이에 'journal lock' 배너(프로세스 재시작)가 있으면 냉
```

## 9. 14판 실측 — 채택 구성(1회 · 3.5s)의 근거

**§7은 폐기 구성을 잰다**(문서 머리의 ⚠). 이 절이 현재 채택값의 정본이다.

### 9.1 전송기 왕복 실측 — 냉연결 표본 20건

13판의 1.3초는 engine.log에서 얻은 **냉 표본 1건(0.754s)** 위에 서 있었다.
14라운드가 전송기를 직접 재서 그것을 기각했다.

```bash
# 읽기 전용 GET. 프로세스가 매번 새로 뜨므로 표본은 전부 냉연결이다.
for i in $(seq 1 20); do
  curl -s -o /dev/null -w "%{time_appconnect} %{time_total}\n" https://ntfy.sh/
done
```

| 통계 | 값 |
|---|---|
| 표본 | **20건 (전부 냉연결)** |
| 평균 total | **1.795 s** |
| 중앙값 | 1.729 s |
| p90 | 2.191 s |
| 최소 | 1.431 s |
| **최대** | **2.721 s** |
| 평균 TLS까지(`time_appconnect`) | 약 0.50 s |

**이것이 13판의 1.3초를 기각한다.** 채택 상수가 **평균보다 작으면** 정상 전송기가
매 관측 실패로 기록된다 — 1.3s는 평균 1.795s의 **0.72배**다.

**3.5초의 여유는 최대 대비 1.29배다.** 이 문서는 그것을 여유가 크다고 적지 않는다.
- 최대 2.721 s는 **첫 표본**이었고 `time_appconnect`가 1.522 s였다 — TLS 왕복이 튄 냉연결이다.
- §4가 창 안의 publish 유발 줄 **39건 중 18건(46%)** 을 냉으로 판정했다.
  **냉이 예외가 아니므로 이 최댓값은 꼬리가 아니라 정상 범위 안이다.**

#### 9.1.1 이 측정이 재지 못하는 것 — 먼저 쓴다

- **topic POST가 아니라 homepage GET이다.** 경계짓는 것은 네트워크 경로이고
  publish 핸들러의 처리 시간이 아니다. 진짜 POST는 그보다 **길 수 있다.**
- 진짜 POST를 재려면 사용자 topic으로 **실제 알림이 발송된다**(휴대폰에 뜬다).
  그것은 사람의 승인 없이 하지 않는다. 절차는 tasks 9.8.
- 표본 20건은 분포가 아니다. 한 시점·한 네트워크에서 잰 것이다.
- **그러므로 3.5초는 측정이 아니라 판단이다.** 측정이 하는 일은 1.3초를 **기각**한
  것이고, 3.5초를 **확정**한 것이 아니다.

### 9.2 누적 체류 재측정 — 관측 5회

프로브를 `internal/obs`에 `-overlay`로 넣고 세 구성을 같은 지연 서버에 걸었다.
같은 조건이 관측마다 다시 관측되는 **하나의 에피소드**를 5회 돌린다.

```bash
go test -overlay=overlay.json -run TestR14Cumulative -v -count=1 ./internal/obs/
```

| 서버 지연 | 오늘 (3회·10s·2s) | 13판 (1회·1.3s) | **14판 (1회·3.5s)** |
|---|---|---|---|
| 500 ms | 0.51 s · 발송 1 · DELIVERED · 래치 0 | 0.51 s · 발송 1 · DELIVERED · 래치 0 | **0.51 s · 발송 1 · DELIVERED · 래치 0** |
| **1.9 s** | 1.91 s · 발송 1 · DELIVERED · 래치 0 | **6.55 s · 발송 5 · PENDING · 래치 1** | **1.91 s · 발송 1 · DELIVERED · 래치 0** |
| **2.125 s** | 2.14 s · 발송 1 · DELIVERED · 래치 0 | **6.55 s · 발송 5 · PENDING · 래치 1** | **2.14 s · 발송 1 · DELIVERED · 래치 0** |
| 4.0 s | 4.02 s · 발송 1 · DELIVERED · 래치 0 | 6.56 s · 발송 5 · PENDING · 래치 1 | **17.57 s · 발송 5 · PENDING · 래치 1** |

**읽는 법 셋.**

1. **실측 분포 안(≤2.721 s)에서 14판은 오늘과 같은 수를 낸다.** 발송 1회, DELIVERED,
   래치 없음. 13판은 같은 구간에서 발송 5회·PENDING·래치였다.
2. **역행은 없어지지 않고 [3.5 s, 10 s]로 옮겨간다.** 4.0 s 행이 그 구간이고,
   그 안에서 14판은 **13판보다 비싸다**(17.57 s vs 6.56 s) — 1회 상한이 클수록
   실패 한 번의 값이 크기 때문이다.
3. **누적의 기전은 `claimOwed`다.** 성공한 행은 `RemindAfter` 1시간이 이후 관측을
   억제하는데, PENDING 행은 창 없이 곧바로 다시 owed다(`outbox.go:277-279`,
   `case AlertPending: return true, false`). 그래서 실패는 **관측마다** 값을 낸다.

**교환의 요지**: 좁아진 구간이 실측 분포 **밖**이고, 넓었던 구간은 실측 분포를
**관통했다**. 구간의 폭이 아니라 측정된 지연이 그 안에 있느냐가 판정 기준이다.
그리고 위 9.1.1이 적듯 그 판정은 **표본 20건짜리 판단**이다.
