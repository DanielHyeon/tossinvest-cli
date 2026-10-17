# operator-console — a114 delta

> a109 D4가 httpapi에 세운 "재부착은 소비자의 것" 계약을 콘솔의 엔진 lifecycle
> 읽기·명령에도 적용한다. 부팅 1회 dial(a109 freeze P2-7)은 엔진 재시작 후의 콘솔을
> 수동 재시작 없이는 눈멀게 한다. freeze 리뷰 P1-3 에 따라 명령 경로의 불변식을 요구로 올린다.

## ADDED Requirements

### Requirement: 콘솔의 엔진 lifecycle 읽기는 부팅 순서와 엔진 재시작에서 독립이다

콘솔의 engine lifecycle(포지션 정책 control plane) client는 콘솔 기동 시 엔진 부재·이후의 엔진 재시작 어느 쪽에서도 콘솔 재시작 없이 회복해야 하며(SHALL), 그 재시도는 렌더·요청 경로 밖의 백그라운드 single-flight여야 하고 화면 요청이 없어도 진행되어야 하며(SHALL), 렌더·요청 경로에서 lifecycle endpoint에 connect를 수반하는 dial을 수행해서는 안 된다(SHALL NOT — 매 읽기 재해결하는 별도 runtime 읽기 endpoint는 이 요구의 대상이 아니다). 엔진이 판정해 돌려준 거절은 endpoint 탈착으로 취급해서는 안 되며(SHALL NOT), Preview·Apply·격리 해제 명령은 호출당 현재 자리의 client로 최대 한 번만 전송되어야 하고 실패 후 새 client로 재전송해서는 안 된다(SHALL NOT). 격리 해제 표면은 재부착 wrapper를 거쳐서도 발견 가능해야 하며(SHALL), 붙기 전 상태를 "이 빌드에 배선되지 않았다"는 표현으로 보고해서는 안 되고(SHALL NOT), 정책 화면의 미배선 표시는 콘솔이 엔진 디렉터리를 해석하지 못한 경우에만 써야 한다(SHALL).

#### Scenario: 엔진이 콘솔보다 늦게 뜬다

- **WHEN** 엔진이 내려간 상태에서 콘솔이 기동하고 이후 엔진이 기동하면
- **THEN** 콘솔은 재시작 없이, 화면 요청이 없더라도 lifecycle client에 부착되고 화면은 엔진
  상태를 반영한다

#### Scenario: 가동 중 엔진 재시작

- **WHEN** 콘솔이 부착된 상태에서 엔진이 재시작하면
- **THEN** 콘솔은 재시작 없이 재부착되고 그 사이 화면은 부재를 상태로 표시하되
  엔진 부재를 단정하는 문구를 쓰지 않는다

#### Scenario: 엔진이 답한 거절은 탈착이 아니다

- **WHEN** 부착된 엔진이 명령을 판정해 거절(version 충돌·capability 만료·내부 오류 코드 등)하면
- **THEN** 그 거절은 운영자에게 그대로 전달되고 콘솔은 탈착 로그나 재-dial을 만들지 않는다
- **AND** 엔진이 토큰을 거절하면(재시작한 엔진의 다른 토큰) 그것은 탈착으로 취급해 재부착한다

#### Scenario: 명령은 다시 보내지 않는다

- **WHEN** Apply 또는 격리 해제 호출이 전송 실패로 끝나고 이후 재부착이 성립하면
- **THEN** 그 명령은 옛 client로 정확히 한 번 전송된 것뿐이며 새 client로 재전송되지 않는다

#### Scenario: 격리 해제는 wrapper 뒤에서도 보인다

- **WHEN** 콘솔이 부착 전이거나 부착된 상태에서 격리 해제 표면을 찾으면
- **THEN** 부착된 엔진이 격리 해제를 제공하면 그대로 전달되고, 부착 전에는 "배선되지 않았다"가
  아니라 붙지 않은 상태로 보고된다
