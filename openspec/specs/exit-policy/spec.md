# exit-policy Specification

## Purpose
손익 극대화 exit 정책 — t0 기준선(진입 손절)·baseline ratchet(워터마크 단조 상승)·profit ladder·발의 pending 수명주기·관측 fail-safe·비용 비게이트 요구사항.

## Requirements

### Requirement: t0 기준선 — 진입 손절가가 첫 독자를 갖는다
포지션이 열리는 순간 보호 기준선은 진입 결정의 손절가(RiskIntent.stop)로 초기화되어야 한다(SHALL — `exit_states.baseline_price`는 NOT NULL이며 빈 기본값이 없다). 기준선 하회 발의 규칙은 t0부터 유효하다(SHALL) — "No Stop = No Trade"의 손절가는 사전 검사 입력에 그치지 않고 체결 후 관측 판정의 기준선이 된다. entry 결정이 없는 포지션(외부·수동 편입)은 exit 정책의 대상이 아니며 발견 시 알림을 발송한다(SHALL — R·기준선의 근거가 없다).

#### Scenario: 개시 직후 손절 하회
- **WHEN** 진입 체결 직후(+0.4R 도달 전) 관측 가격이 진입 손절가를 하회하면
- **THEN** 전량 청산이 RISK_REDUCING 의도로 발의된다

#### Scenario: entry 결정 없는 포지션
- **WHEN** 조정 이벤트로 편입된 외부 포지션이 존재하면
- **THEN** exit 판정이 수행되지 않고 운영자 알림이 발송된다

### Requirement: Baseline Ratchet — 관측 최고가 기반 단조 상승
보호 기준선은 R 트리거로 단계 상승해야 한다(SHALL — StockOS `exit/baseline_ratchet.py` 이식, 수치는 provenance 주석과 §0.9 보수 방향 잠금):

| 트리거(도달 R) | 레벨 | 기준선 후보(R) |
|---|---|---|
| +0.4R | HALF_RISK | −0.5R |
| +0.8R | BREAKEVEN | 실질 본전(비용 차감 `break_even_sell_price`) |
| +1.0R | (레벨 유지) | 부분익절 40% 발의 |
| +1.2R | PARTIAL_LOCK | +0.3R |
| +2.0R | PROFIT_LOCK | +0.8R |

R의 프로브는 **관측 최고가 워터마크**다(SHALL — `exit_states.high_water`, 관측마다 단조 갱신; 원본 `high_since_entry` 의미론). 워터마크는 관측 표본의 최고가이지 참 최고가가 아니며, 표본 사이에 스친 트리거는 놓칠 수 있다(SHALL 명시 — 관측 주기가 허용 오차를 정의한다). 레벨은 워터마크의 함수이므로 단조이며(SHALL), 기준선은 후보 합성으로 결정된다(SHALL — 원본 `compute_protected_stop` 이식): `새 기준선 = max(이전 기준선, 레벨 후보, 레벨 ≥ BREAKEVEN일 때 실질 본전)`, 상승은 strict `>`일 때만 기록한다. R의 분모는 진입 시 확정된 초기 위험(`entry − initial_stop` — `exit_states`에 영속)이며 부분익절·조정 이벤트 후에도 변하지 않는다(SHALL). 입력 검증(초기 손절 ≥ 진입가 등 불변식 위반)은 판정 거부 + 알림이다(SHALL — 원본 `_validate_inputs`).

#### Scenario: 폴링 사이 고점 후 되돌림
- **WHEN** 관측 A(+1.3R)와 관측 B(+0.6R)가 연속되면
- **THEN** 워터마크는 +1.3R로 남아 PARTIAL_LOCK(+0.3R)이 유지되고, 기준선은 하강하지 않는다

#### Scenario: 기준선 단조 property
- **WHEN** 임의의 가격 관측 시퀀스가 평가되면
- **THEN** 기준선·레벨·워터마크 모두 비감소이다 (property 테스트 — 셋 다 검증)

### Requirement: 발의 수명주기 — 레벨당 1회, 미해소 중 억제
부분익절·청산 발의는 pending 수명주기를 가져야 한다(SHALL — `exit_states.pending_action`·`pending_level`(ratchet) 또는 `pending_level`(ladder에서는 rung 인덱스)·`pending_intent_id`): 같은 레벨/rung의 발의는 최대 1회이며(SHALL — ratchet의 40% 부분익절은 레벨 승격과 무관하게 **포지션당 1회**: 원본은 매 평가 재제안을 호출자 dedup에 맡기지만 TossOS는 `taken_ratio_total > 0`이면 ratchet 부분익절을 재발의하지 않는다), 발의가 미해소(미체결·IN_DOUBT)인 동안 새 발의는 억제된다(SHALL NOT 중복 발의). 발의의 체결·거부·취소가 pending을 해소하고, 거부·취소 시 해당 레벨은 재발의 가능해진다(SHALL — 크래시 후 재시작은 pending을 복원해 미재발의·중복발의 둘 다 방지한다). 부분익절 체결 시 `taken_ratio_total`이 이동하며(체결 시점 필드 — 체결 반영 트랜잭션의 원자 apply hook에서만), 누적 비율은 초기 수량 기준, 각 발의 비율은 잔여 수량 기준이다(SHALL — 원본 분모 규칙).

#### Scenario: 미체결 중 재관측
- **WHEN** 40% 부분익절 발의가 미체결인 상태에서 다음 관측이 +1.0R 이상이면
- **THEN** 새 발의는 발생하지 않는다

#### Scenario: 제출 전 크래시
- **WHEN** 발의 기록 후 제출 전에 크래시·재시작되면
- **THEN** pending이 복원되어 같은 레벨이 중복 발의 없이 재개된다

### Requirement: 포지션당 정책 하나
포지션의 exit 정책은 RATCHET 또는 LADDER 중 하나다(SHALL — `exit_states.policy_kind`, 기본값 RATCHET·LADDER는 설정 지정(원본 policy_assignment DEFAULT_ASSIGNMENT 구조); 원본에서 두 모듈은 대안 실행 경로이며 한 포지션의 기준선을 동시에 다투지 않는다). profit ladder는 rung 표(목표%·잠금%·부분비율 — 기본 세트 1.5/2.5/4.0/6.0% 목표, 0/1.0/2.0/3.5% 잠금, 0/0.25/0.25/1.0 부분, `[미검증 — StockOS KOSPI 튜닝값]` provenance)로 정의되고, rung 목표는 단조 증가·잠금은 비감소여야 한다(SHALL — 정책 검증). rung 도달 판정도 워터마크 기반이며 pending 수명주기를 공유한다(SHALL). 원본의 STOP_FIRST 동시 관측 모델은 OHLC 입력이 있는 백테스트 전용이므로 이 change의 SHALL이 아니다(P3 백테스트 도입 시 함께).

#### Scenario: ladder 포지션의 rung 진행
- **WHEN** LADDER 정책 포지션이 rung 1 목표에 도달하면
- **THEN** rung 1 부분익절이 발의되고 잠금이 승격되며, 체결 반영 후 rung 2가 활성화된다

### Requirement: 관측 경로와 fail-safe
exit 판정의 입력은 보유 심볼의 최신가 1점 관측이다(SHALL — 시계열 저장 아님; 주기 기본 5초·보유 심볼 fan-out·§0.4 rate budget 내·체결 감지 SLO에 양보). 관측 실패 시 기준선·워터마크는 유지되고 판정은 보류된다(SHALL). 단, 관측 두절이 staleness 임계(기본 60초)를 넘으면 critical 알림 + ENTRY_BLOCKED 자동 강화가 발동한다(SHALL — "보류"가 무기한 무손절이 되어서는 안 된다; 브로커측 보호가 없는 구성에서 관측 두절은 보호 부재와 같다). 청산 발의가 RECONCILE 확정 하한으로 캡되면 잔여는 pending으로 유지되어 해소 후 같은 레벨 identity로 재발의되며, 캡 발생은 알림된다(SHALL). 미해소 진입 attempt가 같은 심볼의 청산 발의를 지연시키는 창은 해소 절차의 우선·유계성이 한정하며, 지연이 유계를 넘으면 critical 알림한다(SHALL — 브로커측 보호 도입 전의 잔존 리스크로 명시).

#### Scenario: 관측 장기 두절
- **WHEN** 보유 포지션의 가격 관측이 60초 이상 실패하면
- **THEN** critical 알림과 함께 ENTRY_BLOCKED로 자동 강화되고 기준선은 유지된다

#### Scenario: 확정 하한 캡
- **WHEN** RECONCILE 중 전량 청산 발의가 확정 하한 30주로 캡되면
- **THEN** 30주가 제출되고 잔여는 pending으로 남아 해소 후 재발의되며 알림이 발송된다

### Requirement: 비용은 청산 게이트가 아니다
청산·부분익절 발의는 예상 비용을 이유로 차단되지 않는다(SHALL NOT — §0.3, StockOS SELL_COST_BUFFER 미이식). 실질 본전 계산에 쓰이는 비용률은 검증 게이트(비수치·NaN·음수·상한 초과 거부, 상한 `MAX_RATE=0.05` 이식)를 통과한 값만 허용된다(SHALL — 상한 없는 과대 추정은 본전 기준선을 무한히 끌어올려 승자 포지션을 즉시 청산시킨다).

#### Scenario: 비용률 상한 초과 설정
- **WHEN** 상한을 넘는 비용률이 설정되면
- **THEN** 설정이 거부된다 (본전 기준선의 폭주 방지)
