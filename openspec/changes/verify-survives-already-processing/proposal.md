# Change: verify-survives-already-processing

## Why

2026-07-28 00:20 KST의 US 재측정(run-IXCQU5UBZE)이 두 가지를 한꺼번에 보여줬다.

**좋은 쪽**: `order-cancel`이 **pass**로 바뀌었다. 취소가 `409 already-processing`으로 거절됐고
`apply-us-measurement-fixes`의 재시도가 실계좌에서 작동했다 — `order.cancel.retries=1`,
두 번째 시도에서 수락. M16 교정이 실증됐다.

**나쁜 쪽 — 새 실측**: 같은 `409 already-processing`이 **정정(modify)에서도** 온다.
`POST /api/v1/orders` 수락(107ms) 직후 `POST /api/v1/orders/{id}/modify`가
`{"code":"already-processing","data":{"retryAfterSeconds":1}}`로 거절됐다(requestId
`si1tUiUvi8DzWXr5`). 즉 이 코드는 취소 전용이 아니라 **방금 접수한 주문에 대한 모든 후속
변경**에 걸린다 — 브로커가 접수를 처리하는 동안의 짧은 창이다.

그 결과 두 단계가 연쇄로 막혔다.

- `order-amend` **fail**: 정정이 거절됐고, 재시도 규칙이 취소에만 있어 그대로 실패했다.
- 그 단계가 **자기가 낸 주문을 취소하지 않고 반환했다** — 성공 경로에만
  `cancelLiveOrders`가 있다(steps.go의 조기 반환). 남은 주문 1건이 노출 상한을 채웠고
- `sell-boundary` **fail**: `ErrExposureCap`. 아무것도 보내지 못했다.

두 번째 항목이 더 큰 결함이다. "이 도구가 만든 객체는 모두 취소되어 끝난다"는 불변식이
**실패 경로에서 성립하지 않는다**. 실패한 단계마다 잔여물이 하나씩 쌓이고, 그때마다 다음
실행의 정리 prologue가 치워야 한다(`verify-clears-leftovers`가 그 뒤처리를 하고 있을 뿐이다).

## What Changes

- **일시적 거절 재시도를 정정까지 넓힌다.** 조건은 종전과 **똑같이 좁다**: HTTP 409 **그리고**
  본문 `code == "already-processing"`. 대기는 브로커의 `retryAfterSeconds`(상한 5초), 추가 2회.
  재시도 횟수를 `order.amend.retries`로 기록한다.
- **주문 접수·조건주문 생성의 자동 재시도는 계속 금지한다.** 그 둘은 반복하면 두 번째 주문이
  생길 수 있다. 취소와 정정은 생성 연산이 아니다.
- **단계가 실패해도 그 단계가 낸 주문을 취소한다.** 단계 본문이 오류로 끝나면 러너가 그
  단계의 미취소 산출물을 같은 게이트·같은 계획 줄로 정리한다. 계획 밖 요청(`ErrOutsidePlan`)과
  컨텍스트 취소에서는 하지 않는다 — 그때는 아무것도 보내지 않는 것이 규칙이다.

## Non-Goals

- 접수·조건주문 생성 재시도 — 금지 유지.
- 409 전체를 재시도 대상으로 삼기 — `already-processing` 본문 코드가 일치할 때만이다.
- 접수 후 고정 대기 삽입 — 관측값 1건(`retryAfterSeconds:1`)을 상수로 박으면 측정하지 않은
  것을 단언하게 된다. 브로커가 거절하며 알려주는 값을 쓴다.
- 노출 상한 완화 — 무변경.

## Capabilities

### Modified Capabilities

- `order-execution`: `already-processing`은 취소뿐 아니라 **정정**에도 적용된다 / 실패한
  단계도 자기 산출물을 정리한다

## Impact

- Affected code: `internal/verifylive/mutate.go`(amendOrder 재시도),
  `internal/verifylive/runner.go`(실패 단계 sweep)
- 안전 검토(§0): 정정은 **새 노출을 만들지 않는다** — 이 도구의 정정은 체결 불가 지정가를
  시장에서 **한 호가 더 멀리** 옮기는 것이고(`OneTickFurther`), 주문을 생성하지 않고 대체한다.
  재시도는 이미 승인된 동일 요청(같은 주문 ID, 같은 목표가)의 반복이다. 실패 단계 sweep은
  취소만 하며 같은 `r.gate`·`Plan.Authorises`를 통과한다 — 계획에 취소 줄이 없는 단계에서는
  전송되지 않는다.
