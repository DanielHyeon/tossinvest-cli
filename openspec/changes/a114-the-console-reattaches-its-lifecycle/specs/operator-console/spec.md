# operator-console — a114 delta

> a109 D4가 httpapi에 세운 "재부착은 소비자의 것" 계약을 콘솔의 엔진 lifecycle
> 읽기에도 적용한다. 부팅 1회 dial(a109 freeze P2-7)은 엔진 재시작 후의 콘솔을
> 수동 재시작 없이는 눈멀게 한다.

## ADDED Requirements

### Requirement: 콘솔의 엔진 lifecycle 읽기는 부팅 순서와 엔진 재시작에서 독립이다

콘솔의 engine lifecycle 읽기는 콘솔 기동 시 엔진 부재·이후의 엔진 재시작 어느 쪽에서도 콘솔 재시작 없이 회복해야 하며(SHALL), 그 재시도는 렌더·요청 경로 밖의 백그라운드 single-flight여야 하고(SHALL), 렌더·요청 경로에서 connect를 수반하는 dial을 수행해서는 안 된다(SHALL NOT).

#### Scenario: 엔진이 콘솔보다 늦게 뜬다

- **WHEN** 엔진이 내려간 상태에서 콘솔이 기동하고 이후 엔진이 기동하면
- **THEN** 콘솔은 재시작 없이 lifecycle 읽기에 부착되고 화면은 엔진 상태를 반영한다

#### Scenario: 가동 중 엔진 재시작

- **WHEN** 콘솔이 부착된 상태에서 엔진이 재시작하면
- **THEN** 콘솔은 재시작 없이 재부착되고 그 사이 화면은 부재를 상태로 표시하되
  엔진 부재를 단정하는 문구를 쓰지 않는다
