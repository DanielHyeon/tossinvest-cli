# a095 · 손절은 자기가 무엇을 덮는지 알아야 한다

- **Feature**: `FEAT-TOS-009` — Exit line truth and position policy lifecycle
- **Story**: `STORY-TOS-a095`
- **Spec**: `exit-policy` · `engine-safety`
- **위험 등급**: **High-risk** — 손절가·총위험·알림 등급. §0.3·§0.6 적용.
- **base-commit**: `ec29dc72c0fd589daa2069ccf26bad26baeb2a04`

> **작성 순서**: 이 문서의 분기 주장은 전부 `analysis/function-logic/`의 AST 산출물에서
> 나왔다. 산출물이 문서보다 **먼저** 만들어졌다 (`.claude/CLAUDE.md`「단계 건너뛰기 금지」).
> 함수 9개 · 분기 98개 · 진입 62 · 미진입 27 · 자체 블록 없음 9.

## Why

사용자 보고(2026-08-06): *"익절 손절이 잘못 계산되고 있는것 같은데 추가 구매하면 평균
단가가 낮아 지는데 이를 반영 하지 않고 있는 것 같아."*

**절반은 맞고 절반은 반대다. 그리고 진짜 결함은 셋째 것이다.**

### 실측 — 열린 포지션 전부

원장 `positions` ⋈ `exit_states`, 2026-08-07:

| 종목 | 수량 | 평단 | t0 진입가 | `initial_stop` | **유효손절(`baseline_price`)** | **총위험** |
| --- | ---: | ---: | ---: | ---: | ---: | ---: |
| 066570 | 1 | 181,600 | 182,000 | 176,540 | 176,540 | 5,060 |
| 080220 | 12 | 79,000 | 79,800 | 77,406 | 77,406 | **19,128** |
| 272210 | 4 | 68,100 | 68,200 | 66,154 | **69,905** | **−7,220** (이익 고정) |
| 475150 | 32 | 58,000 | 57,900 | 56,163 | **57,900**(본전 승격) | **3,200** |
| TSLA | 0.000154 | 300.01 | 315 | 305.55 | **326.97** | 음수 (먼지) |
| **010170** | **30** | **11,630** | **—** | **—** | **—** | **exit_states 행 없음** |

> **정정(1라운드).** 이 표의 초판은 총위험을 `initial_stop`으로 계산했다. **틀렸다** —
> R3의 공식은 「유효 손절가」이고 그것은 `baseline_price`다. 475150은 이미 본전으로
> 승격돼 있었고(a094 proposal이 같은 원장에서 *"손절선 56,163 → 57,900(본전)"*을 인용한다),
> 272210은 손절이 진입가 **위**에 있어 이익이 고정된 상태다.
> **초판은 475150을 58,784로 적었다 — 실제의 18.4배다.**
> 과소보고를 없애겠다는 change가 대표 사례를 과대보고했다.

**손절은 전부 `entry_price` 대비 정확히 −3.00%다.** `avg_price`는 어디에도 안 쓰인다.

### 유효 손절 대비로 다시 재면 그림이 달라진다 (1라운드 정정)

`baseline_price`(유효 손절) 대비 평단:

| 종목 | 평단 | 유효손절 | 유효손절/평단 | 승격됐나 |
| --- | ---: | ---: | ---: | --- |
| 066570 | 181,600 | 176,540 | −2.79% | 아니오 (`= initial_stop`) |
| 080220 | 79,000 | 77,406 | −2.02% | 아니오 |
| 475150 | 58,000 | 57,900 | **−0.17%** | **예 — 본전** |
| 272210 | 68,100 | 69,905 | **+2.65%** | **예 — 이익 고정** |
| TSLA | 300.01 | 326.97 | **+8.99%** | **예** |

**초판은 이 표를 `initial_stop`으로 만들어 −3.17%·−2.86% 같은 수를 적었다. 전부 틀렸다.**

**그리고 그 정정이 서사를 바꾼다: 손절선은 얼어붙어 있지 않다.** 래칫이 다섯 중 셋을
이미 올렸다. **얼어붙은 것은 `entry_price`** — 모든 레벨이 그것에서 파생되는 기준점이다
(`ladder.go:503-509` `lockPrice(entry, pct)`).

### 사용자가 본 것 (물타기) — 여전히 안전한 방향이다

평단이 내려간 포지션에서 **평단 기준으로 손절을 다시 계산하면 손절이 내려간다.**
066570(−2.79%)·080220(−2.02%)이 그 경우다 — 지금은 의도한 3%보다 **가깝고**,
평단 기준으로 옮기면 **멀어진다.**

**즉 물타기 방향은 고쳐야 할 결함이 아니다.** 고치면 위험이 커진다 —
§6(보수 방향만) 위반이고, StockOS의 SHALL이 정확히 그것을 거부한다.

### 불타기 방향은 래칫이 이미 일부 처리한다

475150은 평단(58,000)이 t0 진입가(57,900)보다 높지만, **래칫이 손절을 본전까지 올려**
유효 손절이 57,900이다 — 평단 대비 −0.17%. **초판이 걱정한 −3.17%는 존재하지 않는다.**

**따라서 「불타기 시 손절이 안 따라 올라간다」는 초판의 주장도 약해진다.** 래칫은
가격이 오르면 올린다. 올리지 못하는 것은 **가격이 오르지 않은 채 수량만 느는 경우**이고,
그때 커지는 것은 주당 위험이 아니라 **총위험**이다 — 다음 절이 그것이다.

### 그러나 훨씬 큰 것은 총위험이다

`initial_risk`는 **주당** 값이고 t0에 얼어붙는다. 수량은 앱에서 계속 늘었다.

- **475150**: 편입 3주 → 현재 **32주**. 유효 손절 57,900 기준 총위험 **3,200원**.
  **수량이 10.7배가 되는 동안 아무도 그 수를 계산하지 않았다** — 그것이 결함이지,
  수가 큰 것이 결함이 아니다
- **080220**: 2주 → **12주**. 유효 손절이 `initial_stop`과 같아 총위험 **19,128원**
- **부호가 음수일 수 있다** — 272210·TSLA는 유효 손절이 평단 위다. 「위험」이 아니라
  **고정된 이익**이며, 보고가 그것을 구별해야 한다

**수량 증가에 상한이 없고, 그 사실을 세는 곳도 없다.**

### 그리고 010170은 지금 무보호다

`instance_seq=2`, 30주, 평단 11,630. `adoption_id`도 `entry_decision_id`도 **없고**,
`exit_states` **0행**, `exit_events` **0건**. **손절도 익절도 걸려 있지 않다.**

## TossOS는 이미 알고 있었다 — 그리고 그 결정의 안전 논거가 무너져 있다

동결은 **의도적 설계**다. `checkExternalIncrease`(`internal/app/engine/adoption.go:437-441`)의
주석이 그렇게 선언한다:

> *The t0 is **deliberately** not recomputed: `exit_states` freezes the entry, the initial
> risk and the initial quantity, and moving any of them would rewrite the denominator
> every R on that position has already been expressed in. **Reporting it is what the
> operator can act on**; re-sizing is a later change.*

**그 논거는 유효하다.** 이미 보고된 R의 분모를 사후에 바꾸는 것은 원장을 다시 쓰는 일이다.

**그리고 감지 코드도 있다.** 같은 함수가
`Title: "편입 후 수량이 늘었고 고정된 t0가 증가분을 덮지 않는다"`로 알린다.

**그런데 그 보고가 나가지 않는다.**

### 사슬 — AST가 열거한 분기

**① `EventExitPositionUnmanaged`가 등급표에 없다**

`criticalEvents`(`internal/obs/event.go:279-298`)는 **18종**을 담는다.
`EventExitPositionUnmanaged`(`:200`)는 **그 안에 없다.**

**② `SeverityOf`는 그 map만 본다 (분기 1개)**

```go
func SeverityOf(t EventType) Severity {   // event.go:309
    if criticalEvents[t] { return SeverityCritical }   // B1 :310 — 진입 실측 예
    return SeverityNormal
}
```

미등재는 **`SeverityNormal`**이다. 주석이 그 설계를 명시한다 —
*"Genuinely critical conditions are named in the table above."*

**③ `Notify`의 분기 하나가 durable 경로를 가른다 (분기 1개)**

`Notifier.Notify`(`internal/obs/notifier.go:107`) **B1** `:111`
`if severity != SeverityCritical` → `publishBestEffort` → **`return nil`**.
outbox 쓰기는 `notifyCritical`에만 있다.

**④ best-effort는 이름이 계약이다 (분기 2개)**

`publishBestEffort`(`:138`) **B1** `:139` `if n.Publisher == nil { return }` — **로그도 없다.**
**B2** `:142`는 전송 실패를 `n.Log != nil`일 때만 남긴다. 주석:
*"treating its failure as an incident would make the grading meaningless."*

**원장 실측이 사슬을 닫는다.** `alert_outbox` 전수(최신 2026-08-07T00:32:11Z):

| | |
| --- | --- |
| 총 행 | **13** |
| severity | **전부 critical** |
| `exit.position_unmanaged` | **0건** |

**운영자는 475150이 3→32주가 된 것도, 010170의 30주가 무보호인 것도 원장에 남는
형태로는 한 번도 통보받지 못했다.** 동결을 정당화하는 유일한 근거가 작동하지 않는다.

### ⑤ 게다가 알림은 프로세스당 한 번이고, 미편입 포지션에는 아예 없다

`checkExternalIncrease`는 **분기 3개**이고 **2개가 미진입**이다(실측):

| 분기 | 자리 | 조건 | 진입 실측 |
| --- | --- | --- | --- |
| **B1** | `:442` | `if d.grown[p.ID]` — **프로세스 메모리 map** | **아니오** |
| **B2** | `:446` | `if err != nil` — `AdoptionOf` 실패 시 **조용히 반환** | **아니오** |
| **B3** | `:450` | `if err != nil \|\| cmp <= 0` | 예 |

- **B1**: 억제는 `d.grown` 메모리 map이다. 재시작하면 초기화되고, 한 프로세스 안에서는
  3주→32주로 계속 늘어도 **한 번만** 운다.
- **B2**: `AdoptionOf`가 실패하면 반환한다. **편입 기록이 없는 포지션은 이 검사를 아예
  받지 않는다** — 엔진이 직접 연 포지션과 010170 같은 미편입 보유가 그것이다.
- **두 분기 다 어떤 시험도 밟지 않는다.**

## StockOS는 어떻게 하는가 — 확인했다

`/mnt/D/project/axipient/stockos/openspec/specs/position-campaign-core/spec.md:52-56`:

> ### Requirement: 유효 손절가는 Leg 추가로 하향되어서는 안 된다
> The system SHALL **reject** any update that would move a long campaign's effective stop
> price **below** its previously stored value, SHALL **record the rejected attempt**
> instead of silently ignoring it, and SHALL keep the prior value in effect.

시나리오가 방향을 못 박는다 — *"평균단가가 내려가도 손절은 따라 내려가지 않는다"*,
*"상향은 무조건 허용된다"*, 그리고 **"손절 갱신 경로는 하나다"**.

**즉 StockOS의 답은 「재계산 금지」가 아니라 「재계산하되 하향은 거부하고 기록한다」 —
래칫이다.**

**단, 그 모듈은 라이브에 배선돼 있지 않다.** 같은 spec `:123-126`:

> The system SHALL keep the campaign core **free of any call site in the live order,
> entry, or exit paths** … and SHALL prove the absence of such wiring **by test**.

**StockOS의 실제 라이브 경로는 TossOS와 똑같이 동결돼 있다.**
따라서 인용하는 것은 **작동 중인 구현이 아니라 검토를 마친 계약**이다. 그렇게 적는다.

## What Changes

**셋을 바꾼다. 손절가 계산 규칙 자체는 물타기 방향으로 바꾸지 않는다.**

### R1 — 무보호와 수량 증가는 critical이다

`EventExitPositionUnmanaged`를 `criticalEvents`에 등재한다. 그것뿐이다 —
`SeverityOf`·`Notify`·`publishBestEffort`의 본문은 건드리지 않는다.
등급이 바뀌면 `Notify` **B1** `:111`이 자동으로 `notifyCritical`로 간다.

**효과**: outbox 행이 생기고, 전달 실패에 재시도가 붙고, 원장에 흔적이 남는다.

**a091과 같은 계열이되 같은 것이 아니다.** a091은 `EventExitProposalCapped`를 다루고
a095는 `EventExitPositionUnmanaged`를 다룬다. 발신 자리는 **4곳**
(`adoption.go:418`·`:456`, `exitloop.go:1501`, `exitwiring.go:104`)이다.

**a091이 적은 유보는 여기서는 성립하지 않는다.** a091은 *"8/2에 운영자가 호출받지 못한
이유는 등급이 아니라 transport 부재"*라 적었다. **지금은 transport가 있다** —
`alert_outbox` 13행이 그 증거이고 그중 4행은 실제 전달을 시도했다. 따라서 a095의
이익은 흔적 하나가 아니라 **흔적 + 호출**이다.

### R2 — 억제는 프로세스가 아니라 사실을 따른다

- **B1**: `d.grown` 메모리 map을 **수량 기준**으로 바꾼다. 마지막으로 알린 수량보다
  또 늘면 다시 알린다. 3→32주가 한 번이 아니라 늘어난 만큼 보인다.
- **B2**: 편입 기록이 없는 포지션에서 **조용히 반환하지 않는다.** 그 포지션은
  `alertUnmanaged` 쪽 사실(무보호)에 해당하므로 그리로 보낸다 — 010170이 그 경우다.

### R3 — 총위험을 재고 상한에 걸린 것을 말한다

`(평단 − 유효손절) × 현재수량`을 재계산해 알림 본문과 원장에 싣는다.
**t0의 `initial_risk`는 그대로 둔다** — R의 분모를 바꾸지 않는다.
새로 만드는 것은 **분모가 아니라 계기판**이다.

StockOS가 같은 것을 한다 —
`projected_risk_amount = (avg_cost − effective_stop_price) × filled_quantity`를
체결마다 원장에서 재산출하고, 과소보고를 시험으로 배제한다.

### 하지 않는 것 — 손절가를 평단 기준으로 다시 계산하기

**물타기(평단↓)에서 평단 기준 손절은 지금 손절보다 낮다.** 그것을 채택하면 손절이
내려간다 — §6 위반이고, StockOS의 SHALL이 정확히 그것을 거부한다.

**불타기(평단↑)에서는 올라간다.** 그 방향만 취하는 래칫은 §6에 부합한다.
**그러나 a095는 그것도 지금 하지 않는다.** 이유는 두 가지이며 둘 다 측정에서 나왔다:

1. **손절가를 다시 세우는 쓰기 자리는 하나뿐**이고
   (`resetExitStateForReadoptTx`, `apply_hook.go:679-684` — 주석이
   *"the only reset writer for the four guarded columns"*라 선언한다) 그것은
   **운영자 행동(`positionpolicy.ActionReadopt`)에서만** 불린다. 자동 경로를
   붙이는 것은 그 규율을 깨는 일이고, StockOS도 같은 자리에 SHALL을 둔다
   (*"손절 갱신 경로는 하나다"*).
2. **모든 선이 `entry_price`에서 나온다.** `EvaluateLadder`(`ladder.go:307`, 분기 32)의
   `lockPrice(entry, pct)`(`:503-509`)가 `entry × (1 + pct/100)`이고 `percentOf`(`:514`)가
   수익률을 `entry`로 잰다. `entry_price`를 움직이면 **레벨·rung·high-water가 전부
   같이 움직인다** — 그 상호작용은 a095의 계측 없이 판단할 수 없다.

**따라서 래칫은 a095가 만드는 계측 위에서, 별도 change로 한다.** 선행 조건을
`issues.md`에 적는다. **침묵한 생략이 아니다.**

## Why these three, and why not fewer

- **R1만**: 등급은 올라가지만 알림이 프로세스당 1회이고 미편입 포지션은 여전히 조용하다.
  **010170은 그대로 안 보인다.**
- **R2만**: 더 자주·더 넓게 울리지만 여전히 `SeverityNormal`이라 **outbox에 안 남는다.**
- **R3만**: 총위험을 재도 그것을 실을 알림이 durable하지 않다.

**R1이 알림을 원장에 남기고, R2가 그 알림이 실제로 울리게 하고, R3이 무엇이 위험한지를
수로 말한다.**

## 지금 열린 것들은 어떻게 되나

**a095는 손절가를 소급해서 바꾸지 않는다.** 배포 후에 바뀌는 것은 **보고**뿐이다.

| | 배포 후 |
| --- | --- |
| 010170 (30주, 무보호) | `alertUnmanaged`가 **critical**로 울리고 outbox에 남는다. **보호가 생기지는 않는다** — 편입은 운영자 행동이다 |
| 475150 (32주, 총위험 3,200) | 수량 증가 알림이 **critical**로, 총위험 수와 함께 |
| 080220 (12주, 19,128) | 같음 |
| 066570·272210 | 물타기 방향이므로 손절이 의도보다 **가깝다.** 알림 대상이나 위험 축소 방향이다 |
| TSLA (먼지) | 손절이 평단 위 — 손실 위험 없음 |

**배포 전까지 010170은 사람이 처리한다.**

## Out of scope — 침묵하지 않고 적는다

- **손절가의 평단 기준 재계산(래칫)** — 위 「하지 않는 것」. 별도 change의 선행 조건을
  `issues.md`에 적는다
- **`positions.avg_price`의 정확성** — `ApplyPositionAdjustment`(`:312`)는
  `firstNonEmpty(req.NewAvgPrice, target.AvgPrice)`로 **브로커가 원가를 안 주면 옛 평단을
  그대로 이어붙인다.** 즉 `avg_price` 자체가 stale일 수 있다. a095는 그 값을 **읽기만**
  하고 고치지 않으며, **그 불확실성을 알림 본문에 명시**한다
- **수량 상한(포지션 사이징)** — a095는 세고 알리기만 한다. 진입을 막는 것은 사이징
  변경이고 §6상 별도 근거가 필요하다
- **a094와 겹치지 않는다** — a094는 409 오분류와 충돌 해소다. 같은 종목이 등장하지만
  원인이 다르다
- **a091과 겹치지 않는다** — 같은 실패 계열이되 **다른 이벤트 종류**다. 두 change가
  같은 map을 건드리므로 병합 순서만 주의한다(`tasks.md` 선후 관계)

## Impact

| | 자리 | 성격 |
| --- | --- | --- |
| R1 | `internal/obs/event.go` `criticalEvents` | **map에 한 줄.** 함수 본문 무변화 |
| R2 | `internal/app/engine/adoption.go` `checkExternalIncrease`·`alertUnmanaged` | 억제 키를 수량 기준으로 · B2의 조용한 반환 제거 |
| R3 | 같은 파일 + 알림 필드 | **읽기와 산술뿐.** 원장의 t0는 안 바꾼다 |

spec: `engine-safety`(등급), `exit-policy`(무보호·수량 증가의 보고 의무와 총위험)

**기존 함수 내부를 고치므로 Function Logic Map 면제는 없다.** 산출물 9개는 이미 있고,
구현 후 재생성한다.
