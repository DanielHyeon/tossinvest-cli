# a129 design — 청산 주문의 생애 줄

기준: base `381e73da`(`base-commit.txt`). 줄 번호는 base 의 것이다. 분기 주장은 이 change 의 AST 번들
(`analysis/function-logic/*/ast.json`)의 열거를 근거로 한다. 2판(2026-10-12): proposal-freeze 적대 리뷰(review.md R-1~R-18) 반영.

## D1. 계좌 원문 — `:1455` 의 인자는 계좌를 싣지 않는다 (영수증)

`engineRuntime` 주석 「관측자 전체 Log 는 넣지 않음 — 기존 줄이 계좌 원문을 실음」을 base 코드로 대조했다.
`ExitObserver` 가 `opts.Log` 에 쓰는 자리는 `o.log` 와 `o.logErr` 둘을 거치는 **열** 곳이다
(`git grep -n 'o\.log(\|o\.logErr('`, 비시험 — `exitloop.go` 8 · `exit_held_proposal.go` 2).

| 자리 | 이벤트 | 싣는 인자 | 계좌 원문 |
|---|---|---|---|
| `exitloop.go:1455` (`submit` B7) | `order.confirmed` | `symbol` · `quantity` · `action` · `level` · `capped` | **없음** |
| `exitloop.go:622` | `exit.snapshot_quarantined` | symbol · position_id · 격리 버전 · selector revision · 고정 문구 | 없음 |
| `exitloop.go:735` · `:763` | `engine.started` | symbol · position_id · (adoption_id) · baseline · initial_risk · 고정 문구 | 없음 |
| `exitloop.go:1856` `logErr` ← `:414` · `:893` · `:898` · `:1837` | `exit.cycle_failed` · `engine.operating_mode` · 알림 이벤트 | **`obs.FieldAccount, o.opts.AccountRef`** 를 매 줄 덧붙임 + `err.Error()` 원문 | **있음** — 계좌 참조(`123-45` 모양, `engine_assembly_test.go:48`)와 하위 호출이 만든 오류 문구 |
| `exit_held_proposal.go:320` · `:337` (`warnEpisode`) | a094 에피소드 | `err.Error()` 원문 · 고정 문구 | 주석상 원장 키만 |

결론: `:1455` 의 인자는 계좌를 싣지 않는다. `Log` 를 켜면 같은 싱크의 `logErr` 줄이 **계좌 참조와 오류 원문을 더** 내보낸다.
그래서 `Log` 는 켜지 않고 a090 `UnobservedLog`·a091 `ZeroFloorLog` 와 같은 **좁은 싱크**를 둔다. 정당화는 「a129 는 계좌 정보를
더하지 않는다」이다.

**정정(리뷰 R-2).** 1판은 「engine.log 는 계좌가 없는 싱크」를 전제로 썼는데 거짓이다. 같은 엔진 로거에 이미 계좌 참조를 싣는 생산자가 있다:
`Context.ReconcileDriver` 는 `opts.Log` 를 `c.Log` 로 채우고(`reconcileloop.go:377-378`) 대사 주기 줄마다 `obs.FieldAccount` 를 싣는다
(`reconcileloop.go:538` · `:543` · `:565`). `Runtime.log` 는 모든 줄 앞에 계좌를 붙이고(`runtime.go:523`) `engineRuntime` 이 `Log: logger` 로
배선한다. **이것은 불변식 8 의 기존 잔여이며 이 change 의 범위 밖이다** — 별도 change 로 다뤄야 한다(완료 보고에 명시).

`position_id` 는 `journal.PositionID`(`position_projection.go:441`)의 sha256 앞 24 hex 이다. 소금 없는 해시이고 같은 줄에 시장·종목이 있으므로,
짧은 숫자 계좌 참조는 오프라인 대입으로 되돌릴 수 있다(리뷰 R-3). a090 줄이 이미 싣고 Manager 화이트리스트가 요구하므로 **싣되 잔여로 적는다**
— 위 기존 잔여가 고쳐질 때 함께 다시 볼 것.

## D2. 전용 로거 `ExitOrderLog`

- `ExitObserverOptions.ExitOrderLog *obs.Logger` — 선택. nil 이면 줄 대기열·체결 따라가기·원장 읽기 모두 0. 종전 동작·출력과 같다.
- 생산 배선: `cmd/tossctl/engine.go` `engineRuntime` 의 `engine.ExitObserverOptions` 리터럴에 `ExitOrderLog: logger` 한 필드
  (a090 `UnobservedLog` 와 같은 자리·같은 로거). `Log` 는 여전히 넣지 않는다 — a090 구조 핀이 계속 그것을 잰다.
- `Context.ExitObserver`(`exitwiring.go:319`)는 이 필드를 덮지 않고 통과시킨다(UnobservedLog 와 같은 규칙).
- 기존 `:1455` `o.log(obs.EventOrderConfirmed, …)` 는 그대로 둔다(토글 OFF = 종전). 생산에서 `Log` 가 nil 이므로 중복 줄은 없다.
  `Log` 와 `ExitOrderLog` 를 **둘 다** 같은 로거에 넣으면 `order.confirmed` 가 두 줄이 된다(시험·도구에서만 가능, 리뷰 R-14).
  참고: 그 기존 줄은 키 `level` 을 써서 slog JSON 의 `level` 과 겹친다 — 토글 OFF 규칙상 손대지 않음.

## D3. 줄의 형식 — 화이트리스트 타입 하나

줄은 `exitOrderLine`(`exit_order_log.go`) 하나에서만 만들어진다. 구조체 필드가 곧 화이트리스트이고 `args()` 가 그 필드만 인자로 바꾼다.

| 키 | 출처 | 비고 |
|---|---|---|
| `origin` | 고정 `"exit"` | 다른 생산자의 `order.*` 와 구분 |
| `symbol` · `market` | `m.position` | 공개 |
| `side` · `order_type` | 고정 `SELL` · `LIMIT` | `sellIntent` 가 만드는 유일한 모양 |
| `quantity` | 제출 수량(`submitQuantity`; B1 은 캡 전 수량) | |
| `price` | 지정가(관측가, 비면 기준선 — `sellIntent` 와 같은 규칙) | |
| `action` · `exit_level` | `proposal.Action` · `proposal.Level` | 키가 `level` 이 아닌 이유: slog JSON 의 `level` 과 겹쳐 읽는 쪽이 뒤의 것만 남김(`log.go:31-41` 과 같은 함정) |
| `observed_price` | 판정 관측가 | |
| `baseline` | `m.state.Baseline` | 이 주기가 읽은 보호선(손절선) |
| `capped` | `applyFloor` 의 캡 여부 | |
| `intent_id` · `position_id` · `attempt_id` · `order_id` | 원장·게이트웨이 키 | `position_id` 잔여는 D1 |
| `to_state` · `reason` | `out.State` · `out.Reason`, 브로커 전 거절은 고정 코드 | 토큰 거름 |
| `order_quantity` · `filled_quantity` · `average_price` · `terminal` | 체결·종결 줄만 — 원장 `FillSnapshotRecord` | |

**싣지 않는 것**: 계좌 참조(관측자·포지션·체결 기록의 `AccountRef`), 모든 `err.Error()`, `out.Detail`, `out.Result`(브로커 응답 본문), Guardian 결정, 잔고·토큰.

**토큰 거름.** `reason`·`to_state` 는 `^[A-Za-z_]{1,64}$` 이거나 빈 값일 때만 그대로 싣고, 아니면 `unrecognised`. base 의
`ReasonCode`·`AttemptState`·원장 사유 상수는 전부 영문자와 밑줄뿐이고(`git grep 'ReasonCode = "'`), `out.Reason` 은 원장 사유에서 옴
(`gateway.go:753`). 숫자를 막으므로 계좌 번호 모양이 사유 문자열에 실려도 통과하지 못한다.

**이벤트.**

| 결과 | 이벤트 | 레벨 |
|---|---|---|
| `Place` → `CONFIRMED` (B7) | `order.confirmed` | info |
| `IN_DOUBT` (B8) · attempt 있는데 입증된 비수용 아님 (B10) | `order.in_doubt` | warn |
| `UNRESOLVED_IN_DOUBT` (B8) | `order.unresolved_in_doubt` | warn |
| 기본 갈래 중 `FAILED_CONFIRMED` (B12) | `order.rejected` | warn |
| 브로커 전 거절 — 하한 산술 실패 B1 `floor_arithmetic_failed` · 하한 0 B2 `confirmed_floor_zero` · Guardian 거절 B3 `guardian_refused` · intent 부착 실패 B4 `attach_failed` · 주문 모양 실패 B5 `intent_invalid` · 종목 진행 중 B9 · 기본 갈래의 나머지 B12 | `order.not_dispatched` | warn |
| 원장 누적 체결이 마지막으로 쓴 값보다 큼 | `fill.observed` | info |
| 늘지 않은 채 종결(취소·만료·거부) | `exit.order_ended` (a129 신설) | warn |
| 원장 intent·체결 기록 읽기 실패(주문당 한 번) | `exit.fill_follow_failed` (a129 신설) | warn |

`fill.observed` 는 「양의 체결 증가」로 정의돼 있어(`event.go:62`) 무체결 종결·읽기 실패를 그 이름으로 세지 않는다(리뷰 R-7).
신설 두 이름은 `internal/obs` 가 이 change 의 편집 범위 밖이라 `exit_order_log.go` 에 둔다(선례 `cmd/tossctl/notificationsettings.go:153`).
등급표에 없으므로 normal 이다 — `obs` 로 옮길지는 Manager 판단.
`order.in_doubt`·`order.unresolved_in_doubt` 의 `severity` 필드는 등급표대로 `critical` 로 찍힌다. 로그 줄일 뿐 알림·outbox 를 만들지 않고,
engine.log 의 `severity` 를 읽는 코드는 없다(리뷰 R-17: `engineproc.go:219` `readLogTail` 은 기동 직후 실패 꼬리만 읽음).

예(가짜 값):

```json
{"time":"2026-10-07T02:08:08Z","level":"INFO","msg":"order.confirmed","event":"order.confirmed","subject":"order","severity":"normal","origin":"exit","symbol":"010170","market":"kr","side":"SELL","order_type":"LIMIT","quantity":"6","price":"18010","action":"STOP_LOSS_LADDER","exit_level":"0","observed_price":"18010","baseline":"18050","capped":false,"intent_id":"exit_3f…","position_id":"pos-9a…","attempt_id":"att-…","order_id":"0000123456","to_state":"CONFIRMED","reason":""}
{"time":"2026-10-07T02:08:13Z","level":"INFO","msg":"fill.observed","event":"fill.observed","subject":"fill","severity":"normal","origin":"exit","symbol":"010170","market":"kr","side":"SELL","order_type":"LIMIT","quantity":"6","price":"18010","action":"STOP_LOSS_LADDER","exit_level":"0","observed_price":"18010","baseline":"18050","capped":false,"intent_id":"exit_3f…","position_id":"pos-9a…","attempt_id":"att-…","order_id":"0000123456","to_state":"FILLED","reason":"","order_quantity":"6","filled_quantity":"6","average_price":"18010","terminal":true}
```

## D4. `submit` 은 줄을 쓰지 않고 대기열에 넣는다 — 기록만 더한다

AST(`internal-app-engine--exitobserver.submit/ast.json`, 분기 13 · return 10)의 모든 return 경로가 `noteExitOrder` 하나를 부른다. 규칙:

1. **명시적 갈래별 호출.** `defer` 로 줄을 내지 않는다 — `defer` 로 쓰면 `Place`·`release` 의 패닉을 삼켜 nil 반환(성공처럼 보임)이 될 수 있다(리뷰 R-4).
2. `noteExitOrder` 는 **메모리 append 만** 한다(쓰기 없음). 해제가 있는 갈래(B2 · B3 · B5 · B9 · B12)는 해제를 먼저 하고 그 반환값을 그대로 돌려준다.
   `Place` 앞에 서는 일은 0 이다.
3. 대기열은 `ObserveOnce` 의 `defer o.endExitOrderCycle(ctx)` 가 그 주기의 **모든 판정·제출·미관측 정리 뒤에** 쓴다(리뷰 R-9 — a090 과 같은 규칙:
   뒤 포지션의 손절 판정 앞에 서지 않음). 패닉으로 주기가 끝나도 defer 가 돌아 그때까지의 줄(예: 이미 나간 주문의 확정)은 쓰인다.
4. **거절 래치(리뷰 R-5, 구현 리뷰 I-1 로 고침).** 거절 갈래(브로커 전 거절·종목 진행 중·거부)는 포지션마다 같은 서명(이벤트·액션·단계·사유·상태)이
   **연속한 주기에** 이어지면 첫 줄만 넣는다. 그 포지션의 거절이 없는 주기가 한 번이라도 끼면(가격 회복 뒤 재이탈 = 새 에피소드) 주기 끝에 래치를 푼다 —
   1판 래치는 에피소드를 넘어 남아, 며칠 뒤 같은 거절이 0 줄이 되는 010170 과 같은 꼴이었다.
   해제된 발의는 다음 주기에 다시 발의되므로 래치가 없으면 막힌 손절이 5초마다 같은 줄을 영원히 쓴다(engine.log 는 회전 없음,
   `tools/engine-autostart.sh:121`). 확정·결과 미상이 오거나 서명이 바뀌면 다시 쓴다. 래치는 메모리(재시작하면 다시 한 줄).
5. 반환값·알림·지연 시계·해제 판정·`:1455` 는 바꾸지 않는다.

## D5. 체결 줄 — 주기 끝의 원장 읽기, 정규 신원으로

체결은 체결 감지가 원장에 쓰고(`filldetect.JournalLedger.Apply` → `Journal.RecordFill`), 체결 감지에는 로거가 없다. 체결 감지·원장은 편집하지 않고
(체결 감지 경로에 I/O 를 더하면 exit 가 기대는 SLO 를 건드림 — 리뷰 R-18), 관측자가 자기가 낸 주문만 따라간다.

- 확정(B7)·결과 미상(B8 · B10)이고 `out.BrokerOrderID` 가 있으면 **intent id** 를 키로 따라가기 항목을 만든다(메모리, I/O 없음 — 주문 번호를 키로 하면
  날을 넘어 재사용된 두 주문이 한 항목으로 겹친다, 구현 리뷰 I-7).
- 주기 끝에 항목마다: 처음 한 번 `Journal.LookupIntent(intentID)` 로 주문의 **정규 신원**(계좌·시장·거래일·종목·방향)을 얻고,
  그 신원으로 `Journal.LookupFillScoped`(원장 읽기, 브로커 호출 0) 한 번. 브로커 주문 번호는 재사용될 수 있으므로(`fills.go:195-197`)
  번호만으로 읽는 `LookupFill` 은 다른 날 같은 번호의 종결 기록을 이 주문의 체결로 잘못 읽고 따라가기를 끝낸다(리뷰 R-1 — 시험이 고정).
- 누적 체결 수량이 마지막으로 쓴 값보다 크면(`riskcalc.CompareDecimal`) `fill.observed`, 늘지 않은 채 종결이면 `exit.order_ended`(리뷰 R-6 —
  「손절이 한 주도 못 팔고 끝났다」), 종결이면 항목을 지운다. `ErrFillNotFound` 는 다음 주기에 다시 본다. intent 가 원장에 없으면 한 번 알리고
  항목을 지운다. 그 밖의 읽기 실패는 연속 실패당 한 번 `exit.fill_follow_failed` 를 쓰고 계속 따라간다(읽기가 한 번 성공하면 래치를 풂).
  `ctx` 가 끝났으면 읽지 않는다. 한 주기의 읽기 전체에 1초 시한(`exitFillReadBudget`)을 준다 — 원장 연결이 막혀도 다음 주기를 붙잡지 않는다(구현 리뷰 I-4).
- 항목 수명은 24시간(`exitFillFollowLimit`)이다. 그 안에 종결 기록이 안 오면(브로커에 닿지 않은 결과 미상 주문 등) `exit.fill_follow_failed`
  `reason: follow_expired` 한 줄 뒤 버린다 — 주기당 원장 읽기가 재시작까지 남지 않게 한다(구현 리뷰 I-2).
- 따라가기 오류는 주기 결과(`ExitCycle.Err`)에 들어가지 않는다 — 기록 실패가 손절 경로의 판정이 되지 않는다.

**잔여(명명).** ① 항목·래치는 메모리에 있다 — 확정 뒤 체결 전에 엔진이 재시작하면 그 주문의 체결 줄은 남지 않는다(확정 줄의 `order_id` 로
원장에서 찾을 수 있다 — 원장이 정본). ② 주기당 한 번 표본을 뜨므로 한 주기 안의 여러 체결 증가는 한 줄로 합쳐진다. ③ 결과 미상인데 주문 번호가
없던 주문(뒤에 해소기가 찾음)은 따라가지 않는다. ④ 24시간 안에 종결 기록이 안 오는 주문은 `follow_expired` 뒤 따라가지 않는다.
⑤ 엔진이 패닉으로 주기를 끝내면 defer 가 대기열을 쓰고 체결 읽기를 한다(시한 1초) — 다른 패닉은 삼키지 않는다.

## D6. 로깅 실패는 청산을 막지도 늦추지도 않는다 (불변식 4)

- 줄 쓰기(`writeExitOrder`)는 그 쓰기 하나만 감싼 `recover` 안에서 한다. Go 의 recover 는 패닉을 일으킨 프레임의 defer 에서만 받으므로
  관측 주기의 다른 패닉(`Place`·원장)은 삼키지 않는다 — 시험이 `Place` 패닉이 종전대로 빠져나감을 고정(리뷰 R-4).
- `slog` 는 쓰기 오류를 버린다(`obs.Logger.emit`). slog 핸들러는 defer 로 잠금을 푸므로 패닉 뒤 잠금이 남지 않는다(리뷰 R-18).
- 줄 쓰기·체결 읽기는 모두 주기 끝(D4-3)이라, 막힌 쓰기 대상도 그 주기의 어떤 제출·해제·판정도 늦추지 못한다. 다음 주기의 시작을 늦출 수는
  있다 — 같은 엔진 로거를 쓰는 a090 · a091 · 대사 · Runtime 줄과 같은 잔여다. 같은 로거의 다른 쓰기에는 recover 가 없으므로
  이 recover 의 생산 보호 범위는 a129 줄에 한정된다(리뷰 R-10).

## D7. 시험 전략

`internal/app/engine/a129_exit_order_log_test.go`(외부 시험 패키지, `newExitHarness` 재사용) + `a129_export_test.go`
(`FollowedExitOrdersForTest` — 항목은 메모리뿐이라 이 창으로만 잼) + `cmd/tossctl/a129_engine_runtime_exit_order_log_test.go`(구조 핀).

1. A 확정 · 부분 체결 · 같은 누적 재관측(줄 없음) · 전량 체결(종결) · 그 뒤 0 줄. B 같은 주기에서 확정 줄이 앞 주문의 체결 줄보다 먼저(판정 뒤 따라가기).
   C 양보 주기에도 따라가기.
2. D 갈래마다 정확히 한 줄 — B1 · B2 · B3 · B4 · B8(IN_DOUBT · UNRESOLVED) · B9 · B10 · B12(거부 · 미발송), 줄이 쓰이는 순간 해제가 이미 끝남.
3. E 계좌 센티널 — 계좌 참조·결과 `Detail`·`Result`·사유 문자열·체결 기록의 계좌 필드에 심고 확정·거부·결과 미상 + 체결 줄 전 출력에 없음.
4. F nil — `Log` 출력이 base 의 두 줄(`engine.started`·`order.confirmed`) 그대로, 새 키 없음, 따라가기 항목 0.
5. G 쓰기 대상 패닉·실패 — 확정은 무장 유지, 거부는 해제·재발의, 주기 오류 없음. K `Place` 패닉은 삼키지 않음.
6. H 다른 날 같은 주문 번호의 종결 기록 — 이 주문의 체결로 읽지 않음. I 무체결 종결 → `exit.order_ended`. J 같은 거절 반복은 한 줄.
7. 생산 배선 — `engineRuntime` 리터럴 키 census(`ExitOrderLog = logger` 있음 · `Log` 없음). `Context.ExitObserver` 통과.
8. 기존 청산·주문 시험 전부 green(`go test -race ./internal/app/engine/ ./cmd/tossctl/`).

9. 구현 리뷰 보강 — M 거절 래치는 에피소드를 넘지 않음 · N 끝나지 않는 따라가기는 24h 뒤 `follow_expired` · O 해제 실패가 종전대로 주기 오류로
   올라옴(B2 · B3 · B10 · B13, 원장을 닫아 해제 실패) · A 의 줄 지정가 = 실제 제출 지정가(`sellIntent` 규칙을 옮겨 적은 값의 고정 — 구현 리뷰 I-8).

변이(`analysis/harness/a129_mutate.py`, 사본·무변이 대조군 GREEN 먼저): 화이트리스트에 계좌(info · warn 두 길) · 배선 삭제 · nil 가드 삭제 · recover 삭제 ·
체결 증가 비교 반전 · 거절 래치 삭제 · 토큰 거름 무력화 · 주기 끝 defer → 직접 호출 · 종결 삭제 누락 · 정규 신원 → 번호만 · 무체결 종결 줄 삭제 ·
확정 줄 삭제 · 래치가 에피소드를 넘음 · 수명 무한 · 해제 실패 삼킴(B2 · B3 · B10 · B13 각각). B5 해제 실패 삼킴은 도달 불가(관측가는 늘 수, 블록 count 0)라
변이를 두지 않는다. 줄 쓰기 쪽 nil 가드만 지우는 변이는 `obs.Logger` 가 nil 안전(`log.go:191`)이라 **동등 변이**다 — 관측 가능한 nil 가드는 대기열·따라가기 쪽이고
`FollowedExitOrdersForTest` 로 잰다.

## 배포 (범위 밖 — 사람 절차)

1. Manager 가 게이트·푸시 후, 사람이 **KST 05:00~09:00**(두 시장 닫힘, 엔진 정지 = 손절 없음 창 규칙)에만 교체한다.
2. `make image CHANGE=a129-an-exit-order-leaves-a-line`(빌드+핀 한 명령, 수동 compose build 금지) → 정본 override 로 컨테이너 재시작.
3. 교체 전 main 과 이미지의 `SchemaVersion` 대조(이 change 는 스키마 무변경).
4. 교체 뒤 확인: `engine.log` 에 `engine.started` 줄이 나오고, 다음 청산 주문에서 `"origin":"exit"` 줄이 남는지 사람이 본다.
   에이전트는 교체·토글·주문을 실행하지 않는다.

롤백: 이미지 핀을 이전 것으로 되돌린다. 원장·설정 변경이 없으므로 데이터 롤백은 없다.
