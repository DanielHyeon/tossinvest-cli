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
amount, so one stopped-out trade could spend the day's loss allowance. The change is in
the conservative direction only (§0.9) and SHALL NOT be relaxed without a new decision.
The budget is derived in one function shared by the production policy and the default
policy, and the startup audit SHALL record the derived value. The Guardian's daily buy notional ceiling SHALL equal the gate's
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
