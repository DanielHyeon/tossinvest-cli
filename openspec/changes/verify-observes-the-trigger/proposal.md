# Proposal: verify-observes-the-trigger

## Why

**2c-A(보호 주문 원장)의 유일한 남은 선행 조건이다.** task 2.5의 미측정 목록에서 2c가 실제로
기대는 항목은 발동 하나다(`verify-execution-capability/measurements.md` "2.5에 남은 미측정").

지금까지 양 시장에서 측정된 것은 **등록·조회·존속·정정·취소·예약 여부**다. 측정되지 않은 것은
**그 조건주문이 실제로 발동해 매도가 체결되는가**다. 그것 없이 2c를 구현하면, 원장이 기록하는
"보호"가 브로커에서 실제로 무엇을 하는지 한 번도 본 적 없는 상태로 실계좌 손절을 맡기게 된다.

### 이 측정만 요구하는 것 — 다른 어떤 단계와도 다르다

`internal/verifylive/pricing.go`의 첫 문장이 이 도구의 전제를 적어 뒀다.

> *"pricing.go decides the one number that makes this tool safe to run: the price of an
> order that must not fill."*

발동 측정은 **그 전제를 의도적으로 반전한다.** 임계는 넘겨야 하고, 발동이 만든 child 시장가
매도는 **체결되어야 한다.** 체결되지 않으면 측정할 것이 없다. 그래서 이 change는 지금까지의
어떤 change보다 위험하고, 아래 네 가지를 **분리해서** 다룬다.

### 지금 코드가 이 측정을 담을 수 없는 네 지점

**G1 — 넘길 의도의 임계를 만드는 함수가 없다.** `FarStopTrigger`는 `FarBuyLimit`에 위임해
"시장 아래 20%, 밴드 안쪽"을 만든다. 발동을 관측하려면 그 반대가 필요하고, 기존 함수에
방향 플래그를 붙이면 **모든 단계의 안전 산술이 한 분기 뒤에 놓인다.**

**G2 — 체결된 객체를 표현할 수 없다.** `Outstanding`([record.go](../../../internal/verifylive/record.go))은
`Cancelled`가 아닌 artifact를 전부 살아 있다고 본다. child 매도가 **체결되면** 그것은 취소된
적이 없으므로 **영원히 outstanding**이다. 결과는 셋이다 — `liveCount`가 `MaxLiveOrders`(1)를
영구히 채워 이후 mutating 단계가 전부 `ErrExposureCap`, `verify status`가 존재하지 않는
주문을 계속 잔여물로 출력, 다음 실행의 정리 prologue가 **체결된 주문의 취소를 승인 목록에
올린다.** 취소·체결은 둘 다 "이제 살아 있지 않다"인데 기록에는 취소만 있다.

**G3 — 기다리는 동안 운영자가 사슬을 끝낼 방법이 없다.** `verify-holds-what-it-awaits`의 I1이다.
발동은 임계를 넘길 때까지 오지 않고, 안 올 수도 있다. 그때 계좌에는 살아 있는 조건주문이 남고
`HeldUntil`이 정리로부터 그것을 보호하므로, **측정을 포기하는 경로가 없으면 붙잡힌 채 끝난다.**
M37을 벗어날 수 없게 만들었던 구조와 같고, 차이는 이번엔 그 상태가 보인다는 것뿐이다.

**G4 — 시각을 브로커에서 읽을 수 없다.** M44: `lastExecutedAt`이 28건 전 건 null이고, US
주문의 `orderedAt`은 날짜만 담는다. M45: `Order.SubmittedAt`은 실제로 `version`(레코드 갱신
시각)이라 접수 시각과 6시간까지 어긋난다. → **네 시각은 전부 이 도구의 관측 시각이어야 하고,
폴링 간격이 그 오차의 상한이다.** 그 간격을 기록에 함께 남기지 않으면 지연 수치가 해석 불가다.

### 왜 지금인가

계측기가 확보됐다. M49로 TSLA를 골랐고(틱 간격 0.18초, 스프레드 0.0200%), M43으로 US 당일
매수분의 당일 매도가 확인돼 "child 매도 거절"을 결제 사유와 혼동할 위험이 US에서는 없다.
`verify-holds-what-it-awaits`(`f0afbcc`)가 child 주문을 정리로부터 지킬 수단을 이미 만들었다.

## What Changes

### 1. 넘길 의도의 임계 — 기존 안전 산술과 분리한다

- `NearStopTrigger(last, bid float64, market string) (SafePrice, error)` 신설. **기존
  `Far*` 함수군은 서명·본문 무변경**이며 방향 플래그를 받지 않는다.
- 임계는 **최근 체결가와 최우선 매수호가 사이**에 놓는다. 두 값이 갈리는 자리에 놓아야
  브로커가 조건을 체결가로 보는지 호가로 보는지(`trigger_price_basis`)를 사후에 좁힐 수 있다
  (I3). 사이에 유효한 틱이 없으면 **에러를 반환하고 단계는 skip** — 추측하지 않는다.
- 수량은 `MinQuantity`(1주), 유형은 **SINGLE + MARKET + SELL 고정**. OCO·OTO·LIMIT·복수
  수량은 이 change가 만들지 않는다.

### 2. 체결을 취소와 나란한 종결 상태로 만든다

- `Artifact.Filled bool` + `FilledAt`(가산·`omitempty`). `Outstanding`은 `Cancelled`
  **또는** `Filled`인 줄을 제외한다.
- `FormatVersion` 무변경 — 기존 줄은 두 필드가 없어 오늘과 같은 판정을 받는다(§0.6).
- 이것 없이는 발동 측정이 성공할수록 도구가 망가진다(G2).

### 3. `conditional-trigger` 단계 — deferred를 걷어내고 실제로 관측한다

- **옵트인 필수**: `--include-trigger`. `FlagIncludeTTLEdge`와 같은 방식이며, 붙이지 않으면
  단계는 오늘과 똑같이 미검증으로 남는다.
- 네 시각을 이 도구의 관측 시각으로 기록한다 — `condition_crossed_at`,
  `trigger_first_observed_at`, `triggered_order_id_first_seen_at`, `child_order_filled_at`.
  **폴링 간격을 함께 기록**하고, 그것이 각 시각의 오차 상한임을 관측값에 명시한다(G4).
- 발동 최초 관측 시점의 **bid/ask/last를 함께 기록**한다(`trigger_price_basis`의 근거, I3).
- child 주문을 `HeldUntil = conditional-trigger` + 사슬의 `ChainID`로 붙잡아 다음 실행의
  정리가 체결 전에 취소하지 못하게 한다(`f0afbcc`가 만든 수단).
- 체결을 확인하면 child artifact를 `Filled`로 종결한다.

### 4. 붙잡힌 사슬을 운영자가 끝내는 경로 (I1)

- `tossctl verify abort` — 이 도구가 붙잡고 있는 사슬을 **승인 목록 위에서** 끝낸다. 살아 있는
  조건주문을 취소하고, 사슬을 기록상 종결하며, 무엇을 취소하는지 실행 전에 나열한다.
- **콘솔에 새 타이핑 확인·추가 승인 마찰을 넣지 않는다.** 기존 승인 레일을 그대로 쓴다.
- 시각 기반 자동 만료로 대체하지 않는다(`verify-holds-what-it-awaits` design.md D3).

### 5. 노출 상한 — I2의 세 선택지 중 ③

- `MaxLiveConditionalsTrigger`(단계 전용 상한). `MaxLiveOrdersTTLEdge`가 이미 쓰는
  방식이고, 일반 상한의 의미("이 도구가 계좌에 만든 살아 있는 노출")를 약화하지 않는다.
  상한의 종류가 주문이 아니라 조건주문인 이유는 design.md D6 — child 주문은 이 도구가
  접수하는 것이 아니라 브로커가 만들고 우리가 발견하는 것이라 주문 상한을 지나가지 않는다.
- 2번(체결을 종결로 표현)이 들어가면 **체결 후에는 상한이 자연히 비므로**, 이 전용 상한이
  필요한 구간은 발동 관측 창 하나뿐이다.

## Non-Goals

- **`ProtectiveCapability` 산출** — 근거 기반 enum이고 근거는 이 change가 만드는 **실측 실행의
  출력**이다. 측정이 돌기 전에는 산출할 값이 없다. 사용자 계획의 ③이며 별도 change다.
- **2c-A 라이브 배선** — 이 측정이 실제로 한 번 돌기 전에는 스키마·상태모델·RED만 허용된다.
- **만료·부분체결·OCO sibling·조건주문과 일반 매도 동시 제출** — 2.5의 나머지 미측정. 2c-B.
- **KR 발동 측정 실행** — 도구는 시장 무관하게 만들지만, 이 change의 실측 실행은 US 1회다.
  KR은 계측기 선정부터 다시 한다(333430은 후보 유니버스 관측 0이라 유리하나 틱 간격 1.8초).
- **M45의 `Order.SubmittedAt` 교정** — High-risk 읽기 경로(`matchesDelayedCancelRecoveryHint`)에
  닿는다. 이 change는 그 필드를 쓰지 않을 뿐이고, 교정은 별도 change다. issues.md에 남긴다.

## Impact

- 위험도 **High-risk**. 이 도구가 처음으로 **체결을 의도한 라이브 주문**을 만든다.
- 영향 파일: `internal/verifylive/{pricing.go,record.go,runner.go,steps.go,verifylive.go,mutate.go}`,
  `cmd/`(abort 하위 명령), 콘솔 표시.
- 기록 스키마: 가산 nullable 2필드(`Filled`, `FilledAt`), `FormatVersion` 무변경.
- **계좌에 남는 변화**: 성공하면 TSLA 1주가 시장가로 매도된다. 되돌릴 수 없다. 사용자가 그
  1주를 이 측정을 위해 매수했고(2026-07-30), 실행은 사람이 그 자리에 있을 때만 한다.
- **토글 OFF 동일성**: `--include-trigger` 없이는 단계가 오늘과 같이 미검증으로 남고, 새
  pricing 함수는 호출되지 않는다. 기존 12개 단계의 산술과 판정은 무변경이다.
