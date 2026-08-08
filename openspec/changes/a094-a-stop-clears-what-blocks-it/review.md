# a094 · review

## 1라운드 (proposal-freeze) — **FAIL**

- 렌즈 셋, 전부 Claude 보이스: **A 적대적 Eng** · **B 근거 대조** · **C 구현 가능성·spec 정합**
- **셋 다 독립 FAIL.** 저장소 무변경(`git status --short` 17항)을 셋 다 확인했다.

### 1.0 교차 모델 — 미충족 (a094 1라운드)

사용자의 명시적 지시("클로드로 돌리세요")로 세 보이스 모두 같은 모델 계열이다.
**a094 1라운드를 교차 모델 미충족으로 기록한다.** a092의 여섯 라운드 연속 미충족과는
별개 건이며, 이유(사용자 지시)는 같다.

이번 라운드에서 모델 편향이 실제로 드러난 자리가 있다 — **A와 C가 R1의 크기에 대해
정반대 결론**을 냈고(A: substring이라 불가 / C: 한 줄이면 된다), 그 충돌을 푸는 것은
보이스가 아니라 **이 change 자신의 spec delta**였다(§1.5). 렌즈 분리는 작동했으나,
같은 모델이 *구조적으로 못 보는 것*은 이번에도 시험되지 않았다.

---

## 1.1 증거 사슬은 깨끗하다 (먼저 적는다)

이 저장소는 「생성된 증거가 커버리지를 거짓 주장한다」로 여러 번 거부당했다.
**이번에는 그 계열이 없다.** 보이스 B가 전수 재현했다.

| 대조 | 결과 |
| --- | --- |
| 문서 인용 `파일:줄` | **60곳 중 60곳 일치** |
| 분기 ID·줄 쌍 | **38개 중 38개 일치** (각 줄의 실제 코드 내용까지) |
| FLM/BTM의 「조건 (원문)」 | **158행 중 158행이 소스와 문자 단위 일치** |
| FLM의 「창의 호출·return」 | **79행 중 79행이 `ast.json` 좌표와 정확히 일치** |
| `ast.json`의 `source_sha256` | **9/9이 현재 HEAD와 일치** — stale 없음 |
| 커버리지 (독립 재측정, 620초) | **분기 79 · 진입 49 · 미진입 27 · 블록없음 3 — 4개 수 전부 일치.** 함수별 미진입 **9/9 일치**, 미진입 분기 **집합까지** 문자 그대로 일치 |
| openapi 인용 | `opposite-pending-order-exists`가 **정확히 한 자리**(`POST /api/v1/orders` → 422)에만 존재. 경로 표기 문자 단위 일치 |
| 원장 인용 | attempt 3건의 상태·requestId 셋·시각 셋·`notes` 0바이트·BUY 0건·`fill_events` 0건·`exit_states` 수치 **전부 일치** |
| 소스 주석 인용 8건 | **전부 원문 일치** |

**따라서 이 판정이 기각하는 것은 근거가 아니라 처방이다.**

---

## 1.2 차단 — 문서가 정본을 지운다 (B·C 독립 수렴)

### 차단 1. `specs/order-execution/spec.md`의 MODIFIED가 기존 요구 본문과 시나리오 6개를 삭제한다

MODIFIED는 요구 블록을 **통째로 치환한다** — 보이스 B가 openspec 1.4.1의 구현으로
확인했다(`@fission-ai/openspec/dist/core/specs-apply.js:207-236`). 이 delta는 기존 본문을
「**종전 조항 (변경 없음)**」이라는 **산문 참조**로 대체했고, 참조는 본문을 보존하지 않는다.

적용되면 정본에서 사라지는 것:

- openapi `clientOrderId` 멱등키와 **유효 10분** TTL
- 재생 진입점의 자기 의무 전부 — IN_DOUBT 상태 확인, attestation 플래그
  `[미측정 — 2b 전 비활성]`, **재생 1회마다** `elapsed < TTL − margin` 재검사
  (margin 기본 60초), 마진 없는 경계 사용 금지(SHALL NOT), 회수 상한 기본 2회·최소 간격,
  wire body 외 본문 구성 불가(SHALL)
- **「재생 응답 분류에 dispatch 분류기를 사용해서는 안 된다(SHALL NOT)」** 와 그 매핑
  (`422 idempotency-key-conflict` → FAILED_CONFIRMED 금지 + UNRESOLVED + critical 알림)
- 조회 대조의 pagination 완주, `PARTIAL_FILLED` 근거, 멱등키로 매칭 불가(SHALL NOT)
- 부재 판정의 **연속 N회(기본 3회)**, 창 오염 시 자동 FAILED_CONFIRMED 금지(SHALL NOT)
- 해소 불능의 「해당 심볼 신규 진입 영구 차단, 운영자 해소만 허용」
- **시나리오 6개 전부.** delta는 새 시나리오 5개만 갖는다

**선례가 정반대다.** `openspec/changes/archive/2026-07-26-extend-execution-contract/`의
같은 요구 MODIFIED는 본문 전문과 시나리오 6개를 **전부 재현**했고, 그 텍스트가 지금
정본이다.

**게이트가 이것을 못 잡는다.** `openspec validate --strict`는 구조만 보고 보존을 보지
않는다 — 두 보이스가 직접 실행해 통과를 확인했다. 이 사실 자체를 기록한다.

---

## 1.3 차단 — R4는 안전 속성을 **제거한다**

### 차단 2. 미체결 매도의 부재를 거짓으로 확증한다 → 초과 매도 (A, Manager 재확인)

`absenceCorroborated`(`internal/execgw/indoubt.go:445-500`)의 증거는 둘뿐이다.

```go
bpDelta := buyingPowerNow - baseline.BuyingPower
if notional > 0 && bpDelta < 0 && math.Abs(bpDelta) >= notional*0.5 {
    return false, "...consistent with this order having been accepted"
}
return true, "the holding and the buying power are unchanged from the pre-dispatch baseline"
```

주석이 모델을 자백한다 — *"a reservation of roughly this order's notional"*.
**매수 예약 모델이다.**

접수됐지만 미체결인 **SELL**은 `Holdings.quantity`도 `CashBuyingPower`도 바꾸지 않는다.
따라서 baseline이 정확해도 두 검사가 **모두 통과**해 `return true` →
`ResolveFailed` → `FAILED_CONFIRMED`(`indoubt.go:339-346`) → 다음 주기가 손절을 다시 낸다
→ **살아 있는 매도 위의 두 번째 매도.**

이것은 이 저장소 전체가 막으려는 실패다(`internal/execgw/classify.go:17-20` —
*a wrong "it did not happen" duplicates a live order*).

**오늘 baseline이 없어서 항상 park하는 것이 안전측이었다.** R4는 그 안전측을 제거하면서
매도용 증거 모델을 만들지 않는다. `proposal.md`와 design D4는 R4를 **순수 이득**으로
서술했다 — **그것이 틀렸다.** §6(보수 방향만)에 정면으로 걸린다.

### 차단 3. R4의 값 원천이 소스에 없다 (A·C 독립 수렴)

`ExitObserverOptions`(`internal/app/engine/exitloop.go:166-223`) 전 필드를 열거하면
Journal·Prices·Retrier·Issuer·Submit·Alerts·Names·Log·Costs·Floor·SLO·Escalate·
Announcer·AccountRef·Clock·Interval·OutageAfter·DelayBound·Ratchet·Ladder·
CommonPolicy·NewID다. **보유수량·매수가능금액 원천이 0개다.**

- `execgw.Baseline`은 `BuyingPower`·`Holding`·`Currency`를 요구한다(`indoubt.go:95-102`)
- `journal.Position`에 **통화가 없다**
- 매수가능금액은 exit loop 어디에서도 읽지 않는다 — 읽는 곳은 `tracer.go:366`(진입)과
  `AccountSweep`(reconcile/filldetect 전용)뿐이고 공유되지 않는다
- `m.position.Quantity`는 **원장의 믿음**이고, `absenceCorroborated`는 그것을 브로커
  `Holdings` 합계와 뺀다. **두 수는 같은 것을 재지 않는다**(엔진 관리 3주 + 앱 보유 7주면
  delta 7이 나와 park하면서 거짓 사유를 남긴다)

**`BuyingPower = 0`으로 채우는 것은 더 나쁘다.** `bpDelta = buyingPowerNow − 0 > 0`이라
가드가 **영원히 발화하지 않고**, spec §3이 요구하는 「매수가능금액·보유수량 delta 교차
확인」의 절반이 **침묵으로 만족된다.** `Baseline`에는 "이 값은 미측정"을 말할 자리가 없다.

**판단: R4를 철회한다.** §1.7 참조.

---

## 1.4 차단 — R3의 진입점이 세션 중 원장을 위조한다 (A·C 독립 수렴)

`design.md`와 tasks 4.5는 *"배선은 이미 있다 — `runtime_wiring.go:175`의
`Context.Recovery`"*라고 썼다. 그것이 주는 `Recovery.Run`은 `RecoverPending`으로 시작한다
(`internal/journal/recovery.go:86-125`):

```go
case StateRecorded:        Settle(StateNotDispatched, "found at startup with no dispatch recorded")
case StateDispatchStarted: MarkInDoubt("process stopped after dispatch started; outcome unknown")
```

**세션 중 `RECORDED`·`DISPATCH_STARTED`는 지금 전송 중인 주문이다.** 그것을 "보낸 적
없음"으로 종결시키는 것은 원장 경합이 아니라 **원장 위조**다. 이어서 `stableSnapshot`
(반복 계정 조회) → `Comparer.Compare` → `Gate.Clear`/`Gate.Block`까지 돈다.
그리고 **`reconcile.New`는 생성자에서 `Gate.Block(ReasonRecoveryIncomplete)`를 건다**
(`internal/reconcile/recovery.go:141-157`) — 주기마다 새로 만들면 매번 진입이 잠긴다.

design D3의 안전 논거는 *"`Resolver`는 mutator를 갖지 않는다"* 하나였고 위 넷을
**하나도** 다루지 않았다. 「침묵한 생략 금지」에 직접 걸린다.

**옳은 진입점은 존재하는데 문서 어디에도 이름이 없다**: `Resolver.Resolve(ctx, attemptID)`
(`indoubt.go:229`) + `Journal.PendingAttempts`, `engine.Context.Resolver`
(`engine.go:185`)로 도달 가능하다.

### 1.4.1 그리고 「미측정」이라 적은 것이 이미 소스에 있었다

`DefaultResolveConfig`(`indoubt.go:156-166`): StableObservations **3** ·
MinObservation **45초** · PollInterval **5초** · MaxDuration **5분** · MaxPages **50**.
`scanBoth`는 관측 1회마다 OPEN·CLOSED **양쪽**을 완주한다(`indoubt.go:415`).

즉 **`Resolve` 1회 = 최소 45초 벽시계 · 최소 6회 페이지 조회**, attempt마다 순차다.
design D3은 *"주기와 동시 실행 상한은 실측으로 정한다"*고만 쓰고 이 상수들을 적지 않았다.
**미측정인 것은 주기이지 1회 비용이 아니었다.** 관측 goroutine에서 동기 호출하면 그
45초~5분 동안 모든 포지션의 관측이 멈춘다(§4).

### 1.4.2 R3의 이득 주장도 틀렸다

`park`의 `Gate.Block`은 **reason으로만 키잉**하므로 계정 전역 차단이다(심볼 범위는
`BlockSymbol`, `internal/execgw/symbolgate.go:64`). 게다가 `windowContaminated`가
세션 중에는 상례로 걸리므로 R3의 현실적 산출은 `FAILED_CONFIRMED`가 아니라 **park**다.

**진짜 이득은 따로 있다** — `checkSymbolFree`는 전면 차단에 `PendingAttempts`만 보고
UNRESOLVED는 그 조회에서 빠진다(`journal/recovery.go:30-33`). 따라서
**IN_DOUBT → UNRESOLVED 전환 자체가 그 심볼의 취소·매도를 푼다.** tasks 4.1이 주장해야
할 것은 그것이다.

---

## 1.5 차단 — R2

### 차단 4. 배선이 없고 tasks가 그것을 추가하지 않는다 (C)

`OrderPager`(`indoubt.go:70-74`)는 `reconcile`·`filldetect`·`flatten`에만 배선돼 있고
`Context.ExitObserver`(`exitwiring.go:319-348`)가 채우는 필드에 없다. tasks 3.9의
「새 API 표면 없음」은 **브로커 API에 대해서만 참이고 엔진 API에 대해서는 거짓**이다.
실제로 필요한 것: `ExitObserverOptions` 새 필드 · nil 허용 여부 결정
(필수로 하면 `NewExitObserver`의 거부 switch가 미배선 빌드의 기동을 막는다) ·
`Context.ExitObserver` 한 줄 · `cmd/tossctl/engine.go` 구성. **파싱 함수도 미지정이다**
(`brokerstate.ParseOfficialOrder` / `execgw.ScanOrders` 중 무엇인지, 어떤 status 그룹을
미체결로 볼 것인지).

### 차단 5. 브로커 오류가 판정 전체를 중단시킨다 (C)

`record`(`exitloop.go:1141-1144`)는 `cleared, err := o.clearTheSymbol(...)` 다음
`if err != nil { return err }`다. 브로커 목록 오류를 같은 자리로 반환하면 브로커 두절
한 번이 `RecordExitJudgementResult` **전에** 판정을 중단시킨다 — 워터마크도 기준선도
전진하지 않고 `noteDelay`도 울리지 않는다. `exitloop.go:1114-1118`이 명시한
「clear 실패는 판정을 멈추지 않는다」 계약보다 **엄격히 나쁘다.**

### 차단 6. R2는 R1의 안전망이 아니다 (A) — 그리고 design의 술어 반전이 그것을 가렸다 (B)

`release(ProposalRefused)` → `pending_action = NULL`(`journal/apply_hook.go:847-848`)
→ 다음 주기 `CancelPendingFirst = false` → 청소는 `isFullExit`으로만 열려
`clearTheSymbol(..., withPending=false)`로 불린다 → `if !buy && !withPending { continue }`
→ **매도 주문을 건너뛴다.** 목록을 브로커로 넓혀도 결과는 같다.

R1이 "이 409는 접수 안 됨을 확정한다"고 판단했는데 사실 그 매도가 살아 있었다면,
다음 주기의 R2는 브로커 목록에서 그것을 **보고도 지나치고** `clear=true`로 새 매도를 낸다.
design D5 다이어그램이 정확히 이 순서를 그리면서 `withPending`이 false인 것을 놓쳤다.

**그리고 design의 R2 설계표가 그 술어를 뒤집어 적었다**(§1.6 오류 1) — 「지금의 눈」을
잘못 읽은 표가 이 구멍을 가린 셈이다.

### 차단 7. §0.3 판정이 같은 문서 안에서 비대칭이다 (A·C 독립 수렴)

`isFullExit`은 `ActionLadderStop`·`ActionBaselineBreach`를 포함하므로 **모든 손절 제출
직전에** `clearTheSymbol`이 돈다. 지금 그 목록은 로컬 SQLite다. R2는 거기에 pagination
브로커 왕복을 넣는다. 그런데 design D4는 **R4에 대해** *"계정 읽기를 손절 제출 경로에
동기로 끼워 넣으면 그 읽기의 지연이 손절의 지연이 된다"*고 금지했다. **같은 경로, 같은
성질, 다른 판정.** tasks 7.2는 「제출 시점이 늦어지지 않음을 호출 수로 보인다」고
약속하는데 R2에 대해서는 **보일 수 없다.** 선례가 옆에 있다 —
`precheckTimeout = 2 * time.Second`(`internal/app/engine/precheck.go:23-43`).

**계약 충돌도 있다**: spec delta는 *"브로커 미체결을 포함해야 한다(SHALL)"*로 폴백을
허용하지 않는데 tasks 3.7은 *"조회 실패해도 저널분 청소는 진행"*을 요구한다. 둘은 동시에
참일 수 없다. 그리고 3.7은 **실패**만 다루고 **지연**은 다루지 않는다 — 30초 타임아웃은
실패가 아니라 30초 늦은 손절이다.

### 차단 8. 손절을 영구 보류시킬 수 있는 주문의 모집단이 무한해진다 (C-H3, A-H2)

`clearTheSymbol` B6·B7이 확정 취소 실패를 `clear=false`로 만들고 `record`가 발의를
보류한다. **오늘 후보는 엔진 자신의 확정 주문뿐이다.** R2 이후 후보는 브로커가 보고하는
무엇이든이고, 엔진이 영원히 취소할 수 없는 주문이 손절을 **영구 보류**시킨다.
구체 경로 둘: (a) 목록 조회와 취소 사이의 체결 → 취소 거절 → `clear=false`
(오늘이라면 그 체결로 반대 주문이 사라져 손절이 나갔을 상황이다),
(b) `floatOf`(`exitloop.go:1673-1679`)가 `ParseFloat("")`을 거절하므로 **가격 없는 주문**이
목록에 오면 `clear=false`가 고정된다.

§0.3상 명시적 결정이 필요하다 — N회 실패 후 손절을 그냥 내보내는가(409를 받더라도),
계속 보류하는가. tasks 3.3은 동작만 고정하고 **그 결과가 무보호**임을 다루지 않는다.

---

## 1.6 문서가 실물과 다른 곳 (B, Manager 재확인)

| # | 문서가 말한 것 | 실물 |
| --- | --- | --- |
| 1 | `design.md` R2 설계표: **B3 — 매수는 `withPending`일 때 치운다** | **반전이다.** `if !buy && !withPending { continue }` — 매수(`buy=true`)는 **항상** 치워지고, `withPending`이 필요한 것은 **자기 방향(매도)**이다. 같은 문서의 경계표와 proposal은 옳게 쓴다 — **문서가 자기와 모순되고 틀린 쪽이 R2 설계표다** |
| 2 | *"272210은 `LADDER_PARTIAL → PROPOSAL_CANCELLED`를 **22시간 동안 5초 주기로** 반복"* | `LADDER_PARTIAL` 326건 중 **325건이 42분 창**에 몰려 있고, 그 사이 21.7시간의 exit_events는 14건(전부 `action=NULL`)뿐. `PROPOSAL_CANCELLED` 1931건은 **약 2시간 54분**. 루프의 **주 action은 `STOP_LOSS_LADDER`(1606건)**이고 `LADDER_PARTIAL`은 17%다 — **소수 쪽 이름을 루프의 이름으로 썼다.** 주기 5초는 맞다(중앙값 5.0초). attempt가 22시간 미정산인 것도 맞다. **틀린 것은 둘의 결합이다** |
| 3 | *"그동안 가격은 **−5.2%**까지 갔고"* + 출처로 `원장(...)` 표기 | 475150의 `exit_events` 9건의 `observed_price` 범위는 **57,700~59,000**(최저 = entry 대비 **−0.35%**). 마지막 관측이 동결 시점이고 **그 이후 관측이 없다.** −5.2%는 원장 어디에도 없다 — 가격 시계열 테이블 자체가 없다. **사용자 보고값을 원장 인용처럼 적었다** |
| 4 | tasks 1.10: *"`reconcile.Run`의 미진입 7개는 해소 경로(B5~B9)에 몰려 있고"* | 미진입은 `B1,B2,B4,B5,B8,B9,B11`. B5~B9에 드는 것은 **3개뿐**이고 나머지 4개는 해소 경로 밖이다 |
| 5 | *"`Baseline`을 넘기는 호출자가 **하나도 없다**"* (근거 2곳) | 비테스트 `PlaceRequest{}` 생성은 **6곳**이고 `strategy_gateway.go:65`는 **실제로 `Baseline: req.Baseline`을 전달한다.** 결론(실전에서 항상 nil)은 성립하나 **전칭 주장의 열거가 불완전하다** |
| 6 | *"`CancelPendingFirst`는 `ladder.go:447`에서 정해진다"* | `ratchet.go:432`도 같은 필드를 정한다(술어 동일, 결론 유지) |
| 7 | *"`Resolver`의 필드는 Journal·Orders·Order·Account·Clock·Gate**뿐**"* | `Config ResolveConfig`도 있다. mutator가 아니므로 안전 결론은 유지되나 「뿐」이 틀렸다 |
| 8 | tasks §1 헤더 *"산출물 (완료 — **문서보다 먼저**)"* | `ast.json` 9개는 proposal보다 **먼저**다(10:09~10:11 vs 11:33) — 분기 주장의 근거 순서는 지켜졌다. 그러나 FLM·BTM **18개는 전부 12:12**로 proposal보다 **38분 늦다.** 헤더의 전칭이 성립하지 않는다 (mtime이 유일 증거임을 함께 적는다) |
| 9 | 패키지 미한정 basename 7종 (`gateway.go`·`recovery.go`·`flatten.go`·`replay.go`·`dispatch.go`) | 각각 3~4개 패키지에 동명 파일이 있어 자동 해석이 실제로 오해석했다. 문맥으로는 풀리나 좌표 인용의 목적이 재검증이라면 패키지를 붙여야 한다 |
| 10 | proposal이 `STORY-TOS-a094`를 선언 | 파일 부재. tasks 7.9가 PM 동기화를 미완으로 두므로 예정된 미완일 수 있다 |

---

## 1.7 판정과 다음 판이 받는 것

**FAIL.** 2판으로 간다.

### R1 — 유지. 단 크기가 3줄이 아니다

**A와 C가 갈렸고, 답은 이 change 자신의 spec에 있다.** C는 *"`containsAny` 한 줄이면
된다"*(실제 payload 기준, 그리고 저널 CHECK 제약도 콘솔 switch도 없음을 grep으로 확인 —
**데이터 마이그레이션 불필요**). A는 *"본문 통짜 substring이라 tasks 2.5를 통과할 수 없다"*.

둘 다 맞다. 이 delta가 *"message 문구로 걸어서도 안 된다(SHALL NOT)"*를 썼으므로
**`error.code` 필드 파싱이 강제된다.** 한 줄로는 자기 SHALL NOT을 만족할 수 없다.

- 필드 파싱으로 구현하고, `containsAny` substring의 취약성을 D0에 **추가로** 적는다
  (`"interactive"` 같은 기존 마커가 그 증거다)
- **`internal/execgw/testdata/reason_codes.golden` 갱신 task를 더한다**
  (`TOSSOS_UPDATE_GOLDEN=1 go test -run TestWriteReasonCodeGolden`)
- **재생 경계를 spec에 못 박는다**: 재생 응답에는 이 code 분류를 적용하지 않는다
  (SHALL NOT). 현재 `classifyReplay`가 코드를 공유하지 않아 우연히 안전하지만,
  재생 attestation이 켜지는 날 R1이 조용히 반대 방향으로 작동한다

### R2 — 재작성

- 배선 4곳(옵션 필드·nil 정책·`Context.ExitObserver`·`engine.go` 구성)과 파싱 함수를
  **문서가 지목한다**
- 브로커 오류는 `clearTheSymbol` **내부에서 흡수**하고 저널분 청소는 계속한다.
  spec의 SHALL을 폴백 가능한 형태로 고친다
- **타임아웃·페이지 상한을 §0.4 예산으로 정한다**(`precheckTimeout = 2s` 선례)
- **자기 방향 미체결의 부재 확인**을 `withPending`과 무관하게 요구한다 — 확인 못 하면
  제출하지 않는다(차단 6)
- **취소 불가 주문이 손절을 영구 보류시키는 문제에 명시적 결정**을 내린다(차단 8)
- design의 B3 술어 반전을 고친다

### R3 — 진입점 교체

- `Resolver.Resolve` + `Journal.PendingAttempts`를 직접 부르는 새 `SupervisedLoop`.
  **`reconcile.Run`·`RecoverPending`·`Recovery` 생성을 재사용하지 않는다(SHALL NOT)**
- **`Resolve` 1회 = 최소 45초 · 최소 6회 조회**를 문서에 적는다. 미측정인 것은 주기다
- **세션 중 park가 계정 전역 진입 차단을 건다**는 운영 결과를 공시한다
- 이득 주장을 정정한다 — 산출은 `FAILED_CONFIRMED`가 아니라 park이고, 이득은
  **UNRESOLVED 전환이 그 심볼의 취소·매도를 푸는 것**이다

### R4 — **철회한다**

값 원천이 없고(차단 3), 억지로 채우면 **살아 있는 매도를 은퇴시킨다**(차단 2).
오늘의 「항상 park」가 안전측이며, 그것을 제거하는 변경은 §6(보수 방향만)에 걸린다.

**대신 남기는 것**: 「부재 확증의 증거 모델은 매도에 대해 아무것도 증명하지 못한다」는
사실을 `issues.md`에 기록하고, 매도용 증거 모델(예: 체결 이벤트 부재 + 브로커 목록
완주의 결합)을 별도 change의 선행 조건으로 세운다. **baseline은 그 모델이 생긴 뒤에
공급한다.**

### spec delta — 재작성

기존 요구 **전문과 시나리오 6개를 재현**하고 그 뒤에 새 조항·시나리오를 덧붙인다.
`openspec validate --strict`가 이 손실을 잡지 못한다는 사실을 tasks의 게이트 절에 적는다.

### 문서 정정

§1.6의 10건. 특히 **1·2·3은 사실 오류**이므로 2판 착수 전에 고친다.

---

## 1.8 막힌 시도 — 설계가 실제로 막은 것

세 보이스가 깨뜨리려 했으나 실패한 것. 설계의 강도를 재는 값이므로 함께 적는다.

1. **R3이 손절을 막는가** → 아니다. `EntryGate` 주석 *"Blocks apply to new entries only.
   Cancels and liquidations are never gated"*, `risk.Evaluate`가 `SideSell`을 entry chain
   밖으로 뺀다
2. **R1이 재생 분류를 오염시키는가** → 아니다. `classifyReplay`가 `classifyMutation`과
   코드를 공유하지 않는다 (계약 공백은 남으므로 §1.7에서 spec에 못 박는다)
3. **`Resolver`가 mutator를 갖는가** → 아니다
4. **R2가 노출을 늘릴 수 있는가** → 아니다. `Gateway.Cancel`이 `raisesExposure: false`로
   고정이고 `clearTheSymbol`은 `Cancel`만 부른다
5. **R1이 409 전체를 확정으로 만드는가** → 아니다. `isDefinitiveRejection`을 건드리지 않고
   B3가 B5보다 먼저 도는 것도 소스대로다
6. **R1이 전면 차단을 푸는가** → 그렇다. `PendingAttempts`에 FAILED_CONFIRMED는 없다
7. **R1이 알림 폭주를 만드는가** → 아니다. `alertProposalRefused`가
   position+action+level을 Key로 쓴다
8. **R2의 비엔진 주문 취소를 게이트웨이가 허용하는가** → 허용한다. `Gateway.Cancel`은
   부모 intent를 조회하지 않고, WTS 없는 precheck도 lineage 없는 주문을 이미 다룬다
9. **exit-policy ADDED가 기존 요구와 충돌하는가** → 아니다. 「발의 수명주기」·「관측 경로와
   fail-safe」 어느 것도 치우기 대상 목록의 원천을 규정하지 않고, ADDED가 스스로
   `record` B3 게이트 보존을 SHALL NOT으로 적는다
10. **게이트 산출물 누락** → 없다. `tools/gate.sh`가 요구하는 tasks·review·issues·
    check_analysis를 0.5와 6.3이 만든다

---

## 1.9 세 보이스가 확인하지 못한 것 (침묵한 생략 아님)

- **A**: 원장 실측 전부(문서 주장을 인용만 했다) · 브로커가 이 code를 낼 때 부분 접수
  후 되돌리는지 · `OrdersPageRaw`의 실 지연·rate limit 예산 · KRX 예약·조건부주문이
  OPEN 목록에 나오는지 · `ast.json` 좌표 1:1 재대조 · a087·a089·a091·a092의 delta 본문
- **B**: `−5.2%`의 외부 출처 · 네 change와의 충돌 · `-race` 회귀·`make sdd-check`·
  `make gate`(mutating) · 산출물 생성 순서는 **mtime만이 증거**(사후 touch면 판정이 달라진다)
- **C**: `go test ./...` 미실행 · 커버리지 재현 안 함(B가 했다) · 네 change delta 미대조 ·
  `official.OrdersFilter`의 status 그룹이 두 시장 모두 서버측 심볼 필터를 지원하는지

---

## 2판 — 1라운드 지시의 반영 (판정 아님)

**이 절은 판정이 아니라 반영 기록이다.** 2라운드 리뷰는 아직 돌지 않았다.
§1.7이 지시한 것과, 그것을 어디에 어떻게 반영했는지를 대조 가능하게 적는다.

### 2.1 R1 — 필드 파싱으로 바꿨다

| §1.7 지시 | 반영 |
| --- | --- |
| 필드 파싱으로 구현 | `design.md` D1「그래서 R1은 함수 하나를 더한다」— `classifyRefusalCode` 신설. `tasks.md` 2.8 |
| `containsAny` substring 취약성을 D0에 추가 | D1「왜 한 줄로는 안 되는가」에 적었다. `"interactive"`가 그 증거 |
| `reason_codes.golden` 갱신 task | `tasks.md` **2.10** — `TOSSOS_UPDATE_GOLDEN=1 go test -run TestWriteReasonCodeGolden` |
| 재생 경계를 spec에 SHALL NOT | `specs/order-execution/spec.md` — *"이 분류를 재생 응답에 적용해서는 안 된다(SHALL NOT)"*. `tasks.md` 2.11 |

**그리고 지시에 없던 것이 측정에서 나왔다 — 본문 모양이 하나가 아니다.**

`mutation_attempts.detail`의 프로덕션 3건은 `{"error":{"requestId":…,"code":…}}`이고
기존 fixture 두 개(`interactive_auth_challenge.json`·`fx_consent_required.json`)는
최상위 `{"code":…}`다. 표기도 다르다(lower-hyphen vs UPPER_SNAKE).

이것이 1라운드의 「`error.code` 필드 파싱」 지시를 **그대로 쓰면 안 되는** 이유다 —
`error.code`만 읽는 파서는 기존 fixture 모양을 놓친다. spec 문장을
「`code` 필드 값(최상위와 `error.code` 둘 다 읽는다)」으로 고쳤고, `tasks.md`
2.5a~2.5d가 네 경우를 나눠 시험한다. **1라운드 판정문의 문구 하나를 측정이 좁혔다.**

### 2.2 R2 — 재작성했다

| §1.7 지시 | 반영 |
| --- | --- |
| 배선 4곳과 파싱 함수를 문서가 지목 | `design.md` D2「배선」표 4행 + 「파싱」표. `tasks.md` **3.A1~3.A5 · 3.B1~3.B5** |
| 브로커 오류는 `clearTheSymbol` 내부에서 흡수, spec의 SHALL을 폴백 가능하게 | D2「브로커 오류는 … 안에서 흡수한다」. `specs/exit-policy` **「목록을 얻지 못하면 제출하지 않는다」+「그러나 그 실패가 판정 자체를 중단시켜서는 안 된다」**. `tasks.md` 3.7 |
| 타임아웃·페이지 상한을 §0.4 예산으로 | D2「§0.4 예산」표 — **2초 · 3페이지**. `tasks.md` 3.E1·3.E2·3.E3 |
| 자기 방향 부재 확인을 `withPending`과 무관하게 | D2「자기 방향 미체결의 부재 확인」. spec 「자기 방향 미체결의 부재는 별도로 확인해야 한다」. `tasks.md` 3.D1~3.D3 |
| 취소 불가 주문의 영구 보류에 명시적 결정 | D2「취소할 수 없는 주문이 …— 결정」— **(b)는 파싱 문제라 없앤다, (a)는 연속 3회에 등급을 올리되 B7은 뒤집지 않는다**. `tasks.md` 3.B3·3.E4 |
| design의 B3 술어 반전 수정 | 1판 정정에서 이미 반영(§1.6 오류 1) |

**파싱 함수는 「무엇을 쓸지」가 아니라 「둘 다 못 쓴다」가 답이었다.**
`brokerstate.ParseOfficialOrder`는 `Side`·`Symbol`·`Market`·`Price`·`Currency`를 주지
않고(`officialOrderPayload`는 6필드짜리 의도적 부분 미러),
`official.Client.Orders()`는 첫 페이지만 남긴다. 새 파서 `execgw.ParseWorkingOrder`를
만든다 — 이것이 1라운드가 지적한 「엔진 API 표면은 새로 생긴다」의 구체다.

**§0.3 약속도 고쳤다.** 1판 tasks 7.2는 *"제출 시점이 늦어지지 않음을 보인다"*였고
R2에 대해 그것은 **보일 수 없었다**(차단 7). 새 7.2는 「상한이 2초임을 보인다」이며,
`precheckTimeout`의 선례가 같은 논거를 쓴다.

### 2.3 R3 — 진입점을 바꿨다

| §1.7 지시 | 반영 |
| --- | --- |
| 새 `SupervisedLoop`, `reconcile.Run`·`RecoverPending`·`Recovery` 생성 재사용 SHALL NOT | `design.md` D3「진입점 — `reconcile.Run`을 재사용하지 않는다」+ 표. spec 「세션 중 해소는 관측 해소 진입점만 불러야 한다(SHALL)」. `tasks.md` 4.4a·4.4b·4.5 |
| `Resolve` 1회 = 최소 45초 · 6회 조회 | D3「주기의 경계」표 — 구성값에서 **유도한 하한**임을 명시. `tasks.md` 4.6 |
| 세션 중 park의 계정 전역 차단 공시 | D3「공시 — park는 계정 전역 진입 차단을 건다」. spec 조항 + 시나리오. `tasks.md` 4.8 |
| 이득 주장 정정 | D3「이득이 무엇인지 정확히」표 — 산출은 park이고 이득은 **취소·매도가 풀리는 것**. `tasks.md` 4.9. proposal도 고쳤다 |

**`RecoverPending`이 왜 세션 중에 위험한지를 표로 열거했다** —
`RECORDED`를 *"found at startup with no dispatch recorded"*로 **종결**시키고
`DISPATCH_STARTED`에 *"process stopped after dispatch started"*를 쓴다.
둘 다 세션 중에는 **거짓이며 원장에 남는다.**

**주기 하한 하나는 새로 정했다** — `MaxDuration`(5분)보다 짧을 수 없다.
그보다 짧으면 앞 `Resolve`가 끝나기 전에 다음이 시작된다. 이것은 미측정값을 박은 것이
아니라 **이미 있는 구성값에서 나오는 제약**이다.

### 2.4 R4 — 철회를 design에도 반영했다

1판 정정에서 `review.md`·`proposal.md`·`tasks.md`는 고쳤으나 **`design.md` D4는
처방을 그대로 두고 있었다.** 2판에서 고쳤다 — 제목에 「철회했다」를 달고, 「무엇을
채우는가」를 **「왜 철회했는가」**로 바꿨다. **진단은 남긴다**: baseline이 비어 있어
부재가 증명되지 않는다는 사실은 참이고 D3의 이득 진술이 그것에 의존한다.

D5 다이어그램에서도 R4를 뺐다(`## D5. 넷의 상호작용` → `셋의 상호작용`).

### 2.5 spec delta — 정본 보존을 프로그램으로 확인했다

`openspec validate --strict` **통과**. 그리고 그것이 잡지 못하는 것을 따로 쟀다:

- 정본 요구 본문 1963자 — delta 안에 **문자열 그대로 존재**(True)
- 정본 시나리오 **6/6** 전부 존재, 누락 0
- delta 총 시나리오 **14개**(정본 6 + a094 8)

`tasks.md` **7.0**에 그 검사를 게이트 항목으로 못 박았다 — *"validate 통과는 정본
보존의 증거가 아니다"*.

### 2.6 아직 안 한 것 (침묵한 생략 아님)

- **2라운드 리뷰를 돌리지 않았다.** 위는 반영 기록이고 판정이 아니다
- **교차 모델**: 1라운드는 Claude 보이스 셋이었다(사용자 지시). a092에서 여섯 라운드
  연속 미충족이므로 2라운드에서 이것을 지켜야 한다
- **FLM·AST 재생성 안 함** — 새 함수 둘(`classifyRefusalCode`·`ParseWorkingOrder`)과
  `SupervisedLoop`은 아직 소스에 없다. 구현 후 tasks 7.5가 만든다.
  **지금 문서가 주장하는 분기는 전부 기존 9함수의 것이고, 그 산출물은 이미 있다**
- **`go test` 미실행** — 이번 판은 문서만 고쳤고 Go diff는 0이다
- **`make sdd-sync`·`make gate` 미실행** (`mutating: true` — 사람이 승인한다)

---

## 2라운드 (gstack plan-eng-review) — **FAIL**

### 2.0 교차 모델 — **미충족** (a094 2라운드)

Codex를 outside voice로 돌렸고 **사용량 한도로 출력 0바이트**였다
(`ERROR: You've hit your usage limit … try again at Aug 8th, 2026 12:36 PM`).
gstack의 폴백대로 Claude 서브에이전트를 돌렸다 — **fresh context이지 다른 모델이 아니다.**

**a092 여섯 라운드 + a094 1·2라운드 = 여덟 라운드 연속 미충족.** 아래 판정은 그 조건
아래에서 읽어야 한다.

### 2.1 차단 P0 — **a094는 475150을 녹이지 못한다. 잠근 것은 attempt가 아니라 무장된 발의다**

1판·2판 전체가 잘못된 잠금을 지목했다. **매니저 재확인 완료.**

**소스 사슬** (전부 HEAD 확인):

| 자리 | 원문 | 결과 |
| --- | --- | --- |
| `exitloop.go:1296-1300` | `case out.State == StateInDoubt \|\| StateUnresolvedInDoubt: return nil` | **`release`를 부르지 않는다.** `pending_action`이 무장된 채 남는다 |
| `ladder.go:439-443` | `if observed < baseline { out.Reason = ReasonStopBreached; if in.State.PendingAction == ActionLadderStop { out.Suppressed = SuppressedPending; return out, nil } }` | **`out.Proposal`을 채우지 않고 반환** = 빈 발의 |
| `exitloop.go:1082` | `orderable := snapshot.Orderable && !proposal.Zero()` | 빈 발의 → `orderable = false` |
| `exitloop.go:1117` | `if orderable && (snapshot.CancelPendingFirst \|\| isFullExit(proposal))` | **게이트가 열리지 않는다** |

RATCHET도 같다(`ratchet.go:423-424`).

**원장 실측 (2026-08-07)**:

| 종목 | `pending_action` | `pending_level` |
| --- | --- | --- |
| **475150** | **`STOP_LOSS_LADDER`** | 0 |
| **080220** | **`STOP_LOSS_LADDER`** | −1 |
| 272210 · 066570 · TSLA | `None` | — |

**따라서 475150·080220에서 `clearTheSymbol`도 `submit`도 도달하지 않는다.**
R1은 `submit`의 분류를 바꾸고 R2는 `clearTheSymbol`의 목록을 넓히는데, **둘 다 실행되지
않는 코드다.** R3은 `mutation_attempts`만 쓰고 `exit_states`는 건드리지 않는다.

`pending_action`을 NULL로 만드는 non-test 경로는 셋뿐이다:
`ResolveExitProposal`(`apply_hook.go:846-849`, 유일 호출자 `exitloop.go:1317` `release`) ·
`ApplyTx.ResolvePending`(체결 적용 안) · `resetExitStateForReadoptTx`(운영자 재편입).
**어느 것도 이 상태에서 불리지 않는다.**

**proposal의 「이미 얼어붙은 셋은 어떻게 되나」 7단계가 475150에 대해 거짓이고,
*"배포는 재시작을 포함하므로 현재의 동결 자체는 배포 시점에 풀린다"*도 거짓이다** —
`RecoverPending`·`reconcile.Run`은 attempt만 만진다.

**272210은 다르다** — `pending_action = None`이므로 매 주기 발의가 나고 a094가 실제로
돕는다. **두 가지 모양이 있었고 change는 하나만 다뤘다.**

### 2.2 차단 — R2의 자기 방향 부재 확인이 **영구 무보호 부류**를 새로 만든다

R1이 발의를 해제하면 `pending_action = NULL` → `CancelPendingFirst = false`(`ladder.go:447`)
→ 이후 모든 주기가 `withPending=false`다. 그 상태에서 spec은 *"자기 방향 미체결이 있으면
제출하지 않는다(SHALL NOT)"*를 요구한다. 그런데 R2는 **반대 방향만 취소한다**
(`exitloop.go:1343`의 `buy || withPending`).

**결과: 사용자가 앱에 넣어 둔 지정가 매도 하나가 그 종목의 모든 보호 청산을 영구
보류시킨다.** 오늘은 그것을 건너뛰고 제출한다. 이 계정의 사건 보고가 정확히
「앱에서 직접 넣은 주문」이었다.

그리고 그 검사가 막으려는 초과 매도는 **이미 막혀 있다** —
`armExitProposalTx`(`apply_hook.go:661-667`)가 발의가 하나 미결이면 두 번째를 거부한다.
D2 차단 8은 (a) 조회-취소 사이 체결 (b) 빈 가격만 열거했고 **가장 흔한 이 경로가 없다.**

### 2.3 차단 — spec delta가 자기와 모순한다 (1라운드 차단 7의 재발, 자리만 이동)

`specs/exit-policy/spec.md`의 시나리오 *"사용자가 아무것도 취소하지 않아도 손절이
나간다 → 그 종목이 무보호로 남지 않는다"*가, 같은 요구의 SHALL NOT 셋과 동시에 참일 수
없다 — 목록 미취득 → 미제출 · 자기 방향 매도 존재 → 미제출 · 자기 방향 부재 미확인 →
미제출.

**1라운드가 잡은 쌍은 고쳤고 새 쌍을 만들었다.**

### 2.4 차단 — a087 상호작용 (선후 관계가 틀렸다)

`LiveOrdersForSymbol`은 `coalesce(i.price,'') AS price`(`fills.go:1859`)이고
`floatOf`는 `ParseFloat("")`을 거절해 `clear=false`로 만든다(`exitloop.go:1673-1679`).

**D2의 결정은 *"빈 price는 0으로 읽되 `floatOf` 거절 경로는 저널분에만 남긴다"*였다.**
그런데 **a087이 보호 청산을 시장가로 바꾼다** — 즉 **저널분에 빈 가격 행을 만들기
시작한다. a094가 실패를 남겨 둔 바로 그쪽이다.**

tasks의 *"a087과 겹치지 않는다 — 가격 문제가 아니라 주문 충돌이다"*는 409만 보고 이것을
놓쳤다. **a087이 먼저 오면 `floatOf` 결정을 양쪽 모두에 대해 뒤집어야 한다.**

### 2.5 차단 — R3의 `Resolve`는 절차 하나가 아니다

`Resolve`는 attempt kind로 분기한다(`indoubt.go:274-279`) → `resolveCancel`/`resolveAmend`.
`resolveCancel`은 `r.Order`가 없으면 **즉시 park**하고(`amend_indoubt.go:52-54`)
park은 **계정 전역 게이트를 latch한다**(`indoubt.go:379-382`).

D3의 스케치와 tasks 4.5는 `PendingAttempts` + `Resolve`만 적는다. 그대로 만들면
**증거가 아니라 배선 누락으로 계정을 막는다.** 올바로 배선된 인스턴스는 이미 있다 —
`Context.Resolver`(`app/engine/gateway.go:233-240`) — **문서가 그것을 지목하지 않는다.**

그리고 D3의 *"관측만 하므로 side effect가 없다"*는 **주문**에 대해서만 참이다.
`Resolve`는 `Journal.Resume`·`ResolveConfirmed`·`ResolveFailed`·`ResolveUnresolved`를
**쓴다**(`indoubt.go:311, 343, 376`). tasks 4.2는 옳은 것을 시험하나 **산문이 과장했다.**

### 2.6 차단 — R2가 R3을 조직적으로 무력화한다

`absenceCorroborated`는 baseline보다 **먼저** `windowContaminated`를 본다
(`indoubt.go:459-461`). **R2의 취소는 같은 종목의 mutation이다.** 따라서 R2가 방금 건드린
종목에서 `Resolve`는 오염으로 park하고 D3·D4가 기대는 baseline 추론에 **도달하지 못한다.**
**두 문서 어디에도 R2↔R3 상호작용이 없다.**

### 2.7 §0.4 — 호출 **빈도** 항이 없다 (매니저 독립 확인)

`clearTheSymbol`은 모든 full exit 직전에 돌고 주기는 5초(`exitloop.go:97`)다.
proposal 자신의 실측: 272210이 `STOP_LOSS_LADDER → PROPOSAL_CANCELLED`를
**2h54m 동안 1931회, 중앙값 5.0초**로 반복했다. **R2 후 그것이 전부 브로커 OPEN 조회다 —
지속 ~11 req/min, 각각 손절 경로에 2초 타임아웃.**

D2의 예산표는 「1종목 × OPEN 1회 × ≤3페이지」에서 멈추고 tasks 7.3도 그대로다.
**R1이 그 라이브락을 닫지도 않는다** — 그것은 `submit` B9 `ReasonSymbolInFlight`
(`exitloop.go:1301`)에서 오고 409뿐 아니라 **모든** 미정산 attempt에 대해 발화한다.

### 2.8 차단 — 이미 있는 브로커 읽기를 다시 만든다 (매니저 발견)

`filldetect.Detector`가 **프로덕션에서 무조건 돌고**(`cmd/tossctl/engine.go:391-396`)
**3초마다 계정 전체 OPEN 목록을 완주한다**(`detect.go:364` `ScanOrders(... Status: statusOpen ...)`,
`detect.go:128` `PollInterval: 3 * time.Second`). 원천은 a094가 새로 넣으려는 것과
**같은 `execgw.OfficialOrders`**다(`engine.go:420`).

**즉 R2는 최대 3초 된 메모리 안의 데이터를 손절 경로에서 동기로 다시 가져온다.**
5종목이 발의 중이면 5초마다 5회 = ~1 req/s로, detector의 ~0.33 req/s 위에 **약 4배**다.

파서도 같다 — D2의 「기존 함수 둘 다 모자란다」 표가 **`filldetect.parseSnapshot`을
평가하지 않았다.** 그 `Snapshot`(`detect.go:194-211`)은 `OrderID`·`Symbol`·`Market`·
`Side`·`Quantity` **7필드 중 5개**를 이미 준다(주문 `Price`는 없다 — 그래서 확장이 답이지
세 번째 파서가 답이 아니다). `PENDING_CANCEL`도 `brokerstate.StateCancelPending`
(`derive.go:421`)으로 이미 모델링돼 있다.

### 2.9 배선 선례를 반대로 골랐고 stale 주석에 기댔다 (매니저 발견)

D2 표는 `Context.ExitObserver`가 nil-fill하고 `cmd/tossctl/engine.go`는 손대지 않는다고
정했다. **그런데 detector 파생 의존성의 기존 선례는 정확히 거기서 주입된다** —
`engine.go:349` `SLO: detectorPressure{detector: detector}`. `Context.ExitObserver`는
`opts.SLO`를 건드리지 않는다.

그리고 그 함수의 doc comment는 빌드와 반대를 말한다 — `exitwiring.go:313-317`
*"this build constructs no fill detector: there is no production polling loop to defer
to yet."* **빌드는 만들고 넘긴다. 주석이 stale이고 D2가 그 위에서 설계했다.**

### 2.10 좌표 오류 — 1라운드 §1.1의 「60/60 일치」는 **2판에 대해서는 성립하지 않는다**

§2.6이 적었듯 2판에서 FLM·AST를 재생성하지 않았다. 현재 본문에서:

| 문서 | 실물 |
| --- | --- |
| D1 ``AllReasonCodes()`(`failclosed.go:250`)` | `:254` (`:250`은 주석) |
| D1 ``classifyRefusalBody`(`failclosed.go:221-238`)` vs D0·proposal의 `223-238` | `223-238`. **design이 자기와 모순** |
| D1 재생 경계가 *"우연히"* 안전 | `classifyReplay` default 분기가 정책을 **의도적으로** 적는다 — *"A first dispatch would call several of these definitive refusals; a replay may not"*(`replay.go:517-520`). 계약 공백은 실재하나 **오늘 코드의 성격 규정이 틀렸다** |

나머지 대조 좌표는 전부 일치했다(§2.11 목록).

### 2.11 3라운드가 받는 것

**FAIL. 차단 8건.**

1. **잠금을 다시 지목한다.** `pending_action`이 무장된 채 남는 것이 475150·080220의
   동결이다. R1/R2/R3 중 그것을 푸는 것이 없다. **B8이 `release`를 부르지 않는 것을
   고칠지, 아니면 change의 주장을 「272210 모양만 고친다」로 축소할지 결정한다.**
   전자는 `submit` B8의 안전 논거(*"releasing here would let the next observation submit
   a second sell"*)를 정면으로 다뤄야 한다
2. **R2의 자기 방향 부재 확인을 철회하거나 근거를 바꾼다.** 초과 매도는
   `armExitProposalTx`가 이미 막고, 이 검사의 한계 효과는 **보호를 withhold하는 것뿐**이다
3. **spec의 새 모순 쌍을 푼다**(§2.3)
4. **a087 선후 관계를 고친다** — `floatOf` 결정이 저널분에도 걸린다(§2.4)
5. **R3은 `Context.Resolver`를 지목하고 `resolveCancel`의 `r.Order` 요구를 적는다.**
   「side effect 없음」을 「주문 side effect 없음」으로 정정한다
6. **R2↔R3 상호작용**(오염으로 인한 park)을 문서에 넣는다
7. **§0.4에 호출 빈도 항을 넣는다.** 그리고 **detector의 OPEN 스냅샷 재사용**을
   대안으로 평가한다(§2.8) — 채택하면 §0.3 지연이 2초에서 ~0이 된다
8. **좌표 3건 정정 + 2판 FLM·AST 재생성**

**분할 권고**: R3은 이 사건의 인과 경로에 없다(proposal 자신이 인정). R1 + 축소된 R2로
줄이고 R3은 자기 §0.4 계측을 가진 별도 change로 낸다.
