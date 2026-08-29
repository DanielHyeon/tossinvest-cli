## ADDED Requirements

### Requirement: 최적화 화면은 공통 청산 정책을 설명한다
콘솔은 `/optimization`에서 현재 승인값, 권장값, 각 등록 정책의 목표·보호선·부분익절·최종 runner 의미를 한국어로 표시해야 한다 (SHALL).

#### Scenario: 미승인 화면
- **WHEN** 공통 정책 설정이 비어 있는 상태에서 인증된 운영자가 `/optimization`을 연다
- **THEN** 현재 동작은 기존 RATCHET이고 HYBRID_50은 권장일 뿐 아직 적용되지 않았다고 표시한다

#### Scenario: 외부 구매 적용 설명
- **WHEN** 운영자가 정책 카드를 본다
- **THEN** 신규 자체 포지션과 향후 편입 외부 매수분에 적용되고 기존 활성 포지션은 바뀌지 않는다고 표시한다

### Requirement: 정책 변경은 세션과 CSRF를 통과한 사람의 POST다
최적화 정책 저장 route는 기존 `session0(mutating(...))` 체인을 사용하고 GET, 무세션, 잘못된 CSRF 요청을 거부해야 한다 (SHALL).

#### Scenario: 정상 저장
- **WHEN** 유효 세션과 CSRF를 가진 운영자가 등록 policy ID를 POST한다
- **THEN** config seam을 정확히 한 번 호출하고 `/optimization`으로 결과 안내와 함께 redirect한다

#### Scenario: CSRF 누락
- **WHEN** 세션은 있지만 CSRF가 없는 정책 저장 요청이 도착한다
- **THEN** 403을 반환하고 config, audit, broker state를 변경하지 않는다

### Requirement: 최적화 화면은 최소 권한 seam만 가진다
최적화 handler가 받는 설정 seam은 공통 exit policy의 load/save만 제공해야 하며 주문, gate, trading toggle, journal mutation capability를 제공해서는 안 된다 (MUST NOT).

#### Scenario: 정적 capability 검사
- **WHEN** console Options와 최적화 handler의 dependency closure를 검사한다
- **THEN** exit-policy 설정 이외의 mutation capability와 account mutation verb가 존재하지 않는다
