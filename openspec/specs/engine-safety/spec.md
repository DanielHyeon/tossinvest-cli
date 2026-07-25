# engine-safety Specification

## Purpose
엔진 배선 안전(official-only)·ExecutionGateway·기동 인터록·flatten saga·알림·audit 요구사항.

## Requirements

### Requirement: 엔진 배선의 구조적 official-only
엔진 프로필(`internal/app` engine wiring)은 공식 Open API 브로커를 **직접** 구성해야 하며(SHALL), hybrid 클라이언트·WTS mutator 타입은 엔진 의존 그래프에 존재해서는 안 된다(SHALL NOT — 정적 import 테스트로 검증). 엔진 프로필은 사용자 config의 `OpenAPI.Enabled/Prefer`를 무시하고, 유효한 공식 자격증명이 없으면 기동을 거부한다(SHALL). place/cancel/amend/조건주문 전 mutation matrix에서 WTS 호출 0회를 spy 테스트로 증명한다(SHALL).

#### Scenario: 자격증명 누락 기동
- **WHEN** 공식 자격증명 없이 엔진 프로필을 기동하면
- **THEN** WTS 폴백 없이 기동이 명시적으로 거부된다

#### Scenario: 전 mutation matrix WTS 미도달
- **WHEN** 엔진 배선으로 place/cancel/amend/조건주문을 각각 실행하면
- **THEN** WTS spy 호출 횟수는 모두 0이다

### Requirement: 엔진 브로커의 cancel/amend 사전 확인
공식 API에는 `GetOrderAvailableActions` 대응이 없으므로, 엔진 브로커 어댑터는 cancel/amend 사전 확인을 `OrderByID` 상태 파생으로 구현하거나 사전 확인을 브로커 선택적으로 만들어야 한다(SHALL). WTS 세션이 없거나 만료된 상태에서도 엔진의 cancel/amend는 동작해야 한다(SHALL — 테스트 필수).

#### Scenario: WTS 세션 만료 중 취소
- **WHEN** WTS 세션이 만료된 상태에서 엔진이 미체결 주문을 취소하면
- **THEN** 공식 API 경로만으로 취소가 완료된다

### Requirement: ExecutionGateway 봉인
엔진의 모든 주문 mutation은 단일 ExecutionGateway를 통해야 하며(SHALL), Gateway는 GuardianDecision(intent 해시, 한도 스냅샷, 만료 시각, one-shot nonce)을 요구하고 브로커 호출 직전 재검증한다(SHALL). Guardian 결정 없는 제출 경로는 컴파일·API 수준에서 존재하지 않아야 한다(SHALL NOT: raw mutator 미노출). Phase 1에서 Guardian 인터페이스는 초안이며 활성화·구현은 Phase 2에서 확정한다. 기존 CLI/MCP 표면은 upstream의 confirm token 게이트를 유지하며 이 change의 대상이 아니다 — MCP 표면이 Gateway를 우회한다는 잔존 리스크는 Phase 4(단일 writer 데몬)까지 문서화된 상태로 유지된다.

#### Scenario: Guardian 결정 없는 제출 시도
- **WHEN** GuardianDecision 없이 Gateway 제출을 시도하면
- **THEN** 컴파일 오류 또는 즉시 거부된다

#### Scenario: nonce 재사용
- **WHEN** 이미 사용된 GuardianDecision으로 재제출을 시도하면
- **THEN** one-shot nonce 검증 실패로 거부된다

### Requirement: 자동화 게이트 기동 인터록
자동 주문 게이트는 기본 OFF이며(SHALL), 게이트 ON 설정 시 다음이 모두 검증되지 않으면 기동을 거부한다(SHALL): (1) Guardian이 0이 아닌 한도로 주입, (2) 유효한 capability attestation(만료 시각·계좌 식별·성공 endpoint 집합을 담은 로컬 durable 기록 — verify-execution-capability change가 생성) 존재·미만료·계좌 일치. 게이트 flip은 사람 승인 절차(§0.7)와 audit 기록을 요구한다(SHALL).

#### Scenario: attestation 만료 상태 기동
- **WHEN** 게이트 ON + attestation 만료 상태로 기동하면
- **THEN** 기동이 거부되고 재검증 안내가 출력된다

### Requirement: Flatten Saga
`tossctl` flatten-all은 durable saga로 구현되어야 한다(SHALL): (1) 신규 진입 차단 → (2) 미체결 각각 취소 + 결과 확정(IN_DOUBT 규칙 적용) → (3) 계좌 재조회 안정화 → (4) 최신 매도가능수량 기준 reduce-only 청산 주문(비-fractional은 공격적 limit) → (5) 반복 reconcile로 잔여 확인. `--dry-run` 모드는 제출할 주문 목록을 mutation 0건으로 출력해야 한다(SHALL). 확인 문자열은 마스킹된 계좌 식별·포지션 수·예상 청산 수량·만료 nonce를 포함하고, TTY 직접 입력만 허용하며 자동화 플래그는 금지된다(SHALL NOT). 크래시 후 재실행 시 saga는 journal 기록에서 안전하게 재개된다(SHALL).

#### Scenario: 취소 결과 불명 중 청산 단계 진입 시도
- **WHEN** 미체결 취소가 IN_DOUBT 상태인데 청산 단계로 진행하려 하면
- **THEN** 해당 심볼 청산은 취소 확정 시까지 보류되고 oversell이 방지된다

#### Scenario: dry-run
- **WHEN** flatten-all --dry-run을 실행하면
- **THEN** 취소·청산 대상 목록이 출력되고 어떤 mutation도 발생하지 않는다

### Requirement: 등급화된 알림
알림은 등급화되어야 한다(SHALL): critical(IN_DOUBT·UNRESOLVED_IN_DOUBT 발생, 자격증명 만료 임박, 영구 불일치, UNKNOWN_BROKER_STATE)은 로컬 durable outbox에 기록 후 전송하고, 전달 실패가 지속되면 신규 진입을 차단한다(SHALL). 일반(체결·상태 전이)은 best-effort. 죽은 프로세스는 스스로 통지할 수 없으므로 heartbeat(예상 주기 초과 시 ntfy 측에서 경보) 방식을 사용한다(SHALL). 알림 이벤트 타입은 확장 가능한 enum으로 정의하고 Phase 2 이벤트(kill switch·운영 모드)는 예약만 한다.

#### Scenario: critical 알림 전달 실패 지속
- **WHEN** critical 이벤트의 전송이 재시도 한도까지 실패하면
- **THEN** 신규 진입이 차단되고 outbox에 미전달 상태로 보존되며, 전달 복구 후 수동 확인으로 해제한다

### Requirement: 테스트·도구의 실 endpoint 기계적 차단
테스트 바이너리와 검증 도구는 실 Toss hostname에 대한 mutation 요청이 구조적으로 불가능해야 한다(SHALL): 테스트는 격리 config 디렉터리 + httptest transport만 사용하고, 실 hostname으로의 POST 시도를 hard fail시키는 transport 가드 테스트를 완료 게이트에 포함한다(SHALL).

#### Scenario: 테스트에서 실 hostname POST 시도
- **WHEN** 테스트 중 실 Toss hostname으로 mutation 요청이 구성되면
- **THEN** transport 가드가 즉시 실패시키고 테스트가 실패한다

### Requirement: 운영 설정 audit
게이트 토글·한도 변경 등 운영 설정 변경은 변경 전후 값·시각·주체를 audit 로그로 기록해야 한다(SHALL).

#### Scenario: 게이트 토글 변경
- **WHEN** 자동화 게이트 설정이 변경되면
- **THEN** audit 로그에 이전 값·새 값·시각이 기록된다
