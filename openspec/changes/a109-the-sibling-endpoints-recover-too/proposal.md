# a109 — 형제 endpoint들도 회복한다

> **상태: 등록만 먼저 했다(2026-08-14).** a108의 A1 적대 리뷰가 실측으로 연 change다.
> **착수는 a108 land 이후다** — a108이 strategy projection에 세운 계약(stage+rename 발행,
> connect probe 생존 판정, 회수의 전체성)을 같은 의례로 이식하는 작업이라, 그 계약이
> 먼저 굳어야 한다.

## Why

a108은 2026-08-13 재부팅 사고의 endpoint(strategy projection)만 고쳤다. A1 리뷰가 나머지
세 엔진 endpoint(position policy command·position policy runtime·alert control)를 실측한
결과, **같은 병이 표면만 다르게 존재한다**:

- **pre-chmod socket 잔재의 영구 거부** — policy runtime과 alert control은
  `PreparePrivateSocket`/`PrepareRuntimeSocket`이 정확-0600이 아닌 socket을 거부한다.
  listen→chmod 사이의 죽음이 umask 077(컨테이너 실측)에서 0700 socket을 남기면 매 부팅이
  거부되고, 세 endpoint 모두 `cmd/tossctl/engine.go`(:272/:277/:294 부근)에서 **fatal**이라
  a108 이전의 strategy projection과 똑같이 엔진 기동 루프 전체가 죽는다. A1이 각 3회
  반복으로 재현했다.
- **alert control은 살아 있는 주인의 socket 위에 두 번째 서버가 올라선다** — strategy
  projection이 a108에서 세운 규칙(수락 중인 socket은 건드리지 않는다)과 정반대다.
  실제 방어는 journal flock뿐인데 코드도 문서도 그것을 근거로 들지 않는다.

a108의 tasks 2.5가 이것을 잡지 못한 이유도 기록돼 있다: 핀이 관용이 자명한 잔재 모양만
만들어 **실패할 수 없었다**(a108 review.md §1 F4). 사고급 모양(pre-chmod socket·산 주인)의
핀은 이 change가 소유한다.

## What Changes

- 세 endpoint의 socket 발행을 a108과 같은 stage+rename 의례로 통일한다(pre-chmod 상태의
  원천 소거).
- 잔재 회수를 a108의 전체성 기준으로 맞춘다 — 검증-사망 잔재의 좁은 perm 완화 포함.
- alert control의 산-주인 탈취를 거부로 바꾸거나, journal flock이 유일한 방어임을 코드
  주석·테스트로 명문화한다(설계에서 결정).
- 사고급 모양의 crash-shape 핀(pre-chmod socket·0바이트 descriptor·산 주인)을 세 endpoint
  전부에 깐다 — a108 2.5의 실패-불가 핀을 대체한다.

## Non-goals

- strategy projection endpoint — a108이 소유한다.
- `engine.go`의 세 fatal을 강등으로 바꾸는 것 — policy command server는 조회 전용이
  아니라 **명령 표면**이므로 a108 D3의 "조회 전용 endpoint" 논증이 그대로 적용되지
  않는다. 강등 가능 여부는 이 change의 설계 단계에서 endpoint별로 따로 판정한다.

## Impact

- Affected specs: `engine-safety` (ADDED 1 — 일반 불변식; a108이 좁힌 요구의 일반형)
- Affected code: `internal/positionpolicyrpc`, `internal/app/engine`의 세 transport
- High-risk: 엔진 기동 경로 — FLM·Pre-Edit는 착수 시.
