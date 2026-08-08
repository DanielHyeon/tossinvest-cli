# a089 proposal-freeze 리뷰

- **날짜**: 2026-08-06
- **대상**: proposal / design / tasks / specs (exit-policy, engine-safety), base `ec29dc72`
- **위험 등급**: High-risk → 적대적 Eng 관점 필수
- **보이스**: Claude Eng (독립·적대적) 실행 · Codex `[codex-unavailable]` 한도 소진(2026-08-08 회복)
  · CEO 보이스 미실행 — Eng가 진단 자체의 blocking 오류를 찾아 거기서 종료
- **판정**: **FREEZE 거부**

## 재검증 — 내 진단 4건 중 2건이 거짓, 1건은 수리 방법이 자기 파괴

WORKFLOW의 "리뷰어 주장은 Manager가 코드로 재검증"에 따라 전부 직접 확인했다.

| # | 내 주장 | 확인 | 결과 |
| --- | --- | --- | --- |
| 1 | `noteDelay`가 거부 경로에 미연결 | 호출자 전수 | **참** — 그러나 수리가 자기 파괴(C1) |
| 2 | 재제안에 보장된 주기가 없다 | `snapshot.go:241-250` | **거짓** |
| 3 | 운영자는 0회 들었다 | `notifier.go:252-285` + 로그 | **거짓** |
| 4 | 400이 한 덩어리다 | `submit` default | 참 — 그러나 세션 분류는 근거 없음(H5) |
| — | **"9분 22초 무보호"** | `exit_events` 전수 | **거짓 — 실제는 약 106초** |

### C3 (가장 중요) — 7분 43초는 무보호 구간이 아니었다

`exit_events`의 005930 15행 전수. 판정이 제안을 만든 시각은 여섯 번뿐이다.

```text
00:54:00.199  STOP_LOSS_LADDER  observed 245750   baseline 245895
00:54:11.194  STOP_LOSS_LADDER  observed 245750
00:55:02.831  STOP_LOSS_LADDER  observed 245750
      ────── 7분 43초: 이 포지션의 행이 0건 ──────
01:02:44.932  STOP_LOSS_LADDER  observed 245750
01:02:51.885  STOP_LOSS_LADDER  observed 245750
01:03:28.826  STOP_LOSS_LADDER  observed 245500 → PROPOSAL_FILLED → COMPLETED
```

공백 구간에 **엔진은 살아 있었고 청산 루프도 돌고 있었다** — reconcile 9사이클,
그리고 **다른 포지션(`pos-7616a68d…`)이 00:59:28에 exit_event를 기록**했다.

`ChangedFromState`가 두 분기 모두 `|| s.Orderable`로 끝나므로(`snapshot.go:244,248`)
**주문 가능한 제안은 반드시 행을 남긴다.** 행이 0건이라는 것은 그 구간에 주문 가능한
제안이 없었다는 뜻이고, baseline 245,895 대비 observed가 245,750이던 종목의 제안이
사라지는 유일한 길은 **가격이 기준선 위로 회복한 것**이다.

즉 실제 무보호는 **약 62초 + 약 44초 ≈ 106초**이고, 가운데 7분 43초는 이탈 자체가 없었다.
나는 **로그 침묵에서 사건을 추론했고**, 이 루프는 정확히 그러지 말라고 설계돼 있다.

> `exitloop.go:380` — "Successful cycles say nothing. One line every five seconds
> forever is not observability."

### C2 — 재제안 주기는 이미 보장돼 있다

```go
// exitpolicy/snapshot.go:241-250
s.Changed = s.HighWater != previousHighWater || s.CurrentProtection != previousProtection ||
    s.RatchetLevel != previousLevel || s.Orderable      // ← 두 분기 모두
```

`!snapshot.Changed` 조기 반환은 주문 가능한 제안에 대해 **도달 불가**다. 내가 근거로 인용한
주석(`exitloop.go:1128`)은 **재판정 중 보류된 익절**의 맥락이고 거부 경로를 설명하지 않는다.
그리고 내 실측 데이터 안에 반증이 있었다 — 01:02:46과 01:02:52 사이 **6초는 정확히 한 사이클**이다.

### C4 — 알림은 전달됐다

`no such alert: 12`는 `requireOneRow`(`outbox.go:282-290`)가 `WHERE id=? AND state='PENDING'`에
0행을 만났을 때만 나온다. 그리고 publisher가 nil이면 `deliver`는 **`Publish` 전에 `break`**
하므로(`notifier.go:252-255`) `Mark*`가 실행되지 않아 이 문자열이 나올 수 없다.

결정적 판별: 전달 소진 시 남는 `"could not be delivered after N attempts"`가 **로그 전체에
0건**이다. 그 소진 경로는 `event`·`alert_id` 필드를 함께 남기는데(`notifier.go:279-282`),
00:11:44(alert 11)과 01:59:41(alert 10)에는 그 필드가 있고 **00:54~01:02의 네 줄에는 없다**.
필드 없는 형태는 `notifier.go:259`·`265`의 `Mark*` 실패뿐이고, 265였다면 소진 줄이 뒤따랐을
것이다. 따라서 **259 — `Publish`가 nil을 반환했다. 즉 성공했다.**

**운영자는 거부마다 알림을 받았다.** 실패한 것은 outbox 장부다.

**여기서 진짜 결함이 나온다(a089 범위 밖):** `EnqueueAlert`가 이벤트 키로 중복 제거하며
이미 `DELIVERED`인 행의 id를 돌려주고, `MarkAlertDelivered`의 `state='PENDING'` 술어가
그것을 거절한다. 결과적으로 **반복되는 critical 조건에 대해 `attempts`·`last_error`·
`UndeliveredCount`가 갱신을 멈춘다.** 게이트의 해제 술어 `remaining == 0`
(`notifier.go:374`)이 자명하게 참이 된다.

### C1 — D1은 발화할 수 없다

`design.md`가 "해제 지점은 `submit`의 `StateConfirmed` 분기 하나다"라고 썼다. 거짓이다.

```go
// exitloop.go:1117
if orderable && (snapshot.CancelPendingFirst || isFullExit(proposal)) {
    ...
    } else {
        o.clearDelay(m.position.ID)      // exitloop.go:1150
    }
}
```

`isFullExit`는 `ActionBaselineBreach`·`ActionLadderStop`을 포함하므로 **모든 보호 제안이
이 분기에 들어온다**. 심볼에 살아 있는 주문이 없으면(거부된 주문은 살아 있지 않다)
`clearTheSymbol`이 `true`를 돌려주고 1150이 실행된다 — **submit보다 먼저**.

그래서 D1을 적용하면 매 사이클이 `clearDelay` → submit → 거부 → `noteDelay`가 되고,
`noteDelay`는 매번 `running == false`를 만나 시계를 재시작만 한다(`exitloop.go:1571-1575`).
경과는 관측 주기를 넘지 못하고 `exit.liquidation_delayed`는 **영원히 안 뜬다.**

### C5 — D4의 요구사항은 이미 존재하고, 둘째 Scenario는 안전을 되돌린다

`EventExitProposalRefused`·`EventExitLiquidationDelayed`는 이미 `criticalEvents`
(`obs/event.go:297-298`)이고 소진 시 `Gate.Block(ReasonAlertUndelivered)`가 이미 걸린다
(`notifier.go:283-285`). ADDED 요구사항이 더하는 것이 없다.

**해로운 쪽**: 내 Scenario는 "전달되거나 확인하면 해제"인데 코드가 전달 절반을 의도적으로 거절한다.

> `notifier.go:305` — "A run that empties the backlog does *not* clear the gate."
> `notifier.go:336-339` — "'the network came back' is not that human."

`Gate.Clear`는 `Acknowledge`에서만 도달 가능하고 운영자 신원을 요구한다. 내 SHALL은
문서화된 human-in-the-loop을 제거한다 — **더 안전하게 만들겠다는 change가 §0.7을 되돌린다.**

## 그 외 확인된 지적

| # | 내용 |
| --- | --- |
| H1 | 재시도가 **재가격 없이** 같은 가격을 다시 보낸다. 상한을 두면 8/5 사건은 9분이 아니라 **영구** 무보호가 됐을 것이다. 그리고 손절 재시도 상한에서 "보수적"은 **더 많은 시도**인데 문서가 방향을 말하지 않았다 |
| H2 | D2가 `ErrProposalPending`·`ArmOutcome`·`SuppressedPending`·`CancelPendingFirst` 네 가드와 충돌하는데 문서가 하나도 언급하지 않았다. tasks가 편집 지점을 "1.x에서 확정"으로 미룬 것은 freeze 시점에 할 결정이다 |
| H3 | retry-matrix §0.1 "주문 mutation은 어떤 오류에도 자동 재시도하지 않는다 — 멱등성 키가 없다"와의 관계를 명시하지 않았다. FAILED_CONFIRMED 논거가 필요하고, 오분류 시 중복 매도 |
| H4 | §0.4 숫자가 없다. K개 포지션 동시 이탈 시 0.2·K req/s가 **mutation 엔드포인트**에 붙고, 429는 AMBIGUOUS→IN_DOUBT라 재시도 폭주가 손절 경로에 IN_DOUBT를 양산한다 |
| H5 | 세션 분류의 인용이 틀렸다(M5가 아니라 M1이고, M1은 **일요일 휴장** 관측이다). M38이 "KR 정규장 밖 일반 주문 접수는 [미측정]"이라고 명시한다. 근거 없는 추측이 **손절을 억제**한다 |
| M1 | 제안이 브로커에 닿지 못하는 경로는 6개인데 D1은 1개만 계측한다 |
| M2 | 재시도 카운터·지연 시계의 재시작 내구성이 freeze 시점에 미결정 |
| M3 | `AllReasonCodes`는 "rename은 데이터 마이그레이션"이라 명시. `official.APIError`는 브로커 `code`/`data`를 파싱하지 않아 분류에 `internal/official` 변경이 필요한데 Impact에 없다 |
| M5 | "9분 22초"는 타임라인상 9분 28초. `notifier.go:373`→374. a088 참조는 존재하지 않는 change |

## 남는 진짜 결함

1. **`noteDelay`가 거부 경로에 미연결** — 참. 다만 잴 대상은 9분이 아니라 62초·44초이고,
   62초는 30초 한계를 넘으므로 첫 창에서는 실제로 울렸을 것이다. 가치는 있고 규모는 작다.
   수리는 C1 때문에 `clearDelay`(1150)와 함께 설계해야 한다.
2. **outbox 장부 결함** — 중복 제거가 DELIVERED 행을 돌려줘 `Mark*`가 실패하고
   `attempts`·`last_error`·`UndeliveredCount`가 갱신을 멈춘다. **a089 범위에 없다.**
3. **거부 사유 미분류** — 참. 단 세션 분류는 측정이 없으므로 제외.

## 판정과 다음

**FREEZE 거부.** 재작성 방향:

- 제목의 전제("9분 무보호")를 **106초**로 교정하고, 그에 맞춰 근거를 다시 쓴다
- C2·C4에 근거한 요구사항 2건을 **철회**한다
- D1을 `clearDelay`(1150)와 함께 재설계한다
- 재시도 계열(H1·H2·H3·H4)은 **재가격이 없는 한 가치가 없고 상한은 해롭다** — a087이
  보낼 새 값을 갖기 전까지 이 축을 빼거나, 상한을 "발주 중단"이 아니라 "에스컬레이션"으로
  바꾼다
- 세션 분류(H5)를 삭제한다
- **범위 후보**: 계측(M1의 6경로 전부) + outbox 장부 결함. 재시도 주기는 이미 있으므로
  건드릴 것이 없다

## 기록해 둘 것 — 세 번 연속 같은 형태로 틀렸다

| 라운드 | 틀린 방식 |
| --- | --- |
| a087 초안 | 실측 1건(n=1)에서 브로커 성질을 단정 |
| a087 교체본 | 선례(StockOS)를 조건 없이 일반화, 호출 사슬 미추적(`internal/trading`) |
| a089 | **로그 침묵에서 사건을 추론**, 호출 사슬 미추적(`clearDelay`·`Changed`·`notifier`) |

공통점은 **가설을 세운 뒤 반증 가능한 곳을 먼저 확인하지 않은 것**이다. 세 번 다 반증은
같은 저장소 안에 있었고 한 번의 조회로 나왔다. 다음 change는 proposal을 쓰기 전에
"이 주장이 거짓이라면 어디에 흔적이 남는가"를 먼저 조회한다.

---

## 재작성 반영 (2026-08-06)

change id를 `a089-a-refused-stop-is-loud-and-retried` → **`a089-an-unserved-stop-is-counted`**로
바꿨다. 재시도가 명시적 non-goal이 됐으므로 옛 이름은 자기 범위를 잘못 말한다.

**위 C3의 "약 106초"도 함께 철회한다.** 그 값은 세 행(00:54:00·00:54:11·00:55:02)이
연속 이탈이라고 가정해 얻은 것인데, 5초 관측 주기에서 62초 구간에 행이 3건뿐이라는 사실이
그 가정을 반증한다 — 그 안에서도 이탈은 간헐적이었다. 관측과 관측 사이의 가격은 원장에
없으므로 **어떤 시간 값도 실측이 아니다.** 실측인 것은 **6판정·5미제출**뿐이고, 재작성본은
그것만 척도로 쓴다.

### 지적별 처리

| # | 지적 | 처리 |
| --- | --- | --- |
| C3 | "9분 22초 무보호"가 거짓 | **전제 교체.** 척도를 시간에서 **횟수**(6판정·5미제출)로 바꿨다. 관측 사이 가격은 `[미측정]`으로 명시 |
| C2 | 재발의 주기는 이미 보장 | **요구사항 철회.** Non-goals에 "고칠 것이 없다"로 기록 |
| C4 | 운영자는 전달받았다 | **요구사항 철회.** 대신 진짜 결함(outbox 재발 장부)을 engine-safety 요구사항으로 |
| C1 | D1이 `clearDelay`(1150)에 지워진다 | **재설계.** 시계의 정의를 "보호가 필요한데 살아 있는 보호 주문이 없다"로 바꾸고 해제를 접수·종료로 한정. tasks 1.3이 이 결함의 RED |
| C5 | D4는 이미 존재하고 둘째 Scenario가 §0.7을 되돌린다 | **삭제.** 게이트 승격 요구사항 자체를 뺐다. 새 요구사항은 게이트를 **더 어렵게** 푼다 |
| H1 | 재시도는 재가격 없이 무가치, 상한은 해롭다 | **축 전체 삭제.** Non-goals에 "상한이 있었다면 영구 무보호였다"로 근거 기록 |
| H2 | 네 가드와의 충돌 미언급 | 재발의를 안 건드리므로 **소멸** |
| H3 | retry-matrix §0.1과의 관계 | 재시도 미추가이므로 **소멸**. tasks 6.3이 요청 수 무변화를 확인 |
| H4 | §0.4 숫자 없음 | 요청 수 무변화이므로 **소멸** |
| H5 | 세션 분류의 인용이 틀렸고 손절을 억제한다 | **삭제.** 사유는 기록만 하고 동작을 분기하지 않는다(spec의 SHALL NOT) |
| M1 | 무제출 경로는 6개인데 1개만 계측 | **9개로 확장.** design의 P1~P9 표. 분기마다 손으로 넣지 않고 불변식 한 지점에서 판정 |
| M2 | 재시작 내구성 미결정 | **결정.** 프로세스 메모리, 재시작 시 0. design D2와 tasks 2.5에 명시 |
| M3 | `official.APIError`가 파싱 안 함 → Impact 누락 | **Impact에 추가.** 가산 메서드 하나, 기존 동작 무편집 |
| M5 | 시각·행번호·존재하지 않는 참조 | 시각은 원장 원문으로 교체, a088 참조는 문맥 설명으로 |

### 이 판이 새로 지는 위험

리뷰어가 공격할 자리를 미리 적어 둔다.

1. **`ACKNOWLEDGED` 행의 재개방**(design D3) — 운영자의 확인을 무르는 동작이다. 근거는
   outbox.go 자신의 원칙이지만, 이것이 §0.7을 되돌리는지 아닌지가 이 change의 최대 쟁점이다
2. **지연 알림의 의미 확대**(design D1) — 가격이 회복한 뒤에도 한 번 뜬다. 더 시끄러운
   방향이므로 채택했으나 오경보로 읽힐 수 있다
3. **계수가 재시작을 넘지 못한다** — 원장이 아니라 메모리다. 스키마 미변경을 택한 대가

---

## a089 proposal-freeze 리뷰 · 2라운드

- **날짜**: 2026-08-06
- **대상**: 재작성본 (`a089-an-unserved-stop-is-counted`), base `ec29dc72`
- **보이스**: Claude Eng(적대적) + Claude 안전-불변식 렌즈, 둘 다 독립 실행 ·
  Codex `[codex-unavailable]` 한도 소진(2026-08-08 회복)
- **판정**: **FREEZE 거부**
- 두 보이스의 지적은 Manager가 전부 코드로 재검증했다. 아래는 검증된 것만이다.

### C1 — 근거 4가 거짓이다. 그리고 1라운드의 교정도 함께 거짓이다

재작성본 proposal의 근거 4: "행이 없다는 것은 주문 가능한 제안이 없었다는 뜻이다."

`Changed`의 `|| s.Orderable`은 **`record`에 도달했을 때만** 행을 보장한다. 그 앞에
**흔적을 남기지 않는** 이탈이 있다.

```go
// exitloop.go:453-460
quote, ok := quotes[strings.ToUpper(strings.TrimSpace(state.position.Symbol))]
if !ok {
    // A symbol the price read did not answer for is a symbol this cycle
    // did not observe. Hold it; ...
    continue          // ← 로그 0, 알림 0, 행 0
}
```

`q.Last <= 0`인 시세도 같은 자리에서 조용히 버려진다(`exitloop.go:750-755`).
그리고 `cycle.Observed`·`cycle.Judged`는 **어디에도 기록되지 않는다** —
`reportCycle`은 `cycle.Err != nil`일 때만 쓴다(`exitloop.go:381-384`).
계정 단위 outage는 **전 종목이 실패해야** 뜬다(`exitloop.go:760-762`).

즉 한 종목만 시세가 안 오면 다른 포지션은 계속 행을 남기고 그 종목만 사라진다 —
내가 "다른 포지션이 00:59:28에 행을 남겼다"를 엔진 생존의 증거로 쓴 그 서명이다.

#### 8/2 사건이 양성 대조군을 준다

같은 원장에 두 번째 사건이 있다(내가 재작성 시점에 몰랐던 것).

```text
2026-08-02 23:23:25 ~ 23:26:21, pos-522745e0 (042660)
  STOP_LOSS_LADDER × 13 → 전부 PROPOSAL_REFUSED
  그 사이 obs=83300 base=83300 인 비-주문 행 7건
```

**관측이 살아 있으면 가격 회복은 행을 남긴다.** 8/2는 남겼고, 8/5의 7분 42초에는
전 계정을 통틀어 관측 유래 행이 **1건**뿐이다(약 92주기).

결론: 8/5 공백의 가장 그럴듯한 설명은 **가격 회복이 아니라 관측 누락**이다. 그러면
1라운드에서 내가 쓴 "실제는 약 106초"도 거짓이고, 초판의 "9분 무보호"가 오히려 사실에
가깝다. **정직한 진술은 하나뿐이다 — 8/5의 무보호 시간은 `[미측정]`이며 현재 원장으로는
구할 수 없다.**

#### 이것이 네 번째 반복이다

1라운드 말미에 내가 이렇게 적었다.

> 다음 change는 proposal을 쓰기 전에 "이 주장이 거짓이라면 어디에 흔적이 남는가"를
> 먼저 조회한다.

그리고 다섯 가지를 조회했다 — 전부 **흔적이 남는** 경로였다. **흔적이 안 남는 단 하나의
경로를 조회하지 않았다.** 규율을 적어 놓고 그 규율이 겨냥한 자리를 비켜 갔다.

### C2 — 더 큰 결함이 이 change 위쪽에 있다 (P0)

C1의 코드가 곧 결함이다. **보유 포지션 하나가 시세 응답을 못 받으면 무기한 조용히
청산 루프에서 빠진다.** 판정이 없으므로 제안도 없고, 제안이 없으므로 a089가 셀 대상도 없다.
계정 단위 outage 시계는 전 종목 실패에만 반응한다.

`exitloop.go:376-377`이 스스로 이렇게 적어 뒀다.

> The conditions that mean a position is actually unprotected raise their own
> critical events from where they happen.

**453-460은 아무것도 올리지 않는다.** 설계 의도와 코드가 어긋난 자리다.

a089의 P1~P9는 전부 "제안이 주문이 되지 못한" 경로다. **P0(판정이 아예 일어나지 않은
포지션)은 그보다 위에 있고 더 위험하다.**

### C3 — "P1~P9는 보호 한정"이 거짓

design 61행의 주장이다. `isFullExit`가 `ActionLadderTakeProfit`을 포함하므로
(`exitloop.go:1208-1211`) 평범한 주기의 익절이 1117 분기에 들어오고, 익절을 빼는 1118의
가드는 `m.reJudge`일 때만 작동한다. 따라서 **익절도 1146의 `noteDelay`를 부른다.**
`submit`의 P4~P9에도 보호 여부 검사가 없다.

필터를 넣지 않으면 익절이 시작한 시계가 새 해제 조건(보호 접수·포지션 종료) 어디에도
안 걸려 **latch되고**, 30초 뒤 보호선 안쪽 포지션에 "보호되지 않는 노출" critical이 뜬다.

### C4 — spec exit-policy Requirement 2가 자기모순

- "…없는 **상태가 지속된 시간**을 재야 한다(SHALL)"
- "해제는 **접수**되었거나 **종료**된 경우로 한정해야 한다(SHALL)"

가격이 회복하면 상태를 벗어나지만 해제되지 않는다. 두 SHALL을 동시에 만족하는 구현은 없다.
그리고 이 모순이 tasks 5.1의 기대값("00:55:02에 경과 62초")을 무효로 만든다 — 해석에 따라
알림이 뜨지도 않는다.

### C5 — P2·P3·InDoubt는 "보호 주문이 없다"를 뜻하지 않는다

spec Requirement 1의 "어느 경로로 멈췄든 포지션에 보호 주문이 없다는 사실은 같다"가 거짓이다.

- `pending_action`은 **접수 성공 후에도 남는다** — `submit`의 `StateConfirmed` 분기는
  `release`를 부르지 않는다(`exitloop.go:1288-1295`). 그래서 `ErrProposalPending`
  (`apply_hook.go:666-667`)은 **호가에 살아 있는 확정 보호 주문**을 가진 포지션에서 나온다
- `StateInDoubt`는 주석이 직접 "**The order may exist.** The proposal stays armed"라고 못박는다
  (`exitloop.go:1296-1299`)
- P3의 `arming refuses a second proposal while the first is outstanding`도 같다
  (`exitloop.go:1109-1111`)

표대로 구현하면 **정상 보호 중인 포지션이 매 주기 미제출로 세어지고 30초 뒤 거짓 critical이
손절 채널에 뜬다.** 이 change의 존재 이유가 그 채널의 신뢰성인데 그것을 스스로 깬다.

### C6 — `EnqueueAlert`에는 `deliver`가 없는 두 번째 호출자가 있다

`execgw/replay.go:551`의 `parkAlert`는 `Notifier`를 거치지 않고 enqueue만 한다.
재개방을 켜면 같은 attempt의 park이 반복될 때마다 행이 `PENDING`으로 돌아오고
**닫는 주체가 없다** → `UndeliveredCount`가 영구 >0 → 운영자가 게이트를 풀 수 없다.

design D3의 "`notifyCritical`이 항상 `deliver`를 부른다"는 논증이 이 호출자에 닿지 않는다.
Impact에 `internal/execgw`가 없다.

### C7 — 기존 테스트가 반대 동작을 의도로 못박아 뒀다

```go
// internal/journal/outbox_test.go:56-61
// The first observation's text is kept: it is the one that describes what was
// seen when the condition arose.
if pending[0].Body != "first" { ... }
```

D3은 이 단언을 뒤집어야 한다. design의 "회귀 0"과 tasks 3.7("첫 발생 무변화")은 이것을
덮지 않는다 — 이것은 **재발 시 본문 선택**에 대한 단언이다. 그리고 그 주석이 제시하는
반대 근거를 design은 한 번도 반박하지 않는다.

### 그 외 확인된 지적 (2라운드)

| # | 내용 |
| --- | --- |
| H1 | `ACKNOWLEDGED` 재개방 시 `acknowledged_at`·`acknowledged_by` 처리가 spec·design 어디에도 없다. 지우면 §0.7 감사 기록 파기, 남기면 `state='PENDING'`인데 `acknowledged_by`가 찬 모순 행 |
| H2 | `Flush`는 `deliver`의 뮤텍스를 안 잡는다. 재개방이 `PENDING` 창을 만들면 `Flush`가 같은 알림을 두 번 발송한다 — "발송량 무변화"는 `notifyCritical` 경로 한정으로만 참 |
| H3 | P3 뒤에는 **arm은 커밋됐고 주문은 없고 release도 없는** 상태가 남는다(`exit_state.go:613-615`). 다음 주기는 `SuppressedPending` → `Changed=false` → 조기 반환이라 "한 지점"에 도달하지 못한다. 가장 세야 할 상태에서 계수가 멈춘다 |
| H4 | 판정 자체가 거부된 포지션(`judgeLadder:918-931`, `961-965`)은 `record`에 도달하지 않아 새 불변식 밖이다 |
| M1 | P1~P9는 전수가 아니다. 최소 4건 누락(`1141-1144`, `1177-1188`, `1239-1241`, `release` 실패) |
| M2 | tasks 1.9의 "한 지점"에 **위치가 없다**. `judge`/`record`/`submit`은 `error`만 반환하고 `ExitCycle`에 포지션별 결과가 없다. 라운드1 H2와 같은 형태의 미룸 |
| M3 | "질의 가능한 별도 필드" SHALL이 충족 불가. durable 자리는 outbox `payload` 하나인데 D3이 재발마다 덮어쓴다. `exit_events`는 free-text 열이 없다(`apply_hook.go:807-809`) |
| M4 | "포지션 종료가 해제한다"의 구현 위치가 설계에 없다. 기존 맵 어느 것도 working set 이탈 시 정리되지 않는다 |
| M5 | `delay_seconds`는 `int()`라 62, 본문은 `Round(time.Second)`라 `1m3s`. tasks 5.1의 기대값이 둘로 갈린다 |
| M6 | proposal 51행의 latch 해제 인용이 `899`(ratchet)인데 005930은 LADDER이므로 `966`이다. 결론은 유효 |
| M7 | 근거 1의 `reconcile.clean`은 **다른 루프**의 증거다. 5초 exit 루프가 그 구간에 관측했는지에 대해 아무 말도 하지 않는다 — C1의 원인 |

### 검증에서 살아남은 것

| 주장 | 증거 |
| --- | --- |
| `clearDelay`(1150)가 submit보다 먼저 실행되고 모든 보호 제안이 그 분기에 들어온다 | `exitloop.go:1117·1150·1208-1211` — 1라운드 C1 진단은 참이고 재설계 동기는 정당하다 |
| `noteDelay` 호출자는 `1146`·`1302` 둘뿐 | 전수 grep |
| outbox 재발 장부 결함 | `outbox.go:127-131 / 158-160 / 174-175 / 195-196` + 실물 원장: 행 12가 `attempts=1`인데 거부는 5회 |
| 8/5에 운영자는 다섯 번 다 호출됐다 | `DefaultCriticalAttempts=3`이므로 Publish 실패 시 거부당 4줄이 남아야 하는데 1줄뿐이고 `alert_id`가 없다. publisher nil 마지막 흔적은 00:11:44로 43분 전 |
| §0.8 노출 증가 0 | 4xx 본문은 이미 `detail`로 로그·outbox payload에 흐른다. `code`·`field`는 진부분집합 |
| 스키마 미변경 판단 | `SchemaVersion=30`, migration append-only. 선례: `delayedSince`·`delayAlerted`·`refused`가 이미 in-memory |
| §0.3 위반 없음 | 편집 지점이 전부 기록 경로. `orderable`을 끄는 1138·1147 무편집, `Submit.Place` 앞에 새 분기 없음 |
| §0.9 위반 없음 | 임계·가격·수량·시점 무변경 |

### 실측으로 확인된 배포 사실

| 사실 | 값 |
| --- | --- |
| outbox 실물 | 12행 — `PENDING` 9(전부 `attempts=0`), `DELIVERED` 3, **`ACKNOWLEDGED` 0** |
| `Notifier.Acknowledge`·`Flush` 비테스트 호출자 | **0건** — 운영자 게이트 해제 경로가 배선돼 있지 않다 |
| `arm_suppressed_reason` | 671행 전부 비어 있음 — **P1·P3은 한 번도 발화한 적이 없다** |
| `effective_source` | `recomputed` 452, `saved` **0** — P3의 `SavedMonotone` 미발화 |
| `PROPOSAL_REFUSED` | **2 포지션 18행** (8/2 13건 + 8/5 5건). 재작성본은 8/5만 인용한다 |
| 8/2 사건의 원인 | 호가 단위가 아니라 **`applyFloor`의 확정 하한 0** (`exit.proposal_capped` 13건, `severity: normal`, outbox 행 0) |

### 판정과 다음 (2라운드)

**FREEZE 거부.** 차단 결함 7건(C1~C7).

가장 중요한 것은 C1·C2다. **이 change가 겨냥한 것보다 위쪽에 더 큰 결함이 있다** —
보유 포지션이 시세 응답을 못 받으면 무기한 조용히 청산 루프에서 빠지고, 계정 단위
outage 시계는 그것을 볼 수 없다. 판정이 없으면 제안도 없고, a089가 셀 대상 자체가 없다.

권고하는 재범위:

- **a090(신설, 선행)** — P0 계측. "이번 주기에 관측되지 않은 보유 포지션"을 종목 단위로
  세고, 연속 미관측이 한계를 넘으면 critical. 이것이 없으면 8/5가 무엇이었는지 다음에도
  알 수 없다
- **a089(축소)** — outbox 재발 장부 **하나만**. `DELIVERED` 재개방 한정,
  `parkAlert` 제외(C6), `outbox_test.go` 단언 반전 명시(C7), `ACKNOWLEDGED`는 범위 밖
  (원장 0행 + 호출자 0건이므로 도달 불가)
- **시계·계수** — C5가 해소될 때까지 보류. "살아 있는 보호 주문"의 판정 출처를 먼저
  정해야 하고(`LiveOrdersForSymbol`이 순수 SQL 읽기라 §0.4 비용 0), P2·P3·InDoubt를
  제외하거나 별도 등급으로 분리해야 한다

### 3라운드 — Function Logic Map이 바꾼 것 (2026-08-06)

사용자 지시로 건너뛴 4·5·6단계를 실행했다. `analysis/function-logic/`에 8함수 산출물이 있다.
아래는 **AST 열거가 리뷰 결론을 바꾼 부분만**이다.

**① 한 주기의 이탈은 26개다.** `ObserveOnce` 6 · `judgeLadder` 6 · `record` 5 · `submit` 9.
주문이 살아나는 것은 `submit:1295` 하나, 존재 미상이 `:1300` 하나, 통과가 3개.
**나머지 21개가 "보유 포지션이 새 주문 없이 주기를 끝내는" 지점**이다. P표는 9개를 적었다 —
C2("전수가 아니다")는 맞았고 과소평가였다.

**② P표의 함수 이름 4줄이 틀렸다.** `judge:1145`·`:1171`·`:1190`은 전부 **`record`** 안이다.
`judge`는 807-840의 다른 함수다. `judgeLadder`의 이탈 6개는 표에 한 줄도 없었다.

**③ 지연 시계는 고장난 것이 아니라 다른 원인을 재고 있다.** 그리고 **기존 테스트가 있다** —
`exitloop_test.go:883` `TestAnUncancellableEntryWithholdsTheLiquidationAndAlertsPastTheBound`가
31초 경과 후 `EventExitLiquidationDelayed`를 critical로 확인한다(`:900`,`:906`,`:910`).

| 미제출 원인 | `clearTheSymbol` | 시계 |
| --- | --- | --- |
| working order를 못 치웠다 (`record` B7 `:1145`) | 실패 | **시작·누적** — 위 테스트가 증명 |
| 브로커가 거부했다 (`submit` B10 `:1304`) | **성공** (거부된 주문은 살아 있지 않다) | `:1150`이 매 주기 초기화 |

따라서 **`:1150`을 그냥 제거하면 안 된다** — B7이 시작한 시계의 유일한 해제점이다.
해제를 `submit`의 `StateConfirmed`로 옮기면 `:914-919`가 그 대체의 회귀 테스트가 된다.

**④ B8(`StateInDoubt`)과 `record` B12·B14를 계수에 넣으면 안 된다.** C5 확정 —
`:1297-1299`가 "The order may exist. The proposal stays armed"라고 명시하고,
`pending_action`은 접수 성공 후에도 남는다(`StateConfirmed`가 `release`를 부르지 않는다).

**⑤ 계측의 올바른 자리는 `ObserveOnce`의 `states` 순회(B5 `:453`)다.** 그것이 보유 포지션을
주기마다 정확히 한 번 보는 유일한 지점이고, 하류 21개 이탈 중 어느 것으로 끝나든 이미 지나갔다.
**하류를 개별 계측하는 P1~P9 설계가 이 한 지점으로 대체된다.**

**⑥ B6(`:455`)의 테스트 부재가 기존 테스트에 가려져 있었다.**
`TestAQuoteWithNoLastTradeIsNotAnObservation`(`:585`)은 **보유 1종목**만 세우므로
`len(out)==0` → `observe` 오류 → **B4**를 단언한다. 보유 2종목 중 1종목만 미응답이면
`len(out)==1` → 오류 없음 → **B6 `continue`** — 이 경로는 한 번도 실행되지 않는다.

**⑦ 4단계만으로는 못 잡았다.** `codegraph callees ObserveOnce`는 8개를 주며 **`o.observe`를
빠뜨린다.** hard evidence와 AST는 대체재가 아니라 보완재다.

**⑧ `record` B4·B7과 B14의 실측 빈도는 0이다.** `arm_suppressed_reason` 671행 전부 공백,
`effective_source`에 `saved` 0건 `[미측정]`.

### 네 번 연속 같은 형태로 틀렸다

| 라운드 | 틀린 방식 |
| --- | --- |
| a087 초안 | 실측 1건(n=1)에서 브로커 성질을 단정 |
| a087 교체본 | 선례를 조건 없이 일반화, 호출 사슬 미추적(`internal/trading`) |
| a089 초판 | 로그 침묵에서 사건을 추론 |
| a089 재작성본 | **흔적이 남는 경로만 조회하고, 흔적이 안 남는 경로를 빠뜨렸다** |

네 번째는 앞의 셋과 다르다. 앞의 셋은 규율이 없어서 틀렸고, 이번엔 **규율을 문서에 적어
놓고 그 규율이 겨냥한 자리를 비켜 갔다.** 다음 조회는 "흔적이 남는 곳"이 아니라
**"흔적이 남지 않는 경로가 무엇인가"**부터 연다.
