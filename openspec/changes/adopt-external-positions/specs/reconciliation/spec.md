# reconciliation Specification (delta)

## MODIFIED Requirements

### Requirement: Reconciliation 계약
로컬 상태와 토스 계좌의 대사는 명시된 계약을 따라야 한다(SHALL): 스냅샷은 (미체결 목록 pagination 완주 → 보유 → 잔고) 고정 순서로 구성하고 as-of 시각을 기록하며, 부분 실패한 스냅샷은 폐기한다(SHALL). 비교 키는 계좌·심볼·lineage 해소 후 현재 주문번호. 수량 허용 오차 0, 평균단가는 decimal 문자열 비교 + 문서화된 epsilon(진입 차단 판정에서 제외). 안정화는 최소 간격(기본 2초)을 둔 연속 2회 동일 스냅샷으로 판정한다(SHALL). 로컬 intent와 매칭되지 않는 브로커 주문·포지션은 external provenance로 분류한다. 충돌 시 토스 계좌가 항상 우선한다(SHALL).

로컬 포지션 상태의 출처는 **Position 투영**(체결 이벤트+조정 이벤트 — position-ledger)이며(SHALL), fills-only 파생과 별도의 두 번째 포지션 계산을 두지 않는다(SHALL NOT). 비교는 심볼 수준에서 수행하고 투영은 비-CLOSED 인스턴스의 합으로 축약한다(SHALL — 보유 스냅샷의 market 차원 제공 여부는 `[미측정]`이며, 제공 전까지 심볼 합산이 비교 단위다). external 분류된 브로커 보유는 조정 이벤트로 투영에 편입하되(SHALL — 청산 수량 판정이 실제 보유를 알아야 한다), exit 관리 자격 없는 포지션의 발견은 `adoption.enabled`와 무관하게 알림을 발송하고(SHALL — 전이 상태 제외; exit-policy 편입 계약이 알림·편입 규칙의 정본), enabled=true이면 편입 후보 판정으로 이어진다(SHALL).

**대사 구동 루프**(SHALL — adopt-external-positions design A6): 엔진은 주기 60초의 구동 루프로 전체 스냅샷 수집(미체결 pagination ≤ MaxPages 50 + holdings 1콜 + 통화별 잔고 N콜) → Stabiliser(수집 2회 필요) → 비교·fold → `Tracker.Observe`(차단·해제 상태기계 — 확정 하한 캡의 전제) → 편입 후보 판정을 수행한다. 이 루프가 "재대사 최소 간격 30초" 요구사항의 주기적 재대사 절차이며, §0.4 계상은 정상 상태 주기당 수집 2회 × (1 + 1 + 통화 수)콜 + 편입 후보 발생 시 심볼당 시세 1콜(observed_price — staleness 상한은 exit-policy 편입 계약), 상한은 MaxPages다(SHALL 문서화). fold의 "entry 결정 상속 금지" 가드는 `entry_decision_id` 명시 비교로 유지한다(SHALL — 편입 포지션(adoption_id만 설정)에 fold가 착지하는 것은 정상 재대사 경로다).

#### Scenario: 외부 수동 주문 발견
- **WHEN** 로컬 journal에 없는 미체결 주문이 계좌 조회에서 발견되면
- **THEN** external로 분류·기록되고 엔진 상태와 분리 추적되며 알림이 발송된다

#### Scenario: 외부 포지션의 투영 편입
- **WHEN** 로컬 투영이 0인 심볼에 브로커 보유가 발견되면
- **THEN** 외부 분류 조정 이벤트로 투영에 편입되고, 무관리 상태로 확정되면 알림이 발송되며, enabled=true이면 exit-policy 편입 계약의 판정으로 이어진다

#### Scenario: 편입 후 재대사
- **WHEN** adoption_id가 설정된 포지션이 다음 대사 주기에 다시 관측되면
- **THEN** fold 가드(entry 결정 상속 금지)에 걸리지 않고 수량 비교가 정상 수행된다
