# operator-console — a115 delta

> a109 A2 P1-1: 콘솔 boot가 전략 projection dial 실패를 nil로 접어, 구분할 줄 아는
> page가 구분할 기회를 잃는다. 접힘을 제거하고 httpapi와 같은 재부착·오귀속 금지를
> 콘솔 전략 화면에 세운다.

## ADDED Requirements

### Requirement: 콘솔 전략 화면은 미구성과 도달 불가를 구분하고 재부착한다

콘솔 전략 화면의 projection 읽기는 부팅 시 dial 실패를 미구성과 동일한 표현으로 접어서는 안 되며(SHALL NOT), endpoint descriptor 가 존재하는 runtime에 도달하지 못하는 상태를 도달 불가로 구분해 표시해야 하고(SHALL), 엔진이 뒤늦게 뜨거나 재시작한 뒤 콘솔 재시작 없이 재부착해야 하며(SHALL), 그 표시는 엔진 부재를 단정해서는 안 되고(SHALL NOT), 화면 요청을 처리하는 goroutine 은 dial 이나 endpoint probe 를 수행해서는 안 된다(SHALL NOT — 재부착 시도는 요청 경로 밖 백그라운드에서만).

#### Scenario: descriptor 가 남은 다운

- **WHEN** 전략 runtime endpoint descriptor 는 남아 있으나 엔진이 연결을 수락하지
  않는 상태(비정상 종료·행업)에서 콘솔이 기동하면
- **THEN** 전략 화면은 미구성(NOT_CONFIGURED)이 아니라 도달 불가를 표시하고, 엔진
  기동 후 콘솔 재시작 없이 실제 runtime 상태로 회복한다

#### Scenario: 미구성은 그대로 미구성이다

- **WHEN** 전략 runtime이 실제로 구성되지 않은 상태에서 콘솔이 기동하면
- **THEN** 전략 화면은 미구성을 표시하며 도달 불가로 오귀속하지 않는다

#### Scenario: descriptor 부재는 미기동 안내이고 재부착으로 회복한다 (선언된 한계)

- **WHEN** 엔진의 깨끗한 정지(descriptor 삭제) 뒤, 또는 autostart 가 endpoint 를
  아직 발행하기 전에 콘솔이 기동하면
- **THEN** 전략 화면은 「runtime endpoint 미기동」 dormant 를 표시하되 엔진 부재를
  단정하지 않으며, 엔진이 endpoint 를 발행하면 콘솔 재시작 없이 실제 runtime
  상태로 회복한다

#### Scenario: 가동 중 엔진 재시작

- **WHEN** 콘솔이 live 로 붙은 뒤 엔진이 재시작하면
- **THEN** 콘솔 재시작 없이, 화면을 열지 않아도 백그라운드 재시도가 새 endpoint 에
  다시 붙는다
