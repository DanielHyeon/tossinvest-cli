# Function Logic Map: `ExitObserver.alert`

- Source: `internal/app/engine/exitloop.go` (1600-1607)
- AST evidence: `ast.json` — branches 2, returns 1, calls 2, assignments 1,
  **defers 0, go_statements 0**
- Risk scan: `risk-pattern-report.md`

이 함수는 exit 관측 루프가 운영자에게 말하는 **여러 자리 중 하나**다. `o.alert`를 부르는
자리는 **7곳**이고(`exitloop.go` 6 + `exit_quarantine_announce.go:71`), 그것이 전부가 아니다.

> **정정**: 이 파일의 첫 판은 "유일한 자리 / 6종"이라고 썼다. 두 군데가 틀렸다.
> (1) 열거가 `exitloop.go` **파일 안**의 `o.alert`만 셌다 —
> `exit_quarantine_announce.go:71`의 일곱 번째가 빠졌다.
> (2) 더 중요하게, **`o.alert`는 열거의 올바른 단위가 아니다.**
> `checkOutage`(`:767-804`)는 같은 사이클에서 `o.alert` `:780`에 더해
> `EscalateOperatingMode(..., o.opts.Announcer)` `:796`으로 **두 번째**로
> `obs.Notifier.Notify`에 도달한다(`journal/operating_mode.go:479` →
> `obs/mode.go:57`). 올바른 단위는 호출자도 파일도 아니라 **`Notify`에 도달하는 경로**다.
> 전수는 `analysis/notify-reach.md`에 있다.

## Inputs and invariants

| Input/state | Valid range | Source of truth | Failure behavior |
|---|---|---|---|
| `ctx` | 관측 루프의 컨텍스트 | `Run:353-363` → `ObserveOnce(ctx)` | **그대로 전달된다** — 이 함수는 기한을 새로 만들지 않는다(AST calls 2: `Notify`와 `logErr`뿐, `context.*` 호출 없음) |
| `e` | `obs.Event` | 호출자 **7곳** (`exitloop.go` `:780`·`:1430`·`:1500`·`:1526`·`:1550`·`:1580`, `exit_quarantine_announce.go:71`) | 등급은 여기서 정해지지 않는다 — `obs.SeverityOf(e.Type)`가 정한다 |
| `o.opts.Alerts` | nil 허용 | `ExitObserverOptions.Alerts` (`exitloop.go:180`), 프로덕션은 `exitwiring.go:341-343`이 `*obs.Notifier` 주입 | nil이면 B1이 조용히 반환 |

## Branches and early returns

| Branch | Condition | Mutation/side effect | Return | 네트워크 접촉 |
|---|---|---|---|---|
| B1 `:1601` | `o.opts.Alerts == nil` | 없음 | `return` `:1602` | ❌ |
| B2 `:1604` | `Notify`가 오류를 돌려줌 | `o.logErr(e.Type, err, "the alert could not be made durable")` `:1605` | 암묵 `:1607` | — |

**이탈은 하나(B1)뿐이다.** `Alerts`가 배선돼 있으면 `Notify`는 **반드시** 호출되고,
이 함수는 그것이 돌아올 때까지 **반환하지 않는다**.

## Calls and live bindings

§0.3 — 이 함수가 호출자를 붙잡아 두는 시간의 예산:

| Callee | Why called | **Error/timeout/retry contract** | Evidence |
|---|---|---|---|
| `o.opts.Alerts.Notify` `:1604` | 운영자 통지 | **동기·무기한.** 이 함수는 기한을 씌우지 않는다(AST calls에 `context.WithTimeout` 없음). 실제 상한은 피호출자가 정한다 — normal **10s**, critical **3×10s + 2×2s = 34s**(아래 유도) | AST calls + `exitwiring.go:341-343` |
| `o.logErr` `:1605` | 실패 기록 | 네트워크 없음, 즉시 반환 | AST calls |

### 상한 34초의 유도 — 전부 AST 산출물에서

| 단계 | 근거 |
|---|---|
| `Notify` B1 `:111`이 등급으로 갈린다 | `internal-obs--notifier.notify/ast.json` |
| normal → `publishBestEffort` `:112`, publish **1회**(루프 없음) | `internal-obs--notifier.publishbesteffort/ast.json` — branches 2, 둘 다 `if` |
| critical → `notifyCritical` `:115` → `deliver` `:182` | `internal-obs--notifier.notifycritical/ast.json` |
| `deliver` B2 `:251` `for attempt := 1; attempt <= attempts` | `internal-obs--notifier.deliver/ast.json` |
| `attempts` = `DefaultCriticalAttempts` = **3** (B1 `:245`가 0 이하를 대체) | `notifier.go:45`. 프로덕션 `newNotifier`(`exitwiring.go:73-81`)는 `Attempts`를 채우지 않는다 |
| 시도 사이 대기 **2회** (B7 `:267` `attempt < attempts`) × `DefaultRetryDelay` **2s** | `notifier.go:48`. `newNotifier`는 `RetryDelay`도 채우지 않는다 |
| publish 1회 상한 **10s** | `internal-obs--ntfy.publish/ast.json` B3 `:96` `timeout <= 0 → 10s`, B7 `:122` `client == nil → &http.Client{Timeout: timeout}`. 프로덕션 `&obs.Ntfy{BaseURL, Topic, Token}`(`notifications.go:101`)는 `Timeout`·`HTTPClient`를 다 비운다 |

**실측** — 전문은 `analysis/delivery-latency.md`. 요약: **publisher를 가진 프로세스가
시작한 뒤**(`engine.log` line 6866 이후) `engine.log`에서 동기 체류를 측정할 수 있는
표본은 **6건**이고

> **창은 커밋이 아니라 프로세스 재시작이다** — 7라운드 H4. 이 산출물은 7판까지
> "publisher 배선 커밋 `e540668f`(2026-08-04) 이후"라고 적었는데,
> `delivery-latency.md` §0이 그 정의를 "**틀렸다**"고 폐기했다. 커밋이 존재하는 것과
> **돌던 프로세스가 그것을 갖고 있는 것**은 다르고, 커밋 다음 날 00:11의 로그가
> `no notification publisher is configured`다. 커밋 기준 창은 정본보다 **약 24.5시간
> 이르다.** **값은 하나도 안 틀렸다**(표본 6건·범위 동일) — 틀린 것은 문장뿐이라
> 값 단위 검사(`tools/check_values.py`)가 못 잡는 종류다.


**전부 exit 관측 루프가 생산했다**(`exit.proposal_refused` 4 · `exit.judgement_refused` 2 —
두 종류 다 프로덕션 생산자가 1곳뿐임을 grep으로 확정). 범위 **0.198 ~ 0.754초**.

**중요**: 이 6건은 전부 publish가 **성공**한 경우다. 배선 이후 publish가 실패한 기록은
`engine.log`에 **한 건도 없다**. 따라서 34초는 **유도된 상한이지 관측된 값이 아니다.**
정상 등급의 유효한 표본은 현재 로그에서 얻을 수 없다(같은 이벤트 종류에 생산자가 4곳이라
줄 간격이 어느 루프의 체류도 재지 못한다).

관측 주기는 **5초**(`DefaultExitObservationInterval`, `exitloop.go:97`)이고,
`Run`은 `ObserveOnce` **뒤에** 그 5초를 잔다(`Run` AST calls `:358`→`:359`) —
즉 사이클 체류는 주기에 **더해진다**.

## State mutations and fallbacks

- **이 함수는 아무 상태도 변경하지 않는다.** AST assignments 1은 B2의 `err :=` 하나뿐이다.
- fallback 없음. `Notify` 실패는 로그 한 줄이고 호출자는 그것을 알 수 없다(결과 0개).
- **goroutine 없음** — AST `go_statements: 0`. 발송은 호출자의 goroutine에서 일어난다.

## Safety conclusion

- **Safe edit boundary**: 이 함수는 판정도 주문도 하지 않는다. 안전 성질은 **호출자를
  붙잡는 시간** 하나다. 그 시간을 줄이는 편집은 §0.3 방향이 옳고 늘리는 편집은 반대다.
- **High-risk impact**: **yes** — `ObserveOnce`(`:412-467`) 안에서 동기로 불린다.
  포지션 순회는 순차이므로(`ObserveOnce` B5 `:453` range) 한 포지션의 알림 체류가
  **같은 사이클 뒤쪽 포지션들의 판정·제출을 그만큼 민다**.
- **같은 저장소가 이미 이 판단을 한 번 내렸다**: `Runtime.alert`(`runtime.go:444-456`)는
  `context.WithTimeout(context.WithoutCancel(ctx), alertDeliveryBound)` = **30초**로
  감싸고 defer로 해제한다(AST defers **1**, calls 7). 그 상수의 주석은 "finite so a dead
  transport cannot hold the shutdown open"(`runtime.go:458-461`)이다. **종료 경로에 씌운
  기한을 관측 경로에는 씌우지 않았다** — 이 함수의 AST defers는 **0**이다.
- `ReconcileDriver.alert`(`reconcileloop.go:552-560`)도 defers **0**으로 같은 상태다.
- **그러나 기한을 이 함수에 씌우는 것은 틀린 자리다.** 같은 사이클의 두 번째 `Notify`
  도달 경로(`checkOutage:796` → `EscalateOperatingMode`)는 이 함수를 지나지 않으므로,
  여기에만 기한을 씌우면 **그 경로가 그대로 남는다.**
  **그리고 저기에 씌우는 것도 "트랜잭션이 잘려서"가 아니다** — 이 산출물은 6판까지
  "그 `ctx`는 journal 트랜잭션과 공유된다"고 적었는데, 그 문장이 근거로 쓰이는 방향이
  틀렸다. `TransitionOperatingMode` AST가 `tx.Commit()` **B25 `:468`** →
  `AnnounceOperatingMode` **B27/B28 `:478-479`**를 열거한다. **announce는 커밋 뒤이고
  트랜잭션 밖이다.** 호출자 기한의 실제 결과는 잘림이 아니라 **`BeginTx`(`:391`) 전에
  만료돼 트랜잭션이 시작조차 안 되는 것**이고, 그러면 운영 모드 승격 자체가 사라진다
  (`…--journal.transitionoperatingmode` FLM · design D1 안 B).
  그래서 예산은 **`Notify` 아래**에서 정해져야 7경로 전부에 걸리고 원장에는 안 걸린다.

  > 2라운드 M5의 정정이 `design.md`에만 착지하고 이 산출물과 `checkoutage` 산출물에는
  > 오지 않았다. 6라운드가 `checkoutage` 쪽을 차단 B3로 잡았고, **7판이 값 단위로 훑어
  > 여기 있는 두 번째 사본을 찾았다** — 리뷰가 가리킨 자리만 고쳤다면 이 사본은
  > 일곱 판째 살아남았을 것이다.
