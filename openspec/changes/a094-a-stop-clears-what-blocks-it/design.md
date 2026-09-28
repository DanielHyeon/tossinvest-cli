# a094 · 설계

> 분기 인용은 전부 `analysis/function-logic/`의 AST 산출물에서 온다
> (함수 **15개** · 분기 **180개** — 3판이 6개 101분기를 더했다; 2026-09-27 refresh 후 HEAD 기준 **182개**, `record` 14→16).
>
> **7판(2026-09-29)**: 바로 아래 **D−5** 가 6라운드 반영이며 D−4 를 이긴다. **D−4**(6판)는 D−3 을, **D−3**(5판)은 D−2 를, **D−2**(4판)는 그 아래 3판 본문을 이긴다.

## D−5. 7판 — 6라운드(codex) 반영 (2026-09-29)

> **이 절이 D−4 를 이긴다.** 6라운드 원문: `review.md` 「6라운드」, `analysis/freeze-review/codex-r6-output.md`(P0 0 — D−4.2-2 의 규약 이탈은
> 「방어 가능」 판정). 방향은 Manager 처분(2026-09-29). 줄 번호는 HEAD `7cf80832` 기준.

### D−5.1 R6-1 — 확정 조건: 공유 판정 하나를 강화한다

**측정(Manager 지시 「측정 먼저」).** 브로커 주문 상세 응답에서 종목 필드가 비는 일이 있는가:

| 원천 | 세는 것 | 결과 |
|---|---|---|
| 계약 `docs/migration/openapi.latest.json` | `GET /api/v1/orders/{orderId}` 의 `result` 스키마 `Order` 의 필수 목록 · 예시 | `symbol` **필수**(`required` 에 있음) · 예시 3/3 이 비어 있지 않음(`005930` · `AAPL` · `AAPL`) |
| 저장소의 실측 기록(goldens · verify-live 기록) | 주문 상세 응답 본문 | **0 건** — 주문 상세 GET 을 바이트로 기록한 산출물이 없다(`analysis/goldens` 는 a112 의 시세만, `verify-execution-capability/measurements.md` 는 재생 결과 서술뿐) |
| 시험 픽스처(주문 상세 모양 — `orderId` + `status` 가 같은 줄) | 종목 없음 | 40 줄 중 13 — `internal/app/engine/precheck_test.go` 5 · `internal/brokerstate/derive_test.go` 4 · `internal/client/trading_test.go` 2 · `internal/app/engine/wts_isolation_test.go` 1(`:96`) · `internal/console/fake_broker_test.go` 1. **빈 문자열·null 종목은 0.** 발주 직후 확인(`roundtrip_test.go` `orderDetailJSON`)의 픽스처는 전부 종목이 있다 |

**판정: 브로커 증거상 공백 0(계약이 필수로 정한다; 실측 표본은 0 건이라 계약이 유일한 증거다).** 그래서 판정을 **둘로 가르지 않고**
공유 판정 하나를 강화한다: `confirmCreatedOrder` 의 종목 비교(`internal/execgw/roundtrip.go:118` — 오늘은 `facts.Symbol != "" && plan.symbol != ""`
일 때만 비교해 **응답 종목이 비면 통과**)를 **「응답 종목이 비었거나 계획 종목과 다르면 확인 실패」** 로 바꾼다. 번호는 종전대로 바이트 일치
(`orderId` 는 불투명 식별자 — `indoubt.go:601-604`), 종목은 `parseOrderFacts` 의 정규화(`ToUpper(TrimSpace)`, `indoubt.go:606`) 뒤 같음이다.

- **발주 직후 확인도 같이 엄격해진다 — 보수 방향이다.** 종목 없는 읽기는 오늘 ACKED→CONFIRMED 였다가 이제 IN_DOUBT(`ack_round_trip_unconfirmed`)
  → 해소 절차로 간다. 브로커가 계약대로면 동작 변화 0.
- **fail-closed 가 거부할 정상 입력**: 계약을 따르는 응답에서는 0. 시험 픽스처는 발주 직후 확인을 **실제로 타는** 것만 고친다 — 후보는
  `wts_isolation_test.go:96`(실제 게이트웨이 + HTTP). 나머지 열두 줄은 다른 판독기(사전 점검 · 상태 도출 · WTS · 콘솔)로 간다 — 구현 로트가 확인 읽기
  경로 도달 여부를 재서 고칠 목록을 확정한다(tasks 4.N4e).
- 반례 시험: 종목 필드 없음 · `null` · 빈 문자열 · 공백만 → 발주 직후 확인과 기동 확정 **둘 다** 거절(같은 함수이므로 한 변이가 둘을 깬다).

### D−5.2 R6-2 — 알림 key 에 에피소드 신원을 넣는다 (outbox 재무장 계약 신설 금지)

**코드가 주는 것.** `Journal.EnqueueAlert` 는 재알림 창 0 으로 기록한다(`internal/journal/outbox.go:142-146`). 창이 0 이면 전달·승인된 행은 다시
빚이 되지 않는다(`claimOwed` `:382-384`). 그래서 같은 key 를 다시 적재하면 **이미 끝난 옛 행이 조용히 재사용된다** — 6판의 "재시작하면 다시 알린다" ·
"기동마다 보인다" · 새 연속이 전부 이 모양이었다.

**결정(Manager).** outbox 의미론은 **바꾸지 않는다**(재무장 계약을 새로 만들지 않는다). 대신 **key 에 에피소드 신원을 넣어 새 에피소드 = 새 행**이 되게
한다. 에피소드 신원은 **유한하고 원장 사실에 결속**된 것만 쓴다(key 폭발 방지):

| 알림 | 에피소드 | key | 유한성·사실 결속 |
|---|---|---|---|
| park 원인(D−4.4) | park 된 **attempt id** | `type|position|attempt` | attempt 는 원장 행 — 포지션당 park 된 attempt 수만큼 |
| 기동 ACKED(D−4.2-2) | 남은 **attempt id** | `type|attempt` | 같음. **attempt 당 1회** — "기동마다 보인다" 는 **정정**: 한 번 적재되고, 전달 뒤에는 운영자 미전달·승인 목록에 남는다 |
| 청소 연속 실패(D−2.7) | 연속이 시작된 **관측 시각**(그 연속의 첫 실패 주기의 `observation.at`) | `type|position|<RFC3339>` | 연속은 "치움 성공 → 실패" 전이마다 하나 — 그 시각은 한 연속에 하나뿐 |
| 기동 행 실패(D−2.6-4) | 따라잡으려던 발의의 **intent id** | `type|position|intent` | 원장 행 |

같은 형태를 a090 이 쓴다 — a090 의 에피소드는 **미관측 연속의 시작 시각**이다(a090 design D4, 교차 인용). 재시작 뒤 같은 에피소드(같은 attempt,
같은 연속 시작 시각)면 같은 행이라 다시 보내지 않는다 — **재시작은 새 에피소드가 아니다.** 관측자 래치(재시작하면 잃는다)는 이제 적재 중복을 줄이는
최적화일 뿐 전송의 근거가 아니다.

### D−5.3 R6-3 — 적재가 실패하면 진입을 outbox 와 독립으로 잠근다

**코드가 주는 것.** 기존 생산자쪽 래치가 같은 상황을 이미 다룬다 — `Notifier.deliver` 는 outbox 기록 자체가 실패하면 "a critical %s alert could not be
recorded in the outbox" 로 **진입만** 잠근다(`internal/obs/notifier.go:262-280`, "Entries only. Exits are untouched: no alert failure may slow a stop.").
직접 `EnqueueAlert` 를 부르면 그 래치를 우회한다(6라운드). a124 원칙 E 의 적용 함수 `EntryGate.BlockUnlessClearedSince(reason, epoch, detail)`
(`internal/execgw/retry.go:571-584`)가 있고, a092 22판이 동기 래치 세 자리에 같은 형태를 썼다(a092 `design.md` 22판 「3. C2 · C27」).

**결정.** 이 change 의 모든 enqueue-only 알림(D−4.6)은 적재 전에 `ClearEpoch(ReasonAlertUndelivered)`(`retry.go:559`)를 읽고, `EnqueueAlert` 가
실패하면 `BlockUnlessClearedSince(ReasonAlertUndelivered, epoch, detail)` 로 진입을 잠그고, 알림 내용이 없는 구조화 로그 한 줄(`EventAlertUndelivered`,
사유만)을 쓴다. **관측 루프는 계속 돈다**(D−4.2-2 유지 — 루프를 세우지 않는다). 청산은 건드리지 않는다.

- **범위: 계정 단위 — Manager 승인(Q7-1, 2026-09-29), 편의가 아니라 정확한 범위다.** 적재 실패는 **원장의 알림 쓰기 자체가 실패**하는 상태라 고장의
  범위가 종목이 아니라 계정(저널)이다 — 종목 한정은 오히려 **과소 차단**이다. 해제는 원칙 E 와 기존 생산자 래치와 같은 경로(운영자 승인으로 미전달 0 이
  되면 `Clear`, `notifier.go:874-876`)를 유지한다. 종목 단위로 좁히려면 새 설계가 필요하다는 영수증 둘: 종목 래치 `BlockSymbol`
  (`internal/execgw/symbolgate.go:64`)에는 해제 세대가 없어 원칙 E 비교가 불가하고, 그것을 푸는 생산 호출자는 대사 경로 하나뿐이다
  (`internal/reconcile/mismatch.go:1482`).

### D−5.4 R6-4 — 형태 B 는 무기한 증거 대기다 (시간 상한 주장 삭제)

6판 D−4.3-4 의 "체결 감지 한 주기(≈3초, SLO 10초)" 는 **상한이 아니다** — 체결 감지가 실패하면 수집을 멈추고(`internal/filldetect/detect.go:364-372` ·
`:428-435`), SLO 는 백분위 목표일 뿐이다(`filldetect/slo.go:50-65`). 종결 스냅숏이 안 오면 **무기한** 기다린다. 정직한 계약:

1. **시간 경과로 해제·제출하지 않는다**(SHALL NOT). 증거가 안 오면 발의는 무장된 채이고 손절은 나가지 않는다.
2. 그동안 침묵하지 않는다 — 유효 시세가 계속 손절 조건을 주면 `clear=false` 가 `noteDelay` 에 닿아 30초 지연 경보가 나고, D−2.7 계수가 는다(6라운드가
   확인: "With continued valid breach observations, `clear=false` reaches `noteDelay`").
3. **사람 복구 경로는 아직 없다** — 해동 명령(D−4.5)은 park 만 다룬다(`resolution.go:144`). 「CONFIRMED 취소인데 종결 증거가 없는 매도」 를 사람이 확인해
   종결로 기록하는 명령은 **해동 명령 가족의 후속 확장 후보**로 명명한다(이 change 범위 밖).
4. §4 대가를 다시 쓴다: 무장 익절 매도 위의 손절은 **종결 증거가 올 때까지** 늦는다 — 정상이면 체결 감지 몇 주기, 체결 감지가 멈추면 무기한이며 그때는
   경보로 드러난다.

### D−5.5 R6-5 — §0.4 완전 계수(인증 요청 포함)

| 요청 | 조건 | 수 |
|---|---|---|
| 주문 상세 `GET /api/v1/orders/{id}` | ACKED PLACE 행마다 | 1 + 401 재시도 ≤2 = **≤3**(`internal/official/client.go:344-359`) |
| 토큰 교환 `POST /oauth2/token` | 401 재시도마다 재발급(`refresh`, `internal/official/token.go:108`)이 교환(`exchange` `:130`)하면; 첫 재발급은 다른 보유자의 토큰을 **채택**할 수 있어 교환이 없을 수 있다(`client.go:333-343` 주석) | 행당 **≤2** |
| 첫 토큰 획득 | 프로세스의 토큰 캐시가 비었을 때(`token.go:60-78`) — 기동의 다른 조회가 먼저 채우는 것이 보통이다(**가정**) | 프로세스당 **≤1** |

그래서 기동 추가 HTTP 요청은 행당 **≤5**(주문 3 + 토큰 2), 프로세스당 +1. 전부 `roundTripTimeout`(3초) 안이다 — `rows × 3s` 는 **이 확정 읽기가 더하는
시간의 상한**이지 기동 전체 시간이 아니다. 관측 루프·손절 경로의 새 요청은 0.

### D−5.6 R6-6 — 되살린 30초 경보의 동기 전송 (잔여 수용)

D−4.7 이 되살린 `noteDelay` 경보는 오늘의 동기 경로(`exitloop.go:1685-1710` → `n.mu`)로 나간다 — enqueue-only 는 이 change 가 **새로 더하는** critical
에만 적용된다(D−4.6). 그 경보가 뒤 포지션의 판정을 늦출 수 있음을 **잔여로 수용**한다. 동기 발송자의 이관은 **a092 소유**다(a092 22판 「범위 밖 동기
발송자」). D−4.7 의 "다른 손절 무변화" 주장은 이 잔여만큼 좁힌다.

### D−5.7 R6-7 — 옛 문구 정리

tasks 의 6판 이전 문구(기동 ACKED "상태 변경 없음" · 해동 명령 "범위 밖" · 옛 ACKED 정산 참조)와 review 6판 표의 "브로커 호출 0" 은 7판에서 정리하거나
대체 표시한다.

## D−4. 6판 — 5라운드(codex) 반영 (2026-09-29)

> **이 절이 D−3 을 이긴다.** 5라운드 원문: `review.md` 「5라운드」, `analysis/freeze-review/codex-r5-output.md`. 방향은 Manager 처분
> (2026-09-29, "축소"). 줄 번호는 HEAD `0c12844a` 기준(인용 Go 는 a094 base `3937e341` 과 `schema.go` 외 바이트 동일).

### D−4.1 한눈에

| 5판 | 6판 | 발견 |
|---|---|---|
| 기동이 ACKED 를 정산한다(PLACE 읽기 → 실패 시 IN_DOUBT → 목록 대조 해소) | **목록 대조 폴백을 뺀다.** 기동은 ACKED PLACE 를 기록 번호로 한 번 읽어 **바이트 일치면 CONFIRMED**(Q6-1 승인), 그 외는 상태 무변경 + 명명 critical. 나머지 정산은 **명명된 후속**(선행 조건: 해소기의 주문 번호 판별자) | R5-1 P0 |
| 취소 ACK(CONFIRMED)면 치움 완료 → 발의 해제 | **매도의 취소 ACK 는 치움의 증거가 아니다.** 그 매도가 원장의 미체결 목록에서 종결 증거로 빠진 뒤에만 치움 완료·발의 해제 | R5-7 P0 |
| park 원인 critical 은 청소 안에서만 | **청소 자격과 무관하게** 무장 발의의 attempt 가 park 면 관측마다 판정해 critical(포지션 key, 1회) — 무장 발의가 손절 자신이어도 | R5-2 P1 |
| 해동 도구는 미래 작업(시그니처 계약만) | **1급 요구로 승격** — `tossctl` mutating 명령, 운영자·승인 참조·note, commit 전 audit, 같은 해제 판정 | R5-3 P1 |
| 새 critical 의 동기 전달은 이름 붙인 잔여 | **enqueue-only 요구** — 이 change 의 새 critical 은 outbox 적재만 하고 관측 루프에서 전송하지 않는다 | R5-4 P1 |
| 6.2 "반복 PROPOSAL_CANCELLED 0" 무조건 | **셋째 기전 확인(코드 추적)** — 다른 intent 의 미종결 attempt 위에서 치움→무장→`SymbolInFlight`→해제가 매 주기 돈다. 청소가 같은 종목의 미종결 attempt 를 `checkSymbolFree` 와 **같은 함수**로 보고 **치움 미완료**로 본다(Q6-2 확정) | R5-5 P1 |
| §0.4: 기동 정산이 행당 읽기 1회 | **전수 재계수** — 새 호출은 기동의 ACKED PLACE 읽기 하나(행당 `OrderRaw` 1회, HTTP 401 재발급 재시도 최대 2회 포함), 관측 루프·손절 경로 0 | R5-6 P1 |

### D−4.2 R5-1 — 기동 ACKED 정산을 빼고 알리기만 한다

**5판이 스스로 적은 정지 조건(D−3.3-4)이 성립했다.** 해소기의 matcher 에는 주문 번호 판별자가 없다 — `matcher` 필드는 종목·방향·수량·
가격·시각 창·`targetOrderID`(CANCEL/AMEND 용)뿐이다(`internal/execgw/indoubt.go:638-650`). PLACE 해소에서 단일 일치가 나오면
`res.BrokerOrderID = order.OrderID` 로 **기록된 번호를 일치한 주문의 번호로 덮는다**(`indoubt.go:307`) — ACKED 에서 온 PLACE 를 이 경로에
보내면 이미 아는 주문 대신 지문이 같은 다른 주문을 접수 확정으로 기록할 수 있다.

**결정(Manager, Q6-1 승인 포함).** 목록 대조 해소로 보내는 폴백을 **뺀다.** 남기는 것은 **번호 일치 해소 그 자체**뿐이다 — 이미 기록된
번호로 읽으므로 matcher·목록 스캔이 필요 없고 덮어쓰기가 일어날 수 없다.

1. **ACKED PLACE**(기록된 `broker_order_id` 있음): 기동(`Recovery.Run` 의 ACKED 건너뛰기 자리, `internal/reconcile/recovery.go:262-272`)이
   그 번호로 주문을 **한 번** 읽는다(`Resolver.Order.OrderRaw`, 생산 배선 `internal/app/engine/gateway.go:276-283`). 응답의 주문 번호가 기록 번호와
   **바이트 일치**하고 종목이 일치하면 `ResolveConfirmed`(ACKED → CONFIRMED, `internal/journal/resolution.go:48-60`). **판정은 발주 직후 확인과
   한 곳**(`Gateway.confirmCreatedOrder` 의 번호·종목 비교, `internal/execgw/roundtrip.go:86-123`)을 게이트웨이 밖에서 부를 수 있게 내보내 쓴다
   (reconcile 은 비공개 `confirmCreatedOrder`·`parseOrderFacts` 를 직접 못 부른다 — 5라운드 부수 확인). 시한은 그 함수의 `roundTripTimeout` 그대로.
2. **그 외 전부 — 상태를 바꾸지 않고 알린다**: 번호 불일치 · 종목 불일치 · 읽기 실패(시한·전송·401 뒤 실패 포함) · 응답 해석 불가 ·
   `broker_order_id` 없음 · CANCEL/AMEND ACKED. 바이트 일치 외의 **어떤 추론도 하지 않는다**(IN_DOUBT 로 올리지도, 해소기로 보내지도 않는다).
   그 attempt 의 id·종류·종목·intent·사유를 명명한 critical 을 attempt 단위 key 로 적재한다(D−4.6 enqueue-only). **읽기 실패는 정산 실패가
   아니다** — (7판 D−5.3: 알림 적재가 실패하면 진입을 잠근다) `ErrRecoveryIncomplete` 를 내지 않고(그것은 관측 루프를 하나도 시작시키지 않는다, `cmd/tossctl/engineready.go:70-75`) 알림으로 남는다.
   `ResolveConfirmed` 의 **원장 쓰기 실패**도 같다 — 전이는 커밋되지 않아 ACKED 로 남고 알림으로 남는다. 이것은 5판 D−3.3-5 의 "이웃(재생·해소)과
   같은 `ErrRecoveryIncomplete`" 를 **뒤집는다**: 그 규약은 복구 전체를 실패시켜 **모든 포지션의 루프가 안 뜬다**(5라운드 부수 확인). ACKED 한 행이
   남는 대가가 그보다 작다. 이웃의 규약 자체는 바꾸지 않는다.
3. **나머지 정산은 명명된 후속**이다. 선행 조건: 해소기 matcher 의 주문 번호 판별자(기록된 `broker_order_id` 와 바이트 일치 요구) + 오답 단일
   일치·복수 일치 반례 시험. 덮어쓰기 좌표 `indoubt.go:307`, matcher 필드 `:638-650` 을 후속 기록에 남긴다. **matcher 편집은 a094 범위 밖.**
4. **대가(정직하게)**: 바이트 일치를 못 얻은 ACKED 의 발의는 재시작마다 얼어 있다(N4 의 사실 일부가 남는다) — 이제 기동마다 **보인다**.
   해동은 D−4.5 의 운영자 명령으로는 안 된다(`OperatorResolve` 는 park 에서만, `resolution.go:144`). 후속 change 가 닫는다.

### D−4.3 R5-7 — 매도의 취소 ACK 는 치움이 아니다

**코드가 주는 것.** 취소에는 확인 읽기가 없다 — `roundTripFor` 는 PLACE 만 확인한다(`internal/execgw/roundtrip.go:72-75`). 취소 ACK 는
CONFIRMED 로 정산되고(`journal/dispatch.go:181-202`), `clearTheSymbol` 은 `out.State == StateConfirmed` 면 그 주문을 치운 것으로 치고
끝에서 발의를 해제한다(`exitloop.go:1485-1495`). 또 미체결 목록 `LiveOrdersForSymbol` 은 소유 intent 가 유일한 주문만 내므로
(`journal/fills.go:1866-1872` `1 = COUNT(DISTINCT owner.intent_id)`) 소유가 모호한 CONFIRMED 주문은 **목록에 없어도 살아 있을 수 있다.**

**두 형태를 재서 고른다(Manager 지시).**

| 형태 | 기존 기계 | 대가 |
|---|---|---|
| (A) 취소 확인 읽기 — `roundtrip.go` 를 CANCEL 로 넓힌다 | `confirmCreatedOrder`(`:86-123`)는 **존재**(번호·종목)만 본다 — 취소됐는지(상태)를 읽는 판정이 없다. 상태 파싱은 해소기 쪽 `parseOrderFacts`(`indoubt.go:601-607`)에 있다 | 게이트웨이의 **모든 취소**에 브로커 왕복 1회 추가(§0.4) · 손절 경로 한가운데 동기 왕복(§0.3) · 새 판정 |
| **(B) 청소 게이트 확장 — 종결 증거는 미체결 목록이 이미 쓰는 것** | `LiveOrdersForSymbol` 이 주문을 빼는 유일한 조건이 **종결 체결 스냅숏**이다(`fills.go:1877-1879` `f.terminal = 1`). 종결은 `brokerstate.State.IsTerminal()` — 체결·취소·부분체결 후 취소·대체·거절(`internal/brokerstate/derive.go:94-100`). 체결 감지가 이것을 기록한다(`filldetect/ledger.go:61`, 주기 3초 `filldetect/detect.go:128`, SLO p95 10초 `slo.go:61`) | 새 브로커 호출 0. 손절은 **취소된 매도가 종결로 기록될 때까지**(체결 감지 한 주기 ≈ 3초, SLO 10초) 기다린다 |

**선택: (B).** 기존 증거(종결 스냅숏)를 그대로 쓰고 새 판정·새 호출이 없다. **규칙**:

1. 청소가 **매도** 주문을 취소했다면 그 주기의 치움은 **완료가 아니다**(`clear=false`, 보호 청산 미제출). 다음 주기에 그 주문이 미체결 목록에
   없으면(= 종결 스냅숏) 완료다. 목록에 남아 있고 그 주문을 대상으로 한 엔진 취소가 이미 CONFIRMED 면 **다시 취소하지 않는다**(원장의
   취소 attempt 로 판정 — D−2.7 의 PENDING_CANCEL 과 같은 원장 조회를 CONFIRMED 까지 넓힌다) — 종결 증거를 기다리는 중이다.
2. **발의 해제(`withPending`)는** 발의 intent 의 CONFIRMED 주문이 **모두 종결 스냅숏을 가질 때만** 한다 — 목록 부재가 아니라 **그 intent 의
   주문 번호로** 종결을 확인한다(소유 모호로 목록에서 빠진 주문이 "없음" 으로 읽히지 않게).
3. **매수**(진입) 주문의 취소는 종전대로 ACK 로 치운다 — 초과 매도 방향이 아니고(매수 체결은 E31 재조정 경로), 흔한 경로(진입 매수 +
   손절)의 즉시성을 바꾸지 않는다.
4. **§4 대가**: 무장 익절(매도)이 걸린 채 손절이 성립하면 손절이 체결 감지 한 주기(≈3초, SLO 10초)만큼 늦는다. 종결 증거가 오지 않으면
   기존 30초 지연 경보(`noteDelay`)와 D−2.7 트리거가 그대로 돈다 — 침묵하지 않는다. 대가로 지키는 것: **살아 있을 수 있는 매도 위에
   두 번째 매도를 얹지 않는다.**
5. 종결 증거를 기다리는 취소(엔진 취소 CONFIRMED · 목록 잔존)는 D−2.7 의 연속 실패 계수에서 **빼지 않는다** — 체결 감지가 SLO 안이면
   계수 한계(3주기) 전에 끝나고, 넘으면 이른 경보가 맞다. D−2.7 의 제외는 종전대로 기록·전송·인수 단계 취소뿐이다.

### D−4.4 R5-2 — park 원인 알림은 청소 자격과 무관하다

**코드가 주는 것.** 무장 발의가 손절 자신이면 평가가 먼저 억제한다 — `EvaluateLadder` 는 손절 조건에서 `PendingAction == ActionLadderStop`
이면 `Suppressed = SuppressedPending` 으로 **발의 없이** 돌아간다(`internal/exitpolicy/ladder.go:441-443`; RATCHET 도 같다, `ratchet.go:422-426`).
그러면 `record` 의 `orderable` 은 거짓이고(`exitloop.go:1185` `snapshot.Orderable && !proposal.Zero()`) 청소 게이트(`:1223`)에 닿지 않는다.
**5판 D−3.2 의 park 원인 critical 은 사건 두 행(무장 발의 = `STOP_LOSS_LADDER`)에서 한 번도 나지 않는다.**

**결정.** park 원인 판정을 청소에서 떼어 **판정 경로의 앞**에 둔다: 관측에서 판정에 닿은 포지션(`ObserveOnce` → `judge`)이 무장 발의를
가지면(`m.state.Pending()`, `journal/apply_hook.go:578`) 그 발의 intent 의 attempt 상태를 원장에서 읽고, `UNRESOLVED_IN_DOUBT` 가 있으면
그 attempt 를 명명한 critical 을 **포지션 key·attempt 당 1회**(관측자 래치 — 재시작하면 다시 알린다) 적재한다. 평가·억제·청소 순서는 바꾸지
않는다(이 판정은 읽기와 알림뿐). 브로커 호출 0.

**FLM 정정.** `record` FLM 「Safety conclusion」 의 "475150은 `pending_action`이 `STOP_LOSS_LADDER`였으므로 그 게이트는 **이미 참이었다**" 는
**거짓**이다 — 무장 발의가 손절이면 `CancelPendingFirst` 가 설정되는 갈래(`ladder.go:446-447`)에 가기 전에 `:441-443` 이 억제로 돌아가
`orderable` 이 거짓이다. 6판에서 그 문장을 고친다.

### D−4.5 R5-3 — park 해동은 운영자 명령(1급 요구)

**코드가 주는 것.** park 의 유일한 출구는 `Journal.OperatorResolve` 이고(`internal/journal/resolution.go:114-150` — 운영자 신원·note 필수,
CONFIRMED 면 브로커 주문 번호 필수), **비시험 호출자가 0** 이다(grep). 자동 해소는 종결 상태에서 즉시 돌아온다(`indoubt.go:253-257`).
이것은 a092 의 운영 모드 완화·a066 의 RISK_OVERAGE 해제에 이어 **「해제 경로 부재」 의 셋째 사례**다 — 자동 경로가 조이는 상태를
사람이 풀 문이 코드에 없다.

**결정 — 요구와 형태(구현은 구현 로트).** 사용자 승인 해제 원칙(2026-09-28: 자동 경로는 조이기만 · 해제는 운영자 + 승인 참조 +
commit 전 audit · 동시 조이기는 보수 쪽 승리 · 진입점은 journal API + `tossctl` mutating 명령 · 콘솔 버튼 없음)을 그대로 따른다.

1. `tossctl` 명령 하나(가칭 `engine attempt-resolve`), `mutating: true` — 대화형 에이전트는 자동 실행하지 않는다.
2. 입력: attempt id · 목표(`FAILED_CONFIRMED` 또는 `CONFIRMED`+브로커 주문 번호) · 운영자 신원 · **승인 참조** · note(무엇을 확인했는가).
3. **audit 줄을 원장 전이 전에** 기록한다 — audit 가 실패하면 아무것도 바꾸지 않는다. 선례: `engine reconcile-resolve`(`cmd/tossctl/
   engine_reconcile.go:79-105` — `--operator`·`--note` 필수, 원장에 운영자·note 기록).
4. 결속: 전이 시점에 attempt 가 여전히 `UNRESOLVED_IN_DOUBT` 가 아니면 거절(stale).
5. `OperatorResolve` 뒤 **그 자리에서** 해제 판정 함수(D−2.5 의 한 곳)를 부른다 — 비수용이면 발의 해제. 두 쓰기 사이 충돌은 기동 따라잡기가
   닫는다(D−2.5). 엔진을 세우지 않는다(엔진 정지 = 손절 없음).
6. 콘솔 버튼 없음.

### D−4.6 R5-4 — 새 critical 은 enqueue-only

이 change 가 더하는 critical(청소 트리거 D−2.7 · park 원인 D−4.4 · 기동 행 실패 D−2.6-4 · 기동 ACKED D−4.2)은 **관측 루프에서 전송하지
않는다.** outbox 에 적재만 하는 기존 경로 `Journal.EnqueueAlert` 를 쓴다(`internal/journal/outbox.go:131` — "Use it when the alert only has
to be *recorded*"; 선례 호출자 `internal/execgw/replay.go:551`). 전송은 a098 의 전달 실행자가, 계속 실패 시 진입 차단은 a124(아카이브)의
실행자 판정이 맡는다. 그래서 전달이 막혀도 **다른 포지션의 손절 판정이 늦어지지 않는다.** a092 를 구현 의존으로 올리지 않는다(Manager —
21판 리뷰 중 결합 금지). 기존 경보(`noteDelay` 등)의 전송 형태는 바꾸지 않는다(범위 밖, a092).

### D−4.7 R5-5 — 셋째 기전: 다른 intent 의 미종결 attempt 위의 반복

**코드 추적(확인).** 무장 발의가 **없는** 포지션에서 손절이 성립하고 같은 종목에 **다른 intent 의** 미종결 attempt(기록·전송·접수·모호)가
있으면:

1. `record` → 청소 게이트(`exitloop.go:1223`, 손절은 `isFullExit`) → `clearTheSymbol` — 미체결 목록은 CONFIRMED 만이라(`fills.go:1862`)
   비어 있다 → `cleared=true` → **`clearDelay`**(`:1255-1256`)가 지연 타이머를 지운다.
2. 무장 → `submit` → 게이트웨이 `checkSymbolFree` 가 같은 종목 미종결을 보고 `SymbolInFlight`(`execgw/gateway.go:804-813`) →
   `noteDelay` 가 타이머를 **지금부터** 다시 시작 → `release(ProposalCancelled)`(`exitloop.go:1407-1409`).
3. 다음 주기 1 로 돌아간다. **타이머는 매 주기 지워지므로 30초 경보에 닿지 않고**, 치움은 성공이라 D−2.7 계수도 늘지 않는다.
   `PROPOSAL_CANCELLED` 가 주기(5초)마다 쌓인다 — 272210 의 기록(1931건 · 중앙값 5.0초)과 모양이 같다(인과 확정은 재생 6.2).

D−3.2 는 **발의 자신의** intent 만 보므로 이것을 못 막는다.

**결정(Manager Q6-2 확정, 2026-09-29).** 청소는 **같은 종목에 미종결 attempt 가 있으면 치움 미완료**로 판정한다(`clear=false`) — 게이트웨이가 어차피
거절할 제출을 무장·해제하지 않는다. 판정은 **`checkSymbolFree` 와 같은 함수**로 한다(목록 `PendingAttempts` + 대상 판정 `attemptTargets`
를 게이트웨이의 한 메서드로 내보내 두 곳이 부른다 — 판정 둘 금지). 결과: 무장·해제 반복 0, `clearDelay` 가 안 돌아 **30초 경보가 난다**,
D−2.7 계수가 는다. 제출 경로 자체의 `SymbolInFlight` 처리(`:1407-1409`)는 경합 대비로 남긴다.

**같은 함수여야 하는 이유.** 오늘 "같은 종목에 미종결이 있으면 안 된다" 는 게이트웨이(`checkSymbolFree`)에만 있고 청소는 다른 목록
(`LiveOrdersForSymbol` — CONFIRMED 만)으로 치움을 판정한다. 두 판정이 갈린 틈이 바로 이 루프다. 청소에 **두 번째 판정을 새로 쓰면**
둘이 서로의 시험을 통과시켜 변이가 살아남는다 — 저장소 교훈 「판정이 둘이면 반증이 죽는다」(규칙 하나는 한 자리에, 한 함수로 합친 뒤엔
호출자의 배관을 변이할 것)의 예방 형태다. 그래서 판정을 게이트웨이의 한 메서드로 내보내고 두 곳이 부르며, 3.R5a 가 그것을 구조와
변이로 고정한다.

### D−4.8 R5-6 — §0.4 전수 재계수

| 경로 | 브로커 호출 |
|---|---|
| R1 3상 분류기(D−3.5) | 0 — 응답 본문 해석 |
| N1 발의 보존(D−3.2) · park 원인(D−4.4) · 셋째 기전(D−4.7) | 0 — 원장 읽기 |
| R5-7 종결 증거(D−4.3) | 0 — 체결 감지의 기존 기록을 읽는다 |
| 기동 ACKED PLACE(D−4.2) | **ACKED PLACE 행마다 `OrderRaw` 1회**(기동, `ready` 앞 — 관측 루프 밖). 클라이언트의 HTTP 401 토큰 재발급 재시도 최대 2회(`internal/official/client.go:344-360`)를 포함하면 행당 최대 3 요청. 시한 `roundTripTimeout`. 재시도 루프·재생·해소·목록 조회 0. CANCEL/AMEND ACKED 는 0 |
| 기동 따라잡기(D−2.5) | 0 — 원장 |
| 운영자 해동(D−4.5) | 0 — 운영자가 확인한 값을 받는다 |
| 알림(D−4.6) | 0 — outbox 적재 |

**이 change 가 더하는 브로커 호출은 기동의 ACKED PLACE 읽기 하나뿐이다**(행당 1회, 401 재발급 포함 최대 3 요청). 관측 루프·손절 경로에는
0 이다. 5판이 빠뜨렸던 재생·해소·목록 조회(5라운드 R5-6)는 폴백을 뺐으므로 0 이다. 기동 시간 상한 = ACKED PLACE 행 수 × `roundTripTimeout`
(401 재발급은 같은 시한 안) — 운영 원장의 ACKED 행 수를 배포 전에 센다(tasks 8.2). 기존 호출(청소가 내는 **취소** mutation)의 수는 바꾸지 않으며, D−4.3-1 의 "재취소 안 함" 은 그것을 줄인다.

### D−4.9 5라운드가 확인한 것(유지)

게이트 좁힘(평범한 IN_DOUBT 는 `checkSymbolFree` 가 막음) · N1 원장 판정의 구현 가능성 · N3 3상의 422 폴백 차단 · 재분류 이연 ·
canonical 바이트 일치. 복구 실패(`ErrRecoveryIncomplete`)가 관측 루프를 하나도 시작시키지 않는다는 지적(`cmd/tossctl/engineready.go:70-75`)에 대해,
6판의 기동 확정은 모든 실패를 알림으로 흡수하므로 새 복구 실패 원천이 없다(D−4.2-2).

## D−3. 5판 — 4라운드(codex) 반영 (2026-09-27)

> **이 절이 D−2 를 이긴다.** D−2 와 충돌하는 자리는 이 절을 따른다. 4라운드 원문: `review.md` 「4라운드」,
> `analysis/freeze-review/codex-r4-output.md`. 방향은 Manager 판정(2026-09-27). 근거 줄은 HEAD 기준 — 인용하는 파일은
> a094 base `3937e341` 이후 바뀌지 않았다. **5라운드는 a089 처분(사용자 답) 뒤에 돈다**(D−3.9).

### D−3.1 한눈에

| 4판 | 5판 | 발견 |
|---|---|---|
| Q4-4 현행 유지 — 청소의 `withPending` 해제가 park 된 익절 위에서 발의를 비워도 손절을 낸다 | **번복.** 발의의 주문이 살아 있을 수 있으면 비우지 않고, 손절은 나가지 않으며, 원인을 명명한 critical 을 낸다 | N1 |
| "재시작 복구가 그 attempt 를 IN_DOUBT 로 만든다" | **거짓(ACKED).** ACKED 는 재시작에서 그대로 남고 복구가 건너뛴다 — 기동 시 ACKED 정산 경로를 설계한다 | N4 |
| 소급 재분류(D−2.3) | **이연(선택 후속).** 사건 두 행은 이미 park — 이득 없음, 전략 PLACE 행 변이 위험만 남는다. 사건 경로는 Q4-1 도구로 일원화 | codex 권고·Manager |
| 코드 분류기 `(ReasonCode, bool)` | **3상**: 확정 거절 · 모호 강제 · 판정 없음 | N3 |
| fixture 4b.6·6.2 가 4판이 못 내는 결말 요구 | 4판·5판 의미론으로 고친다 | N5 |
| "오류가 한 겹 감싸이면 모양이 바뀌어 거절 쪽" | **거짓** — 이연된 절에 정정과 구조적 강화 요구로 남긴다 | N7 |
| 새 critical 들의 전달 비용 무언급 | 이름 붙인 잔여 — `n.mu` 동기 전달은 a092/a124 소관 | N2 |
| N = `obs.DefaultCriticalAttempts` | 유지 — "관례 일관성이지 증거가 아니다" 를 명기 | N8 |

### D−3.2 N1 — 살아 있을 수 있는 주문의 발의는 비우지 않는다 (Q4-4 번복)

> **6판 보강 — D−4.3(취소 ACK 는 치움이 아니다) · D−4.4(park 원인 알림은 청소와 무관) · D−4.7(다른 intent 의 미종결).**

**코드가 주는 것.**

- 청소는 `withPending && m.state.Pending()` 이면 끝에서 무장된 발의를 `ProposalCancelled` 로 푼다(`internal/app/engine/exitloop.go:1492-1495`).
  청소 목록은 `CONFIRMED` 주문뿐이다(`internal/journal/fills.go:1862` 부근 `WHERE a.state = ?` = `StateConfirmed`). 그 발의의 주문
  attempt 가 모호·park 면 목록에 없어 `clear` 는 참이고 해제가 일어난다.
- **평범한 `IN_DOUBT`(·`ACKED`·기록·전송 중) attempt 는 이미 막힌다** — `checkSymbolFree` 는 같은 종목에 미종결(`PendingAttempts`)
  attempt 가 있으면 방향과 무관하게 모든 mutation 을 `SymbolInFlight` 로 거절한다(`internal/execgw/gateway.go:804-813`). 매도가 unresolved
  검사를 건너뛰는 우회(`:815-816`)에 닿는 것은 **park(`UNRESOLVED_IN_DOUBT`, 종결 상태)** 뿐이다. 4판 D−2.11 Q4-4 는 두 경로를 섞었다.
- 발의 attempt 는 `exit_states.pending_intent_id` → `mutation_attempts.intent_id` 로 찾는다(`apply_hook.go:669-673`, `schema.go:216-218`).

**결정(Manager 번복).**

1. 청소가 무장된 발의를 해제하기 **전에** 그 발의 intent 의 attempt 를 본다. 하나라도 주문이 살아 있을 수 있는 상태
   (`RECORDED`·`DISPATCH_STARTED`·`ACKED`·`IN_DOUBT`·`UNRESOLVED_IN_DOUBT`)이면 **해제하지 않고** `clear=false` 로 끝낸다 — 손절은
   그 주기에 나가지 않는다(`armExitProposalTx` 가 두 번째 발의를 거절한다, `apply_hook.go:666-668`). 모든 attempt 가 `CONFIRMED` 면 그 주문은
   청소 목록에 있으므로 종전대로 취소 확정 뒤 해제하고, 비수용 종결뿐이면 D−2.5 의 판정 함수로 해제한다.
2. **침묵하지 않는다.** 그 원인이 **park 된** attempt 면 "무보호 + 차단" critical 을 낸다 — 이벤트는 그 attempt(id·종목·운영자 해소 필요)를
   원인으로 명명하고 key 는 포지션 단위다. 모호·전송 중이면 기존 지연 타이머와 새 트리거(D−2.7)가 그대로 동작한다(그 상태는 재시작 복구가
   종결시킨다).
3. **해동**은 Q4-1 의 운영자 도구(같은 해제 판정 함수) 또는 종결 증거(해소가 비수용·접수로 닫음)뿐이다.

**안전 불변식 §4 와의 관계(정직하게).** 이 규칙은 손절 즉시성을 **양보한다** — 다만 **사람이 에스컬레이션된 park 상태에서만, 그리고
원인을 명명한 critical 과 함께만**이다. 대가로 지키는 것은 넘길 수 없는 규칙 "주문이 살아 있을 수 있는 동안 발의를 비우지 않는다" 다
(F1 에서 철회한 park 해제의 다른 문을 닫는다). 모호·전송 중 상태에서는 오늘도 `checkSymbolFree` 가 손절을 막으므로 5판이 새로 양보하는 것은
없다. 브로커가 보유 초과 매도를 거절하는지는 여전히 **미검증**이며 이 규칙은 그것에 기대지 않는다.

### D−3.3 N4 — ACKED 로 남은 attempt 의 기동 정산

> **6판에서 결정 부분 철회 — D−4.2.** 아래 「결정 — 기동 ACKED 정산」 은 자기 정지 조건(4)이 성립해 빠졌다. 「4판 정정」 과 「코드가 주는 것」 은 사실로 유지한다.

**4판 정정.** D−2.5 의 "재시작 복구가 그 attempt 를 IN_DOUBT 로 만든다" 는 `DISPATCH_STARTED` 에만 참이다. **`ACKED` 는 그대로 남는다** —
`RecoverPending` 은 ACKED·IN_DOUBT 를 "left as they are, still blocking" 으로 두고(`internal/journal/recovery.go:81-83`·`:116-120`),
`Recovery.Run` 은 IN_DOUBT 가 아닌 것을 건너뛴다(`internal/reconcile/recovery.go:262-272` — "ACKED … is not ambiguous, it is unfinished").
`Resolver.Resolve` 도 IN_DOUBT 가 아니면 `ErrNotResolvable` 이다(`internal/execgw/indoubt.go:258-260`). 그래서 D−2.5·4.3c 가 무장을
유지한 ACKED 발의는 **재시작마다 얼어 있다** — 설계가 그것을 허용하면 안 된다(Manager).

**코드가 주는 것 — 기계는 이미 있다.**

- ACKED → CONFIRMED: `ResolveConfirmed` 는 `IN_DOUBT`·`ACKED` 에서 전이한다(`internal/journal/resolution.go:48-60`).
- ACKED → IN_DOUBT: `MarkInDoubt` 는 `DISPATCH_STARTED`·`ACKED` 에서 전이한다(`durability.go:510-516`; 전이표 `lifecycle.go:45`).
- 주문 번호로 읽어 확인: 발주 직후 확인은 `Gateway.confirmCreatedOrder` — `OrderRaw` → `parseOrderFacts` → 주문 번호 **바이트 일치** + 종목
  일치(`internal/execgw/roundtrip.go:86-123`). 같은 읽기 도구를 `Resolver` 도 가진다(`Resolver.Order OrderReader`, `indoubt.go:196-198` —
  주석상 "Required for CANCEL/AMEND resolution; PLACE resolution does not use it"; `resolveAmend` 가 `r.Order.OrderRaw` 를 쓴다,
  `amend_indoubt.go:374`). 생산 배선은 그것을 채운다(`internal/app/engine/gateway.go:276-283` `Order: orders`).
- `Recovery.Run` 은 IN_DOUBT 를 재생 → 해소로 보낸다(`reconcile/recovery.go` 2단계).

**결정 — 기동 ACKED 정산.** `Recovery.Run` 의 미종결 순회에서 ACKED 를 건너뛰는 자리(`reconcile/recovery.go:262-272`)를 바꾼다.

1. **PLACE** ACKED: `broker_order_id` 로 `Resolver.Order.OrderRaw` 를 읽어 `confirmCreatedOrder` 와 **같은 판정**(번호 바이트 일치 · 종목
   일치)을 한다. 확인되면 `ResolveConfirmed`(ACKED → CONFIRMED). 확인되지 않으면(읽기 실패 포함) `MarkInDoubt`(ACKED → IN_DOUBT,
   `ack_round_trip_unconfirmed`) 한 뒤 **같은 순회에서** 종전 IN_DOUBT 경로(재생 → 해소)로 넘긴다. 판정 로직은 `confirmCreatedOrder` 와
   **한 곳**에 둔다(둘로 베끼지 않는다).
2. **CANCEL·AMEND** ACKED: 읽기 확인 규칙이 없다(`roundTripFor` 가 PLACE 만, `roundtrip.go:72-75`). `MarkInDoubt` 로 IN_DOUBT 에 올려
   기존 `resolveCancel`·`resolveAmend` 절차(`indoubt.go` 갈래)에 맡긴다.
3. 그 뒤 D−2.5 의 기동 따라잡기가 종결 결과(비수용이면 해제, 접수면 체결 경로)를 처리한다. 발의는 **종결 증거가 생긴 뒤에만** 풀린다 —
   ACKED 를 풀기 위해 발의를 비우지 않는다.
4. **정지 조건.** `Resolver` 의 IN_DOUBT 해소가 ACKED 에서 온 PLACE(주문 번호가 이미 있음)를 목록 대조로 처리할 수 없다는 사실이 편집 전
   FLM 에서 드러나면(예: matcher 가 번호를 쓰지 않아 다른 주문과 섞임) 멈추고 보고한다.
5. 이 정산은 `Recovery.Run` 의 순서(재시작 규칙 → 재생·해소) 안에서 돈다. **주문 읽기 실패는 정산 실패가 아니다** — 확인되지 않음이므로 1 의 모호 경로로 간다. 정산의 **원장 쓰기 실패**는 같은 순회의 이웃(재생·해소 실패)과 같은 규약을 따른다 — `ErrRecoveryIncomplete` 반환(`reconcile/recovery.go:276-289`). D−2.6-4 의 흡수·critical 규칙은 `Recovery.Run` **밖**의 기동 단계(따라잡기)에만 적용된다 — 복구 본문 안에 두 번째 실패 규약을 만들지 않는다.

이것은 재시작 규칙("ACKED 는 그대로")을 **증거 없이** 바꾸는 것이 아니다 — 확인 읽기라는 증거를 먼저 얻고, 얻지 못하면 모호로 올려 해소
절차가 증거를 찾게 한다. 둘 다 이미 있는 전이다.

### D−3.4 소급 재분류 — 이연(선택 후속)

0.5h 가 보였듯 사건 두 행은 이미 park 이고(D−2.2) 재분류 조건 1 을 못 채운다 — **이득이 없다.** 남는 것은 전략 PLACE 행까지 바꾸는 High-risk
원장 변이뿐이다(YAGNI). 5판은 D−2.3 을 **구현 범위에서 뺀다**: spec delta 의 재분류 요구·시나리오를 지우고, tasks §4bis 를 「선택 후속」 으로
강등한다. D−2.3 본문은 후속 change 가 다시 열 때의 설계 기록으로 남기되 D−3.6(N7 정정)을 함께 읽는다. **사건 경로는 Q4-1 운영자 도구로
일원화한다.** D−2.6 의 기동 순서 ①(재분류)은 사라지고 ②(따라잡기)와 D−3.3(ACKED 정산)이 남는다.

### D−3.5 N3 — 코드 분류기는 3상

**코드가 주는 것.** `classifyMutation` 은 `policyRefusal` → `ClassifyBrokerRefusal`(B3 `internal/execgw/classify.go:46`) → `statusOf` →
`journal.ClassifyHTTPMutation` 순이다(`classify.go:21-79`). 상태 코드 분기에서 422 는 확정 거절이다(`journal/dispatch.go:323-327`·`:351`).
`ClassifyBrokerRefusal` 안의 본문 부분문자열 분류(`classifyRefusalBody`, `failclosed.go:182-183`·`:220-237`)도 먼저 가로챌 수 있다.

**결정.** R1 의 코드 분류기(`classifyRefusalCode`, 3판 D1)는 세 결과를 낸다:

- **확정 거절** — code 가 확정 거절 목록에 있고 두 자리(최상위 `code`·`error.code`)가 모순되지 않음 → `DispatchRejected`.
- **모호 강제** — 두 자리가 **다른 값**으로 모두 있음 → 즉시 `DispatchAmbiguous` 로 반환하고 **뒤의 분류(`ClassifyBrokerRefusal`·상태 코드)를
  타지 않는다.** 그래서 422 본문에 모순 code 가 있어도 확정 거절로 떨어지지 않는다.
- **판정 없음** — JSON 아님 · code 없음 · 목록 밖 → 종전 경로(`ClassifyBrokerRefusal` → 상태 코드).

거절 본문에서 "확정 성공" 은 나오지 않는다 — 분류기의 셋째 상태는 성공이 아니라 판정 없음이다. 분류기는 `classifyMutation` 에서
`ClassifyBrokerRefusal` **앞**에 선다(3판 D1 과 같은 자리).

### D−3.6 N7 — "감싸이면 거절 쪽" 정정 (이연된 절의 요구로)

D−2.3 조건 4 의 "오류가 한 겹 더 감싸이면 모양이 바뀌므로 거절 쪽으로 틀린다" 는 **거짓**이다 — `fmt.Errorf("…: %w", apiErr)` 는 표식
`official: API error <n>: ` 를 정확히 한 번 남기고 그 뒤 JSON 도 그대로다. `statusOf` 는 `errors.As` 로 감싼 오류에서도 상태를 꺼내고
(`classify.go:105-108`) `detail` 에는 감싼 전체가 붙는다(`:70`). 재분류가 다시 열리면 조건 4 는 표식 부분문자열이 아니라 **구조**로 강화한다 —
`detail` 이 `ClassifyHTTPMutation` 의 상태 코드 문장 + `": "` + `APIError.Error()` 로 **정확히** 구성됐는지(앞·뒤 덧붙임 없음), 그리고 감싼
오류(앞·뒤 문맥 덧붙임, 표식 둘) 픽스처가 거절됨을 시험으로 고정한다. 이연 중이므로 5판에서는 요구로만 남긴다.

### D−3.7 N2 — 새 critical 들의 전달 비용 (이름 붙인 잔여)

> **6판에서 해소 — D−4.6 enqueue-only.**

5판이 더하는 critical(청소의 새 트리거 D−2.7 · park 원인 알림 D−3.2 · 기동 행 실패 D−2.6-4)은 오늘의 알림 경로를 탄다 — 관측 루프에 주입된
notifier(`internal/app/engine/exitwiring.go:341-342`)가 동기로 부르고(`exitloop.go:1710`) 포지션을 차례로 처리하며(`:451-465`) 전달은
`n.mu` 아래 동기 발행·재시도다(`internal/obs/notifier.go:254-255`·`:309`·`:428-433`·`:525-527`). 그래서 전달이 막힌 동안 **다른 포지션의 손절이
지연될 수 있다.** 그 뮤텍스와 동기 전달의 제거는 **a092(동기 deliver 제거)와 a124(실행자)** 의 소관이다 — 5판은 그것을 고치지 않고 **잔여로
명명**하며, tasks 「선후 관계」 에서 a092 를 "독립" 이 아니라 **이 잔여의 해소자**로 적는다.

### D−3.8 N8 — N = `obs.DefaultCriticalAttempts` 는 관례 일관성

유지한다. 다만 명기한다: a124 는 **전달 시도** 실패를 세고 a094 는 **청소 주기** 실패를 센다. 같은 상수를 쓰는 것은 두 "계속 실패" 판정을
한 값으로 맞추는 **관례 일관성**이지, 청소 실패 셋이 일시적이라는 **증거가 아니다.** 알림 튜닝이 그 상수를 바꾸면 청소 정책도 함께
바뀐다(결합). 두 정책이 따로 움직여야 한다는 요구가 생기면 청소 전용 상수로 분리한다.

### D−3.9 N6/F7 — a089 처분은 freeze 의 미충족 전제

D−2.8 의 두 결말은 그대로지만 **어느 쪽도 아직 실행되지 않았다.** a094 는 a089 처분 전에 freeze 할 수 없다. 5판은 작성하되 **5라운드는 사용자 답
뒤에** 돈다(Manager). a089 결정의 긴급도는 사용자 큐에서 한 단계 올라갔다(a094 freeze 의 직접 차단자).

**해소(2026-09-28).** 사용자가 a089 불구현 아카이브를 승인했고 `64a1b2b3` 이 `--skip-specs` 로 아카이브했다 — D−2.8 의 「a089 가 아카이브되면」
갈래가 성립했다. a089 R2 의 SHALL NOT 은 main spec 에 들어가지 않았고 a094 R1 의 SHALL 은 그대로다. F7/N6 은 닫혔다.

## D−2. 4판 — 3라운드(codex) 반영 (2026-09-27)

> **이 절이 이긴다.** 아래 D−1~D7 은 3판 본문이며, 이 절과 충돌하는 자리는 이 절을 따른다.
> 근거 줄은 HEAD 기준(3라운드 트리 base `3937e341` 과 Go 동일). 3판 본문의 옛 좌표는
> `analysis/third-round-errata.md` 로 옮겨 읽는다. 코드 영수증이 없는 결정은 만들지 않고
> 「열린 질문」 으로 남긴다. 3라운드 판정·발견 원문: `review.md` 「3라운드」, `analysis/freeze-review/codex-r3-output.md`.

### D−2.1 무엇이 바뀌었나 — 한눈에

| 3판 | 4판 | 발견 |
|---|---|---|
| park(`UNRESOLVED_IN_DOUBT`)에서도 발의를 해제한다 | **해제하지 않는다.** 해제는 **입증된 비수용** 종결(`FAILED_CONFIRMED`·`NOT_DISPATCHED`)에서만 — 세션 중 `submit` 의 해제도 같은 조건으로 좁힌다(결과를 쓰지 못한 제출은 무장 유지) | F1 |
| 소급 재분류: 저장된 IN_DOUBT 본문에 확정 거절 code 가 있으면 종결 | **원 발주 응답임이 전이 기록으로 양성 식별되는** attempt 로 한정 | F2 |
| R2: 청소가 브로커 미체결(엔진 밖 주문 포함)을 보고 취소한다 | **엔진 귀속 주문만.** 사람이 넣은 외부 주문의 취소는 4판에서 **빼고 사용자 결정 대기**(D−2.4) | F3 |
| 종결 → 해제 후처리, position 기준 해제 | **기대 intent 대조** + **기동 시 따라잡기**(충돌 완결성) | F4 |
| detector 스냅샷 재사용 | 4판 R2 는 브로커를 읽지 않으므로 **불요.** 스냅샷 공표·신선도 계약은 D−2.4 가 R2 를 다시 넓힐 때의 **선행 조건**으로 적는다 | F5 |
| tasks 4.4 「`cmd/tossctl/engine.go:374` 무변화」 | 실재 좌표의 **기동 순서 요구**로 재핀 | F6 |
| a089 와의 관계 「독립」 | **규범 충돌**로 기록, a089 처분의 두 결말별 처리 | F7 |
| `PENDING_CANCEL` 제외 | **엔진이 낸 미종결 취소 attempt** 로 정의, 새 계수에만 적용, 기존 지연 타이머는 그대로 | F8 |
| critical 전달 실패 → `ENTRY_BLOCKED` | 프로세스 안 latch + 재시작 시 미전달 알림 재차단은 참, durable 운영 모드 투영은 미배선(AC1) | F9 |

### D−2.2 F1 — 발의 해제는 입증된 비수용에서만

**코드가 주는 것.**

- 종결 상태의 뜻은 `internal/journal/durability.go:37-45` 가 정한다 — `StateConfirmed` "the mutation exists at the broker",
  `StateNotDispatched` "never left the process", `StateFailedConfirmed` "broker definitively rejected it",
  `StateUnresolvedInDoubt` "we could not prove either way".
- `FAILED_CONFIRMED` 에 이르는 길 — dispatch 가 확정 거절로 분류한 경우(`DispatchRejected` → `Settle(StateFailedConfirmed, …)`,
  `dispatch.go:158-163`), 해소가 부재를 증명한 경우(`ResolveFailed`, `resolution.go:63-70` "the mutation was proven not to have
  happened"), 운영자 해소(`OperatorResolve`, `resolution.go:114-150`), 그리고 전략 런타임의 같은 종결(`internal/journal/strategy_dispatch_runtime.go:699`).
  넷 다 "접수되지 않음" 을 뜻하는 같은 상태다. `NOT_DISPATCHED` 는 `DispatchNotSent`(`dispatch.go:151-156`)와 재시작 규칙의
  RECORDED 종결(`journal/recovery.go` — 전송 시작 기록이 없는 attempt).
- 이중 매도의 실제 방벽은 발의 자체다 — `armExitProposalTx` 는 `pending_action` 이 차 있을 때만 두 번째 발의를 거절한다
  (`apply_hook.go:666-668`). 매도(노출 축소)는 gateway 의 unresolved 검사를 건너뛴다(`internal/execgw/gateway.go:815-816`
  — `!plan.raisesExposure` 면 반환). 따라서 park 에서 발의를 비우면 **원 매도가 살아 있어도** 다음 발의가 무장되고 제출된다.
  3판의 근거 1("1차 방벽은 `armExitProposalTx`")은 그 방벽을 해제가 치운다는 사실을 놓쳤다(3라운드 F1).

**결정.**

- 해제 대상은 `FAILED_CONFIRMED`·`NOT_DISPATCHED` 종결뿐이다. `CONFIRMED` 는 3판대로 해제하지 않는다.
- **`UNRESOLVED_IN_DOUBT`(park)에서는 해제하지 않는다.** 원 주문의 존재가 미지인 동안 발의는 무장된 채 남는다.
  이것은 3판이 막으려던 「영구 무보호」를 park 된 포지션에 되살린다 — 그 비용은 park 가 이미 요구하는 **사람의 해소**로만
  푼다(정본 spec order-execution 「해소 불능」: "운영자 해소만 허용"). park 의 유일한 출구는 `Journal.OperatorResolve` 다
  (`internal/journal/resolution.go:114-150` — "the only exit from UNRESOLVED_IN_DOUBT", 대상은 `CONFIRMED`·`FAILED_CONFIRMED`,
  운영자 이름과 메모 필수). **다만 그 함수의 비시험 호출자는 0 이다** — 운영자가 그것을 부를 도구 경로가 저장소에 없다
  (2026-09-27 전수 grep). 운영자가 `FAILED_CONFIRMED` 로 닫으면 그것은 입증된 비수용이 되고 D−2.5 의 기동 따라잡기가 다음
  기동에서 발의를 푼다. 그 **세션 안에서** 곧바로 풀지(도구 경로에서 해제를 부를지)는 코드에 근거가 없다 — **열린 질문 Q4-1**.
- 3판은 이 사건(475150·080220)의 해동을 **D−2.3 의 소급 재분류 → `FAILED_CONFIRMED` → 해제**로 보았다(D3 「단, 이 결정은…」).
  **0.5h 재측정이 그 전제를 무너뜨렸다**(아래) — 두 행은 이미 park 됐고, 해동은 운영자 도구 경로(Q4-1)로만 일어난다. 272210 은 3판 기록상 `PROPOSAL_CANCELLED` 라이브락이며
  (`tasks.md` 6.2) 이 절과 무관하다.
- **0.5h 재측정 결과 (2026-09-27, 운영 원장 읽기 전용 · 브로커 호출 0 · Manager 승인).** 두 행은 **이미 park 됐다** —
  `034e5b79…`(475150) `IN_DOUBT → UNRESOLVED_IN_DOUBT` 2026-08-08T01:10:40Z, `8f68e7c3…`(080220) 같은 전이 2026-08-08T01:09:53Z,
  둘 다 `reason_code = in_doubt_unresolved`. 발의는 **지금도 무장돼 있다**(`exit_states`: `STOP_LOSS_LADDER`, level `0` / `-1`).
  두 행의 모호 전이는 D−2.3 의 조건 2~7 을 모두 채운다(`DISPATCH_STARTED → IN_DOUBT`·`dispatch_outcome_unknown`·ACKED 없음·
  `broker_order_id` 없음·표식 정확히 1회·code 있음·재생 0) — **조건 1(현재 `IN_DOUBT`)만 못 채운다.** 그래서 소급 재분류는 이 두 행을
  건드리지 않고, 이 사건의 해동 경로는 **Q4-1 의 운영자 도구 경로**(운영자가 `OperatorResolve` 로 `FAILED_CONFIRMED` 를 정하면 같은
  해제 판정이 그 자리에서 발의를 푼다 — D−2.11)다. 그 도구는 아직 없다(`OperatorResolve` 비시험 호출자 0).
- **전제의 나이(gstack 문서 리뷰 P1, 위 재측정으로 닫힘).** 그 경로는 두 행이 지금도 `IN_DOUBT` 라는 전제에 선다. 그 근거는 2026-08-07 원장 기록뿐이고
  (design D−1), 재시작은 해소를 돌려 그 행을 park 할 수 있다(tasks 8.2 · proposal 의 재시작 서술). 엔진 컨테이너는 2026-09-08 에
  교체됐다. 두 행이 이미 `UNRESOLVED_IN_DOUBT` 라면 4판에는 그것을 푸는 경로가 없다(park 는 해제하지 않고 `OperatorResolve` 는 호출자가
  없다) — 그 경우 이 사건은 Q4-1 에 막힌다. **두 행의 현재 상태를 4라운드 전에 읽기 전용으로 잰다**(task 0.5h, 운영 원장 읽기라 사람 승인).

### D−2.3 F2 — 소급 재분류는 원 발주 응답만

**코드가 주는 것 — provenance 는 이미 원장에 있다(스키마 추가 없음).**

- `mutation_attempts` 는 `broker_order_id`("assigned once the broker acks", `internal/journal/schema.go:223`)를 갖고,
  `attempt_transitions` 는 **추가 전용** 전이 기록(`from_state`·`to_state`·`reason_code`·`detail`, `schema.go:236-245`)이다.
- 발주 응답의 본문이 `detail` 에 들어가는 길 중 **모호(IN_DOUBT)로 가는 것**은 하나다 — `classifyMutation` 이 **상태 코드를 받은** 오류를
  `journal.ClassifyHTTPMutation` 으로 분류하고 `detail` 에 `": " + err.Error()` 를 붙인다(`internal/execgw/classify.go:64-72`;
  `APIError.Error()` = `"official: API error %d: %s"` 로 본문을 담는다, `internal/official/errors.go:27-28`). 그 분류가 모호면
  `DispatchAmbiguous` → `MarkInDoubt` 이고 이때 전이는 **`DISPATCH_STARTED → IN_DOUBT`**, `reason_code = "dispatch_outcome_unknown"`
  (`dispatch.go:204-209`, `lifecycle.go:96`). (`ClassifyBrokerRefusal` 도 `err.Error()` 를 담지만(`classify.go:54-58`) 그 결과는
  확정 거절 또는 모호 두 갈래이고, 모호 갈래는 확인 요청 분기뿐이다 — reason code 가 달라 아래 조건 3 이 배제한다.)
- 반대로 접수 뒤 readback 실패는 **`MarkAcked`(broker_order_id 기록) 다음** `ACKED → IN_DOUBT`, `reason_code =
  "ack_round_trip_unconfirmed"` 이다(`dispatch.go:181-195`, `:71`). 그 `detail` 의 오류는 readback 의 것이지 발주의 것이 아니다.

**결정 — 재분류 대상은 다음을 모두 만족하는 attempt 로 한정한다(하나라도 아니면 그대로 둔다).**

1. `kind = 'PLACE'`, 현재 `state = 'IN_DOUBT'`, `broker_order_id = ''`.
2. 전이 기록에 `to_state = 'ACKED'` 행이 **없다**.
3. `IN_DOUBT` 로 가는 전이 행이 **정확히 하나**이고 그 행이 `from_state = 'DISPATCH_STARTED'`,
   `reason_code = 'dispatch_outcome_unknown'` 이다.
4. 그 행의 `detail` 에 공식 클라이언트 오류 표식 `official: API error <n>: `(`internal/official/errors.go:27-28` 의 형식)이
   **정확히 한 번** 나오고, 그 뒤가 JSON 으로 읽힌다. 엔진이 덧붙인 산문(`"HTTP <n> does not prove …"`)은 매칭하지 않는다 —
   출처는 조건 1~3 의 전이 컬럼이 정하고, code 는 JSON 본문에서만 읽는다. 모양이 맞지 않으면(표식 0개·2개 이상, JSON 아님)
   그대로 둔다(오류가 한 겹 더 감싸이면 모양이 바뀌므로 거절 쪽으로 틀린다). 전송 실패 분기(`dispatch.go:310-316`)와 상태 없음
   분기(`:330-336`)는 reason code 는 같지만 표식이 없거나 다른 모양이라 여기서 빠진다.
5. 본문의 최상위 `code` 와 `error.code` 가 **둘 다 있으면 같아야 하고**, 하나만 있으면 그것을 쓴다. 둘이 다르면 **모호 —
   그대로 둔다**(3라운드 F2 "conflicting top-level/nested-code handling; ambiguity must remain IN_DOUBT").
6. 그 code 가 확정 거절 목록(4판 = `opposite-pending-order-exists` 하나)에 있다.
7. **재생된 적이 없다** — `coalesce(replay_count,0) = 0` 이고 `last_replay_at IS NULL`. 재생은 attempt 를 끝낼 때만 전이를
   쓰고(`internal/execgw/replay.go:329`·`:355`) 그 밖에는 `replay_count`·`last_replay_at` 만 쓴다(`internal/journal/replay.go:123`,
   컬럼 `execution_contract.go:265-266`). `RefundReplay` 는 횟수를 되돌리지만 시각은 남긴다(`journal/replay.go:146-151`) — 그래서
   시각까지 본다. `409 request-in-progress` 재생 응답은 원 요청이 아직 처리 중이라는 뜻이므로, 재생된 attempt 를 원 409 로 은퇴시키면
   그 사실과 모순된다.

**규칙 문장으로 고정한다 — `MarkAcked` 뒤 readback 실패의 `detail` 은 재분류 근거가 될 수 없다**(2·3 이 구조로 배제).
재생된 attempt 는 조건 7 이 배제한다. (조건 3 의 "정확히 하나" 는 IN_DOUBT 로 들어오는 전이가 `DISPATCH_STARTED`·`ACKED` 둘뿐이라
(`lifecycle.go:43-46`) 재생을 가르지 못한다 — 3판 리뷰 뒤의 gstack 문서 리뷰가 잡았다.)

**범위와 부수 효과.** 조건은 매도·매수를 가르지 않는다 — 전략 런타임의 매수 발주도 같은 모양을 쓴다(`internal/journal/strategy_dispatch_runtime.go:722`·
`:935`). 확정 거절은 방향과 무관하게 "접수되지 않음" 이므로 4판은 **PLACE 전체**를 대상으로 한다. `FAILED_CONFIRMED` 전이는 같은
트랜잭션에서 그 결정의 예약을 푼다(`durability.go:653-663`) — 재분류의 결과도 그렇다. 재분류는 새 reason code 하나로 기록한다
(이름은 구현 로트, 골든·`AllReasonCodes` 갱신 — task 4b.8).

**정지 조건.** 위 여섯은 기존 컬럼만 읽는다 — 스키마 변경이 필요해지면(예: 4 의 문자열 모양이 다른 판본에서 달랐음이
원장에서 확인되면) 그 자리에서 멈추고 보고한다.

### D−2.4 F3 — R2 는 엔진 귀속 주문으로 축소, 외부 주문 취소는 사용자 결정 대기

**코드가 주는 것.**

- 오늘 청소의 목록은 `Journal.LiveOrdersForSymbol`(`internal/journal/fills.go:1849-`)이다 — `CONFIRMED` 인 `PLACE`/`AMEND`
  attempt 로 `broker_order_id` 가 있고, 소유 intent 가 유일하며, 종결 체결 스냅숏이 없는 주문을 lineage 로 현재 번호까지
  따라간 것이다. **이것이 엔진 귀속(positive attribution)의 정본이다.**
- 청소는 `withPending` 이면 같은 종목의 **모든** 매도를 치운다(`exitloop.go:1448-1450` `if !buy && !withPending`). 목록이
  원장뿐인 오늘은 그 "모든" 이 엔진 매도로 닫혀 있다. 3판처럼 목록을 브로커 미체결로 넓히면 다른 포지션·사람의 보호 매도까지
  들어온다 — `withPending` 조건에 귀속 경계가 없다(3라운드 F3). detector 의 비추적 주문 파싱도 범위를 다 채우지 않는다
  (`internal/filldetect/detect.go:407-420`).

**결정.**

- **4판 R2 의 목록은 엔진 귀속 주문뿐이다 — 곧 오늘의 `LiveOrdersForSymbol` 이다.** 브로커 미체결 목록을 청소 대상에 더하지
  않는다. 그러면 3판 R2 의 배선(`ExitObserverOptions` 새 필드 · 스냅샷 주입 · `Snapshot` 필드 추가)과 dedup·감사 요구는
  **필요 없어진다.**
- 4판 R2 에 남는 것: (a) 빈 가격을 치우기 실패로 읽지 않는다(3판 D2 결정, `floatOf`·`exitloop.go:1779-1785`, a087 대비),
  (b) 연속 치우기 실패의 새 알림 트리거와 그 `PENDING_CANCEL` 제외(D−2.7), (c) 발행·수량/가격 해석·취소 실패를 `clear=false`
  로 흡수하는 기존 규칙(`exitloop.go:1465`·`:1471`·`:1485-1486`) 유지. 원장 목록 읽기 실패는 오늘처럼 오류로 반환된다(`:1443-1445`).
- **이 사건에 대한 귀결을 숨기지 않는다.** 475150·080220·272210 의 반대 매수는 엔진이 낸 것이 아니다(proposal 「Why」 —
  세 종목 `intents` 에 BUY 0건). 4판은 그것을 **치우지 않는다.** 그래서 4판 뒤 이 모양은 「영구 동결」 에서 「매 주기 거절되고
  보이는 반복」 으로 바뀐다 — R1 이 409 를 종결로 분류 → 해제(`submit` B10, 오늘 경로) → 다음 관측이 다시 발의 → 같은 409.
  그 반복은 `alertProposalRefused` 로 보고되고(`exitloop.go:1654-`) 전송은 a096 의 재알림 창이 묶는다. **손절은 사람이 반대
  주문을 치울 때까지 나가지 않는다.**
- **그 반복의 비용(gstack 문서 리뷰 P1).** 쓸 수 있는 시세가 있는 주기마다 실제 SELL POST 가 하나씩 나간다(272210 의 라이브락
  간격 중앙값 5.0초 — tasks 6.2). 그중 하나라도 429·5xx 를 받으면 그 응답은 조건 4 의 모양(JSON 본문의 code)이 아니어서 모호로
  남고, 세션 중 해소가 없으므로 재시작 때 park 된다 — 그러면 D−2.2 대로 발의가 무장된 채 영구히 남는다. R1 이 `FAILED_CONFIRMED`
  로 종결한 뒤의 재발의 간격을 둘지는 코드에 근거가 없다 — **Q4-6**.

**사용자 결정 대기 — 엔진 밖(사람이 넣은) 주문의 취소.** 확대는 사용자 몫이다(Manager 방향). 결정이 "넓힌다" 로 나면 그
change(또는 5판)는 다음을 **모두** 선행 조건으로 가져야 한다 — 3라운드 F3·F5 가 요구한 것:

1. 외부 주문의 **귀속 규칙** — 무엇이 "이 포지션의 보호를 막는 주문" 인지(방향·수량·식별자 완결성), 불완전하거나 모르는 주문은 거절.
2. 다른 포지션·사람의 **보호 매도는 취소하지 않는다**는 경계(`withPending` 매도 청소를 발의에 연결된 매도로 제한).
3. **공표된 OPEN 스냅숏 계약**(F5): 계정 범위의 완전한 OPEN 세대를 불변으로 공표, 수집 **시작** 시각 기준 신선도, 「비어 있음」
   과 「읽지 못함」 의 구분, 실패 처리, 청소 쪽 비차단 읽기. 오늘 detector 에는 그것이 없다 — 스냅숏은 한 주기 안에서 소비되고
   (`detect.go:293-352`) 추적 주문 보조 읽기가 섞이며(`:423-455`) 주기는 조회 시간 **더하기** 3초 잠(`:515-525`,
   `PollInterval = 3s` `:128`)이라 "최대 3초 전" 이 성립하지 않는다. 신선도 상한 값은 코드에 근거가 없다 — **열린 질문 Q4-2**.
4. 취소 사실의 감사 기록(3판 D2 「감사」).

### D−2.5 F4 — 종결 → 해제 이양의 충돌 완결성과 기대 intent

**코드가 주는 것.**

- `exit_states.pending_intent_id` 는 발의를 무장할 때 쓰이고(`apply_hook.go:669-673`), attempt 는 `intent_id` 로 intent 에
  묶인다(`schema.go:216-218`). 연결은 이미 있다.
- `ResolveExitProposal` 은 position 으로 찾아 세 컬럼을 비운다 — `pending_intent_id` 를 **읽기만 하고 대조하지 않는다**
  (`apply_hook.go:825-884`). 늦게 도착한 해제가 그 사이 재무장된 **새** 발의를 지울 수 있다.
- attempt 종결과 발의 해제는 다른 트랜잭션이다. 종결된 attempt 는 `PendingAttempts` 에서 빠지므로(`journal/recovery.go:30-33`)
  두 쓰기 사이의 충돌은 재시작 복구가 다시 찾지 못한다.

**결정.**

- **기대 intent 대조.** 해제는 "이 attempt 의 `intent_id` = 현재 `pending_intent_id`" 일 때만 비운다. 다르면 아무것도 하지
  않는다(다른 발의다). `ResolveExitProposal` 의 호출 형태를 바꾸는 편집이며(Function Logic Map 대상, High-risk).
- **기동 시 따라잡기.** 기동 복구에서, `pending_intent_id` 가 찬 `exit_states` 마다 그 intent 의 attempt 가 **모두**
  `FAILED_CONFIRMED`·`NOT_DISPATCHED` 로 종결돼 있거나, 그 intent 의 attempt 가 **하나도 없으면**(발의 무장 뒤 `Prepare` 전의 충돌 —
  `Prepare` 는 전송 전에 커밋하므로(`internal/execgw/gateway.go:540-543`) attempt 가 없으면 아무것도 나가지 않았다) 기대 intent
  대조로 해제한다. 따라잡기는 멱등이다(`ResolveExitProposal` B8). 두 쓰기 사이의 충돌은 다음 기동이 닫는다.
- **세션 중 해제는 `submit` 에서 좁힌다(gstack 문서 리뷰 P0).** 오늘 `submit` 의 `default:` 갈래(`exitloop.go:1410-1416`)는
  `alertProposalRefused` 뒤 발의를 푼다. 그런데 gateway 는 attempt 를 기록한 뒤(`gateway.go:550` `out.AttemptID` 설정) 전송 결과를
  원장에 쓰다 실패하면(`MarkAcked`·`Settle`·`MarkInDoubt` 쓰기 실패, 종료 중 ctx 취소) `State` 가 빈 `out` 을 오류와 함께 돌려준다
  (`gateway.go:747`) — 그 attempt 는 `DISPATCH_STARTED`·`ACKED` 에 남고 주문은 나갔을 수 있다. 4판은 세션 중 해제를
  `out.State ∈ {NOT_DISPATCHED, FAILED_CONFIRMED}` 이거나 **attempt 가 기록되지 않은**(`out.AttemptID == ""`) 경우로 한정한다.
  `State == ""` 이고 `AttemptID != ""` 이면 발의는 무장된 채 남고(재시작 복구가 그 attempt 를 IN_DOUBT 로 만든다) 기동 따라잡기가
  종결 뒤에 처리한다. `SymbolInFlight` 갈래(`:1406-1408`)는 전송 전 거절이므로 그대로 둔다.

### D−2.6 F6 — 기동 순서를 실재 좌표로

**코드가 주는 것.** 기동 복구는 `recoverThenReady(engineRecoverySequence(recovery), ready, …)`(`cmd/tossctl/engine.go:677`)이고
`engineRecoverySequence` 는 `r.Run` 을 돌려준다(`:604-606`). `Recovery.Run`(`internal/reconcile/recovery.go:238-`)은
① `RecoverPending`(재시작 규칙) → ② `PendingAttempts` → 재생 → `Resolver.Resolve` 순이다. 3판 tasks 4.4 가 가리킨
`cmd/tossctl/engine.go:374` 의 호출은 a102 `6cd643ca` 이후 없다(정오표 §2·§4).

**결정 — 순서 요구로 재핀한다(좌표 단언 대신).**

1. **재분류(D−2.3)는 `Resolver.Resolve` 가 IN_DOUBT 후보를 소비하기 전에** 돈다 — `engineRecoverySequence` 이음매에서 `r.Run`
   앞. 해소가 먼저 돌면 후보를 park 하거나 다른 증거로 종결해 원 발주 응답의 증거 자리를 지나간다.
2. **따라잡기(D−2.5)는 `engineRecoverySequence` 클로저 안, `r.Run` 뒤**에 돈다. 루프는 Recover 가 반환한 뒤에만 시작하므로
   (`internal/app/engine/runtime.go:289-295`) 그 자리면 첫 관측 주기 전에 끝난다.
4. **실패 의미(gstack 문서 리뷰 P1).** Recover 가 오류를 돌려주면 런타임은 루프를 하나도 시작하지 않는다(`runtime.go:294`) — 엔진이
   서면 어느 포지션에도 손절이 없다. 그래서 재분류와 따라잡기는 **행 단위로 실패를 흡수**해야 한다: 한 행의 읽기·쓰기 실패는 그 행을
   건드리지 않은 채(오늘 상태 그대로 — 안전 쪽) 기록하고 다음 행으로 가며, Recover 의 반환값을 바꾸지 않는다. 그 실패는 **critical 로
   보낸다**(Q4-5 결정, key 는 포지션 단위).
3. `Recovery.Run` 본문(재시작 규칙 → 재생 → 해소)과 인터록 의미는 바꾸지 않는다.

### D−2.7 F8 — `PENDING_CANCEL` 을 엔진이 낸 미종결 취소로 정의

**코드가 주는 것.**

- `mutation_attempts.kind = 'CANCEL'` 과 `target_order_id` 가 엔진이 낸 취소를 그 대상 주문에 묶는다(`schema.go:219-222`).
- 청소는 취소 결과가 `StateConfirmed` 일 때만 치운 것으로 본다(`exitloop.go:1485-1486`). 취소의 `CONFIRMED` 는 **인수(ack)** 다 —
  dispatch 는 접수를 `CONFIRMED` 로 종결하고(`dispatch.go:198-202`) 취소에는 readback 확인이 없다(3라운드 F8 인용
  `internal/execgw/roundtrip.go:72-75`). 인수와 호가 이탈은 다르다.
- 기존 지연 타이머: 치우지 못하면 `noteDelay` 가 시작되고(`exitloop.go:1252`) 한계 30초(`DefaultExitLiquidationDelayBound`,
  `:116`)를 넘으면 `EventExitLiquidationDelayed` 를 한 번 낸다(`:1675-`). 그 이벤트는 **이미 critical** 이다
  (`internal/obs/event.go:336`).

**결정.**

- 3판의 "연속 3회 → critical 로 올린다" 는 틀린 서술이다 — 이벤트는 이미 critical 이다. 4판은 이것을 **새 트리거**로 적는다:
  같은 포지션의 청소가 연속 N회 `clear=false` 로 끝나면, 기존 30초 타이머와 **별도로** 같은 이벤트 타입을 한 번 낸다(더 이른 신호).
  N 은 `obs.DefaultCriticalAttempts`(a124 관례, `internal/obs/notifier.go:45`) — **Q4-7 결정**. **기존 타이머는 바꾸지 않는다** — 타이머의 시작·해제·한계·중복 방지 무변화.
  그래서 새 트리거는 **타이머와 다른 알림 key** 를 써야 하고(타이머의 key 는 `type|positionID`, `exitloop.go:1688`; 같은 key 면
  알림 중복 제거(`DefaultRemindAfter` 1시간, `internal/obs/notifier.go:50-59`)가 타이머의 경보를 삼킨다) `delayAlerted`
  (`exitloop.go:253`·`:1682`)를 **건드리지 않는다.** 결과는 한 사건에 critical 푸시 둘(이른 것·30초 것)이다.
- **`PENDING_CANCEL` 의 정의**: 그 주기에 치우지 못한 주문 중 **엔진이 낸 `CANCEL` attempt 가 그 `target_order_id` 로
  `RECORDED`·`DISPATCH_STARTED`·`ACKED` 상태**인 주문. **`IN_DOUBT` 취소는 넣지 않는다** — `Submit.Cancel` 은 동기이고 취소에는
  readback 이 없으므로(`internal/execgw/roundtrip.go:72-75`) 다음 주기까지 미종결인 취소는 사실상 IN_DOUBT 이며, 세션 중 해소는 범위
  밖이라 재시작까지 그대로다. 그것을 "치우는 중" 으로 빼면 새 트리거가 세션 내내 꺼진다(gstack 문서 리뷰 P1). 판정은 원장만 읽는다.
- **제외는 새 트리거의 계수에서만.** 그 주기에 치우지 못한 주문이 **전부** 위 정의에 들면 계수를 늘리지 않는다. 하나라도 다른
  이유로 못 치웠으면 센다. 기존 30초 타이머에는 어떤 제외도 적용하지 않는다(지연 경보의 약화 금지 — 안전 불변식 §4).
- 인수(`CONFIRMED`)를 호가 이탈로 믿는 오늘의 청소 규칙은 **바꾸지 않는다.** 이탈 관측을 요구하려면 브로커 읽기가 필요하고
  그것은 D−2.4 의 사용자 결정에 딸린다 — **열린 질문 Q4-3**.

### D−2.8 F7 — a089 R2 와의 규범 충돌: 두 결말

충돌 문장은 정오표 §5. a089 의 처분은 사용자 큐에 있다(2026-09-27, "a089 불구현 아카이브 + a090 신설" 제안). 4판은 별도 작업
없이 그 결말을 따른다.

- **a089 가 아카이브되면**(**성립 — a089 아카이브(`64a1b2b3`)로 해소, 2026-09-28**): a089 R2 의 SHALL NOT("이 기록에 따라 … 동작을 분기해서는 안 된다")은 main spec 에 들어가지 않는다.
  a094 R1 의 SHALL(`code` 로 종결)은 그대로 둔다. a094 쪽 "a089 와 겹치지 않는다" 류 문장(proposal 「관계」·tasks 도식)은
  **4판에서 이미 "규범 충돌" 로 고쳤다** — 아카이브 뒤에는 그 문장에 "a089 아카이브로 해소" 만 덧붙인다.
- **a089 가 유지되면**: a089 R2 의 금지를 **기록 기능에 한정**하도록 a089 쪽 문장을 좁혀야 한다 — "이 기록 필드에 따라 동작을
  분기해서는 안 된다; 별도 요구로 명세된 증거 기반 실행 분류기는 이 금지의 대상이 아니다"(3라운드 F7 권고). 재생·주문 신원
  보호는 완화하지 않는다. 두 delta 를 함께 조정하기 전에는 둘 중 어느 것도 freeze 하지 않는다.
- **그러므로 a094 는 a089 의 처분에 의존한다** — tasks 「선후 관계」 의 "넷 중 어느 것에도 의존하지 않고" 는 4판에서 그렇게 고친다.

### D−2.9 F9 — AC1 의 정확한 범위

- 전달 실패한 critical 알림은 프로세스 안에서 진입 게이트를 잠그고(`internal/obs/notifier.go:570-572`) durable 운영 모드를
  승격한다(`:378-383`). 재시작 시에는 **미전달 알림이 따로 진입을 막는다** — `restoreAlertEntryLatch`
  (`internal/app/engine/gateway.go:153-168`, 호출 `:269`)가 `UndeliveredCount > 0` 이면 `ReasonAlertUndelivered` 로 차단한다.
- 배선되지 않은 것은 **durable 운영 모드의 투영**뿐이다 — `SetModeProjector`(`operating_mode.go:294`)·
  `RestoreOperatingModeProjection`(`:574`) 비시험 호출자 0(a124 AC1). 그래서 전달이 성공한 뒤에는 모드의 사람 해제 요구가
  재시작을 건너지 못한다.
- 3판의 "critical 전달 실패는 `ENTRY_BLOCKED`까지 간다" 는 **부분적으로 참**이다(프로세스 안 latch + 재시작 시 미전달 알림
  재차단). 거짓 critical 의 차단 비용은 실재하므로 D−2.7 의 **좁은** 제외는 유지된다. 제외를 넓히거나 무기한 억제하는 근거는
  되지 않는다.

### D−2.10 F10 — 낡은 논증을 현재 코드로

아래 넷은 **4판에서 본문을 고쳤다**(D2 「경계」 표 두 행 · D2 「결정: (a)…」 문단 · `record` 번들의 FLM Safety conclusion 과
BTM 미진입 요약). 다른 자리의 같은 논증은 이 절대로 읽는다.

- **`record` 의 게이트 논증**(3판 D2 「경계」 표 `record` B3). 현 번호는 **B5**(`exitloop.go:1223` —
  `orderable && (snapshot.CancelPendingFirst || isFullExit(proposal))`). 그 **앞에** a111 의 B1(`:1180-1182`,
  `!o.quoteUsable(quote)` → `return nil`)이 있다 — 쓸 수 있는 시세가 없는 주기에는 청소·기록·지연 처리 어느 것도 일어나지
  않는다. 따라서 3판의 "다음 관측에서 다시 제안된다" 류 문장은 전부 **"쓸 수 있는 시세가 있는 다음 관측에서"** 로 읽는다.
- **「경계」 표의 "자기 방향(SELL)은 부재 확인" 행**은 3판 스스로 철회한 규칙이다(D2 「자기 방향 미체결의 부재 확인 — 3판에서
  철회했다」). 4판에서는 그 행을 "자기 방향은 오늘과 같이 `withPending` 일 때만 치운다(엔진 귀속 매도뿐)" 로 읽는다.
- **"critical 로 올린다"** 는 D−2.7 대로 "이미 critical 인 이벤트의 새 트리거" 로 읽는다.
- `record` 번들의 Function Logic Map·Branch Test Map 요약 문장(미진입 분기 요약)은 refresh 가 줄만 옮긴 자리다 — 4판 편집 전
  산출물 재작성(task 1.x)에서 현재 AST 기준으로 다시 쓴다.

### D−2.11 열린 질문 (코드에 근거 없음 — 지어내지 않는다)

- **Q4-1 — park 된 attempt 의 발의. 결정(Manager, 2026-09-27): 도구 경로에서 즉시 해제, 판정 함수는 하나.** 운영자가 해결했는데
  다음 기동까지 발의가 얼어 있으면 무보호 창이 늘어난다(안전 불변식 §4 — 손절 즉시성). 그래서 운영자 해소 도구(미래 작업)는 해소를
  기록한 뒤 **D−2.5 와 같은 해제 판정 함수**(intent 일치 · 비수용 종결 조건 재검)를 그 자리에서 부른다. 기동 따라잡기는 충돌 백스톱으로
  남는다. 해제 판정을 두 곳에 두지 않는다. 도구는 이 change 의 구현 대상이 아니다 — 설계 문장으로만.
- **Q4-2 — 브로커 OPEN 스냅숏의 신선도 상한**(D−2.4 를 넓힐 때만). 값과 근거.
- **Q4-3 — 취소 인수와 호가 이탈.** 청소가 취소 `CONFIRMED`(인수)만으로 치운 것으로 보는 오늘 규칙을 이탈 관측으로 바꿀지 —
  브로커 읽기를 요구하므로 D−2.4 결정에 딸린다.
- **Q4-4 — `clearTheSymbol` 의 두 번째 해제 문(gstack 문서 리뷰 P1).** 청소는 `withPending && m.state.Pending()` 이면 끝에서 발의를
  `ProposalCancelled` 로 푼다(`exitloop.go:1491-1495`). `withPending` 은 `CancelPendingFirst` — 익절이 무장된 채 손절 선을 넘은 경우다
  (`internal/exitpolicy/ladder.go:447`, `ratchet.go:432`). 그 익절의 attempt 가 `IN_DOUBT`·park 면 `LiveOrdersForSymbol`(CONFIRMED 만)에
  안 보이므로 `clear` 는 참이고 발의가 풀린다 — 그 위에 전량 손절이 나가며, 매도는 unresolved 검사를 건너뛴다(`gateway.go:815-816`).
  D−2.2 와 같은 위험이다. 반대로 무장을 유지하면 손절이 막힌다(§4). **5판에서 번복됨 — D−3.2.** (4판 당시 기록:) **결정(Manager, 2026-09-27): 현행 유지 — 손절을 낸다.** 트레이드오프의
  이름: 「보이지 않는 익절 위에 손절이 나가는 중복 위험」 대 「손절 차단」. TossOS 는 손절 지연 쪽을 늘 기각한다(안전 불변식 §4
  "손절·비상 청산의 즉시성을 약화하거나 지연하지 않는다"). **미검증 잔여**: 그 중복 매도를 브로커가 보유 초과로 거절하는지는 측정되지
  않았다 — 이 문서는 그 한계를 주장하지 않는다.
- **Q4-5 — 기동 단계의 행 단위 실패. 결정(Manager): critical 로 보낸다.** 흡수하되(루프는 뜬다) 침묵하지 않는다 — 조용한 건너뛰기는
  문이다. 알림 key 는 포지션 단위(D−2.6-4).
- **Q4-6 — R1 종결 뒤 재발의 간격. 결정(Manager): 새 간격 없음(즉시성).** 대신 **「429 류 거절 → park」 궤적을 이름 붙인 위험**으로 세운다.
  코드 영수증: 429 는 전송 층에서 타입으로 구별된다(`official.ErrRateLimited`, `internal/official/errors.go:15`·`:54-55` — 429 는
  `APIError` 가 아니라 이 sentinel 로 돌아온다). 그러나 분류는 그것을 상태 429 로 바꿔(`internal/execgw/classify.go:114-117`)
  `journal.ClassifyHTTPMutation` 의 일반 모호 갈래로 보내고(`dispatch_outcome_unknown` — 409 와 같은 reason, `journal/dispatch.go:337-341`),
  park 전이의 reason 은 `in_doubt_unresolved` 다(0.5h 실측과 같은 모양). **원장 reason 수준에서 rate-limit 을 타입으로 가를 수 없다** —
  detail 문자열(`official: rate limited`)뿐이고 그것을 매칭하는 것은 문구 판정이다. 따라서 "rate-limit 으로 park 된 것을 타입으로 구별해
  critical" 은 **구현 로트의 정지 조건**이다 — 구현이 그것을 필요로 하면 새 reason code(분류 층) 추가를 Manager 에게 올리고 멈춘다.
- **Q4-7 — 새 트리거의 연속 실패 횟수 N. 결정(Manager 지시로 영수증 탐색): a124 관례를 따른다 — `N = obs.DefaultCriticalAttempts`.**
  a124 는 freeze 된 8판 D2 에서 "계속 실패하는 발신자" 한도를 상수 `alertAttemptLimit = obs.DefaultCriticalAttempts` 로 두고 시험이 숫자를
  복사하지 않고 그 상수를 인용하게 했다(a124 `design.md:84-86`, 14회차 PASS). 그 상수는 `internal/obs/notifier.go:45` 의
  `DefaultCriticalAttempts = 3` 이고 주석이 근거를 적는다 — "the failures worth retrying through are transient … while every retry delays
  the moment the operator learns". a094 는 그 **형태**(기존 상수 인용, 숫자 복사 금지)를 재사용한다. a124 의 수치 근거(동기 경로와 같은
  시간)는 a124 고유이며 a094 의 5초 주기에 옮겨지지 않는다 — 옮겨지는 것은 "재시도로 기다려 볼 만한 일시 실패의 횟수" 라는 뜻이다.
- **사용자 결정 대기 — 엔진 밖 주문의 취소(D−2.4).**

## D−1. 3판이 고친 것 — **잠금을 잘못 지목하고 있었다**

1판과 2판은 **미정산 attempt**를 동결의 원인으로 봤다. **틀렸다.**
2라운드 리뷰가 잡았고 매니저가 소스와 원장으로 재확인했다(`review.md` §2.1).

**진짜 잠금은 무장된 채 남은 발의(`exit_states.pending_action`)다.**

```text
submit B8  exitloop.go:1296-1300
  case out.State == StateInDoubt || StateUnresolvedInDoubt:
      return nil                      ← release() 를 부르지 않는다
                                        pending_action 이 무장된 채 남는다
        │
        ↓
EvaluateLadder  B25 :439  observed < baseline           ← 손절 조건은 성립한다
                B26 :441  PendingAction == ActionLadderStop
                          → Suppressed = SuppressedPending, return
                                        ← out.Proposal 을 채우지 않는다 = 빈 발의
        │  (RATCHET도 같다 — EvaluateRatchet B17 :423)
        ↓
record  exitloop.go:1082
  orderable := snapshot.Orderable && !proposal.Zero()   → false
        │
        ↓
record  exitloop.go:1117
  if orderable && (CancelPendingFirst || isFullExit)    ← 게이트가 열리지 않는다
        │
        ├─ clearTheSymbol 도달하지 않음   ← R2가 바꾸는 코드
        └─ submit 도달하지 않음           ← R1이 바꾸는 코드
```

**원장 실측 (2026-08-07)** — `pending_intent_id`가 얼어붙은 attempt를 가리킨다:

| 종목 | `pending_action` | `pending_level` | 가리키는 attempt |
| --- | --- | --- | --- |
| **475150** | `STOP_LOSS_LADDER` | `0` | `034e5b79…` `IN_DOUBT` `settled_at=NULL` |
| **080220** | `STOP_LOSS_LADDER` | `-1` | `8f68e7c3…` `IN_DOUBT` `settled_at=NULL` |
| 272210 · 066570 · TSLA | `None` | — | — |

**따라서 1·2판의 R1·R2는 이 두 종목에 대해 실행되지 않는 코드였다.**
272210은 `pending_action`이 비어 있어 매 주기 발의가 나므로 R1·R2가 실제로 돕는다 —
**동결에 두 모양이 있었고 앞선 두 판은 하나만 다뤘다.**

### 그래서 3판이 더하는 고리 하나

`pending_action`을 NULL로 만드는 non-test writer는 셋뿐이다:

| writer | 자리 | 호출자 |
| --- | --- | --- |
| `Journal.ResolveExitProposal` | `apply_hook.go:846-849` (**B9**) | `ExitObserver.release`(`exitloop.go:1317`) **하나** |
| `ApplyTx.ResolvePending` | `apply_hook.go:457-458` | 체결 적용 안 |
| `resetExitStateForReadoptTx` | `apply_hook.go:704-706` | 운영자 `ActionReadopt` |

**어느 것도 이 상태에서 불리지 않는다.** B8이 `release`를 건너뛰었고, 체결은 오지 않고,
운영자는 아직 행동하지 않았다.

**3판의 R3은 그 고리다** — *attempt가 비-CONFIRMED 종결에 이르면 그 intent를 가리키는
발의를 해제한다.* 연결은 이미 원장에 있다(`exit_states.pending_intent_id` →
`mutation_attempts.intent_id`). **이것이 B8 자신의 약속을 참으로 만든다** —
B8의 주석이 *"the proposal stays armed and **the resolver settles it**"*라고 쓰는데,
오늘 그 resolver가 발의까지 settle하지 않는다.

## D0. 하나의 원칙

**브로커가 이름을 준 code는 code가 판단한다. status도 message도 아니다.**

이번 사건이 그 근거를 셋 다 준다.

| 필드 | openapi 계약 | 프로덕션 실물 | 일치 |
| --- | --- | --- | --- |
| status | `422` | `409` | ✗ |
| message | `동일 종목에 반대 방향의 체결 대기 주문이 있습니다.` | `반대 포지션 미체결 주문이 존재합니다.` | ✗ |
| **code** | `opposite-pending-order-exists` | `opposite-pending-order-exists` | **✓** |

세 요청(`6GKYatiUehps5SQX`·`7d3we7ZD3dtxWTMO`·`7k5oRgmEHnoU5Vfi`)에서 같다.

**기존 코드는 이 원칙을 절반만 따른다.** `classifyRefusalBody`(`failclosed.go:223-238`)는
code 토큰(`trade_auth_required`·`fx_consent`·`funding_required`)과 **한국어 message
조각**(`"거래 인증"`·`"환전 동의"`·`"입금"`)을 **함께** 매칭한다. 위 표가 보이듯
message는 계약과 어긋날 수 있으므로 message 매칭은 같은 종류의 취약점이다.
**이 change는 새 항목을 code로만 건다.** 기존 세 항목의 message 매칭은 건드리지 않는다 —
그것을 지우면 지금 잡히던 것이 안 잡히는 방향이고, 이 change는 보수 방향만 취한다.
**그 취약점은 남는다는 사실을 여기 적는다**(침묵한 생략 아님).

## D1. R1 — 어디에 거는가

### 왜 `isDefinitiveRejection`이 아닌가

`journal.isDefinitiveRejection`(`dispatch.go:349-356`, 분기 3)에 409를 더하는 것은
**금지된 방향이다.** 409는 재생 경로의 최빈 응답이기도 하다:

> `409 request-in-progress`(openapi) → 원 요청 처리 중, 대기 후 재시도(상한 미소비)
> — `openspec/specs/order-execution/spec.md` IN_DOUBT 해소 §1

`request-in-progress`는 **원본이 실행됐을 수 있다.** 409 전체를 확정 거절로 바꾸면
그것을 "실행 안 됨"으로 확정하게 되고, 그 오류는 **살아 있는 주문을 은퇴시킨다** —
정확히 반대 방향의 위험이다. 상태 코드 표는 건드리지 않는다.

### 왜 `classifyRefusalBody`인가

`execgw.classifyMutation`(`classify.go:21-79`, 분기 7)의 순서가 이미 맞다.

| 분기 | 자리 | 순서상 의미 |
| --- | --- | --- |
| **B2** | `:32` | `policyRefusal` — 로컬 거부, 브로커 접촉 전 |
| **B3** | `:46` | `ClassifyBrokerRefusal` — **브로커가 요청을 서술하며 거절** |
| **B4** | `:49` | post-prepare 확인만 `DispatchAmbiguous`로 승격 |
| **B5** | `:64` | `statusOf` → `ClassifyHTTPMutation` — **status가 말하게 하는 마지막 수단** |

B3의 주석이 근거를 갖는다:

> *It is classified here, **before the transport tracker gets a say**, because the
> meaning comes from **the answer** and not from how far the bytes got.*

`opposite-pending-order-exists`는 정확히 그런 답이다. B3에 걸리면 B5는 **돌지 않고**,
`class = journal.DispatchRejected`가 되어 attempt가 **종결**한다.

### 무엇을 더하는가

1. `execgw`에 reason code 하나 — `ReasonOppositePendingOrder`.
   `AllReasonCodes()`(`failclosed.go:254`)에도 넣는다. 그 함수의 주석이 이유를 쓴다:
   *"these strings land in the journal and in operator alerts … a rename is a data
   migration and not a refactor."*
2. **본문의 `code` 필드를 파싱**해서 그 값이 `opposite-pending-order-exists`일 때만 건다.
   본문 통짜 substring 매칭이 **아니다**.

`refusalBody(err)`(`failclosed.go:196`)는 이미 브로커 자신의 본문만 읽는다
(*"It never falls back to err.Error()"*). 우리가 만든 문자열을 매칭할 위험이 없다.

### 왜 한 줄(`containsAny`)로는 안 되는가 — 이 delta 자신이 그것을 금지한다

`classifyRefusalBody`(`failclosed.go:223-238`)는 **본문 통짜에 대한 `strings.Contains`**다.
여기에 case 하나를 더하면 `message`에 그 문자열이 들어 있는 본문도 잡힌다.
그런데 이 change의 spec delta는 *"message 문구로 걸어서는 안 된다(SHALL NOT)"*를 쓴다.
**한 줄로는 자기 SHALL NOT을 만족할 수 없다.** tasks 2.5가 그것을 시험한다.

**기존 세 항목이 그 취약성의 증거다.** `"interactive"`는 code 토큰이 아니라 아무 본문에나
나올 수 있는 영어 단어인데 지금 `trade_auth_required`와 같은 case에 묶여 있다.
이 change는 그것을 고치지 않지만(§6 — 지우는 방향은 지금 잡히는 것을 놓친다),
**새 항목을 같은 방식으로 만들지도 않는다.**

### 본문 모양은 하나가 아니다 — 실측

| 출처 | 실물 | code 자리 | 표기 |
| --- | --- | --- | --- |
| `testdata/interactive_auth_challenge.json` | `{"code":"TRADE_AUTH_REQUIRED","message":…}` | **최상위** | UPPER_SNAKE |
| `testdata/fx_consent_required.json` | `{"code":"FX_CONSENT_REQUIRED",…}` | **최상위** | UPPER_SNAKE |
| **프로덕션 409 3건** (원장 `mutation_attempts.detail`) | `{"error":{"requestId":"7k5oRgmEHnoU5Vfi","code":"opposite-pending-order-exists","message":"반대 포지션 미체결 주문이 존재합니다."}}` | **`error` 아래** | lower-hyphen |

**두 가지가 동시에 다르다** — 자리(최상위 vs `error.`)와 표기(UPPER_SNAKE vs
lower-hyphen). 따라서 파서는 둘 다 읽고, 값 비교는 **대소문자 무시 + 전체 일치**다
(substring 아님). 이 표가 D0의 "code만이 안정 필드"를 한 겹 더 좁힌다 — **code의
*값*은 안정적이지만 그 *자리*와 *표기*는 아니다.**

### 그래서 R1은 함수 하나를 더한다

`classifyRefusalCode(body string) (ReasonCode, bool)`

- 본문을 JSON으로 읽어 `code`와 `error.code`만 본다. 다른 필드는 보지 않는다.
- JSON 파싱 실패·필드 부재·빈 값 → **분류하지 않음**(fail-closed: 모르면 종전 경로).
- `classifyRefusalBody`보다 **먼저** 부른다. 기존 세 항목의 substring 매칭은 그대로 둔다.

`internal/execgw/testdata/reason_codes.golden`는 새 code 하나만큼 늘어난다.
갱신은 `TOSSOS_UPDATE_GOLDEN=1 go test -run TestWriteReasonCodeGolden`이며 tasks 2.10이
그것을 명시한다 — 이 파일은 손으로 고치는 파일이 아니다.

### 재생 경계 — 지금 우연히 안전한 것을 계약으로 바꾼다

`classifyReplay`(`replay.go`)는 `classifyMutation`과 코드를 공유하지 않는다. 따라서
R1은 오늘 재생 분류를 오염시키지 않는다 — 그리고 그것은 **우연이 아니다**. `classifyReplay`의
default 분기가 정책을 의도적으로 적는다: *"A first dispatch would call several of these
definitive refusals; a replay may not"*(`replay.go:517-520`). **계약 공백은 여전히 실재한다** —
그 문장은 오늘의 동작을 설명할 뿐 새 code를 구속하지 않는다. 재생 attestation이 켜지는 날
이 code가 재생 응답에 붙으면 "원 요청이 접수 안 됨"으로 확정되어 **정확히 반대 방향**으로
작동한다(재생의 409는 `request-in-progress` 계열이다). spec에 SHALL NOT으로 못 박는다.

### R1이 바꾸는 관측 가능한 결과

| | 전 | 후 |
| --- | --- | --- |
| attempt 상태 | `IN_DOUBT`, `settled_at=NULL` | `FAILED_CONFIRMED`(종결) |
| `PendingAttempts` | 포함 → `checkSymbolFree` B4가 전 주문 차단 | 제외 → 차단 없음 |
| `submit`의 분기 | **B8** `:1296` — 제안 걸린 채 `return nil` | **B10** `:1304` — 알림 + `release(ProposalRefused)` → 레벨 재무장 |
| 운영자 | 아무 소리 없음 | `alertProposalRefused` |

**R1만으로는 팔리지 않는다.** 재무장된 레벨이 다음 주기에 같은 409를 받는다. R1은
*영구 동결*을 *보이고 세어지는 반복*으로 바꾼다. 파는 것은 R2다.

### R1의 소급 — 저장된 IN_DOUBT를 같은 증거로 다시 읽는다 (3판)

**위 표는 앞으로 오는 409에 대해서만 참이다.** 이미 저장된 attempt는 분류가 끝났고,
D−1이 보였듯 그 종목에서는 다음 dispatch 자체가 일어나지 않으므로 **R1이 영영 닿지 않는다.**

**그런데 판단 근거가 원장에 그대로 있다.** `mutation_attempts.detail` 실물:

```text
HTTP 409 does not prove whether the mutation executed: official: API error 409:
{"error":{"requestId":"7k5oRgmEHnoU5Vfi",
          "code":"opposite-pending-order-exists",
          "message":"반대 포지션 미체결 주문이 존재합니다."}}
```

**따라서 기동 시 1회, 저장된 `IN_DOUBT` attempt 중 본문에 확정 거절 code가 있는 것을
`FAILED_CONFIRMED`로 재분류한다.**

**안전 논거는 R1 자신과 같다.** `opposite-pending-order-exists`는 브로커의 **주문 전
검증 거절**이며 그 상태에서 주문은 접수되지 않는다. 같은 증거를 나중에 읽는 것뿐이고,
살아 있는 주문을 은퇴시키는 위험이 없다. openapi가 이 code를 422(확정 거절군)에 둔 것이
같은 판단이다.

**경계 (SHALL NOT)**:

- **code로만 판단한다.** 저장된 `detail`의 **본문 부분**을 파싱하고, 우리가 앞에 붙인
  산문(`"HTTP 409 does not prove…"`)을 매칭 대상으로 삼지 않는다
- **확정 거절 code가 없는 IN_DOUBT는 건드리지 않는다.** `request-in-progress`를 포함해
  모호한 것은 모호한 채 둔다
- **재분류는 attempt 상태만 바꾼다.** 발의 해제는 R3의 고리가 한다 — 두 쓰기를 한
  함수에 섞지 않는다
- **기동 시 1회.** 주기적으로 돌지 않는다 — 새로 생기는 409는 R1이 dispatch 시점에 잡는다

**이것이 475150·080220을 실제로 녹이는 부분이다**: 재분류 → 종결 → R3의 고리가 발의 해제
→ 다음 주기에 발의 → R2가 막고 있던 매수를 치움 → 손절 제출.

## D2. R2 — 청소가 브로커를 본다

> **4판**: 이 절의 「넓히는 방식」·「배선」·「파싱」·「§0.4 — 스냅샷」·「감사」 는 **엔진 밖 주문 취소**를 전제로 한 3판 설계다.
> 4판 R2 는 엔진 귀속 주문만 다루므로(D−2.4) 그 부분은 **사용자 결정 대기**이며 구현 대상이 아니다. 남는 것은 빈 가격 규칙,
> 연속 실패의 새 트리거(D−2.7), 기존 흡수 규칙이다.

### 지금의 눈

`engine.clearTheSymbol`(`exitloop.go:1334-1392`, 분기 9)

| 분기 | 자리 | 역할 |
| --- | --- | --- |
| **B2** | `:1341` | `for _, order := range live` — 목록 순회 |
| **B3** | `:1343` | `if !buy && !withPending { continue }` — **매수(`buy`)는 항상 치우고, `withPending`이 필요한 것은 자기 방향(매도)이다** |
| **B6** | `:1379` | `if err != nil \|\| out.State != journal.StateConfirmed` — 취소 확정 실패는 `clear=false` |
| **B7** | `:1383` | `if !clear` — 하나라도 못 치우면 제출하지 않는다 |
| **B8** | `:1386` | `if withPending && m.state.Pending()` — 마지막에 제안 해제 |

`live`의 원천 `Journal.LiveOrdersForSymbol`(`fills.go:1849-1927`, 분기 7)은
`mutation_attempts JOIN intents`다. **엔진 밖 주문은 0행이다.**

### 넓히는 방식

`live`를 **저널 ∪ 브로커 미체결**로 만든다.

- **저널에 있는 주문**: 지금 그대로. lineage 해소(`fills.go`의 주석 —
  *"the official API answers a modify with a new order number"*)를 유지한다.
- **저널에 없는 주문**: 브로커가 보고한 `orderId`로 취소한다. lineage를 만들 수 없으므로
  **정정(AMEND) 대상이 아니고 취소만 한다.**
- **양쪽에 있는 주문**: `orderId`로 dedup. 이것은 새 규칙이 아니라 IN_DOUBT 조회 대조가
  이미 쓰는 규칙이다(`spec order-execution` §2 — *"유일성 판정 전에 orderId 기준 dedup"*).

### 배선 — 「새 API 표면 없음」은 브로커 API에 대해서만 참이다

1판 tasks 3.9가 *"새 API 표면 없음"*이라 썼다. **엔진 API에 대해서는 거짓이다.**
`OrderPager`(`indoubt.go:70-74`)는 `reconcile`·`filldetect`·`flatten`에만 배선돼 있고
`Context.ExitObserver`(`exitwiring.go:319-347`)가 채우는 필드에 **없다.** 필요한 자리 넷:

| # | 자리 | 무엇을 |
| --- | --- | --- |
| 1 | `ExitObserverOptions`(`exitloop.go:166-223`) | `Orders execgw.OrderPager` 필드 하나 |
| 2 | `NewExitObserver`의 거부 switch(`exitloop.go:263-282`) | **nil을 허용한다** — 필수로 하면 미배선 빌드가 기동조차 못 한다. nil이면 브로커를 보지 않고 **저널만으로 오늘과 똑같이** 동작한다 |
| 3 | `Context.ExitObserver`(`exitwiring.go:338-346`) | `opts.Orders == nil`이면 `execgw.OfficialOrders{Client: c.Official}` — `Floor`·`Names`·`Alerts`가 이미 쓰는 「nil이면 채운다」 형태 그대로 |
| 4 | `cmd/tossctl/engine.go:346` | 구성 리터럴은 **손대지 않는다**(3이 채우므로) |

**nil 허용은 이 change의 토글이 아니다.** 토글은 도입하지 않는다(tasks 7.4). nil 경로는
`ExitObserver`를 직접 만드는 **테스트**의 경로이고, 프로덕션 경로인 `Context.ExitObserver`는
항상 채운다. 그 차이를 tasks 3.11이 시험한다.

### ~~파싱 — 기존 함수 둘 다 모자란다~~ — **열거가 불완전했다 (3판)**

> **아래 표는 유효하지만 완전하지 않다.** 세 번째 후보 `filldetect.parseSnapshot`을
> 평가하지 않았고, 그것이 7필드 중 5개를 이미 준다. **결론은 다음 절(§0.4)이 대체한다** —
> 새 파서가 아니라 기존 것의 확장이다. 이 표는 「왜 `brokerstate`와 `official`로는
> 안 되는가」의 근거로만 남긴다.

#### (1·2판) 두 후보의 한계 — 유효

`clearTheSymbol`은 한 주문에서 **7개 필드**를 쓴다 — `Side`(B3 술어), `OrderID`,
`Symbol`, `Market`, `Quantity`, `Price`, `Currency`(`CancelRequest.Order`).

| 후보 | 주는 것 | 모자란 것 |
| --- | --- | --- |
| `brokerstate.ParseOfficialOrder`(`derive.go:645`) | `OrderID`·`RawStatus`·`Canceled`·`CanceledAt`·`Quantity`·`FilledQuantity` | **`Side`·`Symbol`·`Market`·`Price`·`Currency` 전부 없다.** `officialOrderPayload`(`derive.go:623-631`)가 6필드짜리 의도적 부분 미러다 |
| `official.Client.Orders()` → `domain.Order` | `Side`·`Symbol`·`Market`·`Price`는 있다 | **첫 페이지만 남긴다**(`indoubt.go:726` — *"Upstream's Orders() keeps only the first page"*). 부분 목록은 목록이 아니다 |

따라서 **새 파서 하나를 만든다** — `execgw`에 `ParseWorkingOrder(json.RawMessage)`.
`ScanOrders`(`indoubt.go:730`)가 이미 raw JSON을 완주해서 주므로 페이지 문제는 없고,
새로 필요한 것은 그 raw에서 위 7필드를 읽는 일뿐이다. 이것이 **새 엔진 API 표면**이며
tasks 3.9의 1판 문장을 정정한다.

**status 그룹**: `OrderQuery.Status = "OPEN"` 하나만 쓴다.
openapi가 그것을 정의한다(`indoubt.go:394-397`): *"OPEN: 진행 중 주문 그룹 —
`orders[].status ∈ {PENDING, PARTIAL_FILLED, PENDING_CANCEL, PENDING_REPLACE}`"*.
CLOSED는 조회하지 않는다 — 치울 대상은 미체결뿐이고, CLOSED를 더하면 조회가 두 배가
되면서 얻는 것이 없다. `PENDING_CANCEL`이 목록에 있으면 취소가 이미 진행 중이므로
새 취소를 내지 않고 `clear=false`로 둔다(그 주문은 아직 책 위에 있다).

**종목 필터는 서버측이다** — `OfficialOrders.OrdersPageRaw`(`orders_source.go:47-53`)가
`OrderQuery.Symbol`을 `official.OrdersFilter.Symbol`로 넘기고 그것이 쿼리 파라미터
`symbol`이 된다(`orders_reads.go:93-95`). 계좌 전체를 끌어와 걸러내는 것이 아니다.

### §0.4 — **동기 조회를 넣지 않는다. 이미 있는 스냅샷을 읽는다** (3판)

2판은 손절 경로에 **2초 타임아웃의 동기 브로커 왕복**을 넣고 그것을 §0.3 비용으로
받아들였다. **2라운드가 그 비용이 불필요함을 보였다**(`review.md` §2.8).

**`filldetect.Detector`가 이미 3초마다 계정 전체 OPEN 목록을 완주한다:**

| 사실 | 자리 |
| --- | --- |
| 프로덕션에서 **무조건** 돈다 | `cmd/tossctl/engine.go:391-396` — `{Name: "filldetect", Run: detector.Run, …}` |
| OPEN 목록을 완주한다 | `Detector.collect`(`detect.go:363-…`, 분기 19) **B1** `:365` 직전 — `ScanOrders(ctx, d.Orders, OrderQuery{Status: statusOpen}, cfg.MaxPages)` |
| 주기 3초 | `detect.go:128` — `PollInterval: 3 * time.Second` |
| 원천이 **같다** | `engine.go:420` — `Orders: execgw.OfficialOrders{Client: ectx.Official}` |

**2판이 넣으려던 것과 같은 API, 같은 pager, 같은 호출이다.**

**빈도 실측이 그 낭비를 못 박는다.** `clearTheSymbol`은 모든 full exit 직전에 돌고
주기는 5초(`exitloop.go:97`)다. proposal 자신의 측정: 272210이
`STOP_LOSS_LADDER → PROPOSAL_CANCELLED`를 **2h54m 동안 1931회, 중앙값 5.0초**로 반복했다.
**2판대로면 그것이 전부 브로커 OPEN 조회다** — 5종목이 발의 중이면 ~1 req/s로,
detector의 ~0.33 req/s 위에 **약 4배**다.

**3판의 결정: `clearTheSymbol`은 detector의 마지막 OPEN 스냅샷을 읽는다.**

| | 2판 (동기 조회) | 3판 (스냅샷) |
| --- | --- | --- |
| §0.3 손절 지연 | **최대 2초** | **~0** (메모리 읽기) |
| §0.4 새 호출 | 5초마다 종목당 1회 | **0** |
| 데이터 신선도 | 즉시 | **최대 3초** |
| 종목 필터 | 서버측 | **클라이언트측** (스냅샷은 계정 전체) |
| 목록 못 얻을 때 | 조회 실패·타임아웃 | **스냅샷 부재·상한 초과 노후** |

**신선도 상한은 5초로 둔다** — detector 주기 3초의 여유 1회분이다. 그보다 오래된
스냅샷은 「목록을 얻지 못했다」로 취급한다(아래 절).

**파서도 다시 만들지 않는다.** 2판의 「기존 함수 둘 다 모자란다」 표가
`filldetect.parseSnapshot`을 평가하지 않았다. 그 `Snapshot`(`detect.go:194-224`)은
`OrderID`·`Symbol`·`Market`·`Side`·`Quantity`를 이미 준다 — `clearTheSymbol`이 쓰는
**7필드 중 5개**다. 없는 것은 주문 `Price`와 `Currency`뿐이고(있는 `AveragePrice`는
체결가다), **그것은 확장이 답이지 세 번째 파서가 답이 아니다.**

**상태 판정도 이미 있다** — `PENDING_CANCEL`은 `brokerstate.StateCancelPending`
(`derive.go:421`)으로 모델링돼 있고 `Snapshot.Derived`가 그것을 싣는다. 문자열 비교를
새로 쓰지 않는다.

**배선 선례도 바로잡는다.** 2판 표는 `Context.ExitObserver`가 nil-fill하고
`cmd/tossctl/engine.go`는 손대지 않는다고 정했다. **detector 파생 의존성의 기존 선례는
정확히 거기서 주입된다** — `engine.go:349` `SLO: detectorPressure{detector: detector}`이고
`Context.ExitObserver`는 `opts.SLO`를 건드리지 않는다. 3판은 그 선례를 따라
**`engine.go`에서 주입한다.**

`exitwiring.go:313-317`의 doc comment는 **stale이다** —
*"this build constructs no fill detector: there is no production polling loop to defer to
yet"*라고 쓰는데 빌드는 만들고 넘긴다. 3판이 그 주석도 고친다.


### 브로커 오류는 `clearTheSymbol` 안에서 흡수한다

1판은 오류를 `record`로 반환했다. `record`(`exitloop.go:1141-1144`)는
`cleared, err := o.clearTheSymbol(...)` 다음 `if err != nil { return err }`다.
**브로커 두절 한 번이 `RecordExitJudgementResult` 전에 판정을 통째로 중단시킨다** —
워터마크도 기준선도 전진하지 않고 `noteDelay`도 울리지 않는다. `exitloop.go:1113-1116`이
명시한 계약(*"A clear that does not complete does not stop the judgement"*)보다
**엄격히 나쁘다.**

그래서 브로커 목록 조회의 실패·타임아웃·페이지 초과는 **`clearTheSymbol` 내부에서**
흡수한다:

1. 저널분 청소는 **그대로 수행한다** (오늘과 동일).
2. 브로커분을 보지 못했다는 사실을 `clear = false`로 만든다.
3. `record`의 기존 `!cleared` 가지(`exitloop.go:1145-1148`)가 받는다 —
   `noteDelay` + `ArmSuppressedWorkingOrder`. **알림이 울리고 판정은 계속된다.**

즉 새 실패 모드를 만들지 않고 **이미 있는 실패 모드로 떨어뜨린다.**

**spec의 SHALL을 그 형태로 고친다.** 1판 delta는 *"브로커 미체결을 포함해야 한다(SHALL)"*로
폴백을 허용하지 않았고 tasks 3.7은 *"조회 실패해도 저널분 청소는 진행"*을 요구했다 —
**둘은 동시에 참일 수 없었다.** 새 문장은 「목록을 보지 못하면 제출하지 않는다」다.

### 자기 방향 미체결의 부재 확인 — **3판에서 철회했다**

2판은 「`withPending`과 무관하게 자기 방향(SELL) 미체결이 있으면 제출하지 않는다」를
넣었다. **2라운드가 그것을 차단했고 근거가 맞다**(`review.md` §2.2).

**철회 사유 1 — 막으려던 초과 매도는 이미 막혀 있다.**
`armExitProposalTx`(`apply_hook.go:655-676`, 분기 4) **B3** `:666`:

```go
if strings.TrimSpace(action.String) != "" {
    return fmt.Errorf("%w: %s holds %s", ErrProposalPending, positionID, action.String)
}
```

주석이 논거를 쓴다 — *"A second proposal while one is outstanding is refused rather than
overwritten."* **엔진 자신의 두 번째 매도는 여기서 거부된다.**

**철회 사유 2 — 남는 효과가 「보호를 영구 withhold」뿐이다.**
이 검사가 추가로 잡는 유일한 경우는 **사용자가 앱에 넣은 매도**다. R1이 발의를 해제한
다음 주기부터 `CancelPendingFirst = false`이므로(`ladder.go:447`) 청소는
`withPending=false`로 불리고, R2는 **반대 방향만 취소한다**(B3 `:1343`).
따라서 그 매도는 취소되지 않은 채 남고, 검사는 **매 주기 제출을 막는다.**

**사용자의 지정가 매도 하나가 그 종목의 손절을 영구히 막는다.** 오늘은 그것을 건너뛰고
제출한다. 이 계정의 사건이 정확히 「앱에서 직접 넣은 주문」이었다.

**따라서 R2는 반대 방향 취소만 한다.** 자기 방향은 오늘과 같이 `withPending`일 때만 본다.

### 취소할 수 없는 주문이 손절을 영구 보류시키는 문제 (차단 8) — 결정

오늘 `clear=false`의 후보는 **엔진 자신의 확정 주문뿐**이다. R2 이후 후보는 브로커가
보고하는 무엇이든이고, 엔진이 영원히 취소할 수 없는 주문이 손절을 **영구 보류**시킬 수
있다. 구체 경로 둘:

- (a) 목록 조회와 취소 사이의 체결 → 취소 거절 → `clear=false`
- (b) `floatOf`(`exitloop.go:1673-1679`)가 `ParseFloat("")`을 거절하므로 **가격 없는
  주문**(시장가)이 목록에 오면 `clear=false`가 고정된다

**결정: 빈 가격은 실패가 아니다 — 그리고 3판은 이것을 저널분에도 적용한다.**
2판은 *"`floatOf` 거절 경로는 저널분에만 남긴다"*고 했다. **2라운드가 a087과의 충돌을
잡았다**(`review.md` §2.4): `LiveOrdersForSymbol`은 `coalesce(i.price,'') AS price`
(`fills.go:1859`)이고 **a087이 보호 청산을 시장가로 바꾸면 저널분에 빈 가격 행이 생긴다** —
2판이 실패를 남겨 둔 바로 그쪽이다.

따라서 빈 `price`는 **양쪽 모두** 0으로 읽는다. `execgw.OrderRef.Price`는 취소 요청의
식별에 쓰이지 않으므로 안전하다. **이것으로 a087 선후 관계 제약이 사라진다.**

**결정: (a)는 보류하되 무한 보류하지 않는다.** 같은 종목에서 청소가 **연속 3회**
`clear=false`로 끝나면 `EventExitLiquidationDelayed`(**이미 critical**, `event.go:336`)를 **새 트리거로 한 번 더** 낸다 — 기존 30초 타이머는 그대로다(4판 D−2.7).
그 뒤에도 자동으로 제출하지는 않는다 — 「못 치우면 팔지 않는다」(B7)를 뒤집는 것은
초과 매도 방향이고 §6에 걸린다. **바꾸는 것은 침묵의 길이지 규칙이 아니다.**

**단, 엔진 자신이 방금 낸 취소가 정산 중인 것(`PENDING_CANCEL`)은 그 카운터에서
제외한다** — 그것은 「치우지 못하는 중」이 아니라 「치우는 중」이다. 5초 주기에서
15초 넘게 걸리는 정상 취소가 critical을 만드는 것은 거짓 경보이고, critical 전달 실패는
`ENTRY_BLOCKED`까지 간다(`notifier.go:216-218`). 상태 판정은 문자열 비교가 아니라
`brokerstate.StateCancelPending`(`derive.go:421`)을 쓴다.


### 경계 — 새 권한이 아니라 기존 권한의 눈

| 조건 | 이유 |
| --- | --- |
| `clearTheSymbol`이 불릴 때만 | `record` **B5** `:1223`(base 번호 B3)이 이미 게이트다 — `orderable && (CancelPendingFirst \|\| isFullExit)`. 그 앞에 a111 **B1** `:1180-1182`(쓸 수 있는 시세 없음 → 반환) |
| 같은 계좌·시장·종목 | 충돌의 정의. 종목 필터는 서버측 |
| 반대 방향(`buy`)은 취소, 자기 방향(SELL)은 `withPending` 일 때만 치운다 — **4판: 엔진 귀속 주문만**(D−2.4). 3판이 철회한 부재 확인은 없다 | `clearTheSymbol` **B3** `:1449`(base `:1343`) + 위 절 |
| 취소만. 신규·정정 없음 | 청소는 노출을 늘리지 않는다 |
| 취소 확정 실패 → `clear=false` | **B6** `:1379`의 기존 규칙 그대로. **못 치우면 팔지 않는다** |
| 브로커 조회 실패 → `clear=false` | 새 실패 모드가 아니라 기존 모드로의 낙하 |

**노출을 늘리는 주문은 이 경로로 나가지 않는다.** `raisesExposure`
(`gateway.go:377` — `side == "buy"`)는 취소(`gateway.go:416` — `false`)에 적용되지 않고,
청소는 취소만 낸다.

### 감사

엔진이 내지 않은 주문을 취소하는 것은 **기록되어야 한다.** 취소 intent에 그 사실을
남긴다 — 저널에 원본 intent가 없는 주문을 치웠다는 것, 그리고 그것이 어떤 보호 청산을
위한 것이었는지. `CLAUDE.md`「runtime config 변경은 audit로 추적 가능해야 한다」의
같은 이유가 여기에도 적용된다.

## D3. R3 — **종결된 attempt가 발의를 푼다** (3판에서 전면 교체)

### 무엇을 교체했는가

1·2판의 R3은 「세션 중 IN_DOUBT 해소」였다. **2라운드가 그것이 동결을 풀지 못함을
보였다**(`review.md` §2.1) — `Resolve`는 `mutation_attempts`에만 쓰고 `exit_states`는
건드리지 않는다. attempt를 park시켜도 `pending_action`은 무장된 채이고 사다리는 계속
억제한다.

**3판의 R3은 그 빠진 고리다.**

### 고리 하나

```text
attempt 가 비-CONFIRMED 종결에 이른다
  (FAILED_CONFIRMED · NOT_DISPATCHED · UNRESOLVED_IN_DOUBT)
        │
        ↓
그 attempt 의 intent_id 로 exit_states 를 찾는다
  exit_states.pending_intent_id = attempt.intent_id     ← 이미 있는 연결
        │
        ↓
Journal.ResolveExitProposal(positionID, ProposalRefused)
  B8  :842  이미 비었으면 무동작 (멱등)
  B9  :846  pending_action / pending_level / pending_intent_id → NULL
  B10 :852  LADDER 면 rung 되돌림
  B13 :859  exit_events 1행
        │
        ↓
다음 관측: EvaluateLadder B26 :441 이 억제하지 않는다 → 발의가 난다
        └─→ record :1117 게이트 열림 → clearTheSymbol(R2) → submit(R1)
```

### 안전 확인 — `ResolveExitProposal`이 무엇을 쓰는가 (분기 14, AST)

| 분기 | 자리 | 무엇을 | 안전 |
| --- | --- | --- | --- |
| **B8** | `:842` | `pending_action`이 비면 **무동작 반환** | **멱등** — 중복 호출이 무해하다 |
| **B9** | `:846` | 세 컬럼 → NULL | 해동의 실제 지점 |
| **B10·B11** | `:852`·`:853` | LADDER면 `RungIndex(pending_level)` | 음수 label은 거부된다(`ladder.go:536-538`) → 되돌림 없음 |
| **B12** | `:854` | `rollBackRungTx(rung-1)` | **`active_rung`만 쓴다**(`exit_state.go:980-987`). **손절 가격은 건드리지 않는다** |
| **B13** | `:859` | `exit_events` 1행 | 감사 흔적 |

**§6 확인**: 이 경로에서 `entry_price`·`initial_stop`·`baseline_price` 중 어느 것도
쓰이지 않는다. **손절 가격이 움직이지 않는다.**

원장 실측이 두 경우를 다 보인다 — 475150은 `pending_level="0"` → rung 0에서
되돌림, 080220은 `pending_level="-1"` → `RungIndex` 거부 → 되돌림 없이 해제.
**둘 다 손절 가격은 그대로다.**

### 왜 B8의 안전 논거를 깨지 않는가

`submit` B8의 주석:

> *The order may exist. The proposal stays armed and **the resolver settles it**;
> releasing here would let the next observation submit a second sell on top of one that
> may already be live.*

**그 논거는 「종결 전에 풀면 안 된다」이지 「영원히 풀지 말라」가 아니다.**
3판은 **종결된 뒤에만** 푼다:

| attempt 종결 상태 | 의미 | 발의 |
| --- | --- | --- |
| `CONFIRMED` | 주문이 **실제로 나갔다** | **해제하지 않는다** — 체결 경로가 처리한다 |
| `FAILED_CONFIRMED` | 접수되지 **않았음이 확정** | **해제한다** — 살아 있는 주문이 없다 |
| `NOT_DISPATCHED` | 전송조차 안 됨 | **해제한다** |
| `UNRESOLVED_IN_DOUBT` | **모른다** | **아래 참조** |

**`UNRESOLVED_IN_DOUBT`가 유일한 판단 지점이다.** B8의 논거대로면 모르는 상태에서
푸는 것은 초과 매도 위험이다. **그러나 오늘의 대안은 「영원히 무보호」다.**

**3판의 결정: park에서도 해제한다. 근거 셋.**

1. **초과 매도의 1차 방벽은 발의가 아니라 `armExitProposalTx`다**(`:666`) — 그리고
   그것은 발의를 해제해도 다음 발의가 무장될 때 다시 검사한다. 해제는 「두 번째 발의를
   허용」하는 것이 아니라 「첫 발의를 끝낸 것으로 기록」하는 것이다.
2. **park는 그 자체로 계정 전역 진입을 막는다**(`indoubt.go:379-382`). 즉 그 상태는
   이미 사람의 개입을 요구하는 상태이고, 그 위에 **손절만 막아 두는 것**은 방향이 거꾸로다.
3. **§0.3이 이것을 결정한다.** 「모르는 매도가 있을 수 있다」와 「확실히 보호가 없다」
   중 후자가 더 나쁘다. 안전 불변식 §4가 *"손절·비상 청산의 즉시성을 약화하거나 지연하지
   않는다"*이고, 영구 억제는 무한 지연이다.

**단, 이 결정은 R1의 소급과 짝일 때만 이 사건에 적용된다.** 475150·080220은
소급 재분류로 `FAILED_CONFIRMED`가 되므로 **위 표의 둘째 행**으로 들어간다 —
`UNRESOLVED_IN_DOUBT` 판단에 기대지 않는다. **park 해제는 일반 규칙이고, 이 사건의
해동은 그것에 의존하지 않는다.**

### 어디에 배선하는가

**`reconcile.Run`·`Journal.RecoverPending`을 재사용하지 않는다(SHALL NOT).**
`RecoverPending`(`journal/recovery.go:86-125`, 분기 10)은 세션 중에 부르면 원장을
위조한다 — **B4** `:95` `StateRecorded` → *"found at startup with no dispatch recorded"*로
**종결**시키고, **B6** `:103` `StateDispatchStarted` → *"process stopped after dispatch
started"*를 쓴다. 둘 다 세션 중에는 거짓이다.

**3판이 필요로 하는 것은 그 함수가 아니다.** attempt를 종결시키는 자리는 이미 있다:

| 자리 | 언제 |
| --- | --- |
| `classifyMutation` → `submit` | dispatch 시점 (R1이 여기를 고친다) |
| `Resolver.Resolve` | 재시작 복구 · (별도 change의) 세션 중 해소 |
| R1의 소급 재분류 | 기동 시 1회 (3판이 더한다) |

**R3의 고리는 그 셋 뒤에 붙는 하나의 후처리다** — attempt가 종결되면 그 intent를 가리키는
발의를 찾아 해제한다. **새 루프도, 새 주기도, 새 브로커 조회도 없다.**

### 세션 중 해소는 **별도 change로 분리한다**

1·2판의 R3이 가진 나머지(주기 · `Context.Resolver` 배선 · `resolveCancel`의 `r.Order`
요구 · 계정 전역 park의 운영 결과 · §0.4 계측)는 **이 사건의 인과 경로에 없다.**
proposal 자신이 *"R3만: 이 사건의 원인이 남아 매 주기 다시 언다"*라고 인정했다.

`issues.md`가 선행 조건을 기록한다. **침묵한 생략이 아니다.**

## D4. R4 — baseline — **철회했다 (1라운드 차단 2·3)**

> **아래 절은 진단으로만 남긴다.** 「무엇을 채우는가」의 처방은 **폐기**했다.
> 철회 사유는 이 절 끝의 「왜 철회했는가」에 있고, 판정 원문은 `review.md` §1.3·§1.7,
> 후속 조건은 `tasks.md` §5와 `issues.md`에 있다.
>
> **진단은 유효하다** — baseline이 비어 있어 부재가 증명되지 않는다는 것은 사실이고
> D3의 「이득이 무엇인지 정확히」가 그 사실에 의존한다. 틀린 것은 **그것을 지금 채우자**는
> 처방이다.

### 지금 무슨 일이 일어나는가

`gateway.go:1044` `Notes: EncodeBaseline(plan.baseline)`,
`plan.baseline`은 `gateway.go:378` `baseline: req.Baseline`.

**`Baseline`을 넘기는 호출자가 없다.** `exitloop.go:1281-1286`에도
`tracer.go:442-448`에도 그 필드가 없다. 따라서 `notes`는 항상 빈 문자열이고,
`absenceCorroborated`(`indoubt.go:445-…`)의

```go
baseline, ok := DecodeBaseline(intent.Notes)
if !ok {
    return false, "absence cannot be corroborated: the mutation was submitted without a pre-dispatch account baseline"
}
```

가 **항상** 실패한다. 원장이 그것을 확인한다 — 얼어붙은 세 attempt의 `notes` 전부 0바이트.

`indoubt.go:93-94`가 대가를 명시한다:

> *its absence can never be proven, and it will park as UNRESOLVED_IN_DOUBT.
> **That is the documented price of omitting it.***

### 이것도 적합성 수정이다

spec은 부재 판정을 이미 요구한다:

> **3. 부재 판정**: 최소 관찰 기간에 걸친 연속 N회(기본 3회) 안정화 조회 +
> **매수가능금액·보유수량 delta 교차 확인** 후에만.

delta 교차 확인은 **사전 기준선이 있어야 성립한다.** baseline이 없으면 그 SHALL은
만족될 수 없다. 코드는 필드를 갖고 있고, 부르는 쪽이 안 채웠을 뿐이다.

### ~~무엇을 채우는가~~ — 왜 철회했는가

1판은 *"보호 청산의 `PlaceRequest`에 결정 시점의 `Baseline`을 싣는다 — 이미 관측 주기가
읽어 둔 값에서"*라고 썼다. **그 값이 없다.**

**차단 3 — 값의 원천이 소스에 없다.** `ExitObserverOptions`(`exitloop.go:166-223`)에
계정 상태를 읽는 필드는 **하나도 없다**. `Prices`는 시세이고 `Floor`는 대사 하한이다.
관측 주기는 매수가능금액도 보유수량도 읽지 않으므로 「이미 읽어 둔 값」이 존재하지 않는다.
채우려면 손절 경로에 **새 계정 호출**을 넣어야 하고, 그것은 같은 문서의 D6이
**§0.3을 근거로 금지한 바로 그것**이다.

**차단 2 — 억지로 채우면 살아 있는 매도를 은퇴시킨다.** `absenceCorroborated`
(`indoubt.go:486-500`)의 증거 모델은 **매수 예약 모델**이다:

```go
notional := intentNotional(intent)
bpDelta := buyingPowerNow - baseline.BuyingPower
if notional > 0 && bpDelta < 0 && math.Abs(bpDelta) >= notional*0.5 {
    return false, "... the buying power dropped ... consistent with this order having been accepted"
}
return true, "the holding and the buying power are unchanged from the pre-dispatch baseline"
```

**매도는 매수가능금액을 줄이지 않는다.** 접수된 매도의 매수가능금액 delta는 0이므로
위 `if`가 걸리지 않고 함수는 `true`("부재가 확증됨")를 **반환한다** — 그 매도가 브로커에
살아 있어도. 보호 청산은 전부 매도다. **부재 확증의 증거 모델은 매도에 대해 아무것도
증명하지 못하며, 그 상태에서 baseline을 공급하는 것은 「모름」을 「없음」으로 바꾸는
것이다.**

오늘의 「항상 park」는 그 무지의 **안전측 표현**이고, R4는 그것을 제거한다. §6(보수
방향만)에 걸린다.

**남기는 것**: 매도용 부재 증거 모델(예: 체결 이벤트 부재 + 목록 완주의 결합)이
별도 change의 **선행 조건**이다. baseline은 그 모델이 생긴 뒤에 공급한다.
`tasks.md` 5.3이 `issues.md` 기록을 요구한다.

## D5. 셋의 상호작용 (3판)

```text
[기동 시 1회] R1 소급 — 저장된 IN_DOUBT 중 확정 거절 code를 가진 것
      └─→ FAILED_CONFIRMED 로 재분류        ← 475150 · 080220 이 여기서 종결
                    │
                    ↓
[R3의 고리] attempt 가 비-CONFIRMED 종결 → pending_intent_id 로 exit_states 를 찾아
      Journal.ResolveExitProposal(ProposalRefused)
        B9 :846  pending_action / pending_level / pending_intent_id → NULL
        B12      LADDER면 active_rung 만 되돌림 (손절 가격 불변)
                    │
                    ↓
[다음 관측 5초 뒤] EvaluateLadder B26 :441 이 더는 억제하지 않는다
      → 발의 발생 → record :1117 게이트 열림
                    │
                    ↓
[R2] clearTheSymbol — detector 의 OPEN 스냅샷(≤3초)을 읽는다
        ├─ 반대 방향(BUY) → 취소          ← 막고 있던 앱 매수가 여기서 치워진다
        ├─ 자기 방향(SELL) → 오늘과 같이 withPending 일 때만
        ├─ PENDING_CANCEL → clear=false (연속 3회 카운터에서는 제외)
        └─ 스냅샷 부재·5초 초과 노후 → clear=false → record :1145 noteDelay
                    │
                    ↓
[R1] submit — 그래도 409 가 오면 code 로 확정 거절 → release(ProposalRefused)
      → 다음 주기 재무장 (영구 동결이 아니라 보이고 세어지는 반복)
                    │
                    ↓
                손절 제출 → 체결
```

**셋이 각각 다른 자리를 막는다:**

| | 없으면 |
| --- | --- |
| **R1 소급** | 얼어붙은 attempt가 종결되지 않아 R3의 고리가 걸릴 대상이 없다 |
| **R3의 고리** | attempt가 종결돼도 발의가 무장된 채라 사다리가 계속 억제한다 |
| **R2** | 발의가 나도 브로커의 매수가 그대로라 같은 409를 다시 받는다 |
| **R1 (전방)** | 그 409가 다시 IN_DOUBT가 되어 **같은 동결이 재발한다** |

## D6. 무엇을 하지 않는가

| | 결정 | 근거 |
| --- | --- | --- |
| `isDefinitiveRejection`에 409 추가 | **안 한다** | `request-in-progress`를 확정으로 만들어 살아 있는 주문을 은퇴시킨다 |
| `submit` B8이 즉시 `release`하게 고치기 | **안 한다** | B8의 논거는 옳다 — **종결 전에** 풀면 초과 매도다. 3판은 **종결 뒤에** 푼다(D3) |
| `checkSymbolFree` 미정산 루프에 위험 비증가 면제 | **안 한다** | spec의 SHALL이고 archive `2026-07-26-extend-execution-contract/design.md:63`이 이미 검토·폐기 |
| 재생 attestation 플래그 켜기 | **안 한다** | `[미측정 — 2b 전 비활성]` |
| `classifyRefusalBody`의 기존 message 매칭 제거 | **안 한다** | 지금 잡히는 것이 안 잡히는 방향(D0) |
| **손절 경로에 동기 브로커 조회 추가** | **안 한다 (3판에서 뒤집음)** | detector가 3초마다 같은 목록을 이미 읽는다(§0.4). 2판은 2초를 §0.3 비용으로 받아들였고 **그 비용이 불필요했다** |
| **자기 방향 미체결의 부재 확인** | **안 한다 (3판에서 철회)** | 초과 매도는 `armExitProposalTx` `:666`이 이미 막고, 남는 효과는 **보호의 영구 withhold**뿐 |
| `Journal.RecoverPending`을 세션 중에 호출 | **안 한다** | `RECORDED`를 종결시키고 `DISPATCH_STARTED`에 지어낸 사유를 쓴다(D3) |
| 브로커 조회 오류를 `record`로 반환 | **안 한다** | 브로커 두절 한 번이 판정 전체를 중단시킨다 |
| 「못 치우면 팔지 않는다」(B7)를 N회 후 뒤집기 | **안 한다** | 초과 매도 방향. 바꾸는 것은 침묵의 길이뿐 |
| **세션 중 IN_DOUBT 해소 루프** | **분리한다 (3판)** | 이 사건의 인과 경로에 없다. 주기·`Context.Resolver` 배선·§0.4 계측을 자기 change로 |
| `entry_price`·`initial_stop`·`baseline_price` 쓰기 | **안 한다** | 3판의 어느 경로도 손절 가격을 움직이지 않는다(§6) |

## D7. 실패 모드 재검토 (3판)

| 우려 | 답 |
| --- | --- |
| R1 소급이 실제로 실행된 주문을 "안 됐다"고 확정하면? | 이 code는 **주문 전 검증 거절**이다. 브로커가 반대 주문의 존재를 이유로 거절했고 그 상태에서 주문은 접수되지 않는다. openapi가 422(확정 거절군)에 둔 것이 같은 판단이다. **확정 거절 code가 없는 IN_DOUBT는 건드리지 않는다** |
| R1 소급이 우리가 쓴 산문을 매칭하면? | 저장된 `detail`의 **본문 부분**만 JSON으로 파싱한다. `"HTTP 409 does not prove…"`는 매칭 대상이 아니다 |
| 발의 해제가 초과 매도를 만들면? | 해제는 「두 번째 발의를 허용」이 아니라 「첫 발의를 끝냈다고 기록」이다. 다음 발의가 무장될 때 `armExitProposalTx` `:666`이 다시 검사한다 |
| 발의 해제가 손절 가격을 움직이면? | **움직이지 않는다.** `rollBackRungTx`는 `active_rung`만 쓴다(`exit_state.go:980-987`). AST 분기표로 고정한다(tasks 4.5) |
| 해제가 중복 호출되면? | **B8** `:842`가 멱등을 보장한다 — 이미 비었으면 무동작 반환 |
| 스냅샷이 오래됐는데 그것으로 판단하면? | 5초 상한을 넘으면 **「목록을 얻지 못했다」로 취급**해 `clear=false`다. detector 주기 3초의 여유 1회분이다 |
| R2가 사용자의 의도적 매수를 지우면? | 지운다. 그것이 이 change의 요구다 — 보호 청산이 우선한다. 감사에 남긴다 |
| 정상 취소가 정산 중인데 3회 카운터가 critical을 울리면? | `PENDING_CANCEL`은 카운터에서 제외한다. 판정은 `brokerstate.StateCancelPending`으로 한다 |
| a087이 먼저 오면? | **제약이 사라졌다.** 3판은 빈 가격을 **저널분에도** 0으로 읽는다 — 2판이 남겼던 충돌이 없다 |
| RATCHET 포지션도 같은 동결을 겪는가? | 그렇다. `EvaluateRatchet` **B17** `:423`이 LADDER B26과 같은 억제다. **해동 경로가 두 정책 모두에 적용된다** |
