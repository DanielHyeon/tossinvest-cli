# a129 — 엔진이 낸 청산 주문은 엔진 로그에 한 줄을 남긴다 (an-exit-order-leaves-a-line)

## Why

2026-10-07 02:08:08Z 엔진이 010170 손절을 정상 발동했다 — SELL LIMIT 6주 @18010, 전량 체결, 포지션 CLOSED.
그런데 운영 로그 `~/.config/tossctl/engine.log` 에는 이 주문·체결 줄이 **0건**이었고, 운영자와 Manager 가
"손절 미발동"으로 오판했다. 진실은 저널에만 있었다(Manager 가 저널·코드로 확인, 2026-10-12).

원인(현재 HEAD `381e73da` 에서 확인):

- `internal/app/engine/exitloop.go:1841` `func (o *ExitObserver) log(...)` 는 `o.opts.Log == nil` 이면 즉시 돌아간다.
  주문 확정 줄 `:1455` `o.log(obs.EventOrderConfirmed, ...)` 는 이 함수로만 나간다.
- `cmd/tossctl/engine.go:665-673` `engineRuntime` 은 관측자에 `Log` 를 넘기지 않는다(주석 「관측자 전체 Log 는 넣지 않음 —
  기존 줄이 계좌 원문을 실음」). 넘기는 것은 a090 전용 `UnobservedLog` 하나다. `exitwiring.go` `Context.ExitObserver` 도
  `Log` 를 덮지 않는다(덮는 것은 a091 `ZeroFloorLog` 뿐).
- 체결은 관측자가 아니라 체결 감지(`internal/filldetect`)가 원장에 쓰고, 체결 감지에는 로거가 아예 없다. 그래서 청산 주문의
  체결도 엔진 로그에 남는 자리가 없다.

손절이 났는지를 엔진 로그로 알 수 없으면, 운영자는 맞게 동작한 손절을 장애로 읽고 손으로 개입한다 — 그것이 이번에 일어난 일이다.

## What Changes

1. `ExitObserverOptions` 에 청산 주문 줄만 받는 전용 로거 `ExitOrderLog *obs.Logger` 를 둔다(a090 `UnobservedLog` 선례).
   `Log` 전체는 계속 켜지 않는다.
2. 관측자가 청산 주문을 낸 결과마다 그 로거에 한 줄을 쓴다 — 확정(`order.confirmed`), 결과 미상(`order.in_doubt`),
   브로커 거부(`order.rejected`), 브로커에 닿기 전 거절(`order.not_dispatched`).
3. 확정된(또는 결과 미상이지만 주문 번호가 있는) 청산 주문은 관측 주기 끝에 원장의 체결 기록을 읽어, 누적 체결 수량이
   늘 때마다 `fill.observed` 한 줄을 쓴다. 종결 기록을 보면 따라가기를 멈춘다. 원장 읽기만 하며 브로커 호출은 0 이다.
4. 줄의 필드는 **화이트리스트**(타입 하나)로만 만든다 — 계좌 참조·원문 오류·브로커 응답 본문은 싣지 않는다.
5. `cmd/tossctl/engine.go` `engineRuntime` 이 `ExitOrderLog: logger` 하나만 배선한다.

## What Does NOT Change

- 판정·기준선·발의·주문·해제·취소의 동작과 순서. 브로커 호출 수(0 추가).
- `Log` 의 배선(생산에서 계속 nil)과 기존 `:1455` 줄. 전용 로거가 nil 이면 관측자의 동작·출력·원장 읽기가 종전과 같다.
- 체결 감지(`internal/filldetect`)·원장(`internal/journal`)·게이트웨이(`internal/execgw`)·`internal/obs` 의 코드.
- 엔진 로거의 출력 대상(`engine.log`)과 형식.

## Impact

- Affected specs: `exit-policy`(요구 1개 추가).
- Affected code: `internal/app/engine/exitloop.go`(옵션 필드 · `submit` · `ObserveOnce`), 새 파일
  `internal/app/engine/exit_order_log.go`, `cmd/tossctl/engine.go`(`engineRuntime` 리터럴 필드 하나).
- High-risk: 주문 제출 경로(`ExitObserver.submit`)와 exit 관측 주기(`ObserveOnce`)를 **기록만 더하는 방향으로** 만진다.
  Pre-Edit 선언 + Function Logic Map 필수.
- 배포: 사람이 KST 05:00~09:00(두 시장 닫힘)에 이미지 재빌드·교체(design 「배포」).
