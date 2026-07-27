# Change: apply-us-measurement-fixes

## Why

2026-07-27 US 1차 측정(run-M6WQZ5WKGGE4KS4C)이 실측 두 건을 남겼고, 둘 다 **다음 실행을
막고 있다**(measurements.md M16·M17).

- **M16**: 주문 취소가 `409 already-processing`으로 거절될 수 있다 —
  `{"code":"already-processing","message":"지금은 주문을 변경할 수 없어요. 잠시 후 다시
  시도해주세요.","data":{"retryAfterSeconds":1}}`. 브로커가 **재시도 힌트를 준다**. 도구는
  취소를 재시도하지 않아 order-cancel이 fail로 끝났고, 취소되지 않은 주문 1건이 노출
  상한(1건)을 채워 **order-amend·sell-boundary가 연쇄 차단**됐다. 지금 계좌에 그 주문이
  살아 있고, 재개 실행이 마지막에 이것을 취소하려 할 때 같은 409를 만나면 또 실패한다.
- **M17**: 조건주문 목록의 `status` 필터는 `OPEN`/`CLOSED`만 허용한다(`400
  invalid-request`, `data.field="status"`, `allowedValues=["OPEN","CLOSED"]`). 코드가
  `"WATCHING"`을 보내고 있다 — 일반 주문의 M7과 같은 계열의 결함이다.

## What Changes

- **취소에 한정한 유한 재시도**: `409 already-processing`에 한해 취소를 다시 시도한다.
  브로커가 준 `data.retryAfterSeconds`를 대기값으로 쓰고(상한 5초), 최대 2회 추가 시도한다.
  적용 범위는 **취소뿐**이다 — 주문 접수·정정·조건주문 생성은 종전대로 자동 재시도가 없다.
- **조건주문 목록 필터 교정**: `WATCHING` → `OPEN`. 문서상 `OPEN`은 "진행 중(감시 중·
  일시중지·주문 진행 중 포함)"이므로 이 단계가 확인하려던 집합과 같다. 관측 기록의 의미
  (등록한 조건주문이 목록에 나타나는가)는 보존된다.
- 두 수정 모두 **관측 기록에 정직하게 남긴다**: 재시도 횟수와 최종 결과를 기록하고,
  재시도로 성공한 취소는 "첫 시도에 성공"으로 위장하지 않는다.

## Non-Goals

- 접수·정정·조건주문 생성의 자동 재시도 — 금지 유지(2b task 1.7③).
- 429(rate limited) 정책 변경 — 읽기 전용 백오프는 그대로.
- 취소 실패 시 사람 승인 없는 다른 대체 행동 — 없다. 재시도가 모두 실패하면 종전처럼
  fail로 기록하고 잔여물을 화면·기록에 남긴다.

## Capabilities

### Modified Capabilities

- `order-execution`: "주문 오류 분류" — `already-processing`(409)를 재시도 가능 분류로 성문화

## Impact

- Affected code: `internal/verifylive/mutate.go`(cancelOrder 재시도),
  `internal/verifylive/steps.go`(조건주문 목록 필터)
- 안전 검토(§0): 취소는 **노출을 줄이는 방향**의 유일한 mutation이며(§0.3이 취소 앞에
  확인을 두지 못하게 하는 것과 같은 근거), 재시도는 이미 승인된 같은 취소를 반복한다 —
  새 노출을 만들지 않고 승인 범위를 넓히지 않는다. 대기는 주입된 clock seam을 쓰므로
  테스트가 실제로 자지 않는다. 상한(2회·5초)이 있어 무한 재시도가 되지 않는다.
