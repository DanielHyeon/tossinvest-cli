# exit-policy Specification

## Purpose
손익 극대화 exit 정책 — t0 기준선(진입 손절)·baseline ratchet(워터마크 단조 상승)·profit ladder·발의 pending 수명주기·관측 fail-safe·비용 비게이트 요구사항.
## Requirements
### Requirement: t0 기준선 — 진입 손절가가 첫 독자를 갖는다

포지션이 열리는 순간 보호 기준선은 그 포지션을 정당화하는 기록의 손절가로 초기화되어야 한다(SHALL — `exit_states.baseline_price`는 NOT NULL이며 빈 기본값이 없다): 엔진 진입 포지션은 진입 결정의 손절가(RiskIntent.stop), 편입 포지션은 편입 기록의 합성 손절가("외부 취득 포지션의 자동 편입" 요구사항)다. exit_state 개설 경로는 두 출처를 모두 수용해야 한다(SHALL — 편입 포지션은 진입 결정 조회 자체가 성립하지 않으므로 자격 출처별 분기가 필요하다). 기준선 하회 발의 규칙은 t0부터 유효하다(SHALL). entry 결정도 편입 기록도 없는 포지션은 exit 정책의 대상이 아니며(SHALL — R·기준선의 근거가 없다), **그 발견은 `adoption.enabled` 값과 무관하게 운영자 알림을 발송한다**(SHALL — 엔진이 보호하지 않을 포지션 옆에서 거래 중임을 운영자가 알아야 한다; 무알림은 전이 상태에 한한다 — 편입 요구사항).

#### Scenario: 개시 직후 손절 하회
- **WHEN** 진입 체결 직후(+0.4R 도달 전) 관측 가격이 진입 손절가를 하회하면
- **THEN** 전량 청산이 RISK_REDUCING 의도로 발의된다

#### Scenario: 편입 후 외부 포지션
- **WHEN** 편입 기록이 영속된 외부 포지션의 관측 가격이 합성 손절가를 하회하면
- **THEN** 엔진 진입 포지션과 동일하게 전량 청산이 RISK_REDUCING 의도로 발의된다

#### Scenario: 어떤 기록도 없는 포지션
- **WHEN** entry 결정도 편입 기록도 없는 포지션이 전이 상태 밖에서 확인되면
- **THEN** exit 판정은 수행되지 않고 운영자 알림이 발송된다

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

### Requirement: 외부 취득 포지션의 자동 편입

`adoption.enabled`가 참인 엔진은 기록이 정당화하지 않는 보유(외부·수동 취득)를 발견하면 exit 관리에 편입해야 한다(SHALL — 사용자 결정 2026-07-26). `adoption.enabled`의 기본값은 false이며(SHALL — §0.2 zero-value 안전; false에서의 동작은 무관리 보유 알림을 포함한 기존 동작과 동일하다), true 전환은 audit 기록과 사람의 직접 승인을 요한다(SHALL — §0.5·§0.7; 루프백 콘솔의 세션+CSRF 게이트를 통과한 사람의 저장은 직접 승인이고, audit은 저장 시점 엔트리(operator-console 스펙 소유)와 엔진 기동 시의 설정 변경 기록의 이중이다). 편입은 영속이 선행한다(SHALL): `position_adoptions` 기록(심볼·시장·수량·원가와 그 출처 표기·편입 시점 관측가·합성 손절가·관측 시각·digest — adopt-external-positions design A1이 정본)이 journal에 영속되고 `positions.adoption_id`가 이를 가리킨 뒤에만 exit_state가 열린다.

**종목별 편입(SHALL — 사용자 결정 2026-07-27)**: `adoption.include_symbols`의 심볼은 `enabled`가 false여도 편입 후보다. 후보 판정은 `(enabled ∨ include(심볼)) ∧ ¬exclude(심볼)`이며 `exclude_symbols`가 항상 우선한다(SHALL — include·exclude 동시 등재는 편입하지 않고 무관리 알림이 남는다). include 경유 편입은 신선 조건·staleness·manage-forward·알림 규칙 전부에서 enabled 경유 편입과 동일하다(SHALL — 별도 축소 경로 금지). 빈 include 목록은 이 요구 이전 동작과 동일하다(SHALL — §0.2). include 목록은 정규화되며(SHALL — 대문자·공백 제거·중복 제거) **시장 무관 심볼 단위**다(SHALL — exclude의 기존 의미론 상속; 시장 한정 목록은 별도 스펙 변경 사안). 지정은 상시 규칙이다(SHALL 명시 — CLOSED 후 재매수 인스턴스도 재편입되며, 목록에서의 제거는 장래 편입에만 영향을 주고 이미 편입된 포지션에 영향을 주지 않는다 — 편입 해제 부재는 design A5 유지). `include_symbols`의 변경은 §0.5 audit 대상이다(SHALL — 엔진 기동 설정 기록(recordGateSettings)에 항목으로 포함된다; 특정 실보유를 매도 가능 관리 대상으로 만드는 설정이 무audit이어서는 안 된다).

**manage-forward t0**(SHALL — design A2): EntryPrice는 **편입 판정에 쓴 관측가**이고 InitialStop은 `관측가 × (1 − adoption.default_stop_pct)`다(워터마크는 개설 규칙상 entry로 자동 seed된다). 관측가는 **편입 트랜잭션 직전의 시세 관측**이어야 하며 staleness 15초를 초과하면 편입을 연기한다(SHALL — 묵은 가격으로 동결된 합성 손절은 즉발 청산 경로가 된다). 소스는 exit 관측과 동일한 시세 경로이며 `[기존 제약 — 엔진 가격 경로 float64]`. 원가는 브로커 원문 decimal 문자열로 보존하되(SHALL — cost_basis에 한정) R 산식(분자·분모 모두)에 쓰지 않는다(SHALL NOT — 기록·표시용). **편입 행위 자체는 매도 발의를 생성하지 않는다**(SHALL NOT — 편입 트랜잭션에 exit 판정이 포함되지 않는다). 편입 관측과 첫 exit 관측 사이의 가격 이동으로 인한 첫 틱 발의는 정상 exit 동작이다(SHALL 명시). `adoption.default_stop_pct`는 `0.02 ≤ pct < 1` 범위를 벗어나면 설정 거부다(SHALL — 하한 근거: 관측 노이즈·왕복 비용 규모보다 작은 보호폭은 즉발 청산 장치가 된다, provenance 기록). 범위 검증은 블록이 의미 있을 때 — `enabled`가 참이거나 include 목록이 비어 있지 않을 때 — 요구된다(SHALL — include만으로 편입되는 심볼도 같은 합성 손절 분모를 쓴다; 거부된 블록은 전면 zeroing되어 편입이 전면 비활성으로 남는다). +0.8R 도달 시 기준선이 **편입가 기준** 실질 본전으로 승격된다는 귀결을 운영 문서에 명시한다(SHALL — 편입은 편입일 가격+비용을 보호 바닥으로 만든다).

편입 조건(SHALL): `AutomationStatus.Verified`인 엔진에서, 해당 심볼이 RECONCILE 상태가 아니고, 신선한 보유 확인 — Stabiliser 통과(최소 2초 간격 연속 2회 동일 스냅샷) + 관측 staleness 10초 이내 — 을 만족할 때만. Stabiliser 미수렴 시 편입은 연기되며 이는 fail-closed 의도 동작이다(SHALL 명시). `adoption.exclude_symbols`의 심볼은 편입하지 않는다(SHALL NOT). 편입 이후 래칫·부분익절·관측·pending·staleness 규칙은 엔진 진입 포지션과 동일 규칙이 적용된다(SHALL — 별도 축소 경로 금지; ladder는 정책 배정 경로가 생기면 동일 적용). CLOSED 후 재매수로 생긴 새 인스턴스의 재편입은 의도된 동작이다(SHALL 명시 — include 심볼의 재매수도 동일하다).

알림(SHALL — design A4): 무관리 보유 발견 알림은 enabled와 무관하게 유지된다. enabled=true 또는 include 경유의 편입 성공은 편입가·합성 손절을 담은 이벤트로 기록·통지되어 그 알림을 대체하고, 제외 목록 심볼·편입 실패는 알림이 남는다. 무관리 알림의 사유 문구는 실제 상태를 말해야 한다(SHALL — 사유 행렬: 설정 거부/의도적 제외(enabled·include 무관)/enabled 시도 실패/include 지정 시도 실패/꺼져 있고 미지정): 특히 include 심볼의 편입 실패 알림은 시도와 실패를 말하고 "꺼져 있다"고 말해서는 안 된다(SHALL NOT). 전이 상태(RECONCILE 대기·인터록 미검증·Stabiliser 미수렴)만 무알림이다. 편입은 매수·사이징을 변경하지 않는다(SHALL NOT).

#### Scenario: 수동 매수 종목의 자동 편입
- **WHEN** enabled=true 엔진에서 수동 매수 심볼이 신선 조건을 충족해 편입되면
- **THEN** exit_state가 관측가 기준 t0(R=0)로 열리고, 편입 트랜잭션은 매도 발의를 생성하지 않으며, 편입 이벤트가 통지된다

#### Scenario: 종목별 지정 편입
- **WHEN** enabled=false 엔진에서 include_symbols에 있는 심볼의 무결정 보유가 신선 조건을 충족하면
- **THEN** 그 심볼만 편입되고, 목록에 없는 무결정 보유는 편입 없이 무관리 알림이 유지된다

#### Scenario: include와 exclude 동시 등재
- **WHEN** 같은 심볼이 include_symbols와 exclude_symbols에 모두 있으면
- **THEN** 편입되지 않고 무관리 보유 알림이 발송된다 (exclude 우선)

#### Scenario: 편입 관측과 첫 exit 관측 사이의 가격 이동
- **WHEN** 편입 시 관측가와 다른 가격이 첫 exit 관측에 나타나면
- **THEN** 래칫 규칙이 그 가격에 정상 적용된다(기준선 하회면 청산 발의 — 이는 편입의 결함이 아니라 exit 동작이다)

#### Scenario: enabled=false의 무관리 보유
- **WHEN** `adoption.enabled`가 false이고 include_symbols가 비어 있는 엔진이 무관리 보유를 발견하면
- **THEN** 편입 없이 운영자 알림이 발송된다(기존 동작 유지)

#### Scenario: RECONCILE 중 편입 시도
- **WHEN** 심볼이 RECONCILE 상태일 때 무결정 보유가 관측되면
- **THEN** 편입은 일어나지 않고 해소 후 다음 관측에서 재평가된다(전이 상태 — 무알림)

#### Scenario: 제외 목록 심볼
- **WHEN** `adoption.exclude_symbols`에 있는 심볼의 무결정 보유가 관측되면
- **THEN** 편입 없이 운영자 알림만 발송된다

#### Scenario: 편입 기록 영속 후 크래시
- **WHEN** position_adoptions 영속 후 exit_state open 전에 크래시·재시작되면
- **THEN** 재시작 복구가 영속된 기록으로부터 exit_state open을 완결하고 중복 편입은 발생하지 않는다

#### Scenario: 범위 밖 손절폭 설정
- **WHEN** `adoption.default_stop_pct`가 0.02 미만 또는 1 이상으로 설정되면(enabled 또는 비어 있지 않은 include와 함께)
- **THEN** 설정이 거부되고 편입은 전면 비활성으로 남는다

### Requirement: exit 기준선은 하나의 권위 snapshot으로 계산된다
시스템은 entry, initial stop, current protection, high-water, active rung, next target, next protection, proposal action·ratio·projected quantity를 하나의 immutable decimal snapshot으로 계산해야 한다 (SHALL). 실행과 화면은 이 결과를 별도로 재계산해서는 안 된다 (MUST NOT).
snapshot은 immutable policy ID/version/digest, position generation, observation identity, deterministic snapshot ID와 decision ID를 포함해야 한다 (SHALL).

#### Scenario: 다음 익절과 보호선 계산
- **WHEN** 관리 포지션이 현재 rung과 high-water를 가진 채 평가된다
- **THEN** snapshot은 현재 보호선과 다음 도달 목표·다음 보호선을 함께 반환한다

#### Scenario: 단조 보호선
- **WHEN** 새 후보 보호선이 저장된 보호선보다 낮다
- **THEN** snapshot은 기존 보호선을 유지하고 낮은 값을 적용하지 않는다

### Requirement: 1주는 중간 익절 없이 끝까지 보호한다
whole-share 보유 수량이 정확히 1주일 때 시스템은 중간 부분익절 주문을 생성해서는 안 되며 (MUST NOT), rung과 보호선 승격은 계속 적용해야 한다 (SHALL). 최종 전량익절 또는 보호선 breach는 정확히 1주 전량 청산을 제안해야 한다 (SHALL).

#### Scenario: 1주 중간 목표 도달
- **WHEN** 1주 포지션이 partial ratio 0.25인 중간 rung에 도달한다
- **THEN** 주문 proposal은 없고 active rung과 보호선만 상승하며 수량은 1주로 남는다

#### Scenario: 1주 최종 목표 도달
- **WHEN** 같은 포지션이 final take-full rung에 도달한다
- **THEN** 정확히 1주 전량익절 proposal을 생성한다

#### Scenario: 1주 보호선 이탈
- **WHEN** 같은 포지션의 현재가가 active protection 아래로 내려간다
- **THEN** 중간 익절 이력과 무관하게 정확히 1주 전량보호 proposal을 생성한다

### Requirement: 0수량 주문은 존재할 수 없다
부분익절 projected quantity가 최소 주문 단위보다 작으면 시스템은 state-only transition으로 처리하고 0수량 intent·reservation·broker request를 만들어서는 안 된다 (MUST NOT).

#### Scenario: 내림 결과 0주
- **WHEN** 잔량과 partial ratio의 곱을 whole-share로 내림한 결과가 0이다
- **THEN** exit state는 승격될 수 있지만 journal mutation attempt와 broker call은 0건이다

### Requirement: 공통 정책 descriptor는 계산 기본값과 설명을 제공한다
시스템은 `COMMON_LADDER_BALANCED`, `COMMON_LADDER_RUNNER`, `COMMON_LADDER_HYBRID_50`의 label, summary, recommended 여부, rung별 target/stop/partial, runner gap, 단위와 1주 projection을 server-authoritative descriptor로 제공해야 한다 (SHALL). UI가 descriptor 밖의 기본 수치를 발명해서는 안 된다 (MUST NOT).

#### Scenario: 미승인 공통 정책
- **WHEN** 운영자가 공통 정책을 아직 승인하지 않았다
- **THEN** effective 상태는 `기존 RATCHET 유지`, 추천 선택은 `COMMON_LADDER_HYBRID_50`으로 서로 구분된다

#### Scenario: 1주 descriptor preview
- **WHEN** descriptor를 1주 포지션에 적용해 preview한다
- **THEN** 모든 intermediate partial은 `매도 0주 · 보호선 승격`, 최종 take-full과 protection breach는 `1주 전량`으로 설명된다

### Requirement: 설정 descriptor는 transport-neutral이고 유한한 선택만 허용한다
시스템은 UI나 HTTP 타입에 의존하지 않는 field key/type, control kind, finite stable option ID, default/effective state, apply timing, safety direction과 provenance 계약을 제공해야 한다 (SHALL). 동일 policy ID/version의 canonical digest가 달라지면 해당 descriptor와 snapshot을 거부해야 한다 (SHALL).

#### Scenario: policy identity 충돌
- **WHEN** 같은 policy ID/version으로 서로 다른 rung digest가 로드된다
- **THEN** registry와 exit evaluation은 fail-closed하고 기존 identity의 의미를 덮어쓰지 않는다

#### Scenario: 자유 입력 없는 정책 선택
- **WHEN** transport가 공통 정책 descriptor를 렌더링한다
- **THEN** finite stable option ID만 선택할 수 있고 transport가 별도 수치나 기본값을 발명하지 않는다

### Requirement: 복구된 기준선은 낮아질 수 없다
exit recovery는 저장 snapshot과 현재 재계산 후보 중 더 안전한 기준만 채택하고 baseline과 high-water를 낮춰서는 안 된다 (MUST NOT).
선택은 검증된 coherent snapshot 단위로 수행해야 하며 (SHALL), 서로 다른 policy version/rung/target의
field별 최댓값을 조합해 합성 snapshot을 만들어서는 안 된다 (MUST NOT).
protection/high-water 비교는 같은 policy digest 안에서만 허용하고 (SHALL), rung과 next target/protection은 선택된 immutable policy에서 파생해야 한다 (SHALL). policy digest가 다르거나 안전한 후보 하나를 결정할 수 없으면 해당 포지션을 격리해야 한다 (SHALL).

#### Scenario: 재시작 후 낮은 현재가
- **WHEN** 재시작 시 재계산 후보가 저장된 active protection보다 낮다
- **THEN** 저장된 protection을 유지하고 포지션을 더 낮은 손절선으로 재해석하지 않는다

#### Scenario: 손상 snapshot
- **WHEN** 한 포지션의 snapshot이 invalid decimal 또는 unknown policy version을 포함한다
- **THEN** 해당 포지션의 자동 판정은 fail-closed이고 운영 경고를 남기며 다른 포지션의 emergency exit는 계속한다

#### Scenario: 더 높은 값이 서로 다른 후보에 분산됨
- **WHEN** saved 후보는 더 높은 protection을, recomputed 후보는 더 높은 rung을 가지지만 tuple 조합이 정책상 불가능하다
- **THEN** 검증된 후보 하나를 통째로 선택하거나 해당 포지션을 격리하고 field별 max snapshot을 만들지 않는다

### Requirement: 전역 정책 변경은 활성 포지션을 재해석하지 않는다
시스템은 common policy 변경만으로 기존 exit state의 policy ID/version을 변경해서는 안 된다 (MUST NOT).

#### Scenario: 공통값 변경
- **WHEN** 활성 BALANCED 포지션이 있는 동안 공통값을 RUNNER로 저장한다
- **THEN** 기존 포지션은 BALANCED snapshot을 유지하고 새 lifecycle만 RUNNER를 사용한다

### Requirement: 미국 include-only 편입 경로는 실제 engine boundary에서 회귀 검증된다
미국 보유분은 market-agnostic include 규칙과 기존 신선도/대사 조건을 충족하면 한국 보유분과 동일한 fold→adopt→exit t0 경로를 사용해야 한다 (SHALL). adoption quote는 candidate market과 일치하는 currency provenance(KR→KRW, US→USD)를 가져야 하며 (SHALL), currency가 비었거나 다르거나 같은 symbol이 서로 다른 candidate market에 중복되면 편입을 연기해야 한다 (MUST). 다른 symbol/market의 가격을 사용해서는 안 된다 (MUST NOT).

#### Scenario: 미국 include-only 보유분 편입
- **WHEN** Verified engine에서 adoption.enabled=false, include_symbols=[AAPL], stable/fresh US AAPL holding 두 관측과 fresh official AAPL USD quote 200이 주어진다
- **THEN** 다음 정상 RunOnce는 Folded=1, Adopted=1, Unmanaged=0이고 external-adoption provenance, t0 entry/high-water 200, 5% synthetic initial-stop/baseline 190을 하나의 편입 transaction으로 영속한다

#### Scenario: account-wide quantity mismatch가 미국 편입을 보류한다
- **WHEN** 같은 candidate에 account-wide permanent quantity-mismatch block이 active다
- **THEN** RunOnce는 adoption transaction과 exit state를 만들지 않고 read projector는 미국 미지원이 아닌 `RECONCILE_BLOCKED`로 설명한다

#### Scenario: 미국 candidate에 KRW 또는 무통화 quote가 온다
- **WHEN** US AAPL candidate의 quote currency가 KRW이거나 비어 있다
- **THEN** 편입과 exit t0 생성은 연기되고 잘못된 가격 provenance는 journal에 영속되지 않는다

#### Scenario: 같은 symbol이 KR/US candidate에 동시에 있다
- **WHEN** 동일 symbol이 KR과 US의 미편입 candidate로 동시에 존재하고 quote transport가 symbol-only다
- **THEN** market identity가 모호하므로 두 candidate 모두 편입되지 않는다

### Requirement: Automatic adoption obeys every authoritative RECONCILE state

Before pricing or adopting a holding, the engine SHALL use active journal RECONCILE states for the account as the authority and SHALL block every candidate covered by an account- or symbol-scoped state, regardless of which producer created the state. The runtime management projection SHALL expose the same covering states. Failure to read or update that authority MUST stop the cycle before adoption.

#### Scenario: Non-quantity reconcile cause covers a candidate
- **WHEN** an included or globally enabled holding is covered by `SNAPSHOT_UNAVAILABLE`, `SNAPSHOT_STALE`, `IDENTIFIER_CONFLICT`, or `ATTRIBUTION_FAILED`
- **THEN** the engine performs no adoption and no price read, while `/positions` reports the authoritative reconcile block rather than ordinary adoption pending

#### Scenario: Tracker persistence fails before adoption
- **WHEN** the reconciliation comparison cannot durably enter or release its block state
- **THEN** the cycle returns an error before candidate pricing and the holding remains unadopted
