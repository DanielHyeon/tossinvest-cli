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

---

## 증거 재생성 (2026-09-27, 3라운드 전 — 판정 아님)

Manager 과제: 3판 이후 387+ 커밋이 쌓인 HEAD 에 맞춰 증거만 재생성한다. **3판 본문은 고치지 않았다**(리뷰 전 초안 개정 금지).

### 번들 refresh (3937e341)

- HEAD `ddd39a83` 에서 `check_analysis` 가 stale 로 찍은 `revision: current` 번들 7개를 `go run ./tools/logic-map` 로
  재생성했다. 옛 ast 는 모두 base `ec29dc72` 의 파일 sha 와 일치했다.
- 분기 대응(옛/새 AST 를 (kind, 소스 줄)로 difflib 정렬):

| 번들 | 줄 | 분기 | 대응 | 이웃 |
|---|---|---|---|---|
| `exitobserver.clearthesymbol` | 1334-1392 → 1440-1498 | 9 → 9 | 항등, 본문 바이트 동일 | — |
| `exitobserver.record` | 1077-1197 → 1177-1303 | 14 → 16 | B1..B14 → B3..B16, 새 B1·B2 | a111 `882a0b49` |
| `exitobserver.submit` | 1237-1312 → 1343-1418 | 11 → 11 | 항등, 본문 동일 | — |
| `gateway.checksymbolfree` | 799-834 | 9 → 9 | 항등, 본문 동일 | — |
| `armexitproposaltx` | 655-677 | 4 → 4 | 항등, 본문 동일 | — |
| `journal.resolveexitproposal` | 810-869 → 825-884 | 14 → 14 | 항등, 본문 동일 | — |
| `recovery.run` | 207-296 → 238-329 | 12 → 12 | 항등(본문은 바뀜) | a102 `1c76a580` |

- `record` 의 새 B1·B2 행은 a111 의 Branch Test Map(같은 소스 sha `522d5d81` 의 AST 로 번호를 매긴 표)이 인용한 기존
  시험 `TestA111LeaseIsRecheckedAtTheRecordOrRefreshBoundary`(`internal/app/engine/a111_flat_exit_observation_test.go`)를
  인용했다. 새 시험 저술 0.
- BTM 행의 `:줄` 을 새 AST 줄로 옮겼고 「진입 실측」 열은 base 에서 잰 값 그대로라고 적었다. FLM 에는 Refresh 절을 붙였다
  (본문 줄 번호는 base 기준 그대로). risk report 는 발견 줄만 바뀐 4개를 재생성했다.
- 신선한 번들 8개는 바이트 불변(파일 60개 sha256 전후 대조, 바뀐 것은 stale 7 번들의 파일뿐).
- refresh 뒤 stale·해시 불일치 0.

### base 재고정 (d2f5d3f1)

- 잰 순간: HEAD 3937e341, 모집단 = `git log --full-history -- openspec/changes/a094-…` 커밋 3.
  - a30eb35a(2026-08-09, 7-change 계획 묶음) go 6 — 전부 a096 의 파일(`journal/outbox.go`·`obs/notifier.go`·a096 시험 4),
    `changed_existing_functions` 5 = a096 함수(`EnqueueAlert`·`Acknowledge`·`Flush`·`deliver`·`notifyCritical`)
  - 5bb3b8f9(병합) go 0 · 3937e341(refresh) go 0
  - a094 는 구현 전이므로 **자기 Go 커밋 0**.
- 옛 base ec29dc72 에서 5단계 rc 1 · required 265 · 창에 착지 커밋 406 — 전부 다른 change 의 기존 함수.
- 재고정 ec29dc72 → 3937e341, `base-commit.txt` 만 커밋(d2f5d3f1). 판정은 **격리 detached 워크트리 @d2f5d3f1** 에서:
  `check_analysis` rc 0, required 0, "evidence complete or diff-proven exempt". 공유 워크트리는 병행 로트(a112·a066)의
  미커밋 Go 편집 때문에 required 19 로 오염돼 있어 쓰지 않았다.

### 3라운드 입력

- 정오표: `analysis/third-round-errata.md` — 인용 91건(맞음 57 · 이동 32 · 내용 변경 2), `record` 분기 재번호,
  이웃이 바꾼 것 셋, a094 R1 ↔ a089 R2 문장 대조(같은 `error.code` 를 읽고 동작 분기에서 정반대 — main spec 충돌).
- 검토 자료 목록: `analysis/third-round-review-materials.md` — 교차 모델 요구, 대상 문서, 2라운드 차단 8건의 3판 답
  위치, 사람 몫.
- 교차 모델: codex CLI 가 2026-09-27 재가동 확인됐다(세션 id 3건 — `analysis/third-round-review-materials.md` §A). 3라운드는 Manager 지시 뒤.

---

## 3라운드 (codex 교차 모델, task 0.5d) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

### 3.0 실행 기록

- **교차 모델 충족** — codex-cli 0.154.0, model `gpt-6-astra`, `codex exec -s read-only --ephemeral --skip-git-repo-check -C <트리>`,
  session **`01a0e2d0-a6e1-7322-a506-c5dbde92c8e9`**, 2026-09-27 21:21:59~21:30:55 KST, rc 0, tokens 288,334, 401 없음.
  자격 증명·`~/.codex` 는 읽지도 고치지도 않았다. a124 12회차 종료 뒤 Manager 신호로 시작(동시 실행 금지).
- 프롬프트 = `analysis/freeze-review/codex-r3-prompt.md` 원문(sha256 `04946d6f145b8562…`, 실행 사본과 일치).
  출력 = `analysis/freeze-review/codex-r3-output.md` (최종 메시지, stdout 과 끝 개행 1바이트 외 동일).
- **실행 트리 편차 (기록).** 공유 워크트리에는 병행 로트(a066·a112)의 미커밋 Go 가 있어 a124 방식을 따랐다 — 세션 스크래치에
  `git archive 3937e341`(a094 base; `git diff 3937e341 HEAD -- '*.go'` 0)을 풀고 **워킹트리**의 a094 디렉터리를 겹쳤다
  (`diff -r` 0). 워킹트리를 쓴 것은 세 번째 stale index.lock(21:17:16) 때문에 입력 두 파일을 커밋할 수 없었기 때문이다 —
  Manager 승인, 그 두 파일은 뒤에 바이트 동일로 커밋됐다(093f9f10). 트리 파일 수 17047 → 17047(codex 가 만든 파일 0),
  저장소 변경 0.

### 3.1 2라운드 차단 8건의 판정 (codex)

RESOLVED 3 — spec 모순 쌍 · a087/`floatOf` · R3 `Context.Resolver`(범위 분리로). PARTIAL 5 — 잠금 재지목/B8 · R2 자기 방향 ·
R2↔R3 · §0.4 스냅숏 · 좌표/FLM. 분할 권고 PARTIAL. canonical 「IN_DOUBT 해소」 본문과 시나리오 6은 **그대로 재현됨**(텍스트 보존
통과), 의미 일관성은 park 해제 때문에 실패.

### 3.2 새 발견 — 분류 (Teammate 가 핵심 증거를 재확인, 반영 없음)

| id | codex | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| F1 | P0 | **차단 — 설계** | park 해제가 이중 매도의 실제 방벽을 치운다 | `armExitProposalTx` 는 `pending_action` 이 차 있을 때만 거절(`apply_hook.go:666-668`) — 그것을 비우면 원 매도가 살아 있어도 재무장된다. 축소 방향은 unresolved 검사를 건너뛴다(`gateway.go:815-816`). **확인** |
| F2 | P0 | **차단 — 설계** | 소급 재분류가 저장 본문이 **원 발주 응답**인지 증명하지 않는다 | 접수(`MarkAcked`) 뒤 readback 실패도 IN_DOUBT 와 오류 detail 을 남긴다(`dispatch.go:181-195`). **확인** |
| F3 | P0 | **차단 — 설계** | R2 의 취소 확대에 귀속·포지션 소유 경계가 없다 | 미확인(Manager 결정 뒤 FLM 으로 확인할 몫) |
| F4 | P1 | 설계 보완 | 종결→해제 이양이 충돌 완결적이지 않고 기대 intent 대조가 없다 | 미확인 |
| F5 | P1 | 설계 보완 | 재사용할 OPEN 스냅숏의 공개·신선도 계약이 현재 detector 에 없다 | 미확인 |
| F6 | P1 | 문서·설계 | tasks 4.4 고정 대상이 사라졌고(정오표 §4 와 같은 사실) 마이그레이션 순서 미정 | 정오표 §2·§4 와 일치 |
| F7 | P1 | **교차 change** | a094 R1 ↔ a089 R2 규범 충돌 — codex 권고: a089 의 금지를 텔레메트리로 좁힘 | 정오표 §5 와 일치. a089 처분(불구현 아카이브 제안)이 사용자 큐에 있다 |
| F8 | P1 | 설계 보완 | PENDING_CANCEL·"확인된 취소" 미정의, 인수(ack)와 호가 이탈 구분 없음 | 미확인 |
| F9 | P2 | 기록 | AC1 은 실재하나 "프로세스 안 latch 뿐" 은 불완전 — 재시작 시 미전달 알림이 따로 진입을 막는다 | `restoreAlertEntryLatch`(`internal/app/engine/gateway.go:153-168`). **확인** — 검토 자료 §G 의 "재시작 뒤 되돌리는 경로가 없다" 는 운영 모드에 한해 참 |
| F10 | P2 | 문서 | 좌표는 갱신됐으나 추론(FLM·BTM 요약·산문)이 낡았다 | record FLM `:53`·BTM `:26` 은 refresh 가 본문을 고치지 않은 자리(정오표 방침) |

**판정: REJECT.** 차단 셋(F1·F2·F3)은 3판 설계의 핵심 규칙(park 해제 · 소급 재분류 · 외부 주문 취소)을 겨눈다. 반영 방향 —
park 해제 철회/조건화, 소급 대상을 원 발주 응답으로 한정, 외부 취소의 귀속 경계 — 과 a089 처분(F7)은 Manager·사용자 결정
사항이다. **반영하지 않았다.**

---

## 4판 (2026-09-27) — 3라운드 반영 + gstack 문서 리뷰 1회 · 판정 아님

- 4판 초안 **2fd09cce**: `design.md` D−2(이 절이 이긴다), spec delta 2, tasks §0.5d~0.5h·§3·§4·§4bis·6.1·7.2·7.3, proposal 4판 박스,
  materials §G F9 정정. 방향은 Manager 지시(F1 park 해제 철회 · F2 원 발주 응답 한정 · F3 R2 엔진 귀속 축소, 외부 취소는 사용자 결정 대기 ·
  F4/F5/F8 설계 추가 · F6 재핀 · F7 두 결말 · F9 · F10).
- 격리 detached 워크트리 @2fd09cce 에서 `check_analysis --change a094` rc 0(required 0).
- **gstack 문서 리뷰 1회**(code-reviewer 서브에이전트, 읽기 전용): **P0 1 · P1 10 · P2 8**. 좌표는 두 곳 외 전부 일치.
  Teammate 가 P0(`gateway.go:747` 빈 상태 반환 → `submit` `default:` 해제, `exitloop.go:1410-1416`)와 replay 컬럼(`execution_contract.go:265-266`,
  `journal/replay.go:123`·`:146-151`), 전이표(`lifecycle.go:40-46`)를 재확인했다.

| # | 등급 | 지적 | 처리 |
|---|---|---|---|
| 1 | P0 | 결과를 쓰지 못한 제출(`State==""`, attempt 기록됨)에서 `submit` 이 발의를 푼다 | D−2.5 세션 중 해제 조건 좁힘, spec SHALL NOT + 시나리오, RED 4.3c |
| 2 | P1 | `clearTheSymbol` 의 `withPending` 해제가 IN_DOUBT·park 익절 위에 손절을 얹는다 | 트레이드오프 → **Q4-4**(4판은 동작 불변) |
| 3 | P1 | 조건 3 은 재생을 가르지 못한다 | 조건 7(`replay_count`·`last_replay_at`), RED 4b.2d |
| 4 | P1 | 조건 4 가 엔진 산문을 매칭(spec 금지와 모순) | `official: API error <n>: ` 표식 1회 + JSON, RED 4b.2e |
| 5 | P1 | 전방 R1 에 code 두 자리 불일치 규칙 없음 | spec SHALL NOT, RED 2.5e |
| 6 | P1 | 사건 두 행이 지금도 IN_DOUBT 라는 전제가 7주 묵음 | D−2.2 전제 명기, task 0.5h(읽기 전용 재측정, 사람 승인) |
| 7 | P1 | 기동 단계 실패 의미 없음(Recover 실패 → 루프 0) | 행 단위 흡수 규칙(D−2.6-4, spec), 알림 여부 **Q4-5**, RED 4.3d |
| 8 | P1 | 새 트리거와 30초 타이머의 알림 key 충돌 | 다른 key · `delayAlerted` 무접촉(D−2.7, spec, RED 3.E4) |
| 9 | P1 | IN_DOUBT 취소를 제외하면 새 트리거가 세션 내내 꺼짐 | 제외를 RECORDED·DISPATCH_STARTED·ACKED 로 한정, 시나리오 추가 |
| 10 | P1 | 스냅숏 전제 잔재(tasks 7.2·7.3·안전 표) | 원장만 읽음으로 재서술, 3.E1·3.E3 는 3.X |
| 11 | P1 | "매 주기 거절" 반복의 제출률·park 전락 | D−2.4 「그 반복의 비용」, **Q4-6** |
| 12 | P2 | 따라잡기가 attempt 없는 무장·"마지막" 만 봄 | "모두" + attempt 0 개도 해제(D−2.5, spec, RED 4.3b) |
| 13 | P2 | 재분류 범위·예약 해제 부수 효과 미기재 | PLACE 전체, 예약 해제 명기, reason code task 4b.8, RED 4b.7 |
| 14 | P2 | "유일한 경로" 서술 두 곳 오류 | 경로 넷 · `ClassifyBrokerRefusal` 갈래 명기 |
| 15 | P2 | "연속 3회" 근거 없음 | **Q4-7** |
| 16 | P2 | 따라잡기 순서의 이유가 틀림(`ready` 아닌 Recover 반환) | `engineRecoverySequence` 클로저 안 `r.Run` 뒤(`runtime.go:289-295`) |
| 17 | P2 | tasks 2.9 좌표 낡음 | `:1410-1416` |
| 18 | P2 | a089 관계 자기모순 | "a089 처분에 의존" 으로 정정(tasks · D−2.8) |
| 19 | P2 | exit-policy delta 머리말 3판식 | 재서술 |
| — | 좌표 | `dispatch.go:328-333` · `apply_hook.go:839-864` | `:330-336` · `:825-884` |

남은 사람·Manager 결정: Q4-1(park 발의) · Q4-4(`withPending` 해제) · Q4-5 · Q4-6 · Q4-7 · 외부 주문 취소(사용자) · 0.5h 재측정 승인.
4라운드 codex 는 Manager 대기열 순서대로.

### Manager 판정 (2026-09-27) · 0.5h 재측정

| Q | 판정 | 반영 |
|---|---|---|
| Q4-1 | 운영자 도구 경로에서 **즉시 해제**, 해제 판정 함수는 하나(기동 따라잡기는 백스톱). 도구는 미래 작업 — 설계 문장만 | D−2.11 · spec(park 시나리오·SHALL) · task 4.3e |
| Q4-4 | **현행 유지(손절 발신)**. 트레이드오프를 안전 불변식 §4 로 명명, 브로커의 보유 초과 거절은 미검증 잔여 | D−2.11 |
| Q4-5 | 행별 실패는 **critical**, key 포지션 단위(흡수하되 침묵 안 함) | D−2.6-4 · spec SHALL·시나리오 · task 4.3d |
| Q4-6 | **새 간격 없음**. 「429 류 거절 → park」 를 이름 붙인 위험으로. 영수증: 429 는 전송 층 sentinel(`official.ErrRateLimited`)이나 분류가 `dispatch_outcome_unknown` 으로 합쳐 원장 reason 에서 구별 불가 → **구현 로트 정지 조건** | D−2.11 · task 4.3f |
| Q4-7 | **a124 관례**: `N = obs.DefaultCriticalAttempts`(`notifier.go:45` = 3), 시험은 상수 인용(a124 `design.md:84-86`, freeze 14회차 PASS) | D−2.7 · D−2.11 · task 3.E4 |
| 외부 주문 취소 | 사용자 큐(Manager 등재) | — |

**0.5h — 운영 원장 읽기 전용 재측정(Manager 조건부 승인).** `~/.config/tossctl/journal.db` 를 `mode=ro` + `PRAGMA query_only` 로 열었다.
짧은 SELECT 몇 개, 브로커 호출 0. 읽은 것은 두 attempt 의 상태·전이·id·reason 과 연결된 발의 상태뿐이다. 계좌 정보는 읽지도 기록하지도 않았다.

- `034e5b79…`(475150): PLACE · attempt_no 1 · broker 주문 번호 없음 · replay 0 · `DISPATCH_STARTED → IN_DOUBT`(00:17:46Z,
  `dispatch_outcome_unknown`, 공식 오류 표식 1회, code 있음) → **`UNRESOLVED_IN_DOUBT` 2026-08-08T01:10:40Z**(`in_doubt_unresolved`).
  발의 `STOP_LOSS_LADDER` level `0` 무장.
- `8f68e7c3…`(080220): 같은 모양, park **2026-08-08T01:09:53Z**. 발의 `STOP_LOSS_LADDER` level `-1` 무장.
- 결론: 재분류 조건 2~7 은 채우고 조건 1(현재 IN_DOUBT)은 못 채운다 → **이 사건의 해동 경로는 Q4-1 운영자 도구 경로**다(design D−2.2).

---

## 4라운드 (codex 교차 모델, task 0.5f) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

- 실행: codex-cli 0.154.0, gpt-6-astra, `codex exec -s read-only --ephemeral --skip-git-repo-check -C <트리>`,
  session **`01a0e308-7099-7283-b840-358c484b24fe`**, 2026-09-27 22:22:55~22:27:27 KST, rc 0, tokens 167,944, 401 없음.
  트리 = `git archive d3aa309a`(커밋된 코드 그대로 — 공유 워크트리의 이웃 미커밋 Go 없음), 파일 수 17060 → 17060.
  프롬프트 `analysis/freeze-review/codex-r4-prompt.md`(sha256 `cedbfeac1f08e786…`, 47a61244), 출력 `codex-r4-output.md`.
- 3라운드 발견 판정: RESOLVED 6(F2·F3·F4·F5·F6·F9) · PARTIAL 3(F1 — Q4-4 경로 하나 · F8 — 인수가 호가 이탈을 대신함 ·
  F10 — record FLM 「Branches」 표 옛 번호, task 0.5g 미완) · NOT RESOLVED 1(F7 — a089 처분 미실행).
  canonical 본문 재현 통과, AST 해시 15 일치(182 분기), R2 브로커 읽기 잔재 없음, Q4-6 reason 주장 정확.

| id | codex | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| N1 | P0 | **Manager 재결정 필요** | Q4-4 가 "살아 있을 수 있는 주문 위에서 발의를 비우지 않는다" 를 어긴다 | 좁혀서 **확인** — 평범한 IN_DOUBT 익절은 `checkSymbolFree` 가 같은 종목의 모든 mutation 을 막아(`gateway.go:804-813`) 손절이 못 나간다. 우회는 **park 된** 익절에서만(`:815-816`). D−2.11 Q4-4 서술이 두 경로를 섞었다 |
| N2 | P1 | 설계 — a092 의존 | 새 critical(청소 트리거·기동 행)이 notifier 뮤텍스 아래 동기 전달 → 다른 포지션 손절 지연 가능 | 미확인(구조상 a092 가 다루는 주제) |
| N3 | P1 | 설계 | 코드 불일치의 "모호 강제" 를 `(code,bool)` 분류기가 표현 못 함; 422 면 폴백이 확정 거절 | 미확인 |
| N4 | P1 | **설계 결함 — 확인** | ACK 커밋 뒤 settle 실패로 무장 유지된 발의는 재시작 때 ACKED 가 그대로 남고(`journal/recovery.go:81-83`·`:116-120`) `Recovery.Run` 이 건너뛰며(`reconcile/recovery.go:262-272`) 따라잡기도 제외 → 재시작마다 얼어 있다. D−2.5 의 "재시작 복구가 IN_DOUBT 로 만든다" 는 ACKED 에 대해 거짓 | **확인** |
| N5 | P1 | 문서 | fixture 4b.6·6.2 가 4판이 못 내는 결말을 요구 | 0.5h·D−2.4 와 일치 — 확인 |
| N6 | P1 | 교차 change | a089 처분이 freeze 선결 조건으로 남음 | 사용자 큐 |
| N7 | P2 | 문서 | `fmt.Errorf("…: %w", apiErr)` 은 표식 1회 + JSON 을 유지 → "감싸이면 거절 쪽" 서술 거짓 | 미확인(논리상 참) |
| N8 | P2 | 기록 | a124 상수 차용은 일관성이지 근거가 아니며 notifier 튜닝과 결합 | — |

- codex 권고: **소급 재분류를 미룬다** — 0.5h 뒤 사건 두 행에 이득이 없고, 전략 PLACE 행까지 바꾼다.
- **판정: REJECT. 반영하지 않았다.** N1(Q4-4 재결정) · N4 · N6 · 재분류 연기 여부는 Manager 결정 사항이다.

## 5판 (2026-09-27) — 4라운드 반영 · 판정 아님

Manager 방향(2026-09-27)대로 썼다. 정본 `design.md` D−3. **5라운드는 a089 처분(사용자 답) 뒤**(D−3.9, task 0.5j).

| id | 처분 | 반영 자리 |
|---|---|---|
| N1 (P0) | Q4-4 **번복** — 살아 있을 수 있는 attempt 의 발의는 비우지 않음, `clear=false`, park 원인은 포지션 key critical, 해동은 Q4-1·종결 증거. 좁힘(평범한 IN_DOUBT 는 `checkSymbolFree` 가 이미 막음)을 D−2.11 에 반영 | D−3.2 · D−2.11 Q4-4 줄 · exit-policy delta 요구 1 + 시나리오 1 · tasks 3.0a·3.N1~3.N1d · 안전표 §4 |
| N4 (P1) | D−2.5 정정 + 기동 ACKED 정산 설계(기존 전이 `ResolveConfirmed`·`MarkInDoubt`, 판정은 `confirmCreatedOrder` 와 한 곳). 원장 쓰기 실패는 `Recovery.Run` 의 기존 규약(`ErrRecoveryIncomplete`), 주문 읽기 실패는 모호 경로. 정지 조건: PLACE 해소가 기존 주문 번호를 못 쓰면 멈춤 | D−3.3 · order-execution 요구 1 + 시나리오 1 · tasks 4.0b·4.N4~4.N4d · 7.2·7.3·8.2 |
| 재분류 | 이연(선택 후속) | D−3.4 · delta 에서 요구 1·시나리오 4 삭제 · tasks §4bis 강등 · 4.4 ① 삭제 |
| N3 (P1) | 3상 분류기, 422 모순 code 반례를 계약 시험으로 | D−3.5 · order-execution 요구 문장 + 시나리오 1 · tasks 2.5e~2.5g·2.8 |
| N5 (P1) | fixture 를 5판 의미론으로 | tasks 6.1·6.1a·6.2 · 4b.6 이동 |
| N7 (P2) | "감싸이면 거절 쪽" 정정, 이연 절의 구조적 강화 요구 | D−3.6 · §4bis 머리말 |
| N2 (P1) | 이름 붙인 잔여, a092/a124 교차 인용 | D−3.7 · tasks 선후 관계 |
| N8 (P2) | 유지 + "관례 일관성이지 증거 아님" | D−3.8 |
| N6/F7 | freeze 미충족 전제 | D−3.9 · tasks 0.5j·선후 관계 |
| F10 잔여 | `record` FLM 「Branches」 표 16분기 재번호(BTM 은 앞서 재번호) | FLM · tasks 0.5g 현황 |

**작성 중 정정 1건**: 초안 D−3.3-5 는 ACKED 정산의 행 실패를 D−2.6-4(흡수·critical)로 보냈다. `Recovery.Run` 의 이웃(재생·해소)은 행 실패를
`ErrRecoveryIncomplete` 로 반환한다(`internal/reconcile/recovery.go:276-289`) — 복구 본문 안에 두 번째 실패 규약을 만들지 않도록 고쳤다.

## 5라운드 (codex 교차 모델, task 0.5j) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

- 실행: codex-cli 0.154.0, gpt-6-astra, `codex exec -s read-only --ephemeral --skip-git-repo-check -C <트리>`,
  session **`01a0e857-0cc8-7080-90ee-cd981616d0bf`**, 2026-09-28 23:06:53~23:12:20 KST, rc 0, tokens 170,300, 401 없음.
  트리 = `git archive 0c12844a`(5판 · D−2.8 보충 62277308 · a089 아카이브 64a1b2b3 포함, 커밋된 것만), 파일 수 17283 = `git ls-tree` 17283.
  인용 Go 29파일은 base `3937e341` 대비 `schema.go`(a066 v34 +3줄)만 다르다 — 프롬프트에 명시.
  프롬프트 `analysis/freeze-review/codex-r5-prompt.md`(9a04cb59), 출력 `codex-r5-output.md`.
- 4라운드 발견 판정: RESOLVED 5(N3 · N6 · N7 · N8 · 3라운드 F7) · PARTIAL 3(N1 · N4 · N5) · NOT RESOLVED 1(N2).
  canonical order-execution 32–68 = delta 13–49 바이트 일치, AST 해시 15 일치(182분기 — tasks 의 "180" 은 낡음).

| id | codex | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| R5-1 | P0 | **설계 결함 — 확인** | D−3.3 의 ACKED PLACE 폴백(IN_DOUBT → 해소)은 이미 아는 주문 번호를 안 쓴다 — **5판이 스스로 적은 정지 조건이 지금 성립한다** | 확인 — `matcher` 에 주문 번호 판별자가 없다(`internal/execgw/indoubt.go:638-650`, `targetOrderID` 는 CANCEL/AMEND 용), 단일 일치가 `res.BrokerOrderID = order.OrderID` 로 기록 번호를 덮는다(`:307`) |
| R5-2 | P1 | **설계 결함 — 확인** | park 원인 critical(D−3.2)은 무장된 발의가 **손절 자신**이면 도달 불가 — 평가가 청소 전에 억제한다. 6.1a 가 요구하는 알림이 사건 fixture 에서 안 난다 | 확인 — `EvaluateLadder` `ladder.go:441-443`(PendingAction==ActionLadderStop → `SuppressedPending`, Proposal 없음) → `record` `orderable=false`(`exitloop.go:1185`) → 청소 게이트(`:1223`) 미도달. `record` FLM 「Safety conclusion」 의 "그 게이트는 이미 참이었다" 는 **거짓** — 정정 대상 |
| R5-3 | P1 | 설계 — 사용자/Manager | 좁힌 양보의 해동 경로(운영자 도구)가 이 change 밖이라 실재하지 않는다 | 확인 — `OperatorResolve` 비시험 호출자 0 |
| R5-4 | P1 | 교차 change — Manager | 새 critical 의 동기 전달 지연(N2)을 명명만 했다 — a092 를 구현 의존으로 하거나 enqueue-only 를 요구하라 | 4라운드 N2 와 동일 사실. 처분(의존 승격 여부)은 Manager |
| R5-5 | P1 | 설계 | 6.2 의 무조건 단언("반복 PROPOSAL_CANCELLED 0")은 거짓 — 무장 안 된 포지션이 다른 intent 의 같은 종목 IN_DOUBT 에 막혀 `SymbolInFlight` → 해제 → 재무장을 반복하는 셋째 기전이 D−3.2 밖에 남는다 | 미확인(논리 경로 `exitloop.go:1407-1409` 인용). 6.2 는 이미 "기전이 두 갈래 밖이면 멈추고 보고" 를 적었으나 codex 는 단언 자체를 좁히라고 한다 |
| R5-6 | P1 | 문서 | §0.4 계수 과소 — ACKED 정산의 "행당 1회" 는 직접 호출일 뿐, 뒤따르는 재생·해소(PLACE 두 목록 조회 · CANCEL 반복 읽기)와 401 재시도가 빠졌다. 기동 시간·율 예산 필요 | 미확인 |
| R5-7 | P0 | **설계 결함 — 확인(알려진 F8 계열)** | 유지된 CONFIRMED 경로가 여전히 원 매도가 살아 있을 수 있는 채 발의를 푼다 — 취소 ACK 를 장부 이탈로 받는다(취소엔 readback 없음). 또 소유 필터가 CONFIRMED 행을 빼면 "모두 CONFIRMED → 청소 목록에 있다" 가 거짓 | 확인 — `roundTripFor` 는 PLACE 만(`internal/execgw/roundtrip.go:72-75`), 취소 ACK → CONFIRMED(`journal/dispatch.go:181-202`) → 청소가 해제(`exitloop.go:1485-1495`). 소유 필터 `fills.go:1866-1872` 는 미확인 |

- 부수 확인(codex): 게이트 좁힘(평범한 IN_DOUBT 는 `checkSymbolFree` 가 막음)은 **맞다**. N1 의 원장 검사는 브로커 읽기 없이 구현 가능. ACKED 전이는 전이표상 합법. 복구 실패 규약은 이웃과 일치하나 **모든 루프를 세우지 않는다**(`cmd/tossctl/engineready.go:70-75`) — 읽기 실패가 park 로 가면 손절은 이후 재시작마다 얼어 있다.
- **판정: REJECT. 반영하지 않았다.** P0 둘(R5-1 · R5-7)과 R5-2(설계가 약속한 알림이 사건에서 안 남)·R5-3(해동 경로 실재) 처분은 Manager 결정 사항이다.

## 6판 (2026-09-29) — 5라운드 반영 · 판정 아님

Manager 처분(2026-09-29, 「축소」)대로 썼다. 정본 `design.md` D−4. 6라운드는 codex 대기열(a125 → a090 → a095 r4 → a094 r6).

| id | 처분 | 반영 자리 |
|---|---|---|
| R5-1 (P0) | 5판 정지 조건 성립 인정 — 기동 ACKED 정산 철회, **알림만**(상태 변경·브로커 호출 0). 정산은 명명된 후속(선행: matcher 주문 번호 판별자). **후속 기록용 좌표: `internal/execgw/indoubt.go:307` `res.BrokerOrderID = order.OrderID`(단일 일치가 기록 번호를 덮는다), matcher 필드 `:638-650`** | D−4.2 · D−3.3 배너 · order-execution 요구 교체 + 시나리오 · tasks 4.N4·4.N4a, 4.N4b~d 철회 · 선후 관계 |
| R5-7 (P0) | 두 형태를 코드로 재서 **(B) 청소 게이트 확장** 선택 — (A) 는 취소 상태 판정이 확인 읽기 쪽에 없고(`roundtrip.go:86-123` 은 존재만) 모든 취소에 왕복을 더한다. (B) 는 미체결 목록의 기존 종결 증거(`fills.go:1877-1879` 종결 스냅숏)를 쓴다. 매도 한정, 매수 무변화, 재취소 0, 발의 해제는 intent 의 주문 번호로 종결 확인 | D−4.3 · exit-policy 요구 문단 + 시나리오 2 · tasks 3.R7~3.R7c · 7.2 · 안전표 |
| R5-2 (P1) | park 원인 판정을 청소에서 떼어 판정 경로로. `record` FLM 의 "게이트는 이미 참" 정정(거짓 — `ladder.go:441-443` 억제가 먼저) | D−4.4 · exit-policy 문단 + 시나리오 · tasks 3.R2·3.R2a · 6.1a · `record` FLM |
| R5-3 (P1) | 해동 명령 1급 요구(형태: mutating · 운영자·승인 참조·note · audit 선행 · stale 거절 · 같은 해제 판정 · 엔진 무정지 · 콘솔 없음). 「해제 경로 부재」 셋째 사례(a092 모드 · a066 RISK_OVERAGE 다음)로 명기 | D−4.5 · order-execution 요구 + 시나리오 2 · tasks 4.T·4.Ta · 4.3e |
| R5-4 (P1) | enqueue-only 요구(`Journal.EnqueueAlert`, 선례 `execgw/replay.go:551`). a092 의존 승격 안 함 | D−4.6 · exit-policy 문단 · tasks 3.R4 · 선후 관계 |
| R5-5 (P1) | **코드 추적으로 확인** — 무장 없음 + 다른 intent 미종결: 청소 `cleared=true` → `clearDelay`(`exitloop.go:1255-1256`) → 무장 → `SymbolInFlight` → `noteDelay` 재시작 → 해제(`:1407-1409`), 매 주기. 지연 경보 영구 미도달. **제안(Manager 확인)**: 청소가 같은 종목 미종결을 `checkSymbolFree` 와 같은 함수로 보고 치움 미완료 | D−4.7 · exit-policy 문단 + 시나리오 · tasks 3.R5·3.R5a · 6.2 |
| R5-6 (P1) | **전수 재계수 — 새 브로커 호출 0**(정산 철회로 5판의 과소 계수 대상이 사라졌다) | D−4.8 · tasks 7.2·7.3·8.2 |

**Manager 확인 항목 둘**:
- **Q6-1** — 기동 ACKED 중 **읽기 확인만**(기록 번호 바이트 일치 → CONFIRMED)은 판별자 없이도 안전하다(발주 직후 확인과 같은 판정). 처분 문언("정산 자체는 후속")을 따라 6판은 알림만 쓴다. 이 부분을 남길지.
- **Q6-2** — R5-5 수리(D−4.7)를 6판 범위로 확정할지(Teammate 제안으로 spec·tasks 에 넣었다).

### Manager 판정 (2026-09-29) — Q6-1 · Q6-2

- **Q6-1 승인 — 남긴다.** "기록된 번호로 읽어 바이트 일치 → CONFIRMED" 는 R5-1 이 요구한 번호 일치 해소 그 자체다(matcher 불요 · 목록 스캔 불요 ·
  덮어쓰기 불가). 조건 ① 정확히 그 부분집합만 — 바이트 일치 외 어떤 추론도 알림 경로로 ② §0.4 를 그 읽기 포함으로 재계수 ③ 읽기 실패는 실패가
  아니라 알림 잔존. → D−4.2 1~4 · D−4.8 · order-execution 요구·시나리오 2 · tasks 4.N4~4.N4d · 7.3 · 8.2. **작성 중 결정 하나**: 확정 전이의
  원장 쓰기 실패도 알림 잔존으로 둔다 — 5판의 "이웃과 같은 `ErrRecoveryIncomplete`" 를 뒤집는다(그 규약은 관측 루프를 하나도 시작시키지 않는다). D−4.2-2 에 근거.
- **Q6-2 확정 — 6판 범위.** R5-5 는 재현된 영구 경보 기아 루프이고 수리는 규칙 단일화 + 보수 방향. 「판정이 둘이면 반증이 죽는다」 의 예방
  형태로 D−4.7 에 인용.

## 6라운드 (codex 교차 모델, task 0.5l) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

- 실행: session **`01a0e900-0b80-7841-b617-8ba4a5022d20`**, 2026-09-29 02:11:28~02:15:46 KST, rc 0, tokens 178,574, 401 없음.
  트리 = `git archive 7cf80832`(17397 = ls-tree). 인용 Go 드리프트 3파일(additive·무관)은 프롬프트에 명시. 프롬프트 `codex-r6-prompt.md`(f7a891b7).
- 5라운드 판정: RESOLVED 2(R5-3 · R5-5) · PARTIAL 5(R5-1 · R5-2 · R5-4 · R5-6 · R5-7). **P0 없음.** canonical 바이트 일치.
- **D−4.2-2 규약 이탈(공격 1번)에 대한 codex 판정: 방어 가능** — 확정 전이는 attempt 갱신과 이력이 한 트랜잭션(`durability.go:553-557`·`:633-669`)이라
  반쯤 커밋된 상태가 없고, 남은 ACKED 는 `PendingAttempts` 로 같은 종목 mutation 을 계속 막으며(`gateway.go:800-816`) 따라잡기 해제 규칙을
  채우지 못한다. 이웃의 실패 규약은 그대로다. **빠진 것은 R6-3(그 알림 자체의 적재 실패)** 이지 비대칭 자체가 아니다.

| id | codex | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| R6-1 | P1 | **설계 결함 — 확인** | "같은 판정 재사용" ≠ "번호 + 종목 일치" — `confirmCreatedOrder` 는 응답에 종목이 **없으면** 통과시킨다. 그대로 내보내면 새 확정 조건(Q6-1)을 어긴다 | 확인 — `roundtrip.go:118` `if facts.Symbol != "" && plan.symbol != "" && …` (빈 종목은 불일치로 안 침) |
| R6-2 | P1 | **설계 결함 — 확인** | 관측자 래치 리셋은 outbox 행을 다시 무장하지 않는다 — `EnqueueAlert` 는 remindAfter 0 이라 전달·승인된 같은 key 행을 PENDING 으로 되돌리지 않는다. "재시작하면 다시 알린다"·"기동마다 보인다"·새 연속이 **조용히 재사용**된다. 포지션만의 key 면 다른 park attempt 끼리도 충돌 | 확인 — `outbox.go:142-146`(remindAfter 0) · `claimOwed` `:382-384`(`remindAfter <= 0` → owed 아님) |
| R6-3 | P1 | 설계 | 흡수(D−4.2-2)는 같은 원장이 그 critical 을 기록할 수 있다고 가정 — 확정 실패 + 적재 실패가 겹치면 a124 가 찾을 행이 없다. 직접 enqueue 는 notifier 의 생산자쪽 진입 래치(`notifier.go:262-280`)도 우회 | 미확인(논리). 방향: 적재 실패 시 진입을 outbox 와 독립으로 잠그고 지역 진단 |
| R6-4 | P1 | 설계 | 형태 B 는 "한 주기" 가 아니라 **무기한 증거 대기** — 체결 감지가 실패하면 종결 스냅숏이 영영 안 올 수 있고, CONFIRMED 취소는 재시도되지 않으며, 해동 명령은 park 만 다룬다. 경보는 정지를 드러낼 뿐 풀지 않는다 | 미확인(논리 — `filldetect/detect.go:364-372` 실패 시 수집 중단 인용). 방향: 시간 상한 주장 삭제, 증거 미도착의 사람 복구 경로와 수용 명시(시간 경과로 해제·제출 금지) |
| R6-5 | P1 | 문서 | "401 포함 최대 3 요청" 은 주문 GET 만 — 토큰 재발급 POST `/oauth2/token`(`token.go:60-78`·`:108-145`)이 빠졌다. `rows×3s` 는 추가 읽기 시간이지 기동 전체가 아니다 | 미확인 |
| R6-6 | P2 | 기록 | 되살린 30초 경보(D−4.7)는 여전히 동기 전송 — 그 경보가 뒤 포지션을 늦출 수 있다(enqueue-only 는 **새** critical 만) | 사실(D−4.6 이 명시적으로 제외) — 잔여로 기록 |
| R6-7 | P3 | 편집 | 6판을 거스르는 옛 task 문구(tasks 200-201 · 237-238 · 242, review 859·865) | 정리 대상 |

- **판정: REJECT. 반영하지 않았다.** Manager 결정 사항: R6-1(확정 조건을 공유 판정 쪽에서 강화할지 — 발주 직후 확인도 바뀜), R6-2(내구 알림 에피소드
  신원 — key 에 원인·attempt·에피소드를 넣을지, enqueue-only 재무장 계약을 만들지), R6-3(적재 실패 시 진입 잠금), R6-4(무기한 대기의 수용과 사람 복구).

## 7판 (2026-09-29) — 6라운드 반영 · 판정 아님

Manager 처분(2026-09-29)대로 썼다. 정본 `design.md` D−5. 7라운드는 codex 대기열.

| id | 처분 | 반영 자리 |
|---|---|---|
| R6-1 | **측정 먼저**: 계약 `Order` 스키마 `symbol` 필수 · 예시 3/3 비공백 · 저장소 실측 주문 상세 본문 0 건 · 시험 픽스처 주문 상세 모양 40 줄 중 종목 없음 13(빈 값 0). → 공백 0 → **공유 판정 하나를 강화**(응답 종목 비면 확인 실패, 발주 직후 확인도 같이). 판정 분리 안 함 | D−5.1 · order-execution 요구 문장 + 시나리오 · tasks 4.N4e |
| R6-2 | **에피소드 key**(outbox 재무장 계약 신설 금지): park 원인 = attempt, 기동 ACKED = attempt, 청소 연속 = 연속 시작 관측 시각, 기동 행 실패 = intent. 유한·사실 결속 명시. "기동마다 보인다" 정정 → attempt 당 1회. a090 과 교차 인용 | D−5.2 · exit-policy 문단 + 시나리오 · order-execution 요구 · tasks 4.N4f |
| R6-3 | 적재 실패 → `BlockUnlessClearedSince(ReasonAlertUndelivered)` + 내용 없는 로그, 루프 계속. 형태 인용: 기존 생산자 래치 `notifier.go:262-280` · a092 22판. **범위는 계정 단위로 썼다 — Q7-1** | D−5.3 · exit-policy 문단 + 시나리오 · tasks 4.N4g |
| R6-4 | 시간 상한 주장 삭제, 무기한 대기 + 시간 경과 해제 금지 명시, 사람 복구 = 해동 명령 가족의 후속 확장 후보 | D−5.4 · exit-policy 문단 + 시나리오 · tasks 3.R7d · 7.2 · 안전표 |
| R6-5 | 토큰 POST 포함 완전 계수: 행당 ≤5(주문 GET ≤3 + 토큰 POST ≤2) + 프로세스당 첫 토큰 ≤1(캐시 가정 명시), `rows×3s` 는 확정 읽기가 더하는 시간 | D−5.5 · tasks 7.3 · 8.2 |
| R6-6 | 잔여 수용 + a092(동기 발송자 이관 소유) 교차 인용 | D−5.6 · 안전표 |
| R6-7 | 옛 문구 정리(tasks 4.0 · 4.3e · 4.4 · 7.2 · 안전표) | tasks |

**Manager 확인 항목 하나 — Q7-1(R6-3 의 범위).** 처분은 「그 종목만 잠근다」 였으나 코드에 종목 단위 래치의 **해제 세대가 없고**(`BlockSymbol`
`internal/execgw/symbolgate.go:64` — 원칙 E 비교 불가) **푸는 경로도 없다**(`ClearSymbol` 생산 호출자는 대사 `internal/reconcile/mismatch.go:1482` 하나).
계정 단위는 기존 생산자 래치와 같은 사유·같은 해제(`notifier.go:874-876`)를 그대로 쓴다. 7판은 **계정 단위**로 썼다. 종목 단위가 필요하면 종목 래치의
세대·해제 경로를 새로 만드는 설계가 따라온다(코드 영수증 없음 — 비움).

### Manager 판정 (2026-09-29) — Q7-1

- **Q7-1 = 계정 단위 승인 — 정확한 범위.** 적재 실패 = 원장의 알림 쓰기 자체의 실패 → 고장 범위는 계정(저널). 종목 한정은 과소 차단. 종목 래치의 해제 세대
  부재·해제 경로 1곳은 "좁히려면 새 설계" 의 영수증으로 병기. 해제는 원칙 E·기존 생산자 래치와 같다(Acknowledge). → D−5.3 개정.

## 8판 (2026-09-29) — a092 22라운드 교차 3건 · 판정 아님

Manager 교차 통지(2026-09-29). 정본 `design.md` D−6.

| 건 | 처분 | 반영 자리 |
|---|---|---|
| 1. 해제 세대 읽기 시점 | 7판 D−5.3 의 "적재 **전** 세대 읽기" 는 engine-safety 정본(`spec.md:1468-1470` — 확정 순간 = 오류가 돌아온 순간) 위반 — 인정. 단일 입구(D−6.1)로 a094 가 세대를 직접 읽는 자리는 없어지고, 입구 밖 형태의 규칙으로 "오류 반환 직후" 를 남긴다 | D−6.2 · D−5.3 배너 · exit-policy 문장 · tasks 4.N4g |
| 2. noteDelay 계약 | 에피소드 = key(신원), 재알림 창 = 재전달(주기) — **직교**. 새 기록자는 창 0, noteDelay 는 a092 어댑터의 1시간. "재시작은 새 에피소드 아님" 은 key 규칙 | D−6.3 |
| 3. K6 단일 입구 | 새 기록자들은 알림기의 critical 기록 단일 입구(`RecordAlert`, 창 0)로 기록. D−5.3 의 직접 `BlockUnlessClearedSince` 는 철회(잠금은 입구의 생산자 래치). 구현 순서: a092 입구 착지 뒤. 먼저 구현해야 하면 입구 밖 형태 + 먼저-잠금 규칙(a092 확정 문언 대조) | D−6.1 · exit-policy 문장 · tasks 4.N4g |

### Manager 판정 (2026-09-29) — a092 입구 의존의 층

- **설계 freeze 는 a092 와 독립, 구현은 a092 `RecordAlert` 착지 뒤(하드 조건).** a090 codex 3라운드 R3-4 처분을 a094 D−6.1 에도 같게 적용했다 — 8판의 "입구 밖 + 먼저-잠금" 대안
  문안 삭제(비례 원칙: 가지 않을 경로의 수명주기를 설계하지 않는다). → D−6.1 · D−6.2 · exit-policy 문장 · tasks 4.N4g · 4.N4h · 선후 관계 · proposal.

## 7라운드 (codex 교차 모델, task 0.5n) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

- 실행: session **`01a0e9c0-95ae-7870-8857-29f97891f398`**, 2026-09-29 05:41:47~05:45:22 KST, rc 0, tokens 144,569, 401 없음. 트리 = `git archive 9fa0bb90`
  (8판 7002270f 포함, **8판 개정 904c90e0 이전**). 프롬프트 `codex-r7-prompt.md`(80fc50d1).
- 6라운드 판정: RESOLVED 4(R6-1 · R6-5 · R6-6 · R6-7) · PARTIAL 3(R6-2 · R6-3 · R6-4). **P0 없음.** 세대 읽기 시점(D−6.2) 정본 일치, canonical 바이트 일치.

| id | codex | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| R7-1 | P1 | 설계 | 청소 연속 실패의 에피소드 key(연속 시작 **관측 시각**)는 내구 신원이 아니다 — 재시작이 같은 연속에 새 key 를 만들고, 벽시계 반복이 새 연속을 옛 settled 행에 흡수시킨다 | 확인(논리) — a090 이 같은 문제로 연속 id(`opts.NewID`)로 바꾼 것(a090 3판 N4)과 같다. a094 에 그 형태를 옮기면 된다 |
| R7-2 | P1 | 설계 | "a094 먼저 구현" 대안이 freeze 불가 — a092 는 입구 밖 기록자에 별도 사유의 선행 잠금을 요구한다 | **904c90e0(8판 개정)에서 이미 해소** — 대안 삭제, 구현 하드 의존(이 라운드 트리는 그 이전) |
| R7-3 | P1 | **설계 — 안전 불변식 §4** | 형태 B(매도 취소 ACK ≠ 치움)는 체결 감지가 멈추면 손절을 **영구히** 보류할 수 있다 — 경보는 드러낼 뿐, 운영자 명령은 park 만 다룬다. "후속 후보" 로 미룬 사람 복구 경로는 불변식을 충족하지 못한다 | 확인 — 7판 D−5.4 가 스스로 적은 사실. 방향 둘: 종결 증거 복구 명령(감사·귀속)을 이 change 에 넣거나, 형태 B 를 다시 좁힌다(예: 무기한 대신 사람 에스컬레이션 · 분리) — Manager 결정 |
| R7-4 | P1 | 교차 change | a092 의 "exit 기록은 재무장해야 한다" 는 exit 관측 goroutine 단위로 읽히는데 a094 의 새 기록자(청소·park)도 그 goroutine 에서 돈다 — a092 D0.3h 표(창 0 양립)와 a092 델타 문언이 어긋난다 | a092 는 freeze 됨 — 양쪽 문언 정합은 Manager(a092 는 재개방 없음이면 a094 쪽에서 "a092 델타의 재무장 요구 대상 밖" 을 명시) |
| R7-5 | P2 | 증거 | 측정은 보수적 거절을 정당화할 뿐 생산 영향 0 을 재지 못한다(`minLength` 없음 → 빈 문자열 허용 가능), 인용한 WTS 픽스처는 확인 읽기를 우회 | "생산 빈도 미측정" 명시, 후보 픽스처 정정 |
| R7-6 | P2 | 기록 | 1시간 창이 `noteDelay` 를 시간마다 울리게 하지 않는다 — 생산자 래치가 같은 지연 동안 재기록을 막는다 | D−6.3 에 "재기록이 있을 때만" 명시 |
| R7-7 | P2 | 기록 | 계정 단위 잠금은 방어 가능한 보수 정책이나 "적재 오류 = 계정 범위 고장" 은 과장 — 검증·컨텍스트 오류는 국소적일 수 있다 | Q7-1 근거 문언을 "기록 안 된 critical 의 보수 처리" 로 고치고 과잉 차단 가능성 기록 — Manager 가 격상한 문언이라 확인 필요 |
| R7-8 | P3 | 편집 | 3.R4 는 직접 `EnqueueAlert`(4.N4g 와 모순), 4.7 은 "새 브로커 읽기 없음" | 정리 |

- **판정: REJECT(P1 4 · P2 3 · P3 1). 반영하지 않았다.** R7-2 는 8판 개정으로 이미 닫힘. Manager 결정: R7-3(형태 B 의 무기한 보류 — 복구 명령 편입 vs 형태 재설계), R7-4(a092 델타 문언과의 정합 방식), R7-7(Q7-1 근거 문언).

## 9판 (2026-09-29) — 7라운드 반영 · 판정 아님

Manager 처분(2026-09-29). 정본 `design.md` D−7.

| id | 처분 | 반영 자리 |
|---|---|---|
| R7-1 | a090 3판 N4 해법 이식 — 청소 연속 실패 key = 연속 시작 때 `opts.NewID` 로 만든 연속 id. 재시작은 새 에피소드(보수 중복, 명시) | D−7.1 · tasks 4.N4f |
| R7-2 | 8판 개정 `904c90e0` 으로 이미 닫힘(입구 하드 의존) | — |
| R7-3 | (a) 해동 명령 가족에 **종결 증거 해소 대상** 추가 — 운영자 종결 단언(주문 번호·승인 참조·증거 필드·audit 선행·전제 셋·stale 거절). 자동 경로의 형태 B 는 불변. **체결 합성 금지**(`RecordFill` 은 포지션을 움직인다). 저장은 새 additive 기록 — **Q9-1(Manager)** | D−7.2 · order-execution 요구 + 시나리오 · exit-policy 문장 · tasks 4.Tb · 4.Tc · 3.R7d |
| R7-4 | 의존 쪽 명시 — a092 재무장 요구의 대상은 a092 D0.3h 가 정의한 창 기반 exit 기록자, 에피소드 key · 창 0 기록자는 대상 밖(a092 델타 `:44` 자체의 "다른 기록자는 0 일 수 있다" · D0.3h 4 표 인용). a092 델타 문언 명확화는 a092 구현 로트 erratum 후보(Manager 등록) | D−7.3 |
| R7-5 | 생산 영향 미측정 명기 · `wts_isolation_test.go:96` 후보 정정(서비스 직접 호출로 확인 우회) · 종목 없는 픽스처는 음성 픽스처로 보존 | D−7.5 · D−5.1 |
| R7-6 | 재알림은 생산자 재기록 때만(`delayAlerted` 래치) | D−7.5 · D−6.3 |
| R7-7 | Q7-1 근거 정정 — "기록되지 않은 critical 의 보수 처리, 누락 범위 불명 → 계정 단위가 보수적", 과잉 차단 가능성 기록 | D−7.4 · D−5.3 |
| R7-8 | tasks 3.R4(단일 입구) · 4.7(기동 확정 읽기 한정) 정리 | tasks |

**Manager 확인 항목 — Q9-1**: 운영자 종결 단언을 저장할 **additive 스키마**(단언 테이블 하나). 오늘 사람이 주문 종결을 원장에 쓰는 경로가 없고(비시험 종결 스냅숏 기록자는 체결 감지뿐),
체결 스냅숏 합성은 포지션 투영을 움직여 금지다. High-risk change 에 스키마 추가가 들어간다 — 승인되면 8.1 에 additive·롤백 순서를 추가한다.

### Manager 판정 (2026-09-29) — Q9-1

- **Q9-1 승인 — additive 단언 테이블.** 합성 체결·합성 스냅숏 금지(투영을 움직이는 적용 경로)가 옳고, 운영자 증거의 정직한 표현은 자기 테이블뿐이다. 조건: ① additive-only,
  8.1 에 additive·롤백 순서(v33~v35 선례 형식) ② 배포 노트에 판 증가와 main 대비 경고 ③ 증거 필드 필수. → D−7.2-5 · tasks 4.Tc(체크) · 4.Td · 8.1a.
- R7-4 erratum 은 "필수 아님, 명확화 후보" 로 격하(a092 델타 `:44` 자체 문장으로 정합).

## 8라운드 (codex 좁은 확인, task 0.5q) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

- 실행: session **`01a0e9d8-0a5d-7110-9039-9b0705e91eb9`**, 2026-09-29 06:07:24~06:10:14 KST, rc 0, tokens 122,745, 401 없음. 트리 `git archive 0b5ea307`(17529). 프롬프트 485933e8.
- 7라운드 판정: RESOLVED 6(R7-2 · R7-4 · R7-5 · R7-6 · R7-7 · R7-8) · PARTIAL 2(R7-1 — delta 문장 미반영 · R7-3). **P0 없음.** audit 선행 · 체결 합성 금지 · additive-only 는 옳다고 확인.

| id | codex | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| R8-1 | P1 | 설계 | 운영자 종결 단언의 증거 필드가 **종결 상태를 검증하지 않는다** — 전제 셋이 서도 X 는 살아 있을 수 있다(그래서 형태 B 가 있다). 증거가 `OPEN`·취소 중·모름이어도 받으면 종결 권위가 된다 | 확인(논리). 방향: 받는 종결 증거의 정의(종결 상태 열거 — `brokerstate.IsTerminal` 과 같은 집합)와 그 밖 거절, X 에 결속된 관측 시각 |
| R8-2 | P1 | 설계 | 단언이 **주문 인카네이션에 결속되지 않는다** — 번호 X 만 저장하면 재사용된 번호의 뒤 주문과 맞을 수 있다. 미체결 목록의 종결 필터는 범위·소유 시간순을 요구한다(`fills.go:1877-1889`) | 확인. 방향: 엔진 주문/attempt·정규 범위(계좌·시장·거래일·방향)·그 CANCEL attempt 에 결속, 두 소비자가 정확히 그 인카네이션만 받음 |
| R8-3 | P1 | **설계 — 초과 매도 방향** | 종결이 참이어도 **남은 수량이 안전하지 않을 수 있다** — 원 매도가 취소 전 4주 체결됐는데 체결 감지가 놓쳤으면 원장은 10주를 보유로 보고 손절이 10주를 낸다(6주뿐). 체결 합성 금지는 옳지만 그 빈 증거를 대신할 수량 전제가 없다 | 확인 — 제출 수량은 원장 보유(`exitloop.go:1357-1365`), 브로커 하한은 대사 차단 때만(`exitwiring.go:192-200`). 방향: 단언 뒤 첫 제출 전 **권위 있는 체결 따라잡기 또는 검증된 수량 상한** 요구 |
| R8-4 | P1 | 배포 | 8.1a 의 "롤백 = 백업 복원 2단계" 는 **백업 뒤의 거래 이력**(attempt·체결·투영)을 잃는다 — 단언 행만 사라지는 게 아니다(`internal/journal/backup.go:22-35` 경고) | 확인. 방향: 단순 복원은 "백업 뒤 mutation 0" 이 검증된 창에만, 그 밖은 현 저널 보존 + 통제된 복구·대사 |
| R8-5 | P2 | 편집 | 연속 id 설계가 exit-policy delta 문장(관측 시각 신원)에 미반영 | 확인 — 내 누락. 정정 |

- **판정: REJECT(P1 4 · P2 1). 반영하지 않았다.** R8-1~R8-3 은 9판이 들인 운영자 종결 단언 자체의 안전 결함이다 — Manager 결정: 단언의 증거·결속·수량 전제를 이 change 에서 완성할지, 단언을 좁힐지.

## 10판 (2026-09-29) — 8라운드 반영: 운영자 종결 단언 철회 · 판정 아님

Manager 판정 (A)(2026-09-29). 정본 `design.md` D−8.

- **R8-3 정지 조건 성립 → 단언 철회.** 한 주문 범위 체결 따라잡기가 없고 의도적으로 금지돼 있다(`hints.go:282-286`); 전체 주기는 원자적(`detect.go:398-404`·`:425-432`);
  X 는 추적 주문이라 따라잡기가 성공하면 형태 B 가 스스로 풀리고 실패하면 여전히 보류 — 단언은 어느 경우에도 결과를 못 바꾼다; CLI 경로는 엔진 감지기에 닿지 못한다
  (`detect.go:278`). → D−8.1.
- **Q9-1 Manager 승인 무효화**(위 근거) — 스키마 무변경, tasks 4.Tb·4.Tc·4.Td·8.1a 철회.
- **R8-1 · R8-2 소멸**(대상 없음). **(B)·(C) 기각** 사유 한 줄씩 — D−8.1.
- **불변식 §4 논증**: 형태 B 의 보류는 위반이 아니라 불확실성 하의 집행 — 감지가 죽은 동안 어떤 해제도 초과 매도를 배제할 수 없다. 사람 경로 = 감지 정지 원인 수리
  → 정상 경로 종결·해제. → D−8.2 · exit-policy 문장.
- **감지 복구 절차**: `docs/operations.md` 절 + 보류 경보 본문 한 문장(등급·key·빈도 무변화). → D−8.3 · tasks 3.R7e · 3.R7f.
- **R8-4** 롤백 원칙 → 8.2. **R8-5** exit-policy delta 연속 id 문장 정정.

## 9라운드 (codex 좁은 확인 — 철회 절 + §4 논증, task 0.5s) — **REJECT** · 분류만, 반영은 Manager 결정 뒤

- 실행: session **`01a0e9ed-c575-7693-8c53-2771dc0e58a9`**, 2026-09-29 06:31:08~06:33:24 KST, rc 0, tokens 114,767, 401 없음. 트리 `git archive e25d9b36`. 프롬프트 164eb417.
- 8라운드 판정: R8-1 · R8-2 · R8-3 **MOOT**(단언 철회로), R8-4 · R8-5 RESOLVED. 철회 전파(스펙·tasks 취소선) 일관 확인. **P0 없음.**
- 철회 결정은 정당하나 **증명은 과장**이라는 판정 — "전체 주기 원자적" 은 거짓(수집은 전부/전무지만 **적용은 스냅숏마다 커밋**, `detect.go:316-321`), "성공한 주기는 X 를 결착한다" 는
  보장이 아님(읽기가 OPEN 을 돌려줄 수 있고, 거절된 스냅숏은 **주기 오류 없이** `outage.success` 로 간다, `detect.go:324-347`), "단언은 어느 경우에도 결과를 못 바꾼다" 도
  반례 하나(실제로 취소·체결 0 인데 브로커 기록이 OPEN 으로 읽히는 상태) — 단, 복원 권고가 아니라 문장 반박.

| id | codex | 분류 | 요지 | Teammate 재확인 |
|---|---|---|---|---|
| R9-1 | P1 | **설계 결함 — 확인** | 사람 경로 "감지 수리 → 한 주기 성공 → 종결 증거 → 해제" 가 완결되지 않는다 — **감지가 건강한데** 증거가 영영 비종결·거절일 수 있다. 예: 읽을 수 있는 `CLOSED` 인데 체결·취소 증거가 없으면 도출이 거절하고(`brokerstate/derive.go:567-573`) `RecordFill` 은 **nil 오류로** 거절을 적으며(`fills.go:384-409`) 감지기는 건강으로 리셋(`detect.go:324-347`) — 인증·네트워크 수리로 못 푼다 | 확인. 추가 사실: 그 거절은 게이트 사유 `ReasonBrokerStateUnknown` 으로 **진입을 막을 뿐**(`filldetect/detect.go:669-684` `blockSymbol`, `ledger.go:97-103`) 감지기에서 critical 을 내지 않는다. `derive.go` 는 이 상태를 "resolve by hand or by the IN_DOUBT procedure" 라 적지만 **손 해소 명령은 cmd 에 없다**(grep) |
| R9-2 | P2 | 문서 | "전체 주기 원자적" 거짓 — 수집 완결성과 주문별 트랜잭션 적용을 따로 적어야 | 확인 |
| R9-3 | P2 | 문서 | "보류 중 세 경보가 난다" 는 과장 — 지연·D−2.7 은 판정 가능한 관측이 있어야, filldetect 열화는 **주기 실패**가 있어야 난다. 건강한 감지 + 보류 상태는 셋 다 침묵할 수 있다 | 확인 — 옛 조건부 문구(`design.md:251-252`)가 더 정확했다 |

- **판정: REJECT(P1 1 · P2 2). 반영하지 않았다.** R9-1 은 "건강한 감지 · 비종결/거절 증거" 상태의 사람 경로 — Manager 결정: 이 상태를 **시끄럽게**(명명 critical + 보류 사실)
  하는 것까지를 이 change 로 하고 해소는 이름 붙인 잔여(UNKNOWN_BROKER_STATE 손 해소 부재 — 기존 공백)로 둘지, 해소 수단을 이 change 에 들일지.

## 11판 (2026-09-29) — 9라운드 반영 · 판정 아님

Manager 판정 (i)(2026-09-29). 정본 `design.md` D−9.

| id | 처분 | 반영 자리 |
|---|---|---|
| R9-1 | 「종결 증거 대기」 명명 critical — 판정 진입에서, 엔진 CANCEL CONFIRMED 뒤 청산 지연 한계(30초) 이상 종결 스냅숏 없음, 감지 건강 무관, 기존 `EventExitLiquidationDelayed`, key 에 취소 attempt(에피소드), a092 입구 기록. 판정 불가 포지션은 a090 미관측 경보가 덮음(합집합). **해소 수단은 이름 붙인 후속 후보**(UNKNOWN_BROKER_STATE 손 해소 — `derive.go:86-88` 이 가리키는 명령이 저장소에 없음, 실측) | D−9.3 · D−9.4 · exit-policy 문장 + 시나리오 · tasks 3.R9 · 3.R9a |
| R9-2 | 철회 증명의 과장 셋 정정(적용은 스냅숏별 커밋 · 결착 보장 아님 · 반례 1 — 그러나 가려낼 수단 없음 → 결정 유지) | D−9.1 · D−8.1 배너 |
| R9-3 | 경보는 조건부 — 조건표로 대체 | D−9.2 · D−8.2 배너 |
| (ii) | 기각 — 철회 논리의 역행 | D−9.4 |

## 10라운드 (codex 좁은 확인 — 경보 절 + 정정 문구, task 0.5u) — **PASS**

- 실행: session **`01a0e9f8-f673-7d91-8db4-11efea4fe980`**, 2026-09-29 06:43:21~06:45:24 KST, rc 0, tokens 98,872, 401 없음. 트리 `git archive bc53211f`(17537). 프롬프트 88e05526.
- 9라운드 판정: R9-1 RESOLVED(승인된 에스컬레이션 범위) · R9-2 RESOLVED · R9-3 RESOLVED. **P0·P1 없음.**
- 확인된 것: D−9.3 조건은 원장만으로 구현 가능(브로커 읽기·감지 건강 조회 불요) · 에피소드 key 는 유한·재시작 안정(attempt id 는 기본키) · 건강한 감지 + 거절된 CLOSED 에서 조건이
  참으로 남는다(판정 진입이 억제·조기 반환 앞이어야 함 — D−9.3 이 그 자리를 지정) · 반복 관측·재시작이 새 행을 못 만든다 · D−9.1 의 정정된 추론이 정확하고 철회를 계속 정당화한다
  (금지 근거는 "남은 수량 안전을 세울 수 없음" 이지 "참인 단언은 결과를 못 바꾼다" 가 아님) · D−9.4 의 손 해소 명령 부재 실측 일치(`reconcile-resolve` 는 수량 불일치 RECONCILE 전용).
- 기록된 한계: 30초는 **자격 임계**이지 전달 시한이 아니다 · 벽시계 역행이 자격을 무기한 미룰 수 있다(명명됨) · 중복 제거는 내구적이나 외부 전달의 정확히-한번은 아니다.

| id | codex | 분류 | 요지 | 처리 |
|---|---|---|---|---|
| R10-1 | P2 | 편집 | "합집합이 침묵을 닫는다" 가 a090 의 조건부 보장(지속 B2 · 재시작 반복 잔여)을 넘는다 | **freeze 전 반영** — D−9.3-4 · 3.R9a 를 조건부 상속으로 한정 |

- **판정: PASS.**

## freeze 선언 (2026-09-29) — Manager 확인 대기

- 리뷰 이력: 1라운드(적대 Eng) FAIL · 2라운드(gstack) FAIL · 3~9라운드(codex) REJECT · **10라운드(codex 좁은 확인) PASS**. 판본 1~11.
- 결정 기록: Q4-1~Q4-7 · Q6-1·Q6-2 · Q7-1 · Q9-1(승인 뒤 **무효화**, D−8.1) · (A) 단언 철회 · (i) 명명 critical · a089 아카이브(64a1b2b3)로 F7/N6 해소 · a092 교차(K6 단일 입구 · 세대 시점 ·
  재무장 대상 정합, erratum 은 명확화 후보).
- **freeze 게이트 줄**: tasks 0.1~0.5w 체크. 남은 선행(구현 전): **a092 `RecordAlert` 입구 착지(하드 — 4.N4h)** · base 재고정(구현 로트 첫 행위, 사람 절차) · 편집 전 FLM
  (4.0a · 4.0b · 3.0a).
- 이름 붙인 후속 후보: UNKNOWN_BROKER_STATE 손 해소(D−9.4) · ACKED 정산의 matcher 번호 판별자(D−4.2) · 3.X 엔진 밖 주문 취소(사용자 결정).


## 구현 로트 (2026-09-30~) — 착지 단위별 기록

### 0. base 재고정 승인 (사람 절차 조건 ②)

- 승인 참조: **Manager 상임 지시**(구현 로트의 첫 행위 = base 재고정, freeze 선언 「남은 선행」) + **2026-09-30 a094 구현 재개 지시**.
- 조건 ① 실측: 옛 base `3937e341..1ffe2295` 에서 a094 디렉터리를 만진 비병합 커밋 36 중 `.go` 편집 0. 옛 base 의 required 92 · 창의 착지 223 = 형제 몫.
- 조건 ③: `479fdfa3` — `base-commit.txt` 단독 커밋(옛→새 sha · 실측 · 선례를 메시지에).
- 선행 조건 확인: a092 `RecordAlert` 입구 착지(4.N4h 하드) — a092 아카이브 `75d138b5`. 입구는 `obs.Notifier.RecordCritical(ctx, e, remindAfter)`(`internal/obs/record_only.go`) → `Journal.RecordAlert`. a094 의 새 기록자는 이것을 창 0 으로 부른다.

### 0.1 frozen 좌표의 현재 HEAD 대조 (a092 착지로 움직인 것)

frozen 문서의 줄 번호는 7cf80832 · 0c12844a 기준이다. a092 가 exitloop · obs · journal 을 바꿨으므로 **구현은 좌표가 아니라 심볼로** 짚고, 낡은 좌표는 여기서 정정한다(문서 본문은 frozen 이라 고치지 않는다).

| frozen 좌표 | 현재(HEAD `3260f4eb`) | 비고 |
|---|---|---|
| `classify.go:46` B3 `ClassifyBrokerRefusal` | `:66`(R1 편집 뒤) | R1 이 앞에 switch(`:44`)를 끼움 |
| `exitloop.go:1223` 청소 게이트 · `:1255-1256` `clearDelay` | `record` 안 같은 문장, 편집 전 `:1223`·`:1250` 부근 | 심볼 `ExitObserver.record` |
| `exitloop.go:1407-1416` submit 갈래 | `ExitObserver.submit` `switch` (편집 전 `:1396-1424`) | |
| `exitloop.go:1485-1495` 청소 해제 | `ExitObserver.clearTheSymbol` 끝(편집 전 `:1500-1508`) | |
| `exitloop.go:1688` `noteDelay` key · `:1675-` | `ExitObserver.noteDelay`(편집 전 `:1680-1704`) key `type|positionID` 무변화 | |
| `exitloop.go:1710` 동기 notifier | a092 뒤 **기록 전용**(`obs.RecordOnly`, `exitwiring.go:333`) — D−5.6 의 「동기 전송 잔여」는 a092 가 닫음 | 잔여 소멸 |
| `apply_hook.go:825-884` `ResolveExitProposal` | 같은 자리(편집 전 `:825-884`) | |
| `notifier.go:262-280` 생산자 래치 | 기록 입구의 실패 래치는 `record_only.go` `recordCritical`(`Gate.Block(ReasonAlertUndelivered)` + 승격) | D−6.1 의 「입구의 몫」 실물 |

### 1. 착지 단위 ① — R1 (tasks §2)

**Pre-Edit Gate (2.0)**

- change/task: a094 · 2.0~2.11
- 대상 심볼: `execgw.classifyMutation`(기존 — 분기 추가) · `execgw.AllReasonCodes`(기존 — 목록 한 줄) · 신설 `execgw.classifyRefusalCode`(새 파일 `refusal_code.go`)
- CodeGraph/호출자: `classifyMutation` 의 비시험 호출자 1(`Gateway.submit` → `gateway.go:726`), `ClassifyBrokerRefusal` 비시험 호출자 1(`classifyMutation`), `classifyReplay` 는 둘 다 부르지 않음(grep 전수 2026-09-30)
- CodeGraphContext: not-applicable — 호출자 1 의 순수 함수, grep 전수로 닫힘
- 기존 동작 근거: `TestTransportOutcomeTable` · `TestBrokerBranchesMapToStableReasonCodes` · `TestStatusOfReadsWrappedSentinels`
- FLM/BTM: `analysis/function-logic/internal-execgw--classifymutation/`(편집 뒤 재생성 · 10 분기) · `internal-execgw--allreasoncodes/`(신설)
- upstream 상속 시험 영향: no — 두 자리 code 모순 본문의 422 만 확정 거절 → 모호로 바뀐다(보수 방향, spec 2.5f)
- 실패 시험 선행: yes — `analysis/implementation/r1-red.log`(편집 전 7 실패)
- 설정·DB·journal: 없음(스키마 무변경 · 토글 무도입)
- §0 검토: 통과 — 확정 거절은 목록 code 하나(`opposite-pending-order-exists`)로만 넓고, 409 전체를 확정으로 바꾸지 않음(`request-in-progress` 는 모호 유지 — 2.3)

**구현 요지.** 본문 JSON 의 최상위 `code` · `error.code` 만 읽고(부분문자열 아님), 대소문자 무시 전체 일치. 모순은 모호 강제로 뒤 분류를 건너뜀. 읽을 수 없는 code 값(문자열 · null 아님)은 판정 없음(다른 자리 값으로 확정하지 않음). 읽는 본문은 `official.APIError.Body` 뿐 — 공식 클라이언트는 401·403 을 본문 없는 sentinel 로 바꾸므로 그 둘은 이 분류기에 닿지 않는다(기존 동작).

**검증.** `go test ./internal/execgw/ -count=1` ok · `go vet` ok · 골든 diff = `+opposite_pending_order_exists` 한 줄(생성기). 1.11 소비자 조사: 새 code 는 원장 `reason_code`(자유 문자열) · 알림 detail 로만 흐르고, execgw 코드를 열거하는 소비자는 골든 시험 · a098 census 둘뿐(콘솔 필터 없음 — `internal/console` 의 `ReasonCodes` 는 soak 코드).

**병행 조정.** AllReasonCodes · 골든 · a098 census 는 a112 와 같은 자리였다 — Manager 판정으로 a112(`3260f4eb`) 선행, a094 후행.

### 2. 착지 단위 ② — 발의 해제 판정 · 청소 · 판정 진입 알림 (tasks §3 · §4 일부)

**Pre-Edit Gate (3.0 · 4.0 중 이 단위 몫)**

- 대상 심볼(기존): `journal.Journal.ResolveExitProposal`(호출 형태 — 기대 intent 인자) · `journal.Journal.LiveOrdersForSymbol`(종결 술어를 공유 상수로 — 동작 무변) ·
  `engine.ExitObserver.{judge, record, submit, release, clearTheSymbol}` · `engine.Context.ExitObserver`(Critical 배선) ·
  `execgw.Gateway.checkSymbolFree`(판정 추출) · `execgw.Gateway.confirmCreatedOrder`(판정 추출 + 종목 공백 강화).
  신설: `journal/exit_proposal_release.go`(분류기 · 해제 둘 · 무장 목록 · 취소 조회) · apply_hook.go 의 `clearExitProposalTx` · `armedExitProposalsQ`(가드 컬럼 이름은
  apply_hook.go 에만 — `TestGuardedExitColumnsAreWrittenOnlyByTheApplyHook`) · `engine/exit_held_proposal.go` · `execgw/a094_shared_judgements.go`.
- 호출자(grep 전수): `ResolveExitProposal` 비시험 호출자 1(`ExitObserver.release`) → 이제 0(해제는 판정 함수로), 시험 4. `ExitSubmitter` 구현 = `*execgw.Gateway`
  (생산) + 시험 가짜 4(메서드 추가). `checkSymbolFree` 호출자 1(`Gateway.submit`). `confirmCreatedOrder` 호출자 1(`roundTripFor`).
- 기존 동작 근거: `exitloop_test.go` 전체 · `exit_state_test.go` · `TestTransportOutcomeTable` · `roundtrip_test.go`.
- FLM/BTM: 편집 전 AST 는 `analysis/implementation/pre-edit/*.ast.json`(HEAD `766a8456` 트리, 편집 전에 뽑음). 편집 뒤 번들은 다음 커밋(이 단위의 증거 커밋).
- upstream 상속 시험 영향: **하나 — `TestABreachDisplacesAnOutstandingTakeProfit`**. 의도된 변경(D−4.3): 매도 취소 접수는 치움이 아니므로 손절은 취소한 주기가 아니라
  종결 체결 기록이 있는 다음 주기에 나간다. 시험에 주기 하나를 더하고 그 이유를 주석으로 남겼다(제출 순서 · 무장 intent 단언은 그대로).
- 실패 시험 선행: 이 단위의 새 API(분류기 · 해제 판정 · Critical 옵션 · UnsettledOnSymbol)는 편집 전 트리에 없어 새 시험은 편집 전 트리에서 **컴파일 RED** 다. 행동 RED 는
  구현 조각을 되돌린 변이로 잰다(변이 원장 — 로트 리뷰 전 일괄, Manager 조건).
- 설정·DB·journal 스키마: 없음(원장 쓰기는 기존 컬럼만 — 발의 해제 · exit_events 행).
- §0 검토: 통과 — 손절 즉시성의 양보는 (a) park 된 발의 위, (b) 매도 취소의 종결 증거 대기, (c) 같은 종목의 미종결 attempt 위에서만이며 셋 다 명명 critical 또는 기존 지연 경보와 함께다.
  (c) 는 오늘도 게이트웨이가 거절하던 제출이라 새 양보가 아니다.

**구현 결정(frozen 문서가 비워 둔 자리 — 범위 확장 아님)**

| 자리 | 결정 | 근거 |
|---|---|---|
| park 원인 critical 의 이벤트 타입 | `obs.EventOrderUnresolved`(critical), key `order.unresolved_in_doubt|<position>|attempt:<id>` | 새 타입을 만들지 않음(R9 선례). 정본 등급표 「UNRESOLVED_IN_DOUBT 발생 → critical」 |
| 연속 청소 실패 · 종결 증거 대기 | `obs.EventExitLiquidationDelayed`, key `…|<position>|streak:<연속 id>` · `…|<position>|cancel:<취소 attempt id>` | D−2.7 · D−7.1 · D−9.3 문언 그대로 |
| 새 critical 의 입구 | `ExitObserverOptions.Critical`(`CriticalRecorder`) = `*obs.Notifier.RecordCritical(ctx, e, 0)` — 조립부가 무조건 덮음 | D−6.1(a092 단일 입구, 창 0). 기존 `Alerts`(RecordOnly)는 창 1시간이라 쓰지 않음 |
| 같은 종목 미종결 판정 | `ExitSubmitter.UnsettledOnSymbol` → `Gateway.unsettledFor` 한 함수, `checkSymbolFree` 도 그것을 부름 | D−4.7 · Q6-2. 옛 `checkSymbolFree` 는 첫 일치에서 거절했고 새 판정은 끝까지 모은 뒤 거절 — 뒤 행의 intent 읽기 실패가 거절 대신 오류가 될 수 있다(둘 다 fail-closed, 주문 미발송) |
| 발의 해제 판정 | `journal.ReleaseUnacceptedExitProposal`(판정 하나 — submit · 기동 따라잡기 · 해동 명령) · 청소는 같은 분류기 위의 `ReleaseClearedExitProposal`(ORDERS_CLOSED 도 해제) | D−2.5 · D−3.2 · D−4.3-2. 판정 읽기와 해제 쓰기는 한 트랜잭션 |
| 종결 증거 술어 | `confirmedOrderTerminalEvidence` 상수 하나를 미체결 목록과 해제 판정이 공유 | 판정을 둘로 두지 않음. 소유 유일성 필터는 미체결 목록에만(해제 판정은 intent 의 주문 번호로 봄 — D−4.3-2) |
| 실측 발견 | 같은 범위의 소유 모호(한 주문 번호 · 두 intent)는 미체결 목록에서 **조용히 빠지지 않고 오류**다(`guardTrackedFillIdentity`) → 청소가 오류를 돌려주고 해제 · 제출 없음 | D−4.3 의 「목록에서 빠진 주문」 전제는 이 판본에서 오류 경로로 닫혀 있다 — 3.R7a 는 두 층(journal 판정 · 목록 오류) 모두로 고정 |
| 실측 발견 | 발의 자신의 주문이 종결 체결 기록을 받으면 체결 적용 훅이 같은 트랜잭션에서 발의를 `PROPOSAL_FILLED` 로 끝낸다(`apply_hook.go` ApplyExitFill) | 그래서 형태 B 의 정상 해동은 청소의 해제가 아니라 적용 훅이다. 청소의 ORDERS_CLOSED 해제는 종결 기록이 무장보다 먼저 온 경우의 백스톱 |
| 결과를 못 쓴 제출(4.3c) | `out.AttemptID != ""` 이고 상태가 비수용 종결이 아니면 무장 유지 + 사이클 오류로 드러냄(거절 알림 없음 — 거절이 아니다) | D−2.5 |
| intent 없는 무장 발의 | 청소가 풀지 않음(치움 미완료, 계수 대상) | attempt 를 찾을 수 없으면 살아 있을 수 있음으로 다룸(N1) — 잔여: 옛 판본이 intent 없이 무장한 행이 있으면 사람이 푼다 |
| 알림 기록 실패 | 결과 무변(로그만), 래치 안 함 → 다음 관측이 재시도. 진입 잠금은 입구(`recordCritical`)가 이미 세움 | D−6.1 · 3.R2a |
| 로그 | a094 경로의 경고 줄은 계좌 필드를 싣지 않음(`warnA094`) | 불변식 8 · MaskAccount 선례 |

**§0.3 인접 대가 — 이름 붙은 대가임(Manager 조건, 2026-09-30).** `TestABreachDisplacesAnOutstandingTakeProfit` 의 +1 사이클은 손절 제출이
늦어지는 대가이며, frozen design 이 그것을 이름 대고 적었다: `design.md:403`(D−4.3-4 「무장 익절(매도)이 걸린 채 손절이 성립하면 손절이 체결 감지
한 주기(≈3초, SLO 10초)만큼 늦는다」)과 그 개정 `design.md:311`(D−5.4-4 「종결 증거가 올 때까지 늦는다 — 정상이면 체결 감지 몇 주기, 멈추면 무기한」),
tasks 「안전 불변식 확인」 §4 행(「R5-7 은 무장 익절 매도 위의 손절을 종결 증거가 올 때까지 늦춘다」). 시험의 가짜 제출자는 취소와 **동시에** 종결
스냅숏을 기록하므로 실측 지연은 관측 주기 하나(5초)다 — 생산에서는 체결 감지가 종결을 기록하는 시점까지다. 대가로 지키는 것: 살아 있을 수 있는
매도 위에 두 번째 매도를 얹지 않는다(취소 접수를 치움으로 치던 종전 규칙은 부분 체결을 모른 채 초과 매도로 열리는 fail-open 이었다). **다각 리뷰의
명시 검토 대상**으로 표시한다.

### 3. 착지 단위 ③ — 기동: ACKED 발주 확정 · 따라잡기 (tasks 4.N4* · 4.3b · 4.3d · 4.4)

**Pre-Edit Gate (4.0 의 기동 몫)**: 대상 `reconcile.Recovery.Run`(ACKED 갈래에 `confirmAcked` 한 호출 — 확정되면 `continue`, 아니면 종전 `StillPending`) ·
`engine.Context.Recovery`(Alerts · CatchUp 배선) · `cmd/tossctl` 의 `engineRecoverySequence` **var 클로저**(r.Run 성공 뒤 `r.CatchUpExitProposals`).
신설: `reconcile/acked_boot.go` · `engine/boot_catchup.go`. 호출자: `Recovery.Run` 은 클로저 하나(생산), `Context.Recovery` 는 `engineRuntime` 하나.
`engineRuntime` 은 **편집하지 않았다**(a090 편집 표면 — 따라잡기는 `reconcile.Options.CatchUp` 으로 실었다). var 클로저는 logic-map 도구의 함수 단위가 아니다
(`go run ./tools/logic-map --func engineRecoverySequence` → "not found") — FLM 은 `Function Logic Map: not-applicable` 사유로 구조 시험
`TestA094TheBootCatchUpRunsAfterASuccessfulRecovery`(클로저 문장 순서 Run → 오류 반환 → CatchUp → return 을 AST 로 단언)가 대신한다. §0 검토: 기동
확정은 바이트 일치에서만 원장을 바꾸고, 모든 실패는 알림으로 흡수해 복구를 실패시키지 않는다(D−4.2-2) — 관측 루프가 뜨지 않는 경로를 새로 만들지 않는다.

**구현 결정.** 기동 ACKED 알림은 `obs.EventOrderInDoubt`(critical), key `order.in_doubt|acked:<attempt>` · 창 0. 따라잡기 행 실패는
`obs.EventExitLiquidationDelayed`, key `…|<position>|intent:<intent>` · 창 0. 따라잡기의 목록 읽기 실패는 아무것도 바꾸지 않았으므로 경고 로그만 남기고
복구는 계속한다(잔여: 그 기동에서는 따라잡기가 없다 — 다음 기동이 다시 본다). 번호 없는 PLACE ACKED 는 원장이 `MarkAcked("")` 를 받지 않아 픽스처로 만들 수
없다 — 코드 갈래(`strings.TrimSpace(rec.BrokerOrderID) == ""` → 알림)는 남기고 시험은 CANCEL ACKED 로 「읽기 0 · 알림 1」 을 잰다.

**세 시험 파일의 되돌림은 이동이다(삭제 아님, Manager 조건).** b74875e7 이 `exit_identity_concurrency_test.go`(`journalMutationSubmitter`) ·
`a111_flat_exit_observation_test.go`(`a111SubmitSpy`) · `exit_snapshot_isolation_test.go`(`signallingSubmitter`)에 붙였던 `UnsettledOnSymbol` 메서드 셋은
**단언이 없는 인터페이스 충족용 메서드**(`return nil, nil`)이며, 같은 몸체 그대로 `a094_exit_clear_test.go` 끝으로 옮겼다. 세 파일은 b74875e7 이전 바이트로
돌아갔다 — FLM 게이트가 편집된 파일의 이웃 함수(`newConcurrentObserver` 등)를 요구 집합으로 끌어들이기 때문이다. 약화된 단언 0.

| 옮긴 것 | 옛 자리 | 새 자리 |
|---|---|---|
| `(*journalMutationSubmitter).UnsettledOnSymbol` | `exit_identity_concurrency_test.go` 끝 | `a094_exit_clear_test.go` 끝 |
| `(*a111SubmitSpy).UnsettledOnSymbol` | `a111_flat_exit_observation_test.go` 끝 | 같음 |
| `(*signallingSubmitter).UnsettledOnSymbol` | `exit_snapshot_isolation_test.go` 끝 | 같음 |

### 4. 착지 단위 ④ — park 해동 명령 (tasks 4.T · 4.Ta · 4.3e)

**Pre-Edit Gate**: 기존 함수 편집은 `StartPositionPolicyCommandServer`(route 발견 3줄 — a066 블록 옆, capability 없는 빌드는 route 집합 불변) 하나다. 나머지는 새 파일:
`internal/attemptthaw/`(전송 계약) · `engine/attempt_thaw_{command,transport}.go` · `positionpolicyrpc/attempt_thaw_client.go` · `cmd/tossctl/engine_attempt_resolve.go`.
`cmd/tossctl/engine.go` 의 `AddCommand` 한 줄은 a090 의 engineRuntime 착지 뒤 Manager 창에서 넣는다(그 전까지 명령은 조립돼 있으나 트리에 붙지 않음).

**형태(a066/a092 완화 명령 가족 · D−4.5)**: `mutating: true` · 필수 `--attempt --target --operator --approval --note`(CONFIRMED 는 `--broker-order-id`) · 엔진 제어 endpoint 경유 ·
엔진 lock 없음 · 콘솔 route 0(시험이 `internal/console` 전수 grep). 순서: 엔진 안에서 attempt 가 park 인지 확인(아니면 stale) → **audit 줄**(`attempt_thaw`, 실패 시 아무것도
안 바꿈) → `OperatorResolve`(from=UNRESOLVED 원자 검사, 실패 시 보상 audit 줄 `not_committed`) → 비수용이면 **같은 명령 안에서** `ReleaseUnacceptedExitProposal`(판정 하나,
4.3e). 해제가 실패해도 해소는 유효하고 결과가 그것을 말하며 CLI 는 0 아닌 코드로 「다음 기동이 푼다」를 알린다 — 시험 4.Ta 가 실제로 따라잡기로 풀림을 잰다.
journal API 는 바꾸지 않았다(audit 는 전이 트랜잭션 **앞**에 둔다 — 전이 안의 콜백을 만들면 `Attempt.transition` 편집이 필요해 High-risk 표면이 넓어진다; 선례 a066 도
audit 줄 뒤 commit 실패 시 보상 줄).
