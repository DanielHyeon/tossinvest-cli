## MODIFIED Requirements

### Requirement: Guardian 판정 체인
모든 자동 진입 의도는 Guardian 판정 체인을 통과해야 하며(SHALL), 체인은 고정 순서로 평가되고 첫 실패에서 정지한다: kill switch/운영 모드 → 게이트 상태(진입 차단 latch: 401/403·SLO 위반·RECONCILE·recovery 미완료) → 심볼 allowlist → 구조적 손절 계약(No Stop=No Trade·보호적 stop·실질 본전 미달 target 거부) → 주문 크기 한도 → 최소 RR(신규 검사) → 현금 검증(비용 포함) → 당일 재진입 규칙 → 총 개방 노출 한도 → 일일 손실 한도 → 중복 주문 검사 → ALLOW. 각 거부는 안정적 reason-code로 기록된다(SHALL).

**순서의 권위는 이 표다**(SHALL — StockOS `evaluate_guardian`과의 대응은 `docs/guardian-chain.md` 참고 산출물로 남기되 순서 보존을 규범으로 주장하지 않는다: 원 체인에는 미이식 단계가 섞여 있고 최소 RR은 원 체인에 없는 신규 검사다). 이식 분류의 열거(SHALL):

- **이식**: kill switch·모드, 진입 latch, 손절 계약(reason: STOP_MISSING·STOP_NOT_BELOW_ENTRY·TARGET_NOT_ABOVE_ENTRY·INVALID_TARGET_STOP·TARGET_BELOW_BREAK_EVEN), 주문 크기(INVALID_ORDER_SIZE·MAX_ORDER_EXCEEDED), 현금(INSUFFICIENT_CASH·비용 포함), 당일 재진입(쿨다운 기본 30분 `[미검증]`·당일 심볼당 최대 진입 수 기본 2회 `[미검증]`·**미체결 매수 존재 시 진입 차단 PENDING_BUY_ORDER_BLOCKED**), 총 노출, 일손실, 중복 주문
- **신규**: 최소 RR, 심볼 allowlist, STOP_MISSING(원본은 생성자 강제라 카운터파트 없음), 하루 매수 총액(DAILY_TURNOVER 의 매수 측 — 체인 단계가 아니라 결정·예약 원자 트랜잭션에서 판정한다: 「하루 매수 총액 한도」 요구)
- **제외**: KIS 고유(CASH_ONLY 등), LLM 게이트, capital stage, ARM/DASHBOARD 확인(운영 UI 소관 — 게이트 flip 승인이 대체), LIVE_DISABLED(인터록이 대체), BUY_PAUSED/SELL_ONLY(운영 모드가 대체), 미국장 진입 시간창, MAX_POSITIONS·CANCEL_RATE(P3), SELL_COST_BUFFER(§0.3)
- **구조 대체**: 레버리지/인버스·ETF/ETN 클래스 차단 → 심볼 allowlist(분류 소스 `[미측정]`, P3 재설계)

체인의 모든 입력은 의도 필드·브로커 스냅샷·journal 상태에서 온다(SHALL — 신호 계층 산물 없음: 구조적 RR 계산·등급배수는 P3). 총계 한도의 경계값은 도달 시 차단(≥)이며(SHALL — 예약 트랜잭션의 실장과 일치), 주문 단위 한도는 포함 상한(초과 시 차단)을 유지한다.

#### Scenario: 체인 첫 실패 정지
- **WHEN** kill switch가 활성인 상태에서 진입 의도가 평가되면
- **THEN** 후속 판정 없이 KILL_SWITCH_ACTIVE로 즉시 거부된다

#### Scenario: allowlist 밖 심볼
- **WHEN** allowlist에 없는 심볼의 진입 의도가 평가되면
- **THEN** SYMBOL_NOT_ALLOWED로 거부된다

#### Scenario: 총계 경계값 도달
- **WHEN** 당일 실현 손실이 한도에 정확히 도달한 상태에서 진입 의도가 평가되면
- **THEN** DAILY_LOSS_LIMIT_REACHED로 거부된다 (계좌자본 0 이하이면 즉시 차단)

### Requirement: Automation gate limits are the production Guardian policy source
The production `RiskGuardian` policy SHALL be derived from the automation gate's
five configured limits and one normalized currency. The Guardian's maximum
quantity, maximum order notional, maximum open exposure, maximum daily loss, and
maximum daily loss ratio SHALL produce a limit snapshot byte-for-byte equivalent
to the interlock's audited snapshot, including configured/set bits and currency.
The per-trade risk budget SHALL equal the configured maximum daily-loss amount in
the same currency. The Guardian's daily buy notional ceiling SHALL equal the gate's
resolved `max_daily_buy_notional` (absent or zero resolves to 10,000,000 only when the
currency is KRW) in the same currency, and the startup interlock SHALL compare the
Guardian's declared ceiling with that resolved value; a Guardian that does not declare
one, or declares a different one, SHALL be refused. The ceiling is not part of the
five-field limit snapshot a decision carries, because the reservation transaction —
not the Gateway — is where it is measured.

#### Scenario: USD policy construction
- **WHEN** the gate contains valid USD limits
- **THEN** every Guardian money field and its audited limit snapshot use USD and the configured values without falling back to KRW defaults

#### Scenario: One configured value differs
- **WHEN** any one of the five policy values, the resolved daily buy notional ceiling, or the currency differs from the gate
- **THEN** the startup interlock refuses the Guardian as a single-source mismatch

#### Scenario: Risk budget derivation
- **WHEN** the Guardian policy is created from a valid gate
- **THEN** its per-trade risk budget equals the configured daily-loss amount and grants no larger loss budget

### Requirement: 정책 수치의 provenance

모든 한도·정책 수치는 코드에 출처와 함께 기록되어야 하며(SHALL — 출처는 ① StockOS 파일·심볼과 그 검증 상태, 또는 ② TossOS 실측이다; ②는 `verify-*` change의 `measurements.md`에 남은 관측을 **식별자로 인용**해야 한다(SHALL — 관측 번호·날짜·시장; "관측했다"는 서술은 출처가 아니다)), 사용자 미확정 시 보수 기본값 전체 집합을 사용한다(SHALL — 인터록 5필드를 전부 충족): 주문당 notional 1,000,000 KRW·주문당 수량 100주·총 노출 10,000,000 KRW·일일 손실 100,000 KRW·일일 손실 자본비 1%·통화 KRW. 하루 매수 총액 한도의 기본값은 10,000,000 KRW이며(SHALL), 그 출처는 사용자 결정(2026-10-14, change a136)이다 — 이전에 상한이 없던 새 검사이므로 이 수치의 등재는 완화가 아니다. 이 기본값은 통화가 KRW인 게이트에만 적용되며 다른 통화로 환산해 적용해서는 안 된다(SHALL NOT). USD 축의 승인 집합은 주문당 notional 500 USD·주문당 수량 100주·총 노출 1,500 USD·일일 손실 50 USD·일일 손실 자본비 1%이며, 이 집합은 기본값이 아니라 **운영자가 명시적으로 고를 때만** 적용된다(SHALL NOT — 통화가 하나인 게이트에서 USD를 기본값으로 두면 국내 자동 진입이 조용히 닫힌다). "Toss 검증됨" 표시는 실계좌 검증 결과에만 근거해 전환한다(SHALL NOT — 미검증 상태에서 검증됨 표기 금지). 수치 변경은 §0.9(보수 방향만) 검토와 audit 기록을 요구한다(SHALL). 완화 방향의 수치 등재는 **새 통화 축을 등재하는 경우에 한해**, 각 수치가 이미 승인된 다른 통화의 선호 안쪽에 있고(SHALL — 등가 논증과 **그 논증이 깨지는 환율**을 change 문서에 기록한다), 사람이 그 수치를 명시적으로 승인했으며, 근거를 실측 식별자로 인용할 때만 허용한다(SHALL — 셋 중 하나라도 없으면 등재하지 않는다). 환산 환율은 코드에 넣지 않는다(SHALL NOT — 등가는 등재 시점의 정책 논증이며 런타임이 계산할 값이 아니다; 통화 간 정규화는 FX staleness 경계를 가진 riskcalc의 자리다).

#### Scenario: 한도 완화 시도

- **WHEN** 일일 손실 한도를 높이는 설정 변경이 적용되면
- **THEN** audit 로그에 이전·새 값이 기록되고 변경은 사람 승인 절차를 거친다

#### Scenario: 기본값 게이트 ON 기동

- **WHEN** 사용자 미확정 기본값 집합으로 게이트 ON 기동하면
- **THEN** 인터록 한도 검증(5필드·통화)이 통과한다

#### Scenario: 실측 출처를 가진 수치의 등재

- **WHEN** StockOS에 대응 심볼이 없는 수치를 레지스트리에 등재하면
- **THEN** 코드가 그 수치의 근거를 `measurements.md`의 관측 식별자·날짜·시장으로 인용하고, 등가 논증과 그것이 깨지는 지점이 change 문서에 남는다

#### Scenario: 승인 봉투를 벗어나는 완화 거절

- **WHEN** 새 통화 축의 어떤 수치가 이미 승인된 다른 통화의 선호를 넘으면
- **THEN** 그 수치는 등재되지 않고 기존 값이 유지된다

### Requirement: KR과 US는 하나의 account-base Guardian 권위를 공유한다

Production strategy runtime은 계좌마다 정확히 하나의 account-base-currency Guardian을 사용해야 한다 (SHALL).
그 Guardian은 하나의 account-wide exposure/daily-loss 권위를 소유한다. KR과 US에 quote-currency별
Guardian 또는 독립 account-wide cap을 만들어서는 안 된다 (MUST NOT). `limit_currency`는
account base currency를 뜻하고, 주문 notional·총 노출·일일 손실·equity·risk budget은 그 base
minor unit으로 비교해야 한다 (SHALL). Market cash와 broker order cost는 market quote minor
unit으로 비교해야 하며 base와 quote 숫자를 직접 비교해서는 안 된다 (MUST NOT).

각 exposure-raising 요청은 official source가 mint한 request-scoped frozen FX authority를 사용해야
한다 (SHALL). Authority는 base/quote pair, exact decimal rate, conservative haircut,
source/version/digest, observed-at/fresh-until을 봉인해야 하며 (SHALL), caller가 제공한 raw rate나
digest를 authority로 승격해서는 안 된다 (MUST NOT). KR same-currency path도 봉인된 identity FX를
사용하고 US path는 official quote-to-base FX를 사용해야 한다 (SHALL). 같은 authority가 Guardian
sizing, aggregate와 다섯 bucket reservation, decision limits envelope와 Gateway last-moment
validation에 끝까지 사용되어야 하며 중간 재조회나 다른 환율 사용은 금지한다 (MUST NOT).

일일 손실 측정(「일일 손실 한도의 입력은 원장에서 측정된다」)과 하루 매수 총액 한도의 금액 중 account base currency 로 이미 표시된 것은 환산하지 않으며 authority 도 요구하지 않는다 (SHALL).
다른 시장의 손실 또는 매수 금액은 **그 요청과 같은 paired FX collection(같은 observedAt)에 실린 그 시장의 authority** 로만 환산해야 하며 (SHALL), Guardian 은 precheck 에서
그 authority 의 official source · base/quote pair · 신선도를 검증해야 한다 (SHALL). 그 authority 가 없거나 검증에 실패하면 그 요청의 신규 exposure 를 fail-closed 로 거부해야 한다 (SHALL).
이 환산을 위해 FX 를 재조회하거나, caller 가 제공한 환율 또는 다른 파도의 authority 를 써서는 안 된다 (MUST NOT). 이 조항은 요청의 sizing · aggregate exposure · bucket reservation ·
decision limits envelope · Gateway last-moment validation 이 쓰는 FX 규칙을 바꾸지 않는다 (SHALL NOT).

Base reservation은 exposure를 작게 계산하지 않도록 ceil하고 admissible quantity는 floor해야 한다
(SHALL). FX 없음, stale, 역방향, market/quote/base/digest 불일치 또는 완화된 haircut은 해당
시장의 신규 exposure만 거부하고 (SHALL), protection, fill, reconciliation과 reduce-only exit를
차단해서는 안 된다 (MUST NOT).

#### Scenario: KR identity와 US official conversion의 동시 admission

- **WHEN** 같은 account-base Guardian에 KR identity-FX 요청과 US official quote-to-base 요청이 동시에 들어온다
- **THEN** 두 요청은 commit 순서와 무관하게 하나의 base exposure 잔여를 공유하며 합산 cap을 초과하지 않고 어느 시장도 peer 안정화를 기다리지 않는다

#### Scenario: US FX 만료

- **WHEN** US q_final precheck에 사용한 official FX가 issue 또는 Gateway 직전 만료된다
- **THEN** US exposure-raising authority는 broker request와 부분 reservation 없이 거부되고 KR eligible work와 양 시장 safety lifecycle은 계속된다

#### Scenario: quote cash와 base limit 분리

- **WHEN** US 주문 비용은 USD이고 Guardian exposure limit은 account base currency다
- **THEN** cash는 USD order cost와, exposure는 동일 frozen FX로 환산한 base amount와 비교되며 USD 숫자를 base limit 숫자와 직접 비교하지 않는다

#### Scenario: cross-currency concurrent cap race

- **WHEN** KR과 US 요청이 동시에 마지막 account-wide capacity를 예약하려 한다
- **THEN** aggregate와 다섯 bucket의 base reservation transaction은 합산 cap 안의 요청만 commit하고 loser는 bounded fresh recollection 또는 atomic refusal이며 중복 권한은 0건이다

#### Scenario: FX outage 중 exit

- **WHEN** entry용 official FX authority를 만들 수 없지만 기존 US 포지션의 reduce-only exit가 필요하다
- **THEN** 신규 US entry만 닫히고 exit, protection, reconciliation과 fill 처리는 계속된다

#### Scenario: KR 진입의 일일 손실에 든 US 손실

- **WHEN** KR 진입 precheck 의 일일 손실 측정에 같은 일 키의 US 손실(USD)이 있고 같은 paired FX collection 에 검증되는 US authority 가 있다
- **THEN** 그 손실은 US authority 로 base 환산(올림)되어 측정에 들어가고 KR 요청의 sizing · exposure · reservation 은 KR identity authority 를 그대로 쓴다

#### Scenario: 다른 시장 authority 부재

- **WHEN** KR 진입의 일일 손실 측정에 US 손실이 있는데 같은 paired FX collection 에 US authority 가 없다
- **THEN** 그 KR 진입은 일일 손실 단에서 입력 불가로 거부되고 FX 재조회는 일어나지 않으며 exit · protection · reconciliation 은 계속된다

#### Scenario: base 통화 손실

- **WHEN** account base 가 KRW 이고 측정할 손실이 모두 KR(KRW)이다
- **THEN** 환산과 authority 없이 측정된다

#### Scenario: KR 진입의 하루 매수 합계에 든 US 수동 매수

- **WHEN** KR 진입의 하루 매수 합계에 오늘 US 수동 매수(USD)가 있고 같은 paired FX collection 에 검증되는 US authority 가 있다
- **THEN** 그 금액은 US authority 로 base 환산(올림)되어 합계에 들어가고, authority 가 없으면 그 KR 진입은 거부되며 FX 재조회는 일어나지 않는다


## ADDED Requirements

### Requirement: 하루 매수 총액 한도

Guardian은 계좌 하나의 하루 매수 총액에 상한을 두어야 한다(SHALL). 상한은 `engine.automation_gate.max_daily_buy_notional`(account base 통화 =
`limit_currency`)이며, 키가 없거나 0이면 기본값 10,000,000 KRW를 적용한다(SHALL — 기존 관례: 0은 미설정). 음수·NaN·Inf는 기동 거부다(SHALL).
`limit_currency`가 KRW가 아닌데 값이 없으면 게이트 ON 기동을 거부한다(SHALL — 승인된 다른 통화 기본값이 없고 코드 환율은 금지다).

하루의 경계는 두 시장에 공통인 하나여야 하며(SHALL), **KST 08:00에 넘어간다**: 일 라벨은 `clock.AccountTradingDay(t)` =
`clock.MarketKR.TradingDay(t − 8시간)` 한 곳에서만 계산한다(SHALL — 계좌 단위 일일 판정이 모두 이 함수를 쓴다). 라벨 D 의 하루에는 서울 날짜 D 의
KR 정규장과 뉴욕 날짜 D 의 US 정규장이 함께 들어가며 어느 정규장도 둘로 나뉘지 않는다. 시장별 하루 또는 시장별 별도 한도를 두어서는 안 된다(SHALL NOT).

판정은 체인 단계가 아니라 **결정 영속과 예약 삽입의 원자 journal 트랜잭션 안**에서 해야 한다(SHALL). 오늘의 합계는 같은 트랜잭션에서 읽은
① 오늘(발급일 기준 — 체결일이 아님) 발급된 EXPOSURE_RAISING 결정의 OPEN_EXPOSURE 예약 중 **RELEASED** 이고 체결이 없었음이 증명되지 않은 것 — 제외는 셋뿐이다:
attempt 없이 미소비 만료(EXPIRED_UNCONSUMED)로 해제됨, attempt 가 NOT_DISPATCHED 또는 FAILED_CONFIRMED, 또는 체결 경로 해제(BROKER_TERMINAL)이고 attempt 가 그 결정의
CONFIRMED PLACE 이며 그 주문의 체결 스냅숏이 attempt 의 intent 범위(계좌·시장·거래일·종목·매수)와 attempt 확정 뒤 커밋으로 정확히 한 행 · 종결(fail-closed 아님) ·
누적 체결 0 이고 소유 intent 가 하나이며 그 주문에 체결 이벤트·대체 계보·AMEND 시도가 없음, ② 발급일과 무관하게 아직 HELD인 OPEN_EXPOSURE 예약의 합이며, 새 진입의 예약 금액을 더한 값이 한도에 **도달하면** 발급을 거부하고
LIMIT_REACHED로 기록한다(SHALL — 총계 경계 ≥와 같다). 결정의 발급일과 예약 커밋 순간의 하루가 다르면 그 예약은 거부된다(SHALL — 매수는 발급일에 귀속되므로). 금액은 예약 금액 그대로(`risk.EntryExposureValue` — 비용 포함, account base, US는 요청 단위
frozen FX로 환산·ceil)를 쓰며 새 환율 경로를 만들어서는 안 된다(SHALL NOT). 합계를 계산할 수 없으면 그 매수의 결정과 예약은 함께 롤백된다(SHALL —
fail-closed). EXPOSURE_RAISING 결정의 예약 트랜잭션이 이 한도를 함께 받지 않으면 그 요청은 거부된다(SHALL — 한도를 빠뜨린 발급 문이 조용히 통과하지 못한다).

이 한도는 EXPOSURE_RAISING 진입 발급에만 적용된다(SHALL). RISK_REDUCING 발급, 손절, 익절, 비상 청산, flatten, 보호 주문, 대사, 체결 감지는 이 판정을
지나서는 안 되며 지연되어서도 안 된다(SHALL NOT — §0.3). Guardian 밖의 수동 주문(토스 앱 등)은 이 한도로 막지 않는다(SHALL NOT — OFF = upstream). 그러나 수동 매수는 오늘 합계에 **들어간다**(SHALL —
사용자 결정 2026-10-14 D9=b): 엔진 보조 실행자가 기존 공식 주문 조회(GET /api/v1/orders)로 라벨 D 의 매수 주문을 읽어 시각이 찍힌 스냅숏을 만들고,
이 계좌 attempt 의 `broker_order_id` 와 같은 주문(엔진 자신의 주문 — 예약 원장이 이미 셈)은 뺀다(SHALL). 수동 매수 금액은 체결액과 미체결 잔량 × 지정가의
합이며, 지정가 없는 미체결 주문처럼 금액을 정할 수 없는 행이 있으면 그 스냅숏은 산정 불가다(SHALL). 브로커 조회는 precheck 밖에서만 하며(SHALL — 첫 다리
precheck 의 「journal 변경·브로커 호출 없음」 계약), 스냅숏이 없거나 · 신선도 상한(2분)을 넘었거나 · 산정 불가이거나 · 다른 하루의 것이면 **매수만** 거부한다(SHALL). 실패한 읽기는 산정 불가 스냅숏으로 게시되어 이전 값이 계속 쓰이지 않는다(SHALL).
US 수동 매수의 base 환산은 「KR과 US는 하나의 account-base Guardian 권위를 공유한다」의 paired FX 규칙을 따른다(SHALL).

한도 변경은 기존 한도 설정 seam(콘솔 한도 편집기 → 키 단위 쓰기, `enabled` 비기록)으로 하며 필드별 이전·새 값이 audit 로그에 남아야 한다(SHALL).
설정 파일 수동 편집은 기동 audit 가 해석 값의 변화를 남긴다(SHALL). 콘솔 칸은 현재 한도 통화의 자유 숫자 입력이고 USD 프리셋 값을 만들지 않으며(SHALL NOT),
KRW 에서 미설정이면 적용될 기본값 10,000,000 을, USD 에서 미설정이면 「엔진이 기동을 거부함」을 평이하게 보여 준다(SHALL). 저장이 한도 통화를 바꾸면
`max_daily_buy_notional` 은 0(미설정)으로 지워진다(SHALL — 옛 통화의 숫자가 새 통화로 재해석되지 않게). 콘솔 편집에 타이핑 확인이나 추가 승인 단계를
두지 않는다(SHALL NOT). 가동 중인 엔진은 기동 시점 값으로 동작하고 재시작 때 새 값을 쓴다.

#### Scenario: 동시 매수가 함께 넘지 못한다

- **WHEN** 오늘 합계가 9,500,000 KRW이고 각 400,000 KRW인 두 진입이 동시에 발급을 시도하면
- **THEN** 먼저 커밋한 하나만 발급되고 다른 하나는 LIMIT_REACHED로 결정·예약이 함께 롤백된다

#### Scenario: 체결된 매수도 하루가 끝날 때까지 센다

- **WHEN** 오늘 발급된 진입이 체결되어 OPEN_EXPOSURE 보유가 해제된 뒤 새 진입이 평가되면
- **THEN** 체결된 진입의 예약 금액이 오늘 합계에 그대로 남아 새 진입과 함께 한도에 대조된다

#### Scenario: 체결 없음이 증명된 매수는 빠진다

- **WHEN** 오늘 발급된 진입이 체결 수량 0으로 종결되었거나 미전송으로 끝났으면
- **THEN** 그 예약 금액은 오늘 합계에서 빠진다

#### Scenario: 하루 경계

- **WHEN** KST 07:59에 합계가 한도에 도달해 매수가 거부되고 KST 08:00 이후 새 진입이 평가되면
- **THEN** 전날 하루에 발급된 종결 진입은 합계에서 빠지고, 아직 HELD인 미체결 매수만 남아 대조된다

#### Scenario: 합계를 계산할 수 없음

- **WHEN** 오늘 합계에 들어갈 예약 행의 금액이나 통화를 해석할 수 없으면
- **THEN** 그 매수는 결정과 예약 없이 거부되고, 같은 시각의 매도·청산 발급은 영향을 받지 않는다

#### Scenario: 매도는 한도와 무관하다

- **WHEN** 하루 매수 총액이 한도에 도달한 상태에서 손절 또는 reduce-only 매도가 발급되면
- **THEN** 매도는 이 한도를 평가하지 않고 지연 없이 발급된다

#### Scenario: US 매수의 환산

- **WHEN** account base가 KRW인 계좌에 US 진입이 평가되면
- **THEN** 그 진입이 OPEN_EXPOSURE로 예약하는 frozen FX 환산 금액이 하루 합계에 더해지고 다른 환율은 쓰이지 않는다

#### Scenario: 기본값과 한도 변경 감사

- **WHEN** 설정에 키가 없는 KRW 게이트로 기동했다가 운영자가 콘솔에서 한도를 5,000,000으로 바꾸면
- **THEN** 기동 때는 10,000,000 KRW가 적용되고, 변경은 이전·새 값이 audit 로그에 남으며 재시작 뒤 5,000,000이 적용된다

#### Scenario: 엔진 밖 수동 매수도 합계에 든다

- **WHEN** 오늘 토스 앱으로 4,000,000 KRW 를 산 계좌에서 엔진 합계가 5,000,000 KRW 이고 상한이 10,000,000 KRW 일 때 1,500,000 KRW 진입이 평가되면
- **THEN** 수동 매수는 막히지 않았고, 진입은 4,000,000 + 5,000,000 + 1,500,000 ≥ 10,000,000 으로 LIMIT_REACHED 거부된다

#### Scenario: 수동 매수 스냅숏이 낡음

- **WHEN** 수동 매수 스냅숏이 신선도 상한보다 오래되었거나 없으면
- **THEN** 진입만 거부되고 매도·손절·청산은 영향을 받지 않는다

#### Scenario: 콘솔 통화 전환

- **WHEN** KRW 로 5,000,000 을 넣어 둔 계좌에서 운영자가 콘솔로 한도 통화를 USD 로 저장하면
- **THEN** `max_daily_buy_notional` 은 0 으로 지워지고 콘솔은 USD 미설정이면 엔진이 기동을 거부한다고 보여 준다

#### Scenario: 상한 없는 예약

- **WHEN** EXPOSURE_RAISING 결정의 예약 요청에 하루 매수 상한이 없으면
- **THEN** 그 예약과 결정은 거부되어 남지 않는다
