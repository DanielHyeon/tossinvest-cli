# exit-policy Specification (delta)

## MODIFIED Requirements

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
