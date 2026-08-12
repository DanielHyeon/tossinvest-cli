# a092 설계 — 전송을 루프 밖으로 옮긴다 (19판)

> **이 문서가 상수 유도의 정본이다.** proposal은 7판부터 수치를 복제하지 않고 여기를
> 가리킨다 — 6라운드 차단 2건이 그 복제에서 나왔다. 값이 문서마다 같은지는
> `tools/check_values.py`가 검사한다.
>
> **17판은 재설계다.** 1~16판은 전송을 루프 **안에** 두고 예산으로 묶으려 했다.
> 16라운드가 그 전제를 깼다: 예산은 전송에만 기한을 씌우고 나머지에는 못 씌우므로
> "체류 상한"이 아니고, 시도를 1회로 줄여야만 예산에 들어가는데 그 1회가 a096의
> 3회 계약을 거짓으로 만든다. 두 spec 델타가 **세 번 반복해서** 적은 문장이 답이다 —
> *"상한이 필요하다면 전송이 루프 밖으로 나가야 한다"*(engine-safety:28),
> *"누적을 없애려면 전송이 루프 밖으로 나가야 한다"*(engine-safety:30),
> *"사이클 총합이 유계여야 한다면 전송이 관측 사이클 밖에서 일어나야 한다"*(exit-policy:28).
> 열여섯 판이 그 문장을 **범위 밖**이라고 적었고, 그래서 열여섯 판이 도달하지 못했다.
>
> **D0이 그 재설계이고, D1~D7은 그 앞의 설계다.** D0 끝에 무엇이 무효가 되는지 적는다.

## D0 — 전송을 exit 관측 루프 밖으로 옮긴다 (19판)

### D0.1 — 발견: 루프 밖 전송 기계가 이미 있고 프로덕션이 아무도 부르지 않는다

`Notifier.Flush`가 `internal/obs/notifier.go:427`에 **이미 있다.** 자기 doc comment가
무엇을 기다리는지 직접 말한다:

```go
// Flush retries every pending outbox row.
//
// It is what a supervising loop calls periodically and what an operator triggers
// after fixing the transport. ...
```

그 감독 루프가 없다. 저장소 전체에서 `.Flush(` 호출은 **세 곳이고 전부 테스트다**:
`internal/obs/obs_test.go:440`, `internal/obs/obs_test.go:590`,
`internal/obs/a097_exclusion_is_an_event_test.go:109`. 프로덕션 호출자 **0**.

**이것은 문서의 미배정 후속이 아니라 오늘 살아 있는 결함이다.**
`Gateway.parkAlert`(`internal/execgw/replay.go:534`)는 미확정 주문을 만나면
`g.entry.Block`으로 진입을 잠그고 `g.journal.EnqueueAlert`(:551)로
`order.unresolved_in_doubt`를 outbox에 PENDING으로 넣는다. 그 행을 보내는 코드는
`Flush`뿐이고, `Flush`를 부르는 프로덕션 코드가 없다. 도달 가능성은 배선으로 확인했다:

```text
internal/reconcile/recovery.go:318  r.opts.Replayer.ReplayInDoubt(ctx, attemptID)
  → internal/execgw/replay.go:226   Gateway.ReplayInDoubt
  → internal/execgw/replay.go:358   g.parkAlert(...)
  → internal/execgw/replay.go:551   g.journal.EnqueueAlert(...)   ← 기록됨
                                    ...                            ← 전송 안 됨
```

그리고 `replay.go:101-107`의 프로덕션 주석은 **그 반대를 단언한다**:

```go
// The alert is written
// straight into the journal's outbox, ... — the notifier's Flush picks the row up
// and delivers it — so an engine whose notifier is not wired still records the
// alert rather than losing it.
```

`Flush picks the row up and delivers it`는 오늘 거짓이다. 안전 결과 자체(진입 차단)는
`entry.Block`이 강제하므로 잠기지만, **그 이유가 알림 경로로 사람에게 가지 않는다.**
a092가 감독 루프를 배선하면 이 주석이 참이 된다. 17판이 그것을 고친다.

### D0.2 — 재설계: claim은 동기, publish는 루프 밖

바뀌는 것은 **어느 goroutine이 무엇을 하느냐** 하나다.

| 단계 | 오늘 (HEAD·1~16판) | 17판 |
|---|---|---|
| 구조화 로그 한 줄 | exit 관측 goroutine | exit 관측 goroutine (그대로) |
| `ClaimAlertForDelivery` (outbox 트랜잭션) | exit 관측 goroutine, `n.mu` 보유 | exit 관측 goroutine, **뮤텍스 없이** (D0.3) |
| `Publisher.Publish` × 시도 | **exit 관측 goroutine**, `n.mu` 보유 | **알림 배달 루프** |
| `MarkAlertDelivered` / `MarkAlertAttemptFailed` | exit 관측 goroutine | 알림 배달 루프 |
| 시도 간 대기 | exit 관측 goroutine (`clk.Sleep`) | 루프 주기가 대신한다 (D0.6) |
| `Gate.Block` (시도 소진) | exit 관측 goroutine | 알림 배달 루프 |
| `escalate` → `EscalateOperatingMode` | exit 관측 goroutine | 알림 배달 루프 |
| `Gate.Block` (claim 자체 실패) | exit 관측 goroutine | exit 관측 goroutine (**그대로** — 행이 아예 없는 최악이고, 비용은 실패한 SQLite 쓰기 한 번이다) |

알림 배달 루프는 `SupervisedLoop`로 등록한다. `NewRuntime`(`runtime.go:183`)이
`Name`·`Run`을 요구하고 `Health`는 **optional**이며(`runtime.go:117-122`) `Health`가 nil이면
`Trigger` 검증도 건너뛴다(`runtime.go:201-203`). 이 루프는 `Health: nil`로 등록한다 —
배달 실패의 등급화는 이미 outbox의 `attempts`와 진입 게이트가 하고 있고, 감독자 임계를
하나 더 두면 "알림이 안 나간다"의 정의가 둘이 된다(`runtime.go:118-121`이 exit observer에
대해 같은 판단을 이미 적었다).

### D0.3 — `n.mu`가 지키던 것을 다시 읽는다 — publish가 나가면 배제가 필요 없어진다

**AST가 먼저 이 지점을 위험으로 지목했다.** `analysis/function-logic/internal-obs--notifier.flush/ast.json`의
열거: 분기 6 · 반환 4 · 호출 9 · defer 1 — `n.mu.Lock` + `defer n.mu.Unlock` · 그 안에서 PENDING 행마다
`Publisher.Publish`. 즉 **감독 루프를 그냥 배선하면 배달이 `n.mu`를 행 수만큼의 왕복 동안
쥐고, exit 관측의 claim이 그 뒤에 줄을 선다.** a092가 없애려는 결함이 자리만 옮긴다.
그러므로 배선만으로는 안 되고 뮤텍스의 범위를 다시 정해야 한다.

두 주석이 그 뮤텍스가 무엇을 위한 것인지 말한다.

- `notifier.go:112-115` — *"mu serialises the whole claim-and-send, not just the send.
  Claiming outside it would let two observations of one condition both read a row that
  has not been delivered yet and **both publish**"*
- `outbox.go:165-168` — *"Exclusion against a concurrent claimer is the caller's:
  obs.Notifier holds its delivery mutex across the claim and the send."*

곧 `n.mu`가 막는 결과는 **두 번 보내는 것**이다. 17판에서는 claim 경로가 아예 보내지
않으므로, 두 claimer가 동시에 `owed=true`를 받아도 아무 일도 일어나지 않는다.

남는 물음은 **claim과 배달 루프가 겹칠 때 행이 상하느냐**이고, 답은 SQL에 있다.

| 근거 | 자리 | 내용 |
|---|---|---|
| PENDING 행은 claim이 덮어쓰지 않는다 | `outbox.go:276-278` | `case AlertPending: return true, false` — `rearm=false`이므로 UPDATE 자체가 실행되지 않는다 |
| 재무장은 DELIVERED·ACKNOWLEDGED·미지 상태에서만 | `outbox.go:279-314` | 그 상태의 행은 `PendingAlerts`의 `WHERE state = PENDING`(`outbox.go:393`)에 애초에 들어오지 않는다 |
| 배달 확정은 CAS다 | `outbox.go:342` | `WHERE id = ? AND state = ?` (PENDING) — 사이에 상태가 변하면 0행이 되고 `requireOneRow`가 오류를 낸다 |
| 실패 기록도 CAS다 | `outbox.go:357` | 같은 술어 |
| 운영자 해제도 CAS다 | `outbox.go:378` | 같은 술어 |

~~그러므로 **claim 경로에는 뮤텍스가 필요하지 않다.**~~

> ## ⚠ 18판이 이 결론을 철회한다 — 17라운드 A-P1
>
> 위 표는 **행 하나의 갱신**에 대해서는 옳다. 그러나 17판이 그것으로 뮤텍스를
> 없앨 수 있다고 결론지은 것은 틀렸다. `Acknowledge`가 지키는 것은 행 하나가
> 아니라 **두 연산 사이의 창**이기 때문이다:
>
> ```text
> notifier.go:506   remaining, err := n.Journal.UndeliveredCount(ctx)   ← DB 읽기
>                   ← 이 사이에 PENDING 행이 하나라도 생기면 아래 줄이 틀린다
> notifier.go:510   if remaining == 0 && n.Gate != nil {
> notifier.go:511       n.Gate.Clear(execgw.ReasonAlertUndelivered)     ← 인메모리 쓰기
> ```
>
> **CAS의 범위 밖이다.** CAS는 `WHERE id = ? AND state = ?`로 한 행을 지키지,
> DB 읽기와 인메모리 쓰기 사이를 지키지 않는다. 그리고 그 창을 오늘 닫고 있는
> 것은 **재무장하는 쪽이 같은 뮤텍스를 기다린다**는 사실뿐이다
> (`notifier.go:470-475`가 그렇게 적었다).
>
> **17판의 다리 (d) — "`Acknowledge`는 배달 뮤텍스를 계속 쥔다" — 는 참이지만
> 근거가 되지 못한다.** 잠금은 그것을 잡는 쪽만 배제한다. claim이 놓으면
> `Acknowledge`가 쥐고 있는 것은 상대가 없어진 뒤의 잠금이다.
>
> 창을 여는 사건은 재무장만이 아니다. **새 `event_key`의 INSERT도 PENDING 행을
> 만든다**(`outbox.go:245`). 즉 exit 사이클이 알림을 하나 올리는 것만으로도
> 창이 열린다.

### D0.3a — 18판의 결정: 잠금을 남기되 **원격 전송 위에서는 잡지 않는다**

**사용자 결정 (2026-08-10)**: 세 안 중 **안 1**.

`n.mu`는 기록 경로에 **그대로 남는다.** 17판이 바꾸려던 것은 잠금의 존재였고,
18판이 바꾸는 것은 **잠금이 무엇을 덮느냐**다.

| | 잠금이 덮는 구간 | exit 사이클의 최대 대기 |
|---|---|---|
| HEAD | claim → **publish × 시도 × 대기** → 정산 | 기한 없음 (최대 34초) |
| 17판(철회) | 잠금 없음 | 0 — **그러나 게이트가 잘못 열린다** |
| **18판** | **로컬 SQLite 작업만** | 로컬 DB 연산 하나 |

배달 루프의 한 주기는 이렇게 나뉜다:

```text
n.mu.Lock()    → PendingAlerts(ctx, alertFlushBatch)          로컬
n.mu.Unlock()
               → Publisher.Publish(...)  행마다, 잠금 없이     원격  ← 여기서 절대 안 잡는다
n.mu.Lock()    → MarkAlertDelivered / MarkAlertAttemptFailed   로컬
               → Gate.Block / escalate (필요하면)
n.mu.Unlock()
```

**이 결정이 되돌려 주는 것.**

- a096의 배제 논증이 **한 줄도 안 바뀐다.** `notifier.go:112-115`의 주석과
  `outbox.go:165-168`의 계약이 그대로 참이다.
- 위 SQL 표는 **증명 부담에서 참고 자료로 내려간다.** D0.3이 그것으로 무엇을
  증명할 필요가 없어졌다.
- 보존하기로 한 기존 테스트 셋(`tasks.md` §6.0.1) 중 **둘이 유효하다.**
  17라운드 A-P6=B-P3이 지적한 "구조적으로 통과 불가"는 잠금이 사라진다는
  전제 위에 있었고, 그 전제가 철회됐다.

  > **셋째는 유효하지 않다 (18라운드 A-P1 · 19판이 철회).** 18판은 여기 *"셋 다
  > 유효하다"*고 적었다. 위 순서도가 **publish 위에서 잠금을 놓으므로**,
  > `TestAcknowledgeCannotClearTheGateMidSend`가 성질로 단언하는 전제 —
  > *"발송이 `Publish` 안에 멈춰 있는 동안 그 행은 여전히 PENDING"* — 이
  > 성립하지 않는다. 잠금이 풀린 구간에 `Acknowledge`가 들어오면 그 행은
  > 정착하고, 뒤이은 정산 CAS(`outbox.go:342-343`)는 0행을 갱신한다.

### D0.3b — 19판의 결정: 승인이 발송을 이긴다

**사용자 결정 (2026-08-10, 결정 2 = 안 1).** 정산 CAS가 실패하면 그것은 오류가
아니라 **운영자가 먼저 정산했다**는 뜻이다.

```text
n.mu.Lock()    → ClaimAlertForDelivery                          로컬
n.mu.Unlock()
               → Publisher.Publish(...)                          원격
                     ↑ 이 구간에 Acknowledge가 들어올 수 있다 ── 허용된다
n.mu.Lock()    → MarkAlertDelivered(id, PENDING)                 로컬
                     └ 0행 → ErrAlertNotFound
                         → 오류 아님. "승인이 먼저 왔다"로 기록하고 다음 행으로.
                         → 행을 덮지 않는다. 배치를 중단하지 않는다.
n.mu.Unlock()
```

| | 18판(철회) | 19판 |
|---|---|---|
| 발송 중 `Acknowledge` | 일어날 수 없다 (거짓) | **일어난다** |
| 정산 CAS 실패 | 오류 · 배치 중단 | **경합의 정상 결과** · 배치 계속 |
| 게이트 | 승인이 못 푼다 | **승인이 푼다** |

**왜 승인이 이기는가.** 래치의 사유는 `ReasonAlertUndelivered` — *알림이 사람에게
닿지 않았다*이다. 승인은 사람이 그것을 봤다는 선언이므로 **사유 자체가 소멸한다.**
그 뒤 전송이 실패해도 사유는 되살아나지 않는다: 사람이 이미 아는 사건을 전송
실패가 다시 차단으로 만들면 승인이 의미를 잃는다.

**왜 이것이 "고쳐서 통과시키기"가 아닌가.** 옛 성질은 **경합을 부인**했고
(`Acknowledge`는 발송 중에 들어올 수 없다), 새 성질은 **경합의 승자를 정한다**.
원래 테스트가 막던 사고 — 승인이 조용히 사라지는 것 — 은 새 성질이 더 강하게
막는다: 뒤늦은 정산은 승인을 덮지 못하고, 덮지 않았다는 사실이 기록된다.
R17-13이 `rearm` 경합에 이미 같은 규칙을 쓴다.

**대가.** 그 테스트는 **다시 써야 한다.** 18판의 "고치지 않고 초록" 주장은
그 한 행에 대해 철회한다. 새 RED는 **R19-1**이고 §6.0에 있다.

**이 결정의 대가.**

- **engine-safety 델타의 SHALL NOT을 정밀화해야 한다.** 17판은
  *"배달 잠금을 기록 경로에서 잡아서는 안 된다"*고 썼는데, 금지 대상이
  "잠금"이 아니라 **"원격 왕복 위의 잠금"**이어야 했다. 철회가 아니라 교정이다 —
  막으려던 것(결함이 자리만 옮기는 것)은 그대로 막힌다.
- **exit 사이클은 배달 루프의 로컬 작업 뒤에서 기다릴 수 있다.** 배치 8행 정산이면
  SQLite 8회. 밀리초 단위이고 기한은 없다 — 원장이 멈추면 관측도 멈춘다는
  D0.4의 문장이 여기에도 그대로 적용된다.
- **`Acknowledge`가 남는 무계항이다.** `ids` 없이 부르면 잠금 아래에서
  `PendingAlerts(ctx, 0)` + 행마다 `AcknowledgeAlert` + 카운트를 돈다 — O(미전달 행 수).
  운영자가 직접 부르는 드문 경로이므로 받아들이되, **적어 둔다.**

> **§6의 관측 부담이 바뀐다.** R17-3(claim × 배달 인터리빙)은 여전히 돌지만
> 증명 대상이 아니라 회귀 방지가 된다. 대신 **R17-2가 이 판의 중심 RED가 된다**:
> 배달 루프가 `Publish` 위에서 `n.mu`를 잡지 않는다는 것. 그것이 깨지면
> 18판은 HEAD와 같아진다.

부수 효과 하나: `Flush`는 보낼 내용을 **행에서** 만들고(`notifier.go:445-450`),
`deliver`는 살아 있는 이벤트에서 만든다(`notifier.go:346`). 같은 event key의 두 번째
에피소드는 이유가 다를 수 있으므로(`outbox.go:210-216`), 행에서 만드는 쪽이 사건 후
outbox를 읽는 운영자가 보는 것과 일치한다. 17판은 그 쪽으로 통일된다.

### D0.3c — a099가 착지했다. **D0.3의 표를 다시 읽어야 한다** (2026-08-12, a099 §7.1)

a099(`757550f1`)가 `alert_outbox`에 배달 임차를 넣었다. **D0.3~D0.3b가 인용하는
좌표와 술어의 절반이 그때 바뀌었다.** 여기서 그것을 대조한다 — 위 절들은
**안 고친다**. 고치면 왜 그 결론에 이르렀는지가 사라지고, 남는 것은
「처음부터 맞았던 것처럼 보이는 문서」다.

**1. D0.3의 표 다섯 줄 중 셋이 술어가 늘었다.**

| D0.3의 줄 | a099 이후 |
|---|---|
| 배달 확정 CAS `WHERE id = ? AND state = ?` | **`… AND claim_token = ?`가 붙는다** (`outbox.go:449`) |
| 실패 기록도 같은 술어 | 같음 (`outbox.go:467`) — **성공해도 임차를 안 푼다** |
| 운영자 해제도 같은 술어 | **안 바뀌었다.** `AcknowledgeAlert`는 토큰을 안 받고 임차를 안 본다 — 사람의 확인은 임차 위에 있다 |
| 0행 → `requireOneRow`가 **오류** | **오류가 아니다.** 넷으로 갈린다 — `SettleApplied`·`SettleLeaseLost`·`SettleAlreadySettled`·`SettleNotFound` |
| PENDING 행은 claim이 덮어쓰지 않는다 | **그대로다.** `claimOwed`가 PENDING에 `rearm=false`를 준다 |

**2. D0.3의 결론 문장이 반만 맞았다.**

*"claim 경로에는 뮤텍스가 필요하지 않다"*는 18판이 이미 철회했다. 그러나
철회 이유(`Acknowledge`의 읽기-쓰기 창)와 **별개의 이유**가 하나 더 있었고,
a099가 그것을 찾았다:

> **정산 CAS는 publish가 끝난 뒤에 돈다.** 두 관측이 같은 PENDING 행에 대해
> 각각 `owed=true`를 받으면 **둘 다 보내고**, 두 번째의 CAS만 실패한다.
> **CAS가 막는 것은 이중 정산이지 이중 발송이 아니다.**
> 2026-08-08의 `no such alert` 줄이 그 네 번째 단계였다.

그러므로 D0.3의 표는 **「행이 상하지 않는다」의 증거로는 옳고**,
**「두 번 보내지 않는다」의 증거로는 옳지 않았다.**
후자를 지금 지는 것은 `alert_claim.go`의 `acquireAlertClaimTx`이고,
증거는 `a099-…/analysis/function-logic/internal-journal--acquirealertclaimtx/`다.

**3. D0.3b의 결정은 a099에서 이미 구현되었다.**

「승인이 발송을 이긴다」의 다이어그램이 코드가 됐다. 다만 형태가 다르다:

| D0.3b가 적은 것 | 착지한 것 |
|---|---|
| `MarkAlertDelivered(id, PENDING)` | `MarkAlertDelivered(ctx, id, token) (SettleResult, error)` |
| 0행 → `ErrAlertNotFound` → *"오류 아님"*으로 **호출자가 해석** | **원장이 이름을 준다** — `SettleAlreadySettled` |
| 배치를 중단하지 않는다 | `Flush`가 `continue`하고 `deliver`가 `lost=true`로 나간다 |
| 게이트: 승인이 푼다 | **그대로다.** 경합·임차 상실은 게이트를 **양방향으로 안 건드린다** |

**해석을 호출자에게 맡기지 않은 것이 a099의 선택이다.** 「0행」이라는 하나의
사실에 네 원인이 있고, 그 넷에 호출자가 다르게 반응해야 한다.

**4. R17-3은 이제 회귀 방지가 아니라 「a099 뒤에」 참이다.**

D0.3b의 각주가 *"R17-3은 여전히 돌지만 증명 대상이 아니라 회귀 방지가 된다"*고
적었다. **지금은 그것도 정확하지 않다.** R17-3이 재려는 「claim × 배달 인터리빙이
행을 안 상하게 한다」는 **a099의 임차 없이는 거짓**이다 — 행은 안 상하지만
푸시가 둘 나간다. a092가 재개할 때 R17-3의 성공 기준을 **발송 횟수**로 다시 써야 한다.

**5. D7(Pre-Edit)의 대상이 늘었다.**

a099가 `internal/obs/notifier.go`의 `claimAndDeliver`(분기 4 → 9)·`deliver`(12 → 24)·
`Flush`(6 → 11)를 바꿨다. **a092의 Pre-Edit 선언은 그 세 함수의 지금 모양을
근거로 다시 써야 한다** — 옛 분기 번호로 적힌 선언은 편집 경계를 안 가리킨다.

### D0.4 — 없애는 것과 못 없애는 것

**없앤다.**

- transport가 exit 관측 사이클을 붙잡는 것. 루프에 남는 알림 비용은 로그 한 줄 + outbox 트랜잭션이다.
- `n.mu` **배수**(engine-safety:34). 관측 goroutine은 그 뮤텍스를 계속 쥐지만(D0.3a), 그 아래에 원격 왕복이 없으므로 **전송이 끝날 때까지 기다리는 배수**가 없어진다. 없어지는 것은 잠금이 아니라 **잠금 아래의 네트워크**다.
- 사이클 총합의 무계성(exit-policy:28·50). 알림 수에 비례하는 항이 로컬 쓰기가 된다.
- `alertPublishAttempts = 1`의 **근거**. 재시도가 손절을 붙잡지 않으므로 줄일 이유가 사라진다.

**못 없앤다 — 그리고 이것을 상한이라고 부르지 않는다.**

- outbox 트랜잭션과 구조화 로그에는 여전히 기한이 없다. 원장이 멈추면 관측도 멈춘다.
- 그러므로 두 spec 델타의 *"이 요구가 만드는 것은 예산이지 보장이 아니다"*는 **그대로 산다.**
  17판이 바꾸는 것은 그 예산 안에 무엇이 들어 있느냐이지, 예산이 보장이 되는 것이 아니다.
- 크기는 달라진다: 기한 없는 항이 `네트워크 왕복 × 시도 + 대기`에서 `로컬 SQLite 쓰기 하나`가 된다.

### D0.5 — 대가: 진입 게이트 래치가 늦어진다. 손절은 아니다

게이트가 막는 것은 **진입뿐이다.** `notifier.go:255`가 그 결정을 이미 적어 두었다 —
*"Entries only. Exits are untouched: no alert failure may slow a stop."* 그러므로 래치가
늦어져서 늦어지는 것은 **신규 진입 차단**이고, 손절이 아니다. 손절 경로는 빨라진다.

| | 관측 시점부터 게이트 래치까지 | exit 사이클이 붙잡히는 시간 |
|---|---|---|
| HEAD (obs 기본값 3회·2초) | 3 × (기한 없는 publish) + 2 × 2s | 같은 값 |
| 1~16판 채택 #9 (1회·3.5s) | 3.5s | 3.5s |
| **17판** | 다음 틱까지(≤2s) + 3 × 3.5s + 2 × 2s ≈ **16.5s** | **로컬 쓰기 하나** |
| **18판 (큐 없음)** | 같음 ≈ **16.5s** | 로컬 DB 연산 하나 |
| **18판 (큐 최악)** | 위 + **선행 배치의 서비스 시간** — 아래 참조 | 같음 |

**이 표가 대가다.** 진입 차단이 채택 #9보다 약 13초 늦다. 그 13초 동안 이 계좌는
신규 진입을 할 수 있고, 그 사이 미전달 critical 알림이 outbox에 있다.

> ## ⚠ 18판 정정 — 16.5초는 상한이 아니다 (17라운드 A-P3)
>
> 위 식은 **큐 대기 시간을 뺐다.** 알림이 도착했을 때 앞에 밀린 행이 있으면
> 첫 시도 전에 이미 기다린다. 한 주기의 최악 서비스 시간은
> `alertFlushBatch × alertPublishTimeout = 8 × 3.5s = 28s`이고,
> 그것이 **첫 시도 앞에 통째로 들어올 수 있다.**
>
> 그리고 더 나쁜 것이 있다. `MarkAlertAttemptFailed`는 행을 **PENDING으로 남기고**
> (`outbox.go:356-358`), `PendingAlerts`는 항상 `ORDER BY id` — **오래된 것
> 먼저**다(`outbox.go:393`). 시도를 다 쓴 오래된 실패가 배치 앞자리를 계속
> 차지하면 **새 critical 알림이 굶는다.** 17판의 spec에는 소진된 행을 빼는 규칙도
> 공정성 규칙도 없었다.
>
> **18판이 더하는 규칙 (spec에 SHALL로 내려간다):**
>
> 1. 한 주기의 배치는 `attempts < alertPublishAttempts`인 행을 **먼저** 고른다.
>    시도를 다 쓴 행은 그 주기의 잔여 자리에만 들어간다.
> 2. 소진된 행은 **버려지지 않는다** — PENDING으로 남고 게이트 래치도 유지된다.
>    빠지는 것은 **배치 우선순위**뿐이다. (버리면 내구성 완화이고 안전 불변식 위반이다.)
> 3. 그러므로 상한은 `≤2s + 28s(선행 배치) + 3 × 3.5s + 2 × 2s ≈ **44.5s**`이고,
>    **이것이 이 change가 적는 진입 게이트 래치의 최악값이다.**
>
> 44.5초를 그대로 적는다. 16.5초를 상한이라고 부르면 그것이 다음 편집이
> 인용하는 수가 된다.

> ## ⚠⚠ 19판 정정 — **44.5초도 상한이 아니다** (18라운드 A-P5 = B-P8)
>
> 18판은 16.5초를 고치면서 **같은 형태의 실수를 한 번 더 했다.** 위 3번은
> 선행 배치를 **정확히 하나**로 놓는다. 그것을 정당화하는 것은 아무것도 없다.
>
> **먼저 잰 것 — 오늘의 기계에는 배치가 없다.**
>
> | 잰 것 | 값 | 자리 |
> |---|---|---|
> | `Flush`가 읽는 행 수 | `PendingAlerts(ctx, **0**)` — **전부** | `notifier.go:437` |
> | `PendingAlerts`의 `limit <= 0` 처리 | `LIMIT` 절을 **안 붙인다** | `outbox.go:395-398` |
> | `Flush`가 `n.mu`를 쥐는 구간 | **전체 배수 루프** | `notifier.go:434-435` (`defer`) |
> | 프로덕션 소스의 `alertFlushBatch` | **없다** | 상수는 a092가 **신설할 것** |
>
> `alertFlushBatch = 8`은 현재 기계의 성질이 아니라 **a092가 도입하려는 성질**이다.
> 그것을 이미 있는 것처럼 넣어 28초를 유도했다.
>
> **밀린 행이 N개일 때 참인 식:**
>
> ```text
> 배치가 없는 오늘        : N × alertPublishTimeout          (한 Flush가 n.mu를 쥔 채)
> 배치 B를 도입한 뒤      : ⌈N/B⌉ × B × alertPublishTimeout = N × alertPublishTimeout
> ```
>
> **배치 크기는 이 항을 줄이지 않는다.** 배치는 한 번에 쥐는 시간을 자를 뿐이고,
> 새 알림이 자기 차례를 기다리는 총시간은 **앞에 몇 개가 있느냐**로 정해진다.
> 44.5초는 `N ≤ 8`인 경우의 값이다. `N`에 상한이 없으므로 **상한이 아니다.**
>
> **이것이 드러내는 것은 산술이 아니라 설계 선택이다.** 44.5초를 상한으로 만들려면
> 둘 중 하나를 골라야 하고, 둘은 안전 성질이 다르다.
>
> | | 방법 | 진입 래치 지연 | 대가 |
> |---|---|---|---|
> | **가** | 래치를 **claim 시점**으로 옮긴다 — 발송을 기다리지 않는다 | **≤ 한 틱(2s)**, 백로그와 무관 | 일시적 전송 blip마다 진입이 막힌다. `Acknowledge` 없이는 안 풀린다 |
> | **나** | 백로그 자체에 상한을 준다 (`N > K`면 즉시 래치) | `≤ K × 3.5s` | `K`를 고르는 근거가 아직 없다 |
>
> **a092는 아직 어느 쪽도 고르지 않았다.** 고르기 전까지 이 문서는
> **"진입 게이트 래치에 상한이 없다"**고 적는다 — 44.5초는 `N ≤ 8`의 사례값이고,
> 사례값을 상한이라고 부르는 것이 17·18라운드가 연속으로 잡은 바로 그 형태다.
> **§6.0에 R19-2를 세운다**: 밀린 행 20개 뒤에 도착한 critical 알림의 래치 지연을
> 재고, 그것이 44.5초를 넘는 것을 RED로 관측한다. 그 관측이 선택의 근거가 된다.

> ## ⚠ 19판 정정 — A-P6은 근거가 틀렸다. 공정성 규칙은 **구현 가능하다**
>
> 18라운드 A-P6은 *"공정성 규칙이 `PendingAlerts`의 `LIMIT`만으로는 구현 불가"*라고
> 적었다. 전제가 틀렸다 — **배치가 `LIMIT` 질의에서 나온다고 가정했는데, 오늘의
> `Flush`는 `LIMIT`을 쓰지 않는다.** `PendingAlerts(ctx, 0)`가 PENDING 행을
> **전부** 돌려주므로(`notifier.go:437` → `outbox.go:395-398`), 위 규칙 1의
> *"`attempts < alertPublishAttempts`인 행을 먼저"*는 **읽어 온 슬라이스를 Go에서
> 정렬하는 것**으로 끝난다. `internal/journal`을 한 줄도 안 건드린다 — D7의
> "안 건드리는 것"과 충돌하지 않는다.
>
> `journal.Alert`가 `attempts`를 실어 오는지만 확인하면 되고, 실어 온다 —
> `alertSelect`가 그 열을 고르고(`outbox.go:387`) `Alert.Attempts int`가 받는다
> (`outbox.go:95`). `PendingAlerts`의 선언 주석도 *"A limit of zero means all"*이라고
> 직접 적는다(`outbox.go:390`). **대가는 다른 데 있다**:
> 전부 읽는다는 것은 메모리가 백로그에 비례한다는 뜻이고, 그것은 위 A-P5가
> 고르지 않은 채로 둔 **백로그 상한 문제와 같은 문제**다. 두 지적은 하나다.

받아들이는 근거는 셋이다. (1) 안전 불변식 4는 **손절·비상 청산의 즉시성**이고 그것은
빨라진다. (2) HEAD의 오늘 값은 publish에 기한이 아예 없으므로 17판이 HEAD보다 느리다고
말할 수 없다 — 채택 #9보다 느린 것이지 **현재보다** 느린 것이 아니다. (3) 채택 #9의
3.5s는 시도를 1회로 줄여서 산 값이고, 그 1회가 a096의 계약을 깨는 것이 16라운드 A-3의
차단 사유였다. 13초는 그 차단을 푸는 값이다.

`Publisher`가 nil인 구성은 별도로 다룬다. 오늘 `deliver`는 nil이면 `lastErr`를 세우고
루프를 빠져나와 **래치한다**(`notifier.go:350-352`, :403). `Flush`는 nil이면 `break`하고
아무 시도도 기록하지 않는다(`notifier.go:442-444`). 그대로 배선하면 전송 수단이 없는 엔진은
critical 알림이 쌓여도 **영원히 진입을 막지 않는다.** 17판은 nil publisher를 실패한 시도로
세어 `MarkAlertAttemptFailed`를 남긴다. §6이 이것을 RED로 잡는다.

### D0.6 — 상수: a096의 3회 계약이 문자 그대로 참이 된다

시도 간 대기를 `clk.Sleep`으로 사는 대신 **루프 주기가 대신한다.** 한 사이클에서 한 행은
**한 번** 시도하고, 시도 수는 outbox의 `attempts` 열에 누적된다
(`MarkAlertAttemptFailed`가 `attempts = attempts + 1`, `outbox.go:356`).

```go
// notifications.go — 17판

// alertFlushInterval is the alert delivery loop's period, and it is also the wait
// between two attempts on the same row: one cycle publishes a row at most once.
// It is obs.DefaultRetryDelay's value for that reason and not by coincidence.
const alertFlushInterval = 2 * time.Second

// alertPublishAttempts is three again. a096 landed three, and the only reason the
// 13th revision cut it to one was that the retries were held by the exit loop.
// They are not any more, so the contract stands.
const alertPublishAttempts = 3

// alertPublishTimeout bounds one publish. Unchanged at 3.5s, but its standing has
// changed: it no longer sits on the stop path, so choosing it too low costs a
// wasted cycle rather than a delayed stop.
const alertPublishTimeout = 3500 * time.Millisecond

// alertFlushBatch bounds one cycle's work so a backlog cannot make a cycle
// unbounded. Worst cycle is alertFlushBatch * alertPublishTimeout = 28s.
const alertFlushBatch = 8

// alertLoopShare is what one alert may now cost the exit observation cycle: one
// structured log line and one outbox transaction. No send.
const alertLoopShare = 750 * time.Millisecond
```

- `alertRetryDelay`는 **사라진다.** 그 값을 쓰는 `Notifier.wait`가 claim 경로에서 도달
  불가능해지고, 배달 루프의 대기는 주기다. 16판이 *"never reached ... set anyway"*라고 적은
  상수(design:629-634)는 이제 자리 자체가 없다.
- `alertTransportBudget`·`alertOverheadReserve`·`alertBudget`의 **유도가 무효**가 된다.
  셋 다 "전송이 관측 예산 안에 있다"를 전제로 유도됐다. D3의 컴파일 단언도 같이 바뀐다.
> **⚠⚠ 19판 3차 — `alertFlushInterval`은 더 이상 a092가 정할 값이 아니다.**
>
> 위 블록은 그것을 *"the alert delivery loop's period"*로 선언한다. **19판이
> 배달 루프를 a098로 옮겼고**(D0.10), a098은 `design.md` D1과 `tasks.md` 3.2에
> *"루프 주기는 **재고 정한다.** a092의 값을 인용하지 않는다"*고 적었다.
> 두 change가 같은 상수를 정하면 그것이 **16라운드 B-8이 잡은 형태**이고,
> 이번에는 19판의 분리가 스스로 만들었다.
>
> **경계를 값에 대해 다시 적는다**: a098이 `alertFlushInterval`을 **정하고**,
> a092는 그 값을 **가정으로 인용한다**. a092가 그것을 쓰는 곳은 하나 —
> 진입 게이트 래치가 늦어지는 대가의 산술이고, 그 산술은 이미
> `다음 배달 사이클까지의 시간 + 시도 × 1회 상한 + (시도−1) × 사이클 주기`로
> **주기를 변수로** 적혀 있다(engine-safety 델타). 그러므로 a092는 값을
> 소유하지 않고도 대가를 적을 수 있다. **소유하지 않는다고 적는 것이
> 이 정정의 전부다.**
>
> `tools/check_values.py`의 `ADOPTED_MS`는 아직 `alertFlushInterval: 2000.0`을
> **채택값으로 선언한다.** 집합을 지금 바꾸면 이 값을 인용하는 모든 산문이
> 고아로 뒤집히므로, **바꾸지 않고 이름을 붙여 둔다** → tasks 10.4.5.
> 안 한 것을 안 했다고 적는다.

- `alertLoopShare = 750ms`의 근거는 **보수적 대입**이다. 356.1ms는 claim + 실패 기록 +
  게이트 래치 + 승격 트랜잭션 + 로그 **전부**를 잰 최악이고, 17판의 루프 경로는 그
  **부분집합**(로그 + claim)이다. 부분집합을 상위집합의 실측으로 편성하는 것은 보수적이다.
  750ms는 356.1ms의 2.1배로, 16판이 쓰던 2배 규칙을 그대로 만족한다. **줄어든 경로를 따로
  재지는 않았다** — 그 측정은 §9의 의무로 남는다.

**이것이 16라운드의 여러 건을 뿌리에서 닫는다.** `attempts=1` 논쟁(A-3·B-8)은 상수가
3으로 돌아가면서 사라지고, 4.288s·4.2s·1.3s의 산술(B-2·B-3)은 그것이 유도하던 예산이
없어지면서 사라진다. A-4/ES:32(topic POST 최댓값 미실측)는 **차단에서 기록된 불확실로
내려간다** — 그 값을 낮게 잡아도 이제 손절이 늦지 않고 사이클 하나를 버릴 뿐이다.
그래도 실측 의무 자체는 남는다(§9.8, 사람 승인 필요).

### D0.7 — 17판이 무효로 하는 앞 절

아래는 **삭제하지 않고 기록으로 남긴다** — 어느 라운드가 무엇을 근거로 무엇을 골랐는지가
다음 판의 입력이기 때문이다. 무효의 뜻은 "채택 설계의 근거로 인용할 수 없다"이다.

| 절 | 무효가 되는 이유 |
|---|---|
| D2 전체 (`design.md:46~579`) | 후보 #1~#9 전부가 "전송이 관측 예산 안에 있다"에서 유도됐다 |
| D3의 상수 블록과 컴파일 단언 (`:580~955`) | 유도의 입력이 바뀐다. 단언 집합을 D0.6으로 다시 만든다 |
| D5의 균일성 대가 표 (`:1040~1189`) | 시도 1회의 경로별 대가를 셈한 표다. 시도가 3회로 돌아가면 대가 자체가 없다 |
| `analysis/delivery-latency.md`의 후보 #8 표 | 9.6.1이 세던 고아 51건의 대부분이 여기 산다 |
| 두 spec 델타의 "전송이 루프 밖으로 나가야 한다 — 이 요구는 그것을 정의하지 않는다" 세 문장 | 17판이 그것을 **정의한다**. 델타가 다시 쓰인다 |

D1(코드가 아니라 조립을 바꾼다)·D4(잃는 것)·D6(CLI 시험 발송 10초)·D7(Pre-Edit)은 **유효하다.**
D6은 특히 그대로다 — 사람이 명령을 입력해 기다리는 경로는 루프가 아니므로 루프 밖으로
옮길 것이 없다.

### D0.8 — best-effort 알림도 루프를 붙잡는다. 침묵하지 않는다

exit 관측 루프가 올리는 이벤트는 6종이고, 그중 **2종이 normal 등급**이다 —
`EventExitProposalCapped`(`exitloop.go:1430`)와 `EventExitPositionUnmanaged`(`:1500`).
둘 다 `criticalEvents`(`internal/obs/event.go:279-299`)에 없다. normal은
`publishBestEffort`(`notifier.go:155-167`)로 가고, 그 함수는 outbox를 거치지 않고
**동기로 `Publisher.Publish`를 부른다.** 기한은 transport 자신의 것뿐이다
(`ntfy.go:95`).

곧 critical만 루프 밖으로 옮기면 **이 두 종은 여전히 손절 루프에서 네트워크를 기다린다.**
17판은 best-effort도 배달 루프로 넘긴다: 크기가 정해진 버퍼에 비차단으로 넣고, 가득 차면
버리고 로그 한 줄을 남긴다. 버리는 것이 등급 강등이 아닌 이유는 `notifier.go:160-161`이
이미 적었다 — *"this grade is best-effort by definition, and treating its failure as an
incident would make the grading meaningless."* outbox는 쓰지 않는다. best-effort에
durability를 주는 것은 이 change의 범위가 아니고, 주면 outbox가 사건 아닌 것으로 찬다.

### D0.9 — 18판이 더한 것 — 배달 루프가 죽어도 엔진은 내려가지 않는다

**사용자 결정 (2026-08-10)**: 결정 2 = 안 1.

`Runtime.Run`은 **첫 정지가 전부를 내린다**(`runtime.go:293-301`). 배달 루프를
그대로 등록하면 알림 하위 시스템의 결함 하나가 **엔진 정지**가 되고,
엔진 정지는 exit 관측 루프의 소멸이며, 그것은 **손절이 아예 없어지는 것**이다
(`internal/execgw/protection.go:55-60` — *"protection that ends when the process
does"*).

**그러므로 배달 루프는 이 규칙의 예외다.** 배달 루프의 `Run`이 반환해도
`Runtime`은 다른 루프를 내리지 않는다. 대신 그 반환은:

1. 진입 게이트를 `ReasonAlertUndelivered`로 래치하고,
2. `ModeTriggerCriticalAlertUndelivered`로 운영 모드를 `ENTRY_BLOCKED`로 승격하고
   (`operating_mode.go:518`·`:541-545` — 자동 `HALT_ALL`은 없다),
3. critical 알림을 **남긴다** — 나갈 수 없으므로 outbox 행으로.

**근거는 비대칭이다.** 알림이 안 나가는 상태는 진입을 막으면 유계다.
손절이 없는 상태는 무엇으로도 막을 수 없다. 부분 생존 금지 원칙
(`runtime.go:293-296`)이 지키려는 것은 *"청산을 관측하지 않으면서 대사하는 엔진"*이고,
**배달 루프는 그 감각기 목록에 없다** — 알림은 사람에게 알리는 경로이지
엔진이 세상을 보는 경로가 아니다.

**이것은 `internal/app/engine`의 델타다.** `SupervisedLoop`에 그 구별을 표현하는
필드가 오늘 없다. `NewRuntime`(`runtime.go:183`)이 지금 검증하는 것은
`Name`·`Run`·`Health`·`Trigger`뿐이다. 18판이 하나를 더한다.

> **재시작은 하지 않는다.** `Runtime`의 *"Nothing is restarted"*(`runtime.go:258`)
> 계약은 그대로다. 배달 루프가 죽으면 **죽은 채로 두고 진입을 막는다** —
> 되살리는 것은 사람이다. 재시작 정책을 넣으면 그 계약이 깨지고,
> 깨진 계약은 다음 편집이 다른 루프에도 적용한다.

### D0.10 — 18판이 더한 것 — a092가 재설계를 끝까지 구현한다

**사용자 결정 (2026-08-10)**: 결정 3 = 안 1.

17판의 `proposal.md`는 *"재설계는 만드는 것이 아니라 배선하는 것"*(:30)이라 쓰면서
동시에 **배달 루프·backlog 재시도·`obs` 편집을 non-goal로** 두었다(:415·:417).
17라운드 A-P7 = B-P1·P2가 그 모순을 잡았다. **18판이 non-goal 쪽을 지운다.**

| | 17판 | 18판 | **19판** |
|---|---|---|---|
| 배달 루프 신설 | 미배정 후속 | a092 | **a098** |
| backlog 재시도 | 미배정 후속 | a092 | **a098** |
| `internal/obs` 함수 편집 | non-goal | **a092의 델타** | 그대로 a092 |
| `internal/app/engine` 편집 | 주석만 | `SupervisedLoop` + 배선 | **a098** |
| 잠금 구간·배치 공정성·백로그 상한·등급 이관 | — | a092 | 그대로 **a092** |
| §8.6 `git diff --stat -- internal/obs`가 비어 있음 | 완료 조건 | **폐기** | 폐기 유지 |

> **19판이 이 표의 두 줄을 a098로 옮겼다 (사용자 결정 2026-08-10, 결정 3 = 안 1).**
>
> 18판은 배달 루프 신설과 잠금 재설계와 공정성과 등급 이관을 **한 change에 묶었다.**
> 열여덟 라운드가 그 묶음을 통과시키지 못했다. 결정 3은 **살아 있는 결함만 떼어낸다** —
> `order.unresolved_in_doubt`가 기록되고 게이트를 잠그는데 **아무도 안 보내는 것**.
>
> 새 change: **`a098-nobody-sends-what-the-outbox-keeps`**.
> 경계는 한 문장이다 — **a098은 `Flush`를 얼마나 자주 부르는가를 정하고,
> a092는 한 번 부를 때 무엇이 일어나는가를 바꾼다.** a098은 `obs`를 한 줄도
> 안 건드리므로 두 change가 같은 줄을 청구하지 않는다.
>
> **순서 의존이 없다**: a098은 `Flush`의 현재 시그니처만 쓰고, a092는 본문을 바꾸되
> 시그니처를 안 바꾼다. 어느 쪽이 먼저 들어가도 된다.
>
> 이 표를 여기서 고치는 이유는 16라운드 B-8이다 — *"형제 change 목록이 불완전하다"*.
> 두 change가 같은 일을 청구한 채로 두면 그것이 다음 라운드의 지적이 된다.

배선만 남기면 이 change는 제목의 속성을 다시 지지 않게 되고, 그것이 16라운드가
막은 바로 그 상태다.

## D1 — 코드가 아니라 조립을 바꾼다

34초는 `obs`의 결함이 아니다. `obs`는 세 개의 손잡이를 **이미** 갖고 있고 문서화까지 했다.

| 손잡이 | 선언 | 기본값을 고르는 분기 |
|---|---|---|
| `Notifier.Attempts` | `notifier.go:103-105` | `deliver` B1 `:343` |
| `Notifier.RetryDelay` | `notifier.go:106-107` | `wait` B1 `:412` |
| `Ntfy.Timeout` | `ntfy.go:72-73` | `Publish` B3 `:96` |

> **19판 재고정 (HEAD `285c7619`).** 이 표의 네 좌표가 a097 이전 값이었고, 그중
> **둘은 다른 함수의 줄을 가리키고 있었다** — 옛 `:245`는 지금 `claimAndDeliver`의
> `if err != nil`이고, 옛 `:292`는 `escalate`의 주석이다. 표는 *"이 분기가 기본값을
> 고른다"*를 주장하므로 좌표가 다른 함수로 옮겨 가면 주장이 조용히 거짓이 된다.
> 분기 좌표는 손으로 읽지 않고 `deliver`·`wait`의 `ast.json`에서 가져왔다.

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
`ClaimAlertForDelivery` 1회(outbox 기록) + `MarkAlertAttemptFailed` 시도 수만큼(시도별 실패 기록) + `Gate.Block`(게이트 래치) + **`n.escalate`의 `EscalateOperatingMode` 승격 트랜잭션**(`notifier.go:222` 호출 → `:312` 원장 호출 — 예산이 소진된 그 경로에서만 돈다) + **그 호출이 쓰는 구조화 로그 줄**(`notifier.go:148` `logEvent` Warn·`:399` `deliver` 소진 Error는 항상, `:323` `escalate` 승격 Warn은 승격이 실제로 일어났을 때만).

> **이 열거가 다섯 항인 것은 두 delta의 SHALL과 같아야 한다.** 9판까지 이 줄은
> 네 항이었다 — 로그 쓰기를 D3 주석과 두 delta에만 더하고 **유도 정본인 여기**와
> 제안서와 측정 정본에는 안 더했다(9라운드 차단 2). 유도 정본이 spec보다 적게
> 세면 그 차이만큼 예산의 근거가 비어 있다. `tools/check_values.py`의
> `check_dwell_enumeration`이 이제 이 다섯 문서를 전부 본다.
>
> **19판 재고정 (HEAD `285c7619`).** 이 열거의 좌표 다섯과 이름 하나가 a097 이전
> 것이었다. 이름은 좌표보다 무겁다 — `notifyCritical`은 a097 이후 `EnqueueAlert`를
> 부르지 않고, outbox 기록은 `claimAndDeliver`(`:194`)가 `ClaimAlertForDelivery`(`:244`)로
> 한다. **다섯 항이라는 수와 그중 셋만 원장 커넥션에 줄 선다는 판정은 변하지 않는다** —
> `ClaimAlertForDelivery`는 `EnqueueAlert`가 위임하던 바로 그 문장이다.

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
| 3 (12판 채택 → **13판 기각**) | 3 | 1.3s | 150ms | 4.20s | 800ms | 1.72× |
| 4 | 2 | 2.0s | 250ms | 4.25s | 750ms | 2.65× |
| 5 | 3 | 1.0s | 500ms | 4.00s | 1000ms | 1.33× |
| 6 | 3 | 0.7s | 200ms | 2.50s | 2500ms | **0.93× — 냉 측정보다 작다** |
| 7 (7라운드 M-4) | 3 | 1.3s | **50ms** | 4.00s | **1000ms** | 1.72× |
| 8 (13판 채택 → **14라운드 기각**) | 1 | 1.3s | 150ms (미실행) | 1.30s | 3700ms | 1.72× |
| **9 (14판 채택)** | **1** | **3.5s** | 150ms (미실행) | **3.50s** | **1500ms** | **4.64×** |

### 14판이 #8을 버리고 #9를 택한다 — 실측이 1.3초를 반증했다

**14라운드가 실제 전송기를 재서 #8을 기각했다.** `https://ntfy.sh/`에 읽기 전용
GET **20건**(`curl`이 매번 새 프로세스이므로 **표본은 전부 냉연결**이다):

| 평균 | 중앙값 | p90 | 최소 | **최대** |
|---|---|---|---|---|
| **1.795 s** | 1.729 s | 2.191 s | 1.431 s | **2.721 s** |

`alertPublishTimeout = 1.3s`는 **이 전송기의 평균의 0.72배다.** 20건 중 **20건이
그것을 넘는다.** #8 아래에서는 건강한 ntfy가 매 시도 타임아웃한다 — 즉 #8은
"느린 전송기를 실패로 부른다"가 아니라 **정상 전송기를 실패로 부른다.**

> **⚠ 3.5초의 여유는 최대 대비 1.29배이고, 이 문서는 그것을 크다고 적지 않는다.**
> 첫 8건에서 최댓값은 2.125s였고 20건으로 늘리자 **2.721s**가 나왔다 — 표본이 늘면
> 최댓값이 올라간다는 것 말고는 이 꼬리에 대해 아는 것이 없다.
> §4가 창 안 publish 유발 줄의 **46%를 냉으로 판정**했으므로 이 최댓값은 꼬리가
> 아니라 **정상 범위 안**이다.
>
> **그리고 이것은 홈페이지 GET이지 topic POST가 아니다.** 경계짓는 것은
> **네트워크 경로**이고 publish 핸들러가 아니다 — 진짜 POST는 **더 길 수 있다.**
> 진짜 POST를 재려면 사용자 topic으로 실제 알림이 발송되므로(휴대폰에 뜬다)
> 사람의 승인 없이는 하지 않는다. 재측정 절차는 tasks 9.8.
>
> **그러므로 측정이 한 일은 1.3초를 기각한 것이고, 3.5초를 확정한 것이 아니다.**

**#9가 사는 것과 파는 것을 재서 적는다.** 관측 5회를 순차로 돌리고 누적 체류와
행 상태를 봤다(`internal/obs`에 `-overlay`로 넣은 프로브, `go test` 통과):

| 서버 지연 | 오늘 (3회·10s·2s) | #8 (1회·1.3s) | **#9 (1회·3.5s)** |
|---|---|---|---|
| 500ms | 0.51s · DELIVERED · 래치 0 | 0.51s · DELIVERED · 래치 0 | **0.51s · DELIVERED · 래치 0** |
| **1.9s** (실측 평균 근처) | 1.91s · DELIVERED · 래치 0 | **6.55s · PENDING · 래치 1** | **1.91s · DELIVERED · 래치 0** |
| **2.125s** (첫 8건의 최댓값) | 2.14s · DELIVERED · 래치 0 | **6.55s · PENDING · 래치 1** | **2.14s · DELIVERED · 래치 0** |
| 4.0s | 4.02s · DELIVERED · 래치 0 | 6.56s · PENDING · 래치 1 | **17.57s · PENDING · 래치 1** |

**실측 전송 구간(≤2.721s)에서 #9는 오늘과 같은 수를 낸다** — 같은 1회 발송, 같은
DELIVERED, 래치 없음. #8은 같은 구간에서 5회 발송·PENDING·래치다.

**#9가 파는 것은 [3.5s, 10s] 구간이다.** 그 안에서 오늘은 성공하고 #9는 실패하며,
실패는 `claimOwed`의 PENDING 무창 재-owed 때문에 **관측마다 되풀이된다** —
4.0초 행의 17.57초가 그것이고, 그것은 #8의 6.56초보다 **크다.** 역행 구간은
[1.3s, 10s]에서 [3.5s, 10s]로 **좁아지지만, 그 안의 대가는 커진다.**

이 교환을 택하는 근거는 하나다: **좁아진 구간은 실측 분포 밖이고**(최대 2.721s),
넓었던 구간은 실측 분포를 **통째로 삼켰다**(20건 전부가 1.3s 위다). 구간의 폭이 아니라
**측정된 지연이 그 안에 있느냐**가 판정 기준이다.

**그 기준은 여유 1.29배 위에 서 있다.** 실측 최댓값이 2.721s이고 상한이 3.5s이므로,
꼬리가 780ms만 더 길면 이 근거는 무너진다. **이것이 이 change에서 가장 얇은 근거다.**

### 13판이 #3을 버리고 #8을 택한 근거 (기록 — #8 자체는 위에서 기각됐다)

**12라운드(codex, 교차 모델)가 #3에서 손절이 더 늦어지는 구간을 찾았다.**
위 표의 `transport` 열은 **최악**이고, 최악만 비교하면 중간 구간의 역행이 숨는다.

| publish 실제 지연 | 오늘 (10s × 3, 2s 대기) | #3 (1.3s × 3, 150ms) | **#8 (1.3s × 1)** |
|---|---|---|---|
| timeout 아래 | 즉시 성공 | 즉시 성공 | 즉시 성공 |
| timeout ~ transport 사이 | **1회 성공** | **3회 타임아웃 + 실패 + 래치 — 더 늦다** | **1회 타임아웃 — 오늘보다 짧다** |
| transport 이상 | 성공 또는 34s | 실패 | 실패 |
| 무응답 | **34s** | 4.2s | **1회 상한** |

시도가 하나면 체류는 `min(실제 지연, alertPublishTimeout)`이고, 오늘의 체류는 같은
publish에 대해 **항상 그 이상이다** — 오늘은 첫 시도가 실패하면 재시도로 34초까지 간다.
그러므로 **알림 한 건에 대해 #8은 어떤 지연 구간에서도 오늘보다 나쁘지 않다.**
#3은 그 성질을 갖지 못한다.

### 위 표는 **한 건**의 표다 — 에피소드 누적에서 #8은 오늘보다 나쁘다

**13판이 재서 찾았고, 13라운드가 그 측정을 기각해서 다시 쟀다.**
세 구성을 같은 지연 서버에 걸고 관측을 두 번 돌린다. 프로브는 `-overlay`로
`internal/app/engine`에 넣었고 `go vet` 통과다.

> **⚠ 13판 초판의 이 표는 틀렸다 — 13라운드 P5가 재현으로 기각했다.**
> 두 군데가 틀렸다.
>
> 1. **#3이라고 적은 행이 #3이 아니었다.** 프로브가 `RetryDelay`를 안 채워
>    `obs.DefaultRetryDelay`(2초)가 쓰였다. 잰 8.13초는 `3×1.3 + 2×2.0 = 7.9초`이고
>    그것은 **후보 표에 없는 구성**이다. 진짜 #3은 `3×1.3 + 2×0.15 = 4.2초`다.
> 2. **지연 5.2초는 #3의 역행 구간 밖이다.** #3의 체류는 4.2초로 고정이므로
>    5.2초 지연에서는 #3이 오늘보다 **빠르다.** 역행을 보려면 지연이
>    `alertPublishTimeout`과 #3의 transport 사이여야 한다.
>
> 즉 12라운드 P1을 보이려고 만든 측정이 **P1을 못 보이는 구간에서** 잘못 라벨된
> 구성을 잰 것이다. 아래는 두 구간 × 세 구성 전부를 다시 잰 것이다.

| 서버 지연 | 구성 | 1관측 | 2관측 | 2관측 누적 발송 | 래치 | PENDING |
|---|---|---|---|---|---|---|
| **2.0s** | 오늘 (10s·3회·2s) | **2.01s — 성공** | **0.00s** | **1** | 없음 | 0 |
| 2.0s | #3 (1.3s·3회·150ms) | **4.24s** | **4.23s** | **6** | 래치 | 1 |
| 2.0s | **#8 (1.3s·1회)** | **1.31s** | **1.30s** | **2** | 래치 | 1 |
| **5.2s** | 오늘 | 5.22s — 성공 | 0.00s | 1 | 없음 | 0 |
| 5.2s | #3 | 4.22s | 4.24s | 6 | 래치 | 1 |
| 5.2s | **#8** | **1.31s** | **1.31s** | 2 | 래치 | 1 |

**12라운드 P1의 역행은 2.0초 구간에서 나온다**: 오늘 2.01초 → #3 4.24초.
**#8은 두 구간 모두에서 관측당 오늘보다 짧다** (1.31 < 2.01, 1.31 < 5.22).
그리고 #3은 두 관측에 **여섯 번** 발송한다 — 관측당 세 번의 페이저 폭풍이다.

**그런데 #8도 누적에서는 오늘 이하가 아니다.** 오늘은 발송이 **성공**하므로
`RemindAfter` 1시간이 이후 관측을 전부 억제해 체류가 **0**이 된다(위 표의 `2관측
0.00s`가 그것이다). #8에서는 행이 PENDING으로 남고 `claimOwed`가 PENDING을 창 없이
곧바로 `owed`로 돌려주므로(`outbox.go:277-279`, `case AlertPending: return true, false`)
**매 관측이 1.31초를 낸다.**

| 서버 지연 | 오늘 누적 | #8 누적 | 넘는 시점 |
|---|---|---|---|
| **2.0s** | 2.01s (1회로 끝) | 1.31s/관측 | **2관측 — 약 10초 뒤** |
| 5.2s | 5.22s (1회로 끝) | 1.31s/관측 | 4관측 — 약 20초 뒤 |

**그러므로 §0.3의 문장에는 "알림 한 건에 대해"가 반드시 붙는다.** 13판 초안은
그것 없이 "어떤 publish 지연에서도"라고 적었고 실측이 그 문장을 기각했다.

#### 경계 구간에서 관측당 비교가 뒤집히는가 — 측정 (13라운드 P1b)

위 두 구간(2.0s·5.2s)은 지연이 기한보다 훨씬 크다. **기한을 갓 넘은 구간**에서는 다르다:
오늘은 **성공 경로**(`MarkAlertDelivered` 뒤 즉시 `return true`, `notifier.go:354`)를 타고,
a092는 **실패 경로**(`MarkAlertAttemptFailed` + `Gate.Block` +
`EscalateOperatingMode` 원장 트랜잭션 + 구조화 로그 2줄)를 탄다.
**outbox 기록은 두 경로에 공통이므로 이 차이에 안 들어간다** — `ClaimAlertForDelivery`가
발송보다 먼저 돌고, 성공이든 실패든 같은 값을 낸다. 실패 경로가 더 무거우므로
경계 근처에서 역전이 가능하다.
그것이 13라운드 P1b의 주장이고, **재서 확인했다.**

| 서버 지연 | 오늘 (10s·3회) | #8 (1.3s·1회) | 차 |
|---|---|---|---|
| **1.31s** | 1.324s 성공 | **1.311s** 실패·래치 | #8이 13ms 짧다 |
| 1.35s | 1.356s 성공 | 1.309s | 47ms |
| 1.50s | 1.507s 성공 | 1.313s | 194ms |
| 2.00s | 2.014s 성공 | 1.314s | 700ms |

**역전은 일어나지 않는다.** a092의 실패 경로가 transport 위에 얹는 비용은
`1.311 − 1.300 = 약 11ms`이고, 오늘의 성공 경로도 `1.324 − 1.310 = 약 14ms`를 쓴다.
가장 빠듯한 1.31초 행에서도 #8이 더 짧다.

**이 11ms는 상한이 아니라 관측값이다.** 그 안의 `ClaimAlertForDelivery`는
`BeginTx`(`outbox.go:182`)로 시작하고 그 트랜잭션에는 시간 상한이 없다 — 디스크가 멈추면
11ms가 아니다. **비-transport 작업은 이 예산이 상한을 주는 대상이 아니다**(D2의 `n.mu`
절과 같은 종류의 인정이며, spec이 그것을 말한다).

**그리고 또 다른 대가는 새로 생기는 차단이다.** `alertPublishTimeout`보다 느리지만 **동작하는**
전송기(오늘은 성공으로 세는 것)가 #8 아래에서는 매 관측마다 실패로 기록되고,
게이트가 래치되며, `EscalateOperatingMode`가 `ENTRY_BLOCKED`를 **원장에 쓴다.**
오늘 없던 차단이 새로 생기는 것이다.

> **⚠ 13판 초판은 여기에 "재시작에도 남는다"라고 적었다. 13라운드 보이스 B가 그것을
> 기각했고, 재현해서 확인했다 — 틀린 것은 나다.** 아래가 실측이다.
>
> ```text
> 재시작 전 : alert_latch=true  · mode="ENTRY_BLOCKED" · gate.Blocks()=1
> 재시작 후 : mode(원장)="ENTRY_BLOCKED" · PENDING=1 · alert_latch=false · gate.Blocks()=0
> ```
>
> **기록은 살아남고 강제는 살아남지 않는다.** `EntryGate`의 래치는 메모리 map이고
> (`modegate.go:37,50`), 원장의 모드 행을 새 게이트에 다시 투영하는 경로는
> **프로덕션에서 아무도 부르지 않는다** — `RestoreOperatingModeProjection`
> (`operating_mode.go:574`)·`SetModeProjector`(`:289`)·`CurrentOperatingMode`(`:553`)
> 셋 다 호출자가 테스트뿐이다.
>
> 이것은 a092가 만든 결함이 아니라 **a092가 기대려 한 보상이 없다는 사실**이다.
> 방향은 양쪽으로 간다: **거짓 차단은 재시작으로 스스로 풀리므로 덜 해롭고**(D4 손실①),
> **종료 알림의 보상은 사라진다**(D5 P3). 두 결론 다 아래에 반영했다.
> 배선은 이 change의 범위 밖이고 **미배정 후속에 적는다**(proposal의 이관 목록).

이것을 그대로 안고 #8을 택하는 근거는 셋이다.

1. **진입 차단은 보수 방향이고 청산에 영향이 없다**(안전 불변식 6·4). a092가 지켜야
   하는 것은 손절의 즉시성이고 그것은 지켜진다.
2. **1.3초는 유일한 냉 측정 0.754초의 1.72배다.** 그보다 느린 전송기는 실제로
   degrade된 것이고, 그때 사람을 부르는 것이 이 시스템이 하기로 한 일이다.
3. **오늘 이 구간이 조용한 것은 안전해서가 아니라 관측하지 않아서다.** 10초 기한은
   5초 주기 루프 안에서 이미 계약 위반이다 — 그것이 a092가 존재하는 이유다.

근거 2가 무너지면(실전 ntfy가 1.3초를 넘게 응답하는 것이 관측되면) 택할 것은
timeout 재조정이지 시도 수 복구가 아니다. **그 판정 절차는 tasks 9.8이다.**

**#4를 기각했던 근거(`DefaultCriticalAttempts` 주석의 transient 논거)가 여기서 뒤집힌다.**
그 주석은 **전달 신뢰도**를 근거로 3을 골랐고, 그것은 전송이 손절 루프 안에 있을 때
치르는 값이 무엇인지 세지 않은 판단이다. a092의 목적은 예산의 명시가 아니라 **루프의
해방**이며, 안전 불변식은 전달 신뢰도가 아니라 손절의 즉시성을 우선한다.

> **⚠ 12판은 여기에 "재시도는 사라지지 않는다 — 루프 밖으로 간다(미배정 후속)"라고 적었다.
> 13라운드 P7이 그것을 기각했고 맞다.** 미배정 후속은 아직 없고 `Flush`에는 프로덕션 호출자가
> 없다. **a092가 끝난 시점에 루프 밖 재시도는 존재하지 않는다.** 정확히 말하면:
> exit·대사 루프에서 재시도를 대신하는 것은 **다음 관측**이고(PENDING 행이 창 없이 다시
> owed이므로), 다음 관측이 없는 경로(`runtime.go:453`의 종료 알림)에는 **대신할 것이 없다.**
> 경로별 평가는 D5.

`alertRetryDelay`는 회차가 하나이므로 **실행되지 않는다.** 그래도 채운다:
빈 자리는 `DefaultRetryDelay`(2s)로 돌아가고 그것은 예산 밖이다.

### 12판까지의 채택 근거 (기록으로 남긴다)

**#3을 택했다.**

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
  미배정 후속이 재시도한다. **하드한 쪽에 여유를 준다.**
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
**시도 횟수**이고, 그보다 긴 회복을 사는 것은 **outbox 행의 생존**(미배정 후속)이다.

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

> **⚠ 이 절이 걸어 둔 재검토 조건이 14라운드에 발동했다.** 아래 :413의 조건은
> *"냉연결 publish가 1.3초를 넘는 것이 **한 번이라도** 관측되면 이 값을 다시 유도한다"*
> 였다. **20건 중 20건이 1.3초를 넘었다**(`delivery-latency.md` §9.1, 최소 1.431s).
> 조건대로 다시 유도했고, 그 결과가 후보 #9(1회·3.5s)다.
>
> **아래 본문은 그 재유도 *이전*의 논증이며 기록으로 남긴다.** 1.3초를 채택값으로
> 읽지 않는다. 살아남는 것은 논증의 형태 — *"여유는 여유만큼만 믿는다"* 와
> *"틀렸을 때 무엇이 알려주는지가 정해져 있어야 한다"* — 이고,
> **14판의 3.5초에도 같은 조건이 붙는다**(아래 :413 갱신).

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
| 12판의 1회 상한 | 1.3 s | 유효 최댓값의 1.72배 — **14라운드에 기각됐다** |
| **14판 채택 1회 상한** | **3.5 s** | 실전 냉연결 20건의 최댓값 2.721 s의 **1.29배**(§9.1) |

**만약 무효 표본이 왕복이었다면 1.3초는 작다.** 이 change는 그 가능성을 측정으로
닫지 않았다. 닫는 것은 두 가지다 — design D2의 재검토 조건(냉 publish가 1.3초를 넘는
것이 **한 번이라도** 관측되면 재유도)과 그 관측을 가능하게 하는 **a090의 계측**.
a090 이전까지 이 상한은 **판단이며, 그 판단이 틀렸을 때 무엇이 그것을 알려주는지가
정해져 있다**는 것이 이 절이 주장하는 전부다.

**그래서 이 숫자에는 재검토 조건을 붙인다**: 냉연결 publish가 `alertPublishTimeout`을
넘는 것이 한 번이라도 관측되면 이 값을 다시 유도한다. 그 관측을 가능하게 하는 계측은
a090이고, 판정 절차는 tasks 9.8이다.

> **이 조건은 12판의 1.3초에 대해 이미 발동했다** — 14라운드가 실전 전송기를 재서
> 20건 전부가 그것을 넘는 것을 보였고, 그래서 값이 3.5초로 다시 유도됐다.
> **조건은 새 값에도 그대로 걸린다.** 그리고 새 값의 여유는 1.29배로 **더 얇다** —
> 표본이 20건으로 늘면서 최댓값이 2.125s에서 2.721s로 올라갔기 때문이다.
> 표본을 더 늘리는 것이 이 조건을 다시 시험하는 가장 싼 방법이다.

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

a092는 이것을 **고치지 않는다**(전송이 루프 밖으로 나가야 하고 그것은 미배정 후속이다).
고치지 않는 대신 **spec이 그것을 말하게 한다** — engine-safety 델타는 상한을
"자기 차례"로 한정하고, 직렬화가 그 상한의 배수를 만든다는 것을 실측값과 함께 적는다.

> **4판 초안은 exit-policy ¶20에서 이 사실을 인정해 놓고 같은 요구의 ¶16에서
> 반대로 SHALL NOT을 걸었다.** engine-safety 델타에는 `n.mu`가 한 글자도 없었다.
> 2라운드 C3가 이름 붙인 것이 네 판째 같은 자리에서 재발한 것이고, 그것이 4라운드 C1이다.

### 정상 등급

`publishBestEffort`는 publish 1회다(AST branches 2, 루프 없음). 10s → **`alertPublishTimeout`**.
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

> # ⛔⛔ 이 절 전체는 **16판의 상수 집합**이고 17판이 그것을 대체했다 (19라운드 A-P7 = B-P3)
>
> **아래에서 정의·단언·측정되는 상수 넷은 오늘의 채택값이 아니다** —
> `alertBudget` · `alertTransportBudget` · `alertOverheadReserve` · `alertRetryDelay`.
> 그 사실은 **D0.6이 이미 적었다**(`design.md:405-409`: *"`alertRetryDelay`는
> 사라진다"* · *"셋의 유도가 무효가 된다"*). **그런데 이 절은 그 문장을 안 받았고,
> 절 제목은 여전히 「상수는 한 자리에 모으고」다.** 위에서부터 읽지 않고
> 「상수」를 찾아 들어온 사람은 **이 절을 현재 설계로 읽는다.**
>
> | | 어디 |
> |---|---|
> | **오늘의 채택 다섯** | **`design.md:378-402`의 `// notifications.go — 17판` 블록** |
> | 오늘의 컴파일 단언 여섯 | **`tasks.md` §8.2** — 그 절의 실측 표가 구성 수를 스스로 세고, **못 잡는 구성**을 따로 열거한다 |
> | 아래 절이 정의하는 넷 | **철회됨.** 근거는 D0.6 |
>
> **아래를 지우지 않는 이유**는 D3의 측정 기록(§1333·§1374·§1407)이 *"왜 나노초가
> 아니라 밀리초인가"*(15라운드 A T1)와 *"왜 `[0]struct{}`가 방어를 뚫는가"*를
> 지고 있고, 그 둘은 **17판 단언에도 그대로 걸리는 사실**이기 때문이다.
> 값은 죽었고 **기전은 살아 있다.** 그러므로 이 절은 **측정 기록으로 읽고
> 상수 정의로 읽지 않는다.**
>
> `tasks.md` §8.2.1이 *"D3의 단언 표를 채택 상수 위에서 다시 돌린다"*고 적는데,
> **그 재실행은 §8.2가 이미 했다**(19판 2차·4차). 8.2.1은 그 사실을 가리키도록
> 고쳤다.

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

// alertPublishAttempts is ONE. This is the change, not a restatement.
//
// obs.DefaultCriticalAttempts is three, and three attempts is why an alert can
// hold the exit loop longer than a single slow-but-successful publish would:
// a publish that takes longer than alertPublishTimeout but less than three of
// them succeeds today and fails three times here, costing more wall clock than
// it cost before. The exit loop is where the stop lives, so the loop wins and
// the retry loses. Retries are not gone; they move out of the loop (미배정 후속).
const alertPublishAttempts = 1

// alertPublishTimeout bounds one publish, and with one attempt it is also the
// whole transport budget. It must sit ABOVE the transport's normal response,
// not near it: below it, a healthy ntfy is recorded as a failure on every
// observation, which latches the entry gate and writes ENTRY_BLOCKED.
//
// 3.5s is chosen from twenty read-only GETs against https://ntfy.sh/, every one
// of them a cold connection because curl re-execs -- mean total 1.795s, median
// 1.729s, p90 2.191s, max 2.721s. The margin over the max is 1.29x, which is
// thin, and it is the thinnest evidence in this change. That is a homepage GET
// rather than a topic POST, so it bounds the network path and not the publish
// handler; measuring the real POST would send notifications to the operator's
// phone and needs their say-so (tasks 9.8).
//
// The 1.3s this replaced was derived from a single cold-pool sample of 0.754s
// taken from the engine's own log, and the transport measurement refuted it:
// 1.3s is below the transport's mean. See the change's delivery-latency
// analysis for what that document does and does not measure.
const alertPublishTimeout = 3500 * time.Millisecond

// alertRetryDelay is never reached: one attempt means deliver never calls wait.
// It is set anyway. Leaving it zero would let obs.DefaultRetryDelay -- two
// seconds, which is outside this budget -- come back the moment someone raises
// alertPublishAttempts. A field that is only correct while a neighbouring
// constant keeps its value is a field to fill, not to leave.
const alertRetryDelay = 150 * time.Millisecond

// alertTransportBudget is what the fields above cost on the network. With one
// attempt the retry term is zero; it is written out so that raising the attempt
// count changes this constant instead of silently leaving it behind.
const alertTransportBudget = alertPublishAttempts*alertPublishTimeout +
	(alertPublishAttempts-1)*alertRetryDelay

// alertOverheadReserve is what alertBudget leaves for the work Notify does
// around the sending. The list must be exhaustive for the exhaustion path,
// because a term left out of it is a term the next edit reads as free. Both
// spec deltas make that completeness a SHALL.
//
// At base 285c7619 the exhaustion path is:
//
//	notifier.go:244  ClaimAlertForDelivery -- the outbox insert AND the claim,
//	                 one call, under n.mu. (EnqueueAlert is now a wrapper around
//	                 it, outbox.go:115, and the notifier does not call it.)
//	notifier.go:384  MarkAlertAttemptFailed, once per attempt
//	notifier.go:404  Gate.Block -- deliver's out-of-attempts latch
//	notifier.go:222  n.escalate, which reaches
//	notifier.go:312  EscalateOperatingMode -- a journal transaction
//	notifier.go:148  logEvent's Warn
//	notifier.go:399  deliver's Error on exhaustion
//	notifier.go:323  escalate's Warn, only when the mode actually changed
//
// There are TWO other paths that latch the gate, and they are not this budget's
// exhaustion path but they are not free either:
//
//	notifier.go:262  the claim itself failed (a097). No publish happens at all,
//	                 so no transport time -- but escalate at :213 still runs a
//	                 journal transaction.
//	notifier.go:379  published, then MarkAlertDelivered failed. One publish of
//	                 transport time, then the same journal and log work.
//
// Every line number above was re-derived at base after a097 moved them; the
// citations this comment carried through the 13th revision pointed at unrelated
// lines, and the enumeration also undercounted -- one gate latch where there
// are three, one escalate site where there are two.
//
// Everything but the latch and the log queues on one journal connection;
// EntryGate.Block takes a mutex and writes a map (execgw/retry.go:498).
//
// Measured worst across every probe session is 356ms; the same nominal
// conditions have also produced 32ms, so this number tracks ambient load more
// than it tracks the code.
//
// At one attempt and a 3.5s timeout this derives to 1500ms. The rule that used
// to bind it -- at least twice the largest value ever observed, 712ms -- is  <!-- not-a-measurement: 2배 규칙의 임계값 — 356.1ms 실측의 2배로 유도한 값이다 -->
// still satisfied, at 2.1x, but it is no longer what picks the number: the
// timeout does, and the reserve is what is left. Through the 3-attempt design
// it derived to 800ms and the rule was the only thing keeping it positive.
//
// This is the constraint that bounds how far the timeout can rise, and the
// bound is 4.288s: the reserve is the remainder, and the rule needs twice the  <!-- not-a-measurement: 5.0s − 712ms의 산술이다 -->
// 356.1ms worst, so transport <= 5s - 712ms.  <!-- not-a-measurement: 같은 유도 임계 — 356.1ms 실측의 2배다 -->
//
// The 14th revision wrote here that 4.2s violates the rule. It does not. At
// 4.2s the reserve is 800ms, which is 2.25x the same 356.1ms -- the identical
// figure this design's own candidate #3 recorded as satisfying it. Two lines of
// one document read the same 800ms in opposite directions; this one was wrong.
// At 5s the reserve is zero, which both spec deltas forbid.
//
// What still binds is the direction: the reserve is what is left over, so
// raising the transport budget spends it. Re-measure the overhead before
// trusting either number on a machine other than the one in the change's
// delivery-latency analysis.
const alertOverheadReserve = alertBudget - alertTransportBudget

// The five arrays are the assertion. Each subtracts so that the *smallest legal
// value* is the one that still compiles, and each names its own constant in the
// failure.
//
// Zero is not "unset" at the callee, it is a different, larger number: attempts
// <= 0 becomes DefaultCriticalAttempts, 3 (notifier.go:245), delay <= 0 becomes
// DefaultRetryDelay, 2s (notifier.go:292), and timeout <= 0 becomes 10s
// (ntfy.go:96). At one attempt those fall out unevenly, and the 15th revision
// says so rather than reusing the 3-attempt sentence it inherited:
//
//	alertPublishTimeout = 0   -> callee 10s      -> transport 10s vs a 3.5s budget
//	alertPublishAttempts = 0  -> callee 3        -> transport 10.8s  <!-- not-a-measurement: attempts=0 폴백(피호출자 3회)의 산술 — 3×3.5s + 2×150ms이다 -->
//	alertRetryDelay = 0       -> callee 2s       -> NOTHING. One attempt never
//	                                                waits between attempts.
//
// That last line is why alertRetryDelay is asserted anyway, and the reason is
// not runtime defence: at one attempt the field cannot change behaviour in any
// configuration this file compiles, because raising the attempt count fails the
// budget relation below first. It is asserted because engine-safety requires the
// assembly site to own all three -- attempts, per-publish bound, inter-attempt
// wait -- and a field nobody guards is a field the next edit deletes. An inert
// constant that is explicitly inert is not the same as a missing one.
//
// The first array is the budget itself, and it is strict: both spec deltas say
// the transport budget SHALL NOT equal the observation interval, because the
// work enumerated above is not zero. It is also the one a092 exists for -- the
// engine loop publishes AT MOST ONCE. This is a relation, not a pinned number,
// so it survives retuning the timeout. It is zero when alertPublishAttempts is
// one and negative for any higher count -- which is exactly the configuration
// that made the loop slower than it was.
var _ [alertPublishTimeout - alertTransportBudget]struct{}

var _ [alertPublishAttempts - 1]struct{}

// The three below count MILLISECONDS, not nanoseconds, and that is the whole of
// the 15th revision's change here.
//
// Counting nanoseconds cannot tell 3500ms from 3500ns. A dropped unit --  <!-- not-a-measurement: 상수 리터럴이지 측정이 아니다 -->
// alertPublishTimeout = 3500 -- passed a [alertPublishTimeout - 1] assertion and
// produced a green build whose per-publish timeout is 3500ns, which times out
// before any dial completes: every alert exhausts, every alert latches the entry
// gate and escalates the operating mode. That is worse than the zero the
// nanosecond form caught.
//
// Counting milliseconds catches both. Zero and negative fail the subtraction
// exactly as before, and a dropped or mistyped unit collapses to 0 in the
// division and fails too. The nanosecond forms were strictly subsumed.
//
// They were also unrepresentable. [alertPublishTimeout - 1] is 3,499,999,999 at
// the adopted timeout, which does not fit in int32, so the adopted configuration
// itself did not compile for GOARCH=386 -- measured, not reasoned. Array lengths
// in milliseconds are in the thousands and fit anywhere.
var _ [alertPublishTimeout/time.Millisecond - 100]struct{}
var _ [alertRetryDelay/time.Millisecond - 10]struct{}
var _ [alertOverheadReserve/time.Millisecond - 1]struct{}
```

**단언이 왜 다섯이고 왜 마지막 셋이 밀리초인가 — 7라운드 B1·H1, 15라운드 A T1.**

7판까지의 단언은 `var _ [alertOverheadReserve]struct{}` 하나였다. **`[0]struct{}`는
합법적인 Go 배열이므로 그것이 강제한 것은 `transport ≤ budget`뿐이다.** 두 spec delta는
`SHALL NOT`으로 **`<`** 를 요구한다(engine-safety:26 · exit-policy:16). 3판이 기각당한
후보 #1(T=1.5s·D=250ms, transport = 주기)을 그 단언 아래 넣으면 **BUILD OK**가 난다 —
**기각의 이유였던 바로 그 구성이 컴파일을 통과한다.** 빼기 하나가 그 한 칸을 닫는다.

그리고 단언은 **합**만 보고 **항의 부호**를 안 봤다. 세 필드 전부 피호출자에 0 폴백이
있으므로(위 주석) 비워 둔 자리는 engine-safety:24가 `SHALL NOT`으로 금지한
"그 자리에서 비워 두어 피호출자의 기본값이 쓰이는 구성"이 된다. **세 배열이 그 세
항을 각각 본다.**

> **⚠ 14판까지 이 자리는 `alertRetryDelay = 0`의 대가를 「실제 transport 7.9초」라고  <!-- not-a-measurement: 14판까지 있던 3회 설계 유도값을 인용해 기각한다 — 3×1.3s + 2×2s -->
> 적었고, 그것은 3회 설계의 산술이다**(3×1.3s + 2×2s). 14판이 시도를 1회로 내린 뒤
> **그 문장은 참이 아니다** — 시도가 하나면 시도 사이의 대기가 없으므로
> `alertRetryDelay = 0`은 런타임에서 **아무것도 바꾸지 않는다**(15라운드 B N8).
> 위 주석이 세 폴백의 결과를 1회 기준으로 다시 적는다.
>
> **이것이 이 change에서 같은 형태로 반복된 오류다**: 전제(시도 3회)가 죽었는데
> 그 위에 선 유도값(7.9초·1.1초)이 문장으로 살아남았다. 아래 A T1도 같다.  <!-- not-a-measurement: 같은 기각 인용 — 둘 다 3회 설계의 유도값이다 -->

### 16판이 나노초 단언 셋을 밀리초로 바꾼다 — **채택 구성이 32비트에서 컴파일되지 않았다**

`[alertPublishTimeout - 1]`은 채택 상한에서 **3,499,999,999**이고 int32에 안 들어간다.
**추론이 아니라 측정이다** — 별도 모듈에 채택 상수와 7단언을 그대로 넣고 돌렸다:

```text
GOARCH=386 go build ./...
./main.go: invalid array length alertPublishTimeout - 1
           (constant 3499999999 of int64 type time.Duration)
```

**14판까지 이 절은 정반대를 적었다**(*"32비트 타깃에서도 int32 범위 안이므로 GOARCH
의존이 없다"*). 그 문장이 참이었던 것은 상한이 1.3초였을 때다 —
`[alertPublishTimeout - 1]` = 1,299,999,999. 그것을 떠받친 논거는
*"transport < budget = 5e9이고 T는 그 1/3 아래"*였고, **`1/3`은 시도 3회의 제약이다**
(`3T + 2D ≤ 5s`). 14판이 시도를 1회로 내리는 순간 T는 예산 전체에 가까워질 수 있고
**그 논거는 죽었는데 결론만 남았다.**

**밀리초로 세면 같은 것을 잡으면서 자릿수가 사라진다.** 나노초 형태는 밀리초 형태에
**엄격히 포섭된다**: `[T - 1]`이 잡는 것은 `T ≤ 0`이고
`[T/time.Millisecond - 100]`이 잡는 것은 `T < 100ms`이므로 후자가 전자를 포함한다.  <!-- not-a-measurement: 자릿수 단언의 임계값이지 측정이 아니다 -->
`alertRetryDelay`도 같다(`D ≤ 0` ⊂ `D < 10ms`). 그러므로 **방어가 줄지 않는다** —
아래 표가 구성별로 그것을 보인다. 배열 길이는 3400·140·1499가 되어 어느 GOARCH에서도
표현된다.

**한 가지는 실제로 달라진다.** `alertOverheadReserve`는 나노초 형태가 `≥ 1ns`를
요구했고 밀리초 형태는 `≥ 1ms`를 요구한다. **더 엄격한 쪽이고 그것을 말없이 넣지
않는다** — 2배 규칙이 이미 reserve에 712ms를 요구하므로 이 강화는 실무에서 걸리지  <!-- not-a-measurement: 같은 유도 임계 — 356.1ms의 2배다 -->
않으며, 12판 측정표의 「reserve = 1ns (허용되는 최솟값) BUILD OK」 행은 이제
**FAIL**이다. 1ns는 "이름을 가진 몫"이 아니다.

**실측 — 아래 표의 모든 구성을 `go build -overlay`로 이 패키지에 넣어 돌렸다**
(저장소 무편집, go1.26.5). 메시지는 잘라내지 않았다.

> **구성 수를 산문에 적지 않는다.** 9판까지 여섯 곳이 그 수를 손으로 박았는데
> 같은 절의 표는 아홉 행이었다(9라운드 H-2) — 8라운드 H5가 한 행을 타입별 둘로 가르면서
> 손으로 박은 카운트가 낡았다. *"기대값을 박은 검사는 검사가 아니라 또 하나의 주장이다"*
> 가 이 파일의 원칙이고, 그 원칙은 검사 코드만이 아니라 산문에도 적용된다.
> 이제 `tools/check_values.py`의 `check_derived_counts`가 이 표의 행 수를 세어
> 산문의 "N 구성" 주장과 대조한다 — 그래서 이 문장은 수를 말하지 않는다.

### 13판의 다섯 번째 단언 — **`alertPublishAttempts`는 1이 아니면 컴파일되지 않는다**

`var _ [alertPublishTimeout - alertTransportBudget]struct{}` 는 **관계**를 강제한다.
회차가 하나일 때만 0이고, 둘 이상이면 음수다. 숫자를 박지 않으므로 timeout을 다시
고르더라도 살아남는다. **12라운드 P1이 찾은 역행 구성이 정확히 이 단언에 걸린다.**

> **⚠ 13판 초판은 이 절의 제목을 "엔진 루프는 많아야 한 번 publish한다"라고 적었다.
> 13라운드 P3이 그것을 기각했고 맞다.** 배열 단언이 보는 것은 **상수 두 개의 관계**이며,
> `newNotifier`가 그 상수를 실제로 `Attempts`에 대입하는지는 보지 않는다. 조립부에서
> `Attempts:` 줄을 지우면 이 단언은 그대로 통과하고 `deliver`의 B1(`notifier.go:343`)이
> `DefaultCriticalAttempts`(3)로 되돌린다. **배선을 지키는 것은 컴파일러가 아니라
> R1·R2·R3다.** 이 단언이 지키는 것은 "채택된 상수가 1회 설계와 모순되지 않는다" 하나이며,
> 그 이상을 이 단언에 기대어 주장하는 문장은 이 change 안에 있어서는 안 된다.

**적자마자 돌렸다** (`go build`, 별도 모듈, 상수만 옮겨 심음). **16판은 채택 상수
(1회·T=3.5s·D=150ms)로 다시 쟀고 `GOARCH`를 축으로 더했다** — 이 표의 이전 판은
1.3초 시절의 값이었다.

| 시도 수 | amd64 | 386 | 컴파일러 메시지 |
|---|---|---|---|
| **1 (채택)** | **BUILD OK** | **BUILD OK** | — |
| 2 | **FAIL** | **FAIL** | `invalid array length alertPublishTimeout - alertTransportBudget (constant -3650000000 of int64 type time.Duration)` · `invalid array length alertOverheadReserve / time.Millisecond - 1 (constant -2151 of int64 type time.Duration)` |
| 3 (12판까지의 채택) | **FAIL** | **FAIL** | `… alertPublishTimeout - alertTransportBudget (constant -7300000000 …)` · `… alertOverheadReserve / time.Millisecond - 1 (constant -5801 …)` |
| 0 | **FAIL** | **FAIL** | `invalid array length alertPublishAttempts - 1 (untyped int constant -1)` · `… alertPublishTimeout - alertTransportBudget (constant 3650000000 …)` |

**시도 2·3회가 이제 메시지 둘을 낸다.** 13판의 표는 예산 관계 단언 하나만 적었는데,
1회 설계에서 시도를 늘리면 transport가 예산을 넘으므로 **reserve 단언도 함께 깨진다**.
잡는 것이 늘어난 게 아니라 13판이 한 줄만 적었다.

**단위 누락 단언이 시도 1회에서도 살아 있는지 확인했다**: `alertPublishTimeout = 3500`
(단위 없음, 시도 1회) → `invalid array length alertPublishTimeout / time.Millisecond - 100
(constant -100 of int64 type time.Duration)`. 회차를 줄여도 자릿수 검사는 그대로다.

### 16판의 단언 다섯 (측정 기록 — 채택 상수 기준, GOARCH 둘)

**전 구성을 `GOARCH=amd64`와 `GOARCH=386` 둘로 돌렸고 판정이 15/15 일치한다.**
그래서 결과 칸이 하나다 — 갈라지는 칸이 없다는 것이 이 표의 결론이다.

| 구성 | 결과 | 컴파일러 메시지 |
|---|---|---|
| **채택 #9 (1회·T=3.5s·D=150ms)** | **BUILD OK** | — |
| T=4.2s (reserve 800ms) | **BUILD OK** | — |
| T=4.288s (reserve 712ms — 2배 규칙의 경계) | **BUILD OK** | — |  <!-- not-a-measurement: 구성 라벨의 유도값 — 5.0s − 712ms이고 712ms는 356.1ms의 2배다 -->
| **T=5s (reserve 0)** | **FAIL** | `invalid array length alertOverheadReserve / time.Millisecond - 1 (constant -1 of int64 type time.Duration)` |
| T=5.1s (reserve −100ms) | FAIL | `… alertOverheadReserve / time.Millisecond - 1 (constant -101 …)` |  <!-- not-a-measurement: 반증 구성의 설정값이지 측정이 아니다 -->
| `alertRetryDelay = 0 * time.Millisecond` | **FAIL** | `invalid array length alertRetryDelay / time.Millisecond - 10 (constant -10 of int64 type time.Duration)` |
| `alertRetryDelay = 0` (맨 정수) | **FAIL** | `… alertRetryDelay / time.Millisecond - 10 (constant -10 of int64 type time.Duration)` |
| `alertPublishTimeout = 0 * time.Millisecond` | **FAIL** | `invalid array length alertPublishTimeout / time.Millisecond - 100 (constant -100 of int64 type time.Duration)` |
| `alertPublishAttempts = 0` | **FAIL** | `invalid array length alertPublishAttempts - 1 (untyped int constant -1)` · `… alertPublishTimeout - alertTransportBudget (constant 3650000000 …)` |
| **`alertPublishTimeout = 3500` (단위 누락)** | **FAIL** | `… alertPublishTimeout / time.Millisecond - 100 (constant -100 …)` |
| **`alertRetryDelay = 150` (단위 누락)** | **FAIL** | `… alertRetryDelay / time.Millisecond - 10 (constant -10 …)` |
| **`alertPublishTimeout = 3500 * time.Microsecond` (단위 오타)** | **FAIL** | `… alertPublishTimeout / time.Millisecond - 100 (constant -97 …)` |
| **`alertRetryDelay = 150 * time.Microsecond` (단위 오타)** | **FAIL** | `… alertRetryDelay / time.Millisecond - 10 (constant -10 …)` |

> **굵은 넷이 10판이 더한 것이다** — 9라운드 H-1. 9판의 단언 넷은 **나노초만 세므로
> 3500ms와 3500ns를 구별하지 못했고**, 단위를 빠뜨린 구성이 초록 빌드로 통과했다.  <!-- not-a-measurement: 상수 리터럴이지 측정이 아니다 -->
> 그 빌드에서는 모든 publish가 다이얼 전에 기한을 넘겨 **매 알림이 소진되고 매 알림이
> 진입 게이트를 래치한다** — 9판의 논증이 막겠다던 `0`보다 나쁘다. 자릿수를 세는
> 배열들이 단위가 지고 있던 부분을 본다.

> **⚠ 16판에서 세 행이 달라졌다. 셋 다 여기 적는다 — 표는 이겼다고 적는 자리가 아니다.**
>
> 1. **`reserve = 1ns`가 BUILD OK에서 FAIL이 됐다.** 밀리초 단언이 `≥ 1ms`를 요구한다.
>    의도한 강화이고 위 절에 사유를 적었다.
> 2. **`alertRetryDelay = 0` (맨 정수)의 메시지가 typed 형태로 바뀌었다.**
>    나노초 형태는 `untyped int constant -1`을 냈고 밀리초 형태는 `0 / time.Millisecond`가
>    이미 `Duration`이므로 **언제나** `of int64 type time.Duration`을 낸다.
>    **8라운드 H5가 가른 typed·untyped 두 행은 이제 메시지가 같다** — 둘 다 FAIL이고
>    둘 다 상수 이름을 말하므로 방어는 그대로지만, 그 구별을 근거로 쓰는 문장은
>    이제 근거가 없다.
> 3. **`T=1.6s`·`후보 #1`·`M-4 대안` 행을 뺐다.** 셋 다 1.3초 시절 후보의 구성이고,
>    같은 것을 보는 자리는 `T=5s`·`T=5.1s`가 채택 상수 기준으로 대신한다.  <!-- not-a-measurement: 반증 구성의 설정값이다 -->
>    **뺀 것을 말없이 빼지 않는다.**

**메시지가 상수 이름을 말한다** — 그것이 이 change가 테스트 대신 배열을 쓰는 이유다
(아래). **어느 행도 잘라내지 않았다.**

**타입 부분은 상수를 *어떻게 쓰느냐*에 달렸었고, 16판에 그 구별이 사라졌다**
(8라운드 H5 → 15라운드 A T1). 나노초 단언 아래에서는 `0 * time.Millisecond`로 쓰면
`time.Duration`이라 따옴표 붙은 `"time".Duration`이, **맨 `0`으로 쓰면 무타입 정수**라
`untyped int constant -1`이 나왔다. **밀리초 단언에서는 둘이 같다** —
`0 / time.Millisecond`가 이미 `Duration`이므로 어느 쪽으로 쓰든
`of int64 type time.Duration`이다(위 표 ⚠ 2번, 실측).
`alertPublishAttempts`만 `obs.DefaultCriticalAttempts`(무타입 상수)라 여전히 후자다.
**어느 쪽으로 쓰든 FAIL이므로 방어는 그대로 성립한다** — 없어진 것은 두 표기를
메시지로 가르는 능력이고, **그 구별을 근거로 삼는 문장이 이 change에 남아 있으면
안 된다.** 8판이 행 라벨과 메시지를 어긋나게 실었던 자리이기도 하다.

**배열 길이가 int 범위를 벗어날 수 있다 — 14판까지 이 문단은 정반대를 적었다**
(15라운드 A T1). 16판의 밀리초 단언에서 가장 큰 것은
`[alertOverheadReserve/time.Millisecond - 1]` = **1499**이고, 예산이 5초인 한
어떤 구성에서도 5000을 못 넘으므로 **어느 GOARCH에서도 표현된다** — 원소가 `struct{}`라
배열 자체의 크기는 0이다.

> **⚠ 나노초 단언 아래에서는 이 문단이 거짓이었다.** `[alertPublishTimeout - 1]`은
> 채택 상한에서 **3,499,999,999**로 int32를 넘고, `GOARCH=386`에서 **채택 구성 자체가
> 컴파일되지 않았다**(위 절의 실측). 14판이 근거로 든 *"T는 예산의 1/3 아래"* 는
> **시도 3회의 제약**(`3T + 2D ≤ 5s`)이며, 같은 판이 시도를 1회로 내리면서 그 제약이
> 사라졌는데 결론만 남았다. 전제가 죽고 결론이 살아남은 자리를 이 change 안에서
> 세는 중이라면, 이것이 그중 하나다.

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

> **⚠ 13판이 시도를 1회로 줄여 아래 표의 `a092` 열을 전부 바꿨다.** 12판까지의 이 표는
> 3회를 전제해 `4.2초`와 `3회 실패`와 `2~3건 중복`을 적었다. 13라운드 P6이 그 잔존을
> 짚었다. **한 번의 `Notify` 안에서 publish는 많아야 한 번이다.**

| publish의 실제 거동 | 오늘(10s/시도·3회) | a092(1.3s/시도·**1회**) | 차이 |
|---|---|---|---|
| 빠른 실패 (DNS NXDOMAIN·connection refused·non-2xx) | 3회 실패 | **1회 실패** | 시간만 줄어듦 |
| 타임아웃 (블랙홀) | 30s 쓰고 실패 | **`alertPublishTimeout`** 쓰고 실패 | 없음 (시간만 줄어듦) |
| (1.3s, 10s] 구간에서 **성공** | 성공 | **실패** | **손실 ①** |
| 서버는 **접수**했는데 응답이 1.3s 안에 안 옴 | 대개 성공 | **실패로 기록** | **손실 ②** |

### 손실 ② — 2판이 빠뜨린 모드

`client.Do`(`ntfy.go:125`)가 기한 오류를 돌려주면 `deliver`는 `MarkAlertAttemptFailed`를
쓴다. 13판에서는 **그 자리에서 재시도하지 않는다** — 시도가 1회이므로 `attempt < attempts`
(`notifier.go:387`, B9)가 거짓이고 루프가 끝난다. 메시지는 이미 나갔으므로:

- outbox 행은 PENDING으로 남는다
- `Gate.Block(ReasonAlertUndelivered)`가 래치된다 — **실제로는 배달된 알림에 대해**
- 운영자는 **한 관측당 1건씩 중복 수신**한다. 12판까지 이 줄은 "한 호출 안에서 2~3건"
  이었고 그것은 3회 설계의 수다. 13판에서 중복의 단위는 호출이 아니라 **관측**이다 —
  PENDING 행은 remind 창 없이 곧바로 다시 owed이기 때문이다(`outbox.go:277-279`).
  **총량은 줄지 않는다. 시간축으로 흩어질 뿐이다.**
- `Flush` 프로덕션 호출자가 없으므로 **아무도 정정하지 않는다**

1회 상한을 10초에서 1.3초로 줄이면 이 모드는 **더 있음직해진다.**

#### 늦은 서버에 무엇이 도착하는가 — 측정 (13라운드 P4)

"서버가 받았다"를 핸들러 진입으로만 세면 ntfy가 실제로 그 메시지를 발행했을지 알 수 없다.
그래서 핸들러가 **요청 본문과 `Title` 헤더를 읽어 기록**하게 하고 다시 쟀다.

```text
발송 1 · 서버가 본 것 ["exit judgement refused|A092-PAYLOAD-MARKER"]
```

기한을 4배 넘겨 응답하는 서버에도 **제목과 본문이 온전히 도착한다.** 손실 ②는 "메시지가
안 갔다"가 아니라 "간 메시지를 안 갔다고 기록한다"이다. 이 구분이 손실 ②의 전부다.

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

#### 그 n=1을 이 설계가 어디까지 떠받치는가 (13라운드 P8)

**1.3초는 지연 분류기다.** 그보다 느린 응답을 "실패"로 부르고, 그 판정이 게이트 래치와
`ENTRY_BLOCKED` 행을 만든다. 그 **기록**은 재시작을 넘어 살아남는다 — 같은 프로세스의
핸들이 아니라 **닫았다가 다시 연 원장**에서 읽어 확인했다.

```text
재개봉 후 mode="ENTRY_BLOCKED" · PENDING 1
```

**그러나 강제는 살아남지 않는다**(13라운드 보이스 B, 위 D2의 ⚠). 새 게이트는
`gate.Blocks()=0`이고 원장의 모드를 다시 읽는 프로덕션 경로가 없다.

그래서 이 분류기가 틀렸을 때의 대가는 **영속적이지 않고 프로세스 수명만큼**이다.
거짓 차단은 재시작으로 풀린다. **13판 초판이 여기 적었던 "대가는 영속적이다"는 철회한다.**
그래도 미검증은 그대로다 — 임계값을 정한 냉 표본은 **1건**이고, 프로세스가 사는 동안
새 진입이 거짓으로 막히는 것은 그대로 일어난다.

**"보수 방향이니 괜찮다"는 이 문제의 답이 아니다.** 방향이 보수적이라는 것은 틀렸을 때의
**피해 종류**에 대한 말이고, **틀릴 확률**에 대한 말이 아니다. 냉 1건은 확률에 대해 거의
아무것도 말하지 않는다. 이 설계는 그 미검증을 **감수하고 진행한다** — 근거는 반대편의
비용(34초 동기 체류)이 확률적이지 않고 결정적이라는 것 하나뿐이다. 그 외의 정당화는 없다.
관측을 늘리는 절차는 tasks 9.8이 정한다.

## D5 — 균일성의 대가를 경로별로

**조립부는 하나다.** `newNotifier`(`exitwiring.go:71`)의 산물은 `gateway.go:280`에서 만들어져
엔진 전체가 공유한다. 그러므로 예산은 exit 루프만이 아니라 **이 `Notifier`가 다루는 모든
critical 알림**에 걸린다.

> **⚠ 13판까지 이 자리의 표는 거짓이었다 — 14라운드 보이스 B가 기각했고, 소스에서
> 확인했다. 틀린 것은 나다.** 두 군데가 틀렸다.
>
> 1. **입구를 하나로 셌다.** `.Notify(`만 grep하면 두 번째 공개 입구인
>    `AnnounceOperatingMode`(`mode.go:49`, `journal.ModeAnnouncer` 구현)가 통째로
>    빠진다. 그 입구는 원장이 부르고(`operating_mode.go:479`), 도달 경로는
>    `.Notify(` 목록에 **하나도 나오지 않는다.** 그리고 `mode.go:57`은 호출자가
>    아니라 그 입구의 **내부 홉**인데 호출자로 적혀 있었다.
> 2. **"다음 관측이 다시 보낸다"가 대부분의 경로에서 거짓이다.** `runtime.go:453`
>    하나만 다음 관측이 없다고 적었는데, exit 루프의 critical 알림 **넷이 전부
>    래치를 먼저 세우고 알린다.** 래치가 서면 다음 관측은 `alert`에 닿기 전에
>    early return한다 — 재발송은 없다.

**입구는 둘이고, 두 번째 입구가 표에서 빠져 있었다.**

| 입구 | 정의 | 프로덕션 도달 경로 |
|---|---|---|
| `Notifier.Notify` | `notifier.go:124` | `exitwiring.go:103,141` · `exitloop.go:1604` · `runtime.go:453` · `reconcileloop.go:556` · `flatten.go:694` |
| `Notifier.AnnounceOperatingMode` | `mode.go:49` | 원장이 `operating_mode.go:479`에서 부른다. 승격 site 다섯 중 `Announcer`가 배선된 것 — `exitloop.go:796` · `runtime.go:415` · `retry.go:385`(자격증명 거부) · `riskguardian.go:644`(일일 손실 한도) |

**재발송이 있는가 — 경로별로.** 이 열이 D5의 전부다. 없으면 1회 실패는 곧
**운영자가 영영 모른다**는 뜻이다.

**critical 알림만 센다** — 등급 정본은 `obs/event.go`의 `criticalEvents`(`:279-298`)이고,
기계 전수는 tasks 7.5의 명령 (4)다.

> **⚠ 14판의 이 표도 틀렸다 — 15라운드 보이스 A가 기각했고 소스에서 확인했다.**
> 세 가지가 틀렸다. (1) **normal을 critical 표에 넣었다.** `exitwiring.go:103,141`이
> 내는 `EventExitPositionUnmanaged`·`EventExitPositionClosedExternally`는
> `criticalEvents`(`event.go:279-299`)에 **없다**. (2) **funnel을 산출 지점으로 적었다.**
> `reconcileloop.go:556`은 `d.alert` 헬퍼 안의 `Notify` 줄이고, 그 뒤의 산출 지점은
> `adoption.go:353,417,455` 셋인데 **셋 다 normal이다** — 대사 funnel에는 critical이
> 하나도 없다. (3) **critical 생산자 둘을 한 행에 뭉갰다.** `runtime.go:453`은 funnel이고
> 생산자는 `:316`(`EventEngineLoopFailed`)과 `:396`(`EventEngineLoopDegraded`)이다.
> 열화 쪽에 "프로세스가 죽는 중"은 **거짓이다** — 본문이 "the loop is still running and
> still retrying"이라고 적고 있고, 재발송을 막는 것은 죽음이 아니라 `:383` `takeLatch`다.

| 알림 (critical) | 산출 지점 | 래치 | 재발송 |
|---|---|---|---|
| 관측 두절 | `exitloop.go:767-791` `checkOutage` | `:776` `o.outageRaised = true` **→ `:780`** `o.alert` — **bool 필드다**(map 아님) | **없다** |
| **판정 불가 — 손절도 평가되지 않는다** | `exitloop.go:1516-1538` `alertRefused` | `:1520` `o.refused[id] = true` **→ `:1526`** `o.alert` | **없다** (`clearRefused` 전까지) |
| 판정 격리 | `exit_quarantine_announce.go:53-90` | `:70` `o.quarantineAnnounced[key] = true` **→ `:71`** `o.alert` | **없다** |
| 청산 지연 | `exitloop.go:1567-1595` `noteDelay` | `:1579` `o.delayAlerted[id] = true` **→ `:1580`** `o.alert` | **없다** (`clearDelay` 전까지) |
| 운영 모드 전이 | `AnnounceOperatingMode`(`mode.go:49`) | `direction == 0`이 `operating_mode.go:415`에서 early return — **투영(`:475`)보다 앞이다** | **없다** |
| 엔진 루프 **실패** | `runtime.go:316-317` → funnel `:453` | — 프로세스가 종료 중이다 | **없다** |
| 엔진 루프 **열화** | `runtime.go:396-397` → funnel `:453` | `:383` `takeLatch(loop.Name)` (회복 시 `:379` `clearLatch`) | **없다** — 그러나 **루프는 계속 돈다.** 막는 것은 래치다 |
| 청산 제안 거부 | `exitloop.go:1550` `alertProposalRefused` | **래치 없음** | **있다** — 5초 뒤 |
| flatten 3종 (`FlattenStarted`·`FlattenStalled`·`OrderInDoubt`) | `flatten.go:186,278,403` → funnel `:694` | — | **발송 자체가 없다.** 유일한 프로덕션 리터럴(`cmd/tossctl/flatten.go:247-263`)이 `Notifier`를 안 채운다 → `s.Notifier == nil`이라 `logf`로 빠진다 |

**재발송이 있는 critical은 하나뿐이다**(청산 제안 거부). 없는 것이 일곱이고,
하나는 아예 나가지 않는다. 12·13판의 "하나", 14판의 "셋/여섯"은 둘 다 틀렸고,
**틀린 방향은 언제나 같았다 — 대가를 실제보다 작게 셌다.**

> **normal 등급도 같은 기한을 쓴다.** `Notify`는 `:128`에서 갈라져 normal을
> `publishBestEffort`(`:155`)로 보내는데, 그 경로는 **outbox도 재시도 루프도 타지
> 않는다** — `Publisher.Publish` 1회다. 그러므로 `Attempts`·`RetryDelay`는
> **critical 경로에만** 걸리고, `Ntfy.Timeout`은 **양쪽 다** 걸린다. normal 경로의
> 오늘 최악은 34초가 아니라 **10초**이고 a092 후에는 3.5초다.
> 해당 산출 지점: `exitwiring.go:103,141` · `exitloop.go:1430`(제안 상한) ·
> `exitloop.go:1500`(미관리, `:1499`에서 래치) · `adoption.go:353,417,455`.

> **래치는 일곱이고 그중 critical이 다섯이다** — 7개 `o.alert` 자리 중 5개가 알리기
> **전에** 플래그를 세우고, 여기에 `takeLatch`(감독자 열화)와 `direction == 0`
> (운영 모드)이 더해진다. 등급이 normal인 래치는 `alertUnmanaged`
> (`exitloop.go:1499-1500`)뿐이다.
>
> **어떤 단일 패턴도 이 다섯을 다 잡지 못한다.** `] = true`는
> `o.outageRaised = true`(`:776`)를 **놓친다** — 그것은 map 항목이 아니라 bool
> 필드다. 14판은 이름 목록으로 세다 `alertUnmanaged`를 빠뜨렸고, 15판은 일반형
> 패턴으로 세다 `o.outageRaised`를 빠뜨렸다. **tasks 7.5는 두 패턴을 다 돌려야 한다.**

| 경로 | 등급 | 오늘 | a092 후 (**1회 · 3.5s**) | 평가 |
|---|---|---|---|---|
| **래치되는 critical 넷** (두절·판정 불가·격리·청산 지연) | critical | 34s | **3.5s** | **한 방이다.** 실측 최대 **2.721s** 위 **1.29배** 여유로 그 한 방을 산다 — 그리고 **그 실측은 POST가 아니라 GET이다**(§아래·`analysis/delivery-latency.md` §9.1.1) |
| 청산 제안 거부 | critical | 34s | **3.5s** | **이득.** 재발송이 있는 유일한 critical |
| 감독자 실패·열화 | critical | min(30s, 34s) = 30s | **3.5s** | 한 방(둘 다). `alertDeliveryBound`가 30s라 **오늘 세 번째 시도는 애초에 끝나지 못한다** |
| 감독자 승격 (`runtime.go:415`) | critical | 34s, **기한 없음** | 3.5s, 기한 없음 | 기한이 없다는 사실은 a092가 고치지 않는다 |
| Announcer 승격 (`retry.go:385`·`riskguardian.go:644`) | critical | 34s | 3.5s | 한 방. 오늘 발동 여부는 `analysis/notify-reach.md` |
| exit 관측 결과 (`exitwiring.go:103,141`) | **normal** | **10s** | **3.5s** | 이득. 재시도 루프를 안 타므로 오늘도 34s가 아니다 |
| 제안 상한·미관리 (`exitloop.go:1430,1500`) | **normal** | **10s** | **3.5s** | 이득 |
| 대사/편입 (`adoption.go:353,417,455`) | **normal** | **10s** | **3.5s** | 이득. 이 funnel에 critical은 없다 |
| flatten 3종 | critical | — | — | `Notifier` nil이라 **오늘 아무 데도 안 간다** (`cmd/tossctl/flatten.go:247-263`) |

### 래치되는 넷이 이 change의 상수를 정한다

**"1회"의 진짜 대가는 여기다.** 재발송 없는 경로에서 `alertPublishTimeout`은 체류
상한이 아니라 **배달 확률의 임계값**이다. 그보다 느린 응답은 배달되지 않고, 다시
보낼 것이 없다.

그리고 이 넷 중 하나는 `exit.judgement_refused` — **손절도 평가되지 않는다**는
알림이며 이 시스템에서 가장 중요한 한 건이다.

**#8(1.3s)이 기각된 진짜 이유가 이것이다.** 실측 전송 지연 평균 1.875s 아래에
임계값을 두면 그 알림은 **오늘은 배달되는데 a092 후에는 배달되지 않는다.**
그것은 안전 불변식 3(토글 OFF = upstream)과 6(보수 방향만)을 동시에 깬다.
후보 #9(3.5s)는 실측 최대 2.125s 위 65% 여유를 두어 **오늘 배달되는 것은 계속 배달된다** —
위 §후보의 1.9s·2.125s 행이 그 측정이다.

**남는 미검증은 그대로다**: 표본은 8건이고 homepage GET이다. 3.5s를 넘는 응답이
관측되면 이 넷은 조용히 실패한다. 그 재검토 절차가 tasks 9.8이고, 근본 해결은
**루프 밖 재전송**(`Flush` 배선)이며 a092의 범위 밖이다 — 그리고 그 후속 change는
**아직 만들어지지 않았다**(아래 §후속).

### P3의 대가는 13판에서 **커졌다** — 그리고 그것을 보상하는 것이 a092에는 없다

12판까지 이 자리는 "4.2초, 시도 3회"였다. 13판에서 종료 알림은 **단 한 번, 1.3초** 안에
나가야 한다. 위 표의 마지막 열이 그 차이를 만든다 — **P3에는 다음 관측이 없다.** 다른 모든
경로는 실패한 알림을 다음 주기가 다시 올리지만, 종료 알림을 다시 올릴 주기는 없다.

그래서 12판의 "느리지만 살아 있는 transport라면 6초에 배달됐을 것"은 13판에서 이렇게 읽어야
한다: **`alertPublishTimeout`을 넘겨 응답하는 transport에서 종료 알림은 나가지 않고, 이 change 안에는 그것을
다시 보낼 것이 없다.** `Flush`는 프로덕션 호출자가 없고(미배정 후속 범위), 미배정 후속은 아직 없다.

> **⚠ 13판 초판은 여기에 보상을 적었고, 그 보상은 없다.** 초판의 문장은
> *"게이트가 래치되고 `ENTRY_BLOCKED`가 재시작을 넘어 남으므로, 운영자가 아는 경로가
> 알림에서 **재시작해도 막혀 있는 진입**으로 바뀐다"*였다. **13라운드 보이스 B가
> 기각했고 재현으로 확인했다**(D2의 ⚠) — 새 게이트는 아무것도 막지 않고, 원장의
> 모드를 다시 읽는 프로덕션 경로가 없다. **바뀐 것이 아니라 없어진 것이다.**

**그래서 P3에 남는 것은 이것뿐이다 — 보상이라 부르지 않는다.**

- 알림은 outbox에 PENDING으로 **durable하게 남는다** — 사라지는 것이 아니라 안 나갈 뿐이다.
  다만 **그 행을 다시 보내는 것도 없다**(`Flush` 무배선)
- 그 행의 **시도별 실패 기록**(`MarkAlertAttemptFailed`)도 남는다 — 몇 번 시도했는지는
  세지지만 **세는 것으로 나가지는 않는다**
- `ENTRY_BLOCKED` 행과 구조화 로그가 원장에 남는다. 그 행을 쓴 것은
  `EscalateOperatingMode` 트랜잭션이고 — **읽는 것은 사람뿐이다**
- 게이트 래치는 **죽는 중인 프로세스 안에서만** 산다
- 종료 자체는 막지 않는다 — `alertDeliveryBound`가 상위 기한을 잡고 있다

**그러므로 정직한 문장은 이것이다: 종료 중 전송기가 `alertPublishTimeout`을 넘기면, 운영자에게 가는
자동 신호는 없다.** 오늘도 전송기가 완전히 죽어 있으면 같지만, a092는 **(1.3초, 10초]
구간을 그 상태에 새로 편입시킨다.**

**그래도 진행하는 근거**는 비교의 종류다. 이쪽 손실은 **관측되지 않은 지연 구간의 확률**
(냉 표본 6건 중 그 구간 0건, 최댓값 0.754초)이고, 34초 동기 체류는 **transport가 죽는
즉시 매 사이클 결정적으로** 발생하며 손절 판정이 그 뒤에 줄을 선다. 안전 불변식 4가
가리키는 쪽은 후자다. **그리고 이 선택은 미배정 후속에 두 가지 의무를 남긴다** — `Flush` 배선과
**`RestoreOperatingModeProjection` 배선**. 후자는 a092가 발견한 것이지 만든 것이 아니다.

**없는 것을 있다고 하지 않는다**: a092는 P3에 대해 재전송 수단을 만들지 않는다. 만들 수
있는 자리는 두 곳뿐이고 둘 다 이 change의 범위 밖이다 — 경로별 시도 횟수(`obs.Notifier`에
**필드** 추가 + 호출자 6곳이 저마다 값을 고름)와 루프 밖 재전송(`Flush` 배선 = 미배정 후속).

> **19판이 이 줄의 괄호를 고쳤다 (18라운드 B-P12의 두 번째 사본).**
> 여기 있던 근거는 *"`obs` 패키지 편집, a092가 하지 않기로 한 것"*이었고 18판이 이미
> 뒤집은 전제다 — a092는 `claimAndDeliver`·`Flush`·`publishBestEffort`를 편집한다(D0.10).
> **B-P12는 `proposal.md`의 한 줄로 보고됐지만 같은 값이 여기에도 있었다.**
> 리뷰가 준 file:line만 고쳤으면 이 사본이 살아남아 다음 판의 범위 판단이 다시
> 없는 규칙에 기댔을 것이다. 진짜 근거는 **결정의 종류**다: 재설계는 한 예산을
> *언제* 쓰는가를 바꾸고, 경로별 시도 횟수는 예산이 *몇 개*인가를 바꾼다.

### 발견: 영속 승격이 프로덕션에서 강제되지 않는다 (a092 범위 밖, 미배정 후속 의무)

| 경로 | 정의 | 프로덕션 호출자 | 테스트 호출자 |
|---|---|---|---|
| `Journal.CurrentOperatingMode` | `operating_mode.go:553` | **없음** | — |
| `Journal.RestoreOperatingModeProjection` | `operating_mode.go:574` | **없음** | 3 |
| `Journal.SetModeProjector` | `operating_mode.go:289` | **없음** | 18 (파일 5개) |

`EscalateOperatingMode`가 쓰는 `ENTRY_BLOCKED`는 **산 프로세스에서도 재시작 후에도
아무것도 막지 않는다.** `risk.checkOperatingMode`(`chain.go:261`)는 `Input.Account.Mode`를
읽는데 그 값을 원장에서 채우는 프로덕션 경로가 없다 — 엔진 경로에서 유일하게 잡히는
생산자는 `tracer.go:473`이고 거기서는 `risk.ModeNormal`이 **상수로 박혀 있다.**

> **19판이 이 표의 판정을 넓혔다 (18라운드 A-P4의 답).**
>
> 원래 문장은 *"재시작 후 아무것도 막지 않는다"*였고, 그것은 공백을 재시작 창으로
> 한정한다. 측정은 그보다 넓다. `TransitionOperatingMode`는 커밋 **직후** projector를
> 부르고(`operating_mode.go:475-476`), bind된 projector가 없으면 그 호출이 아무 일도
> 하지 않는다. `ReasonOperatingModeBlocked` 래치를 세우는 자리는 `modegate.go:50`
> 하나뿐이고 그 함수는 projector 인터페이스로만 불린다. 그러므로 승격은
> **재시작을 기다릴 것도 없이 처음부터** 진입 게이트에 닿지 않는다.
>
> **이것이 A-P4를 뒤집는다.** 18라운드 A-P4는 *"`RuntimeOptions`에 `EntryGate`가 없어
> 배달 루프 사망 후 게이트를 잠글 주체가 없다"*고 적었다. 결론은 맞고 **처방이 틀렸다** —
> 필드를 하나 더해도 부족하고, 더할 필요도 없다.
>
> | | A-P4의 처방 | 측정이 말하는 것 |
> |---|---|---|
> | 진단 | `RuntimeOptions`에 `EntryGate` 필드가 없다 | 맞다 (`runtime.go:132-161`) |
> | 함의 | 그래서 게이트를 잠글 수 없다 | **부분만** — `Runtime`은 `Escalate`를 갖고 있다 |
> | 왜 그래도 못 잠그나 | (안 적힘) | `Escalate` → projector 경로가 **프로덕션에서 bind되지 않는다** |
> | 처방 | 필드를 더한다 | 필드를 더해도 `Runtime.escalate`는 여전히 원장만 쓴다. 배선이 문제고 배선은 **범위 밖** |
>
> **그래서 a092가 R18-1에서 기대는 것은 escalation이 아니다.** 배달 루프가 죽었을 때
> 이 프로세스의 진입을 실제로 멈추는 것은 `Notifier`가 **이미 들고 있는** `Gate`
> (`exitwiring.go:77`에서 배선됨)의 `Block(ReasonAlertUndelivered, …)`뿐이다.
> R18-1의 관측 (b)는 그 래치를 보고, (c)의 `ENTRY_BLOCKED` 승격은 **원장 행의 존재**를
> 보는 것이지 그것이 진입을 막는다는 주장이 아니다. 두 관측이 한 문장에 있으면
> 없는 강제력을 있는 것으로 읽게 된다 — tasks §6.0의 R18-1 줄에 이 구분을 적었다.
>
> `runtime.go:400-402`의 알림 본문은 *"New entries are blocked until the loop recovers"*
> 라고 적는다. 위 측정대로면 **그 문장은 오늘 프로덕션에서 참이 아니다.** a092는 이
> 문장을 고치지 않는다(`obs` 델타가 아니라 `engine` 델타이고 R18-1의 관측 대상도
> 아니다). **침묵한 생략이 아니라 여기 적어 두는 것**이고, 배선 change가 이 문장을
> 같이 들고 가야 한다.

**a092는 이것을 고치지 않는다** — 알림 예산의 change가 진입 게이트 복원 배선을 함께 들고
가면 되돌리기 단위가 둘이 된다. 대신 **의존하지 않는다**: 위에서 이 사실에 기대던 문장을
전부 걷어냈다. 배선은 미배정 후속이고, 그때까지 **자동 차단은 원장 행일 뿐**이라는 것이
현재 상태다.

### `runtime.go:459`의 주석은 a092가 고친다

```go
// alertDeliveryBound is how long the runtime waits for an alert it raises while
// stopping. Generous enough for obs.Notifier's three bounded publish attempts,
// finite so a dead transport cannot hold the shutdown open.
const alertDeliveryBound = 30 * time.Second
```

12판은 이 주석의 수정을 미배정 후속으로 미뤘다. 13판은 **a092가 한다.** 이유는 범위가 아니라
**인과**다 — "three bounded publish attempts"를 거짓으로 만드는 것이 a092의 편집이다.
자기가 만든 거짓 진술을 다음 change에 넘기는 것은 침묵한 생략이다. 주석 한 줄이고
동작은 0비트도 바뀌지 않는다(tasks 8.5).

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

> **⚠ 19판이 이 표를 다시 썼다. 이 표는 18판 이후 High-risk 선언으로서 거짓이었다.**
>
> D7은 **Pre-Edit 선언**이다 — 안전 불변식 5가 요구하는, "무엇을 건드리는가"의 정본.
> 그런데 18판이 D0.10에서 `claimAndDeliver`·`Flush`·`publishBestEffort` 편집과
> 배달 루프 신설을 a092 안으로 끌어오는 동안 **이 표는 17판 상태로 남았다.**
> "안 건드리는 것: `obs` 전부"는 D0.10의 표 세 번째 줄과 정면으로 모순된다.
>
> 18라운드 B-P12는 이 모순을 `proposal.md`의 한 줄로 보고했다. 그 한 줄만 고쳤으면
> **같은 거짓이 High-risk 선언 안에 남았을 것이다** — 그리고 그것이 남은 자리는
> 제안서의 non-goal 목록이 아니라 **폭발 반경을 선언하는 표**다. 리뷰가 준
> file:line이 아니라 **값**을 쫓아서 찾았다(사본 셋 중 셋째).

| 항목 | 내용 |
|---|---|
| 편집 대상 | **① 조립부** — `newNotifier`(`exitwiring.go:71-81`), `resolveNotificationPublisher`(`notifications.go`의 `&obs.Ntfy{...}` 리터럴), **`runtime.go`의 `alertDeliveryBound` 주석 한 줄**(13판 — D5) · **② `internal/obs/notifier.go` 함수 본문** — `claimAndDeliver`·`Flush`·`publishBestEffort` (18판 D0.10·D0.3a·D0.3b) |
| 편집 성격 | ①은 **구조체 리터럴 필드 추가 + 상수·import 선언 + 주석 문구**이고 조건문·이탈·호출을 더하지 않는다. **②는 그렇지 않다** — 잠금이 덮는 구간을 바꾸고(claim 아래에서 release → publish → 재취득 → 정산) 새 이탈을 만든다 |
| 무변화 증명 | ①에만 해당한다: 편집 후 AST가 `newNotifier` branches 0/returns 1/calls 0, `resolveNotificationPublisher` branches 5/returns 4 유지. `runtime.go`는 **주석만** 바뀌므로 `Runtime.alert`의 AST가 편집 전과 바이트 단위로 같다. **②는 무변화가 아니라 변화 자체가 델타이므로 Function Logic Map + Branch Test Map으로 진다**(§6.0의 R18-1~R18-4·R19-1·R19-2) |
| 되돌리기 | **되돌리기 단위가 하나가 아니다.** ①은 세 필드를 지우면 정확히 오늘로 돌아온다. ②는 그렇지 않다 — 잠금이 덮는 구간이 되돌아가면 a096 round 1이 막은 이중 발송의 전제가 되살아난다. 스키마·설정·원장 변화는 둘 다 없음 |
| 안 건드리는 것 | `exitloop.go` 전부, `internal/journal` 전부, `internal/risk` 전부, `internal/execgw` 전부. **`obs`는 위 ②가 명시한 세 함수만** — `Notify`·`notifyCritical`·`deliver`·`escalate`·`wait`·`logEvent`는 안 건드린다. **배달 루프 등록과 `SupervisedLoop` 배선은 a098이 진다**(19판 — D0.10의 표). **`cmd/tossctl`은 프로덕션 코드를 안 건드린다 — 다만 6.9가 테스트 파일 하나를 신설한다**(4라운드 N5) |
| 토글 | **없다.** 없애려는 것이 upstream 동작(34초)이므로 OFF 상태가 곧 결함이 된다. §0.6("명확한 근거가 있는 보수 방향")으로 정당화하고 근거는 `analysis/delivery-latency.md` |
| 손절 불변식 | 안전 불변식 4: **②·③ 어느 것도 손절·비상 청산 경로에 들어가지 않는다.** 게이트 래치는 `ReasonAlertUndelivered`로 **진입만** 막고 exit는 건드리지 않는다(`notifier.go:255`의 선언 주석 — "Entries only. Exits are untouched"). D0.5가 이 대가를 따로 적는다 |

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
