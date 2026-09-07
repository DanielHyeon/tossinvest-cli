## ADDED Requirements

### Requirement: 콘솔은 KR과 US strategy runtime을 독립 projection으로 표시한다
콘솔은 KR과 US strategy runtime을 독립 projection으로 표시해야 한다 (SHALL). Shared
server-owned runtime projection을 사용해 KR과 US 각각의 lane desired/effective, evidence
freshness/digest, campaign/leg, horizon risk bucket, scheduler/calendar, activation,
ProtectionReady, reconciliation health, first typed refusal과 observed-at을 read-only로 표시해야
한다 (SHALL). ProtectionReady 값은 정확히 `WIRED` 또는 `UNWIRED`만 사용하고 실패 상세는
별도 typed refusal로 표시해야 한다 (SHALL). 한 시장을 읽지 못하면 해당 시장 status/error
envelope만 unavailable, ProtectionReady는 `UNWIRED`와 typed reason으로 fail-closed하고 다른
시장의 current snapshot을 유지해야 한다 (SHALL). Unavailable을 default, 0 또는 다른 시장의
값으로 대체해서는 안 된다 (MUST NOT). 이 surface는 order, gate, lane/LIVE activation,
autostart 또는 protection-weakening mutation control을 제공해서는 안 된다 (MUST NOT).

#### Scenario: KR blocked와 US eligible
- **WHEN** KR은 stale evidence로 effective OFF이고 US는 current evidence와 scheduler로 eligible이다
- **THEN** 두 market card가 각각의 desired/effective와 first refusal을 표시하고 KR 상태를 US에 복제하지 않는다

#### Scenario: US runtime unavailable
- **WHEN** runtime endpoint가 KR snapshot은 반환하지만 US snapshot을 읽지 못한다
- **THEN** KR current state는 유지되고 US status만 unavailable, ProtectionReady는 `UNWIRED`와 runtime-unavailable refusal이며 0/default 또는 제3 readiness enum으로 꾸미지 않는다

#### Scenario: dormant 배포
- **WHEN** services가 배포됐지만 lane, autostart, automation과 LIVE approval이 활성화되지 않았다
- **THEN** 두 market card는 desired/effective OFF 또는 not-configured와 구체적 blocker를 표시하고 entry-ready로 표시하지 않는다

#### Scenario: 콘솔 권한 회귀
- **WHEN** multi-market runtime 화면의 route table과 HTML을 검사한다
- **THEN** 신규 order/gate/LIVE/activation/protection mutation route와 자유 입력 control은 0건이다

### Requirement: 콘솔 성과는 market과 campaign lineage를 설명한다
콘솔은 성과의 market과 campaign lineage를 설명해야 한다 (SHALL). Lane performance를 market,
lane/version, campaign/leg와 policy version으로 구분하고 완전한 persisted identifier chain만
attributed sample에 포함해야 한다 (SHALL). Partial fill/staged close에는 closed/residual quantity,
gross PnL, entry/exit fees, taxes, FX source/rate/as-of와 net PnL을 함께 표시해야 한다 (SHALL).
Lineage 또는 observation이 없으면 각각 `link_missing`, `not_measured`와 누락 식별자를 표시하고
(SHALL), symbol/time 또는 동일 ticker로 다른 시장·campaign에 귀속해서는 안 된다 (MUST NOT).

#### Scenario: 동일 ticker의 다른 시장
- **WHEN** KR과 US에 동일 문자열 ticker가 있지만 persisted market/campaign lineage는 US만 완전하다
- **THEN** US sample만 해당 US lane/campaign에 귀속되고 KR row를 symbol로 추정하지 않는다

#### Scenario: campaign leg 누락
- **WHEN** closed trade에 lane/version은 있지만 campaign 또는 leg identifier가 없다
- **THEN** attributed result에서 제외하고 link_missing과 누락 필드를 표시한다

#### Scenario: staged close 잔여 수량
- **WHEN** position의 일부만 close fill되고 나머지 수량은 authoritative journal에서 open이다
- **THEN** realized close quantity와 residual open quantity를 분리하고 전체 position을 closed trade로 표시하지 않는다

#### Scenario: FX evidence 누락
- **WHEN** US close fill의 원 currency PnL은 있지만 reporting currency FX source/rate/as-of가 없다
- **THEN** 원 currency metric은 보존하고 reporting-currency net PnL은 `not_measured`로 표시하며 0 또는 current FX로 대체하지 않는다
