# risk-management Specification (delta)

## MODIFIED Requirements

### Requirement: Kill switch와 운영 모드 — 모드×클래스 표
kill switch는 신규 진입 차단 전용(BLOCK-ONLY)이며 어떤 소비자도 강제청산을 유발하지 않는다(SHALL NOT). 운영 모드는 mutation safety class 허용 표로 정의된다(SHALL):

| 모드 | EXPOSURE_RAISING | RISK_REDUCING | PROTECTION_WEAKENING* |
|---|---|---|---|
| NORMAL | 허용 | 허용 | 허용(audit) |
| ENTRY_BLOCKED | 거부 | 허용 | 허용(audit) |
| HALT_ALL | 거부 | 허용 | **거부** |

*PROTECTION_WEAKENING의 발급·소비는 보호주문 도입 change가 정의하며 이 표는 열을 예약한다 — **현재는 landed `RecordDecision`이 전 모드에서 이 class를 거부하므로 허용(audit) 셀은 2c 발급 도입 후에 효력을 갖는다**. HALT_ALL은 이 change 안에서는 ENTRY_BLOCKED와 행동이 같지만, 운영자 전용 진입(자동 강화 없음)이라는 승인 의미와 2c의 PROTECTION_WEAKENING 거부로 구별되므로 유지한다. EXIT_ONLY는 두지 않는다 — ENTRY_BLOCKED와 행동이 동일해지므로(구별 근거 없는 모드는 §0.7 승인 사다리의 무의미 단계다), 실제 행동 차이가 생기는 change가 재도입한다. 수동 flatten-all은 모든 모드에서 통과한다(§0.3).

모드의 강제 지점은 EntryGate 투영이다(SHALL — 모드 전환은 journal 영속과 동시에 EntryGate 계좌 latch로 투영되고, Gateway의 기존 제출 시 재검사가 이를 소비한다; 봉인된 제출 시퀀스는 변경되지 않는다). 전환 승인은 방향 비대칭이다(SHALL): 보수 방향(NORMAL→ENTRY_BLOCKED→HALT_ALL)은 자동·즉시·durable, 완화는 사람 승인(§0.7)+audit. 자동 강화 트리거와 목적 상태의 열거(SHALL): 일일 손실 한도 도달 → ENTRY_BLOCKED, 자격증명 실패(401/403) → ENTRY_BLOCKED, critical 알림 outbox 전달 실패 지속 → ENTRY_BLOCKED, exit 관측 두절 임계 초과 → ENTRY_BLOCKED, **reconcile 사이클 지속 실패(연속 5주기) → ENTRY_BLOCKED, 체결 감지 사이클 지속 실패(연속 5주기) → ENTRY_BLOCKED**(엔진 런타임 change — 대사·체결에 눈이 먼 엔진은 새 진입을 받으면 안 된다) — 전부 메인 스펙의 "신규 진입 차단"과 정합하며 HALT_ALL 자동 진입은 없다(SHALL NOT — HALT_ALL은 운영자 결정). 분석·성과 작업 실패는 트리거가 아니다(SHALL NOT — 대사·체결 감지는 실행 경로이지 분석이 아니다). 모드·kill switch·이력은 journal 영속·재시작 유지(SHALL), 동시 적용 시 보수 우선(SHALL).

#### Scenario: 손실 한도 도달 시 자동 강화
- **WHEN** 일일 손실 한도 도달로 자동 강화가 발동하면
- **THEN** 사람 승인 없이 ENTRY_BLOCKED로 즉시 전환·영속되고 EntryGate에 투영되며 알림이 발송된다

#### Scenario: HALT_ALL 중 청산
- **WHEN** HALT_ALL 상태에서 RISK_REDUCING 청산이 요청되면
- **THEN** 허용된다 (수동 flatten-all 포함)

#### Scenario: 분석 작업 실패는 비트리거
- **WHEN** 성과 집계 작업이 반복 실패하면
- **THEN** 운영 모드는 변하지 않고 분석 재시도만 수행된다

#### Scenario: 재시작 후 모드 유지
- **WHEN** ENTRY_BLOCKED에서 프로세스가 재시작되면
- **THEN** journal에서 복원되어 EntryGate 투영과 함께 유지된다

#### Scenario: 모드 완화 시도
- **WHEN** ENTRY_BLOCKED에서 NORMAL로 되돌리려 하면
- **THEN** 사람 승인 절차와 audit 기록을 거쳐야 한다

#### Scenario: 대사 지속 실패 시 자동 강화
- **WHEN** reconcile 사이클이 연속 5회 실패하면
- **THEN** 사람 승인 없이 ENTRY_BLOCKED로 전환·영속되고 critical 알림이 발송되며 루프는 재시도를 계속한다
