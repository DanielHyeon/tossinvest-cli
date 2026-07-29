# 실계좌 측정 기록 — verify-execution-capability

> 근거: `~/.local/share/tossos/capability-verify.jsonl` (계좌 `*******5921`, digest는 해당 파일의 step 레코드 참조).
> 규율: 브로커 주장은 실측 근거 또는 [미측정] 태그를 동반한다.

## 1차 실행 (2026-07-26, 일요일 — 콘솔 경유)

runs: `run-FOVHBDARTNFK3RKD`(21:31 KST) → `run-WE3EF3ZOHDCNBWGT`(22:46, 승인 거부) → `run-GTXLBOTXQIIMWORW`(22:48, 승인 통과·9단계 기록).

### 실측으로 확정된 사실

| # | 사실 | 근거 |
|---|---|---|
| M1 | **`order-hours-closed`**: 휴장일(일요일) POST /api/v1/orders는 HTTP 422, `code: "order-hours-closed"`, `message: "주문가능일이 아닙니다."` — 주문 오류 분류표에 추가할 실측 코드 (2.1 입력) | order-cancel·order-amend 단계, requestId `qlpibU2Ib7XBFyIX`·`qH2In6Yf5e1k57gi` |
| M2 | **오승인 거부 실전 동작**: 틀린 확인 문자열 → `verdict: refused`, "nothing was sent" — 전송 0건. 배치 승인 레일이 실계좌 조건에서 검증됨 | run-WE3EF3ZOHDCNBWGT approval 레코드 |
| M3 | **프로세스 재시작 간 resume**: 별개 pid 3개에 걸쳐 동일 plan digest `sha256:fac7f233…`로 재개, `approval.resumed=true` | run 2·3 approval 레코드 |
| M4 | **account-seq 해석 429**: GET orders/holdings의 lazy account-seq 해석이 브로커 429(ErrRateLimited는 HTTP 429 매핑 — internal/official/errors.go) 3연속(300ms 내 3회 시도 전부). 직전 ~4분 전 soak cycle 1이 동일 계열 endpoint 호출. **한도 창 크기·집계 단위는 미상** — 1.3 retry matrix의 입력 | run 1 read-fixtures·sellable-baseline·idempotency 단계 |
| M5 | **계좌 보유 KR 종목 0** — sell-boundary·conditional-* 단계가 설계대로 skip("the tool never buys") | run 3 skip 레코드 |
| M6 | costs 단계: 이 실행이 만든 주문 0건 → 수집 대상 없음(설계 일치 — 비용은 실제 체결에서만) | costs 단계 observations |

### soak 기록에서 확정된 추가 사실 (2026-07-27, `tossctl soak status`)

| # | 사실 | 근거 |
|---|---|---|
| M7 | **`GET /api/v1/orders`의 `status`는 필수 파라미터**: status 없는 요청은 HTTP 400 `code: "invalid-request"`, `data.field: "status"`. openapi 스펙과 일치(enum `OPEN`/`CLOSED`, "전체" 그룹 없음; `OPEN`은 cursor/limit 무시·전량 반환, `CLOSED`만 cursor 페이지네이션). soak full walk가 status 없이 호출해 **46/46 사이클 전패** → `GET /api/v1/orders` 미성공 → attestation 영구 차단. → task 1.9 발주 | `~/.local/share/tossos/capability-soak.jsonl` 46 cycle, requestId 예 `rkqNEGbfWwUq0jPx`; `docs/migration/openapi.latest.json` paths./api/v1/orders.get.parameters |

### 1.9/1.10 수정 바이너리 설치 후 실측 (2026-07-27 09:20 KST 설치, soak 사이클)

| # | 사실 | 근거 |
|---|---|---|
| M8 | **orders walk 429 버스트**: 1.9 수정으로 400은 사라졌으나(walk가 5페이지/100건 수집 시작), CLOSED walk가 기본 `limit`(20)으로 페이지를 간격 없이 연발 — 클린 사이클(직전 사이클과 15분 간격, 타 엔드포인트 전부 OK)에서도 **7요청/535ms(≈13 req/s) 후 429**, 직후 `GET /api/v1/orders/{id}` 1요청도 429(패널티 창). 계정 CLOSED 이력 ≥100건이라 기본 limit으로는 매 사이클 재현 → orders 미성공 지속 → attestation 차단. openapi `limit`은 최대 100(기본 20). 기존 실측과 정합: 관측 지속 한도 9.67 req/s(soak status), M4 한도 창 미상. → task 1.11 발주 | capability-soak.jsonl 사이클 `2026-07-27T00:40:42Z`(requests 7·latency 535ms·rate_limited), `00:25:27Z`·`00:25:41Z` 동일 패턴; openapi paths./api/v1/orders.get.parameters.limit |
| M9 | **매수 1주 holdings 반영**: 사용자 토스 앱 KR 1주 매수(2026-07-27 개장 직후) 후 holdings positions 2 → 3 관측 — 2.5·sell-boundary·2.8의 선행 조건 충족 | capability-soak.jsonl positions 추이(41사이클 연속 2 → 3), `00:40:42Z` 사이클 positions 3 |

### 1.11 수정 바이너리 실브로커 확인 (2026-07-27 10:12 KST 설치)

| # | 사실 | 근거 |
|---|---|---|
| M10 | **`GET /api/v1/orders`·`orders/{id}` 최초 성공** (56사이클 만): 설치 후 첫 완주 사이클(10:19:14 KST)에서 6/6 엔드포인트 OK·completeness evaluated·ok. walk는 26페이지/2,320건(limit=100), 도중 429 두 번을 15s 백오프로 흡수(soak.log 한글 안내 2건, 요청 28건 계상, latency 35.2s에 대기 포함). **계정 CLOSED 이력 ≥2,500건**(25페이지 상한 도달 — detail에 명시) — 구 limit 20으로는 125+페이지가 필요해 페이지 상한만으로도 원리적 완주 불가였음을 확증. attest 차단은 streak 3일차(7/28 충족 예정)·token refresh 관측(당일 21:27 KST 만료 예정) 2건만 잔존. 재시작 직후 버스트 사이클 2건은 연속 재시작 SIGINT로 중단된 것(2건 중 1건은 945건/10페이지까지 진행) — account-seq 해석 429는 M4 기지 사항 유지 | capability-soak.jsonl 사이클 `2026-07-27T01:19:14Z`(전건 OK), `01:18:50Z`·`01:19:02Z`(중단 사이클); soak.log 01:19:19·01:19:35 백오프 안내; `tossctl soak status` NOT READY 잔여 2건 |

### 2026-07-27 장중 창 (09:00–15:30 KST) — 측정 0건

| # | 사실 | 근거 |
|---|---|---|
| M11 | **장중 창이 무동작 버튼으로 소모됨**: 이 날 콘솔에서 검증이 두 번 시작됐고(`YW7XSICTO4` 14:09 이전, `ORZEYQU6SI` 18:39 이후) 둘 다 `0 step(s) recorded`(에러 없음)로 끝났다 — 증거 기록 파일의 마지막 쓰기는 여전히 2026-07-26 22:48이고 단계 판정도 1차 실행 그대로다. 원인은 승인 이전이다: 시작 화면의 기본 버튼 [이어하기](`mode=resume`)는 판정이 terminal인 단계를 건너뛰는데(`Runner.settled`) 현재 기록은 전 단계가 terminal이라 **구조적으로 무동작**이며, 실제로 측정하는 [재측정](`mode=redo`)은 아래쪽 별도 섹션에 있었다. 승인 화면·nonce 타이핑에는 도달하지도 않았다. → change `console-click-approval`(무동작 기본값 제거 + 승인 클릭화) 발주 | `~/.local/share/tossos/console-launch.log` 두 줄(`verification … finished — 0 step(s) recorded`), `capability-verify.jsonl` mtime 2026-07-26T22:48:58+0900, `internal/console/pages.go` handleStart·`templates.go` 시작 화면 |

## 2차 실행 — US 시장 (2026-07-27 21:39 KST / 08:39 EDT, 콘솔 경유)

run `run-M6WQZ5WKGGE4KS4C`, 승인 채널 = 콘솔 클릭(console-click-approval), 기록
`~/.local/share/tossos/capability-verify-us.jsonl`, 심볼 MWG(보유 115주), 10단계 기록.
**US 정규장(09:30–16:00 ET) 시작 전 프리마켓 시간대**에 실행됐다.

| # | 사실 | 근거 |
|---|---|---|
| M12 | **SINGLE + MARKET 매도 조건주문이 US에서 등록된다** — 보유(MWG) 대상, 등록 직후 status `WATCHING`, `first_leg_status` `WATCHING`, `triggeredOrderId`는 등록 시 **null**(문서 형태와 일치). 2c의 기본 가설(보호 = SINGLE+MARKET 손절 단독)이 **처음으로 실측 지지**를 받았다. OCO/OTO가 LIMIT 전용이라는 문서 제약과도 정합 | conditional-register 단계 observations `conditional.register.ok/type/order_type/status.after_register/triggered_order_id.at_register` |
| M13 | **조건주문은 매도가능수량을 예약하지 않는다** — 조건주문 등록 전 115, 등록 후에도 115. 브로커가 청산 수량을 잡아주지 않으므로 **2c의 청산 수량 예약은 엔진이 계산·유지해야 한다**(= "한 심볼에 브로커측 매도 청구권 1개" 불변식을 우리가 강제해야 하는 이유) | sellable-reserved 단계 `conditional.reserves_sellable_quantity=false`, `sellable.baseline_recall.MWG=115`, `sellable.with_conditional.MWG=115` |
| M14 | **멱등키 실동작 확인(2.7)** — 동일 `clientOrderId`+동일 본문 재요청이 **같은 orderId를 재반환**하고 두 번째 주문을 만들지 않는다(open orders delta 1). 본문이 다르면 `422 idempotency-key-conflict`. **조건주문도 같은 키로 재생하면 같은 `conditionalOrderId`를 반환**한다. 주문 왕복 지연 최대 **169ms**(3회 중 최악) — 재생 안전 마진 산정 입력. 키의 계좌 스코프는 단일 계좌라 원리적 미검증 | idempotency 단계 observations, conditional-register `idempotency.conditional_replay_returns_same_id` |
| M15 | **US 프리마켓에서 주문·조건주문이 수락된다** — 08:39 EDT(정규장 전)에 지정가 주문 접수와 조건주문 등록이 모두 성공했다. KR의 `order-hours-closed` 422와 **다른 동작**이며, US 휴장·시간외 응답을 [미측정]으로 두던 advisory의 전제가 이 시간대에 한해 해소됐다(휴일·정규장 종료 후는 여전히 미측정) | conditional-register `conditional.register.session="outside US regular hours 08:39 EDT"`, 같은 run의 order 단계 접수 |
| M16 | **주문 취소가 `409 already-processing`으로 거절될 수 있다** — `{"code":"already-processing","message":"지금은 주문을 변경할 수 없어요. 잠시 후 다시 시도해주세요.","data":{"retryAfterSeconds":1}}`. 브로커가 **재시도 힌트(retryAfterSeconds)** 를 준다. 도구는 취소를 재시도하지 않아 order-cancel이 fail로 남았고, 취소되지 않은 주문 1건이 노출 상한(1건)을 채워 **order-amend·sell-boundary가 연쇄 차단**됐다 — 2a 주문 오류 분류표와 2c 취소 경로에 이 코드와 재시도 규칙을 넣어야 한다 | order-cancel 단계 reason, requestId `r4zrJzFgIxqPHPhh` |
| M17 | **조건주문 목록의 `status` 필터는 `OPEN`/`CLOSED`만 허용** — 그 외 값은 `400 invalid-request`, `data.field="status"`, `allowedValues=["OPEN","CLOSED"]`, message "유효하지 않은 주문 상태 필터입니다". 일반 주문의 M7과 같은 계열이며 **조건주문 목록 조회 경로에도 같은 제약**이 있다. 현재 코드가 다른 값을 보내고 있으므로 교정 대상 | conditional-register `conditional.list_by_status.ok=false`, requestId `rK2pMdhWbAID0fLn` |

### 이 실행이 남긴 계좌 잔여물 (2026-07-27 21:40 KST 기준 살아 있음)

- `order OsBakhtsu54X8pIXZjDO1KcOrl14B_7PnX0r75XXe5afmp1j2BNrfMaWwkKGPJyv` (MWG) — M16으로 취소 실패
- `conditional-order hjbGwc27O8eiv3xqvk2t2qp0-4E4qOvRLTYPvV2-JN4` (MWG) — 존속 측정(conditional-persist)을 위해 **의도적으로 존속 중**

`conditional-persist`가 `awaiting-restart`다: 등록한 프로세스가 죽은 뒤에만 존속을 관측할 수
있으므로 콘솔 재시작 후 이어하기가 필요하며, 절차를 마치면 도구가 둘 다 취소한다.

### 미측정으로 남은 항목 (이번 실행 기준)

- 2.1 status enum fixture / 2.7 멱등키 재생·conflict / 2.2 place-cancel-amend / 2.8 sellable 의미: **휴장 + 429로 전부 미측정** — 장중 재실행 필요.
- 2.5 조건주문 전체 / sell-boundary / sellable-reserved: **보유 0으로 원리적 측정 불가** — SELL측 검증(SINGLE+MARKET 손절 가설, 2.6의 임계 입력)은 KR 종목 최소 1주 보유가 선행 조건.
- `conditional.trigger*`: 설계상 deferred(시장 조건 필요) — 변동 없음.

### 발견된 도구 갭 (→ task 1.7)

`fail`/`skipped`는 terminal verdict라 resume이 재시도하지 않는다(runner.go `settled`, `Verdict.Terminal()`).
CLI는 `--redo`로 넘지만 **콘솔은 always-resume + redo 부재**라, 환경 원인(휴장·429·보유 0)으로
실패·생략된 측정을 웹 화면에서 다시 수행할 방법이 없다 — 콘솔 단독 운용(사용자 결정)에서는
검증이 "완료·측정 0건"으로 고착된다. 재측정 경로·장시간 사전 경고·429 강건성을 1.7로 발주.

## 3차 실행 — US 재개 (2026-07-27 22:40 KST / 09:40 EDT, 콘솔 경유)

run `run-G5DKDM3F4JVP5XM2`, 기록 `capability-verify-us.jsonl`, 5단계 기록. 2차 실행이
`awaiting-restart`로 멈춰 둔 조건주문을 **다른 프로세스 인스턴스**가 읽어 마무리했다.

| # | 사실 | 근거 |
|---|---|---|
| M18 | **조건주문은 등록한 프로세스가 죽어도 존속한다** — `proc-PNA2FRT…`가 등록한 조건주문을 `proc-NZF4D55…`가 status `WATCHING`으로 다시 읽었다. 2c의 전제 중 가장 중요한 것: 브로커측 손절은 우리 프로세스의 생존과 무관하게 계좌에 남아 있으므로, **엔진이 죽어 있는 동안에도 보호가 유효하다**. 이것이 앱측 손절이 아니라 브로커측 보호주문을 쓰는 이유 자체다 | conditional-persist `conditional.survives_process_exit=true`, `conditional.status.after_restart=WATCHING` |
| M19 | **US 조건주문 정정은 새 식별자를 발급하고 옛 식별자를 즉시 무효화한다** — `hjbGwc27…` → `H621LabR…`, 옛 id는 `404`로 더 이상 읽히지 않는다. 정정 후 status `WATCHING` 유지. **2c는 정정 응답의 새 id를 원장에 반드시 반영해야 한다** — 옛 id를 들고 있으면 취소도 조회도 실패하고, 그 상태는 "보호가 있다고 믿지만 추적하지 못하는" 최악이다 | conditional-modify `conditional.modify_issues_new_id=true`, `conditional.modify_invalidates_old_id=true` |
| M20 | **US 조건주문 취소가 동작한다** — `DELETE /api/v1/conditional-orders/{id}` 수락 후 식별자가 더 이상 읽히지 않는다. 2.5의 등록·조회·존속·정정·취소가 US에서 모두 pass다(발동만 deferred) | conditional-cancel `conditional.cancel.ok=true`, `conditional.cancel.gone_after=true` |
| M21 | **체결 비용은 여전히 미측정** — 검증 주문 2건 모두 미체결(설계상 체결 불가 지정가), 따라서 수수료·세금 실측은 이 경로로 얻을 수 없다 | costs `costs.orders_filled=0`, `costs.collected=false` |

### 도구 갭 — 잔여물 교착 (→ change `verify-clears-leftovers`)

| # | 사실 | 근거 |
|---|---|---|
| M22 | **도구가 자기 잔여물을 치울 수 없다** — M16으로 취소하지 못한 주문 1건(`OsBakht…`, PENDING)이 남았고, 이 상태에서 ① [이어하기]는 order-cancel이 terminal(`fail`)이라 건너뛰며 그 취소는 어느 계획에도 없어 `ErrOutsidePlan`, ② [재측정]은 세 단계 모두 첫 요청이 노출 상한(`1 live order(s) … cap is 1`)에 걸려 거절, ③ 브로커 앱에서 손으로 취소해도 상한은 **기록**을 보고 세므로 계속 차단. 3차 실행이 5단계를 기록하고도 이 주문을 남긴 채 끝난 것이 실증이다 | order-amend·sell-boundary reason(`ErrExposureCap`), 3차 실행 종료 메시지("객체 1건이 아직 계좌에 살아 있다"), costs `order.status.…afmp1j2=PENDING` |
| M23 | **승인 창 만료가 장중 창 3개를 소모했다** — 22:03·22:27·(그 사이 1회) 세 번의 실행이 모두 `승인 창 만료`로 0단계 종료. 원인 둘: ① 만료 문구가 "확인 문자열이 만료되었다"라고 말하는데 콘솔 승인에는 문자열이 없다(console-click-approval 이후 잔존 문구), ② **끝난 실행이 화면에 있는 동안 시작 제어가 렌더되지 않아** 만료될 때마다 콘솔 재시작이 강제됐다. M11과 같은 계열의 결함이며 원인은 다르다 | `capability-verify-us.jsonl`의 approval 3건(verdict `refused`), `console-launch.log`의 `0 step(s) recorded` 3회, `internal/console/templates.go` 시작 섹션의 `{{else}}` 분기 |

## 4차 실행 — US 재측정 (2026-07-28 00:20 KST / 11:20 EDT, 콘솔 경유)

run `run-IXCQU5UBZE`, 재측정 대상 4단계(승인 목록 8건), 4단계 기록. 직전 정리 실행
(`run-3YX2BQPECCFLVF5O`)이 M22의 잔여 주문을 승인 목록 위에서 취소해 노출 상한이 비어 있었다.

| # | 사실 | 근거 |
|---|---|---|
| M24 | **`409 already-processing`은 취소 전용이 아니다 — 정정에도 온다.** `POST /api/v1/orders` 수락(107ms) **직후**의 `POST /api/v1/orders/{id}/modify`가 같은 코드·같은 `retryAfterSeconds:1`로 거절됐다(requestId `si1tUiUvi8DzWXr5`). 브로커가 접수를 처리하는 동안 그 주문에 대한 **모든 후속 변경**이 이 창에 걸린다고 보는 것이 맞다 — 2c의 손절 정정·취소 경로 전부에 이 코드 처리가 필요하다 | order-amend 단계 reason·calls |
| M25 | **M16 교정이 실계좌에서 작동한다** — order-cancel이 409를 만나 1회 재시도 후 수락, 단계 **pass**. `order.status.after_cancel=PENDING_CANCEL`, `canceledAt` 존재, `filledQuantity=0`. 접수 후 상태는 `PENDING`, `timeInForce=DAY` | order-cancel `order.cancel.retries=1`, `order.cancel.ok=true` |
| M26 | **도구 결함: 실패한 단계가 자기가 낸 주문을 취소하지 않는다** — order-amend가 정정 거절로 조기 반환하면서 방금 접수한 주문을 남겼고(성공 경로에만 `cancelLiveOrders`가 있다), 그 1건이 노출 상한을 채워 **sell-boundary가 `ErrExposureCap`으로 아무것도 보내지 못했다**. "이 도구가 만든 객체는 모두 취소되어 끝난다"는 불변식이 실패 경로에서 성립하지 않는다 | order-amend artifacts(cancelled=false), sell-boundary reason |

## 5차 실행 — US 재측정 2단계 (2026-07-28 00:54 KST / 11:54 EDT, 콘솔 경유)

run `run-EZQRKZZ7WPEMDS2I`, 재측정 대상 2단계(승인 목록 6건). 직전 실행
(`run-QY3MJTSONBTL6CTR`)이 M26의 잔여 주문을 정리 prologue로 취소해 상한이 비어 있었다.
**US 측정이 이 실행으로 완료됐다** — 12단계 pass, `conditional-trigger`만 설계상 deferred,
`idempotency-ttl-edge`는 요청하지 않아 skipped.

| # | 사실 | 근거 |
|---|---|---|
| M27 | **M24·M26 교정이 실계좌에서 작동한다** — order-amend가 409를 만나 1회 재시도 후 수락(**pass**), 그리고 그 단계가 만든 주문 2건이 모두 취소된 채 끝났다(artifacts 4건 = 접수 2 + 취소 2). 연쇄로 막혀 있던 sell-boundary도 **pass** | order-amend `order.amend.retries=1`, `order.amend.ok=true`, artifacts cancelled=true ×2 |
| M28 | **일반 주문 정정은 새 식별자를 발급하고 옛 식별자를 `PENDING_REPLACE`로 남긴다** — 조건주문 정정(M19)이 옛 id를 **404로 즉시 무효화**하는 것과 **다르다**. 2c의 귀속 규칙은 두 경로를 같게 다루면 안 된다: 조건주문은 "옛 id는 없다", 일반 주문은 "옛 id는 대체 중 상태로 읽힌다" | order-amend `order.amend.issues_new_id=true`, `order.amend.original_status=PENDING_REPLACE`, `order.amend.current_status=PENDING` |
| M29 | **미체결 매도 지정가 주문은 매도가능수량을 예약한다** — 매도 1주를 걸어두는 동안 sellable 115 → **114**. **조건주문은 예약하지 않는다(M13)**. 이 대조가 2c 설계의 핵심 입력이다: 브로커측 조건주문 손절은 청산 수량을 잡아주지 않으므로 엔진이 스스로 유지해야 하고, 지정가 매도로 보호를 거는 대안은 수량을 잡아주는 대신 시장가 즉시성을 잃는다 | sell-boundary `sell.reservation.resting_sell_reserves=true`, `sellable_at_start=115`, `sellable_with_resting_sell=114` |
| M30 | **보유 초과 매도는 `422 insufficient-sellable-quantity`로 거절된다**(requestId `sBNQ5vxQ65…`). 부분 매도는 수락된다. **전량 매도는 미검증** — 보유 115주가 도구의 노출 상한(1주)을 넘어 원리적으로 이 경로로는 측정할 수 없다 | sell-boundary `partial_accepted=true`, `over_holding_rejected=true`, `full_accepted=unverified` |

### 남은 도구 결함 (경미)

`order.amend.ok`의 detail이 시장과 무관하게 `"accepted a KR price+quantity amend"`로 고정돼
있다([steps.go:438](../../../internal/verifylive/steps.go)). 실제 요청은 시장별로 갈라진다
(US는 quantity 미전송 — [mutate.go:404-410](../../../internal/verifylive/mutate.go)). 동작은
옳고 **기록이 US 실행에 대해 사실이 아닌 문장을 남긴다**. 기록의 정확성 문제이므로 교정 대상.

### US 측정 최종 상태

| 상태 | 단계 |
|---|---|
| pass (12) | read-fixtures, sellable-baseline, idempotency, order-cancel, order-amend, sell-boundary, conditional-register, sellable-reserved, conditional-persist, conditional-modify, conditional-cancel, costs |
| deferred (1) | conditional-trigger — 체결될 의도의 주문이 필요해 별도 세션 |
| skipped (1) | idempotency-ttl-edge — 두 번째 라이브 주문을 의도적으로 만드는 단계, 요청하지 않음 |

**2c(`add-protection-orders`) 착수에 필요한 실측은 확보됐다**: 조건주문의 프로세스 존속(M18),
US 등록 가능(M12), 정정의 식별자 교체 규칙(M19·M28), 청산 수량 예약 여부의 대조(M13·M29),
취소·정정의 `already-processing` 처리(M16·M24).

## 6차 실행 — KR 장중 (2026-07-28 10:33 KST, 콘솔 경유)

run `run-KGJZOUQY7UFCHCDK`, 기록 `capability-verify.jsonl`, 10단계 기록. **KR 정규장
(09:00–15:30) 안에서 실행된 첫 측정**이다. 계좌 보유가 M5의 "KR 종목 0"에서 3종목
(333430 6주, MWG 115주, TSLA 0.0002주)으로 바뀌어 SELL측 단계가 처음으로 실행 가능해졌다.

| # | 사실 | 근거 |
|---|---|---|
| M31 | **KR 주문 status enum 실측(2.1)** — CLOSED+OPEN 11페이지 197건에서 실제로 관측된 값은 `CANCELED`(119)·`FILLED`(65)·`REJECTED`(13) **셋뿐**이고, 세 값의 필드 구성은 동일하다. 문서에 있으나 이 계좌가 한 번도 만든 적 없는 값: `PENDING`·`PENDING_CANCEL`·`PENDING_REPLACE`·`PARTIAL_FILLED`·`CANCEL_REJECTED`·`REPLACE_REJECTED`·`REPLACED` — 이들에 대한 상태 파생은 **미측정**이다. 특히 2.1이 명시적으로 요구한 **`CANCEL_REJECTED`/`REPLACE_REJECTED`의 "별도 주문 레코드" 형태는 관측되지 않았다**(둘 다 `listed=false`) — 2.1은 이 실행으로 닫히지 않는다 | read-fixtures `order.status.observed`·`order.status.documented_unobserved`·`order.status.{cancel,replace}_rejected.listed` |
| M32 | **KR 일반 주문의 접수·취소·정정이 장중에 동작한다(2.2)** — 최소 수량 지정가 접수 후 `PENDING`/`DAY`, 취소 수락, 정정 수락. 정정은 **새 id를 발급하고 옛 id는 `PENDING`으로 읽힌다**(US M28과 같은 계열). 다만 **취소 수락 직후 status가 여전히 `PENDING`이고 `canceledAt`이 없다** — US M25의 `PENDING_CANCEL`+`canceledAt` 존재와 다르다. 시장 차이인지 읽기 시점 차이인지는 이 관측만으로 가릴 수 없으나, 어느 쪽이든 **2a 상태 파생은 status 하나로 취소를 판정하면 안 된다**는 결론은 같다 | order-cancel `order.status.after_cancel=PENDING`·`order.canceled_at.present=false`, order-amend `order.amend.issues_new_id=true`·`original_status=PENDING` |
| M33 | **KR 매도 경계(2.2·2.8)** — 부분 매도 수락, 보유 초과는 `422 insufficient-sellable-quantity`("주문수량이 일반매도가능수량을 초과하였습니다", requestId `s8bMo3dbo8lp1xqy`). **미체결 매도 지정가는 매도가능수량을 예약한다**(6 → 5). 전량 매도는 보유 6주가 도구의 노출 상한 1주를 넘어 **원리적으로 미측정** — US M30과 같은 한계 | sell-boundary `partial_accepted=true`, `resting_sell_reserves=true`, `over_holding_rejected=true`, `full_accepted=unverified` |
| M34 | **KR 조건주문도 매도가능수량을 예약하지 않는다** — 등록 전 6, 등록 후 6. US M13과 같다. M33의 지정가 매도와의 대조가 **시장 양쪽에서 같은 방향으로** 확인됐다: 브로커측 조건주문 손절은 청산 수량을 잡아주지 않으므로 2c의 예약 공식은 엔진이 계산·유지해야 한다 | sellable-reserved `conditional.reserves_sellable_quantity=false`, `sellable.baseline_recall.333430=6`, `sellable.with_conditional.333430=6` |
| M35 | **KR 멱등키 실동작(2.7)** — 동일 키+동일 본문 재요청이 같은 orderId를 반환하고 두 번째 주문을 만들지 않는다(open delta 1). 본문이 다르면 `422 idempotency-key-conflict`("동일한 clientOrderId 로 다른 내용의 주문을 요청할 수 없습니다"). 조건주문도 같은 키로 같은 `conditionalOrderId`를 반환한다. **주문 왕복 지연 최대 182ms**(US는 169ms) — 재생 안전 마진 입력. 키의 계좌 스코프는 단일 계좌라 미측정 | idempotency observations, conditional-register `idempotency.conditional_replay_returns_same_id=true` |
| M36 | **M17(조건주문 목록 status 필터) 교정이 실계좌에서 확인됐다** — KR에서 목록 조회가 통과하고 새 조건주문을 포함한다. 아울러 read-fixtures가 429를 2회 만나고도 백오프 재시도로 **pass** — 1.11의 교정이 실계좌에서 작동한다 | conditional-register `list_by_status.ok=true`·`contains_new=true`, read-fixtures calls의 `rate limited` 2건 + 단계 pass |

### 이 실행이 만든 교착 — 도구가 자기 측정 대상을 지웠다

| # | 사실 | 근거 |
|---|---|---|
| M37 | **정리 prologue가 존속 측정 대상을 취소해 KR 조건주문 체인이 교착됐다** — 10:34 등록된 `grLKqiGuC…`는 "다음 단계가 프로세스 종료 후 존속을 증명한다"는 이유로 **의도적으로 살려 둔** 객체인데, 1분 뒤 다음 실행(`run-TGPQY66LLIJ45PED`)의 `cleanup` 단계가 10:35:31에 이를 취소했고 **같은 실행의 `conditional-persist`가 같은 초에 "이 검증이 만든 살아 있는 조건주문이 없다"로 skip**했다. 이후 07-28 21:01·21:02, 07-29 20:31의 세 실행도 전부 같은 이유로 conditional-* 4단계를 skip — **2.5는 도구가 만든 교착으로 5일간 측정 0건**이었다. → change `verify-reopens-conditional-chain`로 교정 | `run-TGPQY66LLIJ45PED` cleanup artifact(`cancelled_at=2026-07-28T01:35:31Z`)와 같은 run의 conditional-persist skip reason, 이후 3개 run의 동일 reason |

## 7차 실행 — KR 조건주문 체인 완주 (2026-07-29 21:43–23:16 KST, 콘솔 경유)

runs: `run-B3L57BY47BFBVGZS`(21:43, 등록) → `run-OJRFYBGI4UOBM4MD`(22:23, 존속) →
`run-RC4ENQM5XP5TMYBB`·`run-CBJUABS5FEBHD5YD`(22:28, 정정 2회 연속 fail) →
`run-MFYCBNFYOH7B7TQW`(23:16, 정정·취소 완료). M37의 교착이 풀린 뒤 **KR 조건주문 체인이
처음으로 끝까지 실행됐다.**

| # | 사실 | 근거 |
|---|---|---|
| M38 | **KR SINGLE+MARKET 매도 조건주문이 정규장 밖에서 등록된다** — 21:43 KST(평일 장 종료 후) 등록 수락, status `WATCHING`, `first_leg_status` `WATCHING`, `triggeredOrderId` null. M1의 `order-hours-closed`(일요일 일반 주문 422)와 대조되지만 **두 관측의 조건이 같지 않다**(휴장일 일반 주문 vs 평일 장후 조건주문) — "조건주문만 예외"라고 단정할 수 없고, KR 정규장 밖 **일반** 주문 접수는 여전히 [미측정] | conditional-register `conditional.register.session="outside KR regular hours 21:43 KST"`, `status.after_register=WATCHING` |
| M39 | **KR 조건주문도 등록한 프로세스가 죽어도 존속한다** — `proc-XNRHLSP…`가 등록한 것을 `proc-CEUKLMQ…`가 `WATCHING`으로 다시 읽었다. US M18과 같으며, **시장과 무관한 성질**로 확인됐다 | conditional-persist `conditional.survives_process_exit=true`, `status.after_restart=WATCHING` |
| M40 | **KR 조건주문 정정도 새 식별자를 발급하고 옛 식별자를 즉시 404로 무효화한다** — `p7hQz7HAXc…` → `uiIndFbm…`, 옛 id는 `conditional-order-not-found`("존재하지 않는 설정입니다", requestId `tTXXXQMKNPt7Zuqw`). 정정 후 status `WATCHING`, 발동가 2625. **US M19와 동일** — 조건주문 정정의 식별자 교체 규칙은 시장 불변이고, 일반 주문 정정(M28·M32: 옛 id가 계속 읽힌다)과는 **다르다**. 2c 원장은 정정 응답의 새 id를 반드시 반영해야 한다 | conditional-modify `modify_issues_new_id=true`, `modify_invalidates_old_id=true`, `status.after_modify=WATCHING` |
| M41 | **KR 조건주문의 정정·취소가 정규장 밖에서 접수된다** — 23:16 KST에 `POST …/modify`(124ms)와 `DELETE …/{id}`(77ms)가 모두 수락됐고 취소 후 식별자가 읽히지 않는다. 2.5가 요구한 "정규장 밖 동작" 중 **등록·정정·취소는 이것으로 측정**됐다. **발동은 여전히 미측정**이므로 "장 밖에서 손절이 작동한다"로 확대 해석하면 안 된다 | conditional-modify·conditional-cancel calls의 시각과 수락, `cancel.gone_after=true` |
| M42 | **도구 결함: 승인 계획이 단계가 지목할 객체를 이름하지 않았다** — `conditional-modify`의 계획 줄은 실행의 probe 심볼(`005930`)로 만들어지는데 단계 본문은 살아 있는 조건주문의 심볼(`333430`)로 요청을 만든다. `Plan.Authorises`가 심볼을 정확 비교하므로 22:28의 두 실행이 `ErrOutsidePlan`으로 정지했다 — **전송 0건**, 단계당 GET 1회. 레일은 설계대로 동작했고 승인 밖 요청은 나가지 않았다. 이 결함은 M37의 교착이 풀려 modify가 **처음으로 인가 검사에 도달한 순간** 드러났다(그 전에는 항상 skip). → change `verify-plans-the-object-it-mutates`(`760d213`)로 교정, 같은 계좌에서 23:16 pass 확인 | 두 run의 conditional-modify reason(`… is about to modify-conditional for SELL 1 333430, which is not on the list approved …`), `console-launch.log`의 동일 문구 2회 |

### KR 측정 최종 상태 (2026-07-29 23:16 기준)

| 상태 | 단계 |
|---|---|
| pass (12) | read-fixtures, sellable-baseline, idempotency, order-cancel, order-amend, sell-boundary, conditional-register, sellable-reserved, conditional-persist, conditional-modify, conditional-cancel, costs |
| deferred (1) | conditional-trigger — 체결될 의도의 주문이 필요해 별도 세션 |
| skipped (1) | idempotency-ttl-edge — 두 번째 라이브 주문을 의도적으로 만드는 단계, 요청하지 않음 |

**계좌 잔여물 0건** — 기록 전체를 재생해 계산한 미취소 artifact가 없다. 이 도구가 만든 객체는
모두 취소된 채 끝났다.

### 2.5에 남은 미측정 — 단계 pass는 task 완료가 아니다

`verify run`의 조건주문 단계는 전부 pass지만, task 2.5가 요구하는 항목 중 다음은
**도구가 측정하지 않는다**:

- **발동 관측과 `triggeredOrderId` 노출 지연** — `conditional-trigger`가 설계상 deferred. 2c의 기본 가설(SINGLE+MARKET 손절)이 실제로 **발동해 체결되는지는 양 시장 모두 미측정**이며, 이것이 2.5의 가장 큰 구멍이다.
- **만료** — 1주 만료로 등록하지만 만료 시점의 동작을 관측하는 단계가 없다.
- **부분체결 잔량**과 **OCO sibling 취소 시점** — 단계 없음(OCO/OTO는 LIMIT 전용이라 보호 가설 밖이지만 2.5 문언에는 남아 있다).
- **조건주문과 일반 매도 동시 제출의 거부 의미** — 단계 없음.
- **ProtectiveCapability 속성 기록** — 실측은 JSONL에 있으나 2.5가 요구한 속성 형태로는 아직 산출하지 않았다.

따라서 **2.5는 미완료**다. 등록·조회·존속·정정·취소·예약 여부는 양 시장에서 닫혔고,
남은 것은 발동 계열이다.

## 8차 — 발동 측정 준비 관측 (2026-07-30 00:20–00:50 KST, US 정규장)

주문을 내지 않은 읽기 전용 관측이다. 발동 측정에 쓸 계측기를 고르고, 그 측정이 기댈 수 있는
시각·상태 필드가 무엇인지 확인했다. 사용자가 직접 낸 주문의 결과를 이력에서 읽었다.

| # | 사실 | 근거 |
|---|---|---|
| M43 | **US는 당일 매수분을 당일 매도할 수 있다** — 같은 `userOrderDate 2026-07-30`에 AGRZ(NAS0251001001) 매수 1주 @$0.4077(orderNo 1), 매수 1주 @$0.4101(orderNo 2)가 체결완료된 뒤 매도 2주 @$0.3552(orderNo 7)가 체결완료됐다. 결제 미도래를 이유로 한 거절이 없다. `verify-holds-what-it-awaits/issues.md` I3의 미측정 항목 하나가 닫혔다 — **발동 측정에서 "child 매도 거절"을 결제 사유와 혼동할 위험이 US에서는 없다** | `tossctl orders completed`, 2026-07-30 orderNo 1·2·7, 전 건 `status="체결완료"`, `executedQuantity`가 `orderQuantity`와 같음 |
| M44 | **브로커 완료 이력은 체결 시각을 주지 않는다** — `lastExecutedAt`이 28건 **전 건 null**(양 시장). 더구나 US 주문의 `orderedAt`은 `2026-07-30 00:00:00.000`으로 **날짜만** 있고, KR 주문은 실제 시각(`2026-07-29 13:55:56.354`)을 담는다. → 발동 측정의 네 시각 중 `child_order_filled_at`은 **브로커에서 읽을 수 없고 자체 관측 시각이어야 한다.** US는 주문 접수 시각조차 이력에서 못 얻는다 | 28건 raw payload의 `lastExecutedAt`·`orderedAt` |
| M45 | **도구 결함(미교정): `Order.SubmittedAt`이 접수 시각이 아니다** — `completed_orders.go:258`의 정규화가 `lastExecutedAt → version → orderedAt` 순서로 채우는데 `lastExecutedAt`은 항상 null이므로(M44) 실제로는 `version`(레코드 갱신 시각)이 들어간다. 삼성전자 주문은 `orderedAt 13:55:56` / `version 20:04:04`로 **6시간 어긋난다**. 이 값을 최신성 기준으로 쓰는 곳이 있다(`matchesDelayedCancelRecoveryHint`, `findMatchingCompletedOrder`의 `earliestSubmittedAt`). **발동 측정은 이 필드를 시각 근거로 쓰지 않는다** | `internal/client/completed_orders.go:258`, 2026-07-29/13027844의 orderedAt·version |
| M46 | **완료 이력의 status는 라이브 주문 status enum이 아니다** — 관측값은 한글 표시값 `체결완료`·`취소`·`실패` 셋이고, 사용자 취소와 브로커 거부는 `cancelType`이 가른다: 사용자 취소 = `cancelType="1"` + `orderTransactionType="취소"`, 거부 = `status="실패"` + `cancelType="2"` + `orderTransactionType="정상"`. **M31의 CANCELED/FILLED/REJECTED와 같은 enum이 아니다** — 두 어휘를 섞어 판정하면 안 된다 | 2026-07-30 orderNo 4·6(취소), 2026-07-29/13027844·13242973·11344231(실패) |
| M47 | **`afterMarketOrder`는 장 밖 주문을 나타내지 않는다** — KR 15:39(정규장 종료 후) 체결 주문이 `orderPriceType="81"`을 달았고 `afterMarketOrder`는 그때도 `false`였다. 장 밖 여부의 신호는 `orderPriceType`이다 | 2026-07-28/13229020 |
| M48 | **`quote trades --count` 상한은 50** — 300 요청 시 400 `invalid-request`, `data.field="count"`, `constraint {min:1, max:50}` | requestId `uFUsmaU51r35r0y2` |
| M49 | **계측기 후보 유동성 실측** — 발동을 관측하려면 임계 통과가 관측 가능한 시간 안에 일어나야 한다. 네 종목을 같은 방법(체결 50틱, 호가 1회)으로 재 봤다 | 아래 표 |

### M49 상세 — 계측기 선택 근거

| 종목 | 틱 간격(중앙/최대) | 스프레드 | 호가 단수 | 후보 유니버스 관측 | 1주 가격 |
|---|---|---|---|---|---|
| MWG (US) | 8.2분 / **93분** | 1.12–1.28 = **13.3%** | 1 | 0 | $1.18 |
| AGRZ (US) | 0.24초 / 1초 | 0.38–0.3965 = 4.25% | 1 | 198 | $0.39 |
| 333430 (KR) | 1.8초 / 8초 | 3290–3295 = 0.152% | 10 | 0 | ₩3,290 |
| **TSLA (US)** | **0.18초 / 1초** | 299.88–299.94 = **0.0200%** | 1 | 995 | $300 |

여기서 나온 부수 관측 둘:

- **US 호가 응답은 매수·매도 각 1단만 준다**(KR은 10단). 엔드포인트의 한계이며 실제 호가창이
  1단이라는 뜻이 아니다. → 발동 시점의 호가 깊이를 US에서 근거로 삼을 수 없다.
- **US 체결가가 제시 스프레드 안쪽에서 찍힌다**. 따라서 이 스냅샷만으로는 브로커가 조건을
  체결가로 보는지 호가로 보는지(`trigger_price_basis`)를 가릴 수 없다 — **스프레드가 넓은
  종목일수록 두 기준의 차이가 커서 구별이 쉬워지지만, 넓은 스프레드는 곧 낮은 유동성이라
  발동 자체가 관측되지 않는다.** MWG가 정확히 그 이유로 탈락했다.

**선택: TSLA.** 임계 통과가 초 단위로 일어나고, 스프레드가 좁아 임계를 최근 체결가와 최우선
매수호가 사이에 놓아 두 기준을 가를 수 있다. 대가는 $300 노출과 후보 유니버스 관측 995회
겹침이다(엔진 미가동이지만 측정 전 `exclude_symbols` 고정이 필요하다).
