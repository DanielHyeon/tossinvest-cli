# engine-safety Specification (delta)

## ADDED Requirements

### Requirement: 엔진 런타임 수명주기
엔진 런타임(`tossctl engine run`)의 루프 집합은 **reconcile driver·exit observer·체결 감지(= `filldetect.Detector` 폴링 루프)**다(SHALL — 체결 감지 없는 런타임은 발의 pending이 영구 미해소로 남아 exit 수명주기 계약을 위반한다; 힌트 라우팅(`Hints`)을 포함하는 경우 Refresh 미배선은 감독이 아니라 **조립 시점 검증**으로 거부한다(SHALL); exit 관측의 SLO 양보 지점은 엔진 배선의 어댑터로 체결 감지 상태에 연결한다). 기동 순서(SHALL): ① 게이트 OFF면 기동할 루프 집합이 없으므로 기동을 거부한다 — 이는 이 change가 정의하는 규칙이다(기동 인터록은 게이트 ON에만 정의된다). ② 게이트 ON이면 기동 인터록 검증을 소비하고, 미충족이면 인터록이 반환한 미충족 항목을 열거하며 거부한다(fail-closed). ③ verify runlock이 신선하면 거부한다. ④ **실행 중 엔진 인스턴스가 있으면 두 번째 기동을 거부한다**(SHALL — journal은 단일 writer 설계다). 배타의 기제는 **journal 디렉터리의 flock**이다(SHALL — 동시 기동 경합까지 닫는다; 자문 마커로는 경합이 닫히지 않는다). 이와 별도로 **엔진 활성 마커**(갱신 1분·stale 5분 — runlock 선례 수치)를 유지해 콘솔의 엔진 상태 표시·autostart의 사전 확인이 소비한다(SHALL — 자문 신호이며 배타는 flock이 담당함을 명시). verify 측이 이 마커를 검사해 엔진 실행 중 verify를 거부하는 것은 execution-verification change(2b)의 후속 태스크다 — 이 change는 엔진 측 검사(verify runlock 신선 시 기동 거부)만 소유한다.

감독 계약은 두 층이다(SHALL): ① **방어적 종료 계약** — 루프가 컨텍스트 취소 외의 사유로 반환하면 전체 런타임이 정지하고 critical 알림이 발송된다(현행 루프들은 그런 반환을 하지 않으므로 이는 방어선이다). 컨텍스트 취소에 의한 반환은 정상 종료이며 critical을 발송하지 않는다(SHALL NOT). ② **지속 열화 임계** — 루프가 살아 있으나 사이클이 연속 실패하는 상태를 각 루프에 정의한다: exit 관측은 landed 60초 두절 계약 유지, reconcile driver와 체결 감지는 연속 5주기 실패 시 critical 알림 + ENTRY_BLOCKED 자동 강화(SHALL — 자동 강화 트리거 열거는 risk-management delta가 확장한다; 루프는 계속 재시도한다 — landed "실패한 사이클은 다음 주기에 재시도" 결정과 양립).

종료 시그널은 루프 취소·완주 대기·journal 정합 close로 처리하고(SHALL), 두 번째 시그널은 즉시 종료한다. 재기동 복구는 landed 계약(pending 복원·편입 완결·nonce 재사용 금지)을 소비하며 새 복구 경로를 만들지 않는다(SHALL NOT).

#### Scenario: 게이트 OFF 기동
- **WHEN** 게이트 OFF 상태로 `engine run`을 실행하면
- **THEN** "기동할 루프 집합이 없다(게이트 OFF)"로 거부되고 실패 종료한다 — 인터록 조항 열거는 없다

#### Scenario: 게이트 ON + 인터록 미충족 기동
- **WHEN** 게이트 ON + ProtectionReady 미충족 상태로 `engine run`을 실행하면
- **THEN** 루프가 하나도 시작되지 않고 인터록의 미충족 항목이 열거되며 실패 종료한다

#### Scenario: 두 번째 인스턴스 기동
- **WHEN** 엔진이 실행 중인 머신에서 `engine run`(또는 autostart·콘솔 버튼)이 다시 실행되면
- **THEN** 실행 중 인스턴스가 안내되고 기동이 거부된다

#### Scenario: 루프의 비정상 반환
- **WHEN** 기동된 루프 중 하나가 컨텍스트 취소가 아닌 사유로 반환하면
- **THEN** 나머지 루프도 정지하고 critical 알림이 발송되며 프로세스가 실패로 종료한다

#### Scenario: 정상 종료는 critical이 아니다
- **WHEN** SIGTERM으로 런타임이 graceful 종료하면
- **THEN** 루프가 취소·완주되고 journal이 정합하게 닫히며 critical 알림은 발송되지 않는다

#### Scenario: reconcile driver 지속 실패
- **WHEN** reconcile 사이클이 연속 5회 실패하면
- **THEN** critical 알림과 함께 ENTRY_BLOCKED로 자동 강화되고 루프는 재시도를 계속한다

#### Scenario: 검증 실행 중 기동 시도
- **WHEN** verify runlock이 신선한 상태에서 `engine run`을 실행하면
- **THEN** 기동이 거부되고 검증 종료 후 재시도가 안내된다
