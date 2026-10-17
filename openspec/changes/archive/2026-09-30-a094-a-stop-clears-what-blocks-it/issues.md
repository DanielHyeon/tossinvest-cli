# a094 · issues

> 이 파일은 a094가 **고치지 않고 남기는 것**과 **다른 change의 선행 조건이 되는 것**을
> 기록한다. 침묵한 생략을 만들지 않기 위한 자리다.

## I1. 부재 확증의 증거 모델은 매도에 대해 아무것도 증명하지 못한다

**tasks 5.3의 기록 항목. R4 철회(1라운드 차단 2)의 잔여물이다.**

`execgw.absenceCorroborated`(`internal/execgw/indoubt.go:486-500`)의 증거는 두 가지 —
보유수량 변화와 매수가능금액 delta — 이고, 후자의 판정식이 **매수 예약 모델**이다:

```go
notional := intentNotional(intent)
bpDelta := buyingPowerNow - baseline.BuyingPower
if notional > 0 && bpDelta < 0 && math.Abs(bpDelta) >= notional*0.5 {
    return false, "... the buying power dropped ... consistent with this order having been accepted"
}
return true, "the holding and the buying power are unchanged from the pre-dispatch baseline"
```

**접수됐으나 미체결인 매도는 두 값을 모두 움직이지 않는다.** 따라서 위 `if`가 걸리지
않고 함수는 `true`("부재가 확증됨")를 반환한다 — 그 매도가 브로커에 살아 있어도.

보호 청산은 전부 매도다. 즉 **이 증거 모델은 이 저장소가 가장 자주 내는 주문 종류에
대해 아무것도 증명하지 못한다.**

**오늘 그것이 사고로 이어지지 않는 이유**는 `Baseline`을 넘기는 호출자가 사실상 없어
(`DecodeBaseline`이 항상 실패) 판정이 **항상 park**로 끝나기 때문이다. 무지가
안전측으로 표현되고 있는 것이지, 모델이 옳아서가 아니다.

**따라서 선행 조건**: 매도 mutation에 사전 계정 기준선을 공급하려면 **매도용 부재 증거
모델이 먼저 정의되어야 한다.** 후보는 「체결 이벤트 부재 + OPEN·CLOSED 목록 완주의
결합」이며, 그 자체가 별도 change의 주제다. 그 모델 없이 기준선만 공급하는 것은
**「모름」을 「없음」으로 바꾸는 것**이고, 그 결과는 살아 있는 매도 위의 두 번째 매도다.

a094는 이것을 spec에 SHALL NOT으로 못 박는 것까지만 한다
(`specs/order-execution/spec.md`「부재 판정의 증거 모델은 매도에 대해 성립해야 한다」).

**추가 함정**: 미측정 필드를 0으로 채우는 것도 금지다. `BuyingPower = 0`이면
`bpDelta > 0`이 되어 가드가 **영원히 발화하지 않고**, spec §3의 교차 확인 절반이
침묵으로 만족된다. 0은 "변화 없음"과 구별되지 않는다.

## I2. `classifyRefusalBody`의 message 매칭은 취약한 채로 남는다

`internal/execgw/failclosed.go:221-238`의 기존 세 항목은 code 토큰과 **한국어 message
조각**을 같은 `containsAny`에 묶는다:

| reason | 매칭 문자열 |
| --- | --- |
| `ReasonInteractiveAuthRequired` | `trade_auth_required` · **`interactive`** · `거래 인증` |
| `ReasonFXConsentRequired` | `fx_consent` · `exchange_consent` · `환전 동의` |
| `ReasonFundingRequired` | `funding_required` · `insufficient_deposit` · `입금` |

`"interactive"`는 code 토큰이 아니라 아무 본문에나 나올 수 있는 영어 단어이고,
매칭은 **본문 통짜에 대한 substring**이다. D0의 표가 보이듯 message는 계약과 어긋날 수
있으므로 message 매칭 자체가 같은 종류의 취약점이다.

**a094는 이것을 고치지 않는다.** 지우는 방향은 지금 잡히던 것을 놓치는 방향이고,
이 change는 보수 방향만 취한다(§6). **새 항목을 같은 방식으로 만들지 않는 것**까지가
a094의 범위다.

후속 change의 조건: 기존 세 항목을 code 필드 파싱으로 옮기려면, 각 항목이 실제로 어떤
code로 오는지의 **실물 응답**이 먼저 있어야 한다. `testdata/`의 세 fixture는 최상위
`code`를 갖지만 그것이 프로덕션 모양인지는 미확인이다 —
**프로덕션 409 3건은 `error.code`였다.**

## I3. 재생 결과 (tasks 6.1 · 6.1a · 6.2 · 6.3 — 2026-09-30, fixture)

`internal/app/engine/a094_replay_test.go` 가 세 모양을 재생했다(운영 원장 · 브로커 호출 0).

| 사건 | 재생 | 결말(이 change 의 기대) |
|---|---|---|
| 새 409(`6GKYatiUehps5SQX` · `7d3we7ZD3dtxWTMO` · `7k5oRgmEHnoU5Vfi` 의 모양) | `TestA094ReplayTheNew409BecomesAReportedRepetition` + 게이트웨이 끝의 실제 HTTP 409(`TestA094RefusalCodeClassifiesTheAttempt`) | FAILED_CONFIRMED → 판정 함수가 발의를 풂 → 다음 관측 재발의 → 같은 409 → 매 주기 거절 **보고**. 반대 매수는 엔진 밖 주문이라 치우지 않음(3.X) — **손절은 사람이 반대 주문을 치울 때까지 나가지 않는다** |
| park 된 두 행(475150 level 0 · 080220 level −1) | `TestA094ReplayTheTwoParkedRows` | 발의 무장 유지 · 제출 0 · park 원인 critical 1(판정 진입). 해동 명령(`tossctl engine attempt-resolve --target FAILED_CONFIRMED`) 뒤 6.1 과 같은 결말 |
| 272210 라이브락(PROPOSAL_CANCELLED 1931건, 중앙값 5.0초) | `TestA094ReplayThe272210LivelockStops` | **셋째 기전(D−4.7)으로 재현** — 무장 발의 없음 + 같은 종목 다른 intent 의 IN_DOUBT(옛 409 손절). a094 뒤 3분(36 주기) PROPOSAL_CANCELLED 0, 지연 경보 1. 인과 확정: 청소가 미종결을 안 보게 한 변이(M9)에서 이 시험이 반복으로 실패(변이 원장) |

**상호작용.** a087(보호 청산 시장가): 빈 가격 주문도 치움 대상(3.B3) — 충돌 없음. a089: 아카이브(64a1b2b3), R1 규범 충돌 해소. a091: 알림 등급과 독립.
a092: 새 critical 은 전부 a092 입구(`RecordCritical`, 창 0)로 — 막힌 전송자에서 다른 포지션 손절 무지연을 실측(3.R4). a090: 판정 불가(시세 없음) 포지션의 종결
대기 알림은 a090 미관측 경보가 덮는다(3.R9a — a090 착지 뒤 통합 시험, 이 로트 미실행).

## I5. 후속 기록 (이 change 에서 구현하지 않음)

- **나머지 ACKED 정산(4.N4x)**: 기동은 기록 번호 바이트 일치만 확정한다. 그 밖의 ACKED 를 풀려면 해소기 matcher 에 주문 번호 판별자(기록된 `broker_order_id` 와
  바이트 일치 요구)가 먼저 있어야 한다 — 좌표 `internal/execgw/indoubt.go` 의 matcher 필드 목록(D−4.2 인용 당시 `:638-650`)과 단일 일치 덮어쓰기
  `res.BrokerOrderID = order.OrderID`(당시 `:307`). 반례 시험: 오답 단일 일치 · 복수 일치.
- **UNKNOWN_BROKER_STATE 손 해소(D−9.4)**: 명령이 저장소에 없다 — 이름 붙인 후속 change 후보.
- **3.X 엔진 밖 주문 취소**: 사용자 결정 대기(사람 항목).
- **4.3f 정지 조건 — 미발동**: 구현이 「rate-limit 으로 park 된 attempt」 를 타입으로 가를 필요가 없었다(park 원인 알림은 attempt 를 이름으로 댈 뿐 원인을 분류하지 않음).
- **intent 없는 무장 발의**: 옛 판본이 `pending_intent_id` 없이 무장한 행은 청소 · 따라잡기가 풀지 않는다(attempt 를 찾을 수 없어 살아 있음으로 다룸). 운영 원장에
  그런 행이 있는지 배포 전 읽기 전용으로 센다(8.2 와 함께).

## I4. 현재 얼어붙은 포지션은 이 change가 소급 보호하지 않는다

배포 전까지 475150·080220·272210은 **사람이 처리한다**(tasks 8.4).
`checkSymbolFree`가 미정산 attempt를 이유로 **취소까지** 막으므로,
미체결 매수 주문을 앱에서 취소해도 엔진의 손절은 나가지 않는다 —
푸는 것은 엔진 재시작(그 시점의 recovery가 park시킨다)이거나 사람의 직접 매도다.

**엔진 재시작은 사람이 승인한다**(tasks 8.2).
