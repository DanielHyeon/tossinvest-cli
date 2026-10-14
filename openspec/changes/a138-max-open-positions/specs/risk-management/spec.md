## MODIFIED Requirements

### Requirement: Guardian 판정 체인
모든 자동 진입 의도는 Guardian 판정 체인을 통과해야 하며(SHALL), 체인은 고정 순서로 평가되고 첫 실패에서 정지한다: kill switch/운영 모드 → 게이트 상태(진입 차단 latch: 401/403·SLO 위반·RECONCILE·recovery 미완료) → 심볼 allowlist → 구조적 손절 계약(No Stop=No Trade·보호적 stop·실질 본전 미달 target 거부) → 주문 크기 한도 → 최소 RR(신규 검사) → 현금 검증(비용 포함) → 당일 재진입 규칙 → 총 개방 노출 한도 → 일일 손실 한도 → 중복 주문 검사 → ALLOW. 각 거부는 안정적 reason-code로 기록된다(SHALL).

**순서의 권위는 이 표다**(SHALL — StockOS `evaluate_guardian`과의 대응은 `docs/guardian-chain.md` 참고 산출물로 남기되 순서 보존을 규범으로 주장하지 않는다: 원 체인에는 미이식 단계가 섞여 있고 최소 RR은 원 체인에 없는 신규 검사다). 이식 분류의 열거(SHALL):

- **이식**: kill switch·모드, 진입 latch, 손절 계약(reason: STOP_MISSING·STOP_NOT_BELOW_ENTRY·TARGET_NOT_ABOVE_ENTRY·INVALID_TARGET_STOP·TARGET_BELOW_BREAK_EVEN), 주문 크기(INVALID_ORDER_SIZE·MAX_ORDER_EXCEEDED), 현금(INSUFFICIENT_CASH·비용 포함), 당일 재진입(쿨다운 기본 30분 `[미검증]`·당일 심볼당 최대 진입 수 기본 2회 `[미검증]`·**미체결 매수 존재 시 진입 차단 PENDING_BUY_ORDER_BLOCKED**), 총 노출, 일손실, 중복 주문
- **신규**: 최소 RR, 심볼 allowlist, STOP_MISSING(원본은 생성자 강제라 카운터파트 없음), 하루 매수 총액(DAILY_TURNOVER 의 매수 측 — 체인 단계가 아니라 결정·예약 원자 트랜잭션에서 판정한다: 「하루 매수 총액 한도」 요구), 최대 보유 포지션 수(MAX_POSITIONS — 체인 단계가 아니라 결정·예약 원자 트랜잭션에서 판정한다: 「최대 보유 포지션 수 한도」 요구)
- **제외**: KIS 고유(CASH_ONLY 등), LLM 게이트, capital stage, ARM/DASHBOARD 확인(운영 UI 소관 — 게이트 flip 승인이 대체), LIVE_DISABLED(인터록이 대체), BUY_PAUSED/SELL_ONLY(운영 모드가 대체), 미국장 진입 시간창, CANCEL_RATE(P3), SELL_COST_BUFFER(§0.3)
- **구조 대체**: 레버리지/인버스·ETF/ETN 클래스 차단 → 심볼 allowlist(분류 소스 `[미측정]`, P3 재설계)

체인의 모든 입력은 의도 필드·브로커 스냅샷·journal 상태에서 온다(SHALL — 신호 계층 산물 없음: 구조적 RR 계산·등급배수는 P3). 총계 한도의 경계값은 도달 시 차단(≥)이며(SHALL — 예약 트랜잭션의 실장과 일치), 주문 단위 한도는 포함 상한(초과 시 차단)을 유지한다. 최대 보유 포지션 수는 허용되는 최대 개수이므로 그 개수를 넘게 될 때 차단한다(SHALL — 「최대 보유 포지션 수 한도」).

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
not the Gateway — is where it is measured. The Guardian's maximum open positions
SHALL equal the gate's resolved `max_open_positions` (absent or zero resolves to 6 in
every currency), and the startup interlock SHALL compare the Guardian's declared
maximum with that resolved value; a Guardian that does not declare one, or declares a
different one, SHALL be refused. The maximum is likewise outside the five-field limit
snapshot and is measured in the reservation transaction.

#### Scenario: USD policy construction
- **WHEN** the gate contains valid USD limits
- **THEN** every Guardian money field and its audited limit snapshot use USD and the configured values without falling back to KRW defaults

#### Scenario: One configured value differs
- **WHEN** any one of the five policy values, the resolved daily buy notional ceiling, the resolved maximum open positions, or the currency differs from the gate
- **THEN** the startup interlock refuses the Guardian as a single-source mismatch

#### Scenario: Risk budget derivation
- **WHEN** the Guardian policy is created from a valid gate
- **THEN** its per-trade risk budget equals the configured daily-loss amount and grants no larger loss budget

### Requirement: 정책 수치의 provenance

모든 한도·정책 수치는 코드에 출처와 함께 기록되어야 하며(SHALL — 출처는 ① StockOS 파일·심볼과 그 검증 상태, 또는 ② TossOS 실측이다; ②는 `verify-*` change의 `measurements.md`에 남은 관측을 **식별자로 인용**해야 한다(SHALL — 관측 번호·날짜·시장; "관측했다"는 서술은 출처가 아니다)), 사용자 미확정 시 보수 기본값 전체 집합을 사용한다(SHALL — 인터록 5필드를 전부 충족): 주문당 notional 1,000,000 KRW·주문당 수량 100주·총 노출 10,000,000 KRW·일일 손실 100,000 KRW·일일 손실 자본비 1%·통화 KRW. 하루 매수 총액 한도의 기본값은 10,000,000 KRW이며(SHALL), 그 출처는 사용자 결정(2026-10-14, change a136)이다 — 이전에 상한이 없던 새 검사이므로 이 수치의 등재는 완화가 아니다. 이 기본값은 통화가 KRW인 게이트에만 적용되며 다른 통화로 환산해 적용해서는 안 된다(SHALL NOT). 최대 보유 포지션 수의 기본값은 6이며(SHALL), 그 출처는 사용자 결정(2026-10-15, change a138)이다 — 이전에 상한이 없던 새 검사이므로 이 수치의 등재는 완화가 아니다. 개수는 통화와 무관하므로 이 기본값은 모든 통화의 게이트에 적용된다(SHALL). USD 축의 승인 집합은 주문당 notional 500 USD·주문당 수량 100주·총 노출 1,500 USD·일일 손실 50 USD·일일 손실 자본비 1%이며, 이 집합은 기본값이 아니라 **운영자가 명시적으로 고를 때만** 적용된다(SHALL NOT — 통화가 하나인 게이트에서 USD를 기본값으로 두면 국내 자동 진입이 조용히 닫힌다). "Toss 검증됨" 표시는 실계좌 검증 결과에만 근거해 전환한다(SHALL NOT — 미검증 상태에서 검증됨 표기 금지). 수치 변경은 §0.9(보수 방향만) 검토와 audit 기록을 요구한다(SHALL). 완화 방향의 수치 등재는 **새 통화 축을 등재하는 경우에 한해**, 각 수치가 이미 승인된 다른 통화의 선호 안쪽에 있고(SHALL — 등가 논증과 **그 논증이 깨지는 환율**을 change 문서에 기록한다), 사람이 그 수치를 명시적으로 승인했으며, 근거를 실측 식별자로 인용할 때만 허용한다(SHALL — 셋 중 하나라도 없으면 등재하지 않는다). 환산 환율은 코드에 넣지 않는다(SHALL NOT — 등가는 등재 시점의 정책 논증이며 런타임이 계산할 값이 아니다; 통화 간 정규화는 FX staleness 경계를 가진 riskcalc의 자리다).

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

### Requirement: 최대 보유 포지션 수 한도

Guardian은 계좌 하나가 동시에 드는 포지션의 수에 상한을 두어야 한다(SHALL). 상한은 `engine.automation_gate.max_open_positions`이며, 키가 없거나 0이면
기본값 6을 적용한다(SHALL — 기존 관례: 0은 미설정, 개수는 통화와 무관). 음수·정수가 아님·NaN·Inf는 게이트 ON 기동 거부다(SHALL).

세는 단위는 서로 다른 (시장, 종목) 하나다(SHALL — 같은 종목의 인스턴스·부분 체결 중인 진입과 그 포지션은 하나로 센다). 판정은 체인 단계가 아니라 **결정 영속과
예약 삽입의 원자 journal 트랜잭션 안**에서 해야 하며(SHALL), 같은 트랜잭션에서 읽은 ① 계좌의 살아 있는 포지션(투영에서 CLOSED가 아닌 행 — 엔진 진입·편입·
대사가 접어 넣은 외부 보유를 모두 포함), ② 발급일과 무관하게 아직 HELD인 예약(종류 무관 — Gateway가 제출을 인가하는 술어와 같다)을 가진 EXPOSURE_RAISING 결정의 종목,
③ 이번 결정의 종목, ④ 종목 범위의 활성 RECONCILE 상태 중 ①~③에 없는 것(체결 투영이 거절되어 포지션 행 없이 보유가 남은 종목 — 범위 시장이 있으면 (시장, 종목)으로, 시장이 없는
범위는 종목만으로 ①~③과 맞추고 맞는 것이 없으면 한 칸으로 센다)의
합집합 크기가 상한을 **넘으면** 발급을 거부하고 LIMIT_REACHED로 기록한다(SHALL — 상한은 허용되는 최대 개수다: 상한 6이면 6개는 허용, 7번째 종목은 거부).
이미 집합에 있는 종목에 더 사는 진입은 개수를 늘리지 않으므로 이 한도로 거부되지 않는다(SHALL). 종목은 트랜잭션 안에서 결정의 저장된 preimage에서 읽어야
하며 호출자 값을 써서는 안 된다(SHALL NOT). 개수를 계산할 수 없으면(원장 읽기 실패, preimage 해석 불가) 그 진입의 결정과 예약은 함께 롤백된다(SHALL —
fail-closed, 0으로 대체 금지). EXPOSURE_RAISING 결정의 예약 트랜잭션이 이 상한을 함께 받지 않으면 그 요청은 거부된다(SHALL — 한도를 빠뜨린 발급 문이
조용히 통과하지 못한다).

이 한도는 EXPOSURE_RAISING 진입 발급에만 적용된다(SHALL). RISK_REDUCING 발급, 손절, 익절, 비상 청산, flatten, 보호 주문, 대사, 체결 감지는 이 판정을
지나서는 안 되며 지연되어서도 안 된다(SHALL NOT — §0.3). Guardian 밖의 수동 주문은 이 한도로 막지 않는다(SHALL NOT — OFF = upstream). 명시 한계: 대사가 접어 넣기 전의 외부 매수(대사 주기 60초 이상,
안정 스냅숏이 잡힌 사이클까지), 미체결 수동 매수 주문, 운영자가 HELD 예약을 해제했지만 주문이 살아 있거나 체결된 경우(체결 감지 전), 엔진 진입 주문을 엔진 밖에서
정정한 경우(대사 전)는 세지 않는다. 반대로 엔진 밖에서 판 포지션(대사 수렴 전), 운영자 해제를 기다리는 HELD 예약, 소수 잔량 포지션, RECONCILE 종목은 한 칸씩 센다.
KR과 US는 한 계좌에서 합산된다.

한도 변경은 기존 한도 설정 seam(콘솔 한도 편집기 → 키 단위 쓰기, `enabled` 비기록)으로 하며 이전·새 값이 audit 로그에 남아야 한다(SHALL — 콘솔 저장의 audit 줄은
저장을 되돌리지 않는 best-effort 이고, 기동 audit 가 해석 값을 매 기동 다시 남기는 것이 둘째 기록이다; 0 은 「미설정 — 기본 6」으로 적는다). 설정 파일
수동 편집은 기동 audit가 해석 값의 변화를 남긴다(SHALL). 콘솔 칸은 정수 입력이고, 미설정이면 적용될 기본값 6을 보여 준다(SHALL). 한도 통화를 바꾸는
저장은 이 값을 지우지 않는다(SHALL — 개수는 통화로 재해석되지 않는다). 콘솔 편집에 타이핑 확인이나 추가 승인 단계를 두지 않는다(SHALL NOT). 가동 중인
엔진은 기동 시점 값으로 동작하고 재시작 때 새 값을 쓴다.

#### Scenario: 일곱 번째 종목은 거부된다

- **WHEN** 계좌가 서로 다른 6 종목의 살아 있는 포지션을 들고 있고 상한이 6인 상태에서 새 종목 진입이 평가되면
- **THEN** 진입은 LIMIT_REACHED로 결정·예약이 함께 롤백된다

#### Scenario: 든 종목에 더 사는 진입

- **WHEN** 계좌가 서로 다른 6 종목을 들고 있고 그중 한 종목에 더 사는 진입이 평가되면
- **THEN** 개수가 늘지 않으므로 이 한도로 거부되지 않는다

#### Scenario: 동시 진입이 함께 넘지 못한다

- **WHEN** 5 종목을 든 계좌에서 서로 다른 새 종목 두 진입이 동시에 발급을 시도하면
- **THEN** 먼저 커밋한 하나만 발급되고 다른 하나는 첫 진입의 HELD 예약을 보고 LIMIT_REACHED로 롤백된다

#### Scenario: 편입·외부 보유도 센다

- **WHEN** 계좌가 엔진 진입 4 종목, 편입 1 종목, 대사가 접어 넣은 외부 보유 1 종목을 들고 있고 새 종목 진입이 평가되면
- **THEN** 개수는 6이므로 새 종목 진입은 거부된다

#### Scenario: 투영이 거절된 체결의 종목도 센다

- **WHEN** 진입 주문의 종결 체결이 예약을 해제했지만 포지션 투영이 거절되어 그 종목이 RECONCILE 상태이고 포지션 행이 없을 때, 계좌의 다른 5 종목과 함께 새 종목 진입이 평가되면
- **THEN** RECONCILE 종목이 한 칸으로 세어져 개수는 6이고 새 종목 진입은 거부된다

#### Scenario: 셀 수 없음

- **WHEN** HELD 예약을 가진 결정의 preimage를 해석할 수 없으면
- **THEN** 진입은 결정과 예약 없이 거부되고, 같은 시각의 매도·청산 발급은 영향을 받지 않는다

#### Scenario: 매도는 한도와 무관하다

- **WHEN** 개수가 상한을 넘은 상태(예: 외부 보유가 늘어남)에서 손절 또는 reduce-only 매도가 발급되면
- **THEN** 매도는 이 한도를 평가하지 않고 지연 없이 발급된다

#### Scenario: 기본값과 변경 감사

- **WHEN** 설정에 키가 없는 게이트로 기동했다가 운영자가 콘솔에서 상한을 4로 바꾸면
- **THEN** 기동 때는 6이 적용되고, 변경은 이전·새 값이 audit 로그에 남으며 재시작 뒤 4가 적용된다

#### Scenario: 상한 없는 예약

- **WHEN** EXPOSURE_RAISING 결정의 예약 요청에 최대 보유 포지션 수가 없으면
- **THEN** 그 예약과 결정은 거부되어 남지 않는다
