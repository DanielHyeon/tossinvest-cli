## Context

현재 journal의 Position은 fill과 adjustment의 권위 있는 투영이고 exit lifecycle도 position generation에 귀속된다. 그러나 한 전략이 여러 차례 진입을 계획할 때 계획 단위, 부분체결, 재시도와 재시작을 하나로 묶는 도메인 identity가 없다. 후속 lane마다 별도 scale-in 상태를 만들면 fill 적용과 exit 책임이 분기된다.

campaign core는 이 공백만 채우며 비율·시간 간격·시장 전략을 알지 않는다. 기존 Position 수량, fill watermark, reconciliation과 exit policy가 계속 단일 권위다.

## Goals / Non-Goals

**Goals:**

- first fill 전 prospective generation부터 실제 position generation까지 ordered leg의 영속 identity와 상태기계를 제공한다.
- 계획·제출·부분체결·재시작을 멱등 처리한다.
- EXIT FIRST와 long-only effective stop의 비후퇴를 구조적 불변식으로 만든다.
- journal evidence만으로 campaign을 offline 재구성하고 drift를 탐지한다.

**Non-Goals:**

- 8:4:2, 2:4:8, 7-leg 같은 lane별 상수 또는 다음 leg 조건
- broker 제출, live caller, lane 활성화나 위험 수량 계산
- Position 수량을 campaign projection으로 대체
- reconcile, fill detection 또는 emergency exit 주기 변경

## Decisions

### D1. Campaign과 Leg는 전략 중립 aggregate다

`PositionCampaign`은 account, market, symbol, owning lane identity/version, originating evidence/decision, prospective generation token과 실제 position generation을 연결한다. `CampaignLeg`는 campaign-local sequence, plan identity, requested quantity, intent/attempt 참조와 lifecycle을 가진다. 체결 watermark는 leg aggregate 한 개가 아니라 immutable broker order identity별 child record가 소유한다. 전략 payload는 opaque plan provenance로 참조할 수 있지만 core schema에 비율·cadence column을 두지 않는다.

lane별 별도 상태기계를 택하지 않는 이유는 fill과 exit 불변식이 전략 수만큼 복제되는 것을 막기 위해서다.

### D2. Campaign 생성은 prospective generation CAS를 선행한다

Position이 아직 없거나 이전 generation이 CLOSED인 상태에서 campaign을 만들 때 caller는 `(account, market, symbol, expected_position_generation, expected_position_version)`을 제공한다. journal transaction은 현재 Position projection이 그 기대값과 일치하고 active/prospective campaign이 없음을 검사한 뒤 유일한 `prospective_generation_token`을 예약한다. 동시 caller 하나만 성공하며 stale 기대값은 `GENERATION_CONFLICT`로 거부된다.

첫 accepted entry fill의 기존 Position apply transaction은 successor generation을 생성하면서 prospective token을 실제 `position_generation_id`에 set-once 결합한다. 기대 successor와 다르거나 다른 generation이 먼저 나타나면 수량을 추정 귀속하지 않고 campaign을 RECONCILE로 격리한다. in-memory symbol lock이나 first-fill 이후 owner 생성은 crash/race 창을 남기므로 권위로 사용하지 않는다.

### D3. append-only event와 expected-version transaction을 권위로 둔다

명령은 `(campaign_id, command_kind, deterministic command key)`로 멱등 처리한다. journal transaction은 expected campaign version과 관련 broker order watermark를 비교한 뒤 event append, projection, 명시적 lineage 참조를 함께 commit한다. crash는 전부 commit 또는 전부 rollback이다.

mutable row만 갱신하는 대안은 offline reconstruction과 duplicate 원인 추적이 불가능하므로 기각한다. projection은 읽기 최적화이며 event replay 결과와 불일치하면 RECONCILE로 격리한다.

### D4. Campaign과 Leg의 완전한 상태 전이표를 사용한다

Campaign 전이의 정본은 다음 표다. 표에 없는 정상 명령은 허용하지 않는다.

| 현재 | 사건/증거 | 다음 | 결과 |
|---|---|---|---|
| PLANNED | 동일 command retry 또는 leg plan | PLANNED | 기존 결과 반환 또는 ordered leg 추가 |
| PLANNED | 첫 order submit/link | ACTIVE | exposure-raising 제출 lineage 시작 |
| PLANNED | fill 전 전부 취소 또는 structural invalidation | CLOSED | prospective token 종결, 재사용 금지 |
| PLANNED | lineage/generation 불일치 | RECONCILE | 신규 entry 차단 |
| ACTIVE | 유효 leg plan/submit/fill/cancel | ACTIVE | leg만 표에 따라 전진 |
| ACTIVE | risk-reducing intent, Position CLOSING 또는 exit pending | EXITING | 신규 exposure 차단 |
| ACTIVE | authoritative Position CLOSED/수량 0 | CLOSED | campaign 종결 |
| ACTIVE | version/watermark/generation 불일치 | RECONCILE | 신규 entry 차단 |
| EXITING | risk-reducing fill/observation | EXITING | exit 추적 계속 |
| EXITING | authoritative Position CLOSED/수량 0 | CLOSED | campaign 종결 |
| EXITING | lineage/watermark 불일치 | RECONCILE | safety loop는 계속 |
| RECONCILE | compare-and-append 증거가 OPEN/SCALING 및 exit 부재를 증명 | ACTIVE | 명시적 recovery event 기록 |
| RECONCILE | 증거가 CLOSING/exit pending을 증명 | EXITING | 명시적 recovery event 기록 |
| RECONCILE | 증거가 CLOSED/수량 0/pending 부재를 증명 | CLOSED | 명시적 recovery event 기록 |
| RECONCILE | 불완전·상충 증거 | RECONCILE | 추정 전이 금지 |
| CLOSED | 동일 terminal event retry | CLOSED | 기존 결과 반환 |
| CLOSED | 새로운 fill/order/leg 사실 | CLOSED | event 격리, account reconcile/entry block; reopen 금지 |

Leg와 order 전이의 정본은 다음 표다.

| 현재 | 사건/증거 | 다음 | 결과 |
|---|---|---|---|
| PLANNED | 동일 plan retry | PLANNED | 기존 leg 반환 |
| PLANNED | order identity link/submit | SUBMITTED | 첫 per-order watermark 생성 |
| PLANNED | submit 전 cancel | CANCELLED | terminal zero-fill |
| PLANNED | order lineage 없는 positive fill | RECONCILE | fill 귀속 추정 금지 |
| SUBMITTED | 동일/낮은 cumulative order watermark | SUBMITTED | delta 0 |
| SUBMITTED | `0 < cumulative < order cap` | PARTIAL | 새 delta만 적용 |
| SUBMITTED | cumulative가 order cap 도달 | FILLED | terminal full-fill |
| SUBMITTED | zero-fill cancel/expiry | CANCELLED | terminal zero-fill |
| SUBMITTED | amend/replacement lineage | SUBMITTED | predecessor와 carry baseline을 가진 새 order watermark |
| PARTIAL | 동일/낮은 cumulative order watermark | PARTIAL | delta 0 |
| PARTIAL | 증가하되 cap 미만 | PARTIAL | 새 delta만 적용 |
| PARTIAL | cap 도달 | FILLED | terminal full-fill |
| PARTIAL | 잔량 cancel/expiry | CANCELLED | partial quantity 보존, terminal residual-cancel |
| PARTIAL | amend/replacement lineage | PARTIAL | predecessor와 carry baseline을 가진 새 order watermark |
| RECONCILE | complete broker/position evidence | SUBMITTED/PARTIAL/FILLED/CANCELLED | 명시적 recovery event와 재도출 근거 기록 |
| RECONCILE | 불완전·상충 evidence | RECONCILE | 추정 전이 금지 |
| FILLED/CANCELLED | 동일 terminal observation retry | 동일 | delta 0 |
| FILLED/CANCELLED | replaced/cancelled predecessor의 새 positive delta | 동일 | watermark와 Position을 tx에서 exactly once 전진, replacement remaining/leg aggregate 재계산, campaign RECONCILE 및 신규 entry 차단 |
| FILLED/CANCELLED | immutable broker order identity는 확정됐지만 replacement lineage가 ambiguous한 새 positive delta | 동일 | fill/Position과 관측 evidence는 보존, campaign RECONCILE 및 신규 entry 차단; 추정 replacement 귀속 금지 |

fill의 실제 Position 반영은 기존 tx-scoped apply hook가 계속 소유한다. campaign hook는 같은 transaction에서 order watermark delta와 leg/campaign event만 전진시켜 이중 권위를 만들지 않는다.

### D5. fill 멱등성은 broker order identity별 watermark가 소유한다

각 order record는 immutable broker order identity, submit/amend/replacement attempt identity, predecessor, carry baseline, requested cap, last cumulative filled quantity와 last observation identity를 가진다. cumulative watermark는 후퇴할 수 없고 새 적용량은 해당 order의 검증된 증가분뿐이다. replacement가 새 order identity를 만들면 predecessor의 carry baseline을 명시해 이전 fill을 다시 더하지 않는다.

replaced 또는 cancelled predecessor의 cumulative watermark가 뒤늦게 증가하면 terminal order state를 reopen하지 않더라도 그 immutable broker order identity의 positive delta와 authoritative Position apply는 같은 journal transaction에서 exactly once 전진해야 한다. 그 transaction은 successor replacement의 remaining quantity와 leg aggregate filled/residual quantity를 새 사실에 맞게 재계산한다. 이미 요청 cap을 넘었거나 predecessor/replacement lineage가 ambiguous해도 실제 fill을 버리거나 cap에 맞춰 truncate/rollback해서는 안 된다. 대신 원 fill evidence와 Position을 보존하고 campaign을 `RECONCILE`로 latch해 신규 exposure를 차단한다. retry는 이미 전진한 per-order watermark 때문에 delta 0이고, commit 전 crash는 watermark, Position, replacement remaining과 campaign projection을 모두 rollback한다.

모든 order delta 합이 leg requested quantity를 넘거나 lineage가 ambiguous하면 수량을 산술 보정하지 않고 위 보존 transaction 뒤 RECONCILE로 간다.

leg-level watermark 하나는 서로 다른 order의 cumulative 숫자를 충돌시키거나 replacement fill을 이중 계상할 수 있으므로 기각한다.

### D6. EXIT FIRST를 command admission에 고정한다

position/campaign이 EXITING, CLOSING, RECONCILE이거나 unresolved risk-reducing intent가 있으면 exposure-raising leg plan/submit은 거부한다. risk-reducing fill과 emergency/reconciliation observation은 campaign 상태와 무관하게 기존 우선 경로에서 처리하고, exit 전이는 pending entry leg를 먼저 취소 가능 상태로 만든다.

단순 queue 우선순위보다 admission invariant를 택한 이유는 process restart나 서로 다른 worker에서도 entry가 exit를 추월하지 못하게 하기 위해서다.

### D7. stop은 provenance를 보존하며 단조 합성한다

long-only effective stop은 기존 saved effective stop과 새 validated candidate 중 더 높은 보호 가격이다. missing/invalid candidate가 기존 stop을 NULL 또는 낮은 값으로 바꾸지 못한다. 원본 후보, 선택 결과, policy/source와 observed-at을 함께 저장하며 실제 broker protection mutation은 후속 change의 책임이다.

### D8. offline reconstruction은 read-only 검증 경로다

reconstructor는 event 순서, command key uniqueness, prospective-generation CAS/binding, leg sequence, order별 cumulative fill watermark와 replacement carry lineage, position generation lineage와 effective stop 단조성을 검증해 snapshot과 비교한다. 불일치는 안정 reason과 마지막 valid event를 보고하며 자동 주문이나 임의 수정을 수행하지 않는다.

## Risks / Trade-offs

- [campaign과 Position의 이중 권위] → 수량은 Position만 소유하고 campaign은 계획·lineage·watermark만 투영한다.
- [동시 campaign/leg 명령] → prospective-generation CAS, expected campaign version과 unique command key를 같은 journal transaction에서 검사한다.
- [amend/replacement 이중 체결] → order identity별 monotone watermark와 predecessor carry baseline을 검증한다.
- [replacement/cancel 뒤 predecessor late fill] → terminal state와 무관하게 predecessor watermark와 Position을 tx exactly once 전진하고 successor remaining/leg aggregate를 재계산한 뒤 RECONCILE로 신규 entry만 차단한다.
- [exit와 scale-in race] → EXIT FIRST admission과 tx 재검사로 exposure-raising commit을 거부한다.
- [event 증가로 replay 비용 상승] → 검증 가능한 snapshot을 사용하되 event가 정본이고 주기적 offline replay로 drift를 검사한다.
- [schema rollback] → additive-nullable migration과 ErrSchemaTooNew를 유지한다.

## Migration Plan

1. campaign/leg event, prospective-generation token과 per-order watermark/projection 구조를 additive migration으로 추가한다.
2. 두 완전 상태표, generation CAS, duplicate command, amend/replacement partial fill, replaced/cancelled predecessor late fill, cap 초과·ambiguous lineage, crash boundary와 replay fixture를 먼저 검증한다.
3. 기존 journal을 read-only scan해 campaign 부재를 정상 legacy 상태로 보고한다. 기존 포지션을 추정 campaign으로 backfill하지 않는다.
4. production entry caller 없이 shadow plan/reconstruction API만 배포한다.
5. 후속 lane/runtime change가 명시적 승인 후 campaign command를 호출한다.

Rollback은 caller를 제거하고 구버전 호환 DB를 사용한다. migration 적용 DB를 구버전 바이너리로 열 경우 기존 계약대로 ErrSchemaTooNew로 중단하며 데이터를 변환하지 않는다.

## Open Questions

- 없음. 전략별 leg 수와 cadence는 각 lane change가 versioned 상수로 정의한다.
