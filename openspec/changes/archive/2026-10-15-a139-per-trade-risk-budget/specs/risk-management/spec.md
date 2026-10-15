## MODIFIED Requirements

### Requirement: Automation gate limits are the production Guardian policy source
The production `RiskGuardian` policy SHALL be derived from the automation gate's
five configured limits and one normalized currency. The Guardian's maximum
quantity, maximum order notional, maximum open exposure, maximum daily loss, and
maximum daily loss ratio SHALL produce a limit snapshot byte-for-byte equivalent
to the interlock's audited snapshot, including configured/set bits and currency.
The per-trade risk budget SHALL equal one third of the configured maximum daily-loss
amount in the same currency, rounded down to the currency's minor unit (KRW 1, USD
0.01); a currency with no known minor unit, or a third that rounds to zero, SHALL leave
the policy invalid so the gate refuses to start. The divisor's source is the user's
decision of 2026-10-15 (change a139); before it the budget equalled the whole daily-loss
amount ceiling, so one stopped-out trade could spend the whole amount ceiling (the
ratio-of-equity ceiling, which can be lower, is not considered by this rule). The change is in
the conservative direction only (§0.9) and SHALL NOT be relaxed without a new decision.
The production policy derives the budget in one function, the default policy carries the
same rule's value pinned equal by a test, and the startup audit SHALL record the derived value with the rule in its detail;
when the budget cannot be derived (for example a gate-off configuration with no daily-loss
amount) the audit line SHALL record an empty value with the reason and SHALL NOT fail the
start, so a gate-off start stays the upstream behaviour. The production Guardian's policy
version SHALL change with this rule (`engine.automation_gate/risk-policy-v2`), so every
decision preimage names the sizing rule it was issued under. The Guardian's daily buy notional ceiling SHALL equal the gate's
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
- **THEN** its per-trade risk budget equals one third of the configured daily-loss amount rounded down to the minor unit, and grants no larger loss budget

#### Scenario: Risk budget rounding

- **WHEN** the gate's daily-loss amount is 100000 KRW, 50 USD, or 50.75 USD
- **THEN** the per-trade risk budget is 33333 KRW, 16.66 USD, or 16.91 USD respectively

#### Scenario: Risk budget too small to size

- **WHEN** the gate's daily-loss amount is 0.02 USD so that one third rounds down to zero
- **THEN** the Guardian policy is invalid and the gate-on start is refused

#### Scenario: Gate-off start without a daily loss

- **WHEN** the automation gate is off and the configuration has no daily-loss amount
- **THEN** the start succeeds as before and the startup audit records an empty per-trade risk budget with the reason

### Requirement: 정책 수치의 provenance

모든 한도·정책 수치는 코드에 출처와 함께 기록되어야 하며(SHALL — 출처는 ① StockOS 파일·심볼과 그 검증 상태, 또는 ② TossOS 실측이다; ②는 `verify-*` change의 `measurements.md`에 남은 관측을 **식별자로 인용**해야 한다(SHALL — 관측 번호·날짜·시장; "관측했다"는 서술은 출처가 아니다)), 사용자 미확정 시 보수 기본값 전체 집합을 사용한다(SHALL — 인터록 5필드를 전부 충족): 주문당 notional 1,000,000 KRW·주문당 수량 100주·총 노출 10,000,000 KRW·일일 손실 100,000 KRW·일일 손실 자본비 1%·통화 KRW. 하루 매수 총액 한도의 기본값은 10,000,000 KRW이며(SHALL), 그 출처는 사용자 결정(2026-10-14, change a136)이다 — 이전에 상한이 없던 새 검사이므로 이 수치의 등재는 완화가 아니다. 이 기본값은 통화가 KRW인 게이트에만 적용되며 다른 통화로 환산해 적용해서는 안 된다(SHALL NOT). 최대 보유 포지션 수의 기본값은 6이며(SHALL), 그 출처는 사용자 결정(2026-10-15, change a138)이다 — 이전에 상한이 없던 새 검사이므로 이 수치의 등재는 완화가 아니다. 개수는 통화와 무관하므로 이 기본값은 모든 통화의 게이트에 적용된다(SHALL). 건당 위험 예산은 설정 수치가 아니라 일일 손실 한도의 1/3을 통화 최소 단위로 내린 유도값이며(SHALL — 기본 집합에서 33,333 KRW, USD 승인 집합에서 16.66 USD), 그 나눗수의 출처는 사용자 결정(2026-10-15, change a139)이다 — 이전 유도(일일 손실 전부)보다 작으므로 등재는 강화다. USD 축의 승인 집합은 주문당 notional 500 USD·주문당 수량 100주·총 노출 1,500 USD·일일 손실 50 USD·일일 손실 자본비 1%이며, 이 집합은 기본값이 아니라 **운영자가 명시적으로 고를 때만** 적용된다(SHALL NOT — 통화가 하나인 게이트에서 USD를 기본값으로 두면 국내 자동 진입이 조용히 닫힌다). "Toss 검증됨" 표시는 실계좌 검증 결과에만 근거해 전환한다(SHALL NOT — 미검증 상태에서 검증됨 표기 금지). 수치 변경은 §0.9(보수 방향만) 검토와 audit 기록을 요구한다(SHALL). 완화 방향의 수치 등재는 **새 통화 축을 등재하는 경우에 한해**, 각 수치가 이미 승인된 다른 통화의 선호 안쪽에 있고(SHALL — 등가 논증과 **그 논증이 깨지는 환율**을 change 문서에 기록한다), 사람이 그 수치를 명시적으로 승인했으며, 근거를 실측 식별자로 인용할 때만 허용한다(SHALL — 셋 중 하나라도 없으면 등재하지 않는다). 환산 환율은 코드에 넣지 않는다(SHALL NOT — 등가는 등재 시점의 정책 논증이며 런타임이 계산할 값이 아니다; 통화 간 정규화는 FX staleness 경계를 가진 riskcalc의 자리다).

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
