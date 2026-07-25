# engine-safety Specification (delta)

## ADDED Requirements

### Requirement: 엔진 배선은 official-only
자동매매 엔진용 앱 배선(`internal/app` 엔진 프로필)은 주문 mutation 브로커로 공식 Open API 구현만 주입해야 하며(SHALL), WTS 쓰기 경로(hybrid 라우팅 포함)는 엔진 배선에서 도달 불가능해야 한다(SHALL). 이는 테스트로 증명되어야 한다(SHALL). WTS 전용 의도(fractional KRW 등)는 엔진 경로에서 거부된다.

#### Scenario: 엔진 프로필에서 WTS 전용 주문 시도
- **WHEN** 엔진 배선으로 fractional KRW 매수 intent가 제출되면
- **THEN** WTS로 라우팅되지 않고 미지원 오류로 거부된다

### Requirement: 자동화 게이트 기동 인터록
엔진의 자동 주문 게이트는 기본 OFF이며(SHALL), 위험 엔진(Guardian — Phase 2)이 0이 아닌 한도로 주입·활성화되지 않은 상태에서 게이트 활성화가 설정되어 있으면 엔진은 기동을 거부해야 한다(SHALL). 게이트 활성화 flip은 사람 승인 절차를 거친다(운영 토글 규칙 §0.7).

#### Scenario: Guardian 없는 자동 주문 기동
- **WHEN** 자동 주문 게이트 ON + Guardian 미주입 상태로 데몬을 기동하면
- **THEN** 부팅이 명시적 오류로 실패한다

### Requirement: 비상 청산 명령
`tossctl`은 typed-confirmation을 요구하는 수동 전용 flatten-all 명령을 제공해야 한다(SHALL): 미체결 전량 취소 후 보유 포지션 전량 청산 주문 제출, 진행·실패 내역 즉시 출력. 이 명령은 자동 경로에서 호출되지 않는다(SHALL NOT).

#### Scenario: flatten-all 실행
- **WHEN** 운영자가 flatten-all을 확인 문자열과 함께 실행하면
- **THEN** 미체결 취소 → 청산 주문 제출 순으로 진행되고 각 결과가 출력되며, 실패 항목은 재시도 안내와 함께 명시된다

### Requirement: 관측성과 알림
엔진과 주문 경로는 구조화 로그(주문 생명주기 전이, reconcile 결과, 오류 분류)와 핵심 메트릭을 남겨야 하며(SHALL), 다음 이벤트는 push 알림 채널로 통지되어야 한다(SHALL): kill switch·운영 모드 전환, reconciliation 불일치, 주문 거부·IN_DOUBT 발생, 세션·자격증명 만료 임박, 데몬 비정상 종료.

#### Scenario: IN_DOUBT 발생 통지
- **WHEN** 주문 제출이 IN_DOUBT로 표기되면
- **THEN** 구조화 로그와 push 알림이 모두 발생한다

### Requirement: 공식 API capability 실증
자동매매 활성화 전에 공식 API의 무인 운영 전제를 실증 기록해야 한다(SHALL): 자격증명 무인 갱신 연속 N일(기본 3일) soak, rate limit 실측치, 주문·체결·잔고 조회 완전성 확인. 결과는 docs/에 기록되고 미충족 항목은 해당 기능 범위를 축소한다.

#### Scenario: soak 중 자격증명 갱신 실패
- **WHEN** soak 기간 중 토큰 무인 갱신이 실패하면
- **THEN** 실패 원인이 기록되고 해소 전까지 자동매매 활성화 조건이 미충족으로 유지된다
