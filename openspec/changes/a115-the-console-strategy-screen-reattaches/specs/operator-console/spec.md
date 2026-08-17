# operator-console — a115 delta

> a109 A2 P1-1: 콘솔 boot가 전략 projection dial 실패를 nil로 접어, 구분할 줄 아는
> page가 구분할 기회를 잃는다. 접힘을 제거하고 httpapi와 같은 재부착·오귀속 금지를
> 콘솔 전략 화면에 세운다.

## ADDED Requirements

### Requirement: 콘솔 전략 화면은 미구성과 도달 불가를 구분하고 재부착한다

콘솔 전략 화면의 projection 읽기는 부팅 시 dial 실패를 미구성과 동일한 표현으로 접어서는 안 되며(SHALL NOT), 구성된 runtime에 도달하지 못하는 상태를 도달 불가로 구분해 표시해야 하고(SHALL), 엔진이 뒤늦게 뜨거나 재시작한 뒤 콘솔 재시작 없이 재부착해야 하며(SHALL), 그 표시는 엔진 부재를 단정해서는 안 된다(SHALL NOT).

#### Scenario: 구성된 runtime, 늦게 뜨는 엔진

- **WHEN** 전략 runtime이 구성돼 있으나 엔진이 내려간 상태에서 콘솔이 기동하면
- **THEN** 전략 화면은 미구성(NOT_CONFIGURED)이 아니라 도달 불가를 표시하고, 엔진
  기동 후 콘솔 재시작 없이 실제 runtime 상태로 회복한다

#### Scenario: 미구성은 그대로 미구성이다

- **WHEN** 전략 runtime이 실제로 구성되지 않은 상태에서 콘솔이 기동하면
- **THEN** 전략 화면은 미구성을 표시하며 도달 불가로 오귀속하지 않는다
