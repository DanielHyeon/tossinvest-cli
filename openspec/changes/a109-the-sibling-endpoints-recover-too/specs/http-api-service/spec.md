# http-api-service — a109 delta

> a108 gstack 리뷰가 이관한 기본값을 닫는다: 회복 부팅마다 전략 화면이 httpapi 수동
> 재시작 전까지 소실되는 문제(a108 red-team). 엔진측 in-process 재시도 금지(a108
> D2-2)는 유지되며, 재부착은 소비자인 이 데몬의 것이다.

## ADDED Requirements

### Requirement: 전략 화면은 엔진이 돌아오면 사람 개입 없이 다시 붙는다

httpapi는 전략 runtime reader가 부재이거나 직전 읽기가 실패한 상태일 때(부팅 시 live 부착 후의 엔진 재시작 포함) rate limit 아래 백그라운드 single-flight로 재부착을 시도해 엔진 복귀 후 데몬 재시작 없이 전략 화면을 회복해야 하며(SHALL), 요청 경로의 goroutine에서 dial·connect probe를 실행해서는 안 되고(SHALL NOT), 재부착 전의 응답 값은 기존 부재·unavailable 구분을 그대로 유지해야 한다(SHALL).

#### Scenario: 엔진이 httpapi보다 늦게 뜬다

- **WHEN** httpapi가 엔진보다 먼저 떠서 전략 reader 없이 가동 중이고 이후 엔진이
  endpoint를 발행하면
- **THEN** httpapi는 재시작 없이 rate limit 안에서 전략 화면을 회복한다

#### Scenario: 가동 중 엔진 재시작 후 복귀

- **WHEN** httpapi가 live client를 부착한 뒤 엔진이 재시작해(새 socket·새 토큰) 전략
  읽기가 실패하기 시작하고 이후 엔진이 돌아오면
- **THEN** httpapi는 재시작 없이 전략 화면을 회복하고, 회복 전 요청은 dial·probe를
  거치지 않는 코드 경로로 현재 강등 값을 즉시 반환한다
