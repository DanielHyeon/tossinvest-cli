## ADDED Requirements

### Requirement: 성공한 exit 관측은 상태 변화와 독립적으로 권위 snapshot을 최신화한다
시스템은 관리 포지션의 유효하고 신선한 공식 가격 관측을 성공적으로 평가할 때 가격·high-water·보호선·rung·action이 변하지 않았더라도 권위 exit snapshot의 관측 증거를 영속해야 한다 (SHALL). `SEED`의 첫 성공 관측은 완전한 canonical `EVALUATED` snapshot을 만들어야 하며 (SHALL), 이후 의미가 동일한 관측은 snapshot provenance와 freshness만 원자적으로 최신화하고 exit transition event, proposal, intent, cancel 또는 broker order를 생성해서는 안 된다 (MUST NOT).

#### Scenario: ratchet의 첫 가격이 seed와 같다
- **WHEN** `SEED` ratchet 포지션의 첫 유효 공식 가격이 entry·high-water와 같고 exit 조건을 넘지 않는다
- **THEN** 하나의 canonical `EVALUATED` snapshot과 첫 평가 event가 영속되고 proposal·cancel·order는 0건이다

#### Scenario: ladder의 첫 가격이 adoption t0와 같다
- **WHEN** 자동 편입된 common-ladder 포지션의 첫 유효 공식 가격이 adoption t0와 같다
- **THEN** current protection, next target, next protection을 가진 `EVALUATED` snapshot이 영속되고 편입 또는 첫 평가는 주문을 만들지 않는다

#### Scenario: 같은 가격의 후속 관측
- **WHEN** `EVALUATED` 포지션이 동일한 운영 line을 산출하는 더 최신의 동일 가격을 관측한다
- **THEN** 최신 observation identity와 observed-at이 원자적으로 영속되며 exit event 수, proposal 상태, policy state와 모든 주문 adapter 호출 수는 변하지 않는다

#### Scenario: 가격은 움직였지만 exit line은 같다
- **WHEN** 새 유효 가격이 이전 observed price와 다르지만 protection·high-water·rung·action·projection을 바꾸지 않는다
- **THEN** 새 observed price와 provenance는 refresh되고 semantic exit event는 추가되지 않는다

#### Scenario: 의미가 달라진 후속 평가
- **WHEN** 가격은 같더라도 잔량 변화가 projection을 바꾸거나 pending suppression 등 다른 운영 line 필드가 이전 effective snapshot과 달라진다
- **THEN** 시스템은 freshness-only 갱신으로 덮지 않고 완전한 judgement 경로에서 monotone recovery와 proposal 규칙을 다시 적용한다

#### Scenario: 동시에 더 강한 상태가 기록된다
- **WHEN** flat refresh가 읽은 뒤 같은 lifecycle의 보호선 또는 rung이 다른 transaction에서 전진한다
- **THEN** refresh transaction은 conflict로 아무것도 쓰지 않고 더 강한 권위 상태를 보존한다

#### Scenario: 최신 refresh 뒤 오래된 refresh가 commit을 시도한다
- **WHEN** 더 최신 observation이 먼저 영속된 뒤 같은 line의 더 오래된 observation refresh가 도착한다
- **THEN** 오래된 refresh는 typed stale conflict로 아무 column도 되돌리지 않는다

#### Scenario: 같은 observation이 재전달된다
- **WHEN** 같은 observed-at과 observation identity의 refresh가 다시 전달된다
- **THEN** transaction은 event뿐 아니라 exit state의 `updated_at`도 바꾸지 않는 성공 no-op이다

#### Scenario: frozen clock의 cycle fallback이 전진한다
- **WHEN** zero-fetched-at compatibility source가 같은 engine timestamp에서 더 큰 persisted `cycle:<sequence>`로 다음 유효 관측을 제출한다
- **THEN** refresh는 sequence 순서로 전진하며 낮은 sequence, 같은 sequence의 다른 identity 또는 official same-time/different-identity evidence는 no-write conflict다

#### Scenario: 두 observer가 같은 cycle sequence를 경쟁한다
- **WHEN** 두 observer가 같은 durable maximum에서 같은 `cycle:N`과 timestamp를 발급했지만 서로 다른 price/observation identity로 refresh를 경쟁한다
- **THEN** 먼저 commit된 tuple만 남고 다른 identity는 typed conflict로 state·updated-at·event를 바꾸지 않는다

#### Scenario: restart가 durable cycle sequence를 복구한다
- **WHEN** working set에 `cycle:10` snapshot이 영속된 상태에서 observer가 restart하고 engine clock도 같은 timestamp에 머문다
- **THEN** observer는 working set maximum을 복구해 다음 fallback을 `cycle:11` 이상으로 발급하며 1부터 재시작해 freshness를 굶기지 않는다

#### Scenario: 같은 timestamp에서 fallback과 official source가 교차한다
- **WHEN** persisted `cycle:N` 뒤 같은 timestamp의 valid `quote_fetched_at`이 오거나 persisted official 뒤 같은 timestamp의 fallback이 온다
- **THEN** official은 cycle을 대체할 수 있지만 cycle은 official을 대체할 수 없고, 다른 두 official identity는 ambiguous conflict다

### Requirement: 실패한 관측은 exit-line freshness를 연장하지 않는다
시스템은 누락·invalid·non-finite·future·source-stale 가격, policy/generation 불일치 또는 journal refresh 실패를 성공 관측으로 기록해서는 안 된다 (MUST NOT). 마지막 성공 snapshot은 보존하되 freshness 시계는 연장하지 않아야 한다 (SHALL).

#### Scenario: invalid flat quote
- **WHEN** 마지막 성공 평가 뒤 invalid 또는 source-stale quote만 도착한다
- **THEN** persisted observed-at과 observation identity는 바뀌지 않고 age bound를 넘으면 line은 stale이 된다

#### Scenario: 공식 quote 시각 경계
- **WHEN** batched price 응답의 non-zero fetched-at이 post-read engine clock보다 미래이거나 15초를 초과해 오래됐다
- **THEN** 해당 symbol은 judgement·proposal·refresh에서 제외되며 정확히 15초 old인 quote만 허용된다

#### Scenario: HTTP 성공이지만 모든 quote evidence가 invalid다
- **WHEN** batched Prices 호출은 성공했지만 managed symbol의 quote가 모두 non-positive·non-finite·future 또는 source-stale이다
- **THEN** batch는 typed non-retryable evidence failure로 처리되어 추가 broker retry 없이 `QueryPrice` freshness와 exit outage clock을 성공으로 갱신하지 않는다

#### Scenario: valid sibling과 invalid sibling이 함께 있다
- **WHEN** 한 batched 응답에 유효한 managed-symbol quote와 invalid quote가 함께 있다
- **THEN** query success는 한 번 기록되지만 유효한 symbol만 judgement 대상이고 invalid sibling의 snapshot freshness는 바뀌지 않는다

#### Scenario: 뒤쪽 position 전에 quote가 만료된다
- **WHEN** batched quote가 처음에는 유효했지만 앞선 position 처리 지연으로 뒤쪽 position의 judgement 또는 record 직전 15초 use deadline을 넘는다
- **THEN** 뒤쪽 position은 record·refresh·clear·proposal·order를 수행하지 않고 다음 cycle을 기다린다

#### Scenario: wall clock rollback은 quote lease를 연장하지 않는다
- **WHEN** quote batch 뒤 wall clock이 뒤로 움직여 absolute timestamp상 아직 유효해 보이지만 process-local monotonic elapsed가 15초를 초과했거나 현재 wall time이 official fetched-at보다 앞선다
- **THEN** 해당 position은 judgement·record·refresh·clear·proposal·order를 수행하지 않으며 정확히 15초 elapsed인 경우만 허용된다

#### Scenario: fetched-at이 없는 호환 source
- **WHEN** 성공한 price batch의 quote가 zero fetched-at을 제공한다
- **THEN** 시스템은 post-read engine clock과 durable working-set maximum 다음의 persisted source `cycle:<sequence>`를 사용하되 non-zero invalid 시각을 이 fallback으로 바꾸지 않는다

#### Scenario: refresh journal error
- **WHEN** 의미가 동일한 fresh quote의 snapshot refresh transaction이 실패한다
- **THEN** exit state·event·proposal은 변경되지 않고 해당 관측은 화면 freshness의 증거가 되지 않는다

### Requirement: flat 관측은 기존 cadence와 실행 안전 경계를 보존한다
flat 관측 지원은 기존 exit quote 호출 수와 rate priority를 늘려서는 안 되며 (MUST NOT), pending proposal·re-judgement·실제 state/action 변화의 durable-record-before-submit 순서를 바꾸어서는 안 된다 (MUST NOT).

#### Scenario: 5초 cadence의 flat position
- **WHEN** 한 관리 포지션이 여러 cycle 동안 같은 가격으로 관측된다
- **THEN** 전체 managed symbol set에 대한 기존 batched Prices 호출 1건 이외의 broker 호출은 없고 local snapshot refresh만 수행된다

#### Scenario: 실제 보호선 breach
- **WHEN** flat refresh 사이에 유효 가격이 current protection을 침범한다
- **THEN** 기존 complete judgement, durable arm, submit 순서로 정확히 한 보호 proposal이 처리된다
