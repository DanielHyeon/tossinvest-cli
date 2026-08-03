## ADDED Requirements

### Requirement: KR과 US scheduler binding은 동시에 독립 진행한다
Runtime scheduler는 KR과 US binding을 동시에 독립 진행해야 한다 (SHALL). KR과 US 각각에 별도 calendar generation, activation manifest, session
decision과 endpoint budget binding을 유지하면서 두 시장을 동시에 진행해야 한다 (SHALL).
한 시장의 closed/stale/refused/budget-deferred state를 다른 시장 decision의 입력으로 사용해서는
안 되며 (MUST NOT), per-market binding을 결합한 하나의 승인 scope를 만들거나 복원해서도 안
된다 (MUST NOT). Safety class budget과 exit/reconciliation cadence는 두 entry binding보다
우선해야 한다 (SHALL).

#### Scenario: 동시 상이 상태
- **WHEN** KR은 current calendar에서 WAIT_MARKET이고 US는 ENTRY_ALLOWED다
- **THEN** 두 decision을 같은 cycle에 독립 반환하고 US candidate cadence를 KR 상태 때문에 지연하지 않는다

#### Scenario: 한 시장 budget provenance 누락
- **WHEN** US entry endpoint budget provenance가 stale이고 KR endpoint budget은 current다
- **THEN** US entry poll만 BUDGET_DEFERRED이며 KR entry poll과 두 시장 safety class 호출은 계속된다

#### Scenario: 재시작 복원 범위
- **WHEN** KR auto-resume manifest만 유효하고 US manifest는 만료됐다
- **THEN** KR scheduler binding만 복원하고 US entry는 OFF로 유지하며 새로운 LIVE approval을 만들지 않는다
