# exit-policy — a064 delta

> **ADDED만 쓴다.** 기존 "관측 경로와 fail-safe"는 관측이 **실패**했을 때를 규정하고,
> a064가 더하는 것은 판정이 **거부되도록 만들어진 순간**의 관측 의무다. 기존 SHALL과
> 충돌하지 않는다.

## ADDED Requirements

### Requirement: 판정 격리의 생성은 그 순간에 관측된다

시스템은 exit snapshot 격리를 만든 관측 사이클 안에서 그 사실을 critical 이벤트로 발행해야 한다 (SHALL). 그 이벤트에는 포지션·세대·quarantine version·사유·증거·격리 시각이 실려야 한다 (SHALL). 격리 생성이 다음 사이클의 판정 거부 알림으로만 드러나서는 안 된다 (MUST NOT — 판정 거부 알림은 포지션당 latch되므로 이미 다른 사유로 거부 중이던 포지션의 격리는 어떤 기록도 남기지 못한다).

exit 관측 루프는 실패한 사이클의 사유를 구조화 로그에 남겨야 한다 (SHALL). 사이클 실패 자체가 critical 알림이 되어서는 안 된다 (MUST NOT — 일시적 원장 오류가 outbox 전달 실패를 거쳐 ENTRY_BLOCKED로 이어지는 경로를 만들기 때문이다).

같은 격리에 대한 반복 발행은 억제되어야 하고 (SHALL), 해제 후 새 version으로 다시 격리되면 다시 발행되어야 한다 (SHALL).

#### Scenario: 격리가 만들어진 사이클
- **WHEN** 복구 후보를 하나로 정할 수 없어 판정 기록이 포지션을 격리한다
- **THEN** 같은 사이클에 quarantine version·사유·증거·격리 시각을 실은 critical 이벤트가 발행된다

#### Scenario: 이미 거부 중이던 포지션의 격리
- **WHEN** 판정 거부 알림이 이미 latch된 포지션이 격리된다
- **THEN** 격리 생성 이벤트는 그와 별개로 발행된다

#### Scenario: 같은 격리의 반복 관측
- **WHEN** 같은 quarantine version이 이어지는 사이클에서 계속 관측된다
- **THEN** 생성 이벤트는 다시 발행되지 않는다

#### Scenario: 해제 후 재격리
- **WHEN** 해제된 포지션이 같은 세대에서 새 quarantine version으로 다시 격리된다
- **THEN** 새 version에 대한 생성 이벤트가 발행된다

#### Scenario: 실패한 관측 사이클
- **WHEN** 관측 사이클이 판정 기록 실패로 끝난다
- **THEN** 그 사유가 구조화 로그에 남고, 그것만으로 신규 진입이 차단되지는 않는다
