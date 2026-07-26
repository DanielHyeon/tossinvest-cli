# issues — verify-execution-capability

> WORKFLOW 예외 경로 §분류: ① blocking ② safe local ③ editorial. 아래는 구현 중 발생한 편차와
> Manager 판단이 필요한 항목의 기록이다.

## 2026-07-26 · task 1.5 (`tossctl verify run`) 구현

### ② safe local — `official.ModifyConditionalOrderRef` 신설

**배경**: 2.5 는 조건주문 정정이 "신규 ID 발급·기존 ID 무효화"인지 실측하도록 요구한다.
기존 `official.Client.ModifyConditionalOrder` 는 응답 본문을 버려서(`postAcct(..., nil)`)
새 `conditionalOrderId` 를 알 수 없고, 목록 재조회로 유추하면 조건주문이 2건 이상일 때 추측이
된다.

**조치**: 기존 메서드를 **건드리지 않고** `ModifyConditionalOrderRef` 를 추가했다
(`internal/official/conditional_writes.go`). 기존 호출부(hybrid, app/engine)는 `nil` out 을
넘겨 null result 도 성공으로 처리하는데, 그 동작을 바꾸면 라이브 mutation 의 실패 표면이 넓어진다.
기존 동작 보존 테스트(`TestModifyConditionalOrderStillToleratesANullResult`)를 함께 추가했다.

`internal/app/engine/seal_test.go` 의 `officialMutators` 목록에 새 메서드를 추가했다 —
그 파일의 주석이 "새 mutator 는 의식적으로 추가하라"고 요구한다.

근거: `POST /api/v1/conditional-orders/{id}/modify` 설명, `docs/migration/openapi.latest.json`.

### ② safe local — `--redo <step-id>` 플래그 추가 (설계에 없던 표면)

**배경**: `read-fixtures`(2.1)는 계좌의 기존 주문 이력에서 status enum 을 수집한다. 주문 이력이
없는 계좌에서는 skip 된다. 그런데 이후 단계들이 주문을 만들어 이력을 남기므로, 되돌아갈 방법이
없으면 신규 계좌에서 2.1 이 영구히 미측정으로 남는다.

**조치**: `--redo read-fixtures` 로 기록에 판정이 있는 단계를 다시 돌릴 수 있게 했다. 재실행해도
mutation 은 여전히 매번 typed confirmation 을 거친다. skip 사유 문구가 이 플래그를 직접 안내한다.

### ② safe local — `--max-sell-quantity`(기본 1)와 "최소 오버셀"

**배경**: 2.2 의 매도 경계는 부분/전량/보유초과 3가지다. "전량"과 "보유초과"는 정의상 최소 수량
1주 규칙과 충돌한다.

**조치**:

- 전량 매도는 보유수량이 `--max-sell-quantity`(기본 1주) 이하일 때만 실제로 제출한다. 그보다 큰
  보유에서는 주문을 내지 않고 `sell.boundary.full_accepted = unverified` 로 정직하게 기록한다.
  누군가의 전체 포지션을 시장 위에 걸어두는 것은 이 도구가 할 일이 아니다.
- 보유초과는 `sellable + 1`주 — 가능한 가장 작은 초과 — 로 제출한다. 시장가 대비 멀리 떨어진
  지정가 매도이고, 브로커가 **수락하면** 즉시 취소한 뒤 단계를 FAIL 로 기록한다(청산 예약 공식이
  브로커의 경계 검사에 의존할 수 없다는 뜻이므로).

### ② safe local — 취소도 typed confirmation 을 받는다 (flatten 과 다름)

flatten 은 §0.3 에 따라 취소 앞에 프롬프트를 두지 않는다(취소는 노출을 줄이기만 하므로). 이 도구는
취소 자체가 측정 대상(2.2 의 cancel 경로, 2.5 의 조건주문 취소)이라 확인을 받는다. 프롬프트에
"direction: this REDUCES exposure" 를 명시해 위험 방향을 오해하지 않도록 했다.

### ① Manager 판단 필요 — 보유 없는 계좌에서 2.5 전체가 미측정으로 남는다

**사실**: SINGLE+MARKET 손절은 SELL 이므로 보유가 필요하다. 설계 지시는 "never buy-to-create a
holding automatically; skip-with-reason if none" 이고 그대로 구현했다. 결과적으로 보유가 0인
계좌에서는 `conditional-register / sellable-reserved / conditional-persist /
conditional-modify / conditional-cancel` 이 전부 skip 되고, **2.5·2.6 의 입력이 하나도 생기지
않는다**. attestation 의 조건주문 endpoint 집합도 비므로 게이트는 계속 닫힌다(의도된 방향).

**검토했으나 구현하지 않은 대안**: `--conditional-buy-fallback` — SINGLE + LIMIT + BUY,
triggerPrice 를 시장가보다 훨씬 **위**로(발동 안 됨), orderPrice 를 훨씬 **아래**로(발동해도 체결
불가). 보유 없이 조건주문 등록·조회·존속·정정·취소 endpoint 를 전부 검증할 수 있고 체결 위험은
이중으로 막힌다. 다만 (a) 설계 문구에 없고, (b) 사용자 계좌에 조건부 매수를 만드는 일이라 임의로
넣지 않았다.

**현재 안내**: skip 사유가 "KR 종목 1주 이상을 보유한 뒤 `--resume`(또는 `--holding-symbol`)"
을 지시한다. 사용자 협조 절차에 이 선행 조건을 넣을지, 위 fallback 을 opt-in 으로 추가할지 Manager
판단이 필요하다.

### 측정 불가로 확정된 항목 (report 가 `unverified` 로 노출)

| 항목 | 사유 |
|---|---|
| `idempotency.key_scope` (2.7 계좌 스코프) | 자격증명 1세트 = 계좌 1개. 키가 계좌를 넘는지 관측할 수단이 없다 |
| `conditional.trigger_observed` / `triggered_order_id_exposed` / latency (2.5) | 발동에는 시장이 가격까지 와야 한다. 체결을 의도한 주문을 이 도구는 내지 않는다 — [별도 세션 — 시장 조건 필요] |
| `idempotency.ttl_window_closed` (2.7 유효 창) | `--include-ttl-edge` 옵트인 전에는 미측정. 의도적 이중 주문 절차이므로 기본 생략 |
| US 시장 전체 | mutation 단계는 KR 심볼만 받는다. amend 의 quantity 필수 여부·일일 상하한가 모두 KR 규칙이고, US 규칙으로 대체 검증하면 잘못된 근거가 된다 |
| 정규장 밖 동작 (2.5) | 등록 시각의 KST 세션 라벨만 기록한다. 세션별 반복 실행은 사용자가 시간대를 골라 재실행해야 한다 |

### 2.9 (실측 비용) 는 이 도구만으로는 채워지지 않는다

모든 주문이 체결 불가 가격이므로 `execution.commission` 은 항상 null 이다. `costs` 단계는
`costs.collected = false` 와 "no verification order filled" 사유를 기록한다 — 0원이라고 쓰지
않는다. 2.9 를 채우려면 실제로 체결되는 주문이 필요하고, 그것은 이 도구의 안전 전제(체결 불가
가격) 밖이다. Manager 결정 필요: 사용자의 평소 거래 체결 내역에서 수집할지, 별도 절차를 둘지.
