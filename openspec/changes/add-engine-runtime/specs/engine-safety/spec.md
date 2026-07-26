# engine-safety Specification (delta)

## ADDED Requirements

### Requirement: 엔진 런타임 수명주기
엔진 런타임(`tossctl engine run`)은 기동 인터록 검증을 통과한 뒤에만 루프를 시작해야 한다(SHALL — "자동화 게이트 기동 인터록" 요구사항의 소비자: 게이트 OFF 또는 조항 미충족이면 미충족 항목을 열거하며 기동을 거부한다, fail-closed). verify runlock이 신선한 동안은 기동을 거부한다(SHALL — 검증 절차와 계좌·rate 예산을 다투지 않는다). 기동된 루프 집합(reconcile driver·exit observer)에서 어느 하나라도 오류로 종료하면 전체 런타임이 정지하고 critical 알림이 발송된다(SHALL — 부분 생존 금지: 포지션 보유 중 exit 관측만 죽는 상태는 보호 결손이다). 종료 시그널은 루프 취소·완주 대기·journal 정합 close로 처리하고(SHALL), 두 번째 시그널은 즉시 종료한다. 재기동 복구는 landed 계약(pending 복원·편입 완결·nonce 재사용 금지)을 소비하며 새 복구 경로를 만들지 않는다(SHALL NOT).

#### Scenario: 인터록 미충족 기동
- **WHEN** ProtectionReady 미충족(또는 게이트 OFF) 상태로 `engine run`을 실행하면
- **THEN** 루프가 하나도 시작되지 않고 미충족 항목이 열거되며 종료 코드는 실패다

#### Scenario: 루프 하나의 오류 종료
- **WHEN** 기동된 루프 중 하나가 오류로 종료하면
- **THEN** 나머지 루프도 정지하고 critical 알림이 발송되며 프로세스가 실패로 종료한다

#### Scenario: 검증 실행 중 기동 시도
- **WHEN** verify runlock이 신선한 상태에서 `engine run`을 실행하면
- **THEN** 기동이 거부되고 검증 종료 후 재시도가 안내된다

#### Scenario: 시그널 종료
- **WHEN** 실행 중 런타임에 SIGTERM이 도착하면
- **THEN** 루프가 취소·완주되고 journal이 정합하게 닫힌 뒤 정상 종료한다
