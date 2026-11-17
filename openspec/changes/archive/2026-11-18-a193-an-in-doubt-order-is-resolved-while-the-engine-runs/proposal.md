# a193 — 엔진이 도는 동안 IN_DOUBT 주문을 해소한다

> **상태: 2단계(구현·시험·변이·구현 리뷰), High-risk(주문 해소·알림).** 1단계 Manager 검증 통과(2026-11-18, HEAD `0a48e0138`)·처분 반영(design D0). proposal-freeze.
> Story: `STORY-TOS-a193`(FEAT-TOS-009 — Manager 확인 대상, review.md). base: `791fb6ef1`(`base-commit.txt`).

> **⚠ 이 change 뒤에도 남는 한계 — 먼저 읽을 것.** 엔진 손절(보호 매도) 주문은 제출 전 계좌 기준선(baseline)을 싣지 않는다(`ExitObserver.submit`
> 의 `Submit.Place` 요청, exitloop.go:1556 — 손으로 읽음; 기준선을 채우는 production 자리는 전략 진입 `Gateway.PlaceClaimedStrategy` 의 전달 하나,
> strategy_gateway.go:65). 그래서 세션 중 해소가 **손절 주문의 부재**를 만나면 결과는 지금과 같이 **park(`UNRESOLVED_IN_DOUBT`)** 이고, 손절은 운영자가
> 해동 명령(`tossctl engine attempt-resolve`)으로 닫을 때까지 **동결된 채** 남는다. 이 change 가 바꾸는 것은 **그 사실을 아는 시각**이다 —
> 2026-10-29 사고의 36.8시간 침묵이 약 1분(관찰 45초 + 스캔)으로 줄고, park 원인 critical 이 그때 나간다. **부재를 증명해 손절을 자동으로 다시
> 무장하는 일(매도용 부재 증거 모델 + 기준선)은 후속 Story 후보다**(정본 order-execution 「부재 판정의 증거 모델은 매도에 대해 성립해야 한다」가 그
> 모델 없이 기준선만 싣는 것을 금지한다).

## Why

2026-10-29 운영 사고(Manager 실측 — 운영 저널 사본·엔진 로그, 시각·횟수·상태만):

- 한 포지션의 손절(STOP_LOSS_LADDER, LIMIT 매도)이 10:00:05 KST 부터 약 5초마다 재제출되며 브로커에 거절됐다 — `opposite-pending-order-exists`
  6,206회(10:00~19:59), `invalid-request`(호가 단위) 634회(10:00~15:18), `order-hours-closed` 2,691회(20:00~23:57).
- 23:57:05 KST 그 제출 하나가 HTTP 5xx(`dispatch_outcome_unknown`)로 `IN_DOUBT` 가 됐다. 그 포지션의 `exit_events` 는 23:57:11 에 끊겼다.
- attempt 는 **엔진 재시작 시각 10-31 12:43:01 KST** 에야 `IN_DOUBT → UNRESOLVED_IN_DOUBT` 로 갔다(사유 원문 `absence cannot be corroborated: the mutation was
  submitted without a pre-dispatch account baseline`). `order.unresolved_in_doubt` critical 은 12:43:04 적재·12:43:08 전달. **그 사이 36.8시간 alert_outbox 새 행 0.**
- 포지션은 11-03 09:08 KST 엔진 밖에서 닫혔다. 해동 명령 실행 기록 없음. 10-30(금)·11-02(월) 두 거래일이 손절 없이 지나갔다.
- 저널 전체 `intents` 13,052건이 전부 `source=engine, side=SELL`.

코드 근거(전부 `tools/logic-map` AST — `analysis/cited-ast/`; SQL·상수·stdlib 는 「손으로 읽음」, `analysis/cited-ast/README.md`):

| 사실 | 근거 |
|---|---|
| IN_DOUBT 를 해소하는 production 호출 자리는 둘뿐이고 둘 다 세션 밖이다 | `Resolver.Resolve` 의 production 호출자: `Recovery.Run`(call@300, B3 순회 안)과 수동 flatten `Saga.resolveStep`(call@428, 자기 취소 attempt 만). `Recovery.Run` 은 `Runtime.Run` B2(if@294 `r.opts.Recover`)에서 **루프보다 먼저 한 번** 돈다 |
| exit 제출은 IN_DOUBT 를 받으면 발의를 무장한 채 로그 한 줄만 남긴다 | `ExitObserver.submit` B10(case@1577) → `noteExitOrder`(call@1585) → return@1586(nil). `noteExitOrder` 는 대기열 append(append@219) 뿐 — outbox 행 없음 |
| 그 뒤 판정 진입은 IN_DOUBT 를 침묵으로 지나간다 | `noteHeldProposal`: `ExitIntentAttempts` 판정에서 IN_DOUBT 는 `exitIntentAttemptsQ` B11(case@216 — 소스는 `default:` 절) → live → B15 `ExitIntentLive`. `noteHeldProposal` 의 B4(Unaccepted 해제)·B6(Parked 알림)·B7(Awaiting) 어디에도 안 걸린다 — **IN_DOUBT 를 다루는 갈래 0** |
| 엔진 정본은 IN_DOUBT 발생을 critical 로 정했다 | engine-safety 「등급화된 알림」 첫 문장(critical: IN_DOUBT·UNRESOLVED_IN_DOUBT 발생). 위 두 줄이 그 요구를 지키지 못했다 — 이 change 의 처음 봄 알림(Q-A 처분 (b))이 세션 쪽을 닫는다 |

즉 **세션 중에 IN_DOUBT 가 된 주문은 엔진을 다시 띄울 때까지 아무도 해소하지 않고, 그 사이 알림도 없다.** 정본 order-execution 「IN_DOUBT 해소」는
세션 중 해소를 별도 change 로 미뤘고 그 change 가 **해소 진입점 배선 · 주기와 동시 실행 상한의 실측 · park 가 계정 전역 진입을 차단한다는 공시**
셋을 모두 가져야 한다고 적었다(SHALL). 이 change 는 배선과 공시를 가지고, 실측은 **시험 측정(tasks 3.9)과 배포 전 운영 측정(tasks 1.10)** 으로 가진다 — 정본 delta 는
실측 기록을 운영 활성(배포)의 선행 조건으로 남긴다. 1단계 현재 실측은 **계획**이다.

## What Changes

1. **세션 중 해소(D1·D2)**: 새 보조 실행자(Auxiliary) `in-doubt-settle` 가 기동 복구가 끝난 뒤(`Runtime.Run` 의 B2 다음, 보조 실행자 시작 B4) 시작해 5초마다
   원장의 미종결 attempt 를 읽고(직전 스윕이 끝난 뒤 5초 — 해소가 도는 동안 다음 읽기는 늦어진다) **`IN_DOUBT` 이면서 `PLACE` 인 것만** 하나씩 기존 `execgw.Resolver.Resolve` 로 해소한다. 동시 실행 상한 1(직렬). 증명
   규칙(목록 대조·dedup·안정 관찰·부재 교차 확인·오염 창)은 한 글자도 바꾸지 않는다. 재생(`ReplayInDoubt`)은 부르지 않는다. 재시작 규칙(RECORDED→NOT_DISPATCHED,
   DISPATCH_STARTED→IN_DOUBT)은 세션 중 적용하지 않는다(정본 SHALL NOT 그대로). exit 관측 루프는 해소를 기다리지 않는다(다른 goroutine).
2. **종료는 park 가 아니다(D3)**: 엔진 정지로 ctx 가 취소되면 attempt 는 IN_DOUBT 로 남는다. **Resolver 를 고치지 않는다** — AST 로는 취소가 park 갈래로
   가지만(`waitOrExpire` B1 뒤 `clk.Sleep` 오류 → false → `Resolve` B16/B26 → `park`), park 의 원장 쓰기 `ResolveUnresolved` → `transition` 이 같은 취소된
   ctx 로 `BeginTx`(call@553)를 열어 그 자리에서 실패하므로 원장·진입 게이트는 바뀌지 않는다(`park` B1 return@377 — `Gate.Block` 앞, stdlib
   `database/sql` 의 ctx 검사는 손으로 읽음). 호출자(새 실행자)는 오류 + `ctx.Err() != nil` 이면 결과를 버리고 멈춘다. 이 「우연이 지키는 안전」은 시험으로
   못 박는다(tasks 3.4). 기동 복구 경로도 같은 성질이다(`Recovery.Run` B9 → `Runtime.Run` B2 return@295).
3. **오래 남은 IN_DOUBT 알림(D4)**: PLACE attempt 가 `DispatchStartedAt`(없으면 `RecordedAt`)부터 **10분**(= 2 × `DefaultResolveConfig().MaxDuration` 5분)
   이상 IN_DOUBT 면 실행자가 기존 critical 타입 `order.in_doubt` 로, key `order.in_doubt|stale:<attempt id>`, 알림기의 critical 기록 단일 입구
   (`RecordCritical`, 창 0)로 한 번 기록한다. 실행자 자체가 취소 아닌 사유로 멈추면(패닉 포함) `OnStop` 이 같은 입구로 critical 하나를 기록한다.
   **처음 봄 알림(Q-A 처분)**: 실행자가 IN_DOUBT PLACE 를 처음 볼 때 critical 1회(key `order.in_doubt|attempt:<id>`).
   **park 통지(Q-E 처분)**: 무장된 exit 발의의 attempt 는 기존 `recordParkCause` 가, 그 밖의 PLACE 는 실행자가 critical 1회(key `order.unresolved_in_doubt|attempt:<id>`).
   **원장 읽기 연속 실패(Q-F 처분)**: `PendingAttempts` 가 연속 10분 이상 실패하면 critical 1회. 처음 봄·10분 검사는 스윕 머리에서 해소보다 먼저 돈다(별도 goroutine 없음 —
   늦어짐은 운영 실측 동시 최대 3 으로 유계, design D0.3).
4. **바꾸지 않는 것**: `execgw.Resolver` 전체, `Recovery.Run`, exit 관측 루프(`submit`·`noteHeldProposal`), 원장 스키마, a166 배선 census 의 네 함수
   (`buildGateway`·`newGatewayReplay`·`NewContext`·`NewOrderPath`), 알림 이벤트 열거(`internal/obs` 무편집), 설정·토글.

## 이 change 가 하지 않는 것 (비목표 — 후속 후보는 design §7)

- **손절 주문의 부재 증명**(제출 전 baseline 을 실어 부재면 FAILED_CONFIRMED → 손절 재무장) — 후속 Story 후보. 이 change 뒤에도 손절의 부재는 park 로 끝난다.
- 취소·정정(CANCEL·AMEND) attempt 의 세션 중 해소.
- 재생(`ReplayInDoubt`) 활성화.
- 외부(사람) 매수 주문 취소, 장 마감 뒤 재제출 억제, 호가 단위(a087), 사고 당일 알림 13시간 지연의 원인(미조사).
- 엔진·flatten 프로세스 사이 「검사 → 기록」 경합(기존 공백, Q-H) — 후속 후보.

## Capabilities

### Modified Capabilities

- `order-execution`: 「IN_DOUBT 해소」 — 「재시작 복구가 호출 순서를 소유한다」 문단을 「기동은 재시작 복구가, 세션은 세션 해소 실행자가」로 고치고, 「세션 중 해소는
  이 change 의 범위가 아니다」 문단을 세션 해소의 규칙(대상·주기·상한·조회 대조만·종료 시 IN_DOUBT 유지·오래 남은 IN_DOUBT 알림·park 의 계정 전역 진입
  차단 공시)으로 바꾼다. 시나리오 6개 추가.

## Impact

- 새 파일: `internal/app/engine/in_doubt_settle.go`(실행자 · `Context.InDoubtSettler`), 시험 파일 `internal/app/engine/a193_in_doubt_settle_test.go`,
  `cmd/tossctl/a193_in_doubt_settle_wiring_test.go`.
- 기존 함수 본문 편집(gate ⑤ FLM 대상): `cmd/tossctl/engine.go` `engineRuntime`(보조 실행자 목록에 원소 1), 시험
  `TestProductionRuntimeStartsExactlyOneAlertDeliverer`(기대 목록에 이름 1). 둘 다 a166 census 밖.
- 브로커 호출: IN_DOUBT PLACE 가 있을 때만, 해소 하나당 공식 조회(미체결·종결 목록 + 보유·매수가능) — 상한은 design D2 표(최대 5분, 직렬). 기존 retry matrix 의 조회
  엔드포인트만 씀. 우선순위는 기본(Normal) — 보호 청산 제출(Critical)이 a166 대기열에서 앞선다.
- 원장 스키마 변경 없음. 롤백 = 이전 이미지(세션 해소가 쓴 전이는 기동 복구가 쓰는 것과 같은 전이라 되돌릴 것 없음).
- **운영 결과 공시(정본 SHALL)**: 세션 중 park 는 진입 게이트에 `ReasonUnresolvedInDoubt` 래치를 건다 — **계정 전체의 신규 진입이 그 엔진 프로세스가 끝날 때까지
  막힌다**(래치를 푸는 production 호출자 0, 운영자 해동은 다른 프로세스라 이 래치를 못 푼다 — design D5). 청산은 막지 않는다. 지금도 기동 복구의 park 는
  같다; 바뀌는 것은 그것이 **세션 중에도** 일어난다는 것이다.
