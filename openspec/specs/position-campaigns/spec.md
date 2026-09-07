# position-campaigns Specification

## Purpose

한 position 진입을 여러 leg 로 나눠 실행할 때의 전략 중립 계약이다. campaign/leg identity 와
순서, prospective position generation 예약, broker order identity 별 누적 체결 watermark 와
교체 lineage, EXIT FIRST 진입 통제, 손절 단조 합성, 그리고 주문 없이 원장만으로 하는 offline
재구성을 정의한다.

**본선에 올라간 판본은 좁힌 판본이다.** a065 가 제안한 문장 중 구현·측정되지 않은 것은
그대로 올리지 않고, 무엇이 서 있고 무엇이 아직 배선되지 않았는지를 각 Requirement 안에
적었다. 근거는 `openspec/changes/archive/2026-09-07-a065-add-position-campaign-leg-core/issues.md`
(D-1~D-6) 에 있다.

현재 생산 배선(2026-09-08 측정, 비테스트 호출자 기준):

| 경로 | 상태 |
|---|---|
| `ApplyPositionCampaignFill` (체결 적용) | **live** — engine gateway 가 기동 때 무조건 건다 |
| `campaignExposureBlockedInTx` (EXIT FIRST 진입 판정) | **live** — first-leg dispatch 경로가 부른다 |
| campaign/leg/watermark 행 생성 | **live** — first-leg atomic 경로가 쓴다 |
| 다중 leg scale-in command 표면 | **미배선** — 호출자 0 |
| 손절 합성 (`UpdateCampaignStop`) | **미배선** — 호출자 0 |

## Requirements

### Requirement: Campaign과 leg는 전략 중립 identity와 순서를 가진다

PositionCampaign은 account, market, symbol, owning lane identity/version, originating decision/evidence, prospective generation token 및 실제 position generation을 명시적으로 연결해야 한다(SHALL). 각 CampaignLeg는 campaign 내 불변 sequence, 계획 identity, 요청 수량, intent/attempt/fill lineage와 상태를 가져야 한다(SHALL). 코어는 특정 lane의 비율, 최대 leg 수 또는 cadence 상수를 포함해서는 안 된다(MUST NOT).

중립성은 이름이 아니라 **역할**로 강제된다(SHALL). 금지 문자열을 찾는 검사는 철자를 고르면 통과하므로 계약이 될 수 없다. 강제는 둘이다: 코어 패키지의 import **폐포** 전체가 십진 산술 패키지 하나와 표준 라이브러리로 제한되고, 코어가 담은 숫자 리터럴(10진 문자열 형태 포함)의 집합이 측정으로 얼려져 있어 새 상수는 그것이 무엇이든 먼저 실패한다.

#### Scenario: 서로 다른 scale-in 전략
- **WHEN** 두 lane가 각각 8:4:2와 2:4:8 계획을 CampaignLeg 명령으로 변환한다
- **THEN** core는 동일한 ordered-leg 계약으로 기록하고 두 비율을 도메인 상수로 해석하거나 저장하지 않는다

#### Scenario: 비율표를 다른 철자로 넣기
- **WHEN** 금지 리터럴을 하나도 쓰지 않은 lane 비율표나 시장별 leg cap 상수를 코어에 넣는다
- **THEN** 숫자 열거표 검사가 그 상수를 새 값으로 보고 거부한다

#### Scenario: 청산 후 재진입
- **WHEN** CLOSED position generation의 symbol에 새 진입이 계획된다
- **THEN** 이전 campaign을 재사용하지 않고 새 campaign identity와 새 position generation lineage를 요구한다
- **AND** 이 시나리오는 아직 끝까지 성립하지 않는다: 체결된 campaign을 활성 scope 상태 밖으로
  옮기는 writer가 없어 unique index가 후속 campaign 생성을 막는다(부채 D-1, position
  수명주기를 관측하는 change의 범위)

### Requirement: Campaign은 prospective position generation을 CAS로 예약한다

첫 fill 전 campaign 생성은 `(account, market, symbol, expected_position_generation, expected_position_version)`을 compare-and-swap하고 유일한 prospective generation token을 같은 journal transaction에서 예약해야 한다(SHALL). stale expectation, active campaign 또는 이미 예약된 prospective generation이 있으면 `GENERATION_CONFLICT`로 거부해야 한다(SHALL). 첫 accepted entry fill은 기존 Position apply transaction 안에서 token을 실제 successor position generation에 set-once 결합해야 하며(SHALL), mismatch를 임의 generation에 귀속해서는 안 된다 (MUST NOT).

이 CAS 술어는 현재 저장소에 **두 벌** 있다: 코어의 생성 명령과, 생산에서 실제로 campaign 행을 만드는 first-leg atomic 경로다. 둘은 같은 술어를 강제하지만 같은 규칙의 두 번째 사본이므로 갈릴 수 있다(부채 D-6). 다중 leg를 배선하는 change는 이 둘을 하나로 합쳐야 한다(SHALL).

#### Scenario: 동시 prospective campaign 생성
- **WHEN** 두 lane가 같은 account/market/symbol과 expected generation/version으로 campaign을 동시에 생성한다
- **THEN** 한 CAS만 prospective token을 얻고 다른 생성은 GENERATION_CONFLICT이며 두 번째 campaign은 없다

#### Scenario: 첫 fill의 successor mismatch
- **WHEN** first fill apply가 prospective token이 기대한 successor와 다른 Position generation을 관측한다
- **THEN** campaign은 RECONCILE로 격리되고 fill을 추정 campaign에 귀속하지 않는다

### Requirement: Campaign 명령과 fill 적용은 멱등하고 원자적이다

모든 leg plan, submit-link, cancel 및 fill-application 명령은 deterministic command key를 가져야 하고(SHALL), 같은 key의 retry는 기존 결과를 반환하며 상태·수량·event를 두 번 전진시켜서는 안 된다(MUST NOT). expected campaign version, 해당 broker order identity의 cumulative fill watermark, event append와 projection 갱신은 하나의 journal transaction에서 검증·commit되어야 한다(SHALL). fill watermark는 leg aggregate 하나가 아니라 broker order identity별로 유지되어야 한다(SHALL).

#### Scenario: 부분체결 관측 재전송
- **WHEN** 동일 cumulative fill watermark의 broker observation이 retry 또는 restart 뒤 다시 적용된다
- **THEN** CampaignLeg와 Position 수량은 추가 증가하지 않고 기존 적용 결과가 반환된다

#### Scenario: commit 직전 crash
- **WHEN** leg fill transaction이 commit 전에 process crash를 겪는다
- **THEN** event, campaign projection과 Position apply는 모두 이전 상태이며 재시도가 한 번만 반영한다

#### Scenario: 서로 다른 order watermark 혼합
- **WHEN** 한 leg의 두 broker order identity가 서로 모순되거나 합계가 leg requested quantity를 초과한다
- **THEN** 이미 관측된 fill과 Position delta는 보존하고 산술로 잘라 맞추지 않으며 campaign/leg를 RECONCILE로 격리해 신규 exposure만 차단한다

### Requirement: terminal predecessor의 late fill은 Position에 exactly once 보존된다

replaced 또는 cancelled predecessor order의 immutable broker-order cumulative watermark가 뒤늦게 증가하면 시스템은 그 positive delta, fill evidence와 authoritative Position delta를 같은 journal transaction에서 exactly once 적용해야 한다(SHALL). 같은 transaction은 successor replacement의 remaining quantity와 leg aggregate filled/residual quantity를 재계산해야 한다 (SHALL). late delta가 leg/order cap을 초과하거나 predecessor/replacement lineage가 ambiguous해도 fill 또는 Position apply를 버리거나 truncate/rollback해서는 안 되며(MUST NOT), campaign을 `RECONCILE`로 latch하고 신규 exposure만 차단해야 한다(SHALL). stop, emergency exit, reconciliation과 fill detection은 계속되어야 한다(SHALL).

살아 있는 successor의 remaining quantity는 **자기 cap 잔여와 leg 잔여 중 작은 쪽**이어야 한다 (SHALL). cap 잔여만으로 정의하면 predecessor의 늦은 체결이 두 피연산자 어디에도 들어가지 않아 이 재계산이 no-op이 된다. 이 값의 정본은 하나여야 하며, 원장에 쓰는 쪽과 원장을 다시 읽는 쪽이 같은 함수를 불러야 한다(SHALL).

#### Scenario: cancelled predecessor late positive delta
- **WHEN** partial fill 뒤 cancelled되고 replacement가 연결된 predecessor가 더 높은 cumulative fill watermark를 보고한다
- **THEN** predecessor의 새 delta와 Position이 한 transaction에서 한 번 전진하고 replacement remaining 및 leg aggregate가 재계산되며 campaign은 RECONCILE로 신규 entry를 차단한다

#### Scenario: 늦은 체결이 후속 잔량을 실제로 줄인다
- **WHEN** predecessor의 늦은 체결이 leg 잔여를 후속 주문의 cap 잔여보다 작게 만든다
- **THEN** 후속 주문의 remaining은 leg 잔여로 낮아지고, 그 재계산을 제거하면 시험이 실패한다

#### Scenario: late fill retry
- **WHEN** 같은 predecessor cumulative watermark가 process restart 또는 observation retry로 다시 도착한다
- **THEN** per-order watermark가 이미 반영된 delta를 0으로 만들고 Position, replacement remaining과 leg aggregate는 추가로 변하지 않는다

### Requirement: Campaign 상태기계는 EXIT FIRST를 강제한다

Campaign은 PLANNED, ACTIVE, EXITING, CLOSED, RECONCILE 상태를, Leg는 PLANNED, SUBMITTED, PARTIAL, FILLED, CANCELLED, RECONCILE 상태를 가져야 한다(SHALL). 허용 전이는 완전한 versioned 표 하나로 정의되어야 하며(SHALL), 표에 없는 정상 전이를 허용해서는 안 된다(MUST NOT). 원장에 쓰는 쪽과 원장을 다시 읽는 쪽은 같은 표와 같은 사건 분류 함수를 써야 한다(SHALL) — 판정이 둘이면 각자가 상대의 시험을 통과시켜 갈림이 드러나지 않는다.

EXITING, CLOSED, RECONCILE, Position CLOSING, campaign이 묶인 position generation이 CLOSED, 또는 unresolved risk-reducing intent가 있는 동안 exposure-raising leg plan/submit을 허용해서는 안 된다(MUST NOT). stop, emergency exit, reduce-only fill, reconciliation과 fill detection은 entry 상태나 손실 latch 때문에 지연되어서는 안 된다(MUST NOT).

unresolved risk-reducing 판정은 **현재 position generation이 열린 뒤**의 intent로 한정한다 (SHALL). 경계 없는 판정에는 해소 경로가 없는 모양이 있어(attempt가 없는 intent, terminal 관측이 끝내 오지 않는 CONFIRMED 주문) 옛 intent 하나가 그 종목의 모든 진입을 영구히 막았다. 경계를 알 수 없으면 넓게 본다(SHALL — fail-closed). 이전 generation에서 broker에 남아 있는 SELL 주문은 이 판정이 보지 않으며, 그것은 주문 수명 대사의 범위다(부채 D-4).

진입 통제는 **journal admission port**에서 강제된다. D4 표의 "authoritative Position이 CLOSED면 campaign을 CLOSED로" 라는 **상태 전이 자체**는 아직 구현되지 않았다 — 그 전이를 일으킬 writer가 없다(부채 D-1). 거절은 현재 기존 refusal 코드로 기록되며, 별도의 `EXIT_FIRST_BLOCKED` 코드를 저장하려면 원장 스키마 migration이 필요하다(부채 D-2).

#### Scenario: scale-in과 stop 동시 발생
- **WHEN** 다음 entry leg와 stop exit가 같은 campaign version을 대상으로 경쟁한다
- **THEN** exit 전이가 우선 commit되고 entry leg는 exposure-blocked로 거부된다

#### Scenario: 묶인 generation이 청산됐다
- **WHEN** campaign이 결합된 position generation의 행이 CLOSED가 된 뒤 새 exposure-raising leg가 계획된다
- **THEN** 그 leg는 거부된다

#### Scenario: 이미 나간 주문의 진입 거절
- **WHEN** broker에서 이미 CONFIRMED된 주문을 원장에 연결하려는 시점에 진입이 차단돼 있다
- **THEN** 거절은 그대로 거절이되 campaign을 RECONCILE/entry-block으로 격리하고 거절 사실을 append-only 증거로 남긴다(주문을 되돌릴 수 없으므로 침묵은 그 주문의 실제 체결을 영구히 숨긴다)

#### Scenario: RECONCILE 중 체결 관측
- **WHEN** campaign이 RECONCILE 상태에서 기존 주문의 fill observation이 도착한다
- **THEN** fill detection과 Position 투영은 계속되고 새로운 exposure-raising leg만 거부된다

#### Scenario: CLOSED 뒤 새로운 fill 사실
- **WHEN** CLOSED campaign에 idempotent retry가 아닌 새 positive fill 또는 order 사실이 도착한다
- **THEN** campaign을 reopen하지 않고 event를 격리하며 account reconciliation과 신규 entry block을 발동한다

### Requirement: Effective stop은 불리한 방향으로 후퇴하지 않는다

long-only campaign의 effective stop은 이미 저장된 유효 stop보다 낮아지거나 NULL이 되어서는 안 된다(MUST NOT). 새 candidate는 유효성 검사를 통과한 경우에만 `max(saved_effective_stop, candidate_stop)`으로 합성되어야 하고(SHALL), source, policy/version, observed-at과 선택 provenance를 보존해야 한다(SHALL).

stop evidence가 missing 또는 invalid면 기존 effective stop은 변경하지 않은 채 진입만 fail closed로 건다(SHALL). 이 진입 차단 latch는 campaign 상태에 encode되지 않으므로 **단조**여야 하며(SHALL), 어떤 체결 적용도 그것을 지워서는 안 된다(MUST NOT). 이 latch를 campaign 상태에서 되계산하는 두 번째 구현이 있어서는 안 된다(MUST NOT).

이 요구는 구현돼 있으나 **아직 생산에서 실행되지 않는다** — stop을 합성하는 비테스트 호출자가 없다. 그것을 진입 경로에 강제하는 일은 lane 배선 change의 범위다.

#### Scenario: 더 낮은 새 stop
- **WHEN** 저장된 effective stop보다 낮은 유효 candidate가 제안된다
- **THEN** 저장 stop이 유지되고 candidate와 거부 provenance가 기록된다

#### Scenario: stop evidence 누락
- **WHEN** 새 leg 평가에서 stop evidence가 missing 또는 invalid다
- **THEN** 기존 effective stop은 변경되지 않고 새 exposure-raising leg는 fail closed된다

#### Scenario: 래치가 걸린 campaign에 체결이 도착한다
- **WHEN** 진입이 fail closed로 걸린 campaign에 평범한 부분체결이 적용된다
- **THEN** 체결과 수량은 정상 전진하고 진입 차단은 그대로 유지되며 다음 exposure-raising leg는 계속 거부된다

### Requirement: Offline reconstruction은 주문 없이 campaign을 재현한다

시스템은 append-only journal evidence만 사용해 prospective-generation CAS/binding, campaign 상태, ordered legs, command 결과, broker order identity별 cumulative fill watermarks와 replacement lineage, position generation lineage와 effective stop을 결정적으로 재구성해야 한다(SHALL). reconstruction은 snapshot과의 불일치를 stable reason과 마지막 valid event로 보고해야 하며 (SHALL), broker 호출, 주문 intent 생성, 상태 자동 보정 또는 운영 토글 변경을 수행해서는 안 된다(MUST NOT).

재구성이 검사하는 불변은 원장이 실제로 보장하는 것이어야 한다(SHALL). 교체 주문은 **같은 leg 위에** 새 order/attempt를 만들므로 attempt 불변은 order 수준에 두고 leg 수준에 두어서는 안 된다(MUST NOT) — leg 수준에 두면 정상적인 교체가 drift로 신고된다.

#### Scenario: 재시작 전후 동일 replay
- **WHEN** 같은 ordered journal events를 빈 projection에 replay한다
- **THEN** process restart 전 snapshot과 동일한 campaign/leg 상태와 digest가 생성된다

#### Scenario: 교체가 있는 campaign의 재구성
- **WHEN** predecessor가 부분체결 뒤 교체되고 successor가 체결된 campaign을 재구성한다
- **THEN** 재구성은 valid이며 leg identity 변경으로 신고하지 않는다

#### Scenario: 체결과 잔량 취소가 한 관측으로 도착
- **WHEN** 첫 관측이 부분체결과 잔량 취소를 함께 보고해 leg가 SUBMITTED에서 바로 닫힌다
- **THEN** 원장이 기록한 상태를 재구성이 같은 표로 유도하며 drift로 신고하지 않는다

#### Scenario: 순서가 끊긴 leg event
- **WHEN** replay 중 campaign-local leg sequence gap이 발견된다
- **THEN** reconstruction은 LEG_SEQUENCE_GAP과 마지막 valid event를 보고하고 누락 leg를 추정하지 않는다
