# fill-detection Specification (delta)

## ADDED Requirements

### Requirement: 폴링이 체결 감지의 권위
체결·미체결·잔고 상태의 권위 소스는 공식 Open API 주기 폴링이어야 하며(SHALL), 폴링 주기는 최대 신선도 SLO(기본: 장중 5초)로 정의된다. WTS 세션이 만료되어도 체결 감지는 중단 없이 동작해야 한다(SHALL).

#### Scenario: WTS 세션 만료 중 체결 발생
- **WHEN** WTS 세션이 만료된 상태에서 주문이 체결되면
- **THEN** 공식 API 폴링이 SLO 이내에 체결을 감지하고 상태를 갱신한다

### Requirement: SSE는 지연 단축 힌트
WTS SSE 이벤트는 즉시 재조회를 촉발하는 힌트로만 사용되어야 하며(SHALL), 상태 변경의 근거로 이벤트 페이로드를 직접 사용해서는 안 된다(SHALL NOT). 이벤트 기반 재조회는 토픽별 coalescing(single-flight)과 최소 간격 제한을 적용해 자기 유발 rate limit을 방지해야 한다(SHALL).

#### Scenario: 이벤트 폭주
- **WHEN** 같은 토픽의 SSE 이벤트가 짧은 시간에 다수 도착하면
- **THEN** 재조회는 진행 중 1건으로 합쳐지고 최소 간격이 보장된다

#### Scenario: SSE 단절
- **WHEN** SSE 연결이 장시간 끊겨도
- **THEN** 폴링 권위에 의해 상태 갱신은 SLO 이내로 유지된다

### Requirement: 상태 반영의 멱등성
동일 체결·주문 상태의 중복 감지(폴링과 SSE 재조회의 중복, 재시작 후 재조회)는 상태와 원장 기록을 중복 생성해서는 안 된다(SHALL NOT). 반영은 주문번호·체결 식별 기준으로 멱등해야 한다(SHALL).

#### Scenario: 동일 체결 중복 수신
- **WHEN** 같은 체결이 폴링과 SSE 재조회에서 각각 감지되면
- **THEN** 상태·기록은 한 번만 반영된다
