## ADDED Requirements

### Requirement: 일일 손실 한도의 입력은 원장에서 측정된다

생산 Guardian 은 자동 진입 의도를 판정하기 전에 엔진 원장에서 그날의 실현 손실을 직접 측정해야 한다 (SHALL). 그날은 하루 매수 총액 한도와 같은 일 키 D 이고, 귀속은 `clock.AccountTradingDay(closed_at) == D`(KST 08:00 에 넘어가는 하루, `closed_at` 은 원장 커밋 시각)로 한다 (SHALL). 시장별 현지일로 귀속해서는 안 된다 (SHALL NOT — 늦게 기록된 손실이 끝난 날로 떨어져 세어지지 않는다). 일 키 함수는 하나여야 한다 (SHALL).
측정 대상은 KR · US 의 관리 포지션(진입 결정 또는 편입이 있는 것)이며 그 밖의 시장은 측정 불가다 (SHALL). 각 포지션의 값은 동결된 비용 차감 후 실현 손익이고, 포지션 사이에서 이익이 손실을 상쇄해서는 안 된다 (SHALL NOT — 음수 행만 합산한다). 합산은 `riskcalc.DailyLoss` 의 계약으로 하며 새 공식을 만들어서는 안 된다 (SHALL NOT).
관리 대상이 아닌 포지션, 아직 열린 포지션의 부분 청산, 미실현 손익은 측정에 들어가지 않는다 (SHALL — 명시 한계). 조정으로 수량 0 이 되어 닫힌 관리 포지션은 그 인스턴스 수명 안에 엔진 매도가 브로커에 나갔을 수 있는 기록 — 발송된 엔진 SELL 주문 attempt, 또는 브로커에 놓였을 수 있는 보호 주문(생성 · 교체) — 이 없을 때만 측정에서 빠지며 (SHALL — 사람의 앱 매도), 있으면 측정 불가다 (SHALL). `closed_at` 이 없거나 해석할 수 없는 CLOSED 관리 포지션이 하나라도 있으면 측정 불가다 (SHALL).
체인이 비교하는 값은 호출자 값과 측정값이 **둘 다 유효할 때** 그 중 큰 쪽이어야 한다 (SHALL). 호출자 값이 무효(해석 불가 · 통화 불일치 · 음수)면 측정값으로 대체하지 않고 거절해야 하며 (SHALL), 측정이 실패하면 호출자 값으로 대체해서는 안 된다 (SHALL NOT).
측정을 할 수 없으면 — 원장 읽기 실패, 그날 체결로 닫힌 관리 포지션의 성과 행 부재, 엔진 SELL 이 섞인 조정 종결, 다른 시장 손실의 환산 authority 부재, 원장 값 해석 불가, 측정 결과의 만료 · 일 키 불일치 — 그 진입은 일일 손실 단에서 입력 불가로 거절되어야 하며 (SHALL), 앞 단의 거절 사유를 가려서는 안 된다 (SHALL NOT).
측정은 진입에만 적용된다: 위험 축소 결정 · 손절 · 비상 청산 · 대사 · 체결 감지는 측정을 호출하거나 기다려서는 안 된다 (MUST NOT). 측정은 원장 행을 연 채 네트워크를 타서는 안 되고 (MUST NOT), mutation-free precheck 의 브로커 호출 금지를 깨서는 안 된다 (MUST NOT).
측정이 한도 도달을 판정하면 생산 진입 경로는 운영 모드를 ENTRY_BLOCKED(trigger = 일일 손실 한도)로 강화 · 영속해야 한다 (SHALL). 전략 경로에서 ① 그 거절, ② trigger 가 일일 손실 한도인 운영 모드 행이 만든 모드 투영 거절, ③ 그 모드 변경이 원인인 제출 직전 펜스 거절은 주기 오류 · 레인 결함 · 레인 잠금이 되어서는 안 되며 (SHALL NOT), 계좌 단위 진입 차단 결과로 그 주기를 끝내야 하고 (SHALL), 거절 원인과 trigger 는 타입으로 보존되어야 한다 (SHALL). 다른 trigger 의 모드 차단 · 기동 복원 실패 차단 · 그 밖의 진입 관문과 펜스 거절의 동작은 바뀌어서는 안 된다 (SHALL NOT). 같은 주기에 앞서 건너뛴 범위 거절이 있으면 그 보고는 지금처럼 주기의 오류로 남는다 (SHALL — 조용한 건너뛰기 금지).

#### Scenario: 서명 매니페스트가 0 을 적은 손실일
- **WHEN** 계좌 매니페스트의 일일 손실이 0 이고 원장에 그날 닫힌 손실 합계가 한도 이상이다
- **THEN** 진입은 `DAILY_LOSS_LIMIT_REACHED` 로 거절되고 운영 모드는 ENTRY_BLOCKED 로 강화된다

#### Scenario: 포지션 사이의 이익은 상쇄하지 않는다
- **WHEN** 그날 닫힌 두 포지션이 −60,000 과 +80,000 이고 한도가 50,000 이다
- **THEN** 측정값은 60,000 이고 진입은 거절된다

#### Scenario: 늦게 기록된 손실
- **WHEN** 브로커에서 전날 체결된 손절이 KST 08:00 이후에야 원장에 반영되어 `closed_at` 이 오늘 일 키에 든다
- **THEN** 그 손실은 오늘 측정에 들어간다

#### Scenario: 호출자 값이 더 크다
- **WHEN** 호출자 값과 측정값이 모두 유효하고 호출자 값이 더 크다
- **THEN** 체인은 호출자 값으로 판정한다

#### Scenario: 무효한 호출자 값
- **WHEN** 호출자 값이 음수이고 측정은 성공했다
- **THEN** 진입은 일일 손실 단에서 `INPUT_UNAVAILABLE` 로 거절되고 측정값으로 대체되지 않는다

#### Scenario: 측정 실패는 호출자 값으로 떨어지지 않는다
- **WHEN** 호출자 값은 0 으로 유효하고 측정이 실패한다
- **THEN** 진입은 일일 손실 단에서 `INPUT_UNAVAILABLE` 로 거절된다

#### Scenario: 성과 행이 없는 그날의 청산
- **WHEN** 그날 체결로 닫힌 관리 포지션 하나의 성과 동결이 거절되어 성과 행이 없다
- **THEN** 그 일 키의 진입은 일일 손실 단에서 `INPUT_UNAVAILABLE` 로 거절되고, 매도 · 손절 경로와 다른 포지션의 청산은 영향받지 않는다

#### Scenario: 사람이 앱에서 판 관리 포지션
- **WHEN** 관리 포지션이 조정으로 수량 0 이 되어 닫혔고 그 수명에 발송된 엔진 SELL attempt 가 없다
- **THEN** 그 포지션은 측정에 들어가지 않고 진입을 막지도 않는다

#### Scenario: 엔진 매도가 섞인 조정 종결
- **WHEN** 관리 포지션이 조정으로 닫혔는데 그 수명에 발송된 엔진 SELL attempt 가 있다
- **THEN** 측정 불가로 그 일 키의 진입이 거절된다

#### Scenario: 측정 실패 중 손절
- **WHEN** 원장 측정이 실패하는 상태에서 손절 · 비상 청산 위험 축소 결정이 요청된다
- **THEN** 측정은 호출되지 않고 위험 축소 결정은 지연 없이 발급된다

#### Scenario: kill switch 가 켜진 상태의 측정 실패
- **WHEN** kill switch 가 켜져 있고 원장 측정도 실패한다
- **THEN** 거절 사유는 kill switch 단의 것이다

#### Scenario: 서명 활성화된 레인의 한도 도달
- **WHEN** 같은 주기에 앞서 건너뛴 범위 거절이 없고, 서명 활성화된 전략 레인의 진입이 일일 손실 한도로 거절되고 이어 다른 레인의 진입이 그 ENTRY_BLOCKED 로 거절된다
- **THEN** 두 레인 어느 것도 잠기지 않고 그 주기는 오류 없이 끝나며 운영 모드는 ENTRY_BLOCKED 다

#### Scenario: 앞서 건너뛴 범위 거절과 섞인 한도 도달
- **WHEN** 같은 주기에 한 범위가 버킷 고갈로 건너뛰어진 뒤 다른 범위가 일일 손실 한도로 거절된다
- **THEN** 운영 모드는 ENTRY_BLOCKED 로 강화되고 그 주기의 남은 범위 전달은 멈추며, 건너뛴 범위 거절은 지금처럼 주기의 오류로 보고된다

#### Scenario: 다른 원인의 모드 차단은 그대로
- **WHEN** 운영 모드가 일일 손실이 아닌 trigger 로 ENTRY_BLOCKED 이거나 기동 복원 실패로 진입이 막혀 있다
- **THEN** 전략 경로의 거절 동작(서명 활성화된 레인의 잠금 포함)은 이 change 이전과 같다

#### Scenario: 강화와 경합한 제출 직전 펜스
- **WHEN** 레인 B 의 진입이 강화 전에 승인되었고 제출 직전 재확인에서 일일 손실 한도 강화로 인한 모드 변경을 만나 펜스 거절된다
- **THEN** 그 거절은 계좌 단위 진입 차단 결과이고 레인 B 는 잠기지 않으며 브로커 요청은 없다

#### Scenario: 보호 주문이 있던 조정 종결
- **WHEN** 관리 포지션이 조정으로 닫혔는데 그 수명에 브로커에 놓였을 수 있는 보호 주문(생성 · 교체)이 있었다
- **THEN** 측정 불가로 그 일 키의 진입이 거절된다

#### Scenario: 일 경계
- **WHEN** KST 08:00 을 지나 일 키가 바뀐다
- **THEN** 이전 일 키의 손실과 성과 행 부재는 새 일 키의 측정에 들어가지 않는다. 이미 강화된 ENTRY_BLOCKED 는 사람이 풀 때까지 유지된다

## MODIFIED Requirements

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
