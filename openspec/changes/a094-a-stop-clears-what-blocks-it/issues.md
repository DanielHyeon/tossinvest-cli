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

## I3. 배포 재생 결과 (tasks 6.3 — 미실행)

`[미실행]` — tasks 6.1·6.2가 fixture 재생을 만든 뒤 여기에 결과를 적는다.
a087·a089·a091·a092와의 상호작용도 같은 자리에 적는다.

## I4. 현재 얼어붙은 포지션은 이 change가 소급 보호하지 않는다

배포 전까지 475150·080220·272210은 **사람이 처리한다**(tasks 8.4).
`checkSymbolFree`가 미정산 attempt를 이유로 **취소까지** 막으므로,
미체결 매수 주문을 앱에서 취소해도 엔진의 손절은 나가지 않는다 —
푸는 것은 엔진 재시작(그 시점의 recovery가 park시킨다)이거나 사람의 직접 매도다.

**엔진 재시작은 사람이 승인한다**(tasks 8.2).
