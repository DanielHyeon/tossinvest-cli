# Function Logic Map — verify-observes-the-trigger

기준 HEAD `db6d4c3`. 대상은 **기존 함수 내부 로직이 바뀌는 전부**와, 그 함수를 통해
판정이 바뀌는 소비자다. High-risk이므로 면제 없음 (`.claude/CLAUDE.md` 3).

읽은 근거: `internal/verifylive/{record,cleanup,mutate,runner,plan,steps,retry,report,verifylive}.go`,
`cmd/tossctl/verify.go`, `internal/console/{pages,data}.go`.

---

## 0. 계약과 어긋난 것 두 가지 (읽고 나서 발견)

### 0.1 상한은 주문이 아니라 조건주문에 걸린다 — `MaxLiveOrdersTrigger` → `MaxLiveConditionalsTrigger`

design.md D6과 tasks 1.22–1.23은 `MaxLiveOrdersTrigger`를 적었다. 코드를 읽으니 **주문
상한은 이 단계에서 발동하지 않는다.**

- 주문 상한을 소비하는 것은 `checkOrderCap`뿐이고, 호출부는 `placeOrder`와
  `conflictProbe` 두 개다 (mutate.go:197, :262). 발동 단계는 둘 다 부르지 않는다 —
  **child 주문은 이 도구가 접수하는 것이 아니라 브로커가 만들고 우리가 발견하는 것**이다
  (`triggeredOrderId`를 읽어 `sr.created`로 등록).
- 실제로 막히는 것은 **조건주문 상한**이다. `MaxLiveConditionals = 1` (mutate.go:83),
  `checkConditionalCap`은 `createConditional`이 부른다 (mutate.go:483). 발동 단계는
  자기 조건주문을 **등록**하므로(D4 "자기가 방금 등록한 것을" 취소) 여기에 걸린다.
  전체 실행에서는 `conditional-register`가 남긴 것이 이미 1건 살아 있다.

→ `MaxLiveConditionalsTrigger = 2`로 구현하고 design.md D6·tasks·spec을 고친다.
주문 상한은 **손대지 않는다** — 그리고 손댈 필요가 없는 이유가 D2다: 체결이 종결이 되면
child가 상한을 스스로 비운다.

### 0.2 `sweepStep`이 붙잡힌 child를 취소해 버린다

cleanup.go:277. `sr.abort == nil && ctx.Err() == nil`이면 이 단계가 만들고 남긴
`Kind == "order"` artifact를 전부 취소한다. D4의 결말 ②(발동은 봤으나 체결 미확인 →
`fail`, child를 붙잡은 채 보고)는 `fail`일 뿐 `abort`가 아니므로 **sweep이 돌고 child를
취소한다.** 계약이 "붙잡힌 채 보고된다"고 적은 것이 코드에서 성립하지 않는다.

→ `sweepStep`에 `a.Deliberate` 스킵을 넣는다. 기존 기록·기존 단계에는 **무영향**이다:
`markHeld`는 오늘 조건주문에만 붙고, sweep은 이미 `Kind != KindOrder`를 건너뛴다.

---

## 1. `outstandingLines` — record.go:494 (**직접 수정, High-risk**)

살아 있는 객체를 정하는 단 하나의 함수. `Outstanding`·`cleanupFrom`·`liveCount`·
`liveConditional`·`sweepStep`·`report`·콘솔이 전부 여기서 나온다.

### 현재 로직

| # | 분기 | 조건 | 결과 |
|---|---|---|---|
| L1 | 첫 등장 | `key = Kind\x00ID`가 `latest`에 없음 | `order`에 key 추가(출력 순서 고정) |
| L2 | 역행 가드 | `prev.Cancelled && !a.Cancelled` | `continue` — 되살리지 않는다 |
| L3 | 갱신 | 그 외 | `latest[key] = {a, i}` |
| L4 | 출력 | `!l.Cancelled` | 결과에 넣는다 |

### 변경 후

| # | 분기 | 조건 | 결과 |
|---|---|---|---|
| L1 | 그대로 | | |
| L2′ | 역행 가드 **확장** | `terminal(prev) && !terminal(a)` | `continue` |
| L3 | 그대로 | | |
| L4′ | 출력 | `!terminal(l.Artifact)` | 결과에 넣는다 |

`terminal(a) = a.Cancelled \|\| a.Filled`.

L2′가 D2의 **단조성 주의**다. 확장하지 않으면 체결 뒤에 쓰인 줄(예: 같은 실행의 나중
관측 줄, `verify abort`의 나열)이 객체를 되살린다 — M22가 취소에서 겪은 형태 그대로다.

### Branch Test Map

| 분기 | 테스트 | 기대 |
|---|---|---|
| L2′ 취소 역행 | `TestCancellationIsMonotone`(기존) | 변화 없음 — 회귀 확인 |
| L2′ 체결 역행 | **RED 1.5** 체결 줄 뒤 `Filled:false` 줄 | 여전히 종결 |
| L2′ 교차 | **RED 1.5** 체결 줄 뒤 `Cancelled:true` 줄 | 종결(둘 다 종결이므로 갱신되어도 무해) |
| L4′ 체결 | **RED 1.4** `Filled:true` artifact | `Outstanding`에 없다 |
| L4′ 무필드 | **RED 1.6** `Filled` 없는 기존 줄 | 오늘과 동일 |
| L1 순서 | 기존 | 변화 없음 |

---

## 2. `Outstanding` — record.go:458 (간접)

본문 무변경. `outstandingLines` 위임. 판정만 넓어진다.
소비자 전수 (grep 확인):

| 소비자 | 위치 | 체결된 child에 대한 새 동작 |
|---|---|---|
| `cleanupFrom` | cleanup.go:125 | 정리 대상에서 사라진다 — **G2의 핵심** |
| `liveCount` | mutate.go:626 | 노출 상한을 비운다 |
| `liveConditional` | steps.go:1021 | 발동한 조건주문을 더는 살아 있다고 하지 않는다 |
| `sweepStep` | cleanup.go:282 | 체결된 것을 취소 시도하지 않는다 |
| `cancelLiveOrders` | steps.go:1000 | 무한 루프 방지 — 체결된 것을 계속 고르지 않는다 |
| `Runner.outstanding` → `Summary.Outstanding` | runner.go:604 | 실행 끝 잔여물 0 |
| `undeliberate` | runner.go:783 | 체결은 애초에 목록에 없다 |
| `BuildReport`/`BuildProgress` | report.go:210, :331 | `verify status`가 없는 주문을 안 찍는다 |
| 콘솔 | data.go:315, templates | 같은 이유 |

**회귀 위험**: 없음. 오늘 어떤 기록에도 `Filled` 필드가 없으므로 `terminal()`은
`Cancelled`와 동일하게 평가된다. 2.3의 실기록 재생이 이것을 실측으로 고정한다.

---

## 3. `cleanupFrom` — cleanup.go:123 (간접, **High-risk 경로**)

본문 무변경. `outstandingLines` 결과가 줄어들 뿐.

| # | 분기 | 조건 | 결과 | 변경 |
|---|---|---|---|---|
| C1 | 게이트 없음 | `holdGate(a) == ""` | 대상 | — |
| C2 | 게이트 해제 | `settled(gate) && heldAfter(...)` | 대상 | — |
| C3 | 붙잡힘 | 그 외 | 제외 | — |

체결된 child는 C1~C3에 **도달하지 않는다**. 오늘은 C1로 떨어져(주문의 기본 게이트는
`""`) 승인 목록에 오른다 — 이것이 G2의 세 번째 증상이다.

`holdGate`는 변경 없음. child는 `HeldUntil = StepConditionalTrigger`를 **명시**로 들고
오므로 기본값 분기를 타지 않는다.

### Branch Test Map
| 분기 | 테스트 | 기대 |
|---|---|---|
| C1 체결 | **RED 1.4** | `PendingCleanup`에 없다 |
| C3 붙잡힘 | **RED 1.13** | 판정 전 child는 정리 대상 아님 |
| C2 | 기존 `cleanup_test.go` | 회귀 없음 |

---

## 4. `sweepStep` — cleanup.go:277 (**직접 수정**)

| # | 분기 | 조건 | 오늘 | 변경 후 |
|---|---|---|---|---|
| S1 | 중단 | `sr.abort != nil \|\| ctx.Err() != nil` | return | 동일 |
| S2 | 종류 | `a.Kind != KindOrder` | continue | 동일 |
| S2′ | **의도적 보존** | `a.Deliberate` | (없음 — 취소한다) | **continue** |
| S3 | 취소 실패 | `cancelOrder != nil` | observe + return | 동일 |

S2′는 기존 동작에 무영향(§0.2 근거). RED 1.17이 이것을 고정한다.

---

## 5. `checkConditionalCap` — mutate.go:608 (**직접 수정**)

`checkOrderCap`(mutate.go:596)의 `StepIdempotencyTTLEdge` 분기가 그대로 선례다.

| # | 오늘 | 변경 후 |
|---|---|---|
| K1 | `limit = MaxLiveConditionals` 고정 | `limit = MaxLiveConditionals`; `sr.step.ID == StepConditionalTrigger`이면 `MaxLiveConditionalsTrigger` |
| K2 | `n >= limit` → `ErrExposureCap` | 동일 |

`checkOrderCap`은 **손대지 않는다** (§0.1).

### Branch Test Map
| 분기 | 테스트 | 기대 |
|---|---|---|
| K1 발동 단계 | **RED 1.22** 조건주문 1건 살아 있는 상태에서 발동 단계 등록 | 통과 |
| K1 그 외 | **RED 1.22** 같은 상태에서 `conditional-register` | `ErrExposureCap` |
| K2 | 기존 | 회귀 없음 |

---

## 6. `preflightStatic` — runner.go:530 (**직접 수정, 판정 경로**)

가장 조심할 곳. `preflight`(런타임)와 `Plan`(승인 목록) **양쪽이 같은 함수를 부른다.**

### 현재 분기 순서
| # | 조건 | 결과 |
|---|---|---|
| P1 | `step.Deferred != ""` | **skip 안 함** — 본문이 자기 유예를 기록한다 |
| P2 | `step.OptIn != "" && !optedIn` | skip |
| P3 | `DependsOn` 미통과 | skip |
| P4 | `NeedsHolding && holdingSymbol == ""` | skip |
| P5 | 시장 불일치 | skip |

### 문제

발동 단계는 변경 후 `Deferred != ""` **와** `OptIn != ""` **와** `Mutates: true`를
동시에 갖는다. P1이 먼저 반환하므로 P2가 영영 실행되지 않고, 그러면:

- `Plan`(plan.go:558)이 skip 아님으로 보고 → `step.Mutates`가 참이므로
  **옵트인하지 않았는데 승인 목록에 라이브 조건주문 등록 줄이 오른다.**
- `preflight`(runner.go:499)가 `Mutates && !approvedStep` → 승인에 없다며 skip →
  오늘의 유예 관측이 **사라진다** (1.8의 바이트 동일성 위반).

P2를 P1 앞으로 옮기는 것은 **틀린 수정**이다: 옵트인 없을 때 오늘의 `deferred` 판정 +
세 개의 unverified 관측 대신 `skipped`가 기록되어 1.8과 task 2.6의 입력이 바뀐다.

### 변경 후

`Runner.deferredForm(step) bool` 도입:
```
deferredForm(step) = step.Deferred != "" && (step.OptIn == "" || !r.optedIn(step))
```
= "이 실행에서 이 단계는 유예된 형태로 돈다".

`Runner.mutatesNow(step) bool = step.Mutates && !r.deferredForm(step)`.

| # | 조건 | 결과 |
|---|---|---|
| P1′ | `r.deferredForm(step)` | **skip 안 함** (오늘과 동일) |
| P2 | `step.OptIn != "" && !optedIn` | skip — 이제 옵트인+비유예 단계만 도달 |
| P3–P5 | 그대로 | |

그리고 **두 소비자**를 `mutatesNow`로 바꾼다:

| 위치 | 오늘 | 변경 후 |
|---|---|---|
| `preflight` runner.go:499 | `step.Mutates && !approvedStep` | `r.mutatesNow(step) && !approvedStep` |
| `Plan` plan.go:549,559,565 | `step.Mutates` | `r.mutatesNow(step)` |
| `entryFor` runner.go:618 | `Mutating: sr.step.Mutates` | `Mutating: r.mutatesNow(sr.step)` |

`entryFor`까지 바꾸는 이유: 옵트인 없이 도는 실행의 기록 줄이 `mutating:true`가 되면
1.8의 "바이트 단위로 같은" 미검증 관측이 아니게 된다. 그리고 그 줄은 실제로 아무것도
보내지 않으므로 `false`가 정직하다.

### Branch Test Map
| 분기 | 테스트 | 기대 |
|---|---|---|
| P1′ 유예형 | **RED 1.8** 플래그 없이 실행 | 판정 `deferred`, 관측 3개 오늘과 동일, `Mutating:false` |
| P1′ 유예형 + Plan | **RED 1.8** | 승인 목록에 발동 단계 줄이 **없다** |
| P2 옵트인 | **RED 1.9** 플래그 있음 | 유예 걷히고 목록에 등록·취소 줄이 오른다 |
| P1 다른 단계 | 기존 catalogue에 `Deferred`+`OptIn` 동시 보유 단계 없음 | 회귀 0 |
| P3 | `DependsOn` 제거(§8) | — |

---

## 7. `optedIn` — runner.go:584 (**직접 수정**)

| 오늘 | 변경 후 |
|---|---|
| `step.ID == StepIdempotencyTTLEdge && r.includeTTLEdge` | `switch step.ID { TTLEdge → includeTTLEdge; ConditionalTrigger → includeTrigger; default → false }` |

`default: false`가 fail-closed다. 새 옵트인 단계가 플래그 배선을 잊으면 실행되지 않는다.

---

## 8. `Steps()`의 `StepConditionalTrigger` 항목 — verifylive.go:443 (**직접 수정**)

| 필드 | 오늘 | 변경 후 | 근거 |
|---|---|---|---|
| `Mutates` | `false` | `true` | 조건주문을 등록·취소한다 |
| `Mutations` | 없음 | register-conditional + cancel-conditional | 승인 목록의 정본 |
| `NeedsHolding` | `false` | `true` | 1주를 팔 것이므로 보유가 있어야 한다 |
| `ActsOnConditional` | `false` | `false` | **자기가 등록한** 것을 본다 |
| `OptIn` | 없음 | `FlagIncludeTrigger` | |
| `Deferred` | 있음 | **유지** | 옵트인 없을 때의 문구 |
| `DependsOn` | `[conditional-register]` | **제거** | 남의 객체를 관측하지 않는다 |
| `Procedure` | 1줄 | 관측 절차 | |

`DependsOn` 제거는 오늘 동작에 무영향이다: P1이 먼저 반환해 P3에 도달한 적이 없다
(dead branch). `mutationSymbol`은 `NeedsHolding`으로 `holdingSymbol`을 준다 —
`ActsOnConditional`이었다면 `liveConditional()`이 **남의** 조건주문 종목을 줬을 것이다.

---

## 9. `stepConditionalTrigger` — steps.go:718 (**전면 재작성, High-risk**)

### 오늘
```
observe(trigger_observed=false) → observe(triggered_order_id_exposed=unverified)
→ observe(triggered_order_latency=unverified) → deferStep
```

### 변경 후

| # | 국면 | 조건 | 행동 |
|---|---|---|---|
| T0 | 유예형 | `r.deferredForm(step)` | **오늘 본문 그대로**, 즉시 반환 |
| T1 | 사전 | 매도가능 < 1주 | `skip` — 팔 것이 없다 (J3이 매도 전에 드러나는 자리) |
| T2 | 사전 | `NearStopTrigger` 에러(사이에 유효 틱 없음 등) | `skip`, 추측하지 않는다 |
| T3 | 등록 | `createConditional` | 실패 시 error → resolve |
| T4 | A국면 | 창 만료까지 `TriggerPollIdle` 간격 시세 폴링 | 임계 도달 관측 시 `condition_crossed_at` + B국면 |
| T5 | A국면 만료 | 미도달 | 자기 조건주문 취소 → **T9 재확인** |
| T6 | B국면 | `TriggerPollActive` 간격 조건주문 폴링 | 발동 흔적 → `trigger_first_observed_at` + bid/ask/last |
| T7 | B국면 | `triggeredOrderId` 노출 | `triggered_order_id_first_seen_at`, child를 `HeldUntil`+`ChainID`로 등록 |
| T8 | B국면 | child 체결 확인 | `child_order_filled_at`, child·조건주문을 `Filled`로 종결, `pass` |
| T8′ | B국면 만료 | 발동은 봤으나 체결 미확인 | `fail` — child는 붙잡힌 채 남는다(§0.2 수정이 이걸 지킨다) |
| T9 | **경합** | T5 취소 직후 재확인에 발동 흔적 | 미도달로 끝내지 않고 **T6로 전환** |
| T10 | 취소 성공 + 흔적 없음 | | `skip(INCONCLUSIVE)`, 잔여물 0 |

모든 시각은 `r.now()`이고, 각각 그때의 폴링 간격과 그 구간의 백오프 발생 여부를 함께
기록한다 (M44·M45 → 브로커가 시각을 주지 않는다).

### Branch Test Map
| 분기 | 테스트 |
|---|---|
| T0 | RED 1.8 |
| T1 | RED 1.14 매도가능 0 |
| T2 | RED 1.1/1.14 틱 없음 |
| T4→T6 국면 전환 | RED 1.10 |
| T6/T7 시각+간격+백오프 | RED 1.11 |
| T6 호가 기록 | RED 1.12 |
| T7/T8 hold·chain·체결 | RED 1.13 |
| T5→T10 | RED 1.15 |
| T5→T9→T6 | RED 1.16 |
| T8′ | RED 1.17 |

---

## 10. `WriteSteps` — report.go:368 (**시그니처 변경**)

`WriteSteps(w, includeTTLEdge bool)` → `WriteSteps(w, includeTTLEdge, includeTrigger bool)`.
호출부 2곳: `cmd/tossctl/verify.go:262`, `internal/console/pages.go:167`.
분기는 `if includeX && s.ID == StepY` 하나가 둘로 늘 뿐이다.

---

## 11. 새 함수 (FLM 불요, 신규)

`NearStopTrigger`, `Runner.deferredForm`, `Runner.mutatesNow`, `Runner.Abort`,
`Runner.watchTrigger`(T4–T9), `Artifact.terminal`.

---

## 12. 건드리지 않는 것

- `FarBuyLimit` / `FarSellLimit` / `FarStopTrigger` — 서명·본문 **무변경** (1.3의 AST 가드가 고정)
- `checkOrderCap` / `MaxLiveOrders` (§0.1)
- `holdGate` / `heldAfter` — 위치 기반 해제 규칙 그대로
- `RecordFormatVersion` — 필드 추가는 `omitempty`이므로 옛 리더가 무시한다
- `Order.SubmittedAt` (J1, M45) — 이 change는 쓰지 않을 뿐 고치지 않는다
