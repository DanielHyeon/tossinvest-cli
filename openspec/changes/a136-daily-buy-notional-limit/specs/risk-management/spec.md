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
five configured limits, the resolved daily buy notional ceiling, and one normalized
currency. The Guardian's maximum quantity, maximum order notional, maximum open
exposure, maximum daily loss, maximum daily loss ratio, and maximum daily buy
notional SHALL produce a limit snapshot byte-for-byte equivalent
to the interlock's audited snapshot, including configured/set bits and currency.
The per-trade risk budget SHALL equal the configured maximum daily-loss amount in
the same currency.

#### Scenario: USD policy construction
- **WHEN** the gate contains valid USD limits
- **THEN** every Guardian money field and its audited limit snapshot use USD and the configured values without falling back to KRW defaults

#### Scenario: One configured value differs
- **WHEN** any one of the six policy values or the currency differs from the gate
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

## ADDED Requirements

### Requirement: 하루 매수 총액 한도

Guardian은 계좌 하나의 하루 매수 총액에 상한을 두어야 한다(SHALL). 상한은 `engine.automation_gate.max_daily_buy_notional`(account base 통화 =
`limit_currency`)이며, 키가 없거나 0이면 기본값 10,000,000 KRW를 적용한다(SHALL — 기존 관례: 0은 미설정). 음수·NaN·Inf는 기동 거부다(SHALL).
`limit_currency`가 KRW가 아닌데 값이 없으면 게이트 ON 기동을 거부한다(SHALL — 승인된 다른 통화 기본값이 없고 코드 환율은 금지다).

하루의 경계는 두 시장에 공통인 하나여야 하며(SHALL), **KST 08:00에 넘어간다**(일 라벨 = `clock.MarketKR.TradingDay(now − 8시간)`):
KR 정규장과 그날 밤 US 정규장이 같은 하루에 속하고 어느 정규장도 둘로 나뉘지 않는다. 시장별 하루 또는 시장별 별도 한도를 두어서는 안 된다(SHALL NOT).

판정은 체인 단계가 아니라 **결정 영속과 예약 삽입의 원자 journal 트랜잭션 안**에서 해야 한다(SHALL). 오늘의 합계는 같은 트랜잭션에서 읽은
① 오늘 발급된 EXPOSURE_RAISING 결정의 OPEN_EXPOSURE 예약 금액 중 체결이 없었음이 증명되지 않은 것(미전송 종결·미소비 만료·체결 수량 0 종결만
제외), ② 발급일과 무관하게 아직 HELD인 OPEN_EXPOSURE 예약의 합이며, 새 진입의 예약 금액을 더한 값이 한도에 **도달하면** 발급을 거부하고
LIMIT_REACHED로 기록한다(SHALL — 총계 경계 ≥와 같다). 금액은 예약 금액 그대로(`risk.EntryExposureValue` — 비용 포함, account base, US는 요청 단위
frozen FX로 환산·ceil)를 쓰며 새 환율 경로를 만들어서는 안 된다(SHALL NOT). 합계를 계산할 수 없으면 그 매수의 결정과 예약은 함께 롤백된다(SHALL —
fail-closed).

이 한도는 EXPOSURE_RAISING 진입 발급에만 적용된다(SHALL). RISK_REDUCING 발급, 손절, 익절, 비상 청산, flatten, 보호 주문, 대사, 체결 감지는 이 판정을
지나서는 안 되며 지연되어서도 안 된다(SHALL NOT — §0.3). Guardian 밖의 수동 주문은 이 한도로 막지 않으며(SHALL NOT — OFF = upstream) 합계에도 들어가지
않는다.

한도 변경은 기존 한도 설정 seam(키 단위 쓰기, `enabled` 비기록)으로만 하며 필드별 이전·새 값이 audit 로그에 남아야 한다(SHALL). 콘솔 편집에 타이핑
확인이나 추가 승인 단계를 두지 않는다(SHALL NOT). 가동 중인 엔진은 기동 시점 값으로 동작하고 재시작 때 새 값을 쓴다.

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
