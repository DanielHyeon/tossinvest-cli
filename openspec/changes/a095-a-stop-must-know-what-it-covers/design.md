# a095 · 설계

> 분기 인용은 전부 `analysis/function-logic/`의 AST 산출물에서 온다
> (함수 9개 · 분기 98개 · 진입 62 · 미진입 27 · 자체 블록 없음 9).

## D0. 하나의 원칙

**보호의 기준은 얼려 둔다. 보호가 무엇을 덮는지는 매번 다시 잰다.**

두 문장이 충돌하지 않는 이유가 이 change의 전부다.

- 기준(`entry_price`·`initial_stop`·`initial_risk`)은 **이미 보고된 R의 분모**다.
  움직이면 그 포지션의 과거 보고가 전부 다른 뜻이 된다. TossOS가 그것을 설계로
  선언했고(`adoption.go:437-441`) 그 논거는 옳다.
- 그러나 **덮는 범위**는 기준이 아니라 사실이다. 수량은 앱에서 늘고, 보호 상태가
  아예 없는 보유도 생긴다. 그것을 매번 다시 재는 것은 분모를 건드리지 않는다.

**오늘 무너져 있는 것은 첫째가 아니라 둘째다.** 그리고 첫째의 정당성이 둘째에
의존한다 — 같은 주석이 *"Reporting it is what the operator can act on"*이라 쓴다.

## D1. R1 — 등급표에 한 줄

### 코드가 이 등급을 normal로 둔 이유 — **먼저 인용하고, 명시적으로 뒤집는다** (1라운드 §1.3)

**1판은 이 절을 쓰지 않았다.** `SeverityOf`의 주석만 인용하고, **그 표가 이 이벤트를 뺀
이유가 적혀 있는 자리를 지나쳤다.** 침묵한 생략이었다.

`internal/obs/event.go:190-194`, 대상 이벤트 자신:

> *"Normal — **somebody trading their own account by hand is not a malfunction.**"*

형제 이벤트(`:212-217`)가 기전까지 적는다:

> *"Normal, for the same reason the fold is: a person selling their own shares is not a
> malfunction, and **grading it critical would mean an engine with no alert transport
> configured stops opening positions every time its owner takes a profit by hand.**"*

**그 기전은 실재한다:**

```text
notifyCritical  notifier.go:152
  └─ deliver 실패 (3회 × 2초 대기)      DefaultCriticalAttempts=3 · DefaultRetryDelay=2s
       └─ Gate.Block(ReasonAlertUndelivered)        notifier.go:283-285
       └─ escalate → EscalateOperatingMode(...)     notifier.go:216-218
            └─ ModeEntryBlocked
               "new entries are blocked until an operator acknowledges the alert backlog"
                  └─ Acknowledge 로만 해제           notifier.go:343
```

**a095는 이 결정을 뒤집는다. 근거 셋을 적는다 — 지나친 것이 아니라 뒤집은 것이다.**

**(1) 그 주석이 말하는 사건과 a095가 올리는 사건이 다르다.**
`:212-217`의 논거는 **`EventExitPositionClosedExternally`**(소유자가 자기 주식을 팔았다)에
대한 것이고, 그것은 정말로 오작동이 아니다. **a095는 그 이벤트를 건드리지 않는다.**
`EventExitPositionUnmanaged`가 말하는 것은 *"가진 것에 손절이 걸려 있지 않다"*이며,
그것은 소유자의 행동에 대한 서술이 아니라 **엔진의 보호 상태에 대한 서술**이다.

**(2) 「수동 매매마다」가 성립하지 않는다.** 그 문장이 그리는 폭주는 매매 **횟수**에
비례하는 알림을 전제한다. 그러나 이 이벤트의 억제는 **상태**를 따른다 —
`alertUnmanaged`는 프로세스당 1회(`d.unmanaged`)이고, R2 후의 수량 알림은
**지난 보고보다 늘었을 때만** 운다. 수동 매매를 아무리 많이 해도 수량이 그대로면
한 번도 더 울지 않는다.

**(3) 막히는 것과 지켜지는 것의 교환이 명확하다.** `ModeEntryBlocked`가 막는 것은
**신규 진입뿐**이고 손절·익절·취소는 막지 않는다(§1.10-1 확인). 즉 최악의 경우는
「alert transport가 죽은 동안 새로 사지 못한다」이고, 그 반대편은
「가진 것에 손절이 없는데 아무도 모른다」다. **§4(손절 즉시성)가 이 교환을 결정한다.**

**그러나 대가는 대가다 — 그렇게 적는다.** alert transport가 죽고 무보호 보유가 있으면
**엔진은 신규 진입을 멈춘다.** 그것이 a095가 선택하는 상태이며, 운영 문서(tasks 7절)가
그것을 미리 적는다.

### 왜 함수가 아니라 map인가

`SeverityOf`(`internal/obs/event.go:309-314`, **분기 1개**)는 map 하나만 본다.

| 분기 | 자리 | 조건 | 진입 실측 |
| --- | --- | --- | --- |
| **B1** | `:310` | `if criticalEvents[t]` | 예 |

주석이 설계 의도를 명시한다:

> *An event type this build does not know is graded normal, not critical … Genuinely
> critical conditions are **named in the table above**.*

**따라서 고칠 곳은 표다.** `EventExitPositionUnmanaged`(`:200`)를
`criticalEvents`(`:279-298`, 현재 18종)에 등재한다. 함수 본문은 그대로 둔다.

### 그 한 줄이 무엇을 바꾸는가

`Notifier.Notify`(`notifier.go:107`, **분기 1개**):

| 분기 | 자리 | 조건 | 진입 실측 |
| --- | --- | --- | --- |
| **B1** | `:111` | `if severity != SeverityCritical` → `publishBestEffort` → `return nil` | 예 |

등급이 critical이 되면 B1을 **타지 않고** `notifyCritical`로 간다. 그것이 outbox를 쓰는
유일한 경로다.

`publishBestEffort`(`:138`, **분기 2개**)가 오늘의 종착지다:

| 분기 | 자리 | 조건 | 결과 |
| --- | --- | --- | --- |
| **B1** | `:139` | `if n.Publisher == nil` | **조용히 반환 — 로그도 없다** |
| **B2** | `:142` | `Publish(...) != nil && n.Log != nil` | 실패를 **로그에만** 남긴다 |

주석이 그 조용함을 계약으로 적는다 — *"treating its failure as an incident would make
the grading meaningless."* **이 함수는 결함이 없다.** 결함은 여기로 오는 이벤트다.

### 원장이 사슬을 닫는다

| | 실측 (2026-08-07T00:32:11Z) |
| --- | --- |
| `alert_outbox` 총 행 | 13 |
| severity 분포 | **전부 critical** |
| `exit.position_unmanaged` | **0** |
| 전달 시도(`attempts > 0`) | 4 / 13 |

**normal 등급 이벤트는 outbox에 단 한 행도 없다** — 그 경로가 행을 만들지 않기 때문이다.

### a091과의 관계 — 같은 계열, 다른 이벤트

a091은 `EventExitProposalCapped`를, a095는 `EventExitPositionUnmanaged`를 등재한다.
**같은 map을 건드리므로 병합 순서만 주의하면 되고, 논리적 의존은 없다.**

**a091이 적은 유보는 a095에서 성립하지 않는다.** a091 proposal은
*"8/2에 운영자가 한 번도 호출받지 못한 이유는 등급이 아니라 transport 부재"*이며
*"이 change의 이익은 운영자 호출이 아니라 원장에 남는 흔적 하나"*라고 적었다.

**2026-08-07 현재 transport는 있다** — outbox 13행 중 4행이 실제 전달을 시도했다.
따라서 a095의 이익은 **흔적 + 호출** 둘 다이며, 그렇게 주장한다.

### 발신 자리 넷 — 전부 같은 등급이 된다

| 자리 | 사실 |
| --- | --- |
| `adoption.go:418` | 편입도 진입 결정도 없는 보유 (`alertUnmanaged`) |
| `adoption.go:456` | 편입 후 수량 증가 (`checkExternalIncrease`) |
| `exitloop.go:1501` | 관측 루프가 본 무보호 |
| `exitwiring.go:104` | 배선 시점의 무보호 |

**넷 다 「가진 것에 손절이 안 걸려 있다」는 같은 사실을 말한다.** 등급을 이벤트 종류로
매기는 이상 넷은 같이 올라간다. 그것이 옳은지 D7에서 재검토한다.

## D2. R2 — 억제는 프로세스가 아니라 사실을 따른다

### 지금의 억제

`checkExternalIncrease`(`adoption.go:441-472`, **분기 3개 · 미진입 2개**):

| 분기 | 자리 | 조건 | 진입 실측 |
| --- | --- | --- | --- |
| **B1** | `:442` | `if d.grown[p.ID]` | **아니오** |
| **B2** | `:446` | `if err != nil` (`AdoptionOf` 실패) | **아니오** |
| **B3** | `:450` | `if err != nil \|\| cmp <= 0` | 예 |

**B1과 B2는 어떤 시험도 밟지 않는다.** 이 change가 바로 그 둘 위에 얹히므로
tasks 2.3·2.4가 그것을 먼저 덮는다.

### B1 — 프로세스당 한 번은 사실을 말하지 못한다

`d.grown`은 `map[string]bool`이고 프로세스 메모리다. 한 번 알린 뒤에는 같은 프로세스
안에서 수량이 3주→32주가 되어도 **다시 울지 않는다.** 재시작하면 초기화되므로
「몇 번 울었나」는 프로세스 수명의 함수이지 사실의 함수가 아니다.

**고치는 방식**: 억제 키를 **마지막으로 보고한 수량**으로 바꾼다.
`map[string]bool` → `map[string]string`(보고된 수량). 현재 수량이 그보다 크면 다시 알린다.

**왜 이것이 알림 폭주가 아닌가**: 조건이 「늘었나」가 아니라 「**지난번 보고보다** 늘었나」다.
수량이 그대로면 한 번도 더 울지 않는다. 관측 주기가 아무리 짧아도 마찬가지다.

### B2 — 편입 기록이 없으면 조용히 사라진다

```go
adoption, err := d.opts.Journal.AdoptionOf(ctx, p.ID)
if err != nil {                 // B2 :446
    return                      // 아무 말도 없이
}
```

**이 반환은 두 가지를 삼킨다**:

1. **편입 기록이 없는 보유** — 엔진이 직접 연 포지션, 그리고 미편입 보유.
   원장 실측: 010170 `instance_seq=2`는 `adoption_id`도 `entry_decision_id`도 없다.
2. **진짜 조회 오류** — DB 오류. 이것은 삼켜도 다음 주기가 다시 본다.

**고치는 방식**: 둘을 가른다. 「기록 없음」은 오류가 아니라 **사실**이고,
그 사실의 이름은 「보호 상태가 없는 보유」다 — `alertUnmanaged`가 이미 말하는 것이다.
그리로 보낸다. 진짜 조회 오류만 조용히 반환한다.

### `alertUnmanaged`는 왜 안 바꾸는가

`alertUnmanaged`(`adoption.go:392-436`, **분기 6개 · 미진입 0개**)의 사유 분기
(**B3**`:405`·**B4**`:407`·**B5**`:409`·**B6**`:412`)는 전부 진입하고, why-matrix 주석이
그 설계를 적는다 — *"the one thing this switch may never do is tell the owner of a
tried-and-failed designation that 'adoption is off'."* **옳고 시험된 코드다.**

바뀌는 것은 **등급뿐**이다(R1). 본문은 이미 *"손절·익절이 자동으로 걸려 있지 않다"*고
말한다 — 그 말이 outbox에 남지 않았을 뿐이다.

단 **B1** `:393`의 `d.unmanaged[p.ID]`도 같은 메모리 map이다. 무보호는 수량과 무관하게
**상태**이므로 프로세스당 1회가 곧바로 틀린 것은 아니다. **그러나 critical이 되면
outbox와 재시도가 그 반복을 대신 책임진다** — 억제를 그대로 두는 것이 알림 폭주를
막는 쪽이다. tasks 2.5가 그 판단을 시험으로 고정한다.

## D3. R3 — 총위험을 재는 것은 분모를 바꾸는 것이 아니다

### 왜 지금 수가 틀렸는가

`initial_risk`는 **주당** 값이고 t0에 얼어붙는다. 보고가 그 값을 그대로 쓰면
수량이 10배가 되어도 같은 수를 말한다.

원장 실측:

| 종목 | 편입 수량 | 현재 수량 | 유효손절 | **현재 총위험** |
| --- | ---: | ---: | ---: | ---: |
| 475150 | 3 | **32** | 57,900 (본전 승격) | **3,200** |
| 080220 | 2 | **12** | 77,406 | **19,128** |
| 272210 | — | 4 | 69,905 (평단 위) | **−7,220** (고정된 이익) |

> **정정(1라운드).** 1판은 이 표를 `initial_stop`으로 만들어 475150을 **58,784**로 적었다.
> R3의 공식은 「유효 손절가」이고 그것은 `baseline_price`다. **실제의 18.4배를 적었다** —
> 과소보고를 없애겠다는 절이 과대보고했다. 값의 사본이 proposal·tasks·issues에도 있었고
> 같이 고쳤다(`review.md` §1.2 — **정정의 단위는 좌표가 아니라 값이다**).

**유효 손절이 평단 위면 그 수는 음수다.** 272210·TSLA가 그 경우이며, 그것은 「위험」이
아니라 **고정된 이익**이다. 보고는 둘을 구별해야 한다 — 음수를 위험으로 적으면
운영자가 정반대로 읽는다.

### 무엇을 더하는가

`(평균단가 − 유효 손절가) × 현재 보유수량`을 알림 필드와 로그에 싣는다.

**이것은 읽기와 산술뿐이다.** `exit_states`의 네 기준 컬럼에는 **쓰지 않는다.**
그 컬럼을 쓰는 유일한 자리는 `resetExitStateForReadoptTx`이고(D4), a095는 그것을
부르지 않는다.

### 평균단가는 확실한 수가 아니다 — 그렇게 적는다

`ApplyPositionAdjustment`(`position_adjustments.go:202`, **분기 27개**)의 `:312`:

```go
NewAvgPrice: firstNonEmpty(req.NewAvgPrice, target.AvgPrice),
```

**브로커가 원가를 주지 않으면 직전 평단이 그대로 이어진다.** 즉 `positions.avg_price`가
stale일 수 있다. a095는 그 값을 고치지 않고, **불확실하다는 사실을 보고에 함께 담는다.**

원장이 그 가능성을 보인다 — 475150의 `position_adjustments`는 수량을 3→32로 옮기는
동안 `new_avg_price`가 58,000에서 움직이지 않았다.

### StockOS가 같은 것을 한다

`projected_risk_amount = (avg_cost − effective_stop_price) × filled_quantity`를 체결마다
원장 전수에서 재산출하고, 시험이 과소보고를 명시적으로 배제한다
(`test_projected_risk_is_recomputed_when_a_later_fill_moves_the_average`).

## D4. 왜 손절가를 평단 기준으로 다시 계산하지 않는가

### 방향이 대칭이 아니다

| 상황 | 평단 기준 손절 후보 | 현재 손절 대비 | 위험 |
| --- | --- | --- | --- |
| **물타기**(평단↓) | 낮다 | **내려간다** | **커진다** — §6 위반 |
| **불타기**(평단↑) | 높다 | 올라간다 | 작아진다 — §6 부합 |

**사용자가 본 것은 물타기 쪽이다.** 그 방향으로 고치면 손절이 내려간다.
실측이 그것을 보인다 — 066570·080220·272210은 지금 **의도한 3%보다 가까운** 손절을
갖고 있고, 평단 기준으로 다시 계산하면 셋 다 손절이 내려간다.

StockOS의 spec이 정확히 그것을 금지한다
(`position-campaign-core/spec.md:53-56`): *"SHALL **reject** any update that would move
a long campaign's effective stop price **below** its previously stored value."*

### 불타기 방향도 지금은 하지 않는다 — 이유 둘

**(1) 손절가를 다시 세우는 쓰기 자리는 하나뿐이고, 그것은 운영자 행동이다.**

`resetExitStateForReadoptTx`(`apply_hook.go:684-731`, **분기 6개 · 미진입 4개**)의
주석이 선언한다 — *"the only reset writer for the four guarded columns."*
호출자는 `positionpolicy.ActionReadopt`(`position_policy.go:145`) 하나다.

**미진입 4개**(B1·B3·B4·B5, 전부 오류 처리)가 그 경로의 시험 밀도를 말해 준다.
여기에 자동 경로를 붙이는 것은 **거의 시험되지 않은 쓰기 자리에 새 호출자를 더하는 일**이다.

StockOS도 같은 자리에 SHALL을 둔다 — *"손절 갱신 경로는 하나다 … 비후퇴 검사를 수행하는
단일 저장 경로만이 이 컬럼을 갱신한다."*

**(2) `entry_price`를 움직이면 한 값만 움직이는 것이 아니다.**

`EvaluateLadder`(`ladder.go:307-…`, **분기 32개 · 미진입 11개**)에서:

- `:329` `entry, err := positive("entry price", in.EntryPrice)`
- `:358` `returnPct := percentOf(probe, entry)` — **수익률의 분모**
- `:387` `lock, err := lockPrice(entry, in.Policy.Rungs[newIndex].StopPct)`
- `:503-509` `lockPrice` = `entry × (1 + pct/100)`

정책 표의 `StopPct`도 *"the protected stop relative to the **entry price**"*(`:99-100`)다.

**즉 진입가를 옮기면 레벨 판정·rung 잠금가·high-water 대비가 전부 같이 움직인다.**
그 상호작용은 이 change가 만드는 계측 없이 판단할 수 없다.

### 그래서 무엇을 남기는가

**별도 change의 선행 조건 셋**(spec에 SHALL로 적는다):

1. 유효 손절가의 쓰기 경로가 **하나**임이 유지될 것
2. 하향 시도는 **거부하고 기록**할 것
3. 진입 기준 이동이 사다리 전체에 미치는 범위를 **계측으로 확인**할 것

`issues.md`가 그것을 기록한다. **침묵한 생략이 아니다.**

## D5. 셋의 상호작용

```text
계좌 스냅샷: 앱에서 산 만큼 수량이 늘었다
  │
  ├─ 편입 기록 있음 ──→ checkExternalIncrease
  │                       ├─ B1 :442  이미 알렸나?
  │                       │    └─ R2: 「프로세스당 1회」→「지난 보고 수량보다 늘었나」
  │                       ├─ B2 :446  편입 기록 조회 실패
  │                       │    └─ R2: 「기록 없음」과 「조회 오류」를 가른다
  │                       └─ B3 :450  늘었나 → alert
  │                            └─ R3: 총위험 (평단−손절)×수량 을 싣는다
  │                                   + 평단이 stale일 수 있다는 사실
  │
  └─ 편입 기록 없음 ──→ (오늘) B2가 조용히 반환
                          └─ R2 후: alertUnmanaged 로 — 「보호 상태가 없는 보유」
                                     └─ 010170 30주가 여기서 보인다
                                          │
                    ┌─────────────────────┘
                    ↓
          obs.Event{Type: EventExitPositionUnmanaged}
                    ↓
          SeverityOf B1 :310  ← R1: criticalEvents 에 한 줄
                    ↓
          Notify B1 :111   severity != critical ?
             ├─ (오늘) 예 → publishBestEffort → 흔적 없음, 재시도 없음
             └─ (R1 후) 아니오 → notifyCritical → **alert_outbox 행 + 재시도**
```

**R1만으로는 안 된다.** 등급이 올라가도 B1이 프로세스당 1회로 억제하고 B2가
미편입 포지션을 삼킨다 — **010170은 여전히 안 보인다.**

**R2만으로는 안 된다.** 더 자주·더 넓게 울려도 전부 `SeverityNormal`이라
outbox에 한 행도 안 남는다.

**R3만으로는 안 된다.** 총위험을 재도 그것을 실을 알림이 durable하지 않다.

## D6. 무엇을 하지 않는가

| | 결정 | 근거 |
| --- | --- | --- |
| 손절가를 평단 기준으로 재계산 | **안 한다** | 물타기 방향에서 손절이 **내려간다**(§6). 실측 3종목이 그 경우다 |
| 불타기 래칫(상향만) | **지금은 안 한다** | 쓰기 자리 규율과 사다리 파급을 먼저 계측해야 한다(D4). 선행 조건을 spec과 `issues.md`에 남긴다 |
| `entry_price`·`initial_stop`·`initial_risk` 쓰기 | **안 한다** | 이미 보고된 R의 분모. `checkExternalIncrease` 주석의 논거가 유효하다 |
| `positions.avg_price` 보정 | **안 한다** | `firstNonEmpty`(`:312`)가 만드는 stale은 별개 문제다. **불확실성을 표기**하는 것까지가 a095 |
| `SeverityOf`·`Notify`·`publishBestEffort` 본문 수정 | **안 한다** | 셋 다 설계대로 동작한다. 고칠 것은 그들이 읽는 표다 |
| `alertUnmanaged`의 사유 switch 수정 | **안 한다** | 분기 6개 전부 진입하고 why-matrix가 옳다 |
| 수량 상한(사이징) 도입 | **안 한다** | a095는 세고 알린다. 진입을 막는 것은 사이징 변경이고 별도 근거가 필요하다 |
| `alertUnmanaged` B1의 프로세스당 1회 해제 | **안 한다** | 무보호는 수량과 무관한 **상태**이고, critical이 되면 outbox 재시도가 반복을 대신 책임진다 |

## D7. 실패 모드 재검토

| 우려 | 답 |
| --- | --- |
| 등급을 올리면 알림이 폭주하는가? | R2의 억제가 **「지난 보고 수량보다 늘었나」**이므로 수량이 그대로면 한 번도 더 울지 않는다. 무보호 쪽은 억제를 그대로 둔다(D6) |
| critical 알림 실패가 엔진을 멈추는가? | `notifyCritical`은 전달 실패 시 게이트를 latch한다. **그것이 이 등급의 정의다** — a091이 같은 판단을 했다. 다만 **손절·취소는 게이트에 걸리지 않는다**(entry 전용) |
| 발신 자리 넷을 한꺼번에 올리는 것이 과한가? | 넷 다 「가진 것에 손절이 안 걸려 있다」는 같은 사실이다. 등급을 이벤트 종류로 매기는 이상 분리하려면 **새 이벤트 종류**가 필요하고, 그것은 표를 늘리는 대신 의미를 쪼개는 일이다. tasks 1.3이 넷의 문구가 실제로 같은 사실을 말하는지 확인한다 |
| 총위험 계산이 손절을 늦추는가? | 알림 경로에서만 산출한다. 판정·제출 경로에 새 호출을 넣지 않는다(§0.3) |
| 평단이 stale이면 총위험도 틀리지 않는가? | 틀릴 수 있다. **그래서 보고에 그 사실을 함께 담는다** — 확실하지 않은 수를 확실한 것처럼 말하지 않는 것이 요구다(spec) |
| a091과 같은 map을 건드려 충돌하는가? | 같은 map의 **다른 줄**이다. 텍스트 충돌은 병합 순서로 풀리고 논리 의존은 없다 |
| 010170이 배포로 보호되는가? | **아니다.** a095는 보고만 바꾼다. 편입은 운영자 행동이고, 배포 전까지 사람이 처리한다 |
